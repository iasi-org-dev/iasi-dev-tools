package runners

import (
	"fmt"
	"strings"

	"iasi-dev/internal/commands"
	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

// Release processes discovered IASI targets through the release dispatcher.
func Release(Parms *structures.Parms, depth int) []string {
	if Parms.Debug {
		fmt.Printf("Release: targets=%v\n", Parms.TargetDetails)
	}
	repositories := []string{}
	processed := false

	for _, target := range selectedTargets(*Parms) {
		if target.Repository != "" && isBlackListed(*Parms, target.Repository) {
			continue
		}

		rc := dispatchRelease(target, *Parms, depth)
		if RC.Result(rc) == RC.NothingToDo {
			continue
		}

		processed = true
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

func dispatchRelease(target structures.Target, Parms structures.Parms, depth int) int {
	switch strings.ToLower(strings.TrimSpace(target.Type)) {
	case "r", "quarto", "book", "guide", "website":
		targetMessage(Parms, target, depth, "Release", "Releasing")
		return releaseTarget(target.Path, Parms)
	default:
		return RC.NothingToDo
	}
}

func releaseTarget(target string, Parms structures.Parms) int {
	expression := "rc = iasi::release(); quit(status = as.integer(rc), save = \"no\")"
	result := commands.RunLogged(target, Parms.LogFile, "Rscript", "-e", expression)
	return result.RC
}
