package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"iasi-dev/internal/structures"
)

func TestCreateLogFileUsesRequestedDirectory(t *testing.T) {
	dir := t.TempDir()
	logFile, err := createLogFile("build", dir)
	if err != nil {
		t.Fatal(err)
	}
	defer logFile.Close()

	if filepath.Dir(logFile.Name()) != filepath.Clean(dir) {
		t.Fatalf("log dir = %q, want %q", filepath.Dir(logFile.Name()), filepath.Clean(dir))
	}
}

func TestLogContextWritesTargetTreeFromCanonicalConfigs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "context.log")
	logFile, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}

	repo := filepath.Join("root", "iasi-dev-tools")
	dev := filepath.Join(repo, "iasi-dev")
	guide := filepath.Join(dev, "docs", "01-user-guide")

	context := structures.Context{
		TargetDetails: []structures.Target{
			{Path: repo, Depth: 0},
			{Path: dev, Depth: 1},
			{Path: guide, Depth: 2},
		},
		Configs: map[string]structures.Config{
			filepath.Clean(repo): {IASI: map[string]any{"type": "repository"}},
			filepath.Clean(dev): {IASI: map[string]any{
				"type": "software",
				"software": map[string]any{
					"builder": "go",
					"input-dir": "src",
				},
			}},
			filepath.Clean(guide): {IASI: map[string]any{
				"type": "quarto",
				"publication": map[string]any{
					"strategy": "outlined",
					"numbered": true,
				},
			}},
		},
		LogFile: logFile,
	}

	logParms(context)
	if err := logFile.Close(); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)

	for _, want := range []string{
		"--- CONTEXT",
		"iasi-dev-tools [repository]",
		"iasi-dev [software]",
		"01-user-guide [quarto]",
		"Configs:",
		"type: \"repository\"",
		"software:",
		"builder: \"go\"",
		"input-dir: \"src\"",
		"publication:",
		"strategy: \"outlined\"",
		"numbered: true",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("log does not contain %q:\n%s", want, text)
		}
	}
}
