package tools

import (
	"os"
	"path/filepath"
)

func IsRepo(directory string) bool {
	_, err := os.Stat(filepath.Join(directory, ".git"))
	return err == nil
}

func FindRepo(directory string) string {
	path, err := filepath.Abs(directory)
	if err != nil {
		return ""
	}

	for {
		if IsRepo(path) {
			return filepath.Clean(path)
		}

		parent := filepath.Dir(path)
		if parent == path {
			return ""
		}
		path = parent
	}
}
