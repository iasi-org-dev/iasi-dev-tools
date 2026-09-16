package args

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"iasi-dev/internal/consts"
	"iasi-dev/internal/structures"
)

func TestPrepareDiscoversIASITargets(t *testing.T) {
	root := t.TempDir()
	repository := filepath.Join(root, "repo")
	project := filepath.Join(repository, "project")
	ignored := filepath.Join(repository, "tests", "ignored")

	if err := os.MkdirAll(filepath.Join(repository, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(project, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(ignored, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repository, ".iasi.yml"), []byte("type: repo\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, "_iasi.yml"), []byte("type: project\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ignored, "_iasi.yml"), []byte("type: ignored\n"), 0644); err != nil {
		t.Fatal(err)
	}

	parms := structures.Parms{Targets: []string{repository}, Exclusions: append([]string{}, consts.RequiredExclusions...)}
	Prepare(&parms)

	wantTargets := []string{filepath.Clean(repository), filepath.Clean(project)}
	if !reflect.DeepEqual(parms.Targets, wantTargets) {
		t.Fatalf("Targets = %v, want %v", parms.Targets, wantTargets)
	}
	if !reflect.DeepEqual(parms.Repos, []string{filepath.Clean(repository)}) {
		t.Fatalf("Repos = %v, want [%s]", parms.Repos, repository)
	}
}

func TestPrepareKeepsRequestedSubdirectoryAsDiscoveryScope(t *testing.T) {
	root := t.TempDir()
	repository := filepath.Join(root, "repo")
	projectA := filepath.Join(repository, "a")
	projectB := filepath.Join(repository, "b")

	if err := os.MkdirAll(filepath.Join(repository, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(projectA, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(projectB, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectA, "_iasi.yml"), []byte("type: a\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectB, "_iasi.yml"), []byte("type: b\n"), 0644); err != nil {
		t.Fatal(err)
	}

	parms := structures.Parms{Targets: []string{projectA}, Exclusions: append([]string{}, consts.RequiredExclusions...)}
	Prepare(&parms)

	if !reflect.DeepEqual(parms.Targets, []string{filepath.Clean(projectA)}) {
		t.Fatalf("Targets = %v, want [%s]", parms.Targets, projectA)
	}
	if !reflect.DeepEqual(parms.Repos, []string{filepath.Clean(repository)}) {
		t.Fatalf("Repos = %v, want [%s]", parms.Repos, repository)
	}
}

func TestPrepareDescribesTargetTypesAndDepth(t *testing.T) {
	root := t.TempDir()
	repository := filepath.Join(root, "iasi-dev-tools")
	project := filepath.Join(repository, "iasi-dev")
	manual := filepath.Join(project, "docs", "01-user-guide")
	withoutType := filepath.Join(repository, "without-type")

	if err := os.MkdirAll(filepath.Join(repository, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(manual, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(withoutType, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repository, ".iasi.yml"), []byte("type: repository\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, ".iasi.yml"), []byte("type: software\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(manual, "_iasi.yml"), []byte("type: quarto\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(withoutType, "_iasi.yml"), []byte("name: sample\n"), 0644); err != nil {
		t.Fatal(err)
	}

	parms := structures.Parms{Targets: []string{repository}, Exclusions: append([]string{}, consts.RequiredExclusions...)}
	Prepare(&parms)

	got := map[string]structures.Target{}
	for _, target := range parms.TargetDetails {
		got[filepath.Base(target.Path)] = target
	}

	checks := []struct {
		name, typ string
		depth     int
	}{
		{"iasi-dev-tools", "repository", 0},
		{"iasi-dev", "software", 1},
		{"01-user-guide", "quarto", 2},
		{"without-type", "none", 1},
	}
	for _, check := range checks {
		target, ok := got[check.name]
		if !ok {
			t.Fatalf("missing target %s in %v", check.name, parms.TargetDetails)
		}
		if target.Type != check.typ || target.Depth != check.depth {
			t.Fatalf("%s = type %q depth %d, want type %q depth %d", check.name, target.Type, target.Depth, check.typ, check.depth)
		}
	}
}

func TestPrepareReadsSoftwareBuildConfiguration(t *testing.T) {
	root := t.TempDir()
	project := filepath.Join(root, "01-compiler")
	if err := os.MkdirAll(project, 0755); err != nil {
		t.Fatal(err)
	}
	config := "type: software\n\nbuilder: go\n\nsource-dir: code\noutput-dir: ../bin\nname: compiler-x\n"
	if err := os.WriteFile(filepath.Join(project, ".iasi.yml"), []byte(config), 0644); err != nil {
		t.Fatal(err)
	}

	parms := structures.Parms{Targets: []string{project}, Exclusions: append([]string{}, consts.RequiredExclusions...)}
	Prepare(&parms)
	if len(parms.TargetDetails) != 1 {
		t.Fatalf("TargetDetails = %v", parms.TargetDetails)
	}
	target := parms.TargetDetails[0]
	if target.Type != "software" || target.Builder != "go" || target.SourceDir != "code" || target.OutputDir != "../bin" || target.Name != "compiler-x" {
		t.Fatalf("target = %+v", target)
	}
}

func TestPrepareAppliesSoftwareDefaults(t *testing.T) {
	root := t.TempDir()
	project := filepath.Join(root, "02-my-tool")
	if err := os.MkdirAll(project, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, ".iasi.yml"), []byte("type: software\nbuilder: go\n"), 0644); err != nil {
		t.Fatal(err)
	}

	parms := structures.Parms{Targets: []string{project}, Exclusions: append([]string{}, consts.RequiredExclusions...)}
	Prepare(&parms)
	target := parms.TargetDetails[0]
	if target.SourceDir != "src" || target.OutputDir != "_outputs" || target.Name != "my-tool" {
		t.Fatalf("defaults = source %q output %q name %q", target.SourceDir, target.OutputDir, target.Name)
	}
	if !reflect.DeepEqual(parms.Platforms, []string{"windows", "linux"}) {
		t.Fatalf("Platforms = %v", parms.Platforms)
	}
}

func TestPrepareKeepsRequestedPlatform(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".iasi.yml"), []byte("type: software\nbuilder: go\n"), 0644); err != nil {
		t.Fatal(err)
	}
	parms := structures.Parms{Targets: []string{root}, Platforms: []string{"linux"}, Exclusions: append([]string{}, consts.RequiredExclusions...)}
	Prepare(&parms)
	if !reflect.DeepEqual(parms.Platforms, []string{"linux"}) {
		t.Fatalf("Platforms = %v", parms.Platforms)
	}
}
