package runners

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

func TestPushRepositoryReplacesRemoteMain(t *testing.T) {
	base := t.TempDir()
	repository := filepath.Join(base, "repo-a")
	remote := filepath.Join(base, "repo-a.git")

	gitTest(t, base, "init", "-b", "main", repository)
	gitTest(t, repository, "config", "user.name", "IASI Test")
	gitTest(t, repository, "config", "user.email", "iasi@example.invalid")
	if err := os.WriteFile(filepath.Join(repository, "README.md"), []byte("materialized\n"), 0644); err != nil {
		t.Fatal(err)
	}
	gitTest(t, repository, "add", ".")
	gitTest(t, repository, "commit", "-m", "materialized")
	localHead := strings.TrimSpace(gitTest(t, repository, "rev-parse", "HEAD"))

	gitTest(t, base, "init", "--bare", remote)
	gitTest(t, repository, "remote", "add", "origin", remote)

	rc := RC.OK
	parms := structures.Parms{RC: &rc}
	if !pushRepository(&parms, repository) {
		t.Fatal("pushRepository() = false, want true")
	}

	remoteHead := strings.TrimSpace(gitTest(t, base, "--git-dir", remote, "rev-parse", "refs/heads/main"))
	if remoteHead != localHead {
		t.Fatalf("remote main = %s, want %s", remoteHead, localHead)
	}
}

func gitTest(t *testing.T, directory string, args ...string) string {
	t.Helper()
	command := exec.Command("git", args...)
	command.Dir = directory
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, output)
	}
	return string(output)
}
