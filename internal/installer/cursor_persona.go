package installer

import (
	"fmt"
	"path"
	"strings"

	"github.com/willvelida/code-minions/internal/assistant"
	"github.com/willvelida/code-minions/internal/registry"
)

// CursorGrouping generates a persona agent file for Cursor.
//
// Cursor discovers agents from .cursor/agents/ (Markdown files with
// YAML frontmatter). It also reads skills from .cursor/skills/.
//
// Cursor's agent discovery model is similar to Claude Code — agents
// are Markdown files with YAML frontmatter describing the agent's
// purpose and capabilities. MCP server references use the mcpServers:
// key in the frontmatter, matching Cursor's MCP config format.
//
// Generated file: .cursor/agents/<persona-name>.agent.md
type CursorGrouping struct {
	Config     *assistant.Config
	Resolved   *registry.ResolvedPersona
	Target     string
	DryRun     bool
	Force      bool
	MCPServers []string // MCP server names (set via SetMCPServers)
}

// SetMCPServers provides MCP server names for frontmatter generation.
// Cursor uses mcpServers: in YAML frontmatter — each entry is a
// server name referencing an already-configured server in
// .cursor/mcp.json.
func (g *CursorGrouping) SetMCPServers(servers []string) {
	g.MCPServers = servers
}

// Generate creates the persona agent file at
// .cursor/agents/<persona-name>.agent.md
func (g *CursorGrouping) Generate() ([]string, error) {
	persona := g.Resolved.Persona

	content := g.buildAgentContent()

	// For Cursor: .cursor/agents/senior-dev.agent.md
	outputPath := path.Join(g.Config.AgentDir, persona.Name+".agent.md")

	err := writeGeneratedFile(g.Target, outputPath, []byte(content), g.DryRun, g.Force)
	if err != nil {
		return nil, err
	}

	return []string{outputPath}, nil
}

// buildAgentContent generates the full Markdown content for the
// persona agent, including YAML frontmatter.
func (g *CursorGrouping) buildAgentContent() string {
	persona := g.Resolved.Persona

	var sb strings.Builder

	// YAML frontmatter — Cursor reads this to discover and
	// describe the agent.
	sb.WriteString("---\n")
	fmt.Fprintf(&sb, "description: %s\n", escapeYAMLValue(persona.Description))

	// MCP server references — tells Cursor which MCP servers
	// this persona can use. Each entry references a server name
	// already configured in .cursor/mcp.json.
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
