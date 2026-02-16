package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/willvelida/code-minions/internal/assistant"
	"github.com/willvelida/code-minions/internal/mcp"
)

func newMCPListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List MCP servers configured for a coding assistant",
		Long: `Read and display the MCP server configurations for a specific coding
assistant. Servers are shown in canonical format with their transport
type, command/URL, and environment variables.

Use --for to specify which assistant's config to read.`,
		Example: `  # List Copilot's MCP servers
  code-minions mcp list --for copilot

  # List Claude's MCP servers as JSON
  code-minions mcp list --for claude --json

  # List from a specific directory
  code-minions mcp list --for opencode --target ./my-project`,
		RunE: func(cmd *cobra.Command, args []string) error {
			forAssistant, _ := cmd.Flags().GetString("for")
			target, _ := cmd.Flags().GetString("target")

			if forAssistant == "" {
				return fmt.Errorf("--for is required")
			}

			if target == "" {
				target = "."
			}

			mode := getOutputMode(cmd)

			cfg, warnings, err := mcp.ListServers(forAssistant, target)
			if err != nil {
				return err
			}

			type serverEntry struct {
				Name      string              `json:"name"`
				Transport mcp.ServerTransport `json:"transport"`
				Command   string              `json:"command,omitempty"`
				Args      []string            `json:"args,omitempty"`
				URL       string              `json:"url,omitempty"`
				Headers   map[string]string   `json:"headers,omitempty"`
				Env       map[string]string   `json:"env,omitempty"`
			}

			var entries []serverEntry
			for _, name := range sortedKeys(cfg.Servers) {
				s := cfg.Servers[name]
				entries = append(entries, serverEntry{
					Name:      name,
					Transport: s.Transport,
					Command:   s.Command,
					Args:      nonNil(s.Args),
					URL:       s.URL,
					Headers:   s.Headers,
					Env:       s.Env,
				})
			}

			if mode == OutputJSON {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(struct {
					Assistant string        `json:"assistant"`
					Servers   []serverEntry `json:"servers"`
					Warnings  []string      `json:"warnings"`
				}{
					Assistant: forAssistant,
					Servers:   entries,
					Warnings:  nonNil(warnings),
				})
			}

			quietWarning(cmd, mode)

			w := cmd.OutOrStdout()
			bold := color.New(color.Bold)
			cyan := color.New(color.FgCyan)
			dim := color.New(color.Faint)
			yellow := color.New(color.FgYellow)

			asstCfg, _ := assistant.Get(forAssistant)
			label := forAssistant
			if asstCfg != nil {
				label = asstCfg.Description
			}

			_, _ = bold.Fprintf(w, "\nMCP Servers for %s\n", label)
			_, _ = dim.Fprintln(w, strings.Repeat("-", 50))

			if len(entries) == 0 {
				_, _ = dim.Fprintln(w, "  (no MCP servers configured)")
			}

			for _, e := range entries {
				_, _ = cyan.Fprintf(w, "  %s", e.Name)
				_, _ = dim.Fprintf(w, " (%s)\n", e.Transport)

				if e.Command != "" {
					_, _ = fmt.Fprintf(w, "    command: %s", e.Command)
					if len(e.Args) > 0 {
						_, _ = fmt.Fprintf(w, " %s", strings.Join(e.Args, " "))
					}
					_, _ = fmt.Fprintln(w)
				}
				if e.URL != "" {
					_, _ = fmt.Fprintf(w, "    url: %s\n", e.URL)
				}
				if len(e.Env) > 0 {
					envKeys := sortedKeys(e.Env)
					_, _ = fmt.Fprintf(w, "    env: %s\n", strings.Join(envKeys, ", "))
				}

				verbosePrintf(cmd, mode, "    headers: %d, env: %d\n", len(e.Headers), len(e.Env))
			}

			for _, warn := range warnings {
				_, _ = yellow.Fprintf(w, "  ⚠ %s\n", warn)
			}

			_, _ = fmt.Fprintln(w)
			return nil
		},
	}

	assistants := assistant.List()
	validNames := strings.Join(assistants, ", ")

	cmd.Flags().String("for", "", "Assistant to read from ("+validNames+")")
	cmd.Flags().String("target", ".", "Target directory (default: current directory)")

	_ = cmd.MarkFlagRequired("for")

	return cmd
}
