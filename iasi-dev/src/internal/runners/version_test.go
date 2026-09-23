package runners

import (
	"io"
	"os"
	"strings"
	"testing"

	"iasi-dev/internal/commands"
	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

func TestVersionPrintsPreparedVersion(t *testing.T) {
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer read.Close()

	original := os.Stdout
	os.Stdout = write
	Version(&structures.Parms{Version: "v0.5.0"})
	if err := write.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdout = original

	data, err := io.ReadAll(read)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(string(data)); got != "v0.5.0" {
		t.Fatalf("Version output = %q, want v0.5.0", got)
	}
}

func TestVersionSetsExplicitVersion(t *testing.T) {
	rc := RC.OK
	parms := structures.Parms{
		Organization:  "iasi-org-dev",
		Version:       "v0.5.0",
		TargetVersion: "v0.4.0",
		RC:            &rc,
	}

	commands.SetDryRun(true)
	defer commands.SetDryRun(false)
	Version(&parms)

	if parms.Version != "v0.4.0" {
		t.Fatalf("Version = %q, want v0.4.0", parms.Version)
	}
}

func TestSetOrganizationVersionForUsesExplicitOrganization(t *testing.T) {
	rc := RC.OK
	logFile, err := os.CreateTemp(t.TempDir(), "version-*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer logFile.Close()

	parms := structures.Parms{RC: &rc, LogFile: logFile}
	commands.SetDryRun(true)
	defer commands.SetDryRun(false)
	setOrganizationVersionFor(&parms, "iasi-org", "v0.5.0")

	if err := logFile.Sync(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(logFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if !strings.Contains(text, "gh variable set VERSION --org iasi-org --body v0.5.0") {
		t.Fatalf("version propagation command not found in log: %s", text)
	}
}
