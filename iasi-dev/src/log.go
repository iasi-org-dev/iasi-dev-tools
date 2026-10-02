package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"iasi-dev/internal/cli"
	"iasi-dev/internal/structures"
)

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

func logParms(Context structures.Context) {
	var output strings.Builder

	fmt.Fprintln(&output, "--- CONTEXT ----------------------------------------------")
	fmt.Fprintf(&output, "Verbose: %d\n", Context.Verbose)
	fmt.Fprintf(&output, "All: %t\n", Context.All)
	fmt.Fprintf(&output, "Checkpoints: %t\n", Context.Checkpoints)
	fmt.Fprintf(&output, "Debug: %t\n", Context.Debug)
	fmt.Fprintf(&output, "PrepareOnly: %t\n", Context.PrepareOnly)
	fmt.Fprintf(&output, "DryRun: %t\n", Context.DryRun)
	fmt.Fprintf(&output, "Force: %t\n", Context.Force)
	fmt.Fprintf(&output, "Help: %t\n", Context.Help)
	fmt.Fprintf(&output, "Install: %t\n", Context.Install)
	fmt.Fprintf(&output, "Local: %t\n", Context.Local)
	fmt.Fprintf(&output, "Tolerant: %t\n", Context.Tolerant)
	fmt.Fprintf(&output, "Message: %q\n", Context.Message)
	fmt.Fprintf(&output, "Format: %q\n", Context.Format)
	writeStringSlice(&output, "Platforms", Context.Platforms)
	fmt.Fprintf(&output, "Path: %q\n", Context.Path)
	fmt.Fprintf(&output, "LogDir: %q\n", Context.LogDir)
	fmt.Fprintf(&output, "Organization: %q\n", Context.Organization)
	fmt.Fprintf(&output, "SourceOrganization: %q\n", Context.SourceOrganization)
	fmt.Fprintf(&output, "DestinationOrganization: %q\n", Context.DestinationOrganization)
	fmt.Fprintf(&output, "SourcePath: %q\n", Context.SourcePath)
	fmt.Fprintf(&output, "DestinationPath: %q\n", Context.DestinationPath)
	fmt.Fprintf(&output, "Version: %q\n", Context.Version)
	fmt.Fprintf(&output, "TargetVersion: %q\n", Context.TargetVersion)
	fmt.Fprintf(&output, "NextVersion: %q\n", Context.NextVersion)
	fmt.Fprintf(&output, "MaterializeDestination: %q\n", Context.MaterializeDestination)
	fmt.Fprintf(&output, "Subcommand: %q\n", Context.Subcommand)
	writeStringSlice(&output, "RequestedTargets", Context.RequestedTargets)
	writeTargets(&output, Context)
	writeConfigs(&output, Context)
	writeStringSlice(&output, "Exclusions", Context.Exclusions)
	writeStringSlice(&output, "Repos", Context.Repos)
	writeStringSlice(&output, "BlackList", Context.BlackList)
	fmt.Fprintln(&output, "----------------------------------------------------------")

	message := output.String()
	if Context.LogFile != nil {
		fmt.Fprint(Context.LogFile, message)
	}
	if Context.PrepareOnly || Context.DryRun {
		cli.Preview("%s", message)
	}
}

func writeTargets(output *strings.Builder, Context structures.Context) {
	fmt.Fprintln(output, "Targets:")
	if len(Context.TargetDetails) == 0 {
		for _, value := range Context.Targets {
			fmt.Fprintf(output, "  %s [none]\n", filepath.Base(value))
		}
		return
	}

	for _, target := range Context.TargetDetails {
		indent := strings.Repeat("  ", target.Depth+1)
		fmt.Fprintf(output, "%s%s [%s]\n", indent, filepath.Base(target.Path), Context.Config(target.Path).Type())
	}
}

func writeConfigs(output *strings.Builder, Context structures.Context) {
	fmt.Fprintln(output, "Configs:")

	paths := make([]string, 0, len(Context.Configs))
	for path := range Context.Configs {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	for _, path := range paths {
		fmt.Fprintf(output, "  %s\n", path)
		writeConfigValue(output, Context.Configs[path].IASI, 2)
	}
}

func writeConfigValue(output *strings.Builder, value any, depth int) {
	indent := strings.Repeat("  ", depth)

	switch current := value.(type) {
	case map[string]any:
		keys := make([]string, 0, len(current))
		for key := range current {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			child := current[key]
			switch child.(type) {
			case map[string]any:
				fmt.Fprintf(output, "%s%s:\n", indent, key)
				writeConfigValue(output, child, depth+1)
			default:
				fmt.Fprintf(output, "%s%s: %s\n", indent, key, formatConfigValue(child))
			}
		}
	default:
		fmt.Fprintf(output, "%s%s\n", indent, formatConfigValue(current))
	}
}

func formatConfigValue(value any) string {
	switch current := value.(type) {
	case string:
		return strconv.Quote(current)
	case []string:
		values := make([]string, len(current))
		for index, item := range current {
			values[index] = strconv.Quote(item)
		}
		return "[" + strings.Join(values, ", ") + "]"
	case []any:
		values := make([]string, len(current))
		for index, item := range current {
			values[index] = formatConfigValue(item)
		}
		return "[" + strings.Join(values, ", ") + "]"
	default:
		return fmt.Sprintf("%v", current)
	}
}

func writeStringSlice(output *strings.Builder, name string, values []string) {
	fmt.Fprintf(output, "%s:\n", name)
	for _, value := range values {
		fmt.Fprintf(output, "  - %s\n", value)
	}
}
