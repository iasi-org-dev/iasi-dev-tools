package args

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"iasi-dev/internal/consts"
	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

func writeTOML(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "iasi.toml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestPrepareDiscoversIASITargets(t *testing.T) {
	root := t.TempDir()
	repository := filepath.Join(root, "repo")
	project := filepath.Join(repository, "project")
	ignored := filepath.Join(repository, "tests", "ignored")

	if err := os.MkdirAll(filepath.Join(repository, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	writeTOML(t, repository, "type = \"repository\"\n")
	writeTOML(t, project, "type = \"quarto\"\n")
	writeTOML(t, ignored, "type = \"quarto\"\n")

	rc := RC.OK
	context := structures.Context{
		Targets:    []string{repository},
		Exclusions: append([]string{}, consts.RequiredExclusions...),
		RC:         &rc,
	}
	Prepare(&context)

	wantTargets := []string{filepath.Clean(repository), filepath.Clean(project)}
	if !reflect.DeepEqual(context.Targets, wantTargets) {
		t.Fatalf("Targets = %v, want %v", context.Targets, wantTargets)
	}
	if !reflect.DeepEqual(context.Repos, []string{filepath.Clean(repository)}) {
		t.Fatalf("Repos = %v, want [%s]", context.Repos, repository)
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
	writeTOML(t, projectA, "type = \"quarto\"\n")
	writeTOML(t, projectB, "type = \"quarto\"\n")

	rc := RC.OK
	context := structures.Context{
		Targets:    []string{projectA},
		Exclusions: append([]string{}, consts.RequiredExclusions...),
		RC:         &rc,
	}
	Prepare(&context)

	if !reflect.DeepEqual(context.Targets, []string{filepath.Clean(projectA)}) {
		t.Fatalf("Targets = %v, want [%s]", context.Targets, projectA)
	}
	if !reflect.DeepEqual(context.Repos, []string{filepath.Clean(repository)}) {
		t.Fatalf("Repos = %v, want [%s]", context.Repos, repository)
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
	writeTOML(t, repository, "type = \"repository\"\n")
	writeTOML(t, project, "type = \"software\"\n")
	writeTOML(t, manual, "type = \"quarto\"\n")
	writeTOML(t, withoutType, "[paths]\noutputs = \"_outputs\"\n")

	rc := RC.OK
	context := structures.Context{
		Targets:    []string{repository},
		Exclusions: append([]string{}, consts.RequiredExclusions...),
		RC:         &rc,
	}
	Prepare(&context)

	got := map[string]structures.Target{}
	for _, target := range context.TargetDetails {
		got[filepath.Base(target.Path)] = target
	}

	checks := []struct {
		name, typ string
		depth     int
	}{
		{"iasi-dev-tools", "repository", 0},
		{"iasi-dev", "software", 1},
		{"01-user-guide", "quarto", 2},
		{"without-type", "", 1},
	}

	for _, check := range checks {
		target, ok := got[check.name]
		if !ok {
			t.Fatalf("missing target %s in %v", check.name, context.TargetDetails)
		}
		typ := context.Config(target.Path).Type()
		if typ != check.typ || target.Depth != check.depth {
			t.Fatalf("%s = type %q depth %d, want type %q depth %d", check.name, typ, target.Depth, check.typ, check.depth)
		}
	}
}

func TestPrepareReadsSoftwareBuildConfiguration(t *testing.T) {
	root := t.TempDir()
	project := filepath.Join(root, "01-compiler")
	writeTOML(t, project, `type = "software"

[software]
builder = "go"
input-dir = "code"
output-dir = "../bin"
name = "compiler-x"
`)

	rc := RC.OK
	context := structures.Context{
		Targets:    []string{project},
		Exclusions: append([]string{}, consts.RequiredExclusions...),
		RC:         &rc,
	}
	Prepare(&context)

	config := context.Config(project)
	if config.Type() != "software" ||
		config.SoftwareBuilder() != "go" ||
		config.SoftwareInputDir() != "code" ||
		config.SoftwareOutputDir() != "../bin" ||
		config.SoftwareName("") != "compiler-x" {
		t.Fatalf("config = %+v", config.IASI)
	}
}

func TestPrepareAppliesSoftwareDefaults(t *testing.T) {
	root := t.TempDir()
	project := filepath.Join(root, "02-my-tool")
	writeTOML(t, project, `type = "software"

[software]
builder = "go"
`)

	rc := RC.OK
	context := structures.Context{
		Targets:    []string{project},
		Exclusions: append([]string{}, consts.RequiredExclusions...),
		RC:         &rc,
	}
	Prepare(&context)

	config := context.Config(project)
	if config.SoftwareInputDir() != "src" ||
		config.SoftwareOutputDir() != "_outputs" ||
		config.SoftwareName(defaultTargetName(filepath.Base(project))) != "my-tool" {
		t.Fatalf("software defaults = input %q output %q name %q",
			config.SoftwareInputDir(),
			config.SoftwareOutputDir(),
			config.SoftwareName(defaultTargetName(filepath.Base(project))),
		)
	}
	if !reflect.DeepEqual(context.Platforms, []string{"windows", "linux"}) {
		t.Fatalf("Platforms = %v", context.Platforms)
	}
}

func TestPrepareKeepsRequestedPlatform(t *testing.T) {
	root := t.TempDir()
	writeTOML(t, root, "type = \"software\"\n[software]\nbuilder = \"go\"\n")

	rc := RC.OK
	context := structures.Context{
		Targets:    []string{root},
		Platforms:  []string{"linux"},
		Exclusions: append([]string{}, consts.RequiredExclusions...),
		RC:         &rc,
	}
	Prepare(&context)

	if !reflect.DeepEqual(context.Platforms, []string{"linux"}) {
		t.Fatalf("Platforms = %v", context.Platforms)
	}
}

func TestPrepareLegacyConfigurationAddsAttention(t *testing.T) {
	root := t.TempDir()
	writeTOML(t, root, "type = \"quarto\"\n")
	if err := os.WriteFile(filepath.Join(root, "_iasi.yml"), []byte("type: quarto\n"), 0644); err != nil {
		t.Fatal(err)
	}

	rc := RC.OK
	context := structures.Context{
		Targets:    []string{root},
		Exclusions: append([]string{}, consts.RequiredExclusions...),
		RC:         &rc,
	}
	Prepare(&context)

	if rc&RC.Attention == 0 {
		t.Fatalf("RC = 0x%X, want ATTENTION", rc)
	}
}

func TestPrepareQuartoWithoutTOMLAddsAttention(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "_quarto.yml"), []byte("project:\n  type: book\n"), 0644); err != nil {
		t.Fatal(err)
	}

	rc := RC.OK
	context := structures.Context{
		Targets:    []string{root},
		Exclusions: append([]string{}, consts.RequiredExclusions...),
		RC:         &rc,
	}
	Prepare(&context)

	if rc&RC.Attention == 0 {
		t.Fatalf("RC = 0x%X, want ATTENTION", rc)
	}
	if len(context.TargetDetails) != 0 {
		t.Fatalf("TargetDetails = %v, want no IASI target without iasi.toml", context.TargetDetails)
	}
}
