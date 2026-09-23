package main

import (
	"os"
	"path/filepath"
	"testing"

	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

func TestPreparePathChangesWorkingDirectory(t *testing.T) {
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(original)

	target := t.TempDir()
	rc := RC.OK
	Parms := structures.Parms{Path: target, RC: &rc}
	preparePath(&Parms)

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	want, _ := filepath.Abs(target)
	if filepath.Clean(cwd) != filepath.Clean(want) {
		t.Fatalf("cwd = %q, want %q", cwd, want)
	}
	if filepath.Clean(Parms.Path) != filepath.Clean(want) {
		t.Fatalf("Parms.Path = %q, want %q", Parms.Path, want)
	}
}

func TestPreparePromotePathsResolvesRelativePathsFromCurrentDirectory(t *testing.T) {
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(original)

	base := t.TempDir()
	if err := os.Chdir(base); err != nil {
		t.Fatal(err)
	}

	rc := RC.OK
	Parms := structures.Parms{SourcePath: "iasi-org-dev", DestinationPath: filepath.Join(".", "iasi-org"), RC: &rc}
	preparePromotePaths(&Parms)

	if Parms.SourcePath != filepath.Join(base, "iasi-org-dev") {
		t.Fatalf("SourcePath = %q, want %q", Parms.SourcePath, filepath.Join(base, "iasi-org-dev"))
	}
	if Parms.DestinationPath != filepath.Join(base, "iasi-org") {
		t.Fatalf("DestinationPath = %q, want %q", Parms.DestinationPath, filepath.Join(base, "iasi-org"))
	}
}

func TestPreparePromotePathsUsesPathAsRelativeBase(t *testing.T) {
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(original)

	base := t.TempDir()
	rc := RC.OK
	Parms := structures.Parms{Path: base, SourcePath: "iasi-org-dev", DestinationPath: "iasi-org", RC: &rc}
	preparePath(&Parms)
	preparePromotePaths(&Parms)

	if Parms.SourcePath != filepath.Join(base, "iasi-org-dev") {
		t.Fatalf("SourcePath = %q, want %q", Parms.SourcePath, filepath.Join(base, "iasi-org-dev"))
	}
	if Parms.DestinationPath != filepath.Join(base, "iasi-org") {
		t.Fatalf("DestinationPath = %q, want %q", Parms.DestinationPath, filepath.Join(base, "iasi-org"))
	}
}
