package commands

var debug bool
var dryRun bool

// SetDebug enables or disables temporary debug messages.
func SetDebug(value bool) {
	debug = value
}

// SetDryRun enables or disables command execution while preserving command logging.
func SetDryRun(value bool) {
	dryRun = value
}
