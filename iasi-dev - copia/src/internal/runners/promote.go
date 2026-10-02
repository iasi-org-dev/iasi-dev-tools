package runners

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"iasi-dev/internal/cli"
	"iasi-dev/internal/commands"
	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)


var promotePrepareRemote = promoteRemotePrepare
var promoteSynchronizeRemote = promoteRemoteSynchronize
var promoteSetDestinationVersion = setOrganizationVersionFor

type semanticVersion struct {
	major int
	minor int
	patch int
}

// Promote materializes one frozen organization snapshot without carrying development history.
// The requested version, source path and destination path are explicit; no workspace location is inferred.
func Promote(Parms *structures.Parms) []string {
	requireTargetVersion(Parms, "promote")
	if _, ok := parseSemanticVersion(Parms.TargetVersion); !ok {
		cli.Error(RC.Error, *Parms, "La versión a promover no es válida: %s", Parms.TargetVersion)
	}
	if Parms.SourcePath == "" || Parms.DestinationPath == "" {
		cli.Error(RC.Error, *Parms, "promote requiere source-path y destination-path explícitos.")
	}

	sourceRoot := filepath.Clean(Parms.SourcePath)
	destination := filepath.Clean(Parms.DestinationPath)
	validatePromotePaths(Parms, sourceRoot, destination)

	if Parms.DestinationOrganization == "" {
		Parms.DestinationOrganization = filepath.Base(destination)
	}

	cli.Header(*Parms, "Promote %s", Parms.TargetVersion)
	cli.Info(*Parms, "Source path: %s", sourceRoot)
	cli.Info(*Parms, "Destination path: %s", destination)
	if Parms.SourceOrganization != "" {
		cli.Info(*Parms, "Source organization: %s", Parms.SourceOrganization)
	}
	cli.Info(*Parms, "Destination organization: %s", Parms.DestinationOrganization)
	cli.Info(*Parms, "Validating...")

	sourceRepositories := completeOrganizationRepositories(Parms, sourceRoot)
	if len(sourceRepositories) == 0 {
		cli.Error(RC.Error, *Parms, "No se encontraron repositorios para promover en %s.", sourceRoot)
	}
	validatePromotionTags(Parms, sourceRepositories, Parms.TargetVersion)

	temporary := promoteTemporary(destination, Parms.TargetVersion)

	if Parms.DryRun {
		cli.Info(*Parms, "Preparing workspace...")
		cli.Preview("Promote workspace: %s\n", temporary)
		cli.Info(*Parms, "Promoting repositories...")
		for _, repository := range sourceRepositories {
			name := filepath.Base(repository)
			cli.Preview("Clone %s at %s: %s -> %s\n", name, Parms.TargetVersion, repository, filepath.Join(temporary, name))
			cli.Preview("Create history-free repository: %s -> https://github.com/%s/%s.git\n", name, Parms.DestinationOrganization, name)
		}
		cli.Info(*Parms, "Postprocessing...")
		promotePostprocessPreview(Parms)
		cli.Preview("Replace local destination: %s -> %s\n", temporary, destination)

		repositories := promoteDestinationRepositories(destination, sourceRepositories, Parms.SourceOrganization, Parms.DestinationOrganization)
		promoteRemotePreview(Parms, repositories)
		cli.Preview("Set %s VERSION to %s\n", Parms.DestinationOrganization, Parms.TargetVersion)
		return repositories
	}

	cli.Info(*Parms, "Preparing workspace...")
	if err := preparePromoteTemporary(temporary); err != nil {
		cli.Error(RC.Error, *Parms, "No se pudo preparar el workspace temporal %s: %v", temporary, err)
	}

	completed := false
	defer func() {
		if !completed {
			_ = os.RemoveAll(temporary)
		}
	}()

	cli.Info(*Parms, "Promoting repositories...")
	promotedTemporary := promoteRepositories(Parms, sourceRepositories, temporary, Parms.TargetVersion)

	cli.Info(*Parms, "Postprocessing...")
	promotedTemporary = promotePostprocess(Parms, temporary, promotedTemporary)

	if err := promoteReplace(temporary, destination); err != nil {
		cli.Error(RC.Error, *Parms, "No se pudo instalar la promoción en %s: %v", destination, err)
	}
	completed = true

	repositories := promoteDestinationRepositories(destination, sourceRepositories, Parms.SourceOrganization, Parms.DestinationOrganization)
	Parms.Repos = repositories

	if Parms.Local {
		cli.Info(*Parms, "Preparing remote organization...")
		promotePrepareRemote(Parms, repositories)
	} else {
		cli.Info(*Parms, "Synchronizing remote organization...")
		promoteSynchronizeRemote(Parms, repositories)
	}

	cli.Info(*Parms, "Propagating version...")
	promoteSetDestinationVersion(Parms, Parms.DestinationOrganization, Parms.TargetVersion)

	cli.Success(*Parms, "Promoted %s.", Parms.TargetVersion)
	return repositories
}

func validatePromotePaths(Parms *structures.Parms, source string, destination string) {
	if strings.EqualFold(filepath.Clean(source), filepath.Clean(destination)) {
		cli.Error(RC.Error, *Parms, "source-path y destination-path deben ser rutas distintas.")
	}
	if promotePathContains(source, destination) || promotePathContains(destination, source) {
		cli.Error(RC.Error, *Parms, "source-path y destination-path no pueden contenerse entre sí.")
	}

	info, err := os.Stat(source)
	if err != nil || !info.IsDir() {
		cli.Error(RC.Error, *Parms, "source-path no existe o no es un directorio: %s", source)
	}
	if info, err := os.Stat(destination); err == nil && !info.IsDir() {
		cli.Error(RC.Error, *Parms, "destination-path existe pero no es un directorio: %s", destination)
	} else if err != nil && !os.IsNotExist(err) {
		cli.Error(RC.Error, *Parms, "No se puede acceder a destination-path %s: %v", destination, err)
	}
}

func promotePathContains(parent string, child string) bool {
	relative, err := filepath.Rel(filepath.Clean(parent), filepath.Clean(child))
	if err != nil || relative == "." || relative == ".." {
		return false
	}
	return !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func promoteTemporary(destination string, version string) string {
	name := filepath.Base(filepath.Clean(destination))
	return filepath.Join(os.TempDir(), "."+name+".promote-"+version+".tmp")
}

func promoteBackup(destination string) string {
	name := filepath.Base(filepath.Clean(destination))
	return filepath.Join(os.TempDir(), "."+name+".promote.bak")
}

func preparePromoteTemporary(temporary string) error {
	if err := os.RemoveAll(temporary); err != nil {
		return err
	}
	return os.MkdirAll(temporary, 0755)
}

func validatePromotionTags(Parms *structures.Parms, repositories []string, version string) {
	ref := "refs/tags/" + version
	for _, repository := range repositories {
		result := commands.Run(repository, Parms.LogFile, "git", "rev-parse", "--verify", ref)
		if result.RC != RC.OK {
			cli.Error(RC.Error, *Parms, "El tag %s no existe en %s.", version, filepath.Base(repository))
		}
	}
}

func promoteRepositories(Parms *structures.Parms, repositories []string, root string, version string) []string {
	promoted := make([]string, 0, len(repositories))
	for _, repository := range repositories {
		name := filepath.Base(repository)
		destination := filepath.Join(root, name)
		cli.Step(*Parms, "Cloning %s...", name)
		result := commands.RunLogged(root, Parms.LogFile, "git", "clone", "--branch", version, "--single-branch", repository, destination)
		if result.RC != RC.OK {
			cli.Error(RC.Error, *Parms, "No se pudo clonar %s desde el tag %s. Revisa el log: %s", name, version, logName(*Parms))
		}

		if err := os.RemoveAll(filepath.Join(destination, ".git")); err != nil {
			cli.Error(RC.Error, *Parms, "No se pudo eliminar la historia Git de %s: %v", name, err)
		}
		if !promoteGitInit(Parms, destination) {
			cli.Error(RC.Error, *Parms, "No se pudo inicializar Git en %s. Revisa el log: %s", name, logName(*Parms))
		}
		if !promoteDestinationRemote(Parms, destination, Parms.DestinationOrganization, name) {
			cli.Error(RC.Error, *Parms, "No se pudo configurar origin en %s. Revisa el log: %s", name, logName(*Parms))
		}
		if !promoteCommit(Parms, destination, version) {
			cli.Error(RC.Error, *Parms, "No se pudo crear el snapshot promovido de %s. Revisa el log: %s", name, logName(*Parms))
		}
		promoted = append(promoted, destination)
	}
	return promoted
}

func promoteGitInit(Parms *structures.Parms, repository string) bool {
	result := commands.RunLogged(repository, Parms.LogFile, "git", "init", "-b", "main")
	return result.RC == RC.OK
}

func promoteDestinationRemote(Parms *structures.Parms, repository string, organization string, name string) bool {
	remote := fmt.Sprintf("https://github.com/%s/%s.git", organization, name)
	result := commands.RunLogged(repository, Parms.LogFile, "git", "remote", "add", "origin", remote)
	return result.RC == RC.OK
}

func promoteCommit(Parms *structures.Parms, repository string, version string) bool {
	result := commands.RunLogged(repository, Parms.LogFile, "git", "add", "-A", ".")
	if result.RC != RC.OK {
		return false
	}
	result = commands.RunLogged(repository, Parms.LogFile, "git", "commit", "-m", "IASI organization version "+version)
	return result.RC == RC.OK
}

func promoteReplace(temporary string, destination string) error {
	backup := promoteBackup(destination)
	if err := os.RemoveAll(backup); err != nil {
		return err
	}

	destinationExists := false
	if info, err := os.Stat(destination); err == nil {
		if !info.IsDir() {
			return fmt.Errorf("el destino existe pero no es un directorio: %s", destination)
		}
		destinationExists = true
		if err := os.Rename(destination, backup); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	if err := os.Rename(temporary, destination); err != nil {
		if destinationExists {
			_ = os.Rename(backup, destination)
		}
		return err
	}

	if destinationExists {
		if err := os.RemoveAll(backup); err != nil {
			return err
		}
	}
	return nil
}

func promoteDestinationRepositories(destination string, sourceRepositories []string, sourceOrganization string, destinationOrganization string) []string {
	repositories := make([]string, 0, len(sourceRepositories))
	for _, repository := range sourceRepositories {
		name := promotePostprocessRepositoryName(filepath.Base(repository), sourceOrganization, destinationOrganization)
		repositories = append(repositories, filepath.Join(destination, name))
	}
	return repositories
}

func formatCommandArguments(args []string) string {
	formatted := make([]string, 0, len(args))
	for _, arg := range args {
		if strings.ContainsAny(arg, " \t\"") {
			formatted = append(formatted, strconv.Quote(arg))
			continue
		}
		formatted = append(formatted, arg)
	}
	return strings.Join(formatted, " ")
}

func promoteCommand(Parms *structures.Parms, repository string, args ...string) bool {
	cli.VeryVerbose(*Parms, "%s: git %s", filepath.Base(repository), formatCommandArguments(args))
	result := commands.RunLogged(repository, Parms.LogFile, "git", args...)
	return result.RC == RC.OK
}

func requireTargetVersion(Parms *structures.Parms, operation string) {
	if Parms.TargetVersion == "" {
		cli.Error(RC.Error, *Parms, "%s requiere una versión.", operation)
	}
}

func parseSemanticVersion(value string) (semanticVersion, bool) {
	if !strings.HasPrefix(value, "v") {
		return semanticVersion{}, false
	}

	parts := strings.Split(strings.TrimPrefix(value, "v"), ".")
	if len(parts) != 3 {
		return semanticVersion{}, false
	}

	values := make([]int, 3)
	for i, part := range parts {
		if part == "" || (len(part) > 1 && part[0] == '0') {
			return semanticVersion{}, false
		}

		number, err := strconv.Atoi(part)
		if err != nil || number < 0 {
			return semanticVersion{}, false
		}
		values[i] = number
	}

	return semanticVersion{major: values[0], minor: values[1], patch: values[2]}, true
}

func compareSemanticVersions(left semanticVersion, right semanticVersion) int {
	if left.major != right.major {
		if left.major < right.major {
			return -1
		}
		return 1
	}
	if left.minor != right.minor {
		if left.minor < right.minor {
			return -1
		}
		return 1
	}
	if left.patch != right.patch {
		if left.patch < right.patch {
			return -1
		}
		return 1
	}
	return 0
}
