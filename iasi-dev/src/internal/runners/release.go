package runners

import (
	"fmt"
	"strings"

	"iasi-dev/internal/commands"
	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

func Release(Context *structures.Context, depth int) []string {
	if Context.Debug {
		fmt.Printf("Release: targets=%v\n", Context.TargetDetails)
	}
	repositories := []string{}
	processed := false

	for _, target := range selectedTargets(*Context) {
		if target.Repository != "" && isBlackListed(*Context, target.Repository) {
			continue
		}

		rc := dispatchRelease(target, *Context, depth)
		if RC.Result(rc) == RC.NothingToDo {
			continue
		}

		processed = true
		Context.LastRC = rc
		if RC.Has(handleRC(Context, rc), RC.Skip) {
			if target.Repository != "" {
				addToBlackList(Context, target.Repository)
			}
			continue
		}
		repositories = appendRepository(repositories, target.Repository)
	}

	if !processed {
		Context.LastRC = RC.NothingToDo
		RC.Add(Context.RC, RC.NothingToDo)
	}
	return repositories
}

func dispatchRelease(target structures.Target, Context structures.Context, depth int) int {
	switch strings.ToLower(strings.TrimSpace(Context.Config(target.Path).Type())) {
	case "quarto", "website", "r-package":
		targetMessage(Context, target, depth, "Release", "Releasing")
		return releaseTarget(target.Path, Context)
	default:
		return RC.NothingToDo
	}
}

func releaseTarget(target string, Context structures.Context) int {
	expression := "rc = iasi::release(); quit(status = as.integer(rc), save = \"no\")"
	result := commands.RunLogged(target, Context.LogFile, "Rscript", "-e", expression)
	return result.RC
}
