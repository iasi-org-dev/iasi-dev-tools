package runners

import (
	"path/filepath"
	"strings"

	"iasi-dev/internal/cli"
	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

func Workflow(Parms *structures.Parms) {
	if Parms.Subcommand == "promote" {
		workflowPromote(Parms)
		return
	}

	repositories := append([]string{}, Parms.Repos...)

	for _, repository := range repositories {
		Parms.Repos = []string{repository}
		cli.Header(*Parms, "%s %s", workflowName(Parms.Subcommand), filepath.Base(repository))
		workflowRepository(repository, Parms, 0)
	}
}

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

	if repositoryActive && !Parms.Checkpoints {
		Parms.Repos = Commit(Parms, depth+1)
	}
}

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

func workflowTargetParms(Parms structures.Parms, repository string, target structures.Target) structures.Parms {
	targetParms := Parms
	targetParms.Repos = []string{repository}
	targetParms.Targets = []string{target.Path}
	targetParms.TargetDetails = []structures.Target{target}
	targetParms.BlackList = nil
	targetParms.LastRC = RC.OK
	return targetParms
}

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

func workflowPromote(Parms *structures.Parms) {
	frozenVersion := strings.TrimSpace(Parms.Version)
	nextVersion := strings.TrimSpace(Parms.NextVersion)

	frozen, frozenOK := parseSemanticVersion(frozenVersion)
	next, nextOK := parseSemanticVersion(nextVersion)
	if !frozenOK {
		cli.Error(RC.Error, *Parms, "La versión actual no es válida: %s", frozenVersion)
	}
	if !nextOK {
		cli.Error(RC.Error, *Parms, "La nueva versión no es válida: %s", nextVersion)
	}
	if compareSemanticVersions(next, frozen) <= 0 {
		cli.Error(RC.Error, *Parms, "La nueva versión %s debe ser superior a la versión actual %s.", nextVersion, frozenVersion)
	}

	Freeze(Parms)

	versionParms := *Parms
	versionParms.TargetVersion = nextVersion
	Version(&versionParms)
	Parms.Version = versionParms.Version

	promoteParms := *Parms
	promoteParms.TargetVersion = frozenVersion
	Parms.Repos = Promote(&promoteParms)
}

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

func workflowContinuesAfterBuild(rc int) bool {
	return RC.Result(rc) != RC.NothingToDo
}

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
