package runners

import (
	"fmt"
	"strconv"
	"strings"

	"iasi-dev/internal/commands"
	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

func Build(Context *structures.Context, depth int) []string {
	if Context.Debug {
		fmt.Printf("Build: targets=%v\n", Context.TargetDetails)
	}
	repositories := []string{}
	processed := false

	for _, target := range selectedTargets(*Context) {
		if target.Repository != "" && isBlackListed(*Context, target.Repository) {
			continue
		}

		rc := dispatchBuild(target, *Context, depth)
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

func dispatchBuild(target structures.Target, Context structures.Context, depth int) int {
	targetType := strings.ToLower(strings.TrimSpace(Context.Config(target.Path).Type()))

	switch targetType {
	case "software":
		return buildSoftware(target, Context, depth)
	case "quarto", "website", "r-package":
		targetMessage(Context, target, depth, "Build", "Building")
		return buildR(target.Path, Context)
	default:
		return RC.NothingToDo
	}
}

func buildR(target string, Context structures.Context) int {
	parameters := []string{}

	if Context.Format != "" {
		formats := strings.Split(Context.Format, ",")
		for i, format := range formats {
			formats[i] = strconv.Quote(strings.TrimSpace(format))
		}
		parameters = append(parameters, "format = c("+strings.Join(formats, ", ")+")")
	}

	call := "iasi::build(" + strings.Join(parameters, ", ") + ")"
	expression := "rc = " + call + "; quit(status = as.integer(rc), save = \"no\")"

	result := commands.RunLogged(target, Context.LogFile, "Rscript", "-e", expression)
	return result.RC
}
