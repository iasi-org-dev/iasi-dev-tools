package runners

import (
	"fmt"
	"path/filepath"
	"strings"

	"iasi-dev/internal/cli"
	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

// handleRC accumulates one result and applies the common error policy.
func handleRC(Parms *structures.Parms, rc int) int {
	if Parms.Debug {
		fmt.Printf("handleRC: tolerant=%t rc=%d (0x%02X)\n", Parms.Tolerant, rc, rc)
	}

	if !RC.IsErroneous(rc) {
		RC.Add(Parms.RC, rc)
		return rc
	}

	if Parms.Tolerant {
		RC.Add(Parms.RC, rc)
		return RC.Skip
	}

	cli.ErrorMessage(
		*Parms,
		"La operación ha fallado con RC %d (0x%02X). Revisa el log: %s",
		rc,
		rc,
		logName(*Parms),
	)
	cli.Abort(rc, *Parms)
	return RC.Skip
}

// logName returns the current log path when available.
func logName(Parms structures.Parms) string {
	if Parms.LogFile == nil {
		return "(sin log)"
	}
	return Parms.LogFile.Name()
}

// addToBlackList blacklists repository.
func addToBlackList(Parms *structures.Parms, repository string) {
	if repository == "" || isBlackListed(*Parms, repository) {
		return
	}
	Parms.BlackList = append(Parms.BlackList, filepath.Clean(repository))
}

// isBlackListed reports whether repository has been invalidated by a previous operation.
func isBlackListed(Parms structures.Parms, repository string) bool {
	for _, blocked := range Parms.BlackList {
		if filepath.Clean(blocked) == filepath.Clean(repository) {
			return true
		}
	}
	return false
}

// selectedTargets returns discovered IASI targets belonging to the currently selected repositories.
func selectedTargets(Parms structures.Parms) []structures.Target {
	selected := make([]structures.Target, 0, len(Parms.TargetDetails))
	for _, target := range Parms.TargetDetails {
		if target.Repository == "" || repositorySelected(Parms.Repos, target.Repository) {
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

func targetIndent(target structures.Target) string {
	return strings.Repeat("  ", target.Depth)
}

func targetUsesR(target structures.Target) bool {
	return strings.EqualFold(strings.TrimSpace(target.Type), "r")
}

func appendRepository(repositories []string, repository string) []string {
	if repository == "" || repositorySelected(repositories, repository) {
		return repositories
	}
	return append(repositories, filepath.Clean(repository))
}
