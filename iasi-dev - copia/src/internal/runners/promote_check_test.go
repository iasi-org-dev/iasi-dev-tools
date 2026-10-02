package runners

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPromoteCheckPatternsIncludeSourceNameAndPath(t *testing.T) {
	root := filepath.Join(t.TempDir(), "iasi-org-dev")
	patterns := promoteCheckPatterns(root)

	for _, value := range []string{"iasi-org-dev", root, filepath.ToSlash(root)} {
		found := false
		for _, pattern := range patterns {
			if pattern.value == value {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("pattern %q not found in %v", value, patterns)
		}
	}
}

func TestPromoteCheckScanFindsContentPathAndGitConfig(t *testing.T) {
	root := filepath.Join(t.TempDir(), "iasi-org-dev")
	repository := filepath.Join(root, "iasi-org-dev.github.io")
	if err := os.MkdirAll(filepath.Join(repository, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repository, "index.html"), []byte("<a href=\"https://iasi-org-dev.github.io/iasi-home/\">IASI</a>\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repository, ".git", "config"), []byte("url = https://github.com/iasi-org/iasi-org-dev.github.io.git\n"), 0644); err != nil {
		t.Fatal(err)
	}

	hits, skipped, err := promoteCheckScan(root, promoteCheckPatterns(root))
	if err != nil {
		t.Fatal(err)
	}
	if skipped != 0 {
		t.Fatalf("skipped = %d, want 0", skipped)
	}

	content := false
	gitConfig := false
	pathName := false
	for _, hit := range hits {
		if filepath.Base(hit.path) == "index.html" && hit.line == 1 {
			content = true
		}
		if filepath.Base(hit.path) == "config" && filepath.Base(filepath.Dir(hit.path)) == ".git" {
			gitConfig = true
		}
		if filepath.Base(hit.path) == "iasi-org-dev.github.io" && hit.line == 0 {
			pathName = true
		}
	}
	if !content || !gitConfig || !pathName {
		t.Fatalf("hits missing expected cases: %+v", hits)
	}
}

func TestPromoteCheckScanSkipsBinaryContent(t *testing.T) {
	root := filepath.Join(t.TempDir(), "iasi-org-dev")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "binary.bin"), []byte{'x', 0, 'i', 'a', 's', 'i', '-', 'o', 'r', 'g', '-', 'd', 'e', 'v'}, 0644); err != nil {
		t.Fatal(err)
	}

	hits, skipped, err := promoteCheckScan(root, promoteCheckPatterns(root))
	if err != nil {
		t.Fatal(err)
	}
	if skipped != 1 {
		t.Fatalf("skipped = %d, want 1", skipped)
	}
	for _, hit := range hits {
		if filepath.Base(hit.path) == "binary.bin" && hit.line > 0 {
			t.Fatalf("binary content was scanned: %+v", hit)
		}
	}
}


func TestPromoteCheckScanPromotionScansDestinationOnly(t *testing.T) {
	base := t.TempDir()
	source := filepath.Join(base, "iasi-org-dev")
	destination := filepath.Join(base, "iasi-org")
	if err := os.MkdirAll(source, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(destination, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "source-only.txt"), []byte("https://github.com/iasi-org-dev/source-only\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(destination, "destination.txt"), []byte("https://github.com/iasi-org-dev/destination\n"), 0644); err != nil {
		t.Fatal(err)
	}

	hits, skipped, err := promoteCheckScanPromotion(source, destination)
	if err != nil {
		t.Fatal(err)
	}
	if skipped != 0 {
		t.Fatalf("skipped = %d, want 0", skipped)
	}

	destinationFound := false
	for _, hit := range hits {
		if filepath.Base(hit.path) == "source-only.txt" {
			t.Fatalf("source tree must not be scanned: %+v", hit)
		}
		if filepath.Base(hit.path) == "destination.txt" {
			destinationFound = true
		}
	}
	if !destinationFound {
		t.Fatalf("expected suspicious reference in destination: %+v", hits)
	}
}

func TestPromoteCheckScanIgnoresLogsDirectories(t *testing.T) {
	root := filepath.Join(t.TempDir(), "iasi-org-dev")
	if err := os.MkdirAll(filepath.Join(root, "repo", "logs"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "repo", "logs", "run.log"), []byte("C:/iasi-org-dev/repo\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "repo", "README.md"), []byte("https://github.com/iasi-org-dev/repo\n"), 0644); err != nil {
		t.Fatal(err)
	}

	hits, skipped, err := promoteCheckScan(root, promoteCheckPatterns(root))
	if err != nil {
		t.Fatal(err)
	}
	if skipped != 0 {
		t.Fatalf("skipped = %d, want 0", skipped)
	}

	readmeFound := false
	for _, hit := range hits {
		if strings.Contains(strings.ToLower(hit.path), string(filepath.Separator)+"logs"+string(filepath.Separator)) {
			t.Fatalf("logs directory should be ignored: %+v", hit)
		}
		if filepath.Base(hit.path) == "README.md" {
			readmeFound = true
		}
	}
	if !readmeFound {
		t.Fatalf("expected non-log suspicious reference to remain: %+v", hits)
	}
}

func TestPromoteCheckScanExcludesIasiDevToolsRepository(t *testing.T) {
	base := t.TempDir()
	source := filepath.Join(base, "iasi-org-dev")
	destination := filepath.Join(base, "iasi-org")
	excluded := filepath.Join(destination, "iasi-dev-tools")
	included := filepath.Join(destination, "repo-a")

	if err := os.MkdirAll(excluded, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(included, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(excluded, "config.toml"), []byte("organization = \"iasi-org-dev\"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(included, "LICENSE"), []byte("Copyright (c) 2026 iasi-org-dev\n"), 0644); err != nil {
		t.Fatal(err)
	}

	hits, skipped, err := promoteCheckScan(destination, promoteCheckPatterns(source))
	if err != nil {
		t.Fatal(err)
	}
	if skipped != 0 {
		t.Fatalf("skipped = %d, want 0", skipped)
	}

	includedFound := false
	for _, hit := range hits {
		if strings.Contains(hit.path, "iasi-dev-tools") {
			t.Fatalf("iasi-dev-tools must be excluded from promote check: %+v", hit)
		}
		if filepath.Base(hit.path) == "LICENSE" {
			includedFound = true
		}
	}
	if !includedFound {
		t.Fatalf("expected suspicious reference outside excluded repository: %+v", hits)
	}
}
