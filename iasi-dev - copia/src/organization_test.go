package main

import "testing"

func TestGithubOrganizationFromHTTPS(t *testing.T) {
	if got := githubOrganization("https://github.com/iasi-org-dev/iasi-quarto.git"); got != "iasi-org-dev" {
		t.Fatalf("organization = %q, want iasi-org-dev", got)
	}
}

func TestGithubOrganizationFromSCPSSH(t *testing.T) {
	if got := githubOrganization("git@github.com:iasi-org/iasi-quarto.git"); got != "iasi-org" {
		t.Fatalf("organization = %q, want iasi-org", got)
	}
}

func TestGithubOrganizationRejectsNonGithubRemote(t *testing.T) {
	if got := githubOrganization("https://gitlab.com/iasi-org/iasi-quarto.git"); got != "" {
		t.Fatalf("organization = %q, want empty", got)
	}
}
