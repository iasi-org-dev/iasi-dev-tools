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

// Build processes discovered IASI targets in iasi-dev. For now only targets
// declaring type: r are delegated to the existing R implementation.
func Build(Parms *structures.Parms) []string {
	if Parms.Debug {
		fmt.Printf("Build: targets=%v\n", Parms.TargetDetails)
	}
	repositories := []string{}
	processed := false

	for _, target := range selectedTargets(*Parms) {
		if target.Repository != "" && isBlackListed(*Parms, target.Repository) {
			continue
		}
		if Parms.Subcommand == "" {
			cli.Header(*Parms, "%sBuild %s [%s]", targetIndent(target), filepath.Base(target.Path), target.Type)
		}

		if !targetUsesR(target) {
			continue
		}
		processed = true
		cli.Step(*Parms, "%sBuilding", targetIndent(target))

		rc := buildTarget(target.Path, *Parms)
		Parms.LastRC = rc
		handled := handleRC(Parms, rc)
		if RC.Has(handled, RC.Skip) {
			if target.Repository != "" {
				addToBlackList(Parms, target.Repository)
			}
			continue
		}
		repositories = appendRepository(repositories, target.Repository)
	}

	if !processed {
		Parms.LastRC = RC.NothingToDo
		RC.Add(Parms.RC, RC.NothingToDo)
	}
	return repositories
}

func buildTarget(target string, Parms structures.Parms) int {
	parameters := []string{}

	if Parms.Format != "" {
		formats := strings.Split(Parms.Format, ",")
		for i, format := range formats {
			formats[i] = strconv.Quote(strings.TrimSpace(format))
		}
		parameters = append(parameters, "format = c("+strings.Join(formats, ", ")+")")
	}

	call := "iasi::build(" + strings.Join(parameters, ", ") + ")"
	expression := "rc = " + call + "; quit(status = as.integer(rc), save = \"no\")"

	result := commands.RunLogged(target, Parms.LogFile, "Rscript", "-e", expression)
	return result.RC
}
