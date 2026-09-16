package runners

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"iasi-dev/internal/consts"
	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

func TestWorkflowStopsAfterBuildNothingToDo(t *testing.T) {
	if workflowContinuesAfterBuild(RC.NothingToDo) {
		t.Fatal("workflow must stop after build returns NothingToDo")
	}
}

func TestWorkflowContinuesAfterOtherBuildResults(t *testing.T) {
	for _, rc := range []int{RC.OK, RC.Info, RC.Warning, RC.Attention} {
		if !workflowContinuesAfterBuild(rc) {
			t.Fatalf("workflow unexpectedly stops after build RC 0x%X", rc)
		}
	}
}

func TestWorkflowRunStagePreservesPreviousStateWhenOptionalStageDoesNotApply(t *testing.T) {
	rc := RC.Warning
	parms := structures.Parms{
		All:    true,
		Repos:  []string{"repo-a"},
		LastRC: RC.OK,
		RC:     &rc,
	}

	runner := func(parms *structures.Parms) []string {
		RC.Add(parms.RC, RC.NothingToDo)
		parms.LastRC = RC.NothingToDo
		return nil
	}

	workflowRunStage(&parms, runner, true)

	if len(parms.Repos) != 1 || parms.Repos[0] != "repo-a" {
		t.Fatalf("Repos = %v, want [repo-a]", parms.Repos)
	}
	if rc != RC.Warning {
		t.Fatalf("RC = 0x%X, want 0x%X", rc, RC.Warning)
	}
	if parms.LastRC != RC.OK {
		t.Fatalf("LastRC = 0x%X, want 0x%X", parms.LastRC, RC.OK)
	}
}

func TestWorkflowRunStageKeepsNothingToDoWhenStageIsRequired(t *testing.T) {
	rc := RC.OK
	parms := structures.Parms{
		Repos:  []string{"repo-a"},
		LastRC: RC.OK,
		RC:     &rc,
	}

	runner := func(parms *structures.Parms) []string {
		RC.Add(parms.RC, RC.NothingToDo)
		parms.LastRC = RC.NothingToDo
		return nil
	}

	workflowRunStage(&parms, runner, false)

	if len(parms.Repos) != 0 {
		t.Fatalf("Repos = %v, want empty", parms.Repos)
	}
	if rc != RC.NothingToDo {
		t.Fatalf("RC = 0x%X, want 0x%X", rc, RC.NothingToDo)
	}
	if parms.LastRC != RC.NothingToDo {
		t.Fatalf("LastRC = 0x%X, want 0x%X", parms.LastRC, RC.NothingToDo)
	}
}

func TestWorkflowRepositoryTargetsPreserveProjectOrder(t *testing.T) {
	repoA := filepath.Join("C:", "iasi-org-dev", "repo-a")
	repoB := filepath.Join("C:", "iasi-org-dev", "repo-b")

	parms := structures.Parms{
		TargetDetails: []structures.Target{
			{Path: filepath.Join(repoA, "project-1"), Repository: repoA},
			{Path: filepath.Join(repoB, "project-x"), Repository: repoB},
			{Path: filepath.Join(repoA, "project-2"), Repository: repoA},
		},
	}

	targets := workflowRepositoryTargets(parms, repoA)

	if len(targets) != 2 {
		t.Fatalf("len(targets) = %d, want 2", len(targets))
	}
	if filepath.Base(targets[0].Path) != "project-1" ||
		filepath.Base(targets[1].Path) != "project-2" {
		t.Fatalf("target order = [%s, %s], want [project-1, project-2]",
			filepath.Base(targets[0].Path),
			filepath.Base(targets[1].Path),
		)
	}
}

func TestWorkflowTargetParmsExposeExactlyOneProject(t *testing.T) {
	repository := filepath.Join("C:", "iasi-org-dev", "repo-a")
	first := structures.Target{
		Path:       filepath.Join(repository, "project-1"),
		Type:       "quarto",
		Repository: repository,
	}
	second := structures.Target{
		Path:       filepath.Join(repository, "project-2"),
		Type:       "quarto",
		Repository: repository,
	}

	parms := structures.Parms{
		Repos:         []string{repository},
		Targets:       []string{first.Path, second.Path},
		TargetDetails: []structures.Target{first, second},
		BlackList:     []string{"some-other-repository"},
		LastRC:        RC.Warning,
	}

	targetParms := workflowTargetParms(parms, repository, second)
	selected := selectedTargets(targetParms)

	if len(selected) != 1 {
		t.Fatalf("selected targets = %d, want 1", len(selected))
	}
	if filepath.Clean(selected[0].Path) != filepath.Clean(second.Path) {
		t.Fatalf("selected target = %q, want %q", selected[0].Path, second.Path)
	}
	if len(targetParms.Targets) != 1 ||
		filepath.Clean(targetParms.Targets[0]) != filepath.Clean(second.Path) {
		t.Fatalf("Targets = %v, want [%s]", targetParms.Targets, second.Path)
	}
	if len(targetParms.Repos) != 1 ||
		filepath.Clean(targetParms.Repos[0]) != filepath.Clean(repository) {
		t.Fatalf("Repos = %v, want [%s]", targetParms.Repos, repository)
	}
	if len(targetParms.BlackList) != 0 {
		t.Fatalf("target blacklist = %v, want empty isolated blacklist", targetParms.BlackList)
	}
	if targetParms.LastRC != RC.OK {
		t.Fatalf("LastRC = 0x%X, want 0x%X", targetParms.LastRC, RC.OK)
	}

	// The source parameter set must remain untouched.
	if len(parms.TargetDetails) != 2 || len(parms.Targets) != 2 || len(parms.BlackList) != 1 {
		t.Fatalf("source parameters were modified: %+v", parms)
	}
}

func TestWorkflowOrganizationRoot(t *testing.T) {
	repositories := []string{
		filepath.Join("C:", "iasi-org-dev", "repo-a"),
		filepath.Join("C:", "iasi-org-dev", "repo-b"),
	}

	got := workflowOrganizationRoot(repositories)
	want := filepath.Join("C:", "iasi-org-dev")
	if got != want {
		t.Fatalf("workflowOrganizationRoot() = %q, want %q", got, want)
	}
}

func TestWorkflowPromoteDestination(t *testing.T) {
	rc := RC.OK
	parms := structures.Parms{
		Organization: "iasi-org-dev",
		Repos: []string{
			filepath.Join("C:", "iasi-org-dev", "repo-a"),
			filepath.Join("C:", "iasi-org-dev", "repo-b"),
		},
		RC: &rc,
	}

	got := workflowPromoteDestination(&parms)
	want := filepath.Join("C:", "iasi-org")
	if got != want {
		t.Fatalf("workflowPromoteDestination() = %q, want %q", got, want)
	}
}

func TestWorkflowPromotePushUsesExistingStableOrganization(t *testing.T) {
	base := t.TempDir()
	devRoot := filepath.Join(base, "iasi-org-dev")
	stableRoot := filepath.Join(base, "iasi-org")
	devRepo := filepath.Join(devRoot, "repo-a")
	stableRepo := filepath.Join(stableRoot, "repo-a")
	remote := filepath.Join(base, "repo-a.git")

	if err := os.MkdirAll(devRepo, 0755); err != nil {
		t.Fatal(err)
	}
	gitTest(t, base, "init", "-b", "main", stableRepo)
	gitTest(t, stableRepo, "config", "user.name", "IASI Test")
	gitTest(t, stableRepo, "config", "user.email", "iasi@example.invalid")
	if err := os.WriteFile(filepath.Join(stableRepo, "README.md"), []byte("stable\n"), 0644); err != nil {
		t.Fatal(err)
	}
	gitTest(t, stableRepo, "add", ".")
	gitTest(t, stableRepo, "commit", "-m", "stable")
	gitTest(t, base, "init", "--bare", remote)
	gitTest(t, stableRepo, "remote", "add", "origin", remote)

	rc := RC.OK
	parms := structures.Parms{
		Organization: "iasi-org-dev",
		Push:         true,
		Repos:        []string{devRepo},
		Exclusions:   append([]string{}, consts.RequiredExclusions...),
		RC:           &rc,
	}

	workflowPromotePush(&parms)

	if len(parms.Repos) != 1 || filepath.Clean(parms.Repos[0]) != filepath.Clean(stableRepo) {
		t.Fatalf("Repos = %v, want [%s]", parms.Repos, stableRepo)
	}
	if got := strings.TrimSpace(gitTest(t, base, "--git-dir", remote, "rev-parse", "refs/heads/main")); got == "" {
		t.Fatal("remote main was not pushed")
	}
}
