package runners

import (
	"reflect"
	"testing"
)

func TestPushChangesArgumentsKeepsDevelopmentPushNormal(t *testing.T) {
	got := pushChangesArguments("iasi-org-dev")
	want := []string{"push"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("pushChangesArguments(iasi-org-dev) = %v, want %v", got, want)
	}
}

func TestPushChangesArgumentsReplacesStableMain(t *testing.T) {
	got := pushChangesArguments("iasi-org")
	want := []string{"push", "--force", "origin", "HEAD:main"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("pushChangesArguments(iasi-org) = %v, want %v", got, want)
	}
}

func TestPushChangesArgumentsDoesNotForceWhenOrganizationIsUnknown(t *testing.T) {
	got := pushChangesArguments("")
	want := []string{"push"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("pushChangesArguments(empty) = %v, want %v", got, want)
	}
}
