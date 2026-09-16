package cli

import (
	"io"
	"os"
	"strings"
	"testing"

	"iasi-dev/internal/structures"
)

func TestIndentSizeIsFour(t *testing.T) {
	if IndentSize != 4 {
		t.Fatalf("IndentSize = %d, want 4", IndentSize)
	}
}

func TestStepAtUsesSpacesAndDepth(t *testing.T) {
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer read.Close()

	old := os.Stdout
	os.Stdout = write
	defer func() { os.Stdout = old }()

	parms := structures.Parms{Verbose: int(visibilityVeryVerbose)}
	StepAt(parms, 2, "Building target")
	if err := write.Close(); err != nil {
		t.Fatal(err)
	}

	data, err := io.ReadAll(read)
	if err != nil {
		t.Fatal(err)
	}
	output := string(data)
	want := strings.Repeat(" ", 2*IndentSize) + "Building target"

	if !strings.Contains(output, want) {
		t.Fatalf("StepAt output = %q, want message containing %q", output, want)
	}
	if strings.Contains(output, "\t") {
		t.Fatalf("StepAt output contains a tab: %q", output)
	}
}
