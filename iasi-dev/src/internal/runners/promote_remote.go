package runners

import (
	"path/filepath"
	"sort"
	"strings"

	"iasi-dev/internal/cli"
	"iasi-dev/internal/commands"
	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

type promoteRemotePlan struct {
	Create []string
	Push   []string
	Delete []string
}

// promoteRemoteSynchronize makes the destination GitHub organization mirror the
// already promoted local workspace. Local repositories are the publication source:
// missing remote repositories are created, every current repository is force-pushed,
// and repositories no longer present locally are removed only after all pushes succeed.
func promoteRemoteSynchronize(Parms *structures.Parms, repositories []string) {
	remoteRepositories := promoteRemoteRepositories(Parms)
	plan := promoteRemoteReconciliation(repositories, remoteRepositories)

	for _, name := range plan.Create {
		visibility := promoteRemoteSourceVisibility(Parms, name)
		cli.Step(*Parms, "Creating remote repository %s/%s", Parms.DestinationOrganization, name)
		result := commands.RunLogged(
			".",
			Parms.LogFile,
			"gh",
			"repo",
			"create",
			Parms.DestinationOrganization+"/"+name,
			"--"+visibility,
		)
		if result.RC != RC.OK {
			cli.Error(RC.Error, *Parms, "No se pudo crear el repositorio remoto %s/%s. Revisa el log: %s", Parms.DestinationOrganization, name, logName(*Parms))
		}
	}

	for _, repository := range plan.Push {
		cli.Step(*Parms, "Synchronizing %s", filepath.Base(repository))
		if !pushRepository(Parms, repository) {
			cli.Error(RC.Error, *Parms, "No se pudo sincronizar %s. Revisa el log: %s", filepath.Base(repository), logName(*Parms))
		}
	}

	for _, name := range plan.Delete {
		cli.Step(*Parms, "Removing obsolete remote repository %s/%s", Parms.DestinationOrganization, name)
		result := commands.RunLogged(
			".",
			Parms.LogFile,
			"gh",
			"repo",
			"delete",
			Parms.DestinationOrganization+"/"+name,
			"--yes",
		)
		if result.RC != RC.OK {
			cli.Error(RC.Error, *Parms, "No se pudo eliminar el repositorio remoto obsoleto %s/%s. Revisa el log: %s", Parms.DestinationOrganization, name, logName(*Parms))
		}
	}
}

func promoteRemotePreview(Parms *structures.Parms, repositories []string) {
	cli.Info(*Parms, "Synchronizing remote organization...")
	for _, repository := range repositories {
		name := filepath.Base(repository)
		cli.Preview("Ensure remote repository exists: %s/%s\n", Parms.DestinationOrganization, name)
		cli.Preview("Force remote main from local snapshot: %s\n", repository)
	}
	cli.Preview("Remove remote repositories not present in local destination: %s\n", Parms.DestinationOrganization)
}

func promoteRemoteRepositories(Parms *structures.Parms) []string {
	result := commands.Run(
		".",
		Parms.LogFile,
		"gh",
		"api",
		"--paginate",
		"orgs/"+Parms.DestinationOrganization+"/repos?per_page=100",
		"--jq",
		".[] | .name",
	)
	if result.RC != RC.OK {
		cli.Error(RC.Error, *Parms, "No se pudieron listar los repositorios remotos de %s.", Parms.DestinationOrganization)
	}

	repositories := []string{}
	for _, line := range strings.Split(result.Stdout, "\n") {
		name := strings.TrimSpace(line)
		if name != "" {
			repositories = append(repositories, name)
		}
	}
	return repositories
}

func promoteRemoteReconciliation(repositories []string, remoteRepositories []string) promoteRemotePlan {
	remote := map[string]string{}
	for _, name := range remoteRepositories {
		name = strings.TrimSpace(name)
		if name != "" {
			remote[strings.ToLower(name)] = name
		}
	}

	desired := map[string]string{}
	plan := promoteRemotePlan{
		Create: []string{},
		Push:   append([]string{}, repositories...),
		Delete: []string{},
	}

	for _, repository := range repositories {
		name := filepath.Base(filepath.Clean(repository))
		key := strings.ToLower(name)
		desired[key] = name
		if _, exists := remote[key]; !exists {
			plan.Create = append(plan.Create, name)
		}
	}

	for key, name := range remote {
		if _, exists := desired[key]; !exists {
			plan.Delete = append(plan.Delete, name)
		}
	}

	sort.Strings(plan.Create)
	sort.Strings(plan.Delete)
	return plan
}

func promoteRemoteSourceVisibility(Parms *structures.Parms, destinationName string) string {
	sourceName := promoteRemoteSourceRepositoryName(destinationName, Parms.SourceOrganization, Parms.DestinationOrganization)
	result := commands.Run(
		".",
		Parms.LogFile,
		"gh",
		"api",
		"repos/"+Parms.SourceOrganization+"/"+sourceName,
		"--jq",
		".visibility",
	)
	if result.RC != RC.OK {
		cli.Error(RC.Error, *Parms, "No se pudo obtener la visibilidad de %s/%s.", Parms.SourceOrganization, sourceName)
	}

	visibility := strings.ToLower(strings.TrimSpace(result.Stdout))
	switch visibility {
	case "public", "private", "internal":
		return visibility
	default:
		cli.Error(RC.Error, *Parms, "Visibilidad no soportada para %s/%s: %s", Parms.SourceOrganization, sourceName, visibility)
	}
	return ""
}

func promoteRemoteSourceRepositoryName(destinationName string, sourceOrganization string, destinationOrganization string) string {
	if sourceOrganization == "" || destinationOrganization == "" {
		return destinationName
	}
	if strings.EqualFold(destinationName, destinationOrganization+".github.io") {
		return sourceOrganization + ".github.io"
	}
	return destinationName
}
