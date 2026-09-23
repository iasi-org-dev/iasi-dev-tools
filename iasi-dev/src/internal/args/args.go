package args

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"iasi-dev/internal/cli"
	"iasi-dev/internal/consts"
	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
	"iasi-dev/internal/tools"
)

// Parse parses command-line arguments. Preparation is performed later by main.
func Parse(command string, values []string) structures.Parms {
	subcommand := ""

	if command == "workflow" {
		if len(values) == 0 {
			rc := RC.OK
			cli.Error(RC.Error, structures.Parms{RC: &rc}, "Falta el comando del workflow.")
		}

		subcommand = values[0]
		values = values[1:]
	}

	Parms := parseArguments(command, subcommand, values)
	Parms.Subcommand = subcommand
	validateCheckModes(&Parms)
	validateLocalMode(command, &Parms)
	extractPromoteOperands(command, &Parms)
	extractWorkflowPromoteParameters(command, &Parms)
	extractTargetVersion(command, &Parms)
	extractVersionOperands(command, &Parms)
	validateOrganizationWideCommand(command, &Parms)
	extractMaterializeOperands(command, &Parms)
	if Parms.RequestedTargets == nil {
		Parms.RequestedTargets = append([]string{}, Parms.Targets...)
	}

	return Parms
}

func validateCheckModes(Parms *structures.Parms) {
	if Parms.PrepareOnly && Parms.DryRun {
		cli.Error(RC.Error, *Parms, "-m y -M son incompatibles.")
	}
}

func validateLocalMode(command string, Parms *structures.Parms) {
	if command == "freeze" && Parms.Local {
		cli.Error(RC.Error, *Parms, "-l no está soportado por freeze: freeze siempre publica los tags.")
	}
}

// extractPromoteOperands requires the complete promotion contract explicitly:
//   promote <version> <source-path> <destination-path>
// Relative paths are resolved later against the effective working directory, after --path if present.
func extractPromoteOperands(command string, Parms *structures.Parms) {
	isPromote := command == "promote" || command == "promote-check"
	if !isPromote {
		return
	}
	if len(Parms.Targets) != 3 {
		cli.Error(RC.Error, *Parms, "promote requiere <version> <source-path> <destination-path>.")
	}

	version := strings.TrimSpace(Parms.Targets[0])
	source := strings.TrimSpace(Parms.Targets[1])
	destination := strings.TrimSpace(Parms.Targets[2])
	if !looksLikeSemanticVersion(version) {
		cli.Error(RC.Error, *Parms, "La versión no es válida: %s", version)
	}
	if source == "" || destination == "" {
		cli.Error(RC.Error, *Parms, "source-path y destination-path son obligatorios.")
	}

	sourcePath := filepath.Clean(source)
	destinationPath := filepath.Clean(destination)
	if strings.EqualFold(sourcePath, destinationPath) {
		cli.Error(RC.Error, *Parms, "source-path y destination-path deben ser rutas distintas.")
	}

	Parms.TargetVersion = version
	Parms.SourcePath = sourcePath
	Parms.DestinationPath = destinationPath
	Parms.DestinationOrganization = filepath.Base(destinationPath)
	Parms.Targets = nil
}


// extractWorkflowPromoteParameters requires the orchestration contract explicitly:
//   workflow promote --source <source-path> --dest <destination-path> --version vMAJOR.MINOR.PATCH
// --version is the next development version; the version promoted is the source VERSION read during preparation.
func extractWorkflowPromoteParameters(command string, Parms *structures.Parms) {
	if command != "workflow" || Parms.Subcommand != "promote" {
		return
	}
	if len(Parms.Targets) != 0 {
		cli.Error(RC.Error, *Parms, "workflow promote usa parámetros nombrados: --source, --dest y --version.")
	}
	if strings.TrimSpace(Parms.SourcePath) == "" {
		cli.Error(RC.Error, *Parms, "workflow promote requiere --source <source-path>.")
	}
	if strings.TrimSpace(Parms.DestinationPath) == "" {
		cli.Error(RC.Error, *Parms, "workflow promote requiere --dest <destination-path>.")
	}
	if strings.TrimSpace(Parms.NextVersion) == "" {
		cli.Error(RC.Error, *Parms, "workflow promote requiere --version vMAJOR.MINOR.PATCH.")
	}
	if !looksLikeSemanticVersion(Parms.NextVersion) {
		cli.Error(RC.Error, *Parms, "La nueva versión no es válida: %s", Parms.NextVersion)
	}

	sourcePath := filepath.Clean(Parms.SourcePath)
	destinationPath := filepath.Clean(Parms.DestinationPath)
	if strings.EqualFold(sourcePath, destinationPath) {
		cli.Error(RC.Error, *Parms, "source-path y destination-path deben ser rutas distintas.")
	}

	Parms.SourcePath = sourcePath
	Parms.DestinationPath = destinationPath
	Parms.DestinationOrganization = filepath.Base(destinationPath)
}

// extractTargetVersion separates restore's optional version operand from filesystem targets.
func extractTargetVersion(command string, Parms *structures.Parms) {
	requiresVersion := command == "restore"
	if !requiresVersion || len(Parms.Targets) == 0 {
		return
	}

	Parms.TargetVersion = Parms.Targets[0]
	Parms.Targets = Parms.Targets[1:]
}

// extractVersionOperands supports querying VERSION or explicitly setting it:
//   version
//   version v0.6.0
func extractVersionOperands(command string, Parms *structures.Parms) {
	if command != "version" || len(Parms.Targets) == 0 {
		return
	}
	if len(Parms.Targets) > 1 {
		cli.Error(RC.Error, *Parms, "version acepta como máximo una versión vMAJOR.MINOR.PATCH.")
	}

	version := Parms.Targets[0]
	if !looksLikeSemanticVersion(version) {
		cli.Error(RC.Error, *Parms, "La versión no es válida: %s", version)
	}

	Parms.TargetVersion = version
	Parms.Targets = nil
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

func validateOrganizationWideCommand(command string, Parms *structures.Parms) {
	organizationWide := command == "freeze" || command == "promote" || command == "promote-check" || (command == "workflow" && Parms.Subcommand == "promote")
	if !organizationWide {
		return
	}
	if len(Parms.Targets) != 0 {
		cli.Error(RC.Error, *Parms, "%s opera sobre la organización completa y no acepta targets; usa --path para elegir el workspace.", command)
	}
}

// extractMaterializeOperands separates materialize's destination from the optional source target.
// The destination is resolved before --path changes the working directory. The optional source
// remains a normal target and therefore follows the common preparation flow.
func extractMaterializeOperands(command string, Parms *structures.Parms) {
	if command != "materialize" {
		return
	}

	Parms.RequestedTargets = append([]string{}, Parms.Targets...)
	if len(Parms.Targets) == 0 {
		return
	}
	if len(Parms.Targets) > 2 {
		cli.Error(RC.Error, *Parms, "materialize acepta <destino> y, opcionalmente, [origen].")
	}

	destination, err := filepath.Abs(Parms.Targets[0])
	if err != nil {
		cli.Error(RC.Error, *Parms, "No se puede resolver el destino de materialize: %q", Parms.Targets[0])
	}
	Parms.MaterializeDestination = filepath.Clean(destination)

	if len(Parms.Targets) == 2 {
		Parms.Targets = []string{Parms.Targets[1]}
		return
	}

	// No explicit source: common preparation will use the current directory.
	Parms.Targets = nil
}

func parseArguments(command string, subcommand string, args []string) structures.Parms {
	rc := RC.OK
	Parms := structures.Parms{
		Verbose:    1,
		RC:         &rc,
		Exclusions: append([]string{}, consts.RequiredExclusions...),
	}

	for i := 0; i < len(args); i++ {
		if len(args[i]) == 0 {
			invalidArgument(&Parms, args[i])
		}

		switch args[i][0] {
		case '-':
			parseFlagOrParameter(command, subcommand, args, &i, &Parms)
		default:
			parseTarget(args, i, &Parms)
		}
	}

	return Parms
}

func parseTarget(args []string, i int, Parms *structures.Parms) {
	Parms.Targets = append(Parms.Targets, args[i])
}

func parseFlagOrParameter(command string, subcommand string, args []string, i *int, Parms *structures.Parms) {
	switch len(args[*i]) {
	case 1:
		invalidArgument(Parms, args[*i])
	case 2:
		parseFlag(args, *i, Parms)
	default:
		parseParameter(command, subcommand, args, i, Parms)
	}
}

func parseFlag(args []string, i int, Parms *structures.Parms) {
	if args[i][1] == '-' {
		invalidArgument(Parms, args[i])
	}

	switch args[i][1] {
	case 'a':
		Parms.All = true
	case 'c':
		Parms.Checkpoints = true
	case 'd':
		Parms.Debug = true
	case 'f':
		Parms.Force = true
	case 'h':
		Parms.Help = true
	case 'i':
		Parms.Install = true
	case 'l':
		Parms.Local = true
	case 'm':
		Parms.PrepareOnly = true
	case 'M':
		Parms.DryRun = true
	case 's':
		Parms.Verbose = 0
	case 't':
		Parms.Tolerant = true
	case 'v':
		Parms.Verbose = 3
	case 'V':
		Parms.Verbose = 7
	default:
		invalidArgument(Parms, args[i])
	}
}

func parseParameter(command string, subcommand string, args []string, i *int, Parms *structures.Parms) {
	if args[*i][1] != '-' {
		invalidArgument(Parms, args[*i])
	}
	if *i+1 >= len(args) {
		missingParameterValue(Parms, args[*i])
	}

	name := args[*i][2:]
	(*i)++
	value := args[*i]

	validateParameter(command, subcommand, Parms, name, value)
}

func validateParameter(command string, subcommand string, Parms *structures.Parms, name string, value string) {
	switch name {
	case "exclude":
		processExclusions(Parms, value)
	case "format":
		Parms.Format = value
	case "message":
		Parms.Message = value
	case "log":
		Parms.LogDir = value
	case "path":
		Parms.Path = value
	case "platform":
		Parms.Platforms = []string{strings.ToLower(strings.TrimSpace(value))}
	case "source":
		if command != "workflow" || subcommand != "promote" {
			invalidArgument(Parms, "--"+name)
		}
		Parms.SourcePath = value
	case "dest":
		if command != "workflow" || subcommand != "promote" {
			invalidArgument(Parms, "--"+name)
		}
		Parms.DestinationPath = value
	case "version":
		if command != "workflow" || subcommand != "promote" {
			invalidArgument(Parms, "--"+name)
		}
		Parms.NextVersion = value
	default:
		invalidArgument(Parms, "--"+name)
	}
}

func invalidArgument(Parms *structures.Parms, argument string) {
	cli.Error(RC.Error, *Parms, "Argumento no válido: %q", argument)
}

func missingParameterValue(Parms *structures.Parms, parameter string) {
	cli.Error(RC.Error, *Parms, "Falta el valor del parámetro: %q", parameter)
}

// Prepare resolves the requested scope, discovers Git repositories and IASI targets.
func Prepare(Parms *structures.Parms) {
	preparePlatforms(Parms)
	processTargets(Parms)
}

func preparePlatforms(Parms *structures.Parms) {
	if len(Parms.Platforms) == 0 {
		Parms.Platforms = []string{"windows", "linux"}
		return
	}

	for _, platform := range Parms.Platforms {
		switch strings.ToLower(strings.TrimSpace(platform)) {
		case "windows", "linux":
		default:
			cli.Error(RC.Error, *Parms, "Plataforma no soportada: %q", platform)
		}
	}
}

func processExclusions(Parms *structures.Parms, values string) {
	for _, value := range strings.Split(values, ",") {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}

		if info, err := os.Stat(value); err == nil && !info.IsDir() {
			addExclusionsFile(Parms, value)
			continue
		}

		Parms.Exclusions = append(Parms.Exclusions, value)
	}

	Parms.Exclusions = uniqueStrings(Parms.Exclusions)
}

func addExclusionsFile(Parms *structures.Parms, path string) {
	file, err := os.Open(path)
	if err != nil {
		cli.Error(RC.Error, *Parms, "No se puede leer el fichero de exclusiones: %q", path)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		value := strings.TrimSpace(scanner.Text())
		if value == "" {
			continue
		}
		Parms.Exclusions = append(Parms.Exclusions, value)
	}

	if err := scanner.Err(); err != nil {
		cli.Error(RC.Error, *Parms, "Error leyendo el fichero de exclusiones: %q", path)
	}
}

// processTargets resolves the requested scope, discovers Git repositories and then
// discovers IASI targets below that same scope from ?iasi.yml markers.
func processTargets(Parms *structures.Parms) {
	roots := Parms.Targets
	if len(roots) == 0 {
		roots = []string{"."}
	}

	scopes := []string{}
	Parms.Targets = []string{}
	Parms.Repos = []string{}
	Parms.BlackList = []string{}

	for _, root := range roots {
		path, err := filepath.Abs(root)
		if err != nil {
			cli.Warning(*Parms, "Se ignora %q: no se puede resolver la ruta.", root)
			continue
		}

		info, err := os.Stat(path)
		if err != nil || !info.IsDir() {
			cli.Warning(*Parms, "Se ignora %q: no existe o no es un directorio.", root)
			continue
		}

		path = filepath.Clean(path)
		scopes = append(scopes, path)

		if repository := tools.FindRepo(path); repository != "" {
			Parms.Repos = append(Parms.Repos, repository)
			continue
		}
		discoverRepos(Parms, path)
	}

	Parms.Repos = uniqueStrings(Parms.Repos)
	for _, scope := range uniqueStrings(scopes) {
		discoverIASITargets(Parms, scope)
	}
	Parms.Targets = uniqueStrings(Parms.Targets)
	Parms.TargetDetails = describeTargets(Parms.Targets)
}

// describeTargets reads the small flat target configuration needed by iasi-dev
// and derives the project hierarchy from discovered ancestor targets.
func describeTargets(targets []string) []structures.Target {
	details := make([]structures.Target, 0, len(targets))
	for _, path := range targets {
		config := readTargetConfig(path)
		targetType := strings.TrimSpace(config["type"])
		if targetType == "" {
			targetType = "none"
		}

		detail := structures.Target{
			Path:       filepath.Clean(path),
			Type:       targetType,
			Builder:    strings.TrimSpace(config["builder"]),
			Repository: tools.FindRepo(path),
		}

		if strings.EqualFold(targetType, "software") {
			detail.SourceDir = configValue(config, "source-dir", "src")
			detail.OutputDir = configValue(config, "output-dir", "_outputs")
			detail.Name = configValue(config, "name", defaultTargetName(filepath.Base(path)))
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

func configValue(config map[string]string, name string, fallback string) string {
	value := strings.TrimSpace(config[name])
	if value == "" {
		return fallback
	}
	return value
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

// readTargetConfig intentionally reads only the flat scalar keys currently
// needed by iasi-dev. The on-disk configuration format is transitional.
func readTargetConfig(path string) map[string]string {
	config := map[string]string{}
	marker := targetMarker(path)
	if marker == "" {
		return config
	}

	file, err := os.Open(marker)
	if err != nil {
		return config
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(strings.TrimPrefix(scanner.Text(), "\ufeff"))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])
		if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
			value = value[1 : len(value)-1]
		}
		config[key] = value
	}
	return config
}

func targetType(path string) string {
	value := strings.TrimSpace(readTargetConfig(path)["type"])
	if value == "" {
		return "none"
	}
	return value
}

func targetMarker(path string) string {
	for _, name := range []string{".iasi.yml", "_iasi.yml"} {
		marker := filepath.Join(path, name)
		if info, err := os.Stat(marker); err == nil && !info.IsDir() {
			return marker
		}
	}
	return ""
}

// discoverIASITargets recursively finds directories containing a ?iasi.yml marker.
func discoverIASITargets(Parms *structures.Parms, path string) {
	if isExcluded(Parms, filepath.Base(path)) {
		return
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		cli.Warning(*Parms, "Se ignora %q: no se puede leer.", path)
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() && isIASIMarker(entry.Name()) {
			Parms.Targets = append(Parms.Targets, filepath.Clean(path))
			break
		}
	}

	for _, entry := range entries {
		if !entry.IsDir() || isExcluded(Parms, entry.Name()) {
			continue
		}
		discoverIASITargets(Parms, filepath.Join(path, entry.Name()))
	}
}

func isIASIMarker(name string) bool {
	matched, _ := filepath.Match("?iasi.yml", strings.ToLower(name))
	return matched
}

// discoverRepos recursively discovers Git repositories below path.
func discoverRepos(Parms *structures.Parms, path string) {
	if isExcluded(Parms, filepath.Base(path)) {
		return
	}

	if tools.IsRepo(path) {
		Parms.Repos = append(Parms.Repos, filepath.Clean(path))
		return
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		cli.Warning(*Parms, "Se ignora %q: no se puede leer.", path)
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() || isExcluded(Parms, entry.Name()) {
			continue
		}
		discoverRepos(Parms, filepath.Join(path, entry.Name()))
	}
}

func isExcluded(Parms *structures.Parms, name string) bool {
	for _, exclusion := range Parms.Exclusions {
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
