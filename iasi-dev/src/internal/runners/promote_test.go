package runners

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

func TestParseSemanticVersion(t *testing.T) {
	tests := []struct {
		value string
		ok    bool
	}{
		{"v0.5.0", true},
		{"v1.0.12", true},
		{"v10.20.30", true},
		{"0.5.0", false},
		{"v0.5", false},
		{"v0.5.0.1", false},
		{"v01.5.0", false},
		{"v0.05.0", false},
		{"v0.5.00", false},
		{"v0.x.0", false},
	}

	for _, test := range tests {
		_, ok := parseSemanticVersion(test.value)
		if ok != test.ok {
			t.Fatalf("parseSemanticVersion(%q) ok = %t, want %t", test.value, ok, test.ok)
		}
	}
}

func TestCompareSemanticVersions(t *testing.T) {
	v030, _ := parseSemanticVersion("v0.3.0")
	v060, _ := parseSemanticVersion("v0.6.0")
	v070, _ := parseSemanticVersion("v0.7.0")
	v100, _ := parseSemanticVersion("v1.0.0")

	if compareSemanticVersions(v030, v060) >= 0 {
		t.Fatal("v0.3.0 must be lower than v0.6.0")
	}
	if compareSemanticVersions(v060, v060) != 0 {
		t.Fatal("v0.6.0 must equal v0.6.0")
	}
	if compareSemanticVersions(v070, v060) <= 0 {
		t.Fatal("v0.7.0 must be greater than v0.6.0")
	}
	if compareSemanticVersions(v100, v070) <= 0 {
		t.Fatal("v1.0.0 must be greater than v0.7.0")
	}
}

func TestHighestSemanticVersion(t *testing.T) {
	fallback, _ := parseSemanticVersion("v0.3.0")
	highest, text := highestSemanticVersion([]string{"other", "v0.1.0", "v0.6.0", "v0.4.0"}, fallback, "v0.3.0")
	want, _ := parseSemanticVersion("v0.6.0")

	if compareSemanticVersions(highest, want) != 0 {
		t.Fatal("highest semantic version must be v0.6.0")
	}
	if text != "v0.6.0" {
		t.Fatalf("highest text = %q, want v0.6.0", text)
	}
}

func TestHighestSemanticVersionFallsBackToCurrentVersion(t *testing.T) {
	fallback, _ := parseSemanticVersion("v0.5.0")
	highest, text := highestSemanticVersion([]string{"not-a-version", "release"}, fallback, "v0.5.0")

	if compareSemanticVersions(highest, fallback) != 0 {
		t.Fatal("highest semantic version must fall back to current version")
	}
	if text != "v0.5.0" {
		t.Fatalf("highest text = %q, want v0.5.0", text)
	}
}

func TestPromoteTagsCreatesLocalAndRemoteTags(t *testing.T) {
	base := t.TempDir()
	repositories := []string{}
	remotes := []string{}

	for _, name := range []string{"repo-a", "repo-b"} {
		repository := filepath.Join(base, name)
		remote := filepath.Join(base, name+".git")
		gitTest(t, base, "init", "-b", "main", repository)
		gitTest(t, repository, "config", "user.name", "IASI Test")
		gitTest(t, repository, "config", "user.email", "iasi@example.invalid")
		if err := os.WriteFile(filepath.Join(repository, "README.md"), []byte(name+"\n"), 0644); err != nil {
			t.Fatal(err)
		}
		gitTest(t, repository, "add", ".")
		gitTest(t, repository, "commit", "-m", "initial")
		gitTest(t, base, "init", "--bare", remote)
		gitTest(t, repository, "remote", "add", "origin", remote)
		repositories = append(repositories, repository)
		remotes = append(remotes, remote)
	}

	rc := RC.OK
	parms := structures.Parms{TargetVersion: "v0.7.0", Repos: repositories, RC: &rc}
	promoteTags(&parms)

	for i, repository := range repositories {
		if got := strings.TrimSpace(gitTest(t, repository, "tag", "--list", "v0.7.0")); got != "v0.7.0" {
			t.Fatalf("local tag missing in %s", repository)
		}
		if got := strings.TrimSpace(gitTest(t, base, "--git-dir", remotes[i], "tag", "--list", "v0.7.0")); got != "v0.7.0" {
			t.Fatalf("remote tag missing in %s", remotes[i])
		}
	}
}

func TestPromotePushPublishesStableSibling(t *testing.T) {
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
		Exclusions:   []string{".git", ".github", "tests"},
		RC:           &rc,
	}

	repositories := Promote(&parms)
	if len(repositories) != 1 || filepath.Clean(repositories[0]) != filepath.Clean(stableRepo) {
		t.Fatalf("Promote(-p) = %v, want [%s]", repositories, stableRepo)
	}
	if got := strings.TrimSpace(gitTest(t, base, "--git-dir", remote, "rev-parse", "refs/heads/main")); got == "" {
		t.Fatal("remote main was not pushed")
	}
}
