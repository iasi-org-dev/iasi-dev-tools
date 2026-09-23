package args

import (
	"path/filepath"
	"testing"
)

func TestPromoteExtractsExplicitContract(t *testing.T) {
	Parms := Parse("promote", []string{"v0.6.0", "iasi-org-dev", "iasi-org"})
	source := filepath.Clean("iasi-org-dev")
	destination := filepath.Clean("iasi-org")

	if Parms.TargetVersion != "v0.6.0" {
		t.Fatalf("TargetVersion = %q, want v0.6.0", Parms.TargetVersion)
	}
	if Parms.SourcePath != source {
		t.Fatalf("SourcePath = %q, want %q", Parms.SourcePath, source)
	}
	if Parms.DestinationPath != destination {
		t.Fatalf("DestinationPath = %q, want %q", Parms.DestinationPath, destination)
	}
	if Parms.SourceOrganization != "" {
		t.Fatalf("SourceOrganization = %q, want empty before repository preparation", Parms.SourceOrganization)
	}
	if Parms.DestinationOrganization != "iasi-org" {
		t.Fatalf("DestinationOrganization = %q, want iasi-org", Parms.DestinationOrganization)
	}
	if Parms.Organization != "" {
		t.Fatalf("Organization = %q, want empty before repository preparation", Parms.Organization)
	}
	if len(Parms.Targets) != 0 {
		t.Fatalf("Targets = %v, want none", Parms.Targets)
	}
}

func TestPromoteAcceptsDotRelativePaths(t *testing.T) {
	plain := Parse("promote", []string{"v0.6.0", "iasi-org-dev", "iasi-org"})
	dotted := Parse("promote", []string{"v0.6.0", "./iasi-org-dev", "./iasi-org"})

	if plain.SourcePath != dotted.SourcePath {
		t.Fatalf("source paths differ: %q != %q", plain.SourcePath, dotted.SourcePath)
	}
	if plain.DestinationPath != dotted.DestinationPath {
		t.Fatalf("destination paths differ: %q != %q", plain.DestinationPath, dotted.DestinationPath)
	}
}

func TestPromoteAcceptsAbsolutePaths(t *testing.T) {
	base := t.TempDir()
	source := filepath.Join(base, "iasi-org-dev")
	destination := filepath.Join(base, "iasi-org")
	Parms := Parse("promote", []string{"v0.6.0", source, destination})

	if Parms.SourcePath != source {
		t.Fatalf("SourcePath = %q, want %q", Parms.SourcePath, source)
	}
	if Parms.DestinationPath != destination {
		t.Fatalf("DestinationPath = %q, want %q", Parms.DestinationPath, destination)
	}
}

func TestPromoteLocalKeepsExplicitContract(t *testing.T) {
	Parms := Parse("promote", []string{"v0.6.0", "iasi-org-dev", "iasi-org", "-l"})

	if !Parms.Local {
		t.Fatal("Local = false, want true")
	}
	if filepath.Base(Parms.SourcePath) != "iasi-org-dev" || filepath.Base(Parms.DestinationPath) != "iasi-org" {
		t.Fatalf("unexpected promote paths: %q -> %q", Parms.SourcePath, Parms.DestinationPath)
	}
}

func TestWorkflowPromoteExtractsNamedContract(t *testing.T) {
	Parms := Parse("workflow", []string{"promote", "--source", "iasi-org-dev", "--dest", "iasi-org", "--version", "v0.7.0"})

	if Parms.Subcommand != "promote" {
		t.Fatalf("Subcommand = %q, want promote", Parms.Subcommand)
	}
	if Parms.NextVersion != "v0.7.0" {
		t.Fatalf("NextVersion = %q, want v0.7.0", Parms.NextVersion)
	}
	if Parms.TargetVersion != "" {
		t.Fatalf("TargetVersion = %q, want empty until the workflow selects the frozen VERSION", Parms.TargetVersion)
	}
	if filepath.Base(Parms.SourcePath) != "iasi-org-dev" || filepath.Base(Parms.DestinationPath) != "iasi-org" {
		t.Fatalf("unexpected promote paths: %q -> %q", Parms.SourcePath, Parms.DestinationPath)
	}
}

func TestWorkflowPromoteRejectsPositionalContract(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("workflow promote positional contract did not fail")
		}
	}()
	Parse("workflow", []string{"promote", "v0.7.0", "iasi-org-dev", "iasi-org"})
}

func TestWorkflowPromoteRequiresEveryNamedParameter(t *testing.T) {
	tests := [][]string{
		{"promote", "--dest", "iasi-org", "--version", "v0.7.0"},
		{"promote", "--source", "iasi-org-dev", "--version", "v0.7.0"},
		{"promote", "--source", "iasi-org-dev", "--dest", "iasi-org"},
	}

	for _, values := range tests {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("workflow promote accepted incomplete parameters: %v", values)
				}
			}()
			Parse("workflow", values)
		}()
	}
}

func TestPromoteRejectsIncompleteContract(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("promote without source and destination did not fail")
		}
	}()
	Parse("promote", []string{"v0.6.0"})
}

func TestPromoteRejectsSameSourceAndDestination(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("promote with same source and destination did not fail")
		}
	}()
	Parse("promote", []string{"v0.6.0", "iasi-org", "./iasi-org"})
}

func TestPromoteRejectsRemovedPushFlag(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("-p did not fail")
		}
	}()
	Parse("promote", []string{"v0.6.0", "iasi-org-dev", "iasi-org", "-p"})
}

func TestRestoreExtractsTargetVersion(t *testing.T) {
	Parms := Parse("restore", []string{"v0.3.0"})

	if Parms.TargetVersion != "v0.3.0" {
		t.Fatalf("TargetVersion = %q, want v0.3.0", Parms.TargetVersion)
	}
	if len(Parms.Targets) != 0 {
		t.Fatalf("Targets = %v, want none", Parms.Targets)
	}
}

func TestRestoreWithoutVersionLeavesTargetVersionEmpty(t *testing.T) {
	Parms := Parse("restore", nil)

	if Parms.TargetVersion != "" {
		t.Fatalf("TargetVersion = %q, want empty", Parms.TargetVersion)
	}
	if len(Parms.Targets) != 0 {
		t.Fatalf("Targets = %v, want none", Parms.Targets)
	}
}

func TestVersionWithoutOperandLeavesTargetVersionEmpty(t *testing.T) {
	Parms := Parse("version", nil)

	if Parms.TargetVersion != "" {
		t.Fatalf("TargetVersion = %q, want empty", Parms.TargetVersion)
	}
}

func TestVersionExtractsExplicitNewVersion(t *testing.T) {
	Parms := Parse("version", []string{"v0.6.0"})

	if Parms.TargetVersion != "v0.6.0" {
		t.Fatalf("TargetVersion = %q, want v0.6.0", Parms.TargetVersion)
	}
	if Parms.Organization != "" {
		t.Fatalf("Organization = %q, want empty", Parms.Organization)
	}
}

func TestFreezeHasNoVersionOperand(t *testing.T) {
	Parms := Parse("freeze", nil)
	if Parms.TargetVersion != "" {
		t.Fatalf("TargetVersion = %q, want empty", Parms.TargetVersion)
	}
}

func TestPromoteCheckUsesSameOperandsAsPromote(t *testing.T) {
	Parms := Parse("promote-check", []string{"v1.2.3", "iasi-org-dev", "iasi-org", "-l"})

	if Parms.TargetVersion != "v1.2.3" {
		t.Fatalf("TargetVersion = %q, want v1.2.3", Parms.TargetVersion)
	}
	if filepath.Base(Parms.SourcePath) != "iasi-org-dev" || filepath.Base(Parms.DestinationPath) != "iasi-org" {
		t.Fatalf("unexpected promote-check paths: %q -> %q", Parms.SourcePath, Parms.DestinationPath)
	}
	if !Parms.Local {
		t.Fatal("Local = false, want true")
	}
}
