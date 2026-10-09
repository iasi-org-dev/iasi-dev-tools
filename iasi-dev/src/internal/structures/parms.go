package structures

import "os"

type Diagnostic struct {
	Path string
	Message string
	Emitted bool
}

type Context struct {
	Diagnostics []Diagnostic
	Verbose                int
	All                    bool
	Checkpoints            bool
	Debug                  bool
	PrepareOnly            bool
	DryRun                 bool
	Force                  bool
	Help                   bool
	Install                bool
	Local                  bool
	Tolerant               bool
	Message                string
	Format                 string
	Platforms              []string
	Path                   string
	LogDir                 string
	Organization           string
	SourceOrganization      string
	DestinationOrganization string
	SourcePath              string
	DestinationPath         string
	Version                string
	TargetVersion          string
	NextVersion            string
	MaterializeDestination string
	Subcommand             string
	LogFile                *os.File
	RC                     *int
	LastRC                 int
	Targets                []string
	TargetDetails          []Target
	RequestedTargets       []string
	Exclusions             []string
	Repos                  []string
	BlackList              []string
	Configs                map[string]Config
}

type Parms = Context
