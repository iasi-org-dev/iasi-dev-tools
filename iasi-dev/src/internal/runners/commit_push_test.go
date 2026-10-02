package runners

import (
	"reflect"
	"testing"
)

func TestPushChangesArgumentsKeepsDevelopmentPushNormal(t *testing.T) {
	got := pushChangesArguments("iasi-org-dev"); want := []string{"push"}
	if !reflect.DeepEqual(got,want){ t.Fatalf("got %v want %v",got,want)}
}
func TestPushChangesArgumentsReplacesStableMain(t *testing.T) {
	got := pushChangesArguments("iasi-org"); want := []string{"push","--force","origin","HEAD:main"}
	if !reflect.DeepEqual(got,want){ t.Fatalf("got %v want %v",got,want)}
}
func TestPushChangesArgumentsDoesNotForceWhenOrganizationIsUnknown(t *testing.T) {
	got := pushChangesArguments(""); want := []string{"push"}
	if !reflect.DeepEqual(got,want){ t.Fatalf("got %v want %v",got,want)}
}
