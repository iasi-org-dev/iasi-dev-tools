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

// Build processes each discovered IASI target and delegates its build through
// the type:builder dispatcher.
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

		rc := dispatchBuild(target, *Parms)
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

func dispatchBuild(target structures.Target, Parms structures.Parms) int {
	targetType := strings.ToLower(strings.TrimSpace(target.Type))
	switch targetType {
	case "repository":
		return RC.NothingToDo
	case "software":
		return buildSoftware(target, Parms)
	case "r", "quarto", "book", "guide":
		return buildR(target.Path, Parms)
	default:
		cli.Warning(Parms, "Build no soportado para %s: type=%s", filepath.Base(target.Path), targetType)
		return RC.NothingToDo
	}
}

func buildR(target string, Parms structures.Parms) int {
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
