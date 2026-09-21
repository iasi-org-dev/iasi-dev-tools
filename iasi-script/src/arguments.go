package main

import (
	"fmt"
	"strings"

	"iasi-script/internal/parms"
)

func processArguments(args []string) (parms.Parms, error) {
	var Parms parms.Parms

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "-h" {
			Parms.Help = true
			continue
		}

		if strings.HasPrefix(arg, "--") {
			switch arg {
			case "--file":
				if Parms.File != "" {
					return parms.Parms{}, fmt.Errorf("--file can only be specified once")
				}
				if i+1 >= len(args) || strings.HasPrefix(args[i+1], "-") {
					return parms.Parms{}, fmt.Errorf("--file requires a value")
				}
				i++
				Parms.File = strings.TrimSpace(args[i])
				if Parms.File == "" {
					return parms.Parms{}, fmt.Errorf("--file requires a value")
				}

			case "--root":
				if Parms.Root != "" {
					return parms.Parms{}, fmt.Errorf("--root can only be specified once")
				}
				if i+1 >= len(args) || strings.HasPrefix(args[i+1], "-") {
					return parms.Parms{}, fmt.Errorf("--root requires a value")
				}
				i++
				Parms.Root = strings.TrimSpace(args[i])
				if Parms.Root == "" {
					return parms.Parms{}, fmt.Errorf("--root requires a value")
				}

			default:
				return parms.Parms{}, fmt.Errorf("unknown parameter %s", arg)
			}
			continue
		}

		if strings.HasPrefix(arg, "-") {
			return parms.Parms{}, fmt.Errorf("unknown flag %s", arg)
		}

		if i != len(args)-1 {
			return parms.Parms{}, fmt.Errorf("working directory must be the last argument")
		}
		if Parms.WorkingDir != "" {
			return parms.Parms{}, fmt.Errorf("working directory can only be specified once")
		}
		Parms.WorkingDir = arg
	}

	if Parms.File != "" && Parms.Root != "" {
		return parms.Parms{}, fmt.Errorf("--file and --root cannot be used together")
	}

	return Parms, nil
}
