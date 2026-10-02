package runners

import (
	"fmt"
	"path/filepath"

	"iasi-dev/internal/cli"
	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

func handleRC(Context *structures.Context, rc int) int {
	if Context.Debug {
		fmt.Printf("handleRC: tolerant=%t rc=%d (0x%02X)\n", Context.Tolerant, rc, rc)
	}

	if !RC.IsErroneous(rc) {
		RC.Add(Context.RC, rc)
		return rc
	}

	if Context.Tolerant {
		RC.Add(Context.RC, rc)
		return RC.Skip
	}

	cli.ErrorMessage(
		*Context,
		"La operación ha fallado con RC %d (0x%02X). Revisa el log: %s",
		rc,
		rc,
		logName(*Context),
	)
	cli.Abort(rc, *Context)
	return RC.Skip
}

func logName(Context structures.Context) string {
	if Context.LogFile == nil {
		return "(sin log)"
	}
	return Context.LogFile.Name()
}

func addToBlackList(Context *structures.Context, repository string) {
	if repository == "" || isBlackListed(*Context, repository) {
		return
	}
	Context.BlackList = append(Context.BlackList, filepath.Clean(repository))
}

func isBlackListed(Context structures.Context, repository string) bool {
	for _, blocked := range Context.BlackList {
		if filepath.Clean(blocked) == filepath.Clean(repository) {
			return true
		}
	}
	return false
}

func selectedTargets(Context structures.Context) []structures.Target {
	selected := make([]structures.Target, 0, len(Context.TargetDetails))
	for _, target := range Context.TargetDetails {
		if target.Repository == "" || repositorySelected(Context.Repos, target.Repository) {
			selected = append(selected, target)
		}
	}
	return selected
}

func repositorySelected(repositories []string, repository string) bool {
	for _, selected := range repositories {
		if filepath.Clean(selected) == filepath.Clean(repository) {
			return true
		}
	}
	return false
}

func targetMessageDepth(base int, target structures.Target) int {
	if base < 0 {
		base = 0
	}
	depth := target.Depth
	if depth < 1 {
		depth = 1
	}
	return base + depth
}

func targetLabel(Context structures.Context, target structures.Target) string {
	return fmt.Sprintf("%s [%s]", filepath.Base(target.Path), Context.Config(target.Path).Type())
}

func targetMessage(Context structures.Context, target structures.Target, depth int, headerVerb string, stepVerb string) {
	label := targetLabel(Context, target)
	if Context.Subcommand == "" {
		cli.Header(Context, "%s %s", headerVerb, label)
		return
	}
	cli.StepAt(Context, targetMessageDepth(depth, target), "%s %s", stepVerb, label)
}

func appendRepository(repositories []string, repository string) []string {
	if repository == "" || repositorySelected(repositories, repository) {
		return repositories
	}
	return append(repositories, filepath.Clean(repository))
}
