package runners

import (
	"fmt"
	"path/filepath"

	"iasi-dev/internal/cli"
	"iasi-dev/internal/commands"
	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

// Commit commits the selected repositories and returns the targets that remain active.
func Commit(Parms *structures.Parms) []string {
	if Parms.Debug {
		fmt.Printf("Commit: repos=%v blackList=%v\n", Parms.Repos, Parms.BlackList)
	}
	targets := []string{}

	for _, repository := range Parms.Repos {
		if isBlackListed(*Parms, repository) {
			continue
		}

		if Parms.Subcommand == "" {
			cli.Header(*Parms, "Commit %s", filepath.Base(repository))
		}
		cli.Step(*Parms, "Committing")

		rc := commitRepository(repository, Parms)
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

func commitRepository(repository string, Parms *structures.Parms) int {
	if Parms.Debug {
		fmt.Printf("commitRepository: repository=%s tolerant=%t local=%t\n", repository, Parms.Tolerant, Parms.Local)
	}

	rc := changesPending(repository, *Parms)
	if rc == RC.NothingToDo {
		return rc
	}
	if handleRC(Parms, rc) == RC.Skip {
		return RC.Skip
	}

	rc = addChanges(repository, *Parms)
	if handleRC(Parms, rc) == RC.Skip {
		return RC.Skip
	}

	rc = commitChanges(repository, *Parms)
	if handleRC(Parms, rc) == RC.Skip {
		return RC.Skip
	}

	if !Parms.Local {
		rc = pushChanges(repository, *Parms)
		if handleRC(Parms, rc) == RC.Skip {
			return RC.Skip
		}
	}

	return RC.OK
}

func changesPending(repository string, Parms structures.Parms) int {
	if Parms.Debug {
		fmt.Printf("changesPending: repository=%s\n", repository)
	}
	result := commands.RunFriendly(repository, Parms.LogFile, "git", "status")
	return result.RC
}

func addChanges(repository string, Parms structures.Parms) int {
	if Parms.Debug {
		fmt.Printf("addChanges: repository=%s\n", repository)
	}
	result := commands.RunLogged(repository, Parms.LogFile, "git", "add", "-A", ".")
	if result.RC != RC.OK {
		return RC.Fatal
	}
	return RC.OK
}

func commitChanges(repository string, Parms structures.Parms) int {
	if Parms.Debug {
		fmt.Printf("commitChanges: repository=%s message=%q\n", repository, Parms.Message)
	}
	result := commands.RunLogged(repository, Parms.LogFile, "git", "commit", "-m", Parms.Message)
	if result.RC != RC.OK {
		return RC.Fatal
	}
	return RC.OK
}

func pushChanges(repository string, Parms structures.Parms) int {
	if Parms.Debug {
		fmt.Printf("pushChanges: repository=%s\n", repository)
	}
	result := commands.RunLogged(repository, Parms.LogFile, "git", "push")
	if result.RC != RC.OK {
		return RC.Fatal
	}
	return RC.OK
}
