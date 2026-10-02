package commands

import (
	"fmt"
	"os"

	"iasi-dev/internal/structures"
)

func RunLogged(directory string, logFile *os.File, name string, args ...string) structures.Result {
	if debug {
		fmt.Printf("RunLogged: directory=%s name=%s args=%v\n", directory, name, args)
	}
	return commandLogged(directory, false, logFile, name, args...)
}

func RunFriendlyLogged(directory string, logFile *os.File, name string, args ...string) structures.Result {
	if debug {
		fmt.Printf("RunFriendlyLogged: directory=%s name=%s args=%v\n", directory, name, args)
	}
	return commandLogged(directory, true, logFile, name, args...)
}

func RunLoggedEnv(directory string, logFile *os.File, environment []string, name string, args ...string) structures.Result {
	if debug {
		fmt.Printf("RunLoggedEnv: directory=%s env=%v name=%s args=%v\n", directory, environment, name, args)
	}
	return commandLoggedEnv(directory, false, logFile, environment, name, args...)
}
