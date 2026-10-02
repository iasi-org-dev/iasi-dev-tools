package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"

	"iasi-dev/internal/structures"
)

const FileName = "iasi.toml"

var LegacyFileNames = []string{".iasi.yml", "_iasi.yml"}

func Path(path string) string {
	return filepath.Join(path, FileName)
}

func Exists(path string) bool {
	info, err := os.Stat(Path(path))
	return err == nil && !info.IsDir()
}

func Read(path string) (structures.Config, error) {
	result := structures.Config{IASI: map[string]any{}}

	data, err := os.ReadFile(Path(path))
	if err != nil {
		return result, err
	}

	if err := toml.Unmarshal(data, &result.IASI); err != nil {
		return result, fmt.Errorf("%s: %w", Path(path), err)
	}

	return result, nil
}
