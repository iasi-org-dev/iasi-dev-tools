package runners

import (
	"os"
	"path/filepath"

	"iasi-dev/internal/commands"
	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

func buildGo(target structures.Target, Parms structures.Parms) int {
	source := targetPath(target.Path, target.SourceDir)
	output := targetPath(target.Path, target.OutputDir)

	if !Parms.DryRun {
		if err := os.MkdirAll(output, 0755); err != nil {
			return RC.Error
		}
	}

	rc := RC.OK
	for _, platform := range Parms.Platforms {
		var current int
		switch platform {
		case "windows":
			current = buildGoWindows(target, Parms, source, output)
		case "linux":
			current = buildGoLinux(target, Parms, source, output)
		default:
			current = RC.Error
		}
		rc |= RC.Result(current)
	}
	return rc
}

func buildGoWindows(target structures.Target, Parms structures.Parms, source string, output string) int {
	destination := filepath.Join(output, target.Name+".exe")
	return runGoBuild(Parms, source, destination, "windows")
}

func buildGoLinux(target structures.Target, Parms structures.Parms, source string, output string) int {
	destination := filepath.Join(output, target.Name)
	return runGoBuild(Parms, source, destination, "linux")
}

func runGoBuild(Parms structures.Parms, source string, destination string, platform string) int {
	environment := []string{"GOOS=" + platform, "GOARCH=amd64", "CGO_ENABLED=0"}
	result := commands.RunLoggedEnv(source, Parms.LogFile, environment, "go", "build", "-o", destination, ".")
	return result.RC
}

func targetPath(base string, value string) string {
	if filepath.IsAbs(value) {
		return filepath.Clean(value)
	}
	return filepath.Clean(filepath.Join(base, value))
}
