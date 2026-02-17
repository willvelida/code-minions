package mcp

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// InstallResult captures what happened during MCP installation for a
// single package.
type InstallResult struct {
	ConfigPath  string       // Relative path written, e.g. ".vscode/mcp.json"
	Merge       *MergeResult // Detailed merge outcome
	ServerNames []string     // Successfully installed server names (for manifest tracking)
	DryRun      bool
}

// Install processes a package's mcp.yaml and merges its MCP servers into the
// target assistant's config file.
//
// Returns (nil, nil) when the package has no mcp.yaml — callers can skip
// MCP output entirely in that case.
//
// Flow:
//  1. LoadConfig() from embedded FS
//  2. translator.Translate() to get assistant-specific server map
//  3. Read existing config file (or empty for new)
//  4. Merge() to produce updated JSON
//  5. Write (unless dry-run)
func Install(content fs.FS, packageDir, targetDir string, translator Translator, force, dryRun bool) (*InstallResult, error) {
	// Step 1: parse canonical mcp.yaml
	cfg, err := LoadConfig(content, packageDir)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		// No mcp.yaml in this package — nothing to do
		return nil, nil
	}

	// Step 2: translate to assistant-specific format
	servers, translateWarnings, err := translator.Translate(cfg)
	if err != nil {
		return nil, fmt.Errorf("MCP translation failed: %w", err)
	}

	if len(servers) == 0 {
		// All servers were incompatible with this assistant
		result := &InstallResult{
			ConfigPath: translator.ConfigPath(),
			Merge:      &MergeResult{Warnings: translateWarnings},
			DryRun:     dryRun,
		}
		return result, nil
	}

	// Step 3: read existing config file
	configPath := filepath.Join(targetDir, translator.ConfigPath())
	var existing []byte
	if data, err := os.ReadFile(configPath); err == nil {
		existing = data
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("failed to read %s: %w", translator.ConfigPath(), err)
	}

	// Step 4: merge
	merged, mergeResult, err := Merge(existing, servers, translator.ConfigKey(), force)
	if err != nil {
		return nil, fmt.Errorf("MCP merge failed for %s: %w", translator.ConfigPath(), err)
	}

	// Append translation warnings to merge warnings and sort for deterministic output
	mergeResult.Warnings = append(translateWarnings, mergeResult.Warnings...)
	sort.Strings(mergeResult.Warnings)

	// Step 5: write (unless dry-run)
	if !dryRun && (len(mergeResult.Added) > 0) {
		if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory for %s: %w", translator.ConfigPath(), err)
		}
		if err := os.WriteFile(configPath, merged, 0644); err != nil {
			return nil, fmt.Errorf("failed to write %s: %w", translator.ConfigPath(), err)
		}
	}

	return &InstallResult{
		ConfigPath:  translator.ConfigPath(),
		Merge:       mergeResult,
		ServerNames: mergeResult.Added,
		DryRun:      dryRun,
	}, nil
}

// InstallServers merges pre-translated MCP server definitions into the
// assistant's config file. This is the "persona-aware" variant of Install:
// instead of loading a single package's mcp.yaml, it accepts a combined
// server map that has already been collected and translated from multiple
// packages.
//
// The Merge function handles deduplication automatically:
//   - Same name + same config as existing → skip (deduplicate)
//   - Same name + different config → conflict (unless force)
//   - New server → add
//
// This function is called by PersonaInstaller.installMCP().
func InstallServers(targetDir string, translator Translator, servers map[string]any, force, dryRun bool) (*InstallResult, error) {
	if len(servers) == 0 {
		return nil, nil
	}

	// Read existing config file (may not exist yet).
	configPath := filepath.Join(targetDir, translator.ConfigPath())
	var existing []byte
	if data, err := os.ReadFile(configPath); err == nil {
		existing = data
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("failed to read %s: %w", translator.ConfigPath(), err)
	}

	// Merge the servers into the config.
	merged, mergeResult, err := Merge(existing, servers, translator.ConfigKey(), force)
	if err != nil {
		return nil, fmt.Errorf("MCP merge failed for %s: %w", translator.ConfigPath(), err)
	}

	// Write (unless dry-run or nothing was added).
	if !dryRun && len(mergeResult.Added) > 0 {
		if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory for %s: %w", translator.ConfigPath(), err)
		}
		if err := os.WriteFile(configPath, merged, 0644); err != nil {
			return nil, fmt.Errorf("failed to write %s: %w", translator.ConfigPath(), err)
		}
	}

	return &InstallResult{
		ConfigPath:  translator.ConfigPath(),
		Merge:       mergeResult,
		ServerNames: mergeResult.Added,
		DryRun:      dryRun,
	}, nil
}
