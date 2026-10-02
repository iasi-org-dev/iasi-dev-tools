package RC

import "testing"

func TestResultUsesOnlyLowByte(t *testing.T) {
	if got := Result(Skip | Warning | NothingToDo); got != Warning|NothingToDo {
		t.Fatalf("Result() = 0x%X, want 0x%X", got, Warning|NothingToDo)
	}
}

func TestAddDoesNotAccumulateControlFlags(t *testing.T) {
	current := OK
	Add(&current, Skip|Attention|NothingToDo)

	if current != Attention|NothingToDo {
		t.Fatalf("current = 0x%X, want 0x%X", current, Attention|NothingToDo)
	}
}

func TestValueNeverExposesControlFlags(t *testing.T) {
	current := Skip | Severe | NothingToDo
	if got := Value(&current); got != Severe|NothingToDo {
		t.Fatalf("Value() = 0x%X, want 0x%X", got, Severe|NothingToDo)
	}
}

func TestSkipIsOutsideResultByte(t *testing.T) {
	if Skip <= ResultMask {
		t.Fatalf("Skip = 0x%X, must be above 0x%X", Skip, ResultMask)
	}
	if Result(Skip) != OK {
		t.Fatalf("Result(Skip) = 0x%X, want 0x00", Result(Skip))
	}
}
