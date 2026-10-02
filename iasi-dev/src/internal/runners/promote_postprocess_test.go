package runners

import (
	"strings"
	"testing"
)

func TestPromotePostprocessRepositoryName(t *testing.T) {
	got := promotePostprocessRepositoryName(
		"iasi-org-dev.github.io",
		"iasi-org-dev",
		"iasi-org",
	)
	if got != "iasi-org.github.io" {
		t.Fatalf("got %q", got)
	}
}

func TestPromotePostprocessContentRewritesOrganization(t *testing.T) {
	input := []byte("https://github.com/iasi-org-dev/repo")
	output := string(promotePostprocessContent(input, ".toml", "iasi-org-dev", "iasi-org"))
	if strings.Contains(output, "iasi-org-dev") || !strings.Contains(output, "iasi-org") {
		t.Fatalf("unexpected output: %q", output)
	}
}
