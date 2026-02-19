package installer

import (
	"fmt"
	"path"
	"strings"

	"github.com/willvelida/code-minions/internal/assistant"
	"github.com/willvelida/code-minions/internal/registry"
)

// BuildClaudeMDForPersona generates CLAUDE.md content for a persona
// install. The output uses Claude Code conventions:
//   - @path/to/file imports for skill references
//   - Freeform Markdown with bullet lists
//   - Reference to the persona subagent file
func BuildClaudeMDForPersona(resolved *registry.ResolvedPersona, cfg *assistant.Config) string {
	var sb strings.Builder

	sb.WriteString("# Project Instructions\n\n")
	sb.WriteString("This project uses [code-minions](https://github.com/willvelida/code-minions) for AI-assisted development.\n\n")

	persona := resolved.Persona

	// Persona section
	fmt.Fprintf(&sb, "## Installed Persona: %s\n\n", toTitleCase(persona.Name))

	if persona.Description != "" {
		fmt.Fprintf(&sb, "%s\n\n", persona.Description)
	}

	sb.WriteString("### Skills\n\n")

	for _, rp := range resolved.Packages {
		name := rp.Package.Name
		desc := rp.Package.Description
		if desc == "" {
			desc = "No description"
		}
		displayName := toTitleCase(name)
		// Use @import syntax to reference the skill file.
		// Claude Code resolves these as file references and injects
		// the content into context.
		skillPath := path.Join(cfg.SkillDir, name, "SKILL.md")
		fmt.Fprintf(&sb, "- **%s** — %s (see @%s)\n", displayName, desc, skillPath)
	}

	// Reference the persona subagent
	sb.WriteString("\n### Agent\n\n")
	agentPath := path.Join(cfg.AgentDir, persona.Name+".agent.md")
	fmt.Fprintf(&sb, "See @%s for persona agent configuration.\n", agentPath)

	return sb.String()
}

// BuildClaudeMDForPackages generates CLAUDE.md content for individual
// package installs (--package mode). Each package gets an @import
// reference to its skill file.
func BuildClaudeMDForPackages(packages []string, cfg *assistant.Config) string {
	var sb strings.Builder

	sb.WriteString("# Project Instructions\n\n")
	sb.WriteString("This project uses [code-minions](https://github.com/willvelida/code-minions) for AI-assisted development.\n\n")
	sb.WriteString("## Installed Packages\n\n")

	for _, name := range packages {
		displayName := toTitleCase(name)
		skillPath := path.Join(cfg.SkillDir, name, "SKILL.md")
		fmt.Fprintf(&sb, "- **%s** — @%s\n", displayName, skillPath)
	}

	return sb.String()
}
