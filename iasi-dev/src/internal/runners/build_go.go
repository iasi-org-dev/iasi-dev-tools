package runners

import (
	"os"
	"path/filepath"

	"iasi-dev/internal/commands"
	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

func buildGo(target structures.Target, Context structures.Context) int {
	config := Context.Config(target.Path)
	source := targetPath(target.Path, config.SoftwareInputDir())
	output := targetPath(target.Path, config.SoftwareOutputDir())

	if !Context.DryRun {
		if err := os.MkdirAll(output, 0755); err != nil {
			return RC.Error
		}
	}

	rc := RC.OK
	for _, platform := range Context.Platforms {
		var current int
		switch platform {
		case "windows":
			current = buildGoWindows(target, Context, source, output)
		case "linux":
			current = buildGoLinux(target, Context, source, output)
		default:
			current = RC.Error
		}
		rc |= RC.Result(current)
	}
	return rc
}

func buildGoWindows(target structures.Target, Context structures.Context, source string, output string) int {
	name := Context.Config(target.Path).SoftwareName(defaultSoftwareName(filepath.Base(target.Path)))
	destination := filepath.Join(output, name+".exe")
	return runGoBuild(Context, source, destination, "windows")
}

func buildGoLinux(target structures.Target, Context structures.Context, source string, output string) int {
	name := Context.Config(target.Path).SoftwareName(defaultSoftwareName(filepath.Base(target.Path)))
	destination := filepath.Join(output, name)
	return runGoBuild(Context, source, destination, "linux")
}

func runGoBuild(Context structures.Context, source string, destination string, platform string) int {
	environment := []string{"GOOS=" + platform, "GOARCH=amd64", "CGO_ENABLED=0"}
	result := commands.RunLoggedEnv(source, Context.LogFile, environment, "go", "build", "-o", destination, ".")
	return result.RC
}

func targetPath(base string, value string) string {
	if filepath.IsAbs(value) {
		return filepath.Clean(value)
	}
	return filepath.Clean(filepath.Join(base, value))
}


func defaultSoftwareName(name string) string {
	original := name
	i := 0

	for i < len(name) && name[i] >= '0' && name[i] <= '9' {
		i++
	}
	if i == 0 {
		return name
	}

	for i < len(name) && (name[i] == '-' || name[i] == '_' || name[i] == '.' || name[i] == ' ') {
		i++
	}
	if i >= len(name) {
		return original
	}

	return name[i:]
}
