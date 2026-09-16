package runners

import (
	"iasi-dev/internal/cli"
	"iasi-dev/internal/structures"
)

// Version prints the current organization version loaded during common preparation.
func Version(Parms *structures.Parms) {
	cli.Direct("%s\n", Parms.Version)
}
