package runners

import (
	"fmt"
	"path/filepath"

	"iasi-dev/internal/cli"
	"iasi-dev/internal/commands"
	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

// Release processes discovered IASI targets in iasi-dev. For now only targets
// declaring type: r are delegated to the existing R implementation.
func Release(Parms *structures.Parms) []string {
	if Parms.Debug {
		fmt.Printf("Release: targets=%v\n", Parms.TargetDetails)
	}
	repositories := []string{}
	processed := false

	for _, target := range selectedTargets(*Parms) {
		if target.Repository != "" && isBlackListed(*Parms, target.Repository) {
			continue
		}
		if Parms.Subcommand == "" {
			cli.Header(*Parms, "%sRelease %s [%s]", targetIndent(target), filepath.Base(target.Path), target.Type)
		}

		if !targetUsesR(target) {
			continue
		}
		processed = true
		cli.Step(*Parms, "%sReleasing", targetIndent(target))

		rc := releaseTarget(target.Path, *Parms)
		Parms.LastRC = rc
		if RC.Has(handleRC(Parms, rc), RC.Skip) {
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

func releaseTarget(target string, Parms structures.Parms) int {
	expression := "rc = iasi::release(); quit(status = as.integer(rc), save = \"no\")"
	result := commands.RunLogged(target, Parms.LogFile, "Rscript", "-e", expression)
	return result.RC
}
