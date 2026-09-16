// Command iasi-script materializes templates from the current working directory.
package main

import (
	"fmt"
	"os"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("usage: iasi-script")
	}

	dir, err := workDir()
	if err != nil {
		return err
	}

	descriptor, err := loadDescriptor(dir)
	if err != nil {
		return err
	}

	configPath := resolveConfigPath(dir, descriptor)
	config, err := loadConfig(configPath)
	if err != nil {
		return err
	}

	return build(dir, descriptor, config)
}
