// Command iasi-dev is the IASI development command-line entry point.
package main

import (
	"os"

	"iasi-dev/internal/args"
	"iasi-dev/internal/cli"
	"iasi-dev/internal/commands"
	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/runners"
)

func main() {
	os.Exit(run())
}

func run() (exitCode int) {
	exitCode = RC.OK
	veryVerbose := false

	defer func() {
		if recovered := recover(); recovered != nil {
			switch stop := recovered.(type) {
			case RC.Stop:
				exitCode = stop.Code
			default:
				panic(recovered)
			}
		}

		exitCode = externalRC(exitCode, veryVerbose)
	}()

	if len(os.Args) == 1 {
		printHelp()
		return RC.NothingToDo
	}

	command := os.Args[1]
	Parms := args.Parse(command, os.Args[2:])
	veryVerbose = Parms.Verbose == 7
	commands.SetDebug(Parms.Debug)
	preparePath(&Parms)

	if command == "help" {
		Parms.Help = true
	}

	if Parms.Message == "" {
		Parms.Message = command
	}

	logFile, err := createLogFile(command, Parms.LogDir)
	if err != nil {
		cli.Error(RC.Error, Parms, "No se pudo crear el log: %v", err)
	}
	Parms.LogFile = logFile
	defer Parms.LogFile.Close()

	prepareParms(&Parms)
	logParms(Parms)

	if Parms.PrepareOnly {
		return RC.Value(Parms.RC)
	}

	if Parms.Help {
		printHelp()
		RC.Add(Parms.RC, RC.NothingToDo)
		return RC.Value(Parms.RC)
	}

	switch command {
	case "build":
		runners.Build(&Parms, 0)
	case "publish":
		runners.Publish(&Parms, 0)
	case "commit":
		runners.Commit(&Parms, 0)
	case "release":
		runners.Release(&Parms, 0)
	case "workflow":
		runners.Workflow(&Parms)
	case "sync":
		runners.Sync(&Parms)
	case "version":
		runners.Version(&Parms)
	case "promote":
		runners.Promote(&Parms)
	case "restore":
		runners.Restore(&Parms)
	case "materialize":
		runners.Materialize(&Parms)
	default:
		cli.Error(RC.Error, Parms, "Comando desconocido: %q", command)
	}

	return RC.Value(Parms.RC)
}

// externalRC adapts the internal IASI return code to the process exit code.
// NothingToDo is reported as success unless very-verbose mode (-V) is active.
func externalRC(rc int, veryVerbose bool) int {
	rc = RC.Result(rc)
	if rc == RC.NothingToDo && !veryVerbose {
		return RC.OK
	}
	return rc
}
