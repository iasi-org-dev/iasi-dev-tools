package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

type messageLevel int
type messageVisibility int

const (
	levelVerbose messageLevel = iota
	levelHeader
	levelInfo
	levelSuccess
	levelWarning
	levelError
)

const (
	visibilityNormal messageVisibility = 1 << iota
	visibilityVerbose
	visibilityVeryVerbose
)

const IndentSize = 4

const (
	logBanner   = "============================================================"
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorBlue   = "\033[38;2;0;0;255m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[38;2;255;0;0m"
)

func Direct(format string, args ...any) {
	fmt.Fprintf(os.Stdout, format, args...)
}

// Preview writes preparation and dry-run information to the console in blue.
func Preview(format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	lineBreak := strings.HasSuffix(message, "\n")
	message = strings.TrimSuffix(message, "\n")
	fmt.Fprintf(os.Stdout, "%s%s%s", colorBlue, message, colorReset)
	if lineBreak {
		fmt.Fprintln(os.Stdout)
	}
}

func Verbose(Parms structures.Parms, format string, args ...any) {
	writeMessage(Parms, os.Stdout, visibilityVerbose, levelVerbose, false, format, args...)
}

func VeryVerbose(Parms structures.Parms, format string, args ...any) {
	writeMessage(Parms, os.Stdout, visibilityVeryVerbose, levelSuccess, false, format, args...)
}

func Info(Parms structures.Parms, format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	info(Parms, visibilityNormal, message)
}

// Header writes a highlighted command or workflow header.
// In the log, the header is surrounded by a simple timestamped banner.
func Header(Parms structures.Parms, format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	writeLogMessage(Parms, logBanner)
	writeLogMessage(Parms, message)
	writeLogMessage(Parms, logBanner)

	if !messageVisible(Parms, visibilityNormal) {
		return
	}

	writeConsoleMessage(os.Stdout, levelHeader, true, message)
}

// Step writes a one-level indented Info message only in very verbose mode.
func Step(Parms structures.Parms, format string, args ...any) {
	StepAt(Parms, 1, format, args...)
}

// StepAt writes an Info message at the requested indentation depth.
// Each level is exactly IndentSize spaces.
func StepAt(Parms structures.Parms, depth int, format string, args ...any) {
	if depth < 0 {
		depth = 0
	}
	message := strings.Repeat(" ", depth*IndentSize) + fmt.Sprintf(format, args...)
	info(Parms, visibilityVeryVerbose, message)
}

func info(Parms structures.Parms, visibility messageVisibility, message string) {
	writeMessage(Parms, os.Stdout, visibility, levelInfo, true, "%s", message)
}

func Success(Parms structures.Parms, format string, args ...any) {
	writeMessage(Parms, os.Stdout, visibilityNormal, levelSuccess, false, format, args...)
}

func Warning(Parms structures.Parms, format string, args ...any) {
	writeMessage(Parms, os.Stderr, visibilityNormal, levelWarning, false, format, args...)
}

// ErrorMessage writes a non-terminal error message.
// The log entry is explicitly prefixed with ERROR while the console keeps the normal message.
func ErrorMessage(Parms structures.Parms, format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	writeLogMessage(Parms, "ERROR: "+message)

	if !messageVisible(Parms, visibilityNormal) {
		return
	}

	writeConsoleMessage(os.Stderr, levelError, false, message)
}

// Error writes an error message and aborts execution.
// main owns the final os.Exit().
func Error(rc int, Parms structures.Parms, format string, args ...any) {
	ErrorMessage(Parms, format, args...)
	Abort(rc, Parms)
}

// Abort records the supplied RC bits and aborts execution without printing another message.
func Abort(rc int, Parms structures.Parms) {
	code := RC.Add(Parms.RC, rc)
	panic(RC.Stop{Code: code})
}

func writeMessage(Parms structures.Parms, writer io.Writer, visibility messageVisibility, level messageLevel, bold bool, format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	writeLogMessage(Parms, message)

	if !messageVisible(Parms, visibility) {
		return
	}

	writeConsoleMessage(writer, level, bold, message)
}

func writeConsoleMessage(writer io.Writer, level messageLevel, bold bool, message string) {
	style := messageColor(level)
	if bold {
		style = colorBold + style
	}

	fmt.Fprintf(writer, "%s - %s%s%s\n", time.Now().Format("15:04:05"), style, message, colorReset)
}

func writeLogMessage(Parms structures.Parms, message string) {
	if Parms.LogFile == nil {
		return
	}
	fmt.Fprintf(Parms.LogFile, "%s - %s\n", time.Now().Format("15:04:05"), message)
}

func messageVisible(Parms structures.Parms, visibility messageVisibility) bool {
	return Parms.Verbose&int(visibility) != 0
}

func messageColor(level messageLevel) string {
	switch level {
	case levelVerbose:
		return colorBlue
	case levelHeader:
		return colorCyan
	case levelInfo:
		return colorWhite
	case levelSuccess:
		return colorGreen
	case levelWarning:
		return colorYellow
	case levelError:
		return colorRed
	default:
		return colorReset
	}
}
