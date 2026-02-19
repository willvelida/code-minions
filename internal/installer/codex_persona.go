package installer

import (
	"fmt"
	"path"
	"strings"

	"github.com/willvelida/code-minions/internal/assistant"
	"github.com/willvelida/code-minions/internal/registry"
)

// CodexGrouping generates a persona skill file for Codex CLI.
//
// Codex CLI discovers skills from .agents/skills/<name>/SKILL.md.
// Like Gemini, skills provide passive context — instructions that
// are loaded automatically when the skill is enabled.
//
// Codex uses AGENTS.md for project-level instructions and
// .agents/skills/ for skill discovery. Personas map to skills
// because a persona is a set of instructions describing a role
// and its capabilities.
//
// MCP server references are NOT included in the skill file. Codex
// configures MCP servers in .codex/config.toml, and skills don't
// reference them directly.
//
// Generated file: .agents/skills/<persona-name>/SKILL.md
type CodexGrouping struct {
	Config     *assistant.Config
	Resolved   *registry.ResolvedPersona
	Target     string
	DryRun     bool
	Force      bool
	MCPServers []string // Unused for Codex (no frontmatter references)
}

// SetMCPServers is a no-op for Codex. MCP servers are configured in
// .codex/config.toml — skill files don't reference them.
func (g *CodexGrouping) SetMCPServers(servers []string) {
	g.MCPServers = servers
}

// Generate creates the persona skill file at
// .agents/skills/<persona-name>/SKILL.md
func (g *CodexGrouping) Generate() ([]string, error) {
	persona := g.Resolved.Persona

	content := g.buildSkillContent()

	// For Codex: .agents/skills/senior-dev/SKILL.md
	outputPath := path.Join(g.Config.SkillDir, persona.Name, "SKILL.md")

	err := writeGeneratedFile(g.Target, outputPath, []byte(content), g.DryRun, g.Force)
	if err != nil {
		return nil, err
	}

	return []string{outputPath}, nil
}

// buildSkillContent generates plain Markdown content for the persona
// skill. Codex skills are plain Markdown — no YAML frontmatter.
func (g *CodexGrouping) buildSkillContent() string {
	persona := g.Resolved.Persona

	var sb strings.Builder

	// Skill title
	fmt.Fprintf(&sb, "# %s\n\n", persona.Name)

	// Description
	if persona.Description != "" {
		fmt.Fprintf(&sb, "%s\n\n", strings.TrimSpace(persona.Description))
	}

	// Body text — describes the persona's capabilities
	fmt.Fprintf(&sb, "You are operating as the **%s** persona. You have the following capabilities:\n\n", persona.Name)

	// List each package as a capability
	for _, rp := range g.Resolved.Packages {
		name := rp.Package.Name
		desc := rp.Package.Description
		if desc == "" {
			desc = "No description"
		}
		displayName := toTitleCase(name)
		fmt.Fprintf(&sb, "- **%s** — %s\n", displayName, desc)
	}

	// Persona-level instructions
	if persona.Instructions != "" {
		sb.WriteString("\n")
		sb.WriteString(strings.TrimSpace(persona.Instructions))
		sb.WriteString("\n")
	}

	return sb.String()
}
