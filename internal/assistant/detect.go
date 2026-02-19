package assistant

import (
	"os"
	"path/filepath"
	"sort"
)

// Detect scans the target directory for known assistant configuration files
// and directories, returning the names of all assistants whose markers are
// found. The result is sorted alphabetically.
//
// Detection checks each registered assistant's InstructionsPath and AgentDir
// relative to target. If either exists (file or directory), that assistant
// is included in the result.
func Detect(target string) []string {
	var found []string

	for _, name := range List() {
		cfg := configs[name]
		if assistantPresent(target, &cfg) {
			found = append(found, name)
		}
	}

	sort.Strings(found)
	return found
}

// assistantPresent returns true if any of the assistant's marker paths
// exist within target.
func assistantPresent(target string, cfg *Config) bool {
	markers := []string{
		cfg.InstructionsPath,
		cfg.AgentDir,
		cfg.MCPConfigPath,
	}

	for _, marker := range markers {
		if marker == "" {
			continue
		}
		path := filepath.Join(target, marker)
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}

	return false
}
