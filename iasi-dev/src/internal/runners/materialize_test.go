package runners

import (
	"path/filepath"
	"testing"
)

func TestMaterializeNamingHelpers(t *testing.T) {
	destination := filepath.Join("work", "iasi-org")
	if got := materializeOrganization(destination); got != "iasi-org" {
		t.Fatalf("organization = %q", got)
	}
	if filepath.Base(materializeTemporary(destination)) != ".iasi-org.materialize.tmp" {
		t.Fatalf("unexpected temporary path")
	}
	if filepath.Base(materializeBackup(destination)) != ".iasi-org.materialize.bak" {
		t.Fatalf("unexpected backup path")
	}
}

func TestMaterializeConfirmed(t *testing.T) {
	for _, value := range []string{"s", "S", "si", "sí", "y", "YES"} {
		if !materializeConfirmed(value) {
			t.Fatalf("%q should be accepted", value)
		}
	}
	for _, value := range []string{"", "n", "no", "other"} {
		if materializeConfirmed(value) {
			t.Fatalf("%q should be rejected", value)
		}
	}
}
