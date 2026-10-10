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
	context := structures.Context{Path: target, RC: &rc}
	preparePath(&context)

	cwd, _ := os.Getwd()
	want, _ := filepath.Abs(target)
	if filepath.Clean(cwd) != filepath.Clean(want) {
		t.Fatalf("cwd = %q, want %q", cwd, want)
	}
}
