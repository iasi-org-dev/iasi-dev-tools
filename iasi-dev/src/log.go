package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"iasi-dev/internal/cli"
	"iasi-dev/internal/structures"
)

// createLogFile creates and keeps open the log for the complete execution.
func createLogFile(command string, requestedDir string) (*os.File, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	logDir := requestedDir
	if logDir == "" {
		logDir = filepath.Join(cwd, "logs")
	} else if !filepath.IsAbs(logDir) {
		logDir = filepath.Join(cwd, logDir)
	}
	logDir = filepath.Clean(logDir)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}

	path := filepath.Join(logDir, fmt.Sprintf("iasi-%s-%s.log", command, time.Now().Format("20060102150405")))
	return os.Create(path)
}

// logParms records the fully prepared execution parameters and discovered lists.
// -m and -M mirror the same information to the console; the log is always written.
func logParms(Parms structures.Parms) {
	var output strings.Builder

	fmt.Fprintln(&output, "--- PARMS ------------------------------------------------")
	fmt.Fprintf(&output, "Verbose: %d\n", Parms.Verbose)
	fmt.Fprintf(&output, "All: %t\n", Parms.All)
	fmt.Fprintf(&output, "Checkpoints: %t\n", Parms.Checkpoints)
	fmt.Fprintf(&output, "Debug: %t\n", Parms.Debug)
	fmt.Fprintf(&output, "PrepareOnly: %t\n", Parms.PrepareOnly)
	fmt.Fprintf(&output, "DryRun: %t\n", Parms.DryRun)
	fmt.Fprintf(&output, "Force: %t\n", Parms.Force)
	fmt.Fprintf(&output, "Help: %t\n", Parms.Help)
	fmt.Fprintf(&output, "Install: %t\n", Parms.Install)
	fmt.Fprintf(&output, "Local: %t\n", Parms.Local)
	fmt.Fprintf(&output, "Tolerant: %t\n", Parms.Tolerant)
	fmt.Fprintf(&output, "Message: %q\n", Parms.Message)
	fmt.Fprintf(&output, "Format: %q\n", Parms.Format)
	writeStringSlice(&output, "Platforms", Parms.Platforms)
	fmt.Fprintf(&output, "Path: %q\n", Parms.Path)
	fmt.Fprintf(&output, "LogDir: %q\n", Parms.LogDir)
	fmt.Fprintf(&output, "Organization: %q\n", Parms.Organization)
	fmt.Fprintf(&output, "SourceOrganization: %q\n", Parms.SourceOrganization)
	fmt.Fprintf(&output, "DestinationOrganization: %q\n", Parms.DestinationOrganization)
	fmt.Fprintf(&output, "SourcePath: %q\n", Parms.SourcePath)
	fmt.Fprintf(&output, "DestinationPath: %q\n", Parms.DestinationPath)
	fmt.Fprintf(&output, "Version: %q\n", Parms.Version)
	fmt.Fprintf(&output, "TargetVersion: %q\n", Parms.TargetVersion)
	fmt.Fprintf(&output, "NextVersion: %q\n", Parms.NextVersion)
	fmt.Fprintf(&output, "MaterializeDestination: %q\n", Parms.MaterializeDestination)
	fmt.Fprintf(&output, "Subcommand: %q\n", Parms.Subcommand)
	writeStringSlice(&output, "RequestedTargets", Parms.RequestedTargets)
	writeTargets(&output, Parms)
	writeStringSlice(&output, "Exclusions", Parms.Exclusions)
	writeStringSlice(&output, "Repos", Parms.Repos)
	writeStringSlice(&output, "BlackList", Parms.BlackList)
	fmt.Fprintln(&output, "----------------------------------------------------------")

	message := output.String()
	if Parms.LogFile != nil {
		fmt.Fprint(Parms.LogFile, message)
	}
	if Parms.PrepareOnly || Parms.DryRun {
		cli.Preview("%s", message)
	}
}

func writeTargets(output *strings.Builder, Parms structures.Parms) {
	fmt.Fprintln(output, "Targets:")
	if len(Parms.TargetDetails) == 0 {
		for _, value := range Parms.Targets {
			fmt.Fprintf(output, "  %s [none]\n", filepath.Base(value))
		}
		return
	}

	for _, target := range Parms.TargetDetails {
		indent := strings.Repeat("  ", target.Depth+1)
		fmt.Fprintf(output, "%s%s [%s]\n", indent, filepath.Base(target.Path), target.Type)
	}
}

func writeStringSlice(output *strings.Builder, name string, values []string) {
	fmt.Fprintf(output, "%s:\n", name)
	for _, value := range values {
		fmt.Fprintf(output, "  - %s\n", value)
	}
}
