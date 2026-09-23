package runners

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

func TestPromotePostprocessContentKeepsNarrativeDevReferences(t *testing.T) {
	input := strings.Join([]string{
		"La organización de desarrollo es iasi-org-dev.",
		`iasi-dev build --path C:\iasi-org-dev`,
		"https://iasi-org-dev.github.io/iasi-home/",
		"https://github.com/iasi-org-dev/iasi-home",
		`remotes::install_github("iasi-org-dev/iasi-r")`,
	}, "\n")

	got := string(promotePostprocessContent([]byte(input), ".qmd", "iasi-org-dev", "iasi-org"))

	if !strings.Contains(got, "La organización de desarrollo es iasi-org-dev.") {
		t.Fatalf("narrative development reference was changed:\n%s", got)
	}
	if !strings.Contains(got, `C:\iasi-org-dev`) {
		t.Fatalf("explicit Windows source path was changed:\n%s", got)
	}
	for _, want := range []string{
		"https://iasi-org.github.io/iasi-home/",
		"https://github.com/iasi-org/iasi-home",
		`remotes::install_github("iasi-org/iasi-r")`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("postprocessed content does not contain %q:\n%s", want, got)
		}
	}
}

func TestPromotePostprocessContentReplacesRawOrganizationInConfiguration(t *testing.T) {
	input := []byte("organization = \"iasi-org-dev\"\n")
	got := string(promotePostprocessContent(input, ".toml", "iasi-org-dev", "iasi-org"))
	if got != "organization = \"iasi-org\"\n" {
		t.Fatalf("configuration postprocess = %q, want destination organization", got)
	}
}

func TestPromotePostprocessFilesLeavesScriptsUntouched(t *testing.T) {
	root := t.TempDir()
	repository := filepath.Join(root, "repo-a")
	gitTest(t, root, "init", "-b", "main", repository)
	gitTest(t, repository, "config", "user.name", "IASI Test")
	gitTest(t, repository, "config", "user.email", "iasi@example.invalid")

	files := map[string]string{
		"config.yml": "organization: iasi-org-dev\n",
		"README.md": "Repository: https://github.com/iasi-org-dev/repo-a\nDevelopment: iasi-org-dev\n",
		"script.sh": "TARGET_ORG=iasi-org-dev\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(repository, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	gitTest(t, repository, "add", ".")
	gitTest(t, repository, "commit", "-m", "initial")

	rc := RC.OK
	parms := structures.Parms{
		SourceOrganization:      "iasi-org-dev",
		DestinationOrganization: "iasi-org",
		RC:                      &rc,
	}

	processed := promotePostprocess(&parms, root, []string{repository})
	if len(processed) != 1 || processed[0] != repository {
		t.Fatalf("processed repositories = %v, want %s", processed, repository)
	}

	config, err := os.ReadFile(filepath.Join(repository, "config.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if string(config) != "organization: iasi-org\n" {
		t.Fatalf("config.yml = %q, want destination organization", string(config))
	}

	readme, err := os.ReadFile(filepath.Join(repository, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(readme), "https://github.com/iasi-org/repo-a") {
		t.Fatalf("README repository URL was not postprocessed: %s", readme)
	}
	if !strings.Contains(string(readme), "Development: iasi-org-dev") {
		t.Fatalf("README narrative development reference was changed: %s", readme)
	}

	script, err := os.ReadFile(filepath.Join(repository, "script.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if string(script) != "TARGET_ORG=iasi-org-dev\n" {
		t.Fatalf("script.sh was modified: %q", string(script))
	}

	if status := strings.TrimSpace(gitTest(t, repository, "status", "--porcelain")); status != "" {
		t.Fatalf("postprocess left a dirty repository: %s", status)
	}
	if count := strings.TrimSpace(gitTest(t, repository, "rev-list", "--count", "HEAD")); count != "1" {
		t.Fatalf("postprocess created %s commits, want one amended snapshot", count)
	}
}

func TestPromotePostprocessRenamesGitHubPagesRepository(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "iasi-org-dev.github.io")
	gitTest(t, root, "init", "-b", "main", source)
	gitTest(t, source, "config", "user.name", "IASI Test")
	gitTest(t, source, "config", "user.email", "iasi@example.invalid")
	if err := os.WriteFile(filepath.Join(source, "index.html"), []byte("https://iasi-org-dev.github.io/iasi-home/\n"), 0644); err != nil {
		t.Fatal(err)
	}
	gitTest(t, source, "add", ".")
	gitTest(t, source, "commit", "-m", "initial")
	gitTest(t, source, "remote", "add", "origin", "https://github.com/iasi-org/iasi-org-dev.github.io.git")

	rc := RC.OK
	parms := structures.Parms{
		SourceOrganization:      "iasi-org-dev",
		DestinationOrganization: "iasi-org",
		RC:                      &rc,
	}

	processed := promotePostprocess(&parms, root, []string{source})
	destination := filepath.Join(root, "iasi-org.github.io")
	if len(processed) != 1 || processed[0] != destination {
		t.Fatalf("processed repositories = %v, want %s", processed, destination)
	}
	if _, err := os.Stat(source); !os.IsNotExist(err) {
		t.Fatalf("source GitHub Pages directory still exists: %v", err)
	}
	if info, err := os.Stat(destination); err != nil || !info.IsDir() {
		t.Fatalf("destination GitHub Pages directory missing: %v", err)
	}

	remote := strings.TrimSpace(gitTest(t, destination, "remote", "get-url", "origin"))
	if remote != "https://github.com/iasi-org/iasi-org.github.io.git" {
		t.Fatalf("GitHub Pages origin = %q, want destination repository", remote)
	}
	data, err := os.ReadFile(filepath.Join(destination, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "https://iasi-org.github.io/iasi-home/\n" {
		t.Fatalf("GitHub Pages redirect = %q, want destination URL", string(data))
	}
	if status := strings.TrimSpace(gitTest(t, destination, "status", "--porcelain")); status != "" {
		t.Fatalf("postprocessed GitHub Pages repository is dirty: %s", status)
	}
	if count := strings.TrimSpace(gitTest(t, destination, "rev-list", "--count", "HEAD")); count != "1" {
		t.Fatalf("GitHub Pages postprocess created %s commits, want one amended snapshot", count)
	}
}

func TestPromotePostprocessFilesExcludesIasiDevToolsRepository(t *testing.T) {
	root := t.TempDir()
	excluded := filepath.Join(root, "iasi-dev-tools")
	included := filepath.Join(root, "repo-a")

	for _, repository := range []string{excluded, included} {
		if err := os.MkdirAll(filepath.Join(repository, ".git"), 0755); err != nil {
			t.Fatal(err)
		}
	}

	excludedFile := filepath.Join(excluded, "config.toml")
	includedFile := filepath.Join(included, "config.toml")
	if err := os.WriteFile(excludedFile, []byte("organization = \"iasi-org-dev\"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(includedFile, []byte("organization = \"iasi-org-dev\"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	changed, err := promotePostprocessFiles(root, "iasi-org-dev", "iasi-org")
	if err != nil {
		t.Fatal(err)
	}
	if len(changed) != 1 || changed[0] != included {
		t.Fatalf("changed repositories = %v, want only %s", changed, included)
	}

	excludedData, err := os.ReadFile(excludedFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(excludedData) != "organization = \"iasi-org-dev\"\n" {
		t.Fatalf("iasi-dev-tools was postprocessed: %q", excludedData)
	}

	includedData, err := os.ReadFile(includedFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(includedData) != "organization = \"iasi-org\"\n" {
		t.Fatalf("normal repository was not postprocessed: %q", includedData)
	}
}
