package main

import (
	"testing"

	"iasi-dev/internal/consts/RC"
)

func TestExternalRCHidesNothingToDoWithoutVeryVerbose(t *testing.T) {
	if got := externalRC(RC.NothingToDo, false); got != RC.OK {
		t.Fatalf("externalRC(1, false) = %d, want 0", got)
	}
}

func TestExternalRCPreservesNothingToDoWithVeryVerbose(t *testing.T) {
	if got := externalRC(RC.NothingToDo, true); got != RC.NothingToDo {
		t.Fatalf("externalRC(1, true) = %d, want 1", got)
	}
}

func TestExternalRCDoesNotRewriteOtherCodes(t *testing.T) {
	cases := []int{
		RC.OK,
		RC.Info | RC.NothingToDo,
		RC.Error | RC.NothingToDo,
		RC.Severe,
		RC.Fatal,
	}

	for _, rc := range cases {
		if got := externalRC(rc, false); got != RC.Result(rc) {
			t.Fatalf("externalRC(0x%X, false) = 0x%X, want 0x%X", rc, got, RC.Result(rc))
		}
	}
}

func TestExternalRCNeverExposesControlFlags(t *testing.T) {
	if got := externalRC(RC.Skip|RC.NothingToDo, false); got != RC.OK {
		t.Fatalf("externalRC(Skip|NothingToDo, false) = 0x%X, want 0x00", got)
	}
}
