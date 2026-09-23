package runners

import (
	"fmt"

	"iasi-dev/internal/cli"
	"iasi-dev/internal/commands"
	"iasi-dev/internal/consts"
	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

// Version shows the current organization version or explicitly sets a new one.
// Version selection is never inferred or incremented automatically.
func Version(Parms *structures.Parms) {
	if Parms.TargetVersion == "" {
		cli.Direct("%s\n", Parms.Version)
		return
	}

	current, ok := parseSemanticVersion(Parms.Version)
	if !ok {
		cli.Error(RC.Error, *Parms, "La versión actual no es válida: %s", Parms.Version)
	}
	target, ok := parseSemanticVersion(Parms.TargetVersion)
	if !ok {
		cli.Error(RC.Error, *Parms, "La nueva versión no es válida: %s", Parms.TargetVersion)
	}
	if compareSemanticVersions(target, current) <= 0 {
		cli.Error(RC.Error, *Parms, "La nueva versión %s debe ser superior a %s.", Parms.TargetVersion, Parms.Version)
	}

	setOrganizationVersion(Parms, Parms.TargetVersion)
	Parms.Version = Parms.TargetVersion
	cli.Success(*Parms, "Organization version set to %s.", Parms.TargetVersion)
}

func setOrganizationVersion(Parms *structures.Parms, version string) {
	endpoint := fmt.Sprintf("orgs/%s/actions/variables/%s", Parms.Organization, consts.OrganizationVersionVariable)
	result := commands.Run(
		".",
		Parms.LogFile,
		"gh",
		"api",
		"--method",
		"PATCH",
		endpoint,
		"-f",
		"value="+version,
	)
	if result.RC != RC.OK {
		cli.Error(RC.Error, *Parms, "No se pudo actualizar %s de %s a %s.", consts.OrganizationVersionVariable, Parms.Organization, version)
	}
}
