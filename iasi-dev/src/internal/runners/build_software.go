package runners

import (
	"path/filepath"
	"strings"

	"iasi-dev/internal/cli"
	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

// buildSoftware dispatches software targets to the builder declared by the target.
func buildSoftware(target structures.Target, Parms structures.Parms) int {
	builder := strings.ToLower(strings.TrimSpace(target.Builder))

	switch builder {
	case "go":
		return buildGo(target, Parms)
	case "r":
		return buildR(target.Path, Parms)
	case "iasi-script":
		return buildIASIScript(target, Parms)
	default:
		cli.Warning(Parms, "Builder de software no soportado para %s: %s", filepath.Base(target.Path), builder)
		return RC.NothingToDo
	}
}
