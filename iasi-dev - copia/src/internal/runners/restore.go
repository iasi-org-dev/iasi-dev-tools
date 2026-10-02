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

type restoreState struct {
	repository string
	branch     string
	commit     string
}

// Restore restores every repository to a tagged organization version.
// Without a target version it restores the repositories to main.
func Restore(Parms *structures.Parms) []string {
	target := restoreTarget(Parms)

	if len(Parms.Repos) == 0 {
		cli.Error(RC.Error, *Parms, "No se encontraron repositorios para restaurar.")
	}

	if target != "main" {
		if _, ok := parseSemanticVersion(target); !ok {
			cli.Error(RC.Error, *Parms, "La versión destino no es válida: %s", target)
		}
	}

	for _, repository := range Parms.Repos {
		validateRestoreTarget(Parms, repository, target)
	}

	states := make([]restoreState, 0, len(Parms.Repos))
	for _, repository := range Parms.Repos {
		states = append(states, repositoryRestoreState(Parms, repository))
	}

	fmt.Printf("Organization: %s\n", Parms.Organization)
	fmt.Printf("Current version: %s\n", Parms.Version)
	fmt.Printf("Restore target: %s\n", target)

	restoreRepositories(Parms, target, states)

	return append([]string{}, Parms.Repos...)
}

func restoreTarget(Parms *structures.Parms) string {
	if Parms.TargetVersion == "" {
		return "main"
	}
	return Parms.TargetVersion
}

func validateRestoreTarget(Parms *structures.Parms, repository string, target string) {
	ref := "refs/heads/main"
	if target != "main" {
		ref = "refs/tags/" + target
	}

	result := commands.Run(repository, Parms.LogFile, "git", "show-ref", "--verify", "--quiet", ref)
	if result.RC != RC.OK {
		cli.Error(RC.Error, *Parms, "No existe %s en %s", target, repository)
	}
}

func repositoryRestoreState(Parms *structures.Parms, repository string) restoreState {
	result := commands.Run(repository, Parms.LogFile, "git", "symbolic-ref", "--quiet", "--short", "HEAD")
	if result.RC == RC.OK {
		branch := strings.TrimSpace(result.Stdout)
		if branch != "" {
			return restoreState{repository: repository, branch: branch}
		}
	}

	result = commands.Run(repository, Parms.LogFile, "git", "rev-parse", "HEAD")
	if result.RC != RC.OK {
		cli.Error(RC.Error, *Parms, "No se pudo determinar el estado actual de %s", repository)
	}

	return restoreState{repository: repository, commit: strings.TrimSpace(result.Stdout)}
}

// restoreRepositories executes the restore transaction.
func restoreRepositories(Parms *structures.Parms, target string, states []restoreState) {
	restored := []restoreState{}

	for _, state := range states {
		args := restoreSwitchArguments(target)
		if !restoreCommand(Parms, state.repository, args...) {
			cli.ErrorMessage(*Parms, "No se pudo restaurar %s a %s", state.repository, target)
			rollbackRestore(Parms, restored)
			cli.Abort(RC.Error, *Parms)
		}
		restored = append(restored, state)
	}
}

func restoreSwitchArguments(target string) []string {
	if target == "main" {
		return []string{"switch", "main"}
	}
	return []string{"switch", "--detach", target}
}

func rollbackRestore(Parms *structures.Parms, states []restoreState) {
	failed := false

	for i := len(states) - 1; i >= 0; i-- {
		state := states[i]
		args := restoreRollbackArguments(state)
		if restoreCommand(Parms, state.repository, args...) {
			continue
		}

		failed = true
		cli.ErrorMessage(*Parms, "No se pudo restaurar el estado previo de %s", state.repository)
	}

	if failed {
		cli.ErrorMessage(*Parms, "El rollback de restore no se completó; el sistema puede haber quedado en un estado inconsistente.")
	}
}

func restoreRollbackArguments(state restoreState) []string {
	if state.branch != "" {
		return []string{"switch", state.branch}
	}
	return []string{"switch", "--detach", state.commit}
}

func restoreCommand(Parms *structures.Parms, repository string, args ...string) bool {
	cli.VeryVerbose(*Parms, "%s: git %s", filepath.Base(repository), formatCommandArguments(args))
	result := commands.RunLogged(repository, Parms.LogFile, "git", args...)
	return result.RC == RC.OK
}
