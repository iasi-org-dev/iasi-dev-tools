package args

import "testing"

func TestPrepareOnlyFlag(t *testing.T) {
	Parms := Parse("build", []string{"-m"})
	if !Parms.PrepareOnly {
		t.Fatal("PrepareOnly = false, want true")
	}
	if Parms.DryRun {
		t.Fatal("DryRun = true, want false")
	}
}

func TestDryRunFlag(t *testing.T) {
	Parms := Parse("build", []string{"-M"})
	if !Parms.DryRun {
		t.Fatal("DryRun = false, want true")
	}
	if Parms.PrepareOnly {
		t.Fatal("PrepareOnly = true, want false")
	}
}

func TestLogParameter(t *testing.T) {
	Parms := Parse("build", []string{"--log", "trace"})
	if Parms.LogDir != "trace" {
		t.Fatalf("LogDir = %q, want trace", Parms.LogDir)
	}
}

func TestCheckModesAreIncompatible(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("Parse(-m -M) did not abort")
		}
	}()
	Parse("build", []string{"-m", "-M"})
}
