package runners

import (
	"iasi-dev/internal/cli"
	"iasi-dev/internal/commands"
	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

func push(Parms *structures.Parms) []string {
	if len(Parms.Repos) == 0 {
		cli.Error(RC.Error, *Parms, "No se encontraron repositorios para publicar.")
	}

	cli.Step(*Parms, "Pushing organization")
	for _, repository := range Parms.Repos {
		if !pushRepository(Parms, repository) {
			cli.Error(RC.Error, *Parms, "No se pudo publicar %s. Revisa el log: %s", repository, logName(*Parms))
		}
	}

	cli.Success(*Parms, "Organization pushed.")
	return append([]string{}, Parms.Repos...)
}

func pushRepository(Parms *structures.Parms, repository string) bool {
	result := commands.RunLogged(repository, Parms.LogFile, "git", "push", "-u", "--force", "origin", "HEAD:main")
	return result.RC == RC.OK
}
