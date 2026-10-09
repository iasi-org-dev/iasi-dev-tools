// Command iasi-dev is the IASI development command-line entry point.
package main

import (
	"os"

	"iasi-dev/internal/args"
	"iasi-dev/internal/cli"
	"iasi-dev/internal/commands"
	"iasi-dev/internal/consts/RC"
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

	if command == "-h" || command == "--help" {
		printHelp()
		return RC.NothingToDo
	}

	entry, valid := Registry[command]
	if !valid {
		cli.Direct("Comando desconocido: %q\n", command)
		return RC.Error
	}

	Context := args.Parse(command, os.Args[2:])
	veryVerbose = Context.Verbose == 7
	commands.SetDebug(Context.Debug)
	preparePath(&Context)

	if Context.Message == "" { Context.Message = command }

	logFile, err := createLogFile(command, Context.LogDir)
	if err != nil { cli.Error(RC.Error, Context, "No se pudo crear el log: %v", err) }
	Context.LogFile = logFile
	defer Context.LogFile.Close()

	prepareParms(command, &Context)
	logParms(Context)

	if Context.PrepareOnly { return RC.Value(Context.RC) }

	if Context.Help {
		printHelp()
		RC.Add(Context.RC, RC.NothingToDo)
		return RC.Value(Context.RC)
	}

	entry.Run(&Context)
	args.EmitRemainingDiagnostics(Context)

	return RC.Value(Context.RC)
}

func externalRC(rc int, veryVerbose bool) int {
	rc = RC.Result(rc)
	if rc == RC.NothingToDo && !veryVerbose {
		return RC.OK
	}
	return rc
}
