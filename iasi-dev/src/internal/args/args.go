package args

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"iasi-dev/internal/cli"
	iasiconfig "iasi-dev/internal/config"
	"iasi-dev/internal/consts"
	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
	"iasi-dev/internal/tools"
)

// Parse parses command-line arguments. Preparation is performed later by main.
func Parse(command string, values []string) structures.Context {
	subcommand := ""

	if command == "workflow" {
		if len(values) == 0 {
			rc := RC.OK
			cli.Error(RC.Error, structures.Context{RC: &rc}, "Falta el comando del workflow.")
		}

		subcommand = values[0]
		values = values[1:]
	}

	Context := parseArguments(command, subcommand, values)
	Context.Subcommand = subcommand
	validateCheckModes(&Context)
	validateLocalMode(command, &Context)
	extractPromoteOperands(command, &Context)
	extractWorkflowPromoteParameters(command, &Context)
	extractTargetVersion(command, &Context)
	extractVersionOperands(command, &Context)
	validateOrganizationWideCommand(command, &Context)
	extractMaterializeOperands(command, &Context)
	if Context.RequestedTargets == nil {
		Context.RequestedTargets = append([]string{}, Context.Targets...)
	}

	return Context
}

func validateCheckModes(Context *structures.Context) {
	if Context.PrepareOnly && Context.DryRun {
		cli.Error(RC.Error, *Context, "-m y -M son incompatibles.")
	}
}

func validateLocalMode(command string, Context *structures.Context) {
	if command == "freeze" && Context.Local {
		cli.Error(RC.Error, *Context, "-l no está soportado por freeze: freeze siempre publica los tags.")
	}
}

func extractPromoteOperands(command string, Context *structures.Context) {
	isPromote := command == "promote" || command == "promote-check"
	if !isPromote {
		return
	}
	if len(Context.Targets) != 3 {
		cli.Error(RC.Error, *Context, "promote requiere <version> <source-path> <destination-path>.")
	}

	version := strings.TrimSpace(Context.Targets[0])
	source := strings.TrimSpace(Context.Targets[1])
	destination := strings.TrimSpace(Context.Targets[2])
	if !looksLikeSemanticVersion(version) {
		cli.Error(RC.Error, *Context, "La versión no es válida: %s", version)
	}
	if source == "" || destination == "" {
		cli.Error(RC.Error, *Context, "source-path y destination-path son obligatorios.")
	}

	sourcePath := filepath.Clean(source)
	destinationPath := filepath.Clean(destination)
	if strings.EqualFold(sourcePath, destinationPath) {
		cli.Error(RC.Error, *Context, "source-path y destination-path deben ser rutas distintas.")
	}

	Context.TargetVersion = version
	Context.SourcePath = sourcePath
	Context.DestinationPath = destinationPath
	Context.DestinationOrganization = filepath.Base(destinationPath)
	Context.Targets = nil
}

func extractWorkflowPromoteParameters(command string, Context *structures.Context) {
	if command != "workflow" || Context.Subcommand != "promote" {
		return
	}
	if len(Context.Targets) != 0 {
		cli.Error(RC.Error, *Context, "workflow promote usa parámetros nombrados: --source, --dest y --version.")
	}
	if strings.TrimSpace(Context.SourcePath) == "" {
		cli.Error(RC.Error, *Context, "workflow promote requiere --source <source-path>.")
	}
	if strings.TrimSpace(Context.DestinationPath) == "" {
		cli.Error(RC.Error, *Context, "workflow promote requiere --dest <destination-path>.")
	}
	if strings.TrimSpace(Context.NextVersion) == "" {
		cli.Error(RC.Error, *Context, "workflow promote requiere --version vMAJOR.MINOR.PATCH.")
	}
	if !looksLikeSemanticVersion(Context.NextVersion) {
		cli.Error(RC.Error, *Context, "La nueva versión no es válida: %s", Context.NextVersion)
	}

	sourcePath := filepath.Clean(Context.SourcePath)
	destinationPath := filepath.Clean(Context.DestinationPath)
	if strings.EqualFold(sourcePath, destinationPath) {
		cli.Error(RC.Error, *Context, "source-path y destination-path deben ser rutas distintas.")
	}

	Context.SourcePath = sourcePath
	Context.DestinationPath = destinationPath
	Context.DestinationOrganization = filepath.Base(destinationPath)
}

func extractTargetVersion(command string, Context *structures.Context) {
	if command != "restore" || len(Context.Targets) == 0 {
		return
	}

	Context.TargetVersion = Context.Targets[0]
	Context.Targets = Context.Targets[1:]
}

func extractVersionOperands(command string, Context *structures.Context) {
	if command != "version" || len(Context.Targets) == 0 {
		return
	}
	if len(Context.Targets) > 1 {
		cli.Error(RC.Error, *Context, "version acepta como máximo una versión vMAJOR.MINOR.PATCH.")
	}

	version := Context.Targets[0]
	if !looksLikeSemanticVersion(version) {
		cli.Error(RC.Error, *Context, "La versión no es válida: %s", version)
	}

	Context.TargetVersion = version
	Context.Targets = nil
}

func looksLikeSemanticVersion(value string) bool {
	if !strings.HasPrefix(value, "v") {
		return false
	}
	parts := strings.Split(strings.TrimPrefix(value, "v"), ".")
	if len(parts) != 3 {
		return false
	}
	for _, part := range parts {
		if part == "" || (len(part) > 1 && part[0] == '0') {
			return false
		}
		for _, char := range part {
			if char < '0' || char > '9' {
				return false
			}
		}
	}
	return true
}

func validateOrganizationWideCommand(command string, Context *structures.Context) {
	organizationWide := command == "freeze" || command == "promote" || command == "promote-check" || (command == "workflow" && Context.Subcommand == "promote")
	if !organizationWide {
		return
	}
	if len(Context.Targets) != 0 {
		cli.Error(RC.Error, *Context, "%s opera sobre la organización completa y no acepta targets; usa --path para elegir el workspace.", command)
	}
}

func extractMaterializeOperands(command string, Context *structures.Context) {
	if command != "materialize" {
		return
	}

	Context.RequestedTargets = append([]string{}, Context.Targets...)
	if len(Context.Targets) == 0 {
		return
	}
	if len(Context.Targets) > 2 {
		cli.Error(RC.Error, *Context, "materialize acepta <destino> y, opcionalmente, [origen].")
	}

	destination, err := filepath.Abs(Context.Targets[0])
	if err != nil {
		cli.Error(RC.Error, *Context, "No se puede resolver el destino de materialize: %q", Context.Targets[0])
	}
	Context.MaterializeDestination = filepath.Clean(destination)

	if len(Context.Targets) == 2 {
		Context.Targets = []string{Context.Targets[1]}
		return
	}

	Context.Targets = nil
}

func parseArguments(command string, subcommand string, args []string) structures.Context {
	rc := RC.OK
	Context := structures.Context{
		Verbose:    1,
		RC:         &rc,
		Exclusions: append([]string{}, consts.RequiredExclusions...),
		Configs:    map[string]structures.Config{},
	}

	for i := 0; i < len(args); i++ {
		if len(args[i]) == 0 {
			invalidArgument(&Context, args[i])
		}

		switch args[i][0] {
		case '-':
			parseFlagOrParameter(command, subcommand, args, &i, &Context)
		default:
			parseTarget(args, i, &Context)
		}
	}

	return Context
}

func parseTarget(args []string, i int, Context *structures.Context) {
	Context.Targets = append(Context.Targets, args[i])
}

func parseFlagOrParameter(command string, subcommand string, args []string, i *int, Context *structures.Context) {
	switch len(args[*i]) {
	case 1:
		invalidArgument(Context, args[*i])
	case 2:
		parseFlag(args, *i, Context)
	default:
		parseParameter(command, subcommand, args, i, Context)
	}
}

func parseFlag(args []string, i int, Context *structures.Context) {
	if args[i][1] == '-' {
		invalidArgument(Context, args[i])
	}

	switch args[i][1] {
	case 'a':
		Context.All = true
	case 'c':
		Context.Checkpoints = true
	case 'd':
		Context.Debug = true
	case 'f':
		Context.Force = true
	case 'h':
		Context.Help = true
	case 'i':
		Context.Install = true
	case 'l':
		Context.Local = true
	case 'm':
		Context.PrepareOnly = true
	case 'M':
		Context.DryRun = true
	case 's':
		Context.Verbose = 0
	case 't':
		Context.Tolerant = true
	case 'v':
		Context.Verbose = 3
	case 'V':
		Context.Verbose = 7
	default:
		invalidArgument(Context, args[i])
	}
}

func parseParameter(command string, subcommand string, args []string, i *int, Context *structures.Context) {
	if args[*i][1] != '-' {
		invalidArgument(Context, args[*i])
	}
	if *i+1 >= len(args) {
		missingParameterValue(Context, args[*i])
	}

	name := args[*i][2:]
	(*i)++
	value := args[*i]

	validateParameter(command, subcommand, Context, name, value)
}

func validateParameter(command string, subcommand string, Context *structures.Context, name string, value string) {
	switch name {
	case "exclude":
		processExclusions(Context, value)
	case "format":
		Context.Format = value
	case "message":
		Context.Message = value
	case "log":
		Context.LogDir = value
	case "path":
		Context.Path = value
	case "platform":
		Context.Platforms = []string{strings.ToLower(strings.TrimSpace(value))}
	case "source":
		if command != "workflow" || subcommand != "promote" {
			invalidArgument(Context, "--"+name)
		}
		Context.SourcePath = value
	case "dest":
		if command != "workflow" || subcommand != "promote" {
			invalidArgument(Context, "--"+name)
		}
		Context.DestinationPath = value
	case "version":
		if command != "workflow" || subcommand != "promote" {
			invalidArgument(Context, "--"+name)
		}
		Context.NextVersion = value
	default:
		invalidArgument(Context, "--"+name)
	}
}

func invalidArgument(Context *structures.Context, argument string) {
	cli.Error(RC.Error, *Context, "Argumento no válido: %q", argument)
}

func missingParameterValue(Context *structures.Context, parameter string) {
	cli.Error(RC.Error, *Context, "Falta el valor del parámetro: %q", parameter)
}

func Prepare(Context *structures.Context) {
	preparePlatforms(Context)
	processTargets(Context)
}

func preparePlatforms(Context *structures.Context) {
	if len(Context.Platforms) == 0 {
		Context.Platforms = []string{"windows", "linux"}
		return
	}

	for _, platform := range Context.Platforms {
		switch strings.ToLower(strings.TrimSpace(platform)) {
		case "windows", "linux":
		default:
			cli.Error(RC.Error, *Context, "Plataforma no soportada: %q", platform)
		}
	}
}

func processExclusions(Context *structures.Context, values string) {
	for _, value := range strings.Split(values, ",") {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}

		if info, err := os.Stat(value); err == nil && !info.IsDir() {
			addExclusionsFile(Context, value)
			continue
		}

		Context.Exclusions = append(Context.Exclusions, value)
	}

	Context.Exclusions = uniqueStrings(Context.Exclusions)
}

func addExclusionsFile(Context *structures.Context, path string) {
	file, err := os.Open(path)
	if err != nil {
		cli.Error(RC.Error, *Context, "No se puede leer el fichero de exclusiones: %q", path)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		value := strings.TrimSpace(scanner.Text())
		if value == "" {
			continue
		}
		Context.Exclusions = append(Context.Exclusions, value)
	}

	if err := scanner.Err(); err != nil {
		cli.Error(RC.Error, *Context, "Error leyendo el fichero de exclusiones: %q", path)
	}
}

func processTargets(Context *structures.Context) {
	roots := Context.Targets
	if len(roots) == 0 {
		roots = []string{"."}
	}

	scopes := []string{}
	Context.Targets = []string{}
	Context.TargetDetails = nil
	Context.Repos = []string{}
	Context.BlackList = []string{}
	Context.Configs = map[string]structures.Config{}

	for _, root := range roots {
		path, err := filepath.Abs(root)
		if err != nil {
			cli.Warning(*Context, "Se ignora %q: no se puede resolver la ruta.", root)
			continue
		}

		info, err := os.Stat(path)
		if err != nil || !info.IsDir() {
			cli.Warning(*Context, "Se ignora %q: no existe o no es un directorio.", root)
			continue
		}

		path = filepath.Clean(path)
		scopes = append(scopes, path)

		if repository := tools.FindRepo(path); repository != "" {
			Context.Repos = append(Context.Repos, repository)
			continue
		}
		discoverRepos(Context, path)
	}

	Context.Repos = uniqueStrings(Context.Repos)

	for _, scope := range uniqueStrings(scopes) {
		discoverIASITargets(Context, scope)
	}

	Context.Targets = uniqueStrings(Context.Targets)
	Context.TargetDetails = describeTargets(Context, Context.Targets)
}

func describeTargets(Context *structures.Context, targets []string) []structures.Target {
	details := make([]structures.Target, 0, len(targets))

	for _, path := range targets {
		path = filepath.Clean(path)

		config, err := iasiconfig.Read(path)
		if err != nil {
			cli.Error(RC.Error, *Context, "Configuración IASI inválida en %q: %v", path, err)
		}

		Context.Configs[path] = config

		detail := structures.Target{
			Path:       path,
			Repository: tools.FindRepo(path),
		}

		for _, ancestor := range targets {
			if isAncestorTarget(ancestor, path) {
				detail.Depth++
			}
		}

		details = append(details, detail)
	}

	return details
}

func defaultTargetName(name string) string {
	original := name
	i := 0

	for i < len(name) && name[i] >= '0' && name[i] <= '9' {
		i++
	}
	if i == 0 {
		return name
	}

	for i < len(name) && (name[i] == '-' || name[i] == '_' || name[i] == '.' || name[i] == ' ') {
		i++
	}
	if i >= len(name) {
		return original
	}

	return name[i:]
}

func isAncestorTarget(parent string, child string) bool {
	parent = filepath.Clean(parent)
	child = filepath.Clean(child)

	if parent == child {
		return false
	}

	relative, err := filepath.Rel(parent, child)
	if err != nil || relative == "." || relative == ".." {
		return false
	}

	return !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func targetType(Context structures.Context, path string) string {
	value := Context.Config(path).Type()
	if value == "" {
		return "none"
	}
	return value
}

func discoverIASITargets(Context *structures.Context, path string) {
	if isExcluded(Context, filepath.Base(path)) {
		return
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		cli.Warning(*Context, "Se ignora %q: no se puede leer.", path)
		return
	}

	hasTOML := false
	hasQuarto := false
	legacy := []string{}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		switch {
		case strings.EqualFold(entry.Name(), iasiconfig.FileName):
			hasTOML = true
		case strings.EqualFold(entry.Name(), "_quarto.yml"):
			hasQuarto = true
		default:
			for _, name := range iasiconfig.LegacyFileNames {
				if strings.EqualFold(entry.Name(), name) {
					legacy = append(legacy, entry.Name())
					break
				}
			}
		}
	}

	if hasTOML {
		Context.Targets = append(Context.Targets, filepath.Clean(path))

		if len(legacy) != 0 {
			cli.Attention(
				*Context,
				"Multiple IASI configuration files found in: %s. Using iasi.toml; legacy files: %s",
				path,
				strings.Join(legacy, ", "),
			)
		}
	} else if len(legacy) != 0 || hasQuarto {
		cli.Attention(*Context, "Missing iasi.toml in: %s", path)
	}

	for _, entry := range entries {
		if !entry.IsDir() || isExcluded(Context, entry.Name()) {
			continue
		}
		discoverIASITargets(Context, filepath.Join(path, entry.Name()))
	}
}

func discoverRepos(Context *structures.Context, path string) {
	if isExcluded(Context, filepath.Base(path)) {
		return
	}

	if tools.IsRepo(path) {
		Context.Repos = append(Context.Repos, filepath.Clean(path))
		return
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		cli.Warning(*Context, "Se ignora %q: no se puede leer.", path)
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() || isExcluded(Context, entry.Name()) {
			continue
		}
		discoverRepos(Context, filepath.Join(path, entry.Name()))
	}
}

func isExcluded(Context *structures.Context, name string) bool {
	for _, exclusion := range Context.Exclusions {
		if name == exclusion {
			return true
		}
	}
	return false
}

func uniqueStrings(values []string) []string {
	unique := []string{}
	seen := map[string]bool{}

	for _, value := range values {
		key := filepath.Clean(value)
		if seen[key] {
			continue
		}

		seen[key] = true
		unique = append(unique, value)
	}

	return unique
}
