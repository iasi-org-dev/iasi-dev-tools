package structures

import "path/filepath"

func (context Context) Config(path string) Config {
	if context.Configs == nil {
		return Config{}
	}
	return context.Configs[filepath.Clean(path)]
}
