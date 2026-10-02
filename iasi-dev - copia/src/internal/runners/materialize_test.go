package runners

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

func TestMaterializeOrganizationUsesDestinationDirectoryName(t *testing.T) {
	destination := filepath.Join("work", "iasi-org")
	if got := materializeOrganization(destination); got != "iasi-org" {
		t.Fatalf("materializeOrganization() = %q, want iasi-org", got)
	}
}

func TestMaterializeTemporaryIsSiblingOfDestination(t *testing.T) {
	destination := filepath.Join("work", "iasi-org")
	want := filepath.Join("work", ".iasi-org.materialize.tmp")
	if got := materializeTemporary(destination); got != want {
		t.Fatalf("materializeTemporary() = %q, want %q", got, want)
	}
}

func TestMaterializeConfirmed(t *testing.T) {
	accepted := []string{"s", "S", "si", "sí", "y", "YES"}
	for _, value := range accepted {
		if !materializeConfirmed(value) {
			t.Fatalf("materializeConfirmed(%q) = false, want true", value)
		}
	}

	rejected := []string{"", "n", "no", "anything"}
	for _, value := range rejected {
		if materializeConfirmed(value) {
			t.Fatalf("materializeConfirmed(%q) = true, want false", value)
		}
	}
}

func TestMaterializeLocalCreatesFreshRepositories(t *testing.T) {
	base := t.TempDir()
	sourceRoot := filepath.Join(base, "iasi-org-dev")
	source := filepath.Join(sourceRoot, "repo-a")
	destination := filepath.Join(base, "iasi-org")

	if err := os.MkdirAll(destination, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(destination, "stale.txt"), []byte("remove me"), 0644); err != nil {
		t.Fatal(err)
	}

	gitTest(t, base, "init", "-b", "main", source)
	gitTest(t, source, "config", "user.name", "IASI Test")
	gitTest(t, source, "config", "user.email", "iasi@example.invalid")
	if err := os.MkdirAll(filepath.Join(source, ".github", "workflows"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "one.txt"), []byte("one\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, ".github", "workflows", "test.yml"), []byte("name: test\n"), 0644); err != nil {
		t.Fatal(err)
	}
	gitTest(t, source, "add", ".")
	gitTest(t, source, "commit", "-m", "first")
	if err := os.WriteFile(filepath.Join(source, "two.txt"), []byte("two\n"), 0644); err != nil {
		t.Fatal(err)
	}
	gitTest(t, source, "add", ".")
	gitTest(t, source, "commit", "-m", "second")

	t.Setenv("GIT_AUTHOR_NAME", "IASI Test")
	t.Setenv("GIT_AUTHOR_EMAIL", "iasi@example.invalid")
	t.Setenv("GIT_COMMITTER_NAME", "IASI Test")
	t.Setenv("GIT_COMMITTER_EMAIL", "iasi@example.invalid")

	rc := RC.OK
	parms := structures.Parms{
		Verbose:                0,
		Organization:           "iasi-org-dev",
		Message:                "materialize",
		MaterializeDestination: destination,
		Repos:                  []string{source},
		RC:                     &rc,
	}

	repositories := materializeLocal(&parms)
	if len(repositories) != 1 {
		t.Fatalf("materializeLocal returned %v, want one repository", repositories)
	}

	materialized := filepath.Join(destination, "repo-a")
	if _, err := os.Stat(filepath.Join(destination, "stale.txt")); !os.IsNotExist(err) {
		t.Fatalf("stale destination content still exists: %v", err)
	}
	for _, file := range []string{"one.txt", "two.txt", filepath.Join(".github", "workflows", "test.yml")} {
		if _, err := os.Stat(filepath.Join(materialized, file)); err != nil {
			t.Fatalf("materialized file %s missing: %v", file, err)
		}
	}
	if _, err := os.Stat(filepath.Join(materialized, ".git")); err != nil {
		t.Fatalf("fresh .git missing: %v", err)
	}

	if got := strings.TrimSpace(gitTest(t, materialized, "rev-list", "--count", "HEAD")); got != "1" {
		t.Fatalf("materialized history has %s commits, want 1", got)
	}
	if got := strings.TrimSpace(gitTest(t, materialized, "branch", "--show-current")); got != "main" {
		t.Fatalf("materialized branch = %q, want main", got)
	}
	if got := strings.TrimSpace(gitTest(t, materialized, "remote", "get-url", "origin")); got != "https://github.com/iasi-org/repo-a.git" {
		t.Fatalf("origin = %q", got)
	}
	if _, err := os.Stat(materializeTemporary(destination)); !os.IsNotExist(err) {
		t.Fatalf("temporary workspace still exists: %v", err)
	}
	if _, err := os.Stat(materializeBackup(destination)); !os.IsNotExist(err) {
		t.Fatalf("backup workspace still exists: %v", err)
	}
}
