package runners

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"iasi-dev/internal/cli"
	"iasi-dev/internal/commands"
	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

// Materialize recreates an organization workspace in destination without carrying Git history.
// Unless -l is active, the materialized organization is then published through its remotes.
func Materialize(Parms *structures.Parms) []string {
	repositories := materializeLocal(Parms)
	if len(repositories) == 0 || Parms.Local {
		return repositories
	}

	Parms.Repos = repositories
	return push(Parms)
}

// materializeLocal performs only the local materialization transaction.
// Publication is deliberately a separate primitive so workflows can compose both stages.
func materializeLocal(Parms *structures.Parms) []string {
	if Parms.MaterializeDestination == "" {
		cli.Error(RC.Error, *Parms, "materialize requiere un destino.")
	}
	if len(Parms.Repos) == 0 {
		cli.Error(RC.Error, *Parms, "No se encontraron repositorios para materializar.")
	}

	destination := filepath.Clean(Parms.MaterializeDestination)
	confirmed, _ := confirmMaterializeDestination(Parms, destination)
	if !confirmed {
		RC.Add(Parms.RC, RC.NothingToDo)
		return nil
	}

	destinationOrganization := materializeOrganization(destination)
	temporary := materializeTemporary(destination)

	cli.Header(*Parms, "Materialize %s", Parms.Organization)
	cli.Info(*Parms, "Source organization: %s", Parms.Organization)
	cli.Info(*Parms, "Destination: %s", destination)
	cli.Info(*Parms, "Destination organization: %s", destinationOrganization)
	cli.Info(*Parms, "Temporary workspace: %s", temporary)
	cli.Info(*Parms, "Mode: local")

	if err := prepareMaterializeTemporary(temporary); err != nil {
		cli.Error(RC.Error, *Parms, "No se pudo preparar el workspace temporal %s: %v", temporary, err)
	}

	completed := false
	defer func() {
		if !completed {
			_ = os.RemoveAll(temporary)
		}
	}()

	destinationRepositories := make([]string, 0, len(Parms.Repos))
	for _, repository := range Parms.Repos {
		name := filepath.Base(repository)
		temporaryRepository := filepath.Join(temporary, name)
		destinationRepository := filepath.Join(destination, name)
		destinationRepositories = append(destinationRepositories, destinationRepository)

		cli.Step(*Parms, "Materializing %s", name)
		if err := materializeCopy(repository, temporaryRepository); err != nil {
			cli.Error(RC.Error, *Parms, "No se pudo copiar %s: %v", repository, err)
		}
		if !materializeGitInit(Parms, temporaryRepository) {
			cli.Error(RC.Error, *Parms, "No se pudo inicializar Git en %s. Revisa el log: %s", name, logName(*Parms))
		}
		if !materializeRemote(Parms, temporaryRepository, destinationOrganization, name) {
			cli.Error(RC.Error, *Parms, "No se pudo configurar origin en %s. Revisa el log: %s", name, logName(*Parms))
		}
		if !materializeCommit(Parms, temporaryRepository) {
			cli.Error(RC.Error, *Parms, "No se pudo crear el commit materializado de %s. Revisa el log: %s", name, logName(*Parms))
		}
	}

	if err := materializeReplace(temporary, destination); err != nil {
		cli.Error(RC.Error, *Parms, "No se pudo reemplazar %s por el workspace materializado: %v", destination, err)
	}
	completed = true

	cli.Success(*Parms, "Organization materialized.")
	return destinationRepositories
}

func materializeOrganization(destination string) string {
	return filepath.Base(filepath.Clean(destination))
}

func materializeTemporary(destination string) string {
	parent := filepath.Dir(filepath.Clean(destination))
	name := filepath.Base(filepath.Clean(destination))
	return filepath.Join(parent, "."+name+".materialize.tmp")
}

func materializeBackup(destination string) string {
	parent := filepath.Dir(filepath.Clean(destination))
	name := filepath.Base(filepath.Clean(destination))
	return filepath.Join(parent, "."+name+".materialize.bak")
}

func prepareMaterializeTemporary(temporary string) error {
	if err := os.RemoveAll(temporary); err != nil {
		return err
	}
	return os.MkdirAll(temporary, 0755)
}

// materializeCopy copies the current working tree but deliberately omits Git metadata.
func materializeCopy(source string, destination string) error {
	return filepath.WalkDir(source, func(path string, item os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		if relative == ".git" || strings.HasPrefix(relative, ".git"+string(filepath.Separator)) {
			if item.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		target := destination
		if relative != "." {
			target = filepath.Join(destination, relative)
		}

		info, err := item.Info()
		if err != nil {
			return err
		}
		if item.IsDir() {
			return os.MkdirAll(target, info.Mode().Perm())
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return copySymlink(path, target)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("tipo de fichero no soportado: %s", path)
		}
		return copyFile(path, target)
	})
}

func materializeGitInit(Parms *structures.Parms, repository string) bool {
	result := commands.RunLogged(repository, Parms.LogFile, "git", "init", "-b", "main")
	return result.RC == RC.OK
}

func materializeRemote(Parms *structures.Parms, repository string, organization string, name string) bool {
	remote := fmt.Sprintf("https://github.com/%s/%s.git", organization, name)
	result := commands.RunLogged(repository, Parms.LogFile, "git", "remote", "add", "origin", remote)
	return result.RC == RC.OK
}

func materializeCommit(Parms *structures.Parms, repository string) bool {
	result := commands.RunLogged(repository, Parms.LogFile, "git", "add", "-A", ".")
	if result.RC != RC.OK {
		return false
	}

	message := materializeCommitMessage(Parms)
	result = commands.RunLogged(repository, Parms.LogFile, "git", "commit", "-m", message)
	return result.RC == RC.OK
}

func materializeCommitMessage(Parms *structures.Parms) string {
	if Parms.TargetVersion != "" {
		return "IASI organization version " + Parms.TargetVersion
	}
	if strings.TrimSpace(Parms.Message) != "" {
		return Parms.Message
	}
	return "IASI materialization"
}

// materializeReplace swaps the complete destination transactionally. If the final rename fails,
// the previous destination is restored from the sibling backup.
func materializeReplace(temporary string, destination string) error {
	backup := materializeBackup(destination)
	if err := os.RemoveAll(backup); err != nil {
		return err
	}

	destinationExists := false
	if info, err := os.Stat(destination); err == nil {
		if !info.IsDir() {
			return fmt.Errorf("el destino existe pero no es un directorio: %s", destination)
		}
		destinationExists = true
		if err := os.Rename(destination, backup); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	if err := os.Rename(temporary, destination); err != nil {
		if destinationExists {
			_ = os.Rename(backup, destination)
		}
		return err
	}

	if destinationExists {
		if err := os.RemoveAll(backup); err != nil {
			return err
		}
	}
	return nil
}

func confirmMaterializeDestination(Parms *structures.Parms, destination string) (bool, bool) {
	info, err := os.Stat(destination)
	if err == nil {
		if !info.IsDir() {
			cli.Error(RC.Error, *Parms, "El destino existe pero no es un directorio: %s", destination)
		}
		return true, false
	}
	if !os.IsNotExist(err) {
		cli.Error(RC.Error, *Parms, "No se puede comprobar el destino %s: %v", destination, err)
	}

	fmt.Printf("El directorio %s no existe. ¿Quieres crearlo? [s/N] ", destination)
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	if !materializeConfirmed(answer) {
		return false, false
	}
	return true, true
}

func materializeConfirmed(answer string) bool {
	switch strings.ToLower(strings.TrimSpace(answer)) {
	case "s", "si", "sí", "y", "yes":
		return true
	default:
		return false
	}
}
