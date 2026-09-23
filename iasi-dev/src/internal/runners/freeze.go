package runners

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"iasi-dev/internal/args"
	"iasi-dev/internal/cli"
	"iasi-dev/internal/commands"
	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

const freezeManifestName = ".iasi-freeze.json"

type freezeManifest struct {
	Organization string            `json:"organization"`
	Version      string            `json:"version"`
	Repositories map[string]string `json:"repositories"`
}

// Freeze closes the current development-organization version into an immutable
// sibling workspace. The current VERSION is used as-is; choosing the next
// development version is a separate, explicit `version vX.Y.Z` operation.
func Freeze(Parms *structures.Parms) []string {
	if _, ok := parseSemanticVersion(Parms.Version); !ok {
		cli.Error(RC.Error, *Parms, "La versión actual no es válida: %s", Parms.Version)
	}

	root := organizationWorkspaceRoot(Parms)
	if root == "" {
		cli.Error(RC.Error, *Parms, "No se puede deducir el workspace completo de la organización.")
	}

	repositories := completeOrganizationRepositories(Parms, root)
	if len(repositories) == 0 {
		cli.Error(RC.Error, *Parms, "No se encontraron repositorios para congelar.")
	}

	destination := freezeDestination(root, Parms.Organization, Parms.Version)
	if _, err := os.Stat(destination); err == nil {
		cli.Error(RC.Error, *Parms, "La versión %s ya está congelada en %s.", Parms.Version, destination)
	} else if !os.IsNotExist(err) {
		cli.Error(RC.Error, *Parms, "No se puede comprobar el destino de freeze %s: %v", destination, err)
	}

	cli.Header(*Parms, "Freeze %s", Parms.Organization)
	cli.Info(*Parms, "Version: %s", Parms.Version)
	cli.Info(*Parms, "Source: %s", root)
	cli.Info(*Parms, "Destination: %s", destination)

	if Parms.DryRun {
		for _, repository := range repositories {
			name := filepath.Base(repository)
			cli.Preview("Command [%s]: git tag -a %s -m \"IASI organization version %s\"\n", repository, Parms.Version, Parms.Version)
			if !Parms.Local {
				cli.Preview("Command [%s]: git push origin %s\n", repository, Parms.Version)
			}
			cli.Preview("Command [%s]: git clone --no-hardlinks %s %s\n", filepath.Dir(freezeTemporary(destination)), repository, filepath.Join(freezeTemporary(destination), name))
		}
		cli.Preview("Freeze snapshot: %s\n", destination)
		return append([]string{}, repositories...)
	}

	validateFreezeWorkingTrees(Parms, repositories)
	createdTags := []string{}
	pushedTags := []string{}
	tagsCommitted := false
	defer func() {
		if tagsCommitted {
			return
		}
		rollbackFreezeRemoteTags(Parms, Parms.Version, pushedTags)
		rollbackFreezeLocalTags(Parms, Parms.Version, createdTags)
	}()
	ensureFreezeTags(Parms, repositories, Parms.Version, &createdTags, &pushedTags)

	temporary := freezeTemporary(destination)
	if err := os.RemoveAll(temporary); err != nil {
		cli.Error(RC.Error, *Parms, "No se pudo limpiar el workspace temporal %s: %v", temporary, err)
	}
	if err := os.MkdirAll(temporary, 0755); err != nil {
		cli.Error(RC.Error, *Parms, "No se pudo crear el workspace temporal %s: %v", temporary, err)
	}

	completed := false
	defer func() {
		if !completed {
			_ = os.RemoveAll(temporary)
		}
	}()

	manifest := freezeManifest{
		Organization: Parms.Organization,
		Version:      Parms.Version,
		Repositories: map[string]string{},
	}
	frozenRepositories := make([]string, 0, len(repositories))

	for _, repository := range repositories {
		name := filepath.Base(repository)
		destinationRepository := filepath.Join(temporary, name)
		cli.Step(*Parms, "Freezing %s", name)

		commit := repositoryHead(Parms, repository)
		manifest.Repositories[name] = commit

		if !freezeCloneRepository(Parms, repository, destinationRepository, Parms.Version) {
			cli.Error(RC.Error, *Parms, "No se pudo congelar %s. Revisa el log: %s", name, logName(*Parms))
		}
		frozenRepositories = append(frozenRepositories, filepath.Join(destination, name))
	}

	if err := writeFreezeManifest(temporary, manifest); err != nil {
		cli.Error(RC.Error, *Parms, "No se pudo escribir el manifest de freeze: %v", err)
	}
	if err := os.Rename(temporary, destination); err != nil {
		cli.Error(RC.Error, *Parms, "No se pudo cerrar la versión congelada en %s: %v", destination, err)
	}
	completed = true
	tagsCommitted = true

	cli.Success(*Parms, "Organization frozen at %s.", Parms.Version)
	return frozenRepositories
}

func completeOrganizationRepositories(Parms *structures.Parms, root string) []string {
	discovery := *Parms
	discovery.Targets = []string{root}
	discovery.RequestedTargets = nil
	discovery.TargetDetails = nil
	discovery.Repos = nil
	discovery.BlackList = nil
	args.Prepare(&discovery)
	return append([]string{}, discovery.Repos...)
}

func organizationWorkspaceRoot(Parms *structures.Parms) string {
	current, err := os.Getwd()
	if err == nil {
		for {
			if strings.EqualFold(filepath.Base(current), Parms.Organization) {
				return filepath.Clean(current)
			}
			parent := filepath.Dir(current)
			if parent == current {
				break
			}
			current = parent
		}
	}
	return workflowOrganizationRoot(Parms.Repos)
}

func freezeDestination(root string, organization string, version string) string {
	stable := strings.TrimSuffix(organization, "-dev")
	if stable == "" || stable == organization {
		stable = organization
	}
	return filepath.Join(filepath.Dir(filepath.Clean(root)), stable+"-"+version)
}

func freezeTemporary(destination string) string {
	return filepath.Join(filepath.Dir(destination), "."+filepath.Base(destination)+".freeze.tmp")
}

func validateFreezeWorkingTrees(Parms *structures.Parms, repositories []string) {
	for _, repository := range repositories {
		result := commands.Run(repository, Parms.LogFile, "git", "status", "--porcelain")
		if result.RC != RC.OK {
			cli.Error(RC.Error, *Parms, "No se pudo comprobar el estado de %s", repository)
		}
		if strings.TrimSpace(result.Stdout) != "" {
			cli.Error(RC.Error, *Parms, "No se puede congelar %s: hay cambios sin commit.", repository)
		}
	}
}

func repositoryHead(Parms *structures.Parms, repository string) string {
	result := commands.Run(repository, Parms.LogFile, "git", "rev-parse", "HEAD")
	if result.RC != RC.OK {
		cli.Error(RC.Error, *Parms, "No se pudo obtener HEAD de %s", repository)
	}
	return strings.TrimSpace(result.Stdout)
}

func ensureFreezeTags(Parms *structures.Parms, repositories []string, version string, created *[]string, pushed *[]string) {
	for _, repository := range repositories {
		head := repositoryHead(Parms, repository)
		result := commands.Run(repository, Parms.LogFile, "git", "rev-list", "-n", "1", version)
		if result.RC == RC.OK {
			if strings.TrimSpace(result.Stdout) != head {
				cli.Error(RC.Error, *Parms, "El tag %s de %s no apunta a HEAD.", version, repository)
			}
		} else {
			message := "IASI organization version " + version
			if !promoteCommand(Parms, repository, "tag", "-a", version, "-m", message) {
				cli.Error(RC.Error, *Parms, "No se pudo crear el tag %s en %s", version, repository)
			}
			*created = append(*created, repository)
		}

		if Parms.Local {
			continue
		}

		remoteCommit, exists := freezeRemoteTagCommit(Parms, repository, version)
		if exists {
			if remoteCommit != head {
				cli.Error(RC.Error, *Parms, "El tag remoto %s de %s no apunta a HEAD.", version, repository)
			}
			continue
		}

		if !promoteCommand(Parms, repository, "push", "origin", version) {
			cli.Error(RC.Error, *Parms, "No se pudo publicar el tag %s en %s", version, repository)
		}
		*pushed = append(*pushed, repository)
	}
}

func freezeRemoteTagCommit(Parms *structures.Parms, repository string, version string) (string, bool) {
	peeled := "refs/tags/" + version + "^{}"
	result := commands.Run(repository, Parms.LogFile, "git", "ls-remote", "--exit-code", "origin", peeled)
	if result.RC == RC.OK {
		fields := strings.Fields(result.Stdout)
		if len(fields) > 0 {
			return fields[0], true
		}
	}

	ref := "refs/tags/" + version
	result = commands.Run(repository, Parms.LogFile, "git", "ls-remote", "--exit-code", "origin", ref)
	if result.RC != RC.OK {
		return "", false
	}
	fields := strings.Fields(result.Stdout)
	if len(fields) == 0 {
		return "", false
	}
	return fields[0], true
}

func rollbackFreezeLocalTags(Parms *structures.Parms, version string, repositories []string) {
	for i := len(repositories) - 1; i >= 0; i-- {
		if !promoteCommand(Parms, repositories[i], "tag", "-d", version) {
			cli.ErrorMessage(*Parms, "No se pudo revertir el tag local %s en %s", version, repositories[i])
		}
	}
}

func rollbackFreezeRemoteTags(Parms *structures.Parms, version string, repositories []string) {
	for i := len(repositories) - 1; i >= 0; i-- {
		if !promoteCommand(Parms, repositories[i], "push", "origin", "--delete", version) {
			cli.ErrorMessage(*Parms, "No se pudo revertir el tag remoto %s en %s", version, repositories[i])
		}
	}
}

func freezeCloneRepository(Parms *structures.Parms, source string, destination string, version string) bool {
	result := commands.RunLogged(filepath.Dir(destination), Parms.LogFile, "git", "clone", "--no-hardlinks", source, destination)
	if result.RC != RC.OK {
		return false
	}

	origin := commands.Run(source, Parms.LogFile, "git", "remote", "get-url", "origin")
	if origin.RC != RC.OK {
		return false
	}
	if commands.RunLogged(destination, Parms.LogFile, "git", "remote", "set-url", "origin", strings.TrimSpace(origin.Stdout)).RC != RC.OK {
		return false
	}
	if commands.RunLogged(destination, Parms.LogFile, "git", "checkout", "--detach", version).RC != RC.OK {
		return false
	}
	return true
}

func writeFreezeManifest(root string, manifest freezeManifest) error {
	names := make([]string, 0, len(manifest.Repositories))
	for name := range manifest.Repositories {
		names = append(names, name)
	}
	sort.Strings(names)
	ordered := make(map[string]string, len(names))
	for _, name := range names {
		ordered[name] = manifest.Repositories[name]
	}
	manifest.Repositories = ordered

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(filepath.Join(root, freezeManifestName), data, 0644)
}

func readFreezeManifest(root string) (freezeManifest, error) {
	data, err := os.ReadFile(filepath.Join(root, freezeManifestName))
	if err != nil {
		return freezeManifest{}, err
	}
	manifest := freezeManifest{}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return freezeManifest{}, err
	}
	return manifest, nil
}

func validateFrozenOrganization(Parms *structures.Parms, root string, version string) []string {
	manifest, err := readFreezeManifest(root)
	if err != nil {
		cli.Error(RC.Error, *Parms, "No se puede leer el freeze %s: %v", root, err)
	}
	if manifest.Version != version {
		cli.Error(RC.Error, *Parms, "El freeze %s declara %s y se solicitó %s.", root, manifest.Version, version)
	}
	if manifest.Organization != Parms.Organization {
		cli.Error(RC.Error, *Parms, "El freeze %s pertenece a %s y no a %s.", root, manifest.Organization, Parms.Organization)
	}

	discovery := *Parms
	discovery.Targets = []string{root}
	discovery.RequestedTargets = nil
	discovery.TargetDetails = nil
	discovery.Repos = nil
	discovery.BlackList = nil
	args.Prepare(&discovery)

	if len(discovery.Repos) != len(manifest.Repositories) {
		cli.Error(RC.Error, *Parms, "El freeze %s no contiene el conjunto completo de repositorios.", root)
	}

	for _, repository := range discovery.Repos {
		name := filepath.Base(repository)
		expected, ok := manifest.Repositories[name]
		if !ok {
			cli.Error(RC.Error, *Parms, "El repositorio %s no figura en el manifest de freeze.", name)
		}
		if Parms.DryRun {
			continue
		}
		if head := repositoryHead(Parms, repository); head != expected {
			cli.Error(RC.Error, *Parms, "El repositorio congelado %s ha cambiado: %s != %s.", name, head, expected)
		}
		result := commands.Run(repository, Parms.LogFile, "git", "status", "--porcelain")
		if result.RC != RC.OK || strings.TrimSpace(result.Stdout) != "" {
			cli.Error(RC.Error, *Parms, "El repositorio congelado %s no está limpio.", name)
		}
	}

	return append([]string{}, discovery.Repos...)
}

func frozenOrganizationPath(Parms *structures.Parms, version string) string {
	root := organizationWorkspaceRoot(Parms)
	if root == "" {
		cli.Error(RC.Error, *Parms, "No se puede deducir el workspace de la organización.")
	}
	return freezeDestination(root, Parms.Organization, version)
}
