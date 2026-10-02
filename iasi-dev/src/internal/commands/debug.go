package commands

var debug bool
var dryRun bool

func SetDebug(value bool) {
	debug = value
}

func SetDryRun(value bool) {
	dryRun = value
}
