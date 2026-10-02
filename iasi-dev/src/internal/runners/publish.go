package runners

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"iasi-dev/internal/commands"
	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

func Publish(Context *structures.Context, depth int) []string {
	if Context.Debug {
		fmt.Printf("Publish: targets=%v\n", Context.TargetDetails)
	}
	repositories := []string{}
	processed := false

	for _, target := range selectedTargets(*Context) {
		if target.Repository != "" && isBlackListed(*Context, target.Repository) {
			continue
		}

		rc := dispatchPublish(target, *Context, depth)
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

func dispatchPublish(target structures.Target, Context structures.Context, depth int) int {
	switch strings.ToLower(strings.TrimSpace(Context.Config(target.Path).Type())) {
	case "quarto", "website", "r-package":
		targetMessage(Context, target, depth, "Publish", "Publishing")
		return publishTarget(target.Path, Context)
	default:
		return RC.NothingToDo
	}
}

func publishTarget(target string, Context structures.Context) int {
	parameters := []string{}

	if Context.Format != "" {
		formats := strings.Split(Context.Format, ",")
		for i, format := range formats {
			formats[i] = strconv.Quote(strings.TrimSpace(format))
		}
		parameters = append(parameters, "format = c("+strings.Join(formats, ", ")+")")
	}

	parameters = append(parameters, "path = "+strconv.Quote(filepath.Clean(target)))

	call := "iasi::publish(" + strings.Join(parameters, ", ") + ")"
	expression := "rc = " + call + "; quit(status = as.integer(rc), save = \"no\")"

	result := commands.RunLogged(target, Context.LogFile, "Rscript", "-e", expression)
	return result.RC
}
