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

func testContextForTarget(path string, typ string) structures.Context {
	rc := RC.OK
	return structures.Context{
		RC: &rc,
		Configs: map[string]structures.Config{
			filepath.Clean(path): {IASI: map[string]any{"type": typ}},
		},
	}
}

func TestDispatchBuildRoutesCanonicalRTypesToSameRBuilder(t *testing.T) {
	commands.SetDryRun(true)
	defer commands.SetDryRun(false)

	for _, targetType := range []string{"quarto", "website", "r-package"} {
		t.Run(targetType, func(t *testing.T) {
			root := t.TempDir()
			logPath := filepath.Join(root, "build.log")
			logFile, err := os.Create(logPath)
			if err != nil {
				t.Fatal(err)
			}

			target := structures.Target{Path: root}
			context := testContextForTarget(root, targetType)
			context.DryRun = true
			context.LogFile = logFile

			if rc := dispatchBuild(target, context, 0); rc != RC.OK {
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

func TestDispatchBuildRejectsHistoricalBookAndGuideTypes(t *testing.T) {
	for _, targetType := range []string{"book", "guide", "r"} {
		root := t.TempDir()
		target := structures.Target{Path: root}
		context := testContextForTarget(root, targetType)

		if rc := dispatchBuild(target, context, 0); rc != RC.NothingToDo {
			t.Fatalf("dispatchBuild(%s) RC = %d, want %d", targetType, rc, RC.NothingToDo)
		}
	}
}

func TestDispatchBuildDoesNotSupportIASIScript(t *testing.T) {
	root := t.TempDir()
	target := structures.Target{Path: root}
	context := testContextForTarget(root, "software")
	context.Configs[filepath.Clean(root)] = structures.Config{
		IASI: map[string]any{
			"type": "software",
			"software": map[string]any{
				"builder": "iasi-script",
			},
		},
	}

	if rc := dispatchBuild(target, context, 0); rc != RC.NothingToDo {
		t.Fatalf("dispatchBuild(software:iasi-script) RC = %d, want %d", rc, RC.NothingToDo)
	}
}
