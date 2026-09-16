package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestResolveConfigPathDefaultsToConfigToml(t *testing.T) {
	workDir := filepath.Join("tmp", "project")
	got := resolveConfigPath(workDir, Descriptor{})
	want := filepath.Join(workDir, defaultConfigName)
	if got != want {
		t.Fatalf("resolveConfigPath() = %q, want %q", got, want)
	}
}

func TestLoadDescriptorReadsBuildMetadata(t *testing.T) {
	workDir := t.TempDir()
	content := []byte(`type: software
builder: iasi-script
name: iasi-net
input-dir: src
output-dir: ../../bin
config: custom.toml

targets:
  - powershell
  - bash
`)
	if err := os.WriteFile(filepath.Join(workDir, defaultDescriptorName), content, 0644); err != nil {
		t.Fatal(err)
	}

	descriptor, err := loadDescriptor(workDir)
	if err != nil {
		t.Fatal(err)
	}
	if descriptor.Config != "custom.toml" {
		t.Fatalf("Config = %q, want custom.toml", descriptor.Config)
	}
	if descriptor.Name != "iasi-net" {
		t.Fatalf("Name = %q, want iasi-net", descriptor.Name)
	}
	if descriptor.InputDir != "src" {
		t.Fatalf("InputDir = %q, want src", descriptor.InputDir)
	}
	if descriptor.OutputDir != "../../bin" {
		t.Fatalf("OutputDir = %q, want ../../bin", descriptor.OutputDir)
	}
	if !reflect.DeepEqual(descriptor.Targets, []string{"powershell", "bash"}) {
		t.Fatalf("Targets = %#v", descriptor.Targets)
	}
}

func TestLoadDescriptorAcceptsLegacyDotFile(t *testing.T) {
	workDir := t.TempDir()
	content := []byte("name: legacy\ntargets:\n  - powershell\n")
	if err := os.WriteFile(filepath.Join(workDir, legacyDescriptorName), content, 0644); err != nil {
		t.Fatal(err)
	}

	descriptor, err := loadDescriptor(workDir)
	if err != nil {
		t.Fatal(err)
	}
	if descriptor.Name != "legacy" {
		t.Fatalf("Name = %q, want legacy", descriptor.Name)
	}
}

func TestLoadConfigSeparatesTemplateMetadataFromData(t *testing.T) {
	workDir := t.TempDir()
	path := filepath.Join(workDir, "config.toml")
	content := []byte("template = \"network\"\nsubnet = \"10.77.0.0/24\"\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}
	config, err := loadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if config.Template != "network" {
		t.Fatalf("Template = %q, want network", config.Template)
	}
	if _, ok := config.Data["template"]; ok {
		t.Fatal("template metadata leaked into Data")
	}
	if config.Data["subnet"] != "10.77.0.0/24" {
		t.Fatalf("subnet = %#v", config.Data["subnet"])
	}
}
