package runners

import (
	"os"
	"path/filepath"
	"strings"

	"iasi-dev/internal/args"
	"iasi-dev/internal/cli"
	"iasi-dev/internal/commands"
	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

// Freeze closes the current development-organization version by tagging the
// current HEAD of every repository. A freeze is always published to origin;
// it never creates a copied or materialized workspace.
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

	validateFreezeVersion(Parms, repositories, Parms.Version)

	cli.Header(*Parms, "Freeze %s", Parms.Organization)
	cli.Info(*Parms, "Version: %s", Parms.Version)
	cli.Info(*Parms, "Workspace: %s", root)
	cli.Info(*Parms, "Mode: tags")

	if Parms.DryRun {
		for _, repository := range repositories {
			cli.Preview("Command [%s]: git tag -a %s -m \"IASI organization version %s\"\n", repository, Parms.Version, Parms.Version)
			cli.Preview("Command [%s]: git push origin %s\n", repository, Parms.Version)
		}
		return append([]string{}, repositories...)
	}

	validateFreezeWorkingTrees(Parms, repositories)

	createdTags := []string{}
	pushedTags := []string{}
	committed := false
	defer func() {
		if committed {
			return
		}
		rollbackFreezeRemoteTags(Parms, Parms.Version, pushedTags)
		rollbackFreezeLocalTags(Parms, Parms.Version, createdTags)
	}()

	ensureFreezeTags(Parms, repositories, Parms.Version, &createdTags, &pushedTags)
	committed = true

	cli.Success(*Parms, "Organization frozen at %s.", Parms.Version)
	return append([]string{}, repositories...)
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

func validateFreezeVersion(Parms *structures.Parms, repositories []string, version string) {
	candidate, ok := parseSemanticVersion(version)
	if !ok {
		cli.Error(RC.Error, *Parms, "La versión actual no es válida: %s", version)
	}

	var highest semanticVersion
	highestTag := ""
	highestRepository := ""

	for _, repository := range repositories {
		for _, tag := range freezeRepositoryTags(Parms, repository) {
			parsed, ok := parseSemanticVersion(tag)
			if !ok {
				continue
			}
			if highestTag == "" || compareSemanticVersions(parsed, highest) > 0 {
				highest = parsed
				highestTag = tag
				highestRepository = repository
			}
		}
	}

	if highestTag != "" && compareSemanticVersions(candidate, highest) <= 0 {
		cli.Error(
			RC.Error,
			*Parms,
			"La versión %s debe ser superior a todos los tags existentes; el mayor es %s en %s.",
			version,
			highestTag,
			filepath.Base(highestRepository),
		)
	}
}

func freezeRepositoryTags(Parms *structures.Parms, repository string) []string {
	tags := map[string]bool{}

	local := commands.Run(repository, Parms.LogFile, "git", "tag", "--list")
	if local.RC != RC.OK {
		cli.Error(RC.Error, *Parms, "No se pudieron leer los tags locales de %s", repository)
	}
	for _, tag := range strings.Fields(local.Stdout) {
		tags[tag] = true
	}

	remote := commands.Run(repository, Parms.LogFile, "git", "ls-remote", "--tags", "origin")
	if remote.RC != RC.OK {
		cli.Error(RC.Error, *Parms, "No se pudieron leer los tags publicados de %s", repository)
	}
	for _, line := range strings.Split(remote.Stdout, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		ref := strings.TrimPrefix(fields[1], "refs/tags/")
		ref = strings.TrimSuffix(ref, "^{}")
		if ref != "" {
			tags[ref] = true
		}
	}

	result := make([]string, 0, len(tags))
	for tag := range tags {
		result = append(result, tag)
	}
	return result
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

func ensureFreezeTags(Parms *structures.Parms, repositories []string, version string, created *[]string, pushed *[]string) {
	for _, repository := range repositories {
		message := "IASI organization version " + version
		if !promoteCommand(Parms, repository, "tag", "-a", version, "-m", message) {
			cli.Error(RC.Error, *Parms, "No se pudo crear el tag %s en %s", version, repository)
		}
		*created = append(*created, repository)

		if _, exists := freezeRemoteTagCommit(Parms, repository, version); exists {
			cli.Error(RC.Error, *Parms, "El tag %s ya existe en origin para %s.", version, repository)
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
