package main

import (
	"io"
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

func TestLogParmsIncludesPreparedListsAndModes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "parms.log")
	logFile, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}

	Parms := structures.Parms{
		PrepareOnly:      true,
		DryRun:           false,
		LogDir:           "trace",
		RequestedTargets: []string{"."},
		Targets:          []string{"/workspace"},
		Repos:            []string{"/workspace/repo"},
		Exclusions:       []string{".git", "tests"},
		LogFile:          logFile,
	}
	logParms(Parms)
	if err := logFile.Close(); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(content)
	for _, want := range []string{
		"PrepareOnly: true",
		"DryRun: false",
		`LogDir: "trace"`,
		"RequestedTargets:\n  - .",
		"Targets:\n  workspace [none]",
		"Repos:\n  - /workspace/repo",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("log does not contain %q:\n%s", want, text)
		}
	}
}

func TestLogParmsPrepareOnlyMirrorsPreparedData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "parms.log")
	logFile, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer logFile.Close()

	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer read.Close()

	old := os.Stdout
	os.Stdout = write
	defer func() { os.Stdout = old }()

	Parms := structures.Parms{PrepareOnly: true, RequestedTargets: []string{"."}, Targets: []string{"/workspace"}, Repos: []string{"/workspace/repo"}, LogFile: logFile}
	logParms(Parms)
	if err := write.Close(); err != nil {
		t.Fatal(err)
	}

	console, err := io.ReadAll(read)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(console), "RequestedTargets:\n  - .") {
		t.Fatalf("prepare-only output missing requested targets: %q", string(console))
	}
	if !strings.Contains(string(console), "Repos:\n  - /workspace/repo") {
		t.Fatalf("prepare-only output missing repos: %q", string(console))
	}
}

func TestLogParmsWithoutCheckModeDoesNotMirror(t *testing.T) {
	path := filepath.Join(t.TempDir(), "parms.log")
	logFile, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer logFile.Close()

	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer read.Close()

	old := os.Stdout
	os.Stdout = write
	defer func() { os.Stdout = old }()

	logParms(structures.Parms{Targets: []string{"/workspace"}, LogFile: logFile})
	if err := write.Close(); err != nil {
		t.Fatal(err)
	}

	console, err := io.ReadAll(read)
	if err != nil {
		t.Fatal(err)
	}
	if len(console) != 0 {
		t.Fatalf("normal execution mirrored parms: %q", string(console))
	}
}

func TestLogParmsWritesTargetTree(t *testing.T) {
	path := filepath.Join(t.TempDir(), "parms.log")
	logFile, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}

	Parms := structures.Parms{
		TargetDetails: []structures.Target{
			{Path: filepath.Join("root", "iasi-dev-tools"), Type: "repository", Depth: 0},
			{Path: filepath.Join("root", "iasi-dev-tools", "iasi-dev"), Type: "r", Depth: 1},
			{Path: filepath.Join("root", "iasi-dev-tools", "iasi-dev", "docs", "01-user-guide"), Type: "quarto", Depth: 2},
		},
		LogFile: logFile,
	}
	logParms(Parms)
	if err := logFile.Close(); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := "Targets:\n  iasi-dev-tools [repository]\n    iasi-dev [r]\n      01-user-guide [quarto]\n"
	if !strings.Contains(string(data), want) {
		t.Fatalf("target tree missing:\n%s", string(data))
	}
}
