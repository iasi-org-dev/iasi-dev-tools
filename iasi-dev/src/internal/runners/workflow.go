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

		switch Parms.Subcommand {
		case "build":
			workflowBuild(true, Parms)
		case "publish":
			workflowPublish(true, Parms)
		case "release":
			workflowRelease(true, Parms)
		default:
			cli.Error(RC.Error, *Parms, "Workflow desconocido: %q", Parms.Subcommand)
		}
	}
}

// workflowBuild builds and commits when standalone or used as a checkpoint.
func workflowBuild(standalone bool, Parms *structures.Parms) {
	Parms.LastRC = RC.OK
	Parms.Repos = Build(Parms)
	if !workflowContinuesAfterBuild(Parms.LastRC) {
		Parms.Repos = nil
		return
	}
	if standalone || Parms.Checkpoints {
		Parms.Repos = Commit(Parms)
	}
}

// workflowPublish optionally builds first, publishes and commits when required.
func workflowPublish(standalone bool, Parms *structures.Parms) {
	if Parms.All {
		workflowBuild(false, Parms)
	}
	if len(Parms.Repos) == 0 {
		return
	}

	Parms.Repos = Publish(Parms)
	if standalone || Parms.Checkpoints {
		Parms.Repos = Commit(Parms)
	}
}

// workflowRelease optionally runs previous stages, releases and commits when required.
func workflowRelease(standalone bool, Parms *structures.Parms) {
	if Parms.All {
		workflowPublish(false, Parms)
	}
	if len(Parms.Repos) == 0 {
		return
	}

	Parms.Repos = Release(Parms)
	if standalone || Parms.Checkpoints {
		Parms.Repos = Commit(Parms)
	}
}

// workflowPromote promotes the complete development organization, materializes
// the resulting state locally as the stable organization and, unless -l is active, pushes it.
func workflowPromote(standalone bool, Parms *structures.Parms) {
	if Parms.Push {
		workflowPromotePush(Parms)
		return
	}

	Parms.Repos = Promote(Parms)
	if len(Parms.Repos) == 0 {
		return
	}

	destination := workflowPromoteDestination(Parms)
	cli.Step(*Parms, "Materializing stable organization")
	Parms.MaterializeDestination = destination
	Parms.Repos = materializeLocal(Parms)
	if len(Parms.Repos) == 0 || Parms.Local {
		return
	}

	Parms.Repos = push(Parms)
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
