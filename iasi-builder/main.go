// Command iasi-builder materializes a template using TOML configuration.
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
    workDir, err := resolveWorkDir(args)
    if err != nil { return err }

    configPath, err := findConfig(workDir)
    if err != nil { return err }

    config, err := loadConfig(configPath)
    if err != nil { return err }

    return build(workDir, config)
}
