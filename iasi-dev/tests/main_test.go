package main

import (
	"testing"

	"iasi-dev/internal/consts/RC"
)

func TestExternalRCHidesOnlyPureNothingToDo(t *testing.T) {
	if got := externalRC(RC.NothingToDo, false); got != RC.OK {
		t.Fatalf("got 0x%X, want 0", got)
	}
	if got := externalRC(RC.Warning|RC.NothingToDo, false); got != RC.Warning|RC.NothingToDo {
		t.Fatalf("got 0x%X, want warning+nothing", got)
	}
}

func TestExternalRCPreservesNothingToDoWithVeryVerbose(t *testing.T) {
	if got := externalRC(RC.NothingToDo, true); got != RC.NothingToDo {
		t.Fatalf("got %d", got)
	}
}
