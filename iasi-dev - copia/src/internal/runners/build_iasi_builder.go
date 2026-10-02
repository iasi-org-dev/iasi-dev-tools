package runners

import (
	"iasi-dev/internal/commands"
	"iasi-dev/internal/structures"
)

func buildIASIBuilder(target structures.Target, Parms structures.Parms) int {
	result := commands.RunLogged(target.Path, Parms.LogFile, "iasi-builder")
	return result.RC
}
