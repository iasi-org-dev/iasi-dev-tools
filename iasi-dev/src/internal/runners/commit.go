package runners

import (
	"fmt"
	"path/filepath"
	"strings"

	"iasi-dev/internal/cli"
	"iasi-dev/internal/commands"
	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

func Commit(Parms *structures.Parms, depth int) []string {
	if Parms.Debug {
		fmt.Printf("Commit: repos=%v blackList=%v\n", Parms.Repos, Parms.BlackList)
	}
	targets := []string{}

	for _, repository := range Parms.Repos {
		if isBlackListed(*Parms, repository) {
			continue
		}

		rc := commitRepository(repository, Parms, depth)
		switch rc {
		case RC.OK, RC.NothingToDo:
			targets = append(targets, repository)
		case RC.Skip:
		default:
			cli.Error(RC.Error, *Parms, "Resultado inesperado en commit de %s: rc=0x%02X.", filepath.Base(repository), rc)
		}
	}

	return targets
}

func commitRepository(repository string, Parms *structures.Parms, depth int) int {
	if Parms.Debug {
		fmt.Printf("commitRepository: repository=%s tolerant=%t local=%t\n", repository, Parms.Tolerant, Parms.Local)
	}

	rc := changesPending(repository, *Parms)
	if rc != RC.NothingToDo {
		if handleRC(Parms, rc) == RC.Skip {
			return RC.Skip
		}

		if Parms.Subcommand == "" {
			cli.Header(*Parms, "Commit %s", filepath.Base(repository))
		} else {
			cli.StepAt(*Parms, depth, "Committing %s", filepath.Base(repository))
		}

		rc = addChanges(repository, *Parms)
		if handleRC(Parms, rc) == RC.Skip {
			return RC.Skip
		}

		rc = commitChanges(repository, *Parms)
		if handleRC(Parms, rc) == RC.Skip {
			return RC.Skip
		}
	}

	if !Parms.Local {
		if Parms.Subcommand == "" {
			cli.Header(*Parms, "Push %s", filepath.Base(repository))
		} else {
			cli.StepAt(*Parms, depth, "Pushing %s", filepath.Base(repository))
		}

		rc = pushChanges(repository, *Parms)
		if handleRC(Parms, rc) == RC.Skip {
			return RC.Skip
		}
	}

	return RC.OK
}

func changesPending(repository string, Parms structures.Parms) int {
	result := commands.RunFriendly(repository, Parms.LogFile, "git", "status")
	return result.RC
}

func addChanges(repository string, Parms structures.Parms) int {
	result := commands.RunLogged(repository, Parms.LogFile, "git", "add", "-A", ".")
	if result.RC != RC.OK {
		return RC.Fatal
	}
	return RC.OK
}

func commitChanges(repository string, Parms structures.Parms) int {
	result := commands.RunLogged(repository, Parms.LogFile, "git", "commit", "-m", Parms.Message)
	if result.RC != RC.OK {
		return RC.Fatal
	}
	return RC.OK
}

func pushChanges(repository string, Parms structures.Parms) int {
	args := pushChangesArguments(Parms.Organization)
	result := commands.RunLogged(repository, Parms.LogFile, "git", args...)
	if result.RC != RC.OK {
		return RC.Fatal
	}
	return RC.OK
}

func pushChangesArguments(organization string) []string {
	organization = strings.TrimSpace(strings.ToLower(organization))
	if organization == "" || strings.HasSuffix(organization, "-dev") {
		return []string{"push"}
	}
	return []string{"push", "--force", "origin", "HEAD:main"}
}
