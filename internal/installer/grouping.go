package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/willvelida/code-minions/internal/assistant"
	"github.com/willvelida/code-minions/internal/registry"
)

// GroupingGenerator generates the assistant-specific file(s) that
// express persona grouping. Each coding assistant has a different
// way of saying "these skills and agents belong together."
//
// The generator is responsible for creating files — NOT for
// installing packages (that's the PersonaInstaller's job).
type GroupingGenerator interface {
	// Generate creates the grouping artefact(s) and returns the
	// list of file paths that were generated (relative to target).
	Generate() ([]string, error)

	// SetMCPServers provides the list of MCP server names that
	// were installed for this persona. Generators use this to
	// include MCP tool routing in the persona's agent file.
	// Each generator formats the server names in its own
	// assistant-specific syntax.
	SetMCPServers(servers []string)
}

// NewGroupingGenerator returns the appropriate generator for the
// given assistant. This is a factory function — it picks the right
// implementation based on the assistant name.
//
// Think of it like picking the right adapter for an electrical
// socket: different countries (assistants) need different plugs
// (generators), but they all deliver the same power (persona info).
func NewGroupingGenerator(
	cfg *assistant.Config,
	resolved *registry.ResolvedPersona,
	target string,
	dryRun bool,
	force bool,
) (GroupingGenerator, error) {
	switch cfg.Name {
	case "copilot":
		return &CopilotGrouping{
			Config:   cfg,
			Resolved: resolved,
			Target:   target,
			DryRun:   dryRun,
			Force:    force,
		}, nil
	case "claude":
		return &ClaudeGrouping{
			Config:   cfg,
			Resolved: resolved,
			Target:   target,
			DryRun:   dryRun,
			Force:    force,
		}, nil
	case "opencode":
		return &OpenCodeGrouping{
			Config:   cfg,
			Resolved: resolved,
			Target:   target,
			DryRun:   dryRun,
			Force:    force,
		}, nil
	case "cursor":
		return &CursorGrouping{
			Config:   cfg,
			Resolved: resolved,
			Target:   target,
			DryRun:   dryRun,
			Force:    force,
		}, nil
	case "gemini":
		return &GeminiGrouping{
			Config:   cfg,
			Resolved: resolved,
			Target:   target,
			DryRun:   dryRun,
			Force:    force,
		}, nil
	case "codex":
		return &CodexGrouping{
			Config:   cfg,
			Resolved: resolved,
			Target:   target,
			DryRun:   dryRun,
			Force:    force,
		}, nil
	default:
		// For assistants we don't have a specific generator for,
		// return a no-op generator that doesn't create any files.
		// This means persona installation still works — you just
		// don't get the grouping artefact.
		return &NoopGrouping{}, nil
	}
}

// --- Helper functions shared by all generators ---

// writeGeneratedFile writes content to a file in the target directory.
// It handles creating parent directories, dry-run, and force modes.
// Returns the relative path of the file or an error.
func writeGeneratedFile(target, relativePath string, content []byte, dryRun, force bool) error {
	fullPath := filepath.Join(target, relativePath)

	// Check if the file already exists
	if !force {
		if _, err := os.Stat(fullPath); err == nil {
			return fmt.Errorf("file already exists: %s (use --force to overwrite)", relativePath)
		}
	}

	if dryRun {
		return nil
	}

	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", relativePath, err)
	}

	// Write the file
	if err := os.WriteFile(fullPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", relativePath, err)
	}

	return nil
}

// buildPackageSummary generates a Markdown list of packages in the
// persona. Used by multiple generators.
//
// Example output:
//
//   - **Git Workflow** (`git-workflow`) — Git workflow skill
//   - **Raise Pull Requests** (`raise-pull-requests`) — Pull request skill
func buildPackageSummary(resolved *registry.ResolvedPersona) string {
	var sb strings.Builder
	for _, rp := range resolved.Packages {
		name := rp.Package.Name
		desc := rp.Package.Description
		if desc == "" {
			desc = "No description"
		}
		// Convert "git-workflow" → "Git Workflow" for display
		displayName := toTitleCase(name)
		fmt.Fprintf(&sb, "- **%s** (`%s`) — %s\n", displayName, name, desc)
	}
	return sb.String()
}

// escapeYAMLValue wraps a string in quotes if it contains characters
// that could break YAML parsing (colons, quotes, newlines, etc.).
// Simple values are returned as-is.
func escapeYAMLValue(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return `""`
	}
	// If value contains any YAML-sensitive characters, wrap in double
	// quotes and escape internal double quotes and backslashes.
	if strings.ContainsAny(s, ":\"'{}[]|>&*!#%@`,\n\r\t") {
		escaped := strings.ReplaceAll(s, `\`, `\\`)
		escaped = strings.ReplaceAll(escaped, `"`, `\"`)
		escaped = strings.ReplaceAll(escaped, "\n", `\n`)
		escaped = strings.ReplaceAll(escaped, "\r", `\r`)
		escaped = strings.ReplaceAll(escaped, "\t", `\t`)
		return `"` + escaped + `"`
	}
	return s
}

// toTitleCase converts "git-workflow" → "Git Workflow".
func toTitleCase(s string) string {
	parts := strings.Split(s, "-")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, " ")
}

// --- NoopGrouping ---

// NoopGrouping is a generator that does nothing. Used for assistants
// that don't have a specific grouping mechanism (or that we haven't
// implemented yet).
type NoopGrouping struct{}

func (n *NoopGrouping) Generate() ([]string, error) {
	return nil, nil
}

func (n *NoopGrouping) SetMCPServers(_ []string) {}
