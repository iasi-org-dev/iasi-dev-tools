package runners

import (
	"os"
	"path/filepath"

	"iasi-dev/internal/cli"
	"iasi-dev/internal/commands"
	"iasi-dev/internal/consts"
	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

// Version shows the current organization version or explicitly sets it.
// A successful execution always materializes the effective version in <root>/VERSION.
func Version(Parms *structures.Parms) {
	if Parms.TargetVersion == "" {
		cli.Direct("%s\n", Parms.Version)
		writeWorkspaceVersion(Parms)
		return
	}

	if _, ok := parseSemanticVersion(Parms.TargetVersion); !ok {
		cli.Error(RC.Error, *Parms, "La versión no es válida: %s", Parms.TargetVersion)
	}

	setOrganizationVersionFor(Parms, Parms.Organization, Parms.TargetVersion)
	Parms.Version = Parms.TargetVersion
	writeWorkspaceVersion(Parms)
	cli.Success(*Parms, "Organization version set to %s.", Parms.TargetVersion)
}

func writeWorkspaceVersion(Parms *structures.Parms) {
	if Parms.DryRun {
		cli.Preview("Write %s\n", filepath.Join(Parms.Root, consts.OrganizationVersionVariable))
		return
	}
	if Parms.Root == "" {
		cli.Error(RC.Error, *Parms, "No se ha determinado la raíz del workspace.")
	}

	path := filepath.Join(Parms.Root, consts.OrganizationVersionVariable)
	if err := os.WriteFile(path, []byte(Parms.Version+"\n"), 0644); err != nil {
		cli.Error(RC.Error, *Parms, "No se pudo escribir %s.", path)
	}
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
