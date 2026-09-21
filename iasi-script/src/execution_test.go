package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"iasi-script/internal/parms"
)

func TestPrepareExecutionDefaultsToConfigToml(t *testing.T) {
	workingDir := t.TempDir()
	configPath := filepath.Join(workingDir, defaultConfigName)
	if err := os.WriteFile(configPath, []byte("name = \"IASI\"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	execution, err := prepareExecution(parms.Parms{}, Environment{CurrentDir: workingDir})
	if err != nil {
		t.Fatal(err)
	}
	if execution.WorkingDir != workingDir {
		t.Fatalf("WorkingDir = %q, want %q", execution.WorkingDir, workingDir)
	}
	if !reflect.DeepEqual(execution.ConfigPaths, []string{configPath}) {
		t.Fatalf("ConfigPaths = %#v, want %#v", execution.ConfigPaths, []string{configPath})
	}
}

func TestPrepareExecutionResolvesFileFromWorkingDir(t *testing.T) {
	workingDir := t.TempDir()
	configPath := filepath.Join(workingDir, "iasi-tools.toml")
	if err := os.WriteFile(configPath, []byte("name = \"iasi-tools\"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	Parms := parms.Parms{File: "iasi-tools.toml"}
	execution, err := prepareExecution(Parms, Environment{CurrentDir: workingDir})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(execution.ConfigPaths, []string{configPath}) {
		t.Fatalf("ConfigPaths = %#v, want %#v", execution.ConfigPaths, []string{configPath})
	}
}

func TestPrepareExecutionFindsRootTomlRecursively(t *testing.T) {
	workingDir := t.TempDir()
	root := filepath.Join(workingDir, "machines")
	nested := filepath.Join(root, "linux")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatal(err)
	}

	paths := []string{
		filepath.Join(root, "iasi-ai.toml"),
		filepath.Join(nested, "iasi-tools.toml"),
	}
	for _, path := range paths {
		if err := os.WriteFile(path, []byte("name = \"IASI\"\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(nested, "notes.txt"), []byte("ignored"), 0644); err != nil {
		t.Fatal(err)
	}

	Parms := parms.Parms{Root: "machines"}
	execution, err := prepareExecution(Parms, Environment{CurrentDir: workingDir})
	if err != nil {
		t.Fatal(err)
	}

	want := []string{paths[0], paths[1]}
	if !reflect.DeepEqual(execution.ConfigPaths, want) {
		t.Fatalf("ConfigPaths = %#v, want %#v", execution.ConfigPaths, want)
	}
}
