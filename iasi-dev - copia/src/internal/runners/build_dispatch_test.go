package runners

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"iasi-dev/internal/commands"
	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

func TestDispatchBuildRoutesDocumentTypesToSameRBuilder(t *testing.T) {
	commands.SetDryRun(true)
	defer commands.SetDryRun(false)

	for _, targetType := range []string{"quarto", "book", "guide", "website"} {
		t.Run(targetType, func(t *testing.T) {
			root := t.TempDir()
			logPath := filepath.Join(root, "build.log")
			logFile, err := os.Create(logPath)
			if err != nil {
				t.Fatal(err)
			}

			target := structures.Target{Path: root, Type: targetType}
			parms := structures.Parms{DryRun: true, LogFile: logFile}

			if rc := dispatchBuild(target, parms, 0); rc != RC.OK {
				_ = logFile.Close()
				t.Fatalf("dispatchBuild(%s) RC = %d, want %d", targetType, rc, RC.OK)
			}
			if err := logFile.Close(); err != nil {
				t.Fatal(err)
			}

			data, err := os.ReadFile(logPath)
			if err != nil {
				t.Fatal(err)
			}
			text := string(data)
			if !strings.Contains(text, "Rscript -e") || !strings.Contains(text, "iasi::build()") {
				t.Fatalf("dispatchBuild(%s) did not use R build backend: %q", targetType, text)
			}
		})
	}
}

func TestDispatchBuildRoutesSoftwareToIASIScript(t *testing.T) {
	commands.SetDryRun(true)
	defer commands.SetDryRun(false)

	root := t.TempDir()
	logPath := filepath.Join(root, "build.log")
	logFile, err := os.Create(logPath)
	if err != nil {
		t.Fatal(err)
	}

	target := structures.Target{Path: root, Type: "software", Builder: "iasi-script"}
	parms := structures.Parms{DryRun: true, LogFile: logFile}

	if rc := dispatchBuild(target, parms, 0); rc != RC.OK {
		_ = logFile.Close()
		t.Fatalf("dispatchBuild(software:iasi-script) RC = %d, want %d", rc, RC.OK)
	}
	if err := logFile.Close(); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "Command ["+root+"]: iasi-script") {
		t.Fatalf("iasi-script command missing: %q", string(data))
	}
}

func TestDispatchBuildNothingToDoIsSilent(t *testing.T) {
	root := t.TempDir()
	logPath := filepath.Join(root, "build.log")
	logFile, err := os.Create(logPath)
	if err != nil {
		t.Fatal(err)
	}

	target := structures.Target{Path: filepath.Join(root, "mystery"), Type: "software", Builder: "unknown"}
	parms := structures.Parms{LogFile: logFile}

	if rc := dispatchBuild(target, parms, 0); rc != RC.NothingToDo {
		_ = logFile.Close()
		t.Fatalf("dispatchBuild unsupported RC = %d, want %d", rc, RC.NothingToDo)
	}
	if err := logFile.Close(); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(data)) != "" {
		t.Fatalf("NothingToDo wrote to the log: %q", string(data))
	}
}
