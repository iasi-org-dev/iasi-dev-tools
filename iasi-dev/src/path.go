package main

import (
	"os"
	"path/filepath"

	"iasi-dev/internal/cli"
	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

// preparePath changes the process working directory when --path was provided.
func preparePath(Parms *structures.Parms) {
	if Parms.Path == "" {
		return
	}

	path, err := filepath.Abs(Parms.Path)
	if err != nil {
		cli.Error(RC.Error, *Parms, "No se puede resolver --path %q.", Parms.Path)
	}

	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		cli.Error(RC.Error, *Parms, "--path no existe o no es un directorio: %q", Parms.Path)
	}

	if err := os.Chdir(path); err != nil {
		cli.Error(RC.Error, *Parms, "No se puede cambiar al directorio indicado por --path: %q", Parms.Path)
	}

	Parms.Path = filepath.Clean(path)
}
