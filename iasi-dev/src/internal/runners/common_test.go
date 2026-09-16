package runners

import (
	"testing"

	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

func TestHandleRCContinuesForNotice(t *testing.T) {
	rc := RC.OK
	parms := structures.Parms{RC: &rc}

	got := handleRC(&parms, RC.Warning)

	if got != RC.Warning {
		t.Fatalf("handleRC() = 0x%X, want 0x%X", got, RC.Warning)
	}
	if rc != RC.Warning {
		t.Fatalf("accumulated RC = 0x%X, want 0x%X", rc, RC.Warning)
	}
}

func TestHandleRCTolerantSkipsErroneousResult(t *testing.T) {
	rc := RC.OK
	parms := structures.Parms{
		RC:       &rc,
		Tolerant: true,
	}

	got := handleRC(&parms, RC.Error)

	if got != RC.Skip {
		t.Fatalf("handleRC() = 0x%X, want RC.Skip", got)
	}
	if rc != RC.Error {
		t.Fatalf("accumulated RC = 0x%X, want 0x%X", rc, RC.Error)
	}
}

func TestHandleRCNonTolerantAbortsOnErroneousResult(t *testing.T) {
	rc := RC.OK
	parms := structures.Parms{RC: &rc}

	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("handleRC() did not abort")
		}

		stop, ok := recovered.(RC.Stop)
		if !ok {
			t.Fatalf("handleRC() panicked with %T, want RC.Stop", recovered)
		}
		if stop.Code != RC.Error {
			t.Fatalf("RC.Stop.Code = 0x%X, want 0x%X", stop.Code, RC.Error)
		}
		if rc != RC.Error {
			t.Fatalf("accumulated RC = 0x%X, want 0x%X", rc, RC.Error)
		}
	}()

	handleRC(&parms, RC.Error)
}

func TestTargetMessageDepth(t *testing.T) {
	tests := []struct {
		base  int
		depth int
		want  int
	}{
		{0, 0, 1},
		{0, 1, 1},
		{0, 2, 2},
		{1, 0, 2},
		{1, 2, 3},
	}

	for _, test := range tests {
		target := structures.Target{Depth: test.depth}
		if got := targetMessageDepth(test.base, target); got != test.want {
			t.Fatalf("targetMessageDepth(%d, depth=%d) = %d, want %d", test.base, test.depth, got, test.want)
		}
	}
}
