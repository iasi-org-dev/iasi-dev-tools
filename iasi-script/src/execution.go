package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"iasi-script/internal/parms"
)

type Environment struct {
	CurrentDir string
}

type Execution struct {
	WorkingDir  string
	ConfigPaths []string
}

func prepareEnvironment() (Environment, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return Environment{}, fmt.Errorf("cannot determine current directory: %w", err)
	}

	currentDir, err = filepath.Abs(currentDir)
	if err != nil {
		return Environment{}, fmt.Errorf("cannot resolve current directory: %w", err)
	}

	return Environment{CurrentDir: filepath.Clean(currentDir)}, nil
}

func prepareExecution(Parms parms.Parms, environment Environment) (Execution, error) {
	workingDir := strings.TrimSpace(Parms.WorkingDir)
	if workingDir == "" {
		workingDir = environment.CurrentDir
	} else if !filepath.IsAbs(workingDir) {
		workingDir = filepath.Join(environment.CurrentDir, workingDir)
	}
	workingDir = filepath.Clean(workingDir)

	info, err := os.Stat(workingDir)
	if err != nil {
		return Execution{}, fmt.Errorf("cannot inspect working directory %s: %w", workingDir, err)
	}
	if !info.IsDir() {
		return Execution{}, fmt.Errorf("working directory is not a directory: %s", workingDir)
	}

	configPaths, err := resolveConfigPaths(Parms, workingDir)
	if err != nil {
		return Execution{}, err
	}

	return Execution{
		WorkingDir:  workingDir,
		ConfigPaths: configPaths,
	}, nil
}

func resolveConfigPaths(Parms parms.Parms, workingDir string) ([]string, error) {
	if strings.TrimSpace(Parms.File) != "" {
		path := resolveExecutionPath(workingDir, Parms.File)
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("cannot inspect configuration file %s: %w", path, err)
		}
		if info.IsDir() {
			return nil, fmt.Errorf("configuration file is a directory: %s", path)
		}
		return []string{path}, nil
	}

	if strings.TrimSpace(Parms.Root) != "" {
		root := resolveExecutionPath(workingDir, Parms.Root)
		return findConfigPaths(root)
	}

	path := filepath.Join(workingDir, defaultConfigName)
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("cannot inspect configuration file %s: %w", path, err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("configuration file is a directory: %s", path)
	}
	return []string{path}, nil
}

func resolveExecutionPath(workingDir, value string) string {
	value = strings.TrimSpace(value)
	if filepath.IsAbs(value) {
		return filepath.Clean(value)
	}
	return filepath.Clean(filepath.Join(workingDir, value))
}

func findConfigPaths(root string) ([]string, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("cannot inspect configuration root %s: %w", root, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("configuration root is not a directory: %s", root)
	}

	var paths []string
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if strings.EqualFold(filepath.Ext(entry.Name()), ".toml") {
			paths = append(paths, filepath.Clean(path))
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("cannot scan configuration root %s: %w", root, err)
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("no TOML configuration files found under %s", root)
	}

	sort.Strings(paths)
	return paths, nil
}
