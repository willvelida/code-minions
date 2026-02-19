package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
)

// MergeResult reports what happened during a merge operation.
type MergeResult struct {
	Added    []string // Server names that were added
	Skipped  []string // Server names already present (identical config)
	Conflict []string // Server names already present (different config)
	Warnings []string // e.g. empty env vars, unsupported transport
}

// Merge reads existing JSON config bytes (may be nil/empty for a new file),
// merges the translated MCP servers under the given configKey, and returns
// the resulting JSON bytes plus a report of what changed.
//
// Merge semantics:
//   - New server → add
//   - Existing server, identical config → skip
//   - Existing server, different config → skip (unless force=true)
//   - Non-MCP content in the file → preserved unchanged
//
// The returned JSON uses 2-space indent (standard for VS Code config files).
func Merge(existing []byte, translated map[string]any, configKey string, force bool) ([]byte, *MergeResult, error) {
	doc, err := unmarshalJSON(existing)
	if err != nil {
		return nil, nil, err
	}

	result := mergeInto(doc, translated, configKey, force)

	out, err := marshalJSON(doc)
	if err != nil {
		return nil, nil, err
	}

	detectEmptyEnvWarnings(result, translated)

	return out, result, nil
}

// MergeTOML reads existing TOML config bytes, merges translated MCP servers
// under the given configKey, and returns the resulting TOML bytes plus a
// report of what changed.
//
// The merge semantics are identical to Merge (JSON variant).
func MergeTOML(existing []byte, translated map[string]any, configKey string, force bool) ([]byte, *MergeResult, error) {
	doc, err := unmarshalTOML(existing)
	if err != nil {
		return nil, nil, err
	}

	result := mergeInto(doc, translated, configKey, force)

	out, err := marshalTOML(doc)
	if err != nil {
		return nil, nil, err
	}

	detectEmptyEnvWarnings(result, translated)

	return out, result, nil
}

// MergeForFile dispatches to Merge or MergeTOML based on the config file path.
// Files ending in .toml use TOML encoding; all others use JSON.
func MergeForFile(configPath string, existing []byte, translated map[string]any, configKey string, force bool) ([]byte, *MergeResult, error) {
	if isTOMLFile(configPath) {
		return MergeTOML(existing, translated, configKey, force)
	}
	return Merge(existing, translated, configKey, force)
}

// isTOMLFile returns true if the file path has a .toml extension.
func isTOMLFile(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".toml")
}

// unmarshalJSON parses JSON bytes into a generic map.
func unmarshalJSON(data []byte) (map[string]any, error) {
	doc := make(map[string]any)
	if len(data) > 0 {
		if err := json.Unmarshal(data, &doc); err != nil {
			return nil, fmt.Errorf("failed to parse existing config: %w", err)
		}
	}
	return doc, nil
}

// unmarshalTOML parses TOML bytes into a generic map.
func unmarshalTOML(data []byte) (map[string]any, error) {
	doc := make(map[string]any)
	if len(data) > 0 {
		if err := toml.Unmarshal(data, &doc); err != nil {
			return nil, fmt.Errorf("failed to parse existing config: %w", err)
		}
	}
	return doc, nil
}

// marshalJSON serialises a map to JSON with 2-space indent and trailing newline.
func marshalJSON(doc map[string]any) ([]byte, error) {
	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}
	out = append(out, '\n')
	return out, nil
}

// marshalTOML serialises a map to TOML with trailing newline.
func marshalTOML(doc map[string]any) ([]byte, error) {
	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	if err := enc.Encode(doc); err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}
	return buf.Bytes(), nil
}

// mergeInto merges translated servers into the document under configKey.
// This is the shared merge logic used by both JSON and TOML paths.
func mergeInto(doc map[string]any, translated map[string]any, configKey string, force bool) *MergeResult {
	result := &MergeResult{}

	// Get or create the MCP servers section
	var existingServers map[string]any
	if raw, ok := doc[configKey]; ok {
		if m, ok := raw.(map[string]any); ok {
			existingServers = m
		} else {
			existingServers = make(map[string]any)
		}
	} else {
		existingServers = make(map[string]any)
	}

	// Merge each translated server into the existing map
	// Sort names for deterministic output
	names := make([]string, 0, len(translated))
	for name := range translated {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		newServer := translated[name]

		if existingServer, exists := existingServers[name]; exists {
			// Compare: normalise both through JSON round-trip to avoid
			// type mismatches (e.g. float64 vs int from JSON unmarshal)
			if ServersEqual(existingServer, newServer) {
				result.Skipped = append(result.Skipped, name)
			} else if force {
				existingServers[name] = newServer
				result.Added = append(result.Added, name)
			} else {
				result.Conflict = append(result.Conflict, name)
			}
		} else {
			existingServers[name] = newServer
			result.Added = append(result.Added, name)
		}
	}

	doc[configKey] = existingServers

	return result
}

// detectEmptyEnvWarnings checks newly-added servers for empty env vars.
func detectEmptyEnvWarnings(result *MergeResult, translated map[string]any) {
	addedSet := make(map[string]bool, len(result.Added))
	for _, n := range result.Added {
		addedSet[n] = true
	}

	names := make([]string, 0, len(translated))
	for name := range translated {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		if !addedSet[name] {
			continue
		}
		serverMap, ok := translated[name].(map[string]any)
		if !ok {
			continue
		}
		// Check both "env" (Copilot/Claude) and "environment" (OpenCode)
		for _, envKey := range []string{"env", "environment"} {
			if envRaw, ok := serverMap[envKey]; ok {
				if envMap, ok := envRaw.(map[string]string); ok {
					for k, v := range envMap {
						if v == "" {
							result.Warnings = append(result.Warnings,
								fmt.Sprintf("server %q has empty env var: %s", name, k))
						}
					}
				} else if envMap, ok := envRaw.(map[string]any); ok {
					for k, v := range envMap {
						if str, ok := v.(string); ok && str == "" {
							result.Warnings = append(result.Warnings,
								fmt.Sprintf("server %q has empty env var: %s", name, k))
						}
					}
				}
			}
		}
	}
	sort.Strings(result.Warnings)
}

// ServersEqual compares two server definitions for equality by normalising
// them through JSON serialisation. This handles type differences that occur
// when JSON is unmarshalled into map[string]any (e.g. numbers become float64).
func ServersEqual(a, b any) bool {
	ajson, err1 := json.Marshal(a)
	bjson, err2 := json.Marshal(b)
	if err1 != nil || err2 != nil {
		return reflect.DeepEqual(a, b)
	}

	var anorm, bnorm any
	_ = json.Unmarshal(ajson, &anorm)
	_ = json.Unmarshal(bjson, &bnorm)
	return reflect.DeepEqual(anorm, bnorm)
}

// MergeCanonical combines multiple canonical MCP configs into a single
// config. Configs are applied in order: later configs override earlier
// ones when the same server name appears.
//
// This is used for three-layer merging during team install:
//
//	layer 1: package-level servers (one Config per package)
//	layer 2: team-level servers (team's Config)
//
// The caller typically passes package configs first, then the team config
// last so that team-level definitions take priority.
//
// Returns nil if all inputs are nil or empty.
func MergeCanonical(configs ...*Config) *Config {
	merged := make(map[string]Server)
	for _, cfg := range configs {
		if cfg == nil {
			continue
		}
		for name, srv := range cfg.Servers {
			merged[name] = srv
		}
	}
	if len(merged) == 0 {
		return nil
	}
	return &Config{Servers: merged}
}
