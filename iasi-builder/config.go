package main

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/pelletier/go-toml/v2"
)

type Config struct {
    Template string
    Data     map[string]any
}

func resolveWorkDir(args []string) (string, error) {
    if len(args) > 1 { return "", fmt.Errorf("usage: iasi-builder [path]") }
    if len(args) == 0 { return os.Getwd() }
    return filepath.Abs(args[0])
}

func findConfig(workDir string) (string, error) {
    matches, err := filepath.Glob(filepath.Join(workDir, "*.toml"))
    if err != nil { return "", err }
    if len(matches) != 1 {
        return "", fmt.Errorf("expected exactly one TOML configuration in %s, found %d", workDir, len(matches))
    }
    return matches[0], nil
}

func loadConfig(path string) (Config, error) {
    content, err := os.ReadFile(path)
    if err != nil { return Config{}, fmt.Errorf("cannot read %s: %w", path, err) }

    data := map[string]any{}
    if err := toml.Unmarshal(content, &data); err != nil {
        return Config{}, fmt.Errorf("invalid TOML %s: %w", path, err)
    }

    templateName, _ := data["template"].(string)
    if templateName == "" { return Config{}, fmt.Errorf("template is required in %s", path) }

    return Config{Template: templateName, Data: data}, nil
}
