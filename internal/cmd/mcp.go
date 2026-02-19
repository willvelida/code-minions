package cmd

import (
	"github.com/spf13/cobra"
)

func newMCPCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Manage MCP server configurations across coding assistants",
		Long: `Translate and inspect MCP (Model Context Protocol) server configurations
between coding assistants such as GitHub Copilot, Claude Code, Cursor, and OpenCode.

Sub-commands allow you to translate server configs from one assistant's format
to another, or list the MCP servers currently configured for a specific assistant.`,
	}

	cmd.AddCommand(newMCPTranslateCommand())
	cmd.AddCommand(newMCPListCommand())

	return cmd
}
