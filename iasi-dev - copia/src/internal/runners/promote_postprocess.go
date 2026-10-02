package runners

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"iasi-dev/internal/cli"
	"iasi-dev/internal/commands"
	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

var promotePostprocessExcludedRepositories = map[string]bool{
	"iasi-dev-tools": true,
}

var promotePostprocessExtensions = map[string]bool{
	".yml":  true,
	".yaml": true,
	".toml": true,
	".qmd":  true,
	".md":   true,
	".html": true,
	".js":   true,
	".json": true,
}

var promotePostprocessRawOrganizationExtensions = map[string]bool{
	".yml":  true,
	".yaml": true,
	".toml": true,
	".js":   true,
	".json": true,
}

func promotePostprocess(Parms *structures.Parms, root string, repositories []string) []string {
	processed, err := promotePostprocessGitHubPages(Parms, root, repositories)
	if err != nil {
		cli.Error(RC.Error, *Parms, "No se pudo postprocesar el repositorio GitHub Pages: %v", err)
	}

	changed, err := promotePostprocessFiles(root, Parms.SourceOrganization, Parms.DestinationOrganization)
	if err != nil {
		cli.Error(RC.Error, *Parms, "No se pudieron postprocesar los ficheros promovidos: %v", err)
	}

	for _, repository := range changed {
		if !promotePostprocessAmend(Parms, repository) {
			cli.Error(RC.Error, *Parms, "No se pudo actualizar el snapshot postprocesado de %s. Revisa el log: %s", filepath.Base(repository), logName(*Parms))
		}
	}

	return processed
}

func promotePostprocessPreview(Parms *structures.Parms) {
	sourcePages := Parms.SourceOrganization + ".github.io"
	destinationPages := Parms.DestinationOrganization + ".github.io"
	if Parms.SourceOrganization != "" && Parms.DestinationOrganization != "" && !strings.EqualFold(sourcePages, destinationPages) {
		cli.Preview("Postprocess GitHub Pages repository when present: %s -> %s\n", sourcePages, destinationPages)
	}
	cli.Preview("Postprocess files by extension: %s\n", strings.Join(promotePostprocessExtensionList(), ", "))
}

func promotePostprocessGitHubPages(Parms *structures.Parms, root string, repositories []string) ([]string, error) {
	if Parms.SourceOrganization == "" || Parms.DestinationOrganization == "" {
		return append([]string{}, repositories...), nil
	}

	sourceName := Parms.SourceOrganization + ".github.io"
	destinationName := Parms.DestinationOrganization + ".github.io"
	processed := append([]string{}, repositories...)

	for index, repository := range processed {
		if !strings.EqualFold(filepath.Base(repository), sourceName) {
			continue
		}

		destination := filepath.Join(root, destinationName)
		if !strings.EqualFold(filepath.Clean(repository), filepath.Clean(destination)) {
			if _, err := os.Stat(destination); err == nil {
				return nil, fmt.Errorf("el destino especial ya existe: %s", destination)
			} else if !os.IsNotExist(err) {
				return nil, err
			}
			if err := os.Rename(repository, destination); err != nil {
				return nil, err
			}
		}

		remote := fmt.Sprintf("https://github.com/%s/%s.git", Parms.DestinationOrganization, destinationName)
		if !promoteCommand(Parms, destination, "remote", "set-url", "origin", remote) {
			return nil, fmt.Errorf("no se pudo configurar origin de %s", destinationName)
		}

		processed[index] = destination
		cli.Step(*Parms, "Postprocessed GitHub Pages repository: %s -> %s", sourceName, destinationName)
		break
	}

	return processed, nil
}

func promotePostprocessRepositoryName(name string, sourceOrganization string, destinationOrganization string) string {
	if sourceOrganization == "" || destinationOrganization == "" {
		return name
	}
	if strings.EqualFold(name, sourceOrganization+".github.io") {
		return destinationOrganization + ".github.io"
	}
	return name
}

func promotePostprocessFiles(root string, sourceOrganization string, destinationOrganization string) ([]string, error) {
	if sourceOrganization == "" || destinationOrganization == "" || strings.EqualFold(sourceOrganization, destinationOrganization) {
		return nil, nil
	}

	changedRepositories := map[string]bool{}

	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if promotePostprocessExcludedRepository(root, path, entry.Name()) {
				return filepath.SkipDir
			}
			if entry.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return nil
		}

		extension := strings.ToLower(filepath.Ext(entry.Name()))
		if !promotePostprocessExtensions[extension] {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		processed := promotePostprocessContent(data, extension, sourceOrganization, destinationOrganization)
		if bytes.Equal(data, processed) {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}
		if err := os.WriteFile(path, processed, info.Mode().Perm()); err != nil {
			return err
		}

		repository := promotePostprocessRepositoryRoot(root, path)
		if repository != "" {
			changedRepositories[repository] = true
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	changed := make([]string, 0, len(changedRepositories))
	for repository := range changedRepositories {
		changed = append(changed, repository)
	}
	sort.Strings(changed)
	return changed, nil
}

func promotePostprocessExcludedRepository(root string, path string, name string) bool {
	if strings.EqualFold(filepath.Clean(path), filepath.Clean(root)) {
		return false
	}
	if !strings.EqualFold(filepath.Clean(filepath.Dir(path)), filepath.Clean(root)) {
		return false
	}
	for repository := range promotePostprocessExcludedRepositories {
		if strings.EqualFold(name, repository) {
			return true
		}
	}
	return false
}

func promotePostprocessRepositoryRoot(root string, path string) string {
	relative, err := filepath.Rel(root, path)
	if err != nil || relative == "." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return ""
	}
	parts := strings.Split(relative, string(filepath.Separator))
	if len(parts) < 2 || parts[0] == "" {
		return ""
	}
	repository := filepath.Join(root, parts[0])
	if info, err := os.Stat(filepath.Join(repository, ".git")); err != nil || !info.IsDir() {
		return ""
	}
	return repository
}

func promotePostprocessContent(data []byte, extension string, sourceOrganization string, destinationOrganization string) []byte {
	replacements := []string{
		sourceOrganization + ".github.io", destinationOrganization + ".github.io",
		"github.com/" + sourceOrganization, "github.com/" + destinationOrganization,
		"github.com%2F" + sourceOrganization, "github.com%2F" + destinationOrganization,
		"git@github.com:" + sourceOrganization + "/", "git@github.com:" + destinationOrganization + "/",
		"ssh://git@github.com/" + sourceOrganization + "/", "ssh://git@github.com/" + destinationOrganization + "/",
		sourceOrganization + "/", destinationOrganization + "/",
		sourceOrganization + "%2F", destinationOrganization + "%2F",
	}

	if promotePostprocessRawOrganizationExtensions[strings.ToLower(extension)] {
		replacements = append(replacements, sourceOrganization, destinationOrganization)
	}

	replacer := strings.NewReplacer(replacements...)
	return []byte(replacer.Replace(string(data)))
}

func promotePostprocessAmend(Parms *structures.Parms, repository string) bool {
	result := commands.RunLogged(repository, Parms.LogFile, "git", "add", "-A", ".")
	if result.RC != RC.OK {
		return false
	}
	result = commands.RunLogged(repository, Parms.LogFile, "git", "commit", "--amend", "--no-edit")
	return result.RC == RC.OK
}

func promotePostprocessExtensionList() []string {
	extensions := make([]string, 0, len(promotePostprocessExtensions))
	for extension := range promotePostprocessExtensions {
		extensions = append(extensions, extension)
	}
	sort.Strings(extensions)
	return extensions
}
