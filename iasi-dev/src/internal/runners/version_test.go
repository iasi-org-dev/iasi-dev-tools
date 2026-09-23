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

func TestVersionSetsExplicitHigherVersion(t *testing.T) {
	rc := RC.OK
	parms := structures.Parms{
		Organization:  "iasi-org-dev",
		Version:       "v0.5.0",
		TargetVersion: "v0.7.0",
		RC:            &rc,
	}

	commands.SetDryRun(true)
	defer commands.SetDryRun(false)
	Version(&parms)

	if parms.Version != "v0.7.0" {
		t.Fatalf("Version = %q, want v0.7.0", parms.Version)
	}
}
