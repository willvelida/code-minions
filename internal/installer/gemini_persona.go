package installer

import (
	"fmt"
	"path"
	"strings"

	"github.com/willvelida/code-minions/internal/assistant"
	"github.com/willvelida/code-minions/internal/registry"
)

// GeminiGrouping generates a persona agent file for Gemini CLI.
//
// Gemini CLI discovers context from GEMINI.md and the .gemini/
// directory. Agents are placed in .gemini/agents/ as Markdown files
// with YAML frontmatter. Skills go in .gemini/skills/.
//
// Gemini's agent discovery model is similar to Claude Code and
// Cursor — agents are Markdown files with YAML frontmatter
// describing the agent's purpose and capabilities. MCP server
// references use the mcpServers: key in the frontmatter, matching
// Gemini's MCP config format.
//
// Generated file: .gemini/agents/<persona-name>.agent.md
type GeminiGrouping struct {
	Config     *assistant.Config
	Resolved   *registry.ResolvedPersona
	Target     string
	DryRun     bool
	Force      bool
	MCPServers []string // MCP server names (set via SetMCPServers)
}

// SetMCPServers provides MCP server names for frontmatter generation.
// Gemini uses mcpServers: in YAML frontmatter — each entry is a
// server name referencing an already-configured server in
// .gemini/settings.json.
func (g *GeminiGrouping) SetMCPServers(servers []string) {
	g.MCPServers = servers
}

// Generate creates the persona agent file at
// .gemini/agents/<persona-name>.agent.md
func (g *GeminiGrouping) Generate() ([]string, error) {
	persona := g.Resolved.Persona

	content := g.buildAgentContent()

	// For Gemini: .gemini/agents/senior-dev.agent.md
	outputPath := path.Join(g.Config.AgentDir, persona.Name+".agent.md")

	err := writeGeneratedFile(g.Target, outputPath, []byte(content), g.DryRun, g.Force)
	if err != nil {
		return nil, err
	}

	return []string{outputPath}, nil
}

// buildAgentContent generates the full Markdown content for the
// persona agent, including YAML frontmatter.
func (g *GeminiGrouping) buildAgentContent() string {
	persona := g.Resolved.Persona

	var sb strings.Builder

	// YAML frontmatter — Gemini reads this to discover and
	// describe the agent.
	sb.WriteString("---\n")
	fmt.Fprintf(&sb, "description: %s\n", escapeYAMLValue(persona.Description))

	// MCP server references — tells Gemini which MCP servers
	// this persona can use. Each entry references a server name
	// already configured in .gemini/settings.json.
	if len(g.MCPServers) > 0 {
		sb.WriteString("mcpServers:\n")
		for _, server := range g.MCPServers {
			fmt.Fprintf(&sb, "  - %s\n", server)
		}
	}

	sb.WriteString("---\n\n")

	// Agent title
	fmt.Fprintf(&sb, "# %s\n\n", persona.Name)

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
