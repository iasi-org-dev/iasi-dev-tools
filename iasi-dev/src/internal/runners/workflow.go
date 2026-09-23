package runners

import (
	"os"
	"path/filepath"
	"strings"

	"iasi-dev/internal/args"
	"iasi-dev/internal/cli"
	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

// Workflow executes repository workflows one repository at a time.
// Within each repository, targets execute their complete workflow one at a time.
// Promote is organization-wide and therefore runs once with the complete repository set.
func Workflow(Parms *structures.Parms) {
	if Parms.Subcommand == "promote" {
		if Parms.Push {
			cli.Header(*Parms, "Pushing %s", strings.TrimSuffix(Parms.Organization, "-dev"))
		} else {
			cli.Header(*Parms, "%s %s", workflowName(Parms.Subcommand), Parms.Organization)
		}
		workflowPromote(true, Parms)
		return
	}

	repositories := append([]string{}, Parms.Repos...)

	for _, repository := range repositories {
		Parms.Repos = []string{repository}
		cli.Header(*Parms, "%s %s", workflowName(Parms.Subcommand), filepath.Base(repository))
		workflowRepository(repository, Parms, 0)
	}
}

// workflowRepository executes the requested workflow target by target.
//
// This deliberately restores the original workflow semantics: one project
// completes its lifecycle before the next project starts. A tolerated failure
// invalidates the repository for the final commit/push, but does not prevent
// sibling targets in the same repository from being processed.
func workflowRepository(repository string, Parms *structures.Parms, depth int) {
	if isBlackListed(*Parms, repository) {
		return
	}

	targets := workflowRepositoryTargets(*Parms, repository)
	repositoryActive := false
	repositoryFailed := false

	for _, target := range targets {
		rcBefore := RC.Value(Parms.RC)

		targetParms := workflowTargetParms(*Parms, repository, target)

		// Once one target has failed, keep processing sibling targets in tolerant
		// mode but do not create later checkpoints for a repository that is
		// already known to be invalid.
		if repositoryFailed {
			targetParms.Checkpoints = false
		}

		switch targetParms.Subcommand {
		case "build":
			workflowBuild(false, &targetParms, depth)
		case "publish":
			workflowPublish(false, &targetParms, depth)
		case "release":
			workflowRelease(false, &targetParms, depth)
		default:
			cli.Error(RC.Error, targetParms, "Workflow desconocido: %q", targetParms.Subcommand)
		}

		// A target for which the workflow does not apply is transparent. In
		// particular, repository/none targets must not contaminate the global
		// result merely because their build returns NothingToDo.
		if RC.Result(targetParms.LastRC) == RC.NothingToDo &&
			len(targetParms.Repos) == 0 &&
			!isBlackListed(targetParms, repository) {
			if Parms.RC != nil {
				*Parms.RC = rcBefore
			}
			continue
		}

		Parms.LastRC = targetParms.LastRC

		if isBlackListed(targetParms, repository) {
			repositoryFailed = true
			continue
		}

		if len(targetParms.Repos) != 0 {
			repositoryActive = true
		}
	}

	Parms.Repos = []string{repository}

	if repositoryFailed {
		addToBlackList(Parms, repository)
	}

	// Without checkpoints, commit/push once after every valid target in the
	// repository has completed. With checkpoints, the stage workflows already
	// performed the requested commits.
	if repositoryActive && !Parms.Checkpoints {
		Parms.Repos = Commit(Parms, depth+1)
	}
}

// workflowRepositoryTargets returns the discovered targets that belong to one
// repository, preserving discovery order.
func workflowRepositoryTargets(Parms structures.Parms, repository string) []structures.Target {
	targets := []structures.Target{}

	for _, target := range Parms.TargetDetails {
		if target.Repository != "" &&
			filepath.Clean(target.Repository) == filepath.Clean(repository) {
			targets = append(targets, target)
		}
	}

	return targets
}

// workflowTargetParms creates the single-target view used by workflows.
// The cumulative RC and log handle remain shared; selection and blacklist are
// isolated so one tolerated target failure cannot hide its sibling targets.
func workflowTargetParms(Parms structures.Parms, repository string, target structures.Target) structures.Parms {
	targetParms := Parms
	targetParms.Repos = []string{repository}
	targetParms.Targets = []string{target.Path}
	targetParms.TargetDetails = []structures.Target{target}
	targetParms.BlackList = nil
	targetParms.LastRC = RC.OK
	return targetParms
}

// workflowBuild builds and commits when standalone or used as a checkpoint.
func workflowBuild(standalone bool, Parms *structures.Parms, depth int) {
	Parms.LastRC = RC.OK
	Parms.Repos = Build(Parms, depth)
	if !workflowContinuesAfterBuild(Parms.LastRC) {
		Parms.Repos = nil
		return
	}
	if standalone || Parms.Checkpoints {
		Parms.Repos = Commit(Parms, depth+1)
	}
}

// workflowPublish optionally builds first, publishes and commits when required.
func workflowPublish(standalone bool, Parms *structures.Parms, depth int) {
	if Parms.All {
		workflowBuild(false, Parms, depth)
	}
	if len(Parms.Repos) == 0 {
		return
	}

	workflowRunStage(Parms, func(parms *structures.Parms) []string {
		return Publish(parms, depth)
	}, Parms.All)
	if len(Parms.Repos) == 0 {
		return
	}

	if standalone || Parms.Checkpoints {
		Parms.Repos = Commit(Parms, depth+1)
	}
}

// workflowRelease optionally runs previous stages, releases and commits when required.
func workflowRelease(standalone bool, Parms *structures.Parms, depth int) {
	if Parms.All {
		workflowPublish(false, Parms, depth)
	}
	if len(Parms.Repos) == 0 {
		return
	}

	workflowRunStage(Parms, func(parms *structures.Parms) []string {
		return Release(parms, depth)
	}, Parms.All)
	if len(Parms.Repos) == 0 {
		return
	}

	if standalone || Parms.Checkpoints {
		Parms.Repos = Commit(Parms, depth+1)
	}
}

// workflowRunStage executes one workflow stage. When -a is active, a stage that
// does not apply to the current repository is transparent: the repository and
// accumulated result from the preceding stage are preserved.
func workflowRunStage(Parms *structures.Parms, runner func(*structures.Parms) []string, optional bool) {
	repositories := append([]string{}, Parms.Repos...)
	rc := RC.Value(Parms.RC)
	lastRC := Parms.LastRC

	Parms.Repos = runner(Parms)

	if optional && RC.Result(Parms.LastRC) == RC.NothingToDo {
		Parms.Repos = repositories
		if Parms.RC != nil {
			*Parms.RC = rc
		}
		Parms.LastRC = lastRC
	}
}

// workflowPromote delegates the organization-wide promotion transaction.
// Promotion always consumes a previously frozen version.
func workflowPromote(standalone bool, Parms *structures.Parms) {
	Parms.Repos = Promote(Parms)
}

// workflowPromotePush publishes only the already materialized stable organization.
// It deliberately skips promotion and materialization.
func workflowPromotePush(Parms *structures.Parms) {
	destination := workflowPromoteDestination(Parms)
	info, err := os.Stat(destination)
	if err != nil || !info.IsDir() {
		cli.Error(RC.Error, *Parms, "No existe la organización estable local para publicar: %s", destination)
	}

	pushParms := *Parms
	pushParms.Targets = []string{destination}
	pushParms.Repos = nil
	pushParms.BlackList = nil
	args.Prepare(&pushParms)
	if len(pushParms.Repos) == 0 {
		cli.Error(RC.Error, *Parms, "No se encontraron repositorios en la organización estable local: %s", destination)
	}

	Parms.Repos = pushParms.Repos
	Parms.Repos = push(Parms)
}

// workflowPromoteDestination returns the sibling stable organization workspace.
// Example: C:\iasi-org-dev -> C:\iasi-org.
func workflowPromoteDestination(Parms *structures.Parms) string {
	stableOrganization := strings.TrimSuffix(Parms.Organization, "-dev")
	if stableOrganization == Parms.Organization || stableOrganization == "" {
		cli.Error(RC.Error, *Parms, "No se puede deducir la organización estable desde %q.", Parms.Organization)
	}

	root := workflowOrganizationRoot(Parms.Repos)
	if root == "" {
		cli.Error(RC.Error, *Parms, "No se puede deducir el workspace de la organización.")
	}

	return filepath.Join(filepath.Dir(root), stableOrganization)
}

// workflowOrganizationRoot finds the common parent containing the organization's repositories.
func workflowOrganizationRoot(repositories []string) string {
	if len(repositories) == 0 {
		return ""
	}

	root := filepath.Dir(filepath.Clean(repositories[0]))
	for _, repository := range repositories[1:] {
		directory := filepath.Dir(filepath.Clean(repository))
		for !workflowPathContains(root, directory) {
			parent := filepath.Dir(root)
			if parent == root {
				return ""
			}
			root = parent
		}
	}
	return root
}

func workflowPathContains(parent string, child string) bool {
	relative, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	return relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)))
}

// workflowContinuesAfterBuild reports whether later workflow stages apply.
// NothingToDo means build succeeded but found no buildable IASI project.
func workflowContinuesAfterBuild(rc int) bool {
	return RC.Result(rc) != RC.NothingToDo
}

// workflowName returns the display name of a workflow.
func workflowName(name string) string {
	switch name {
	case "build":
		return "Building"
	case "publish":
		return "Publishing"
	case "release":
		return "Releasing"
	case "promote":
		return "Promoting"
	default:
		return name
	}
}
