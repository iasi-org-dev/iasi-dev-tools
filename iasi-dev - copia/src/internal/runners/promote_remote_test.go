package runners

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestPromoteRemoteReconciliationCreatesPushesAndDeletes(t *testing.T) {
	root := filepath.Join("work", "iasi-org")
	repositories := []string{
		filepath.Join(root, "repo-a"),
		filepath.Join(root, "repo-new"),
		filepath.Join(root, "iasi-org.github.io"),
	}
	remote := []string{"repo-a", "repo-old", "IASI-ORG.GITHUB.IO"}

	plan := promoteRemoteReconciliation(repositories, remote)

	if !reflect.DeepEqual(plan.Create, []string{"repo-new"}) {
		t.Fatalf("Create = %v, want [repo-new]", plan.Create)
	}
	if !reflect.DeepEqual(plan.Push, repositories) {
		t.Fatalf("Push = %v, want %v", plan.Push, repositories)
	}
	if !reflect.DeepEqual(plan.Delete, []string{"repo-old"}) {
		t.Fatalf("Delete = %v, want [repo-old]", plan.Delete)
	}
}

func TestPromoteRemoteReconciliationCreatesEveryRepositoryWhenDestinationIsEmpty(t *testing.T) {
	root := filepath.Join("work", "iasi-org")
	repositories := []string{
		filepath.Join(root, "repo-b"),
		filepath.Join(root, "repo-a"),
	}

	plan := promoteRemoteReconciliation(repositories, nil)

	if !reflect.DeepEqual(plan.Create, []string{"repo-a", "repo-b"}) {
		t.Fatalf("Create = %v, want every local repository", plan.Create)
	}
	if len(plan.Delete) != 0 {
		t.Fatalf("Delete = %v, want none", plan.Delete)
	}
}

func TestPromoteRemoteSourceRepositoryNameMapsGitHubPagesBackToDevelopment(t *testing.T) {
	if got := promoteRemoteSourceRepositoryName("iasi-org.github.io", "iasi-org-dev", "iasi-org"); got != "iasi-org-dev.github.io" {
		t.Fatalf("source repository = %q, want iasi-org-dev.github.io", got)
	}
	if got := promoteRemoteSourceRepositoryName("repo-a", "iasi-org-dev", "iasi-org"); got != "repo-a" {
		t.Fatalf("normal source repository = %q, want repo-a", got)
	}
}

func TestPromoteRemoteLocalModeCreatesMissingWithoutPushOrDelete(t *testing.T) {
	root := filepath.Join("work", "iasi-org")
	repositories := []string{
		filepath.Join(root, "repo-a"),
		filepath.Join(root, "repo-new"),
	}
	remote := []string{"repo-a", "repo-old"}

	plan := promoteRemotePlanForMode(repositories, remote, true)

	if !reflect.DeepEqual(plan.Create, []string{"repo-new"}) {
		t.Fatalf("Create = %v, want [repo-new]", plan.Create)
	}
	if len(plan.Push) != 0 {
		t.Fatalf("Push = %v, want none in local mode", plan.Push)
	}
	if len(plan.Delete) != 0 {
		t.Fatalf("Delete = %v, want none in local mode", plan.Delete)
	}
}

func TestPromoteRemoteNonLocalModeFullyReconciles(t *testing.T) {
	root := filepath.Join("work", "iasi-org")
	repositories := []string{
		filepath.Join(root, "repo-a"),
		filepath.Join(root, "repo-new"),
	}
	remote := []string{"repo-a", "repo-old"}

	plan := promoteRemotePlanForMode(repositories, remote, false)

	if !reflect.DeepEqual(plan.Create, []string{"repo-new"}) {
		t.Fatalf("Create = %v, want [repo-new]", plan.Create)
	}
	if !reflect.DeepEqual(plan.Push, repositories) {
		t.Fatalf("Push = %v, want %v", plan.Push, repositories)
	}
	if !reflect.DeepEqual(plan.Delete, []string{"repo-old"}) {
		t.Fatalf("Delete = %v, want [repo-old]", plan.Delete)
	}
}
