package mcp

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// TranslateResult captures the outcome of a pairwise MCP config translation.
type TranslateResult struct {
	SourceAssistant string       `json:"sourceAssistant"`
	TargetAssistant string       `json:"targetAssistant"`
	ConfigPath      string       `json:"configPath"` // Target config file written
	Merge           *MergeResult `json:"merge"`      // Detailed merge outcome
	Warnings        []string     `json:"warnings"`   // All warnings combined
	DryRun          bool         `json:"dryRun,omitempty"`
}

// TranslateOptions configures a pairwise translation.
type TranslateOptions struct {
	From      string // Source assistant name
	To        string // Target assistant name
	TargetDir string // Working directory (default ".")
	Server    string // Translate only this server (empty = all)
	Force     bool   // Overwrite conflicting servers
	DryRun    bool   // Preview without writing
}

// Translate reads MCP config from the source assistant, translates to the
// target assistant's format, and merges into the target's config file.
//
// Flow:
//  1. Validate: source ≠ target, both are recognised assistants
//  2. Read source config file
//  3. Parse via Reader into canonical Config
//  4. Filter by server name if specified
//  5. Translate canonical → target format via Translator
//  6. Read existing target config file
//  7. Merge translated servers into target
//  8. Write (unless dry-run)
func Translate(opts TranslateOptions) (*TranslateResult, error) {
	// Validate
	if opts.From == opts.To {
		return nil, fmt.Errorf("source and target assistant cannot be the same (%q)", opts.From)
	}

	reader, err := NewReader(opts.From)
	if err != nil {
		return nil, fmt.Errorf("unknown source assistant %q", opts.From)
	}

	translator, err := NewTranslator(opts.To)
	if err != nil {
		return nil, fmt.Errorf("unknown target assistant %q", opts.To)
	}

	targetDir := opts.TargetDir
	if targetDir == "" {
		targetDir = "."
	}

	// Step 2: Read source config file
	sourceConfigPath := filepath.Join(targetDir, reader.ConfigPath())
	sourceData, err := os.ReadFile(sourceConfigPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("no MCP config found for %s at %s", opts.From, reader.ConfigPath())
		}
		return nil, fmt.Errorf("failed to read %s: %w", reader.ConfigPath(), err)
	}

	// Step 3: Parse into canonical Config
	cfg, readWarnings, err := reader.Read(sourceData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s MCP config: %w", opts.From, err)
	}

	var allWarnings []string
	allWarnings = append(allWarnings, readWarnings...)

	if len(cfg.Servers) == 0 {
		allWarnings = append(allWarnings, fmt.Sprintf("no MCP servers found in %s config", opts.From))
		return &TranslateResult{
			SourceAssistant: opts.From,
			TargetAssistant: opts.To,
			ConfigPath:      translator.ConfigPath(),
			Merge:           &MergeResult{},
			Warnings:        allWarnings,
			DryRun:          opts.DryRun,
		}, nil
	}

	// Step 4: Filter by server name if specified
	if opts.Server != "" {
		s, ok := cfg.Servers[opts.Server]
		if !ok {
			return nil, fmt.Errorf("server %q not found in %s config", opts.Server, opts.From)
		}
		cfg = &Config{
			Servers: map[string]Server{
				opts.Server: s,
			},
		}
	}

	// Step 5: Translate canonical → target format
	servers, translateWarnings, err := translator.Translate(cfg)
	if err != nil {
		return nil, fmt.Errorf("translation failed: %w", err)
	}
	allWarnings = append(allWarnings, translateWarnings...)

	if len(servers) == 0 {
		sort.Strings(allWarnings)
		return &TranslateResult{
			SourceAssistant: opts.From,
			TargetAssistant: opts.To,
			ConfigPath:      translator.ConfigPath(),
			Merge:           &MergeResult{Warnings: allWarnings},
			Warnings:        allWarnings,
			DryRun:          opts.DryRun,
		}, nil
	}

	// Step 6: Read existing target config file
	targetConfigPath := filepath.Join(targetDir, translator.ConfigPath())
	var existing []byte
	if data, err := os.ReadFile(targetConfigPath); err == nil {
		existing = data
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("failed to read %s: %w", translator.ConfigPath(), err)
	}

	// Step 7: Merge
	merged, mergeResult, err := Merge(existing, servers, translator.ConfigKey(), opts.Force)
	if err != nil {
		return nil, fmt.Errorf("merge failed for %s: %w", translator.ConfigPath(), err)
	}

	allWarnings = append(allWarnings, mergeResult.Warnings...)

	// Check for potential secrets in env vars
	hasEnvVars := false
	for _, s := range cfg.Servers {
		if len(s.Env) > 0 {
			hasEnvVars = true
			break
		}
	}
	if hasEnvVars {
		allWarnings = append(allWarnings, "MCP config may contain secrets. Review before committing.")
	}

	sort.Strings(allWarnings)

	// Overwrite mergeResult.Warnings with the unified set so callers
	// see all warnings (reader + translator + merge + secrets) in one place.
	mergeResult.Warnings = allWarnings

	// Step 8: Write (unless dry-run)
	if !opts.DryRun && len(mergeResult.Added) > 0 {
		if err := os.MkdirAll(filepath.Dir(targetConfigPath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory for %s: %w", translator.ConfigPath(), err)
		}
		if err := os.WriteFile(targetConfigPath, merged, 0644); err != nil {
			return nil, fmt.Errorf("failed to write %s: %w", translator.ConfigPath(), err)
		}
	}

	return &TranslateResult{
		SourceAssistant: opts.From,
		TargetAssistant: opts.To,
		ConfigPath:      translator.ConfigPath(),
		Merge:           mergeResult,
		Warnings:        allWarnings,
		DryRun:          opts.DryRun,
	}, nil
}

// ListServers reads the MCP config for the named assistant from the given
// directory and returns the servers in canonical format.
func ListServers(assistant, targetDir string) (*Config, []string, error) {
	reader, err := NewReader(assistant)
	if err != nil {
		return nil, nil, err
	}

	if targetDir == "" {
		targetDir = "."
	}

	configPath := filepath.Join(targetDir, reader.ConfigPath())
	data, err := os.ReadFile(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Config{Servers: map[string]Server{}}, nil, nil
		}
		return nil, nil, fmt.Errorf("failed to read %s: %w", reader.ConfigPath(), err)
	}

	return reader.Read(data)
}
