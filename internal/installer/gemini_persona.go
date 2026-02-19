package installer

import (
	"fmt"
	"path"
	"strings"

	"github.com/willvelida/code-minions/internal/assistant"
	"github.com/willvelida/code-minions/internal/registry"
)

// GeminiGrouping generates a persona skill file for Gemini CLI.
//
// Gemini CLI discovers skills from .gemini/skills/<name>/SKILL.md.
// Skills provide passive context — instructions that are loaded
// automatically when the skill is enabled.
//
// Unlike Claude/Copilot/Cursor, Gemini doesn't have an agents/
// directory. Personas map to skills because a persona is essentially
// a set of instructions describing a role and its capabilities —
// exactly what a Gemini skill is.
//
// MCP server references are NOT included in the skill file. Gemini
// configures MCP servers in .gemini/settings.json, and skills don't
// reference them directly.
//
// Generated file: .gemini/skills/<persona-name>/SKILL.md
type GeminiGrouping struct {
	Config     *assistant.Config
	Resolved   *registry.ResolvedPersona
	Target     string
	DryRun     bool
	Force      bool
	MCPServers []string // Unused for Gemini (no frontmatter references)
}

// SetMCPServers is a no-op for Gemini. MCP servers are configured in
// .gemini/settings.json — skill files don't reference them.
func (g *GeminiGrouping) SetMCPServers(servers []string) {
	g.MCPServers = servers
}

// Generate creates the persona skill file at
// .gemini/skills/<persona-name>/SKILL.md
func (g *GeminiGrouping) Generate() ([]string, error) {
	persona := g.Resolved.Persona

	content := g.buildSkillContent()

	// For Gemini: .gemini/skills/senior-dev/SKILL.md
	outputPath := path.Join(g.Config.SkillDir, persona.Name, "SKILL.md")

	err := writeGeneratedFile(g.Target, outputPath, []byte(content), g.DryRun, g.Force)
	if err != nil {
		return nil, err
	}

	return []string{outputPath}, nil
}

// buildSkillContent generates plain Markdown content for the persona
// skill. Gemini skills are plain Markdown — no YAML frontmatter.
func (g *GeminiGrouping) buildSkillContent() string {
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
