package runners

import (
	"iasi-dev/internal/commands"
	"iasi-dev/internal/structures"
)

func buildIASIScript(target structures.Target, Parms structures.Parms) int {
	result := commands.RunLogged(target.Path, Parms.LogFile, "iasi-script")
	return result.RC
}
