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

type semanticVersion struct {
	major int
	minor int
	patch int
}

// Promote materializes one previously frozen organization version into the
// stable organization. Frozen versions are read from Git tags, never from a
// persistent copied freeze workspace.
func Promote(Parms *structures.Parms) []string {
	if Parms.Push {
		workflowPromotePush(Parms)
		return append([]string{}, Parms.Repos...)
	}

	requireTargetVersion(Parms, "promote")
	if _, ok := parseSemanticVersion(Parms.TargetVersion); !ok {
		cli.Error(RC.Error, *Parms, "La versión a promover no es válida: %s", Parms.TargetVersion)
	}

	root := organizationWorkspaceRoot(Parms)
	if root == "" {
		cli.Error(RC.Error, *Parms, "No se puede deducir el workspace completo de la organización.")
	}
	sourceRepositories := completeOrganizationRepositories(Parms, root)
	if len(sourceRepositories) == 0 {
		cli.Error(RC.Error, *Parms, "No se encontraron repositorios para promover.")
	}

	promotion := *Parms
	promotion.Repos = sourceRepositories
	promotion.Targets = nil
	promotion.TargetDetails = nil
	promotion.BlackList = nil
	promotion.MaterializeDestination = workflowPromoteDestination(&promotion)

	temporary := promoteTemporary(promotion.MaterializeDestination, Parms.TargetVersion)

	fmt.Printf("Organization: %s\n", Parms.Organization)
	fmt.Printf("Frozen version: %s\n", Parms.TargetVersion)
	fmt.Printf("Source: Git tags\n")
	fmt.Printf("Temporary workspace: %s\n", temporary)

	if Parms.DryRun {
		for _, repository := range sourceRepositories {
			name := filepath.Base(repository)
			cli.Preview("Promote source [%s]: origin tag %s -> %s\n", name, Parms.TargetVersion, filepath.Join(temporary, name))
		}
		cli.Preview("Promote tagged version: %s -> %s\n", Parms.TargetVersion, promotion.MaterializeDestination)
		if !Parms.Local {
			cli.Preview("Publish stable organization: %s\n", promotion.MaterializeDestination)
		}
		return append([]string{}, sourceRepositories...)
	}

	validatePromotionTags(Parms, sourceRepositories, Parms.TargetVersion)

	if err := os.RemoveAll(temporary); err != nil {
		cli.Error(RC.Error, *Parms, "No se pudo limpiar el workspace temporal %s: %v", temporary, err)
	}
	if err := os.MkdirAll(temporary, 0755); err != nil {
		cli.Error(RC.Error, *Parms, "No se pudo crear el workspace temporal %s: %v", temporary, err)
	}
	defer os.RemoveAll(temporary)

	promotion.Repos = clonePromotionVersion(Parms, sourceRepositories, temporary, Parms.TargetVersion)

	cli.Step(*Parms, "Materializing stable organization")
	repositories := materializeLocal(&promotion)
	if len(repositories) == 0 {
		Parms.Repos = repositories
		return repositories
	}
	if Parms.Local {
		Parms.Repos = repositories
		cli.Success(*Parms, "Organization promoted locally from tag %s.", Parms.TargetVersion)
		return repositories
	}

	promotion.Repos = repositories
	repositories = push(&promotion)
	Parms.Repos = repositories
	cli.Success(*Parms, "Organization promoted from tag %s.", Parms.TargetVersion)
	return repositories
}

func promoteTemporary(destination string, version string) string {
	parent := filepath.Dir(filepath.Clean(destination))
	name := filepath.Base(filepath.Clean(destination))
	return filepath.Join(parent, "."+name+".promote-"+version+".tmp")
}

func validatePromotionTags(Parms *structures.Parms, repositories []string, version string) {
	for _, repository := range repositories {
		if _, exists := freezeRemoteTagCommit(Parms, repository, version); !exists {
			cli.Error(RC.Error, *Parms, "El tag %s de %s no está publicado en origin.", version, repository)
		}
	}
}

func promotionOrigin(Parms *structures.Parms, repository string) string {
	result := commands.Run(repository, Parms.LogFile, "git", "remote", "get-url", "origin")
	if result.RC != RC.OK || strings.TrimSpace(result.Stdout) == "" {
		cli.Error(RC.Error, *Parms, "No se puede leer origin de %s.", repository)
	}
	return strings.TrimSpace(result.Stdout)
}

func clonePromotionVersion(Parms *structures.Parms, repositories []string, root string, version string) []string {
	clones := make([]string, 0, len(repositories))
	for _, repository := range repositories {
		name := filepath.Base(repository)
		destination := filepath.Join(root, name)
		origin := promotionOrigin(Parms, repository)
		cli.Step(*Parms, "Reading %s at %s", name, version)

		result := commands.RunLogged(root, Parms.LogFile, "git", "clone", "--no-hardlinks", "--branch", version, "--single-branch", origin, destination)
		if result.RC != RC.OK {
			cli.Error(RC.Error, *Parms, "No se pudo leer %s desde el tag publicado %s. Revisa el log: %s", name, version, logName(*Parms))
		}
		clones = append(clones, destination)
	}
	return clones
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
