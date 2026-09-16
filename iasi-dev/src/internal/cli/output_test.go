package cli

import (
	"io"
	"os"
	"strings"
	"testing"

	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

func TestRedIsPureRed(t *testing.T) {
	if colorRed != "\033[38;2;255;0;0m" {
		t.Fatalf("colorRed = %q, want pure RGB red", colorRed)
	}
}

func TestBlueIsPureRGBBlue(t *testing.T) {
	if colorBlue != "\033[38;2;0;0;255m" {
		t.Fatalf("colorBlue = %q, want pure RGB blue", colorBlue)
	}
}

func TestPreviewIsBlueNotBold(t *testing.T) {
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer read.Close()

	old := os.Stdout
	os.Stdout = write
	defer func() { os.Stdout = old }()

	Preview("preview\n")
	if err := write.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(read)
	if err != nil {
		t.Fatal(err)
	}
	output := string(data)
	if !strings.Contains(output, colorBlue+"preview"+colorReset) {
		t.Fatalf("preview output is not blue: %q", output)
	}
	if strings.Contains(output, colorBold+colorBlue+"preview") {
		t.Fatalf("preview output is still bold blue: %q", output)
	}
}

func TestVerboseMessageIsBlueNotBold(t *testing.T) {
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer read.Close()

	old := os.Stdout
	os.Stdout = write
	defer func() { os.Stdout = old }()

	parms := structures.Parms{Verbose: int(visibilityVerbose)}
	Verbose(parms, "processing")
	if err := write.Close(); err != nil {
		t.Fatal(err)
	}

	data, err := io.ReadAll(read)
	if err != nil {
		t.Fatal(err)
	}
	output := string(data)
	if !strings.Contains(output, colorBlue+"processing"+colorReset) {
		t.Fatalf("verbose output is not blue: %q", output)
	}
	if strings.Contains(output, colorBold+colorBlue+"processing") {
		t.Fatalf("verbose output is still bold blue: %q", output)
	}
}

func TestErrorMessageIsRedNotBold(t *testing.T) {
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer read.Close()

	old := os.Stderr
	os.Stderr = write
	defer func() { os.Stderr = old }()

	parms := structures.Parms{Verbose: int(visibilityNormal)}
	ErrorMessage(parms, "boom")
	if err := write.Close(); err != nil {
		t.Fatal(err)
	}

	data, err := io.ReadAll(read)
	if err != nil {
		t.Fatal(err)
	}
	output := string(data)

	if !strings.Contains(output, colorRed+"boom"+colorReset) {
		t.Fatalf("error output is not red: %q", output)
	}
	if strings.Contains(output, colorBold+colorRed+"boom") {
		t.Fatalf("error output is still bold red: %q", output)
	}
}

func TestHeaderWritesTimestampedBannerToLog(t *testing.T) {
	log, err := os.CreateTemp(t.TempDir(), "iasi-log-*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()

	parms := structures.Parms{Verbose: 0, LogFile: log}
	Header(parms, "Releasing %s", "iasi-quarto")

	if err := log.Sync(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(log.Name())
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 3 {
		t.Fatalf("header log has %d lines, want 3: %q", len(lines), string(data))
	}

	for i, line := range lines {
		if len(line) < 11 || line[2] != ':' || line[5] != ':' || line[8:11] != " - " {
			t.Fatalf("line %d has no timestamp prefix: %q", i+1, line)
		}
	}

	if !strings.HasSuffix(lines[0], logBanner) {
		t.Fatalf("first banner line missing: %q", lines[0])
	}
	if !strings.HasSuffix(lines[1], "Releasing iasi-quarto") {
		t.Fatalf("header line missing: %q", lines[1])
	}
	if !strings.HasSuffix(lines[2], logBanner) {
		t.Fatalf("last banner line missing: %q", lines[2])
	}
}

func TestHeaderMessageIsBoldCyan(t *testing.T) {
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer read.Close()

	message := "Releasing iasi-quarto"
	writeConsoleMessage(write, levelHeader, true, message)
	if err := write.Close(); err != nil {
		t.Fatal(err)
	}

	data, err := io.ReadAll(read)
	if err != nil {
		t.Fatal(err)
	}
	output := string(data)

	if !strings.Contains(output, colorBold+colorCyan+message+colorReset) {
		t.Fatalf("header output is not bold cyan: %q", output)
	}
}

func TestErrorMessageDoesNotAbort(t *testing.T) {
	rc := RC.OK
	parms := structures.Parms{RC: &rc, Verbose: int(visibilityNormal)}

	completed := false
	func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				t.Fatalf("ErrorMessage aborted: %v", recovered)
			}
		}()
		ErrorMessage(parms, "rollback warning")
		completed = true
	}()

	if !completed {
		t.Fatal("ErrorMessage did not return normally")
	}
	if rc != RC.OK {
		t.Fatalf("rc = 0x%02X, want 0x%02X", rc, RC.OK)
	}
}

func TestErrorStopsAfterReporting(t *testing.T) {
	rc := RC.OK
	parms := structures.Parms{RC: &rc, Verbose: int(visibilityNormal)}

	defer func() {
		recovered := recover()
		stop, ok := recovered.(RC.Stop)
		if !ok {
			t.Fatalf("Error recovered %T, want RC.Stop", recovered)
		}
		if stop.Code != RC.Error {
			t.Fatalf("stop code = 0x%02X, want 0x%02X", stop.Code, RC.Error)
		}
		if rc != RC.Error {
			t.Fatalf("rc = 0x%02X, want 0x%02X", rc, RC.Error)
		}
	}()

	Error(RC.Error, parms, "boom")
}

func TestAbortStopsWithoutAdditionalMessage(t *testing.T) {
	rc := RC.OK
	parms := structures.Parms{RC: &rc}

	defer func() {
		recovered := recover()
		stop, ok := recovered.(RC.Stop)
		if !ok {
			t.Fatalf("Abort recovered %T, want RC.Stop", recovered)
		}
		if stop.Code != RC.Error {
			t.Fatalf("stop code = 0x%02X, want 0x%02X", stop.Code, RC.Error)
		}
		if rc != RC.Error {
			t.Fatalf("rc = 0x%02X, want 0x%02X", rc, RC.Error)
		}
	}()

	Abort(RC.Error, parms)
}
