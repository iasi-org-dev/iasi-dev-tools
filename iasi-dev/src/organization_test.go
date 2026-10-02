package main

import "testing"

func TestGithubOrganization(t *testing.T) {
	tests := map[string]string{
		"https://github.com/iasi-org-dev/iasi-r.git": "iasi-org-dev",
		"git@github.com:iasi-org/iasi-r.git":          "iasi-org",
		"https://gitlab.com/iasi-org/iasi-r.git":     "",
	}
	for remote, want := range tests {
		if got := githubOrganization(remote); got != want {
			t.Fatalf("githubOrganization(%q) = %q, want %q", remote, got, want)
		}
	}
}
