package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

const (
	defaultDescriptorName = "iasi.yml"
	legacyDescriptorName  = ".iasi.yml"
	defaultConfigName     = "config.toml"
)

type Descriptor struct {
	Name      string
	InputDir  string
	OutputDir string
	Targets   []string
}

type Config struct {
	Template  string            `toml:"template"`
	Templates map[string]string `toml:"templates"`
	Data      map[string]any    `toml:"-"`
}

func descriptorPath(workDir string) (string, error) {
	for _, name := range []string{defaultDescriptorName, legacyDescriptorName} {
		path := filepath.Join(workDir, name)
		info, err := os.Stat(path)
		if err == nil && !info.IsDir() {
			return path, nil
		}
		if err != nil && !os.IsNotExist(err) {
			return "", fmt.Errorf("cannot inspect %s: %w", path, err)
		}
	}
	return "", fmt.Errorf("cannot find %s in %s", defaultDescriptorName, workDir)
}

// loadDescriptor reads the small subset of the current YAML descriptor needed
// by iasi-script. The descriptor will migrate to TOML later, so this parser is
// intentionally limited to top-level scalars plus the top-level targets list.
func loadDescriptor(workDir string) (Descriptor, error) {
	path, err := descriptorPath(workDir)
	if err != nil {
		return Descriptor{}, err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return Descriptor{}, fmt.Errorf("cannot read %s: %w", path, err)
	}

	var descriptor Descriptor
	inTargets := false
	for lineNumber, raw := range strings.Split(string(content), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if inTargets && strings.HasPrefix(line, "-") {
			value := cleanYAMLValue(strings.TrimSpace(strings.TrimPrefix(line, "-")))
			if value == "" {
				return Descriptor{}, fmt.Errorf("invalid empty target in %s:%d", path, lineNumber+1)
			}
			descriptor.Targets = append(descriptor.Targets, value)
			continue
		}
		inTargets = false

		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = cleanYAMLValue(value)

		switch key {
		case "name":
			descriptor.Name = value
		case "input-dir":
			descriptor.InputDir = value
		case "output-dir":
			descriptor.OutputDir = value
		case "targets":
			inTargets = true
			if value != "" {
				return Descriptor{}, fmt.Errorf("inline targets are not supported yet in %s:%d", path, lineNumber+1)
			}
		}
	}

	return descriptor, nil
}

func cleanYAMLValue(value string) string {
	value = strings.TrimSpace(value)
	if i := strings.Index(value, " #"); i >= 0 {
		value = strings.TrimSpace(value[:i])
	}
	return strings.Trim(value, "\"'")
}

func loadConfig(path string) (Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("cannot read %s: %w", path, err)
	}

	data := map[string]any{}
	if err := toml.Unmarshal(content, &data); err != nil {
		return Config{}, fmt.Errorf("invalid TOML %s: %w", path, err)
	}

	var config Config
	if err := toml.Unmarshal(content, &config); err != nil {
		return Config{}, fmt.Errorf("invalid TOML %s: %w", path, err)
	}
	if strings.TrimSpace(config.Template) != "" && len(config.Templates) > 0 {
		return Config{}, fmt.Errorf("%s cannot define both template and templates", path)
	}

	// template/templates describe how to materialize the artifact; the rest of
	// config.toml is template data.
	delete(data, "template")
	delete(data, "templates")
	config.Data = data
	return config, nil
}
