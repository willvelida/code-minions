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

func newMCPTranslateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "translate",
		Short: "Translate MCP server configs between coding assistants",
		Long: `Read MCP server configurations from one coding assistant's format and
write them into another's. This enables sharing MCP server setups across
coding assistants without manual config editing.

The source config is read from the --from assistant, translated to the
canonical format, and then written into the --to assistant's config file.

Multiple targets can be specified with comma-separated values, e.g.
--to copilot,opencode.

Existing servers in the target are preserved. Use --force to overwrite
servers that already exist with different configurations. Use --dry-run
to preview changes without writing any files.`,
		Example: `  # Translate all MCP servers from Copilot to Claude
  code-minions mcp translate --from copilot --to claude

  # Translate a specific server only
  code-minions mcp translate --from copilot --to claude --server github

  # Translate to multiple targets
  code-minions mcp translate --from copilot --to claude,opencode

  # Preview what would change
  code-minions mcp translate --from copilot --to claude --dry-run

  # Overwrite conflicting servers
  code-minions mcp translate --from copilot --to claude --force

  # Output results as JSON
  code-minions mcp translate --from copilot --to claude --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			from, _ := cmd.Flags().GetString("from")
			toRaw, _ := cmd.Flags().GetString("to")
			server, _ := cmd.Flags().GetString("server")
			force, _ := cmd.Flags().GetBool("force")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			target, _ := cmd.Flags().GetString("target")

			if from == "" {
				return fmt.Errorf("--from is required")
			}
			if toRaw == "" {
				return fmt.Errorf("--to is required")
			}

			rawTargets := strings.Split(toRaw, ",")
			seen := make(map[string]bool)
			var targets []string
			for _, t := range rawTargets {
				t = strings.TrimSpace(t)
				if t == "" || seen[t] {
					continue
				}
				seen[t] = true
				targets = append(targets, t)
			}
			if len(targets) == 0 {
				return fmt.Errorf("no valid targets specified in --to")
			}

			// Pre-validate: reject targets that match the source to avoid
			// partial side-effects when earlier translations succeed but a
			// later same-as-source target fails.
			for _, to := range targets {
				if to == from {
					return fmt.Errorf("target %q is the same as source --from; cannot translate to self", to)
				}
			}

			if target == "" {
				target = "."
			}

			mode := getOutputMode(cmd)

			var results []*mcp.TranslateResult
			for _, to := range targets {
				verbosePrintf(cmd, mode, "translating from %s to %s...\n", from, to)

				result, err := mcp.Translate(mcp.TranslateOptions{
					From:      from,
					To:        to,
					TargetDir: target,
					Server:    server,
					Force:     force,
					DryRun:    dryRun,
				})
				if err != nil {
					return fmt.Errorf("translate %s → %s: %w", from, to, err)
				}
				results = append(results, result)
			}

			if mode == OutputJSON {
				if len(results) == 1 {
					return json.NewEncoder(cmd.OutOrStdout()).Encode(results[0])
				}
				return json.NewEncoder(cmd.OutOrStdout()).Encode(results)
			}

			if mode == OutputQuiet {
				return nil
			}

			w := cmd.OutOrStdout()
			bold := color.New(color.Bold)
			green := color.New(color.FgGreen)
			yellow := color.New(color.FgYellow)
			dim := color.New(color.Faint)

			for _, r := range results {
				_, _ = bold.Fprintf(w, "\n%s → %s\n", r.SourceAssistant, r.TargetAssistant)
				_, _ = dim.Fprintln(w, strings.Repeat("-", 50))

				if r.DryRun {
					_, _ = yellow.Fprintln(w, "  (dry-run — no files written)")
				}

				_, _ = dim.Fprintf(w, "  config: %s\n", r.ConfigPath)

				if len(r.Merge.Added) > 0 {
					_, _ = green.Fprintf(w, "  added:    %s\n", strings.Join(r.Merge.Added, ", "))
				}
				if len(r.Merge.Skipped) > 0 {
					_, _ = dim.Fprintf(w, "  skipped:  %s\n", strings.Join(r.Merge.Skipped, ", "))
				}
				if len(r.Merge.Conflict) > 0 {
					_, _ = yellow.Fprintf(w, "  conflict: %s (use --force to overwrite)\n", strings.Join(r.Merge.Conflict, ", "))
				}

				if len(r.Merge.Added) == 0 && len(r.Merge.Skipped) == 0 && len(r.Merge.Conflict) == 0 {
					_, _ = dim.Fprintln(w, "  no changes")
				}

				for _, warn := range r.Warnings {
					_, _ = yellow.Fprintf(w, "  ⚠ %s\n", warn)
				}

				verbosePrintf(cmd, mode, "  servers translated: %d\n",
					len(r.Merge.Added)+len(r.Merge.Skipped)+len(r.Merge.Conflict))
			}

			_, _ = fmt.Fprintln(w)
			return nil
		},
	}

	assistants := assistant.List()
	validNames := strings.Join(assistants, ", ")

	cmd.Flags().String("from", "", "Source assistant ("+validNames+")")
	cmd.Flags().String("to", "", "Target assistant(s), comma-separated ("+validNames+")")
	cmd.Flags().String("server", "", "Translate only this server name")
	cmd.Flags().Bool("force", false, "Overwrite conflicting servers in the target")
	cmd.Flags().Bool("dry-run", false, "Preview changes without writing files")
	cmd.Flags().String("target", ".", "Target directory (default: current directory)")

	_ = cmd.MarkFlagRequired("from")
	_ = cmd.MarkFlagRequired("to")

	return cmd
}
