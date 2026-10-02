package runners

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"iasi-dev/internal/commands"
	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

func TestDispatchPublishAndReleaseSupportWebsite(t *testing.T) {
	commands.SetDryRun(true)
	defer commands.SetDryRun(false)

	for _, test := range []struct {
		name string
		run  func(structures.Target, structures.Parms, int) int
		want string
	}{
		{"publish", dispatchPublish, "iasi::publish("},
		{"release", dispatchRelease, "iasi::release()"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			logPath := filepath.Join(root, test.name+".log")
			logFile, err := os.Create(logPath)
			if err != nil {
				t.Fatal(err)
			}

			target := structures.Target{Path: root, Type: "website"}
			parms := structures.Parms{DryRun: true, LogFile: logFile}

			if rc := test.run(target, parms, 0); rc != RC.OK {
				_ = logFile.Close()
				t.Fatalf("%s RC = %d, want %d", test.name, rc, RC.OK)
			}
			if err := logFile.Close(); err != nil {
				t.Fatal(err)
			}

			data, err := os.ReadFile(logPath)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(data), test.want) {
				t.Fatalf("%s backend missing: %q", test.name, string(data))
			}
		})
	}
}

func TestDispatchLifecycleNothingToDoIsSilent(t *testing.T) {
	for _, test := range []struct {
		name string
		run  func(structures.Target, structures.Parms, int) int
	}{
		{"publish", dispatchPublish},
		{"release", dispatchRelease},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			logPath := filepath.Join(root, test.name+".log")
			logFile, err := os.Create(logPath)
			if err != nil {
				t.Fatal(err)
			}

			target := structures.Target{Path: root, Type: "software"}
			parms := structures.Parms{LogFile: logFile}

			if rc := test.run(target, parms, 0); rc != RC.NothingToDo {
				_ = logFile.Close()
				t.Fatalf("%s RC = %d, want %d", test.name, rc, RC.NothingToDo)
			}
			if err := logFile.Close(); err != nil {
				t.Fatal(err)
			}

			data, err := os.ReadFile(logPath)
			if err != nil {
				t.Fatal(err)
			}
			if strings.TrimSpace(string(data)) != "" {
				t.Fatalf("%s NothingToDo wrote to the log: %q", test.name, string(data))
			}
		})
	}
}
