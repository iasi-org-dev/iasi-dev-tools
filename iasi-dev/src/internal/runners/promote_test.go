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


func TestPromoteTemporaryPathsUseSystemTemp(t *testing.T) {
	destination := filepath.Join("C:", "iasi-org")
	temporary := promoteTemporary(destination, "v0.5.0")
	backup := promoteBackup(destination)

	if filepath.Clean(filepath.Dir(temporary)) != filepath.Clean(os.TempDir()) {
		t.Fatalf("temporary parent = %q, want system temp %q", filepath.Dir(temporary), os.TempDir())
	}
	if filepath.Clean(filepath.Dir(backup)) != filepath.Clean(os.TempDir()) {
		t.Fatalf("backup parent = %q, want system temp %q", filepath.Dir(backup), os.TempDir())
	}
	if filepath.Base(temporary) != ".iasi-org.promote-v0.5.0.tmp" {
		t.Fatalf("temporary name = %q", filepath.Base(temporary))
	}
	if filepath.Base(backup) != ".iasi-org.promote.bak" {
		t.Fatalf("backup name = %q", filepath.Base(backup))
	}
}

func TestFreezePublishesTagsWithoutCreatingWorkspace(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "iasi-org-dev")
	repositories := []string{}
	remotes := map[string]string{}

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
		remotes[name] = remote
	}

	rc := RC.OK
	parms := structures.Parms{
		Organization: "iasi-org-dev",
		Version:      "v0.5.0",
		Repos:        repositories,
		Exclusions:   []string{".git", ".github", "tests"},
		RC:           &rc,
	}

	frozen := Freeze(&parms)
	if len(frozen) != 2 {
		t.Fatalf("Freeze() returned %d repositories, want 2", len(frozen))
	}

	for _, repository := range repositories {
		name := filepath.Base(repository)
		localCommit := strings.TrimSpace(gitTest(t, repository, "rev-parse", "HEAD"))
		localTag := strings.TrimSpace(gitTest(t, repository, "rev-list", "-n", "1", "v0.5.0"))
		remoteTag := strings.TrimSpace(gitTest(t, base, "--git-dir", remotes[name], "rev-list", "-n", "1", "v0.5.0"))
		if localTag != localCommit {
			t.Fatalf("local tag in %s = %s, want %s", name, localTag, localCommit)
		}
		if remoteTag != localCommit {
			t.Fatalf("remote tag in %s = %s, want %s", name, remoteTag, localCommit)
		}
	}

	if _, err := os.Stat(filepath.Join(base, "iasi-org-v0.5.0")); !os.IsNotExist(err) {
		t.Fatalf("freeze created a versioned workspace: %v", err)
	}
}

func TestPromoteMaterializesTaggedVersionNotCurrentHead(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "iasi-org-dev")
	repository := filepath.Join(root, "repo-a")
	remote := filepath.Join(base, "repo-a.git")
	stable := filepath.Join(base, "iasi-org")

	gitTest(t, base, "init", "-b", "main", repository)
	gitTest(t, repository, "config", "user.name", "IASI Test")
	gitTest(t, repository, "config", "user.email", "iasi@example.invalid")
	if err := os.WriteFile(filepath.Join(repository, "README.md"), []byte("frozen\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repository, "config.yml"), []byte("organization: iasi-org-dev\n"), 0644); err != nil {
		t.Fatal(err)
	}
	gitTest(t, repository, "add", ".")
	gitTest(t, repository, "commit", "-m", "frozen")
	gitTest(t, base, "init", "--bare", remote)
	gitTest(t, repository, "remote", "add", "origin", remote)

	rc := RC.OK
	freezeParms := structures.Parms{
		Organization: "iasi-org-dev",
		Version:      "v0.5.0",
		Repos:        []string{repository},
		Exclusions:   []string{".git", ".github", "tests"},
		RC:           &rc,
	}
	Freeze(&freezeParms)

	if err := os.WriteFile(filepath.Join(repository, "README.md"), []byte("development\n"), 0644); err != nil {
		t.Fatal(err)
	}
	gitTest(t, repository, "add", ".")
	gitTest(t, repository, "commit", "-m", "next development")

	if err := os.MkdirAll(stable, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stable, "OLD.txt"), []byte("old destination\n"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_AUTHOR_NAME", "IASI Test")
	t.Setenv("GIT_AUTHOR_EMAIL", "iasi@example.invalid")
	t.Setenv("GIT_COMMITTER_NAME", "IASI Test")
	t.Setenv("GIT_COMMITTER_EMAIL", "iasi@example.invalid")

	rc = RC.OK
	promoteParms := structures.Parms{
		Organization:            "iasi-org-dev",
		SourceOrganization:      "iasi-org-dev",
		DestinationOrganization: "iasi-org",
		SourcePath:              root,
		DestinationPath:         stable,
		TargetVersion:           "v0.5.0",
		Local:                   true,
		Repos:                   []string{repository},
		Exclusions:              []string{".git", ".github", "tests"},
		RC:                      &rc,
	}

	promoted := Promote(&promoteParms)
	if len(promoted) != 1 {
		t.Fatalf("Promote() returned %v, want one repository", promoted)
	}

	data, err := os.ReadFile(filepath.Join(stable, "repo-a", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "frozen\n" {
		t.Fatalf("promoted content = %q, want frozen tagged content", string(data))
	}
	config, err := os.ReadFile(filepath.Join(stable, "repo-a", "config.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if string(config) != "organization: iasi-org\n" {
		t.Fatalf("postprocessed config = %q, want destination organization", string(config))
	}

	commitCount := strings.TrimSpace(gitTest(t, filepath.Join(stable, "repo-a"), "rev-list", "--count", "HEAD"))
	if commitCount != "1" {
		t.Fatalf("promoted history has %s commits, want 1", commitCount)
	}
	remoteURL := strings.TrimSpace(gitTest(t, filepath.Join(stable, "repo-a"), "remote", "get-url", "origin"))
	if remoteURL != "https://github.com/iasi-org/repo-a.git" {
		t.Fatalf("promoted origin = %q, want destination organization", remoteURL)
	}
	if _, err := os.Stat(filepath.Join(stable, "OLD.txt")); !os.IsNotExist(err) {
		t.Fatalf("old destination content survived promotion: %v", err)
	}

	temporary := promoteTemporary(stable, "v0.5.0")
	if _, err := os.Stat(temporary); !os.IsNotExist(err) {
		t.Fatalf("promotion temporary workspace still exists: %v", err)
	}
}

func TestFreezeRequiresVersionGreaterThanEveryExistingTag(t *testing.T) {
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

	gitTest(t, repositories[0], "tag", "v0.6.0")
	gitTest(t, repositories[1], "tag", "v0.7.0")
	gitTest(t, repositories[1], "push", "origin", "v0.7.0")
	gitTest(t, repositories[1], "tag", "-d", "v0.7.0")

	for _, version := range []string{"v0.6.9", "v0.7.0"} {
		t.Run(version, func(t *testing.T) {
			rc := RC.OK
			parms := structures.Parms{
				Organization: "iasi-org-dev",
				Version:      version,
				Repos:        repositories,
				Exclusions:   []string{".git", ".github", "tests"},
				RC:           &rc,
			}

			defer func() {
				recovered := recover()
				if recovered == nil {
					t.Fatalf("Freeze() with %s did not fail", version)
				}
				if _, ok := recovered.(RC.Stop); !ok {
					t.Fatalf("Freeze() with %s panicked with %T, want RC.Stop", version, recovered)
				}
				if output := strings.TrimSpace(gitTest(t, repositories[0], "tag", "--list", version)); output != "" {
					t.Fatalf("Freeze() created %s before rejecting the version", version)
				}
			}()

			Freeze(&parms)
		})
	}
}

func TestPromoteDiscardsExistingTemporaryWorkspace(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "iasi-org-dev")
	repository := filepath.Join(root, "repo-a")
	remote := filepath.Join(base, "repo-a.git")
	stable := filepath.Join(base, "iasi-org")
	temporary := promoteTemporary(stable, "v0.5.0")

	gitTest(t, base, "init", "-b", "main", repository)
	gitTest(t, repository, "config", "user.name", "IASI Test")
	gitTest(t, repository, "config", "user.email", "iasi@example.invalid")
	if err := os.WriteFile(filepath.Join(repository, "README.md"), []byte("frozen\n"), 0644); err != nil {
		t.Fatal(err)
	}
	gitTest(t, repository, "add", ".")
	gitTest(t, repository, "commit", "-m", "frozen")
	gitTest(t, base, "init", "--bare", remote)
	gitTest(t, repository, "remote", "add", "origin", remote)

	rc := RC.OK
	freezeParms := structures.Parms{
		Organization: "iasi-org-dev",
		Version:      "v0.5.0",
		Repos:        []string{repository},
		Exclusions:   []string{".git", ".github", "tests"},
		RC:           &rc,
	}
	Freeze(&freezeParms)

	if err := os.MkdirAll(temporary, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(temporary, "STALE.txt"), []byte("stale\n"), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("GIT_AUTHOR_NAME", "IASI Test")
	t.Setenv("GIT_AUTHOR_EMAIL", "iasi@example.invalid")
	t.Setenv("GIT_COMMITTER_NAME", "IASI Test")
	t.Setenv("GIT_COMMITTER_EMAIL", "iasi@example.invalid")

	rc = RC.OK
	promoteParms := structures.Parms{
		Organization:            "iasi-org-dev",
		SourceOrganization:      "iasi-org-dev",
		DestinationOrganization: "iasi-org",
		SourcePath:              root,
		DestinationPath:         stable,
		TargetVersion:           "v0.5.0",
		Local:                   true,
		Repos:                   []string{repository},
		Exclusions:              []string{".git", ".github", "tests"},
		RC:                      &rc,
	}

	Promote(&promoteParms)

	if _, err := os.Stat(filepath.Join(stable, "STALE.txt")); !os.IsNotExist(err) {
		t.Fatalf("stale temporary workspace content survived promotion: %v", err)
	}
	if _, err := os.Stat(temporary); !os.IsNotExist(err) {
		t.Fatalf("temporary workspace still exists after promotion: %v", err)
	}
}
