package runners

import (
	"os"
	"path/filepath"
	"testing"

	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

func TestBuildGoBuildsWindowsAndLinuxIntoSameOutputDirectory(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "src")
	output := filepath.Join(root, "bin")
	if err := os.MkdirAll(source, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "go.mod"), []byte("module example.test/tool\n\ngo 1.22\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "main.go"), []byte("package main\nfunc main() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	logFile, err := os.Create(filepath.Join(root, "build.log"))
	if err != nil {
		t.Fatal(err)
	}
	defer logFile.Close()

	target := structures.Target{
		Path:      root,
		Type:      "software",
		Builder:   "go",
		SourceDir: "src",
		OutputDir: "bin",
		Name:      "tool",
	}
	parms := structures.Parms{Platforms: []string{"windows", "linux"}, LogFile: logFile}

	if rc := buildGo(target, parms); rc != RC.OK {
		t.Fatalf("buildGo RC = %d", rc)
	}
	for _, path := range []string{filepath.Join(output, "tool.exe"), filepath.Join(output, "tool")} {
		if info, err := os.Stat(path); err != nil || info.IsDir() {
			t.Fatalf("missing built artifact %s: %v", path, err)
		}
	}
}
