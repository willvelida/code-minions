package installer

import (
	"fmt"
	"strings"

	"github.com/willvelida/code-minions/internal/assistant"
	"github.com/willvelida/code-minions/internal/registry"
)

// CopilotGrouping generates a persona custom agent file for Copilot.
//
// VS Code detects .agent.md files in .github/agents/ and shows them
// in the agents dropdown. The user selects which persona to use
// from this dropdown, and the file's frontmatter controls which
// tools (including MCP servers) are available.
//
// Generated file: .github/agents/<persona-name>.agent.md
//
// Example output:
//
//	---
//	description: A senior developer persona
//	tools:
//	  - github/*
//	  - linear/*
//	---
//
//	# senior-dev
//
//	You are operating as the **senior-dev** persona with these capabilities:
//
//	- **Git Workflow** — Agent: .github/agents/git-workflow.agent.md
//	  Skill: skills/git-workflow/SKILL.md
//	  ...
type CopilotGrouping struct {
	Config     *assistant.Config
	Resolved   *registry.ResolvedPersona
	Target     string
	DryRun     bool
	Force      bool
	MCPServers []string // MCP server names (set via SetMCPServers)
}

// SetMCPServers provides MCP server names for frontmatter generation.
// Copilot uses the format: tools: ['server-name/*'] to grant access
// to all tools from an MCP server.
func (g *CopilotGrouping) SetMCPServers(servers []string) {
	g.MCPServers = servers
}

// Generate creates the persona custom agent file at
// .github/agents/<persona-name>.agent.md
//
// This replaces the previous AGENTS.md section approach. The
// .agent.md format supports YAML frontmatter with a tools: field,
// which is required for MCP tool routing. It also appears in
// VS Code's agents dropdown, making persona selection intuitive.
func (g *CopilotGrouping) Generate() ([]string, error) {
	persona := g.Resolved.Persona

	content := g.buildAgentContent()

	// The file goes in the assistant's agent directory.
	// For Copilot: .github/agents/senior-dev.agent.md
	outputPath := g.Config.AgentDir + "/" + persona.Name + ".agent.md"

	err := writeGeneratedFile(g.Target, outputPath, []byte(content), g.DryRun, g.Force)
	if err != nil {
		return nil, err
	}

	return []string{outputPath}, nil
}

// buildAgentContent generates the full Markdown content for the
// persona custom agent, including YAML frontmatter.
func (g *CopilotGrouping) buildAgentContent() string {
	persona := g.Resolved.Persona

	var sb strings.Builder

	// YAML frontmatter — VS Code reads this to understand the
	// agent's purpose and available tools.
	sb.WriteString("---\n")
	fmt.Fprintf(&sb, "description: %s\n", strings.TrimSpace(persona.Description))

	// Tools list: MCP server references use the <server>/* format
	// to grant access to all tools from that server.
	if len(g.MCPServers) > 0 {
		sb.WriteString("tools:\n")
		for _, server := range g.MCPServers {
			fmt.Fprintf(&sb, "  - %s/*\n", server)
		}
	}

	sb.WriteString("---\n\n")

	// Agent title
	fmt.Fprintf(&sb, "# %s\n\n", persona.Name)

	// Body text — describes the persona's capabilities and routes
	// to individual package agent and skill files.
	fmt.Fprintf(&sb, "You are operating as the **%s** persona with these capabilities:\n\n", persona.Name)

	for _, rp := range g.Resolved.Packages {
		name := rp.Package.Name
		displayName := toTitleCase(name)

		// Agent path (as Copilot would see it after path mapping)
		agentPath := g.Config.AgentDir + "/" + name + ".agent.md"
		skillPath := g.Config.SkillDir + "/" + name + "/SKILL.md"

		fmt.Fprintf(&sb, "- **%s** — Agent: [%s](%s), Skill: [%s](%s)", displayName, agentPath, agentPath, skillPath, skillPath)

		if rp.Package.Description != "" {
			fmt.Fprintf(&sb, " — %s", rp.Package.Description)
		}
		sb.WriteString("\n")
	}

	// Persona-level instructions
	if persona.Instructions != "" {
		sb.WriteString("\n")
		sb.WriteString(strings.TrimSpace(persona.Instructions))
		sb.WriteString("\n")
	}

	return sb.String()
}
