package assistant

import (
	"fmt"
	"path"
	"sort"
	"strings"
)

// Config describes a coding assistant's directory layout for agents and skills,
// and its MCP configuration file location.
type Config struct {
	Name             string // Assistant identifier, e.g. "copilot", "claude", "opencode"
	Description      string // Human-readable label, e.g. "GitHub Copilot"
	AgentDir         string // Where agent .md files go, e.g. ".github/agents"
	SkillDir         string // Where skill directories go, e.g. "skills"
	InstructionDir   string // Where file-pattern instructions go, e.g. ".github/instructions"
	PromptDir        string // Where prompt/command files go, e.g. ".github/prompts"
	MCPConfigPath    string // Where MCP server config lives, e.g. ".vscode/mcp.json"
	MCPConfigKey     string // Top-level JSON key for MCP servers, e.g. "servers"
	InstructionsPath string // Where global instructions live, e.g. ".github/copilot-instructions.md"
}

var configs = map[string]Config{
	"copilot": {
		Name:             "copilot",
		Description:      "GitHub Copilot",
		AgentDir:         ".github/agents",
		SkillDir:         "skills",
		InstructionDir:   ".github/instructions",
		PromptDir:        ".github/prompts",
		MCPConfigPath:    ".vscode/mcp.json",
		MCPConfigKey:     "servers",
		InstructionsPath: ".github/copilot-instructions.md",
	},
	"claude": {
		Name:             "claude",
		Description:      "Claude Code",
		AgentDir:         ".claude/agents",
		SkillDir:         ".claude/skills",
		InstructionDir:   ".claude/instructions",
		PromptDir:        ".claude/commands",
		MCPConfigPath:    ".claude/settings.local.json",
		MCPConfigKey:     "mcpServers",
		InstructionsPath: "CLAUDE.md",
	},
	"opencode": {
		Name:             "opencode",
		Description:      "OpenCode",
		AgentDir:         ".opencode/agents",
		SkillDir:         ".opencode/skills",
		InstructionDir:   ".opencode/instructions",
		PromptDir:        ".opencode/commands",
		MCPConfigPath:    "opencode.json",
		MCPConfigKey:     "mcp",
		InstructionsPath: ".opencode/instructions.md",
	},
	"cursor": {
		Name:             "cursor",
		Description:      "Cursor",
		AgentDir:         ".cursor/agents",
		SkillDir:         ".cursor/skills",
		InstructionDir:   ".cursor/rules",
		PromptDir:        ".cursor/rules",
		MCPConfigPath:    ".cursor/mcp.json",
		MCPConfigKey:     "mcpServers",
		InstructionsPath: ".cursor/rules/instructions.mdc",
	},
	"gemini": {
		Name:             "gemini",
		Description:      "Gemini CLI",
		AgentDir:         ".gemini/agents",       // package agent files (reference only — Gemini has no native agents dir)
		SkillDir:         ".gemini/skills",       // skills AND persona grouping target
		InstructionDir:   ".gemini/instructions", // file-pattern instruction files
		PromptDir:        ".gemini/commands",     // custom slash commands (.toml files)
		MCPConfigPath:    ".gemini/settings.json",
		MCPConfigKey:     "mcpServers",
		InstructionsPath: "GEMINI.md",
	},
	"codex": {
		Name:             "codex",
		Description:      "Codex CLI",
		AgentDir:         ".agents/agents",
		SkillDir:         ".agents/skills",
		InstructionDir:   ".agents/instructions",
		PromptDir:        ".agents/prompts",
		MCPConfigPath:    ".codex/config.toml",
		MCPConfigKey:     "mcp_servers",
		InstructionsPath: "AGENTS.md",
	},
}

// Get returns the configuration for the named assistant, or an error if the
// name is not recognised.
func Get(name string) (*Config, error) {
	cfg, ok := configs[name]
	if !ok {
		return nil, fmt.Errorf("unknown assistant %q, valid options: %s", name, strings.Join(List(), ", "))
	}

	// Return a copy so callers can't accidentally modify the registry
	return &cfg, nil
}

// List returns the sorted names of all registered assistants.
func List() []string {
	names := make([]string, 0, len(configs))
	for name := range configs {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// FlagUsage returns a comma-separated list of all registered assistant
// names, suitable for embedding in --for / --from / --to flag help text.
// The list is sorted alphabetically and built from the registry so it
// stays in sync automatically when assistants are added or removed.
func FlagUsage() string {
	return strings.Join(List(), ", ")
}

// NewPathMapper returns a function that remaps file paths from the
// generic package layout (agents/, skills/, standards/) to the
// directories this assistant expects.
//
// For example, with the Copilot config:
//
//	agents/my-agent.md   → .github/agents/my-agent.md
//	skills/foo/SKILL.md  → skills/foo/SKILL.md  (unchanged — Copilot uses skills/ as-is)
//
// With the Claude config:
//
//	agents/my-agent.md   → .claude/agents/my-agent.md
//	skills/foo/SKILL.md  → .claude/skills/foo/SKILL.md
//
// Paths that don't match any known prefix pass through unchanged.
func (c *Config) NewPathMapper() func(string) string {
	// Build the remap table: source prefix → target directory
	// Using a slice of pairs so we check prefixes in a defined order
	type remap struct {
		from string // prefix to match, e.g. "agents"
		to   string // replacement directory, e.g. ".github/agents"
	}

	remaps := []remap{
		{from: "agents", to: c.AgentDir},
		{from: "skills", to: c.SkillDir},
		{from: "instructions", to: c.InstructionDir},
		{from: "prompts", to: c.PromptDir},
	}

	return func(p string) string {
		for _, r := range remaps {
			// Check if the path starts with this prefix (e.g. "agents/..." or exactly "agents")
			if p == r.from || strings.HasPrefix(p, r.from+"/") {
				// If source and target are the same, no work needed
				if r.from == r.to {
					return p
				}
				// Strip the old prefix and prepend the new one
				rest := strings.TrimPrefix(p, r.from)
				return path.Join(r.to, rest)
			}
		}
		// No matching prefix — pass through unchanged
		return p
	}
}

// NewReversePathMapper returns a function that remaps assistant-specific
// paths back to the generic package layout.
//
// For example, with the Copilot config:
//
//	.github/agents/my-agent.md → agents/my-agent.md
//	skills/foo/SKILL.md        → skills/foo/SKILL.md  (unchanged — Copilot uses skills/ as-is)
//
// With the Claude config:
//
//	.claude/agents/my-agent.md  → agents/my-agent.md
//	.claude/skills/foo/SKILL.md → skills/foo/SKILL.md
//
// Paths that don't match any known prefix pass through unchanged.
func (c *Config) NewReversePathMapper() func(string) string {
	type remap struct {
		from string // assistant-specific prefix to match, e.g. ".github/agents"
		to   string // generic prefix to replace with, e.g. "agents"
	}

	remaps := []remap{
		{from: c.AgentDir, to: "agents"},
		{from: c.SkillDir, to: "skills"},
		{from: c.InstructionDir, to: "instructions"},
		{from: c.PromptDir, to: "prompts"},
	}

	return func(p string) string {
		// Collect all entries whose prefix matches the path. When an
		// assistant uses the same directory for multiple generic prefixes
		// (e.g. Cursor maps both instructions/ and prompts/ to
		// .cursor/rules/), we need to pick the best match using the
		// file extension as a tie-breaker.
		var candidates []remap
		for _, r := range remaps {
			if p == r.from || strings.HasPrefix(p, r.from+"/") {
				if r.from == r.to {
					return p
				}
				candidates = append(candidates, r)
			}
		}

		if len(candidates) == 1 {
			rest := strings.TrimPrefix(p, candidates[0].from)
			return path.Join(candidates[0].to, rest)
		}

		if len(candidates) > 1 {
			// Disambiguate by file extension when multiple generic prefixes
			// map to the same assistant directory (e.g. Cursor maps both
			// instructions/ and prompts/ to .cursor/rules/).
			//
			// For Cursor, both instructions and prompts become .mdc files,
			// and instruction files lose their ".instructions" infix
			// (security.instructions.md → security.mdc). This means a bare
			// .mdc file is ambiguous. We can only disambiguate when the
			// canonical extension is still intact (.prompt.md, .instructions.md)
			// or when a distinctive extension is present (.toml for Gemini).
			//
			// When no hint matches, the first candidate wins (instructions
			// before prompts, matching the remaps ordering).
			extHints := map[string][]string{
				"instructions": {".instructions.md", ".instructions.mdc"},
				"prompts":      {".prompt.md", ".toml"},
			}
			// Check instructions first (more specific patterns)
			for _, priority := range []string{"instructions", "prompts"} {
				for _, r := range candidates {
					if r.to != priority {
						continue
					}
					if exts, ok := extHints[r.to]; ok {
						for _, ext := range exts {
							if strings.HasSuffix(p, ext) {
								rest := strings.TrimPrefix(p, r.from)
								return path.Join(r.to, rest)
							}
						}
					}
				}
			}
			// No extension match — fall through to first candidate
			rest := strings.TrimPrefix(p, candidates[0].from)
			return path.Join(candidates[0].to, rest)
		}

		return p
	}
}
