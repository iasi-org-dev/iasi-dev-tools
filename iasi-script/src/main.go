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
	Parms, err := processArguments(args)
	if err != nil {
		return err
	}

	if Parms.Help {
		printHelp()
		return nil
	}

	environment, err := prepareEnvironment()
	if err != nil {
		return err
	}

	execution, err := prepareExecution(Parms, environment)
	if err != nil {
		return err
	}

	descriptor, err := loadDescriptor(execution.WorkingDir)
	if err != nil {
		return err
	}

	for _, configPath := range execution.ConfigPaths {
		config, err := loadConfig(configPath)
		if err != nil {
			return err
		}

		if err := build(execution.WorkingDir, descriptor, config); err != nil {
			return err
		}
	}

	return nil
}
