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

// Publish processes discovered IASI targets through the publish dispatcher.
func Publish(Parms *structures.Parms, depth int) []string {
	if Parms.Debug {
		fmt.Printf("Publish: targets=%v\n", Parms.TargetDetails)
	}
	repositories := []string{}
	processed := false

	for _, target := range selectedTargets(*Parms) {
		if target.Repository != "" && isBlackListed(*Parms, target.Repository) {
			continue
		}

		rc := dispatchPublish(target, *Parms, depth)
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

func dispatchPublish(target structures.Target, Parms structures.Parms, depth int) int {
	switch strings.ToLower(strings.TrimSpace(target.Type)) {
	case "r", "quarto", "book", "guide", "website":
		targetMessage(Parms, target, depth, "Publish", "Publishing")
		return publishTarget(target.Path, Parms)
	default:
		return RC.NothingToDo
	}
}

func publishTarget(target string, Parms structures.Parms) int {
	parameters := []string{}

	if Parms.Format != "" {
		formats := strings.Split(Parms.Format, ",")
		for i, format := range formats {
			formats[i] = strconv.Quote(strings.TrimSpace(format))
		}
		parameters = append(parameters, "format = c("+strings.Join(formats, ", ")+")")
	}

	parameters = append(parameters, "path = "+strconv.Quote(filepath.Clean(target)))

	call := "iasi::publish(" + strings.Join(parameters, ", ") + ")"
	expression := "rc = " + call + "; quit(status = as.integer(rc), save = \"no\")"

	result := commands.RunLogged(target, Parms.LogFile, "Rscript", "-e", expression)
	return result.RC
}
