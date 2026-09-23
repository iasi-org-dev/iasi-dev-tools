package runners

import (
	"iasi-dev/internal/cli"
	"iasi-dev/internal/commands"
	"iasi-dev/internal/consts"
	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

// Version shows the current organization version or explicitly sets it.
func Version(Parms *structures.Parms) {
	if Parms.TargetVersion == "" {
		cli.Direct("%s\n", Parms.Version)
		return
	}

	if _, ok := parseSemanticVersion(Parms.TargetVersion); !ok {
		cli.Error(RC.Error, *Parms, "La versión no es válida: %s", Parms.TargetVersion)
	}

	setOrganizationVersionFor(Parms, Parms.Organization, Parms.TargetVersion)
	Parms.Version = Parms.TargetVersion
	cli.Success(*Parms, "Organization version set to %s.", Parms.TargetVersion)
}

func setOrganizationVersionFor(Parms *structures.Parms, organization string, version string) {
	result := commands.Run(
		".",
		Parms.LogFile,
		"gh",
		"variable",
		"set",
		consts.OrganizationVersionVariable,
		"--org",
		organization,
		"--body",
		version,
	)
	if result.RC != RC.OK {
		cli.Error(RC.Error, *Parms, "No se pudo establecer %s de %s a %s.", consts.OrganizationVersionVariable, organization, version)
	}
}
