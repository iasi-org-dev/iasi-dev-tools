package commands

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"iasi-dev/internal/cli"
	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

// Run executes a command using its specific implementation when available.
func Run(directory string, logFile *os.File, name string, args ...string) structures.Result {
	if debug {
		fmt.Printf("Run: directory=%s name=%s args=%v\n", directory, name, args)
	}
	switch name {
	case "git":
		return commandGit(directory, false, logFile, args...)
	default:
		return command(directory, false, logFile, name, args...)
	}
}

// RunFriendly executes a command producing friendly output when supported.
func RunFriendly(directory string, logFile *os.File, name string, args ...string) structures.Result {
	if debug {
		fmt.Printf("RunFriendly: directory=%s name=%s args=%v\n", directory, name, args)
	}
	switch name {
	case "git":
		return commandGit(directory, true, logFile, args...)
	default:
		return command(directory, true, logFile, name, args...)
	}
}

// RunDirect executes a command without using a specific implementation.
func RunDirect(directory string, logFile *os.File, name string, args ...string) structures.Result {
	if debug {
		fmt.Printf("RunDirect: directory=%s name=%s args=%v\n", directory, name, args)
	}
	return command(directory, false, logFile, name, args...)
}

// RunDirectLogged executes a direct command and writes its output to the log.
func RunDirectLogged(directory string, logFile *os.File, name string, args ...string) structures.Result {
	if debug {
		fmt.Printf("RunDirectLogged: directory=%s name=%s args=%v\n", directory, name, args)
	}
	return commandLogged(directory, false, logFile, name, args...)
}

// command executes a generic command, logs it and captures its output.
func command(directory string, friendly bool, logFile *os.File, name string, args ...string) structures.Result {
	if debug {
		fmt.Printf("command: directory=%s friendly=%t name=%s args=%v\n", directory, friendly, name, args)
	}

	if name == "Rscript" && len(args) >= 2 && args[0] == "-e" {
		args = append([]string{}, args...)
		args[1] = "options(warn = 1); " + args[1]
	}

	writeCommand(logFile, directory, name, args...)
	if dryRun {
		return structures.Result{RC: RC.OK}
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.Command(name, args...)
	cmd.Dir = directory
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := structures.Result{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		RC:     RC.OK,
	}

	if err != nil {
		result.RC = RC.Error
		if name == "Rscript" {
			if exitError, ok := err.(*exec.ExitError); ok {
				code := exitError.ExitCode()
				if code >= 0 && code <= 0xFF {
					result.RC = code
				}
			}
		}
	}

	return result
}

// commandLogged executes a generic command, logs it and writes its output directly to the log.
func commandLogged(directory string, friendly bool, logFile *os.File, name string, args ...string) structures.Result {
	return commandLoggedEnv(directory, friendly, logFile, nil, name, args...)
}

func commandLoggedEnv(directory string, friendly bool, logFile *os.File, environment []string, name string, args ...string) structures.Result {
	if debug {
		fmt.Printf("commandLoggedEnv: directory=%s friendly=%t env=%v name=%s args=%v\n", directory, friendly, environment, name, args)
	}

	if name == "Rscript" && len(args) >= 2 && args[0] == "-e" {
		args = append([]string{}, args...)
		args[1] = "options(warn = 1); " + args[1]
	}

	writeCommandEnv(logFile, directory, environment, name, args...)
	if dryRun {
		return structures.Result{RC: RC.OK}
	}

	cmd := exec.Command(name, args...)
	cmd.Dir = directory
	if len(environment) != 0 {
		cmd.Env = append(os.Environ(), environment...)
	}
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	err := cmd.Run()

	result := structures.Result{RC: RC.OK}
	if err != nil {
		result.RC = RC.Error
		if name == "Rscript" {
			if exitError, ok := err.(*exec.ExitError); ok {
				code := exitError.ExitCode()
				if code >= 0 && code <= 0xFF {
					result.RC = code
				}
			}
		}
	}

	return result
}

// writeCommand writes the complete command to the log and mirrors it during dry-run.
func writeCommand(logFile *os.File, directory string, name string, args ...string) {
	writeCommandEnv(logFile, directory, nil, name, args...)
}

func writeCommandEnv(logFile *os.File, directory string, environment []string, name string, args ...string) {
	commandLine := name
	if len(args) != 0 {
		commandLine += " " + formatCommandArguments(args)
	}
	if len(environment) != 0 {
		commandLine = strings.Join(environment, " ") + " " + commandLine
	}
	if logFile != nil {
		fmt.Fprintf(logFile, "%s - Command [%s]: %s\n", time.Now().Format("15:04:05"), directory, commandLine)
	}
	if dryRun {
		cli.Preview("Command [%s]: %s\n", directory, commandLine)
	}
}

func formatCommandArguments(args []string) string {
	formatted := make([]string, 0, len(args))
	for _, arg := range args {
		if strings.ContainsAny(arg, " \t\"") {
			formatted = append(formatted, strconv.Quote(arg))
			continue
		}
		formatted = append(formatted, arg)
	}
	return strings.Join(formatted, " ")
}
