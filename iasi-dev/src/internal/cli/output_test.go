package cli

import (
	"testing"

	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

func TestWarningAccumulatesRC(t *testing.T) {
	rc := RC.OK
	context := structures.Context{RC: &rc}
	Warning(context, "warning")
	if rc != RC.Warning {
		t.Fatalf("RC = 0x%X, want 0x%X", rc, RC.Warning)
	}
}

func TestAttentionAccumulatesRC(t *testing.T) {
	rc := RC.Warning
	context := structures.Context{RC: &rc}
	Attention(context, "attention")
	if rc != RC.Warning|RC.Attention {
		t.Fatalf("RC = 0x%X, want 0x%X", rc, RC.Warning|RC.Attention)
	}
}

func TestErrorMessageDoesNotAccumulateOrAbort(t *testing.T) {
	rc := RC.OK
	context := structures.Context{RC: &rc}
	ErrorMessage(context, "diagnostic")
	if rc != RC.OK {
		t.Fatalf("RC = 0x%X, want 0", rc)
	}
}

func TestAbortAccumulatesAndStops(t *testing.T) {
	rc := RC.Warning
	context := structures.Context{RC: &rc}

	defer func() {
		recovered := recover()
		stop, ok := recovered.(RC.Stop)
		if !ok {
			t.Fatalf("recovered %T, want RC.Stop", recovered)
		}
		if stop.Code != RC.Warning|RC.Error {
			t.Fatalf("stop code = 0x%X", stop.Code)
		}
	}()

	Abort(RC.Error, context)
}
