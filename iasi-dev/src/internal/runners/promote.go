package runners

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"iasi-dev/internal/cli"
	"iasi-dev/internal/commands"
	"iasi-dev/internal/consts"
	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

type semanticVersion struct {
	major int
	minor int
	patch int
}

// Promote creates a new stable organization version from the current organization state.
// With -p it performs only the publication step for the current local organization.
func Promote(Parms *structures.Parms) []string {
	if Parms.Push {
		workflowPromotePush(Parms)
		return append([]string{}, Parms.Repos...)
	}

	requireTargetVersion(Parms, "promote")

	current, ok := parseSemanticVersion(Parms.Version)
	if !ok {
		cli.Error(RC.Error, *Parms, "La versión actual no es válida: %s", Parms.Version)
	}

	target, ok := parseSemanticVersion(Parms.TargetVersion)
	if !ok {
		cli.Error(RC.Error, *Parms, "La versión destino no es válida: %s", Parms.TargetVersion)
	}

	if len(Parms.Repos) == 0 {
		cli.Error(RC.Error, *Parms, "No se encontraron repositorios para promover.")
	}

	validatePromoteWorkingTrees(Parms)

	highest, highestText := repositoryHighestVersion(Parms, Parms.Repos[0], current, Parms.Version)

	for _, repository := range Parms.Repos[1:] {
		repositoryVersion, repositoryVersionText := repositoryHighestVersion(Parms, repository, current, Parms.Version)
		if compareSemanticVersions(repositoryVersion, highest) != 0 {
			cli.Error(
				RC.Error,
				*Parms,
				"La organización no está alineada: %s está en %s y se esperaba %s",
				repository,
				repositoryVersionText,
				highestText,
			)
		}
	}

	fmt.Printf("Organization: %s\n", Parms.Organization)
	fmt.Printf("Current version: %s\n", Parms.Version)
	fmt.Printf("Highest version: %s\n", highestText)
	fmt.Printf("Target version: %s\n", Parms.TargetVersion)

	comparison := compareSemanticVersions(target, highest)
	if comparison < 0 {
		cli.Error(RC.Error, *Parms, "La versión destino %s no puede ser inferior a %s", Parms.TargetVersion, highestText)
	}
	if comparison == 0 {
		if Parms.Version != Parms.TargetVersion || !repositoriesHaveTag(Parms, Parms.TargetVersion) {
			cli.Error(RC.Error, *Parms, "La versión destino %s ya existe, pero la organización no está en un estado reanudable.", Parms.TargetVersion)
		}
		fmt.Printf("Version validation: already promoted; resuming\n")
		return append([]string{}, Parms.Repos...)
	}

	fmt.Printf("Version validation: OK\n")

	promoteTags(Parms)
	setOrganizationVersion(Parms)
	Parms.Version = Parms.TargetVersion

	cli.Success(*Parms, "Organization promoted to %s.", Parms.TargetVersion)
	return append([]string{}, Parms.Repos...)
}

// promoteTags creates the organization version tag in every repository and publishes it.
// A partial transaction is rolled back both remotely and locally.

func validatePromoteWorkingTrees(Parms *structures.Parms) {
	for _, repository := range Parms.Repos {
		result := commands.Run(repository, Parms.LogFile, "git", "status", "--porcelain")
		if result.RC != RC.OK {
			cli.Error(RC.Error, *Parms, "No se pudo comprobar el estado de %s", repository)
		}
		if strings.TrimSpace(result.Stdout) != "" {
			cli.Error(RC.Error, *Parms, "No se puede promover %s: hay cambios sin commit.", repository)
		}
	}
}

func repositoriesHaveTag(Parms *structures.Parms, version string) bool {
	ref := "refs/tags/" + version
	for _, repository := range Parms.Repos {
		result := commands.Run(repository, Parms.LogFile, "git", "show-ref", "--verify", "--quiet", ref)
		if result.RC != RC.OK {
			return false
		}
	}
	return true
}

func promoteTags(Parms *structures.Parms) {
	created := []string{}
	message := "IASI organization version " + Parms.TargetVersion

	for _, repository := range Parms.Repos {
		if !promoteCommand(Parms, repository, "tag", "-a", Parms.TargetVersion, "-m", message) {
			cli.ErrorMessage(*Parms, "No se pudo crear el tag %s en %s", Parms.TargetVersion, repository)
			rollbackLocalTags(Parms, Parms.TargetVersion, created)
			cli.Abort(RC.Error, *Parms)
		}
		created = append(created, repository)
	}

	pushed := []string{}
	for _, repository := range Parms.Repos {
		if !promoteCommand(Parms, repository, "push", "origin", Parms.TargetVersion) {
			cli.ErrorMessage(*Parms, "No se pudo publicar el tag %s en %s", Parms.TargetVersion, repository)
			rollbackRemoteTags(Parms, Parms.TargetVersion, pushed)
			rollbackLocalTags(Parms, Parms.TargetVersion, created)
			cli.Abort(RC.Error, *Parms)
		}
		pushed = append(pushed, repository)
	}
}

func rollbackLocalTags(Parms *structures.Parms, version string, repositories []string) {
	failed := false
	for i := len(repositories) - 1; i >= 0; i-- {
		if promoteCommand(Parms, repositories[i], "tag", "-d", version) {
			continue
		}
		failed = true
		cli.ErrorMessage(*Parms, "No se pudo eliminar el tag local %s en %s", version, repositories[i])
	}
	if failed {
		cli.ErrorMessage(*Parms, "El rollback local no se completó; el sistema puede haber quedado en un estado inconsistente.")
	}
}

func rollbackRemoteTags(Parms *structures.Parms, version string, repositories []string) {
	failed := false
	for i := len(repositories) - 1; i >= 0; i-- {
		if promoteCommand(Parms, repositories[i], "push", "origin", "--delete", version) {
			continue
		}
		failed = true
		cli.ErrorMessage(*Parms, "No se pudo eliminar el tag remoto %s en %s", version, repositories[i])
	}
	if failed {
		cli.ErrorMessage(*Parms, "El rollback remoto no se completó; el sistema puede haber quedado en un estado inconsistente.")
	}
}

func promoteCommand(Parms *structures.Parms, repository string, args ...string) bool {
	cli.VeryVerbose(*Parms, "%s: git %s", filepath.Base(repository), formatCommandArguments(args))
	result := commands.RunLogged(repository, Parms.LogFile, "git", args...)
	return result.RC == RC.OK
}

func setOrganizationVersion(Parms *structures.Parms) {
	// Use the REST update endpoint so changing VERSION does not accidentally change
	// the organization variable visibility configured in GitHub.
	endpoint := fmt.Sprintf("orgs/%s/actions/variables/%s", Parms.Organization, consts.OrganizationVersionVariable)
	result := commands.Run(
		".",
		Parms.LogFile,
		"gh",
		"api",
		"--method",
		"PATCH",
		endpoint,
		"-f",
		"value="+Parms.TargetVersion,
	)
	if result.RC == RC.OK {
		return
	}

	cli.ErrorMessage(*Parms, "No se pudo actualizar %s de %s a %s.", consts.OrganizationVersionVariable, Parms.Organization, Parms.TargetVersion)
	rollbackRemoteTags(Parms, Parms.TargetVersion, Parms.Repos)
	rollbackLocalTags(Parms, Parms.TargetVersion, Parms.Repos)
	cli.Abort(RC.Error, *Parms)
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

func repositoryHighestVersion(Parms *structures.Parms, repository string, current semanticVersion, currentText string) (semanticVersion, string) {
	result := commands.Run(repository, Parms.LogFile, "git", "tag", "--list")
	if result.RC != RC.OK {
		cli.Error(RC.Error, *Parms, "No se pudieron leer los tags de %s", repository)
	}

	return highestSemanticVersion(strings.Fields(result.Stdout), current, currentText)
}

func highestSemanticVersion(tags []string, fallback semanticVersion, fallbackText string) (semanticVersion, string) {
	highest := fallback
	highestText := fallbackText

	for _, tag := range tags {
		version, valid := parseSemanticVersion(tag)
		if !valid {
			continue
		}
		if compareSemanticVersions(version, highest) > 0 {
			highest = version
			highestText = tag
		}
	}

	return highest, highestText
}

func requireTargetVersion(Parms *structures.Parms, operation string) {
	if Parms.TargetVersion == "" {
		cli.Error(RC.Error, *Parms, "%s requiere una versión destino.", operation)
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
