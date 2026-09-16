package main

import "iasi-dev/internal/cli"

func printHelp() {
	cli.Direct(`IASI Dev
	
	Usage:
	  iasi-dev <command> [-h] [-s] [-v|-V] [-m|-M] [-a] [-c] [-d] [-f] [-l] [-p] [-t] [-i] [--path value] [--log directory] [--exclude value[,value]*] [--format value] [--message value] [target...]
	  iasi-dev workflow <build|publish|release> [-h] [-s] [-v|-V] [-m|-M] [-a] [-c] [-d] [-f] [-l] [-t] [-i] [--path value] [--log directory] [--exclude value[,value]*] [--format value] [--message value] [target...]
	  iasi-dev workflow promote vMAJOR.MINOR.PATCH [-l] [-h] [-s] [-v|-V] [-m|-M] [-d] [--path value] [--log directory] [target...]
	  iasi-dev workflow promote -p [-h] [-s] [-v|-V] [-m|-M] [-d] [--path value] [--log directory]
	  iasi-dev promote vMAJOR.MINOR.PATCH [-h] [-s] [-v|-V] [-m|-M] [-d] [--path value] [--log directory] [target...]
	  iasi-dev promote -p [-h] [-s] [-v|-V] [-m|-M] [-d] [--path value] [--log directory]
	  iasi-dev restore [vMAJOR.MINOR.PATCH] [-h] [-s] [-v|-V] [-m|-M] [-d] [--path value] [--log directory] [target...]
	  iasi-dev materialize [-l] <destination> [source] [-h] [-s] [-v|-V] [-m|-M] [-d] [--path value] [--log directory]
	  iasi-dev version [organization] [-h] [-s] [-v|-V] [-m|-M] [-d] [--path value] [--log directory]
	
	Commands:
	  help         Show help
	  build        Build through iasi.quarto
	  publish      Publish through iasi.quarto
	  release      Release through iasi.quarto
	  commit       Commit selected repositories
	  promote      Promote an organization version; with -p, push only
	  restore      Restore repositories to a tagged version; without a version, restore main
	  materialize  Materialize an organization into a destination workspace without Git history
	  workflow     Run a development workflow
	  sync         Sync shared files from iasi-common
	  version      Show VERSION for the explicit or current organization
	
	Workflows:
	  build        Build and commit
	  publish      Publish and commit; optionally include previous stages with -a
	  release      Release and commit; optionally include previous stages with -a
	  promote      Promote the development organization, materialize it locally and push it unless -l; with -p, push only
	
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
	  -l         Keep the operation local; do not publish to GitHub when supported.
	  -m         Show the fully prepared execution and stop before running the command.
	  -M         Dry-run: show the prepared execution and external commands without executing them.
	  -p         Push only; with promote, do not promote or materialize.
	  -t         Continue when an operation fails.
	
	Parameters:
	  --path value               Change to this directory before resolving projects.
	  --exclude value[,value]*  Add exclusions. Existing files are read one exclusion per line; .git, .github and tests are always excluded.
	  --format value            Output format passed to build.
	  --message value           Commit message.
	  --log directory           Directory where execution logs are written.
	`)
}
