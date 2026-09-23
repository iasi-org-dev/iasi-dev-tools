package main

import (
	"net/url"
	"path/filepath"
	"strings"

	"iasi-dev/internal/args"
	"iasi-dev/internal/cli"
	"iasi-dev/internal/commands"
	"iasi-dev/internal/consts"
	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

// prepareParms discovers the effective repositories, resolves the organization and loads its VERSION.
func prepareParms(command string, Parms *structures.Parms) {
	if command == "promote-check" {
		preparePromotePaths(Parms)
		commands.SetDryRun(Parms.DryRun)
		return
	}

	if command == "promote" {
		preparePromotePaths(Parms)
		Parms.Targets = []string{Parms.SourcePath}
		args.Prepare(Parms)
		preparePromoteOrganizations(Parms)
		commands.SetDryRun(Parms.DryRun)
		return
	}

	if command == "workflow" && Parms.Subcommand == "promote" {
		preparePromotePaths(Parms)
		Parms.Targets = []string{Parms.SourcePath}
		args.Prepare(Parms)
		preparePromoteOrganizations(Parms)
		loadOrganizationVersion(Parms)
		commands.SetDryRun(Parms.DryRun)
		return
	}

	args.Prepare(Parms)
	prepareOrganization(Parms)
	if !isVersionWrite(command, Parms) {
		loadOrganizationVersion(Parms)
	}

	// Check modes become effective only after the real preparation has completed.
	commands.SetDryRun(Parms.DryRun)
}

// preparePromoteOrganizations derives promotion organization names from the explicit workspace paths.
func preparePromoteOrganizations(Parms *structures.Parms) {
	source := filepath.Base(filepath.Clean(Parms.SourcePath))
	destination := filepath.Base(filepath.Clean(Parms.DestinationPath))
	if source == "" || source == "." {
		cli.Error(RC.Error, *Parms, "No se puede deducir la organización origen desde %s.", Parms.SourcePath)
	}
	if destination == "" || destination == "." {
		cli.Error(RC.Error, *Parms, "No se puede deducir la organización destino desde %s.", Parms.DestinationPath)
	}

	Parms.Organization = source
	Parms.SourceOrganization = source
	Parms.DestinationOrganization = destination
}

func isVersionWrite(command string, Parms *structures.Parms) bool {
	return command == "version" && Parms.TargetVersion != ""
}

// prepareOrganization keeps an explicit organization or deduces one from the discovered Git origins.
func prepareOrganization(Parms *structures.Parms) {
	if Parms.Organization != "" {
		return
	}
	if len(Parms.Repos) == 0 {
		cli.Error(RC.Error, *Parms, "No se puede deducir la organización: no se han descubierto repositorios Git.")
	}

	organization := ""
	for _, repository := range Parms.Repos {
		candidate := repositoryOrganization(Parms, repository)
		if organization == "" {
			organization = candidate
			continue
		}
		if candidate != organization {
			cli.Error(RC.Error, *Parms, "Los repositorios descubiertos pertenecen a organizaciones distintas: %q y %q.", organization, candidate)
		}
	}

	Parms.Organization = organization
}

// repositoryOrganization returns the GitHub organization owning repository's origin remote.
func repositoryOrganization(Parms *structures.Parms, repository string) string {
	result := commands.Run(repository, Parms.LogFile, "git", "remote", "get-url", "origin")
	if result.RC != RC.OK {
		cli.Error(RC.Error, *Parms, "No se puede leer origin de %s: %s", repository, organizationCommandError(result))
	}

	organization := githubOrganization(strings.TrimSpace(result.Stdout))
	if organization == "" {
		cli.Error(RC.Error, *Parms, "No se puede deducir una organización GitHub de origin en %s: %q", repository, strings.TrimSpace(result.Stdout))
	}
	return organization
}

// githubOrganization extracts the owner from common github.com remote URL forms.
func githubOrganization(remote string) string {
	remote = strings.TrimSpace(remote)
	if remote == "" {
		return ""
	}

	const scpPrefix = "git@github.com:"
	if strings.HasPrefix(remote, scpPrefix) {
		return firstPathPart(strings.TrimPrefix(remote, scpPrefix))
	}

	const sshPrefix = "ssh://git@github.com/"
	if strings.HasPrefix(remote, sshPrefix) {
		return firstPathPart(strings.TrimPrefix(remote, sshPrefix))
	}

	parsed, err := url.Parse(remote)
	if err != nil || !strings.EqualFold(parsed.Hostname(), "github.com") {
		return ""
	}
	return firstPathPart(strings.TrimPrefix(parsed.Path, "/"))
}

func firstPathPart(path string) string {
	path = strings.TrimSuffix(strings.TrimSpace(path), ".git")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 || parts[0] == "" {
		return ""
	}
	return parts[0]
}

// loadOrganizationVersion loads VERSION from the resolved GitHub organization.
// Failure is fatal because this call is also the common GitHub connectivity/authentication checkpoint.
func loadOrganizationVersion(Parms *structures.Parms) {
	result := commands.Run(".", Parms.LogFile, "gh", "variable", "get", consts.OrganizationVersionVariable, "--org", Parms.Organization)
	if result.RC != RC.OK {
		cli.Error(RC.Error, *Parms, "No se pudo leer %s de %s: %s", consts.OrganizationVersionVariable, Parms.Organization, organizationCommandError(result))
	}

	version := strings.TrimSpace(result.Stdout)
	if version == "" {
		cli.Error(RC.Error, *Parms, "%s de %s está vacía.", consts.OrganizationVersionVariable, Parms.Organization)
	}

	Parms.Version = version
}

func organizationCommandError(result structures.Result) string {
	message := strings.TrimSpace(result.Stderr)
	if message == "" {
		message = strings.TrimSpace(result.Stdout)
	}
	if message == "" {
		message = "gh devolvió un error sin mensaje"
	}
	return message
}
