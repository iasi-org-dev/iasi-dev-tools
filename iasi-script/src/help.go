package main

import "fmt"

func printHelp() {
	fmt.Print(`IASI Script

Usage:
  iasi-script
  iasi-script [working-dir]
  iasi-script --file <file> [working-dir]
  iasi-script --root <root> [working-dir]
  iasi-script -h

Flags:
  -h                  Show this help and exit.

Parameters:
  --file <file>       Process one TOML configuration file.
  --root <root>       Process all TOML configuration files under root recursively.

Execution:
  working-dir         Working directory. Defaults to the current directory.
  config.toml         Default configuration when --file and --root are not specified.

Paths supplied to --file and --root are resolved from the working directory
unless they are absolute. --file and --root are mutually exclusive.
`)
}
