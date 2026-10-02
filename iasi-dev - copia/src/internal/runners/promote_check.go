package runners

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"iasi-dev/internal/cli"
	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

var promoteCheckExcludedRepositories = map[string]bool{
	"iasi-dev-tools": true,
}

type promoteCheckPattern struct {
	label string
	value string
}

type promoteCheckHit struct {
	path    string
	line    int
	text    string
	matches []string
}

// PromoteCheck scans the promotion destination for references to the source that may require postprocessing.
// It never modifies the source, destination or Git repositories.
func PromoteCheck(Parms *structures.Parms) []string {
	requireTargetVersion(Parms, "promote-check")
	if _, ok := parseSemanticVersion(Parms.TargetVersion); !ok {
		cli.Error(RC.Error, *Parms, "La versión a comprobar no es válida: %s", Parms.TargetVersion)
	}
	if Parms.SourcePath == "" || Parms.DestinationPath == "" {
		cli.Error(RC.Error, *Parms, "promote-check requiere source-path y destination-path explícitos.")
	}

	source := filepath.Clean(Parms.SourcePath)
	destination := filepath.Clean(Parms.DestinationPath)
	validatePromotePaths(Parms, source, destination)

	cli.Header(*Parms, "Promote check %s", Parms.TargetVersion)
	cli.Info(*Parms, "Source path: %s", source)
	cli.Info(*Parms, "Destination path: %s", destination)
	cli.Info(*Parms, "Scanning suspicious references...")

	hits, skipped, err := promoteCheckScanPromotion(source, destination)
	if err != nil {
		cli.Error(RC.Error, *Parms, "No se pudo completar el escaneo: %v", err)
	}

	if len(hits) == 0 {
		cli.Success(*Parms, "No suspicious references found. No files modified.")
		return nil
	}

	for _, hit := range hits {
		relative, err := filepath.Rel(destination, hit.path)
		if err != nil {
			relative = hit.path
		}
		location := relative
		if hit.line > 0 {
			location = fmt.Sprintf("%s:%d", relative, hit.line)
		}
		cli.Info(*Parms, "%s [%s]", location, strings.Join(hit.matches, ", "))
		if hit.text != "" {
			cli.Direct("    %s\n", hit.text)
		}
	}

	cli.Info(*Parms, "%d suspicious reference(s) found. No files modified.", len(hits))
	if skipped > 0 {
		cli.Info(*Parms, "Skipped binary/unreadable files: %d", skipped)
	}
	return nil
}

func promoteCheckScanPromotion(source string, destination string) ([]promoteCheckHit, int, error) {
	return promoteCheckScan(destination, promoteCheckPatterns(source))
}

func promoteCheckPatterns(source string) []promoteCheckPattern {
	values := []promoteCheckPattern{}
	seen := map[string]bool{}
	add := func(label string, value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		key := strings.ToLower(value)
		if seen[key] {
			return
		}
		seen[key] = true
		values = append(values, promoteCheckPattern{label: label, value: value})
	}

	add("source-name", filepath.Base(source))
	add("source-path", source)
	add("source-path", filepath.ToSlash(source))
	add("source-path", strings.ReplaceAll(source, string(filepath.Separator), "\\"))

	return values
}

func promoteCheckScan(root string, patterns []promoteCheckPattern) ([]promoteCheckHit, int, error) {
	hits := []promoteCheckHit{}
	skipped := 0

	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			skipped++
			return nil
		}

		if entry.IsDir() && promoteCheckExcludedRepository(root, path, entry.Name()) {
			return filepath.SkipDir
		}
		if entry.IsDir() && promoteCheckIgnoredDir(entry.Name()) {
			return filepath.SkipDir
		}

		if path != root {
			if nameMatches := promoteCheckMatches(entry.Name(), patterns); len(nameMatches) > 0 {
				hits = append(hits, promoteCheckHit{path: path, matches: nameMatches})
			}
		}

		if entry.IsDir() {
			if entry.Name() == ".git" {
				config := filepath.Join(path, "config")
				if info, err := os.Stat(config); err == nil && !info.IsDir() {
					fileHits, binary, err := promoteCheckScanFile(config, patterns)
					if err != nil || binary {
						skipped++
					} else {
						hits = append(hits, fileHits...)
					}
				}
				return filepath.SkipDir
			}
			return nil
		}

		fileHits, binary, err := promoteCheckScanFile(path, patterns)
		if err != nil || binary {
			skipped++
			return nil
		}
		hits = append(hits, fileHits...)
		return nil
	})
	if err != nil {
		return nil, skipped, err
	}

	sort.SliceStable(hits, func(i int, j int) bool {
		if hits[i].path == hits[j].path {
			return hits[i].line < hits[j].line
		}
		return hits[i].path < hits[j].path
	})
	return hits, skipped, nil
}


func promoteCheckExcludedRepository(root string, path string, name string) bool {
	if strings.EqualFold(filepath.Clean(path), filepath.Clean(root)) {
		return false
	}
	if !strings.EqualFold(filepath.Clean(filepath.Dir(path)), filepath.Clean(root)) {
		return false
	}
	for repository := range promoteCheckExcludedRepositories {
		if strings.EqualFold(name, repository) {
			return true
		}
	}
	return false
}

func promoteCheckIgnoredDir(name string) bool {
	return strings.EqualFold(name, "logs")
}

func promoteCheckScanFile(path string, patterns []promoteCheckPattern) ([]promoteCheckHit, bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, false, err
	}
	defer file.Close()

	binary, err := promoteCheckBinary(file)
	if err != nil {
		return nil, false, err
	}
	if binary {
		return nil, true, nil
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, false, err
	}

	hits := []promoteCheckHit{}
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 16*1024*1024)
	line := 0
	for scanner.Scan() {
		line++
		text := strings.TrimSpace(scanner.Text())
		matches := promoteCheckMatches(text, patterns)
		if len(matches) == 0 {
			continue
		}
		hits = append(hits, promoteCheckHit{
			path:    path,
			line:    line,
			text:    promoteCheckSnippet(text),
			matches: matches,
		})
	}
	if err := scanner.Err(); err != nil {
		return hits, false, err
	}
	return hits, false, nil
}

func promoteCheckBinary(file *os.File) (bool, error) {
	buffer := make([]byte, 8192)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return false, err
	}
	for _, value := range buffer[:n] {
		if value == 0 {
			return true, nil
		}
	}
	return false, nil
}

func promoteCheckMatches(text string, patterns []promoteCheckPattern) []string {
	lower := strings.ToLower(text)
	matches := []string{}
	seen := map[string]bool{}
	for _, pattern := range patterns {
		if !strings.Contains(lower, strings.ToLower(pattern.value)) || seen[pattern.label] {
			continue
		}
		seen[pattern.label] = true
		matches = append(matches, pattern.label)
	}
	return matches
}

func promoteCheckSnippet(text string) string {
	const limit = 240
	text = strings.TrimSpace(text)
	if len(text) <= limit {
		return text
	}
	return text[:limit] + "..."
}
