package mcp

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/BurntSushi/toml"
)

// UninstallResult captures what happened during MCP server removal for a
// single package.
type UninstallResult struct {
	ConfigPath string   // Relative config file path
	Removed    []string // Server names that were removed
	NotFound   []string // Server names that weren't in the config
	Warnings   []string
}

// Uninstall removes the named MCP servers from the assistant's config file.
//
// It reads the existing config, removes servers whose names appear in
// serverNames, and writes the result. If no servers match, the file is
// left untouched.
//
// Returns (nil, nil) when the config file does not exist — nothing to undo.
func Uninstall(targetDir string, translator Translator, serverNames []string, dryRun bool) (*UninstallResult, error) {
	if len(serverNames) == 0 {
		return nil, nil
	}

	configPath := filepath.Join(targetDir, translator.ConfigPath())
	data, err := os.ReadFile(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read %s: %w", translator.ConfigPath(), err)
	}

	useTOML := isTOMLFile(translator.ConfigPath())

	var doc map[string]any
	if useTOML {
		if err := toml.Unmarshal(data, &doc); err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", translator.ConfigPath(), err)
		}
	} else {
		if err := json.Unmarshal(data, &doc); err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", translator.ConfigPath(), err)
		}
	}

	result := &UninstallResult{
		ConfigPath: translator.ConfigPath(),
	}

	// Get the servers section
	configKey := translator.ConfigKey()
	raw, ok := doc[configKey]
	if !ok {
		// No MCP section at all — all servers are "not found"
		result.NotFound = append(result.NotFound, serverNames...)
		return result, nil
	}
	servers, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s: %q key is not an object", translator.ConfigPath(), configKey)
	}

	// Sort for deterministic output
	sort.Strings(serverNames)

	for _, name := range serverNames {
		if _, exists := servers[name]; exists {
			delete(servers, name)
			result.Removed = append(result.Removed, name)
		} else {
			result.NotFound = append(result.NotFound, name)
		}
	}

	if len(result.Removed) == 0 {
		// Nothing changed — don't rewrite the file
		return result, nil
	}

	// Update the doc
	if len(servers) == 0 {
		// Remove the entire MCP key if no servers remain
		delete(doc, configKey)
	} else {
		doc[configKey] = servers
	}

	if !dryRun {
		// If the doc is now empty (only had MCP servers), remove the file
		// and attempt to clean up the parent directory if it is now empty.
		if len(doc) == 0 {
			if err := os.Remove(configPath); err != nil && !errors.Is(err, os.ErrNotExist) {
				return nil, fmt.Errorf("failed to remove empty %s: %w", translator.ConfigPath(), err)
			}
			// Best-effort: remove empty parent dir (e.g. .vscode/, .claude/).
			// os.Remove fails on non-empty dirs, which is the desired behaviour.
			_ = os.Remove(filepath.Dir(configPath))
		} else {
			var out []byte
			var marshalErr error
			if useTOML {
				out, marshalErr = marshalTOML(doc)
			} else {
				out, marshalErr = marshalJSON(doc)
			}
			if marshalErr != nil {
				return nil, fmt.Errorf("failed to marshal %s: %w", translator.ConfigPath(), marshalErr)
			}
			if err := os.WriteFile(configPath, out, 0644); err != nil {
				return nil, fmt.Errorf("failed to write %s: %w", translator.ConfigPath(), err)
			}
		}
	}

	return result, nil
}
