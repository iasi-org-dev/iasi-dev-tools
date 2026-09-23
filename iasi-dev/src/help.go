package main

import "iasi-dev/internal/cli"

func printHelp() {
	cli.Direct(`IASI Dev
	
	Usage:
	  iasi-dev <command> [-h] [-s] [-v|-V] [-m|-M] [-a] [-c] [-d] [-f] [-l] [-t] [-i] [--path value] [--log directory] [--exclude value[,value]*] [--format value] [--platform value] [--message value] [target...]
	  iasi-dev workflow <build|publish|release> [-h] [-s] [-v|-V] [-m|-M] [-a] [-c] [-d] [-f] [-l] [-t] [-i] [--path value] [--log directory] [--exclude value[,value]*] [--format value] [--platform value] [--message value] [target...]
	  iasi-dev workflow promote --source <source-path> --dest <destination-path> --version vMAJOR.MINOR.PATCH [-l] [-h] [-s] [-v|-V] [-m|-M] [-d] [--path value] [--log directory]
	  iasi-dev freeze [-h] [-s] [-v|-V] [-m|-M] [-d] [--path value] [--log directory]
	  iasi-dev promote vMAJOR.MINOR.PATCH <source-path> <destination-path> [-l] [-h] [-s] [-v|-V] [-m|-M] [-d] [--path value] [--log directory]
	  iasi-dev promote-check vMAJOR.MINOR.PATCH <source-path> <destination-path> [-l] [-h] [-s] [-v|-V] [-m|-M] [-d] [--path value] [--log directory]
	  iasi-dev restore [vMAJOR.MINOR.PATCH] [-h] [-s] [-v|-V] [-m|-M] [-d] [--path value] [--log directory] [target...]
	  iasi-dev materialize [-l] <destination> [source] [-h] [-s] [-v|-V] [-m|-M] [-d] [--path value] [--log directory]
	  iasi-dev version [vMAJOR.MINOR.PATCH] [-h] [-s] [-v|-V] [-m|-M] [-d] [--path value] [--log directory]
	
	Commands:
	  help         Show help
	  build        Build discovered IASI targets through their configured builders
	  publish      Publish through iasi.quarto
	  release      Release through iasi.quarto
	  commit       Commit selected repositories
	  freeze       Freeze the complete development organization at its current VERSION using published Git tags
	  promote      Promote an explicit frozen version between explicit source and destination paths without development history
	  promote-check Scan an explicit promotion source for suspicious references without modifying files
	  restore      Restore repositories to a tagged version; without a version, restore main
	  materialize  Materialize an organization into a destination workspace without Git history
	  workflow     Run a development workflow
	  sync         Sync shared files from iasi-common
	  version      Show VERSION or explicitly set the organization version
	
	Workflows:
	  build        Build and commit
	  publish      Publish and commit; optionally include previous stages with -a
	  release      Release and commit; optionally include previous stages with -a
	  promote      Freeze current VERSION, advance development to --version, then promote the frozen VERSION and publish unless -l
	
	Options:
	  -h         Show help.
	  -s         Silent output.
	  -v         Verbose output.
	  -V         Very verbose output; preserves process RC 1 (NothingToDo).
	  -a         Execute previous workflow stages too.
	  -c         Use intermediate workflows as checkpoints.
	  -d         Show debug messages.
	  -f         Force the operation when supported.
	  -i         Install the artifact when applicable.
	  -l         Keep the operation local when supported. Freeze does not support -l and always publishes its tags.
	  -m         Show the fully prepared execution and stop before running the command.
	  -M         Dry-run: show the prepared execution and external commands without executing them.
	  -t         Continue when an operation fails.
	
	Parameters:
	  --path value               Change to this directory before resolving projects and relative paths.
	  --exclude value[,value]*  Add exclusions. Existing files are read one exclusion per line; .git, .github and tests are always excluded.
	  --format value            Output format passed to build.
	  --platform value          Build only the selected platform: windows or linux. Without it, build prepares both.
	  --message value           Commit message.
	  --log directory           Directory where execution logs are written.
	  --source source-path       Source workspace for workflow promote.
	  --dest destination-path    Destination workspace for workflow promote.
	  --version vX.Y.Z           Next development VERSION for workflow promote; the current VERSION is the one promoted.
	`)
}
