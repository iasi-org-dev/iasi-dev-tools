package structures

import "os"

type Context struct {
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
