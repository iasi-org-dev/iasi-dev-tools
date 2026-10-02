package main

import (
	"os"
	"path/filepath"
	"strings"

	"iasi-dev/internal/cli"
	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

func preparePath(Context *structures.Context) {
	if Context.Path == "" {
		return
	}

	path, err := filepath.Abs(Context.Path)
	if err != nil {
		cli.Error(RC.Error, *Context, "No se puede resolver --path %q.", Context.Path)
	}

	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		cli.Error(RC.Error, *Context, "--path no existe o no es un directorio: %q", Context.Path)
	}

	if err := os.Chdir(path); err != nil {
		cli.Error(RC.Error, *Context, "No se puede cambiar al directorio indicado por --path: %q", Context.Path)
	}

	Context.Path = filepath.Clean(path)
}

func preparePromotePaths(Context *structures.Context) {
	if Context.SourcePath == "" || Context.DestinationPath == "" {
		return
	}

	source, err := filepath.Abs(Context.SourcePath)
	if err != nil {
		cli.Error(RC.Error, *Context, "No se puede resolver source-path %q.", Context.SourcePath)
	}
	destination, err := filepath.Abs(Context.DestinationPath)
	if err != nil {
		cli.Error(RC.Error, *Context, "No se puede resolver destination-path %q.", Context.DestinationPath)
	}

	source = filepath.Clean(source)
	destination = filepath.Clean(destination)
	if strings.EqualFold(source, destination) {
		cli.Error(RC.Error, *Context, "source-path y destination-path deben ser rutas distintas.")
	}

	Context.SourcePath = source
	Context.DestinationPath = destination
	Context.DestinationOrganization = filepath.Base(destination)
}
