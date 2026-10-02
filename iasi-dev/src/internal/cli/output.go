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

func Preview(format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	lineBreak := strings.HasSuffix(message, "\n")
	message = strings.TrimSuffix(message, "\n")
	fmt.Fprintf(os.Stdout, "%s%s%s", colorBlue, message, colorReset)
	if lineBreak {
		fmt.Fprintln(os.Stdout)
	}
}

func Verbose(Context structures.Context, format string, args ...any) {
	writeMessage(Context, os.Stdout, visibilityVerbose, levelVerbose, false, format, args...)
}

func VeryVerbose(Context structures.Context, format string, args ...any) {
	writeMessage(Context, os.Stdout, visibilityVeryVerbose, levelSuccess, false, format, args...)
}

func Info(Context structures.Context, format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	info(Context, visibilityNormal, message)
}

func Header(Context structures.Context, format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	writeLogMessage(Context, logBanner)
	writeLogMessage(Context, message)
	writeLogMessage(Context, logBanner)

	if !messageVisible(Context, visibilityNormal) {
		return
	}

	writeConsoleMessage(os.Stdout, levelHeader, true, message)
}

func Step(Context structures.Context, format string, args ...any) {
	StepAt(Context, 1, format, args...)
}

func StepAt(Context structures.Context, depth int, format string, args ...any) {
	if depth < 0 {
		depth = 0
	}
	message := strings.Repeat(" ", depth*IndentSize) + fmt.Sprintf(format, args...)
	info(Context, visibilityVeryVerbose, message)
}

func info(Context structures.Context, visibility messageVisibility, message string) {
	writeMessage(Context, os.Stdout, visibility, levelInfo, true, "%s", message)
}

func Success(Context structures.Context, format string, args ...any) {
	writeMessage(Context, os.Stdout, visibilityNormal, levelSuccess, false, format, args...)
}

func Warning(Context structures.Context, format string, args ...any) {
	RC.Add(Context.RC, RC.Warning)
	writeMessage(Context, os.Stderr, visibilityNormal, levelWarning, false, format, args...)
}

func Attention(Context structures.Context, format string, args ...any) {
	RC.Add(Context.RC, RC.Attention)
	writeMessage(Context, os.Stderr, visibilityNormal, levelWarning, false, "ATTENTION: "+format, args...)
}

func ErrorMessage(Context structures.Context, format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	writeLogMessage(Context, "ERROR: "+message)

	if !messageVisible(Context, visibilityNormal) {
		return
	}

	writeConsoleMessage(os.Stderr, levelError, false, message)
}

func Error(rc int, Context structures.Context, format string, args ...any) {
	ErrorMessage(Context, format, args...)
	Abort(rc, Context)
}

func Abort(rc int, Context structures.Context) {
	code := RC.Add(Context.RC, rc)
	panic(RC.Stop{Code: code})
}

func writeMessage(Context structures.Context, writer io.Writer, visibility messageVisibility, level messageLevel, bold bool, format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	writeLogMessage(Context, message)

	if !messageVisible(Context, visibility) {
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

func writeLogMessage(Context structures.Context, message string) {
	if Context.LogFile == nil {
		return
	}
	fmt.Fprintf(Context.LogFile, "%s - %s\n", time.Now().Format("15:04:05"), message)
}

func messageVisible(Context structures.Context, visibility messageVisibility) bool {
	return Context.Verbose&int(visibility) != 0
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
