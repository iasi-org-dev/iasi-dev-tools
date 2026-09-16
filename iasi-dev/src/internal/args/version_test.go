package args

import "testing"

func TestPromoteExtractsTargetVersion(t *testing.T) {
	Parms := Parse("promote", []string{"v0.6.0", "iasi-quarto"})

	if Parms.TargetVersion != "v0.6.0" {
		t.Fatalf("TargetVersion = %q, want v0.6.0", Parms.TargetVersion)
	}
	if len(Parms.Targets) != 1 || Parms.Targets[0] != "iasi-quarto" {
		t.Fatalf("Targets = %v, want [iasi-quarto]", Parms.Targets)
	}
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

func TestPromoteWorkflowExtractsTargetVersion(t *testing.T) {
	Parms := Parse("workflow", []string{"promote", "v0.7.0", "iasi-quarto"})

	if Parms.Subcommand != "promote" {
		t.Fatalf("Subcommand = %q, want promote", Parms.Subcommand)
	}
	if Parms.TargetVersion != "v0.7.0" {
		t.Fatalf("TargetVersion = %q, want v0.7.0", Parms.TargetVersion)
	}
	if len(Parms.Targets) != 1 || Parms.Targets[0] != "iasi-quarto" {
		t.Fatalf("Targets = %v, want [iasi-quarto]", Parms.Targets)
	}
}

func TestVersionExtractsOrganization(t *testing.T) {
	Parms := Parse("version", []string{"iasi-org"})

	if Parms.Organization != "iasi-org" {
		t.Fatalf("Organization = %q, want iasi-org", Parms.Organization)
	}
	if len(Parms.Targets) != 0 {
		t.Fatalf("Targets = %v, want none", Parms.Targets)
	}
}

func TestPathParameter(t *testing.T) {
	Parms := Parse("version", []string{"--path", "C:/iasi-org", "iasi-org"})

	if Parms.Path != "C:/iasi-org" {
		t.Fatalf("Path = %q, want C:/iasi-org", Parms.Path)
	}
	if Parms.Organization != "iasi-org" {
		t.Fatalf("Organization = %q, want iasi-org", Parms.Organization)
	}
}

func TestPromotePushOnlyHasNoTargetVersion(t *testing.T) {
	Parms := Parse("promote", []string{"-p"})

	if !Parms.Push {
		t.Fatal("Push = false, want true")
	}
	if Parms.TargetVersion != "" {
		t.Fatalf("TargetVersion = %q, want empty", Parms.TargetVersion)
	}
	if len(Parms.Targets) != 0 {
		t.Fatalf("Targets = %v, want none", Parms.Targets)
	}
}

func TestWorkflowPromotePushOnlyHasNoTargetVersion(t *testing.T) {
	Parms := Parse("workflow", []string{"promote", "-p"})

	if Parms.Subcommand != "promote" {
		t.Fatalf("Subcommand = %q, want promote", Parms.Subcommand)
	}
	if !Parms.Push {
		t.Fatal("Push = false, want true")
	}
	if Parms.TargetVersion != "" {
		t.Fatalf("TargetVersion = %q, want empty", Parms.TargetVersion)
	}
}
