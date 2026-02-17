package installer

import (
	"fmt"
	"strings"

	"github.com/willvelida/code-minions/internal/assistant"
	"github.com/willvelida/code-minions/internal/registry"
)

// OpenCodeGrouping generates a persona agent file for OpenCode.
//
// OpenCode discovers agents from .opencode/agents/ (Markdown files
// with YAML frontmatter). It also auto-discovers skills from
// .opencode/skills/ via a built-in `skill` tool.
//
// The persona agent just needs to mention skills by name — OpenCode's
// name-based discovery handles the rest. The generated file is very
// similar to the Claude generator, but uses OpenCode-specific
// frontmatter fields (mode, tools).
//
// Generated file: .opencode/agents/<persona-name>.md
type OpenCodeGrouping struct {
	Config     *assistant.Config
	Resolved   *registry.ResolvedPersona
	Target     string
	DryRun     bool
	Force      bool
	MCPServers []string // MCP server names (set via SetMCPServers)
}

// SetMCPServers provides MCP server names for frontmatter generation.
// OpenCode uses tools: in frontmatter with wildcard patterns like
// server-name_*: true to grant access to all tools from an MCP server.
func (g *OpenCodeGrouping) SetMCPServers(servers []string) {
	g.MCPServers = servers
}

// Generate creates the persona agent file at
// .opencode/agents/<persona-name>.md
func (g *OpenCodeGrouping) Generate() ([]string, error) {
	persona := g.Resolved.Persona

	content := g.buildAgentContent()

	// For OpenCode: .opencode/agents/senior-dev.md
	// Note: OpenCode uses .md (not .agent.md) for agent files.
	outputPath := g.Config.AgentDir + "/" + persona.Name + ".md"

	err := writeGeneratedFile(g.Target, outputPath, []byte(content), g.DryRun, g.Force)
	if err != nil {
		return nil, err
	}

	return []string{outputPath}, nil
}

// buildAgentContent generates the Markdown content for the persona
// agent with OpenCode-specific frontmatter.
func (g *OpenCodeGrouping) buildAgentContent() string {
	persona := g.Resolved.Persona

	var sb strings.Builder

	// YAML frontmatter — OpenCode-specific fields.
	// "description" is required for agent discovery.
	// "mode: primary" makes this agent show up as a selectable option.
	// "tools" controls which tools this agent can use.
	sb.WriteString("---\n")
	fmt.Fprintf(&sb, "description: %s\n", strings.TrimSpace(persona.Description))
	sb.WriteString("mode: primary\n")
	sb.WriteString("tools:\n")
	sb.WriteString("  skill: true\n")

	// MCP server tool access — OpenCode uses wildcard patterns
	// like server-name_*: true to enable all tools from a server.
	for _, server := range g.MCPServers {
		fmt.Fprintf(&sb, "  %s_*: true\n", server)
	}

	sb.WriteString("---\n\n")

	// Agent title
	fmt.Fprintf(&sb, "# %s\n\n", persona.Name)

	// Body text
	fmt.Fprintf(&sb, "You are operating as the **%s** persona with these capabilities:\n\n", persona.Name)

	// List each package as a skill reference.
	// We include the skill name in parentheses because OpenCode's
	// `skill` tool uses name-based discovery — mentioning the
	// skill name helps the LLM know what to invoke.
	for _, rp := range g.Resolved.Packages {
		name := rp.Package.Name
		desc := rp.Package.Description
		if desc == "" {
			desc = "No description"
		}
		displayName := toTitleCase(name)
		fmt.Fprintf(&sb, "- %s (skill: %s) — %s\n", displayName, name, desc)
	}

	// Persona-level instructions
	if persona.Instructions != "" {
		sb.WriteString("\n")
		sb.WriteString(strings.TrimSpace(persona.Instructions))
		sb.WriteString("\n")
	}

	return sb.String()
}
