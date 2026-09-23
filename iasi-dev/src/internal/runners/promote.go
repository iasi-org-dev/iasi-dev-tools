package runners

import (
	"fmt"
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

// Promote promotes one previously frozen organization version to the stable
// organization. It never reads release content from the live development tree.
func Promote(Parms *structures.Parms) []string {
	if Parms.Push {
		workflowPromotePush(Parms)
		return append([]string{}, Parms.Repos...)
	}

	requireTargetVersion(Parms, "promote")
	if _, ok := parseSemanticVersion(Parms.TargetVersion); !ok {
		cli.Error(RC.Error, *Parms, "La versión a promover no es válida: %s", Parms.TargetVersion)
	}

	frozenRoot := frozenOrganizationPath(Parms, Parms.TargetVersion)
	frozenRepositories := validateFrozenOrganization(Parms, frozenRoot, Parms.TargetVersion)

	fmt.Printf("Organization: %s\n", Parms.Organization)
	fmt.Printf("Frozen version: %s\n", Parms.TargetVersion)
	fmt.Printf("Frozen workspace: %s\n", frozenRoot)

	promotion := *Parms
	promotion.Repos = frozenRepositories
	promotion.Targets = nil
	promotion.TargetDetails = nil
	promotion.BlackList = nil
	promotion.MaterializeDestination = workflowPromoteDestination(&promotion)

	if Parms.DryRun {
		cli.Preview("Promote frozen workspace: %s -> %s\n", frozenRoot, promotion.MaterializeDestination)
		if !Parms.Local {
			cli.Preview("Publish stable organization: %s\n", promotion.MaterializeDestination)
		}
		return append([]string{}, frozenRepositories...)
	}

	cli.Step(*Parms, "Materializing stable organization")
	repositories := materializeLocal(&promotion)
	if len(repositories) == 0 {
		Parms.Repos = repositories
		return repositories
	}
	if Parms.Local {
		Parms.Repos = repositories
		cli.Success(*Parms, "Organization promoted locally from frozen version %s.", Parms.TargetVersion)
		return repositories
	}

	promotion.Repos = repositories
	repositories = push(&promotion)
	Parms.Repos = repositories
	cli.Success(*Parms, "Organization promoted from frozen version %s.", Parms.TargetVersion)
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
