package runners

import (
	"reflect"
	"testing"
)

func TestPromoteRemoteReconciliation(t *testing.T) {
	repositories := []string{
		"C:/work/iasi-org/repo-a",
		"C:/work/iasi-org/repo-b",
	}
	remote := []string{"repo-b", "repo-c"}

	plan := promoteRemoteReconciliation(repositories, remote)

	if !reflect.DeepEqual(plan.Create, []string{"repo-a"}) {
		t.Fatalf("Create = %v", plan.Create)
	}
	if !reflect.DeepEqual(plan.Delete, []string{"repo-c"}) {
		t.Fatalf("Delete = %v", plan.Delete)
	}
	if len(plan.Push) != 2 {
		t.Fatalf("Push = %v", plan.Push)
	}
}

func TestPromoteRemoteLocalModeIsNonDestructive(t *testing.T) {
	plan := promoteRemotePlanForMode(
		[]string{"C:/work/repo-a"},
		[]string{"repo-b"},
		true,
	)
	if len(plan.Push) != 0 || len(plan.Delete) != 0 {
		t.Fatalf("local plan must not push/delete: %+v", plan)
	}
}
