package assistant

import (
	"fmt"
	"path"
	"sort"
	"strings"
)

// Config describes a coding assistant's directory layout for agents and skills.
type Config struct {
	Name        string // Assistant identifier, e.g. "copilot", "claude", "opencode"
	Description string // Human-readable label, e.g. "GitHub Copilot"
	AgentDir    string // Where agent .md files go, e.g. ".github/agents"
	SkillDir    string // Where skill directories go, e.g. "skills"
}

var configs = map[string]Config{
	"copilot": {
		Name:        "copilot",
		Description: "GitHub Copilot",
		AgentDir:    ".github/agents",
		SkillDir:    "skills",
	},
	"claude": {
		Name:        "claude",
		Description: "Claude Code",
		AgentDir:    ".claude/agents",
		SkillDir:    ".claude/skills",
	},
	"opencode": {
		Name:        "opencode",
		Description: "OpenCode",
		AgentDir:    ".opencode/agents",
		SkillDir:    ".opencode/skills",
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
