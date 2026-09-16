package args

import (
	"bufio"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf16"

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

	Parms := parseArguments(values)
	Parms.Subcommand = subcommand
	validateCheckModes(&Parms)
	validatePushMode(command, &Parms)
	extractTargetVersion(command, &Parms)
	extractVersionOrganization(command, &Parms)
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

func validatePushMode(command string, Parms *structures.Parms) {
	isPromote := command == "promote" || (command == "workflow" && Parms.Subcommand == "promote")
	if !Parms.Push {
		return
	}
	if !isPromote {
		cli.Error(RC.Error, *Parms, "-p solo está soportado por promote y workflow promote.")
	}
	if Parms.Local {
		cli.Error(RC.Error, *Parms, "-p y -l son incompatibles.")
	}
	if len(Parms.Targets) != 0 {
		cli.Error(RC.Error, *Parms, "-p no acepta versión ni targets: solo publica el estado local existente.")
	}
}

// extractTargetVersion separates organization version operands from filesystem targets.
func extractTargetVersion(command string, Parms *structures.Parms) {
	requiresVersion := command == "promote" || command == "restore" || (command == "workflow" && Parms.Subcommand == "promote")
	if !requiresVersion || len(Parms.Targets) == 0 {
		return
	}

	Parms.TargetVersion = Parms.Targets[0]
	Parms.Targets = Parms.Targets[1:]
}

// extractVersionOrganization separates the optional organization operand from filesystem targets.
func extractVersionOrganization(command string, Parms *structures.Parms) {
	if command != "version" || len(Parms.Targets) == 0 {
		return
	}
	if len(Parms.Targets) > 1 {
		cli.Error(RC.Error, *Parms, "version acepta como máximo una organización.")
	}

	Parms.Organization = Parms.Targets[0]
	Parms.Targets = nil
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

func parseArguments(args []string) structures.Parms {
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
			parseFlagOrParameter(args, &i, &Parms)
		default:
			parseTarget(args, i, &Parms)
		}
	}

	return Parms
}

func parseTarget(args []string, i int, Parms *structures.Parms) {
	Parms.Targets = append(Parms.Targets, args[i])
}

func parseFlagOrParameter(args []string, i *int, Parms *structures.Parms) {
	switch len(args[*i]) {
	case 1:
		invalidArgument(Parms, args[*i])
	case 2:
		parseFlag(args, *i, Parms)
	default:
		parseParameter(args, i, Parms)
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
	case 'p':
		Parms.Push = true
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

func parseParameter(args []string, i *int, Parms *structures.Parms) {
	if args[*i][1] != '-' {
		invalidArgument(Parms, args[*i])
	}
	if *i+1 >= len(args) {
		missingParameterValue(Parms, args[*i])
	}

	name := args[*i][2:]
	(*i)++
	value := args[*i]

	validateParameter(Parms, name, value)
}

func validateParameter(Parms *structures.Parms, name string, value string) {
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
	processTargets(Parms)
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

// describeTargets reads the literal type from each IASI marker and derives the
// project hierarchy from discovered ancestor targets.
func describeTargets(targets []string) []structures.Target {
	details := make([]structures.Target, 0, len(targets))
	for _, path := range targets {
		detail := structures.Target{
			Path:       filepath.Clean(path),
			Type:       targetType(path),
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

// targetType returns the literal type value from the target's IASI marker.
// Both .iasi.yml and _iasi.yml are valid markers. Missing or empty type values
// are reported as "none".
func targetType(path string) string {
	for _, name := range []string{".iasi.yml", "_iasi.yml"} {
		marker := filepath.Join(path, name)
		info, err := os.Stat(marker)
		if err != nil || info.IsDir() {
			continue
		}
		if value := readTargetType(marker); value != "" {
			return value
		}
		return "none"
	}
	return "none"
}

func readTargetType(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	scanner := bufio.NewScanner(strings.NewReader(decodeIASIMarker(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(strings.TrimPrefix(scanner.Text(), "\ufeff"))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		separator := strings.Index(line, ":")
		if separator < 0 || !strings.EqualFold(strings.TrimSpace(line[:separator]), "type") {
			continue
		}
		return strings.TrimSpace(line[separator+1:])
	}
	return ""
}

// decodeIASIMarker accepts the UTF-8 used normally by IASI and also UTF-16
// files that can be produced by Windows tooling such as legacy PowerShell.
func decodeIASIMarker(data []byte) string {
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return string(data[3:])
	}
	if len(data) >= 2 && data[0] == 0xFF && data[1] == 0xFE {
		return decodeUTF16(data[2:], binary.LittleEndian)
	}
	if len(data) >= 2 && data[0] == 0xFE && data[1] == 0xFF {
		return decodeUTF16(data[2:], binary.BigEndian)
	}
	if order, ok := utf16ByteOrderWithoutBOM(data); ok {
		return decodeUTF16(data, order)
	}
	return string(data)
}

// utf16ByteOrderWithoutBOM recognises the common ASCII-heavy UTF-16 layout
// used by YAML even when the file has no byte-order mark.
func utf16ByteOrderWithoutBOM(data []byte) (binary.ByteOrder, bool) {
	pairs := len(data) / 2
	if pairs < 2 {
		return nil, false
	}

	evenZeros := 0
	oddZeros := 0
	for i := 0; i < pairs*2; i += 2 {
		if data[i] == 0 {
			evenZeros++
		}
		if data[i+1] == 0 {
			oddZeros++
		}
	}

	threshold := pairs / 2
	if oddZeros > threshold && evenZeros <= threshold/2 {
		return binary.LittleEndian, true
	}
	if evenZeros > threshold && oddZeros <= threshold/2 {
		return binary.BigEndian, true
	}
	return nil, false
}

func decodeUTF16(data []byte, order binary.ByteOrder) string {
	if len(data)%2 != 0 {
		data = data[:len(data)-1]
	}
	units := make([]uint16, len(data)/2)
	for i := range units {
		units[i] = order.Uint16(data[i*2 : i*2+2])
	}
	return string(utf16.Decode(units))
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
