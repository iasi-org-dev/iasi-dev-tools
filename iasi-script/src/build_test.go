package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"iasi-script/internal/consts/EXT"
)

func TestResolveTemplatePrefersTargetSpecific(t *testing.T) {
	inputDir := t.TempDir()
	generic := filepath.Join(inputDir, "network.tpl")
	specific := filepath.Join(inputDir, "network.ps1.tpl")
	if err := os.WriteFile(generic, []byte("generic"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(specific, []byte("specific"), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := resolveTemplate(inputDir, "network", "powershell")
	if err != nil {
		t.Fatal(err)
	}
	if got != specific {
		t.Fatalf("resolveTemplate() = %q, want %q", got, specific)
	}
}

func TestResolveTemplateFallsBackToGeneric(t *testing.T) {
	inputDir := t.TempDir()
	generic := filepath.Join(inputDir, "network.tpl")
	if err := os.WriteFile(generic, []byte("generic"), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := resolveTemplate(inputDir, "network", "bash")
	if err != nil {
		t.Fatal(err)
	}
	if got != generic {
		t.Fatalf("resolveTemplate() = %q, want %q", got, generic)
	}
}

func TestResolveTemplateReportsBothCandidates(t *testing.T) {
	inputDir := t.TempDir()
	_, err := resolveTemplate(inputDir, "network", "powershell")
	if err == nil {
		t.Fatal("resolveTemplate() returned nil error")
	}
	for _, want := range []string{"network.ps1.tpl", "network.tpl"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error %q does not mention %q", err, want)
		}
	}
}

func TestArtifactNameUsesConfiguredName(t *testing.T) {
	if got := artifactName(filepath.Join("tmp", "iasi-net"), "network"); got != "network" {
		t.Fatalf("artifactName() = %q, want network", got)
	}
}

func TestArtifactNameFallsBackToDirectory(t *testing.T) {
	if got := artifactName(filepath.Join("tmp", "iasi-net"), ""); got != "iasi-net" {
		t.Fatalf("artifactName() = %q, want iasi-net", got)
	}
}

func TestTargetExtension(t *testing.T) {
	tests := map[string]string{
		"powershell": EXT.PowerShell,
		"bash":       EXT.Bash,
	}
	for target, want := range tests {
		got, err := targetExtension(target)
		if err != nil {
			t.Fatalf("targetExtension(%q): %v", target, err)
		}
		if got != want {
			t.Fatalf("targetExtension(%q) = %q, want %q", target, got, want)
		}
	}
}

func TestBuildUsesIasiDescriptorForNameDirsAndTargets(t *testing.T) {
	root := t.TempDir()
	workDir := filepath.Join(root, "artifacts", "iasi-net")
	inputDir := filepath.Join(workDir, "src")
	outputDir := filepath.Join(root, "bin")
	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inputDir, "network.ps1.tpl"), []byte("Hello {{.who}}"), 0644); err != nil {
		t.Fatal(err)
	}

	descriptor := Descriptor{
		Name:      "iasi-net",
		InputDir:  "src",
		OutputDir: "../../bin",
		Targets:   []string{"powershell"},
	}
	config := Config{Template: "network", Data: map[string]any{"who": "IASI"}}
	if err := build(workDir, descriptor, config); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filepath.Join(outputDir, "iasi-net.ps1"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "Hello IASI" {
		t.Fatalf("output = %q, want %q", content, "Hello IASI")
	}
	if _, err := os.Stat(filepath.Join(workDir, "_outputs")); !os.IsNotExist(err) {
		t.Fatalf("unexpected _outputs directory")
	}
}

func TestBuildFallsBackToDirectoryNameAndDefaultDirs(t *testing.T) {
	root := t.TempDir()
	workDir := filepath.Join(root, "iasi-net")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workDir, "network.tpl"), []byte("#!/usr/bin/env bash"), 0644); err != nil {
		t.Fatal(err)
	}

	descriptor := Descriptor{Targets: []string{"bash"}}
	config := Config{Template: "network", Data: map[string]any{}}
	if err := build(workDir, descriptor, config); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(workDir, "_outputs", "iasi-net.sh")); err != nil {
		t.Fatal(err)
	}
}

func TestBuildAllDescriptorTargets(t *testing.T) {
	workDir := t.TempDir()
	inputDir := filepath.Join(workDir, "src")
	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{"powershell", "bash"} {
		if err := os.WriteFile(filepath.Join(inputDir, "network."+map[string]string{"powershell": EXT.PowerShell, "bash": EXT.Bash}[target]+".tpl"), []byte(target), 0644); err != nil {
			t.Fatal(err)
		}
	}

	descriptor := Descriptor{Name: "iasi-net", InputDir: "src", OutputDir: "out", Targets: []string{"powershell", "bash"}}
	config := Config{Template: "network", Data: map[string]any{}}
	if err := build(workDir, descriptor, config); err != nil {
		t.Fatal(err)
	}
	for _, output := range []string{"iasi-net." + EXT.PowerShell, "iasi-net." + EXT.Bash} {
		if _, err := os.Stat(filepath.Join(workDir, "out", output)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestBuildAllowsConfigWithoutTemplates(t *testing.T) {
	if err := build(t.TempDir(), Descriptor{}, Config{Data: map[string]any{"name": "IASI"}}); err != nil {
		t.Fatalf("build() returned error: %v", err)
	}
}
