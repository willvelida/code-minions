package installer

import (
	"fmt"
	"strings"

	"github.com/willvelida/code-minions/internal/assistant"
	"github.com/willvelida/code-minions/internal/registry"
)

// ClaudeGrouping generates a persona subagent file for Claude Code.
//
// Claude Code discovers agents automatically from .claude/agents/.
// A persona subagent is a Markdown file with YAML frontmatter that
// describes the persona and references its constituent packages.
//
// Claude Code's subagent model supports "context: fork" which means
// the persona agent can delegate to package-level agents. We don't
// set that here — instead, the persona agent describes its
// capabilities and lets Claude Code's routing figure out the rest.
//
// Generated file: .claude/agents/<persona-name>.agent.md
type ClaudeGrouping struct {
	Config     *assistant.Config
	Resolved   *registry.ResolvedPersona
	Target     string
	DryRun     bool
	Force      bool
	MCPServers []string // MCP server names (set via SetMCPServers)
}

// SetMCPServers provides MCP server names for frontmatter generation.
// Claude uses mcpServers: in YAML frontmatter — each entry is a
// server name referencing an already-configured server.
func (g *ClaudeGrouping) SetMCPServers(servers []string) {
	g.MCPServers = servers
}

// Generate creates the persona subagent file at
// .claude/agents/<persona-name>.agent.md
func (g *ClaudeGrouping) Generate() ([]string, error) {
	persona := g.Resolved.Persona

	// Build the agent file content with YAML frontmatter.
	// Frontmatter tells Claude Code about the agent; the body
	// provides detailed instructions.
	content := g.buildAgentContent()

	// The file goes in the assistant's agent directory.
	// For Claude: .claude/agents/senior-dev.agent.md
	outputPath := g.Config.AgentDir + "/" + persona.Name + ".agent.md"

	err := writeGeneratedFile(g.Target, outputPath, []byte(content), g.DryRun, g.Force)
	if err != nil {
		return nil, err
	}

	return []string{outputPath}, nil
}

// buildAgentContent generates the full Markdown content for the
// persona subagent, including YAML frontmatter.
func (g *ClaudeGrouping) buildAgentContent() string {
	persona := g.Resolved.Persona

	var sb strings.Builder

	// YAML frontmatter — Claude Code reads this to understand
	// the agent's purpose and capabilities.
	sb.WriteString("---\n")
	fmt.Fprintf(&sb, "description: %s\n", strings.TrimSpace(persona.Description))

	// MCP server references — tells Claude which MCP servers
	// this persona can use. Each entry references a server name
	// that's already configured in .claude/settings.local.json.
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
