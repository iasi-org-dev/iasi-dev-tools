package structures

import "strings"

type Config struct {
	IASI map[string]any
}

func (config Config) Value(path ...string) any {
	var current any = config.IASI

	for _, name := range path {
		table, ok := current.(map[string]any)
		if !ok {
			return nil
		}

		current, ok = table[name]
		if !ok {
			return nil
		}
	}

	return current
}

func (config Config) String(path ...string) string {
	value := config.Value(path...)
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(text)
}

func (config Config) StringDefault(fallback string, path ...string) string {
	value := config.String(path...)
	if value == "" {
		return fallback
	}
	return value
}

func (config Config) Type() string {
	return config.String("type")
}

func (config Config) SoftwareBuilder() string {
	return config.String("software", "builder")
}

func (config Config) SoftwareName(fallback string) string {
	return config.StringDefault(fallback, "software", "name")
}

func (config Config) SoftwareInputDir() string {
	return config.StringDefault("src", "software", "input-dir")
}

func (config Config) SoftwareOutputDir() string {
	return config.StringDefault("_outputs", "software", "output-dir")
}
