package main

import (
	"testing"
)

func TestProcessArgumentsDefaults(t *testing.T) {
	Parms, err := processArguments(nil)
	if err != nil {
		t.Fatal(err)
	}
	if Parms.Help || Parms.File != "" || Parms.Root != "" || Parms.WorkingDir != "" {
		t.Fatalf("unexpected parameters: %#v", Parms)
	}
}

func TestProcessArgumentsFileAndWorkingDir(t *testing.T) {
	Parms, err := processArguments([]string{"--file", "iasi-tools.toml", "workspace"})
	if err != nil {
		t.Fatal(err)
	}
	if Parms.File != "iasi-tools.toml" {
		t.Fatalf("File = %q, want iasi-tools.toml", Parms.File)
	}
	if Parms.WorkingDir != "workspace" {
		t.Fatalf("WorkingDir = %q, want workspace", Parms.WorkingDir)
	}
}

func TestProcessArgumentsRootAndWorkingDir(t *testing.T) {
	Parms, err := processArguments([]string{"--root", "machines", "workspace"})
	if err != nil {
		t.Fatal(err)
	}
	if Parms.Root != "machines" {
		t.Fatalf("Root = %q, want machines", Parms.Root)
	}
	if Parms.WorkingDir != "workspace" {
		t.Fatalf("WorkingDir = %q, want workspace", Parms.WorkingDir)
	}
}

func TestProcessArgumentsRejectsFileAndRootTogether(t *testing.T) {
	_, err := processArguments([]string{"--file", "one.toml", "--root", "machines"})
	if err == nil {
		t.Fatal("processArguments() returned nil error")
	}
}

func TestProcessArgumentsRequiresWorkingDirLast(t *testing.T) {
	_, err := processArguments([]string{"workspace", "--file", "iasi-tools.toml"})
	if err == nil {
		t.Fatal("processArguments() returned nil error")
	}
}

func TestProcessArgumentsHelpFlag(t *testing.T) {
	Parms, err := processArguments([]string{"-h"})
	if err != nil {
		t.Fatal(err)
	}
	if !Parms.Help {
		t.Fatal("Help = false, want true")
	}
}
