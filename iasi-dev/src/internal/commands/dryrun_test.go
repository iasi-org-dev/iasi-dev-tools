package commands

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"iasi-dev/internal/consts/RC"
)

func TestDryRunLogsWithoutExecuting(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "dry-run.log")
	logFile, err := os.Create(logPath)
	if err != nil {
		t.Fatal(err)
	}
	defer logFile.Close()

	SetDryRun(true)
	defer SetDryRun(false)

	result := Run(".", logFile, "iasi-command-that-must-not-exist", "arg")
	if result.RC != RC.OK {
		t.Fatalf("RC = %d, want %d", result.RC, RC.OK)
	}

	if err := logFile.Sync(); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "iasi-command-that-must-not-exist arg") {
		t.Fatalf("log does not contain simulated command: %s", content)
	}
}

func TestDryRunGitStatusDoesNotBecomeNothingToDo(t *testing.T) {
	SetDryRun(true)
	defer SetDryRun(false)

	result := RunFriendly(".", nil, "git", "status")
	if result.RC != RC.OK {
		t.Fatalf("RC = %d, want %d", result.RC, RC.OK)
	}
}

func TestDryRunMirrorsCommandInOneLine(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "dry-run.log")
	logFile, err := os.Create(logPath)
	if err != nil { t.Fatal(err) }
	defer logFile.Close()

	read, write, err := os.Pipe()
	if err != nil { t.Fatal(err) }
	defer read.Close()

	old := os.Stdout
	os.Stdout = write
	defer func() { os.Stdout = old }()

	SetDryRun(true)
	defer SetDryRun(false)
	Run("C:/workspace", logFile, "tool", "one", "two words")
	if err := write.Close(); err != nil { t.Fatal(err) }

	console, err := io.ReadAll(read)
	if err != nil { t.Fatal(err) }
	text := string(console)
	if !strings.Contains(text, `Command [C:/workspace]: tool one "two words"`) { t.Fatalf("dry-run command not mirrored as one line: %q", text) }
	if strings.Count(strings.TrimSpace(text), "\n") != 0 { t.Fatalf("dry-run command spans multiple lines: %q", text) }
}
