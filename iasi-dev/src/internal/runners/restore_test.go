package runners

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

func TestRestoreTargetDefaultsToMain(t *testing.T) {
	Parms := structures.Parms{}
	if got := restoreTarget(&Parms); got != "main" {
		t.Fatalf("restoreTarget() = %q, want main", got)
	}
}

func TestRestoreTargetUsesRequestedVersion(t *testing.T) {
	Parms := structures.Parms{TargetVersion: "v0.5.0"}
	if got := restoreTarget(&Parms); got != "v0.5.0" {
		t.Fatalf("restoreTarget() = %q, want v0.5.0", got)
	}
}

func TestRestoreSwitchArgumentsForMain(t *testing.T) {
	want := []string{"switch", "main"}
	if got := restoreSwitchArguments("main"); !reflect.DeepEqual(got, want) {
		t.Fatalf("restoreSwitchArguments(main) = %v, want %v", got, want)
	}
}

func TestRestoreSwitchArgumentsForTag(t *testing.T) {
	want := []string{"switch", "--detach", "v0.5.0"}
	if got := restoreSwitchArguments("v0.5.0"); !reflect.DeepEqual(got, want) {
		t.Fatalf("restoreSwitchArguments(tag) = %v, want %v", got, want)
	}
}

func TestRestoreRollbackArgumentsForBranch(t *testing.T) {
	state := restoreState{branch: "main"}
	want := []string{"switch", "main"}
	if got := restoreRollbackArguments(state); !reflect.DeepEqual(got, want) {
		t.Fatalf("restoreRollbackArguments(branch) = %v, want %v", got, want)
	}
}

func TestRestoreRollbackArgumentsForDetachedHead(t *testing.T) {
	state := restoreState{commit: "abc123"}
	want := []string{"switch", "--detach", "abc123"}
	if got := restoreRollbackArguments(state); !reflect.DeepEqual(got, want) {
		t.Fatalf("restoreRollbackArguments(detached) = %v, want %v", got, want)
	}
}

func TestRestoreRepositoriesSwitchesToTaggedVersion(t *testing.T) {
	base := t.TempDir()
	repository := filepath.Join(base, "repo-a")
	gitTest(t, base, "init", "-b", "main", repository)
	gitTest(t, repository, "config", "user.name", "IASI Test")
	gitTest(t, repository, "config", "user.email", "iasi@example.invalid")

	if err := os.WriteFile(filepath.Join(repository, "value.txt"), []byte("one\n"), 0644); err != nil {
		t.Fatal(err)
	}
	gitTest(t, repository, "add", ".")
	gitTest(t, repository, "commit", "-m", "one")
	gitTest(t, repository, "tag", "-a", "v0.1.0", "-m", "v0.1.0")
	tagCommit := strings.TrimSpace(gitTest(t, repository, "rev-list", "-n", "1", "v0.1.0"))

	if err := os.WriteFile(filepath.Join(repository, "value.txt"), []byte("two\n"), 0644); err != nil {
		t.Fatal(err)
	}
	gitTest(t, repository, "add", ".")
	gitTest(t, repository, "commit", "-m", "two")

	rc := RC.OK
	parms := structures.Parms{RC: &rc}
	state := repositoryRestoreState(&parms, repository)
	restoreRepositories(&parms, "v0.1.0", []restoreState{state})

	if got := strings.TrimSpace(gitTest(t, repository, "rev-parse", "HEAD")); got != tagCommit {
		t.Fatalf("HEAD = %s, want tagged commit %s", got, tagCommit)
	}
	if got := strings.TrimSpace(gitTest(t, repository, "branch", "--show-current")); got != "" {
		t.Fatalf("branch = %q, want detached HEAD", got)
	}
}
