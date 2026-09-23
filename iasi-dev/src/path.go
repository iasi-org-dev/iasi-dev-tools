package main

import (
	"os"
	"path/filepath"
	"strings"

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

// preparePromotePaths resolves explicit promote operands against the effective working directory.
// No workspace discovery or sibling inference is performed.
func preparePromotePaths(Parms *structures.Parms) {
	if Parms.SourcePath == "" || Parms.DestinationPath == "" {
		return
	}

	source, err := filepath.Abs(Parms.SourcePath)
	if err != nil {
		cli.Error(RC.Error, *Parms, "No se puede resolver source-path %q.", Parms.SourcePath)
	}
	destination, err := filepath.Abs(Parms.DestinationPath)
	if err != nil {
		cli.Error(RC.Error, *Parms, "No se puede resolver destination-path %q.", Parms.DestinationPath)
	}

	source = filepath.Clean(source)
	destination = filepath.Clean(destination)
	if strings.EqualFold(source, destination) {
		cli.Error(RC.Error, *Parms, "source-path y destination-path deben ser rutas distintas.")
	}

	Parms.SourcePath = source
	Parms.DestinationPath = destination
	Parms.DestinationOrganization = filepath.Base(destination)
}
