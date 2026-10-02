package runners

import (
	"reflect"
	"testing"

	"iasi-dev/internal/structures"
)

func TestRestoreTargetDefaultsToMain(t *testing.T) {
	parms := structures.Parms{}
	if got := restoreTarget(&parms); got != "main" {
		t.Fatalf("got %q", got)
	}
}

func TestRestoreSwitchArguments(t *testing.T) {
	if got := restoreSwitchArguments("main"); !reflect.DeepEqual(got, []string{"switch", "main"}) {
		t.Fatalf("main args = %v", got)
	}
	if got := restoreSwitchArguments("v1.0.0"); !reflect.DeepEqual(got, []string{"switch", "--detach", "v1.0.0"}) {
		t.Fatalf("tag args = %v", got)
	}
}

func TestRestoreRollbackArguments(t *testing.T) {
	if got := restoreRollbackArguments(restoreState{branch: "main"}); !reflect.DeepEqual(got, []string{"switch", "main"}) {
		t.Fatalf("branch rollback = %v", got)
	}
	if got := restoreRollbackArguments(restoreState{commit: "abc123"}); !reflect.DeepEqual(got, []string{"switch", "--detach", "abc123"}) {
		t.Fatalf("commit rollback = %v", got)
	}
}
