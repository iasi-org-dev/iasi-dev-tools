package runners

import (
	"iasi-dev/internal/cli"
	"iasi-dev/internal/commands"
	"iasi-dev/internal/consts"
	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
 "strings"
)

func Version(Parms *structures.Parms) {
	if Parms.TargetVersion == "" {
		cli.Direct("%s\n", Parms.Version)
		return
	}

	if _, ok := parseSemanticVersion(Parms.TargetVersion); !ok {
		cli.Error(RC.Error, *Parms, "La versión no es válida: %s", Parms.TargetVersion)
	}

	// The previous VERSION may carry the old v prefix during migration.
 if Parms.Version != "" {
  old,ok:=parseLegacySemanticVersion(Parms.Version)
  if !ok {cli.Error(RC.Error,*Parms,"La versión actual no es válida: %s",Parms.Version)}
  next,_:=parseSemanticVersion(Parms.TargetVersion)
  // One-time migration: vX.Y.Z -> X.Y.Z is not a downgrade.
  migration := strings.HasPrefix(Parms.Version,"v") && compareSemanticVersions(next,old)==0
  if compareSemanticVersions(next,old)<=0 && !migration {
   cli.Error(RC.Error,*Parms,"La nueva versión %s debe ser superior a la actual %s.",Parms.TargetVersion,Parms.Version)
  }
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
