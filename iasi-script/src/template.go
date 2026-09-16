package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// resolveTemplate resolves one logical template basename for a target.
// Target-specific templates take precedence over the generic fallback:
//
//	<template>.<target-extension>.tpl
//	<template>.tpl
func resolveTemplate(inputDir, templateName, target string) (string, error) {
	templateName = strings.TrimSpace(templateName)
	if templateName == "" {
		return "", fmt.Errorf("template name is empty")
	}
	if strings.HasSuffix(strings.ToLower(templateName), ".tpl") {
		return "", fmt.Errorf("template %q must be a basename without .tpl", templateName)
	}

	target = strings.ToLower(strings.TrimSpace(target))
	if target == "" {
		return "", fmt.Errorf("target is required to resolve template %q", templateName)
	}

	extension, err := targetExtension(target)
	if err != nil {
		return "", err
	}

	base := templateName
	if filepath.IsAbs(base) {
		base = filepath.Clean(base)
	} else {
		base = filepath.Join(inputDir, base)
	}

	candidates := []string{
		base + "." + extension + ".tpl",
		base + ".tpl",
	}
	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err == nil {
			if info.IsDir() {
				continue
			}
			return candidate, nil
		}
		if !os.IsNotExist(err) {
			return "", fmt.Errorf("cannot inspect template %s: %w", candidate, err)
		}
	}

	return "", fmt.Errorf(
		"template %q for target %q not found; tried %s and %s",
		templateName,
		target,
		candidates[0],
		candidates[1],
	)
}
