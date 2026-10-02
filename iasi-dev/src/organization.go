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

func prepareParms(command string, Context *structures.Context) {
	if command == "promote-check" {
		preparePromotePaths(Context)
		commands.SetDryRun(Context.DryRun)
		return
	}

	if command == "promote" {
		preparePromotePaths(Context)
		Context.Targets = []string{Context.SourcePath}
		args.Prepare(Context)
		preparePromoteOrganizations(Context)
		commands.SetDryRun(Context.DryRun)
		return
	}

	if command == "workflow" && Context.Subcommand == "promote" {
		preparePromotePaths(Context)
		Context.Targets = []string{Context.SourcePath}
		args.Prepare(Context)
		preparePromoteOrganizations(Context)
		loadOrganizationVersion(Context)
		commands.SetDryRun(Context.DryRun)
		return
	}

	args.Prepare(Context)
	prepareOrganization(Context)
	if !isVersionWrite(command, Context) {
		loadOrganizationVersion(Context)
	}

	commands.SetDryRun(Context.DryRun)
}

func preparePromoteOrganizations(Context *structures.Context) {
	source := filepath.Base(filepath.Clean(Context.SourcePath))
	destination := filepath.Base(filepath.Clean(Context.DestinationPath))
	if source == "" || source == "." {
		cli.Error(RC.Error, *Context, "No se puede deducir la organización origen desde %s.", Context.SourcePath)
	}
	if destination == "" || destination == "." {
		cli.Error(RC.Error, *Context, "No se puede deducir la organización destino desde %s.", Context.DestinationPath)
	}

	Context.Organization = source
	Context.SourceOrganization = source
	Context.DestinationOrganization = destination
}

func isVersionWrite(command string, Context *structures.Context) bool {
	return command == "version" && Context.TargetVersion != ""
}

func prepareOrganization(Context *structures.Context) {
	if Context.Organization != "" {
		return
	}
	if len(Context.Repos) == 0 {
		cli.Error(RC.Error, *Context, "No se puede deducir la organización: no se han descubierto repositorios Git.")
	}

	organization := ""
	for _, repository := range Context.Repos {
		candidate := repositoryOrganization(Context, repository)
		if organization == "" {
			organization = candidate
			continue
		}
		if candidate != organization {
			cli.Error(RC.Error, *Context, "Los repositorios descubiertos pertenecen a organizaciones distintas: %q y %q.", organization, candidate)
		}
	}

	Context.Organization = organization
}

func repositoryOrganization(Context *structures.Context, repository string) string {
	result := commands.Run(repository, Context.LogFile, "git", "remote", "get-url", "origin")
	if result.RC != RC.OK {
		cli.Error(RC.Error, *Context, "No se puede leer origin de %s: %s", repository, organizationCommandError(result))
	}

	organization := githubOrganization(strings.TrimSpace(result.Stdout))
	if organization == "" {
		cli.Error(RC.Error, *Context, "No se puede deducir una organización GitHub de origin en %s: %q", repository, strings.TrimSpace(result.Stdout))
	}
	return organization
}

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

func loadOrganizationVersion(Context *structures.Context) {
	result := commands.Run(".", Context.LogFile, "gh", "variable", "get", consts.OrganizationVersionVariable, "--org", Context.Organization)
	if result.RC != RC.OK {
		cli.Error(RC.Error, *Context, "No se pudo leer %s de %s: %s", consts.OrganizationVersionVariable, Context.Organization, organizationCommandError(result))
	}

	version := strings.TrimSpace(result.Stdout)
	if version == "" {
		cli.Error(RC.Error, *Context, "%s de %s está vacía.", consts.OrganizationVersionVariable, Context.Organization)
	}

	Context.Version = version
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
