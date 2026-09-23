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

func TestFreezeDestinationUsesCurrentVersion(t *testing.T) {
	root := filepath.Join("C:", "iasi-org-dev")
	got := freezeDestination(root, "iasi-org-dev", "v0.5.0")
	want := filepath.Join("C:", "iasi-org-v0.5.0")
	if got != want {
		t.Fatalf("freezeDestination() = %q, want %q", got, want)
	}
}

func TestWriteAndReadFreezeManifest(t *testing.T) {
	root := t.TempDir()
	manifest := freezeManifest{
		Organization: "iasi-org-dev",
		Version:      "v0.5.0",
		Repositories: map[string]string{"repo-a": "abc", "repo-b": "def"},
	}
	if err := writeFreezeManifest(root, manifest); err != nil {
		t.Fatal(err)
	}
	got, err := readFreezeManifest(root)
	if err != nil {
		t.Fatal(err)
	}
	if got.Organization != manifest.Organization || got.Version != manifest.Version || len(got.Repositories) != 2 {
		t.Fatalf("manifest = %+v, want %+v", got, manifest)
	}
}

func TestValidateFrozenOrganizationUsesManifestCommits(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "iasi-org-v0.5.0")
	repository := filepath.Join(root, "repo-a")
	gitTest(t, base, "init", "-b", "main", repository)
	gitTest(t, repository, "config", "user.name", "IASI Test")
	gitTest(t, repository, "config", "user.email", "iasi@example.invalid")
	if err := os.WriteFile(filepath.Join(repository, "README.md"), []byte("frozen\n"), 0644); err != nil {
		t.Fatal(err)
	}
	gitTest(t, repository, "add", ".")
	gitTest(t, repository, "commit", "-m", "frozen")
	commit := repositoryHead(&structures.Parms{RC: intPtr(RC.OK)}, repository)

	manifest := freezeManifest{
		Organization: "iasi-org-dev",
		Version:      "v0.5.0",
		Repositories: map[string]string{"repo-a": commit},
	}
	if err := writeFreezeManifest(root, manifest); err != nil {
		t.Fatal(err)
	}

	rc := RC.OK
	parms := structures.Parms{
		Organization: "iasi-org-dev",
		TargetVersion: "v0.5.0",
		Exclusions: []string{".git", ".github", "tests"},
		RC: &rc,
	}
	repositories := validateFrozenOrganization(&parms, root, "v0.5.0")
	if len(repositories) != 1 || filepath.Clean(repositories[0]) != filepath.Clean(repository) {
		t.Fatalf("repositories = %v, want [%s]", repositories, repository)
	}
}

func intPtr(value int) *int {
	return &value
}

func TestFreezeCreatesCompleteVersionedWorkspace(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "iasi-org-dev")
	repositories := []string{}

	for _, name := range []string{"repo-a", "repo-b"} {
		repository := filepath.Join(root, name)
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
	}

	rc := RC.OK
	parms := structures.Parms{
		Organization: "iasi-org-dev",
		Version:      "v0.5.0",
		Local:        true,
		Repos:        repositories,
		Exclusions:   []string{".git", ".github", "tests"},
		RC:           &rc,
	}

	frozen := Freeze(&parms)
	freezeRoot := filepath.Join(base, "iasi-org-v0.5.0")
	if len(frozen) != 2 {
		t.Fatalf("Freeze() returned %d repositories, want 2", len(frozen))
	}
	if _, err := os.Stat(filepath.Join(freezeRoot, freezeManifestName)); err != nil {
		t.Fatalf("freeze manifest missing: %v", err)
	}
	for _, name := range []string{"repo-a", "repo-b"} {
		frozenRepo := filepath.Join(freezeRoot, name)
		if _, err := os.Stat(filepath.Join(frozenRepo, ".git")); err != nil {
			t.Fatalf("frozen git history missing for %s: %v", name, err)
		}
		if got := strings.TrimSpace(gitTest(t, frozenRepo, "tag", "--list", "v0.5.0")); got != "v0.5.0" {
			t.Fatalf("frozen tag missing in %s", name)
		}
	}
}
