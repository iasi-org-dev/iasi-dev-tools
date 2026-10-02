package runners

import (
	"strings"

	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

// buildSoftware dispatches software targets to the builder declared by the target.
func buildSoftware(target structures.Target, Parms structures.Parms, depth int) int {
	builder := strings.ToLower(strings.TrimSpace(target.Builder))

	switch builder {
	case "go":
		targetMessage(Parms, target, depth, "Build", "Building")
		return buildGo(target, Parms)
	case "r":
		targetMessage(Parms, target, depth, "Build", "Building")
		return buildR(target.Path, Parms)
	case "iasi-script":
		targetMessage(Parms, target, depth, "Build", "Building")
		return buildIASIScript(target, Parms)
	default:
		return RC.NothingToDo
	}
}
