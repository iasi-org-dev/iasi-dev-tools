package runners

import (
	"strings"

	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

func buildSoftware(target structures.Target, Context structures.Context, depth int) int {
	builder := strings.ToLower(strings.TrimSpace(Context.Config(target.Path).SoftwareBuilder()))

	switch builder {
	case "go":
		targetMessage(Context, target, depth, "Build", "Building")
		return buildGo(target, Context)
	case "r":
		targetMessage(Context, target, depth, "Build", "Building")
		return buildR(target.Path, Context)
	default:
		return RC.NothingToDo
	}
}
