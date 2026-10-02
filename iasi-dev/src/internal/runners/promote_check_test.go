package runners

import (
	"strings"
	"testing"
)

func TestPromoteCheckPatternsContainSourceName(t *testing.T) {
	patterns := promoteCheckPatterns("C:/work/iasi-org-dev")
	found := false
	for _, pattern := range patterns {
		if pattern.label == "source-name" && strings.EqualFold(pattern.value, "iasi-org-dev") {
			found = true
		}
	}
	if !found {
		t.Fatal("source-name pattern missing")
	}
}

func TestPromoteCheckSnippetIsBounded(t *testing.T) {
	text := strings.Repeat("x", 300)
	if got := promoteCheckSnippet(text); len(got) > 243 {
		t.Fatalf("snippet too long: %d", len(got))
	}
}
