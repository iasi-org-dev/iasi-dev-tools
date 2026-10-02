package args

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestMaterializeDefaultsSourceToCurrentDirectory(t *testing.T) {
	Parms := Parse("materialize", []string{"target-org"})

	wantDestination, err := filepath.Abs("target-org")
	if err != nil {
		t.Fatal(err)
	}
	if Parms.MaterializeDestination != filepath.Clean(wantDestination) {
		t.Fatalf("MaterializeDestination = %q, want %q", Parms.MaterializeDestination, filepath.Clean(wantDestination))
	}
	if len(Parms.Targets) != 0 {
		t.Fatalf("Targets = %v, want empty so common preparation uses current directory", Parms.Targets)
	}
	if !reflect.DeepEqual(Parms.RequestedTargets, []string{"target-org"}) {
		t.Fatalf("RequestedTargets = %v, want [target-org]", Parms.RequestedTargets)
	}
}

func TestMaterializeUsesOptionalSourceAsNormalTarget(t *testing.T) {
	Parms := Parse("materialize", []string{"target-org", "source-org"})

	if !reflect.DeepEqual(Parms.Targets, []string{"source-org"}) {
		t.Fatalf("Targets = %v, want [source-org]", Parms.Targets)
	}
	if !reflect.DeepEqual(Parms.RequestedTargets, []string{"target-org", "source-org"}) {
		t.Fatalf("RequestedTargets = %v, want [target-org source-org]", Parms.RequestedTargets)
	}
}
