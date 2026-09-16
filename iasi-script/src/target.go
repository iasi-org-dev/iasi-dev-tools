package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"iasi-script/internal/consts/EXT"
)

func artifactName(workDir, configuredName string) string {
	name := strings.TrimSpace(configuredName)
	if name != "" {
		return name
	}
	return filepath.Base(filepath.Clean(workDir))
}

// targetExtension maps the logical target name to the materialized file
// extension. Targets are semantic names, not extensions.
func targetExtension(target string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(target)) {
	case "powershell":
		return EXT.PowerShell, nil
	case "bash":
		return EXT.Bash, nil
	default:
		return "", fmt.Errorf("unsupported target %q", target)
	}
}

func resolveDirectory(workDir, configured, fallback string) string {
	value := strings.TrimSpace(configured)
	if value == "" {
		value = fallback
	}
	if filepath.IsAbs(value) {
		return filepath.Clean(value)
	}
	return filepath.Clean(filepath.Join(workDir, value))
}

func resolveInputDir(workDir string, descriptor Descriptor) string {
	return resolveDirectory(workDir, descriptor.InputDir, ".")
}

func resolveOutputDir(workDir string, descriptor Descriptor) string {
	return resolveDirectory(workDir, descriptor.OutputDir, "_outputs")
}

func resolveOutput(outputDir, name, target string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("artifact name is empty")
	}

	extension, err := targetExtension(target)
	if err != nil {
		return "", err
	}

	return filepath.Join(outputDir, name+"."+extension), nil
}
