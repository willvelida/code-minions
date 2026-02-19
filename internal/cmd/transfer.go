package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/willvelida/code-minions/internal/assistant"
	"github.com/willvelida/code-minions/internal/installer"
	"github.com/willvelida/code-minions/internal/mcp"
)

func newTransferCommand(content fs.FS) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer",
		Short: "Transfer agent and skill files between coding assistant layouts",
		Long: `Copy agent and skill files from one coding assistant's directory layout
to another. This enables switching between assistants (e.g. GitHub Copilot
to Claude Code) or supporting multiple assistants simultaneously without
manually moving files.

Files are copied by default — the source layout is left in place. Use
--cleanup to delete the source files after a successful copy.

MCP server configurations are automatically translated between the source
and target assistant formats when present.

An AGENTS.md routing file is always regenerated (not copied) in the target
layout because it contains assistant-specific paths.`,

		Example: `  # Transfer from Copilot to Claude
  code-minions transfer --from copilot --to claude

  # Preview what would be transferred
  code-minions transfer --from copilot --to claude --dry-run

  # Overwrite existing files at the destination
  code-minions transfer --from copilot --to claude --force

  # Transfer and remove old Copilot layout
  code-minions transfer --from copilot --to claude --cleanup

  # Transfer in a specific directory
  code-minions transfer --from claude --to opencode --target ./my-project`,
		RunE: func(cmd *cobra.Command, args []string) error {
			from, _ := cmd.Flags().GetString("from")
			to, _ := cmd.Flags().GetString("to")
			target, _ := cmd.Flags().GetString("target")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			force, _ := cmd.Flags().GetBool("force")
			cleanup, _ := cmd.Flags().GetBool("cleanup")

			if from == "" {
				return fmt.Errorf("--from is required\n\nUsage: code-minions transfer --from <assistant> --to <assistant>\n\nAvailable assistants: %s",
					strings.Join(assistant.List(), ", "))
			}
			if to == "" {
				return fmt.Errorf("--to is required\n\nUsage: code-minions transfer --from <assistant> --to <assistant>\n\nAvailable assistants: %s",
					strings.Join(assistant.List(), ", "))
			}

			if target == "" {
				target = "."
			}

			mode := getOutputMode(cmd)

			if dryRun && (mode == OutputNormal || mode == OutputVerbose) {
				_, _ = color.New(color.FgYellow, color.Bold).Println("Dry run - no files will be written")
				fmt.Println()
			}

			// Step 1: Transfer agent/skill files
			transferResult, err := installer.Transfer(installer.TransferOptions{
				FromAssistant: from,
				ToAssistant:   to,
				TargetDir:     target,
				Force:         force,
				DryRun:        dryRun,
				Cleanup:       cleanup,
			})
			if err != nil {
				return err
			}

			// Step 2: Translate MCP config (best-effort — don't fail if no MCP config)
			var mcpResult *mcp.TranslateResult
			mcpResult, err = mcp.Translate(mcp.TranslateOptions{
				From:      from,
				To:        to,
				TargetDir: target,
				Force:     force,
				DryRun:    dryRun,
			})
			if err != nil {
				// MCP translation failure is non-fatal — add as warning
				transferResult.Warnings = append(transferResult.Warnings,
					fmt.Sprintf("MCP config translation skipped: %v", err))
				mcpResult = nil
			}

			// Step 3: Regenerate AGENTS.md in target layout.
			// Remove any existing copy first so OnInstall always writes a fresh
			// file with the correct assistant-specific paths.
			toCfg, err := assistant.Get(to)
			if err != nil {
				return fmt.Errorf("failed to get assistant config for %s: %w", to, err)
			}
			agentsMDPath := toCfg.NewPathMapper()("agents/AGENTS.md")

			if !dryRun {
				_ = os.Remove(filepath.Join(target, agentsMDPath))
			}

			handlerStdout := io.Writer(os.Stdout)
			if mode == OutputJSON || mode == OutputQuiet {
				handlerStdout = io.Discard
			}
			handler := &installer.AgentsMDHandler{
				Target: target,
				DryRun: dryRun,
				Stdin:  os.Stdin,
				Stdout: handlerStdout,
			}

			action, agentsMDErr := handler.OnInstall(agentsMDPath, []byte(installer.DefaultAgentsMDContent))
			if agentsMDErr != nil {
				transferResult.Errors = append(transferResult.Errors, fmt.Sprintf("AGENTS.md: %v", agentsMDErr))
			} else if action == "created" {
				transferResult.Copied = append(transferResult.Copied, agentsMDPath)
			}

			// --- Output ---
			return renderTransferOutput(cmd, mode, from, to, dryRun, cleanup, transferResult, mcpResult)
		},
	}

	cmd.Flags().String("from", "", "Source coding assistant ("+assistant.FlagUsage()+")")
	cmd.Flags().String("to", "", "Target coding assistant ("+assistant.FlagUsage()+")")
	cmd.Flags().String("target", ".", "Target directory")
	cmd.Flags().Bool("dry-run", false, "Preview changes without writing files")
	cmd.Flags().Bool("force", false, "Overwrite existing files at the destination")
	cmd.Flags().Bool("cleanup", false, "Delete source agent/skill files after successful copy")

	return cmd
}

func renderTransferOutput(
	cmd *cobra.Command,
	mode OutputMode,
	from, to string,
	dryRun, cleanup bool,
	result *installer.TransferResult,
	mcpResult *mcp.TranslateResult,
) error {
	// JSON output
	if mode == OutputJSON {
		type mcpJSON struct {
			ConfigPath string   `json:"config_path"`
			Added      []string `json:"added"`
			Skipped    []string `json:"skipped"`
			Conflicts  []string `json:"conflicts"`
			Warnings   []string `json:"warnings"`
		}

		type filesJSON struct {
			Copied   []string `json:"copied"`
			Skipped  []string `json:"skipped"`
			Errors   []string `json:"errors"`
			Cleaned  []string `json:"cleaned"`
			Warnings []string `json:"warnings"`
		}

		output := struct {
			From    string    `json:"from"`
			To      string    `json:"to"`
			DryRun  bool      `json:"dry_run"`
			Cleanup bool      `json:"cleanup"`
			Files   filesJSON `json:"files"`
			MCP     *mcpJSON  `json:"mcp,omitempty"`
		}{
			From:    from,
			To:      to,
			DryRun:  dryRun,
			Cleanup: cleanup,
			Files: filesJSON{
				Copied:   nonNil(result.Copied),
				Skipped:  nonNil(result.Skipped),
				Errors:   nonNil(result.Errors),
				Cleaned:  nonNil(result.Cleaned),
				Warnings: nonNil(result.Warnings),
			},
		}

		if mcpResult != nil && mcpResult.Merge != nil {
			output.MCP = &mcpJSON{
				ConfigPath: mcpResult.ConfigPath,
				Added:      nonNil(mcpResult.Merge.Added),
				Skipped:    nonNil(mcpResult.Merge.Skipped),
				Conflicts:  nonNil(mcpResult.Merge.Conflict),
				Warnings:   nonNil(mcpResult.Merge.Warnings),
			}
		}

		if err := json.NewEncoder(cmd.OutOrStdout()).Encode(output); err != nil {
			return err
		}
		if len(result.Errors) > 0 {
			return fmt.Errorf("transfer completed with %d errors", len(result.Errors))
		}
		return nil
	}

	// Quiet mode
	if mode == OutputQuiet {
		for _, e := range result.Errors {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "  error: %s\n", e)
		}
		if len(result.Errors) > 0 {
			return fmt.Errorf("transfer completed with %d errors", len(result.Errors))
		}
		return nil
	}

	// Normal / Verbose mode
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	red := color.New(color.FgRed)
	cyan := color.New(color.FgCyan)
	bold := color.New(color.Bold)

	fromCfg, _ := assistant.Get(from)
	toCfg, _ := assistant.Get(to)
	verbosePrintf(cmd, mode, "transfer: %s → %s\n", fromCfg.Description, toCfg.Description)

	// Warnings
	for _, w := range result.Warnings {
		_, _ = yellow.Printf("  warning: %s\n", w)
	}

	// File results
	for _, f := range result.Copied {
		if dryRun {
			_, _ = yellow.Printf("  would copy: %s\n", f)
		} else {
			_, _ = green.Printf("  copied: %s\n", f)
		}
	}
	for _, f := range result.Skipped {
		_, _ = yellow.Printf("  skipped (exists): %s\n", f)
		verbosePrintf(cmd, mode, "    → use --force to overwrite\n")
	}
	for _, e := range result.Errors {
		_, _ = red.Fprintf(os.Stderr, "  error: %s\n", e)
	}

	// Cleanup results
	if cleanup && len(result.Cleaned) > 0 {
		fmt.Println()
		_, _ = bold.Println("Cleanup:")
		for _, f := range result.Cleaned {
			if dryRun {
				_, _ = yellow.Printf("  would remove: %s\n", f)
			} else {
				_, _ = cyan.Printf("  removed: %s\n", f)
			}
		}
	}

	// MCP results
	if mcpResult != nil && mcpResult.Merge != nil {
		hasOutput := len(mcpResult.Merge.Added) > 0 ||
			len(mcpResult.Merge.Skipped) > 0 ||
			len(mcpResult.Merge.Conflict) > 0 ||
			len(mcpResult.Merge.Warnings) > 0
		if hasOutput {
			fmt.Println()
			_, _ = bold.Println("MCP servers:")
			for _, s := range mcpResult.Merge.Added {
				if dryRun {
					_, _ = yellow.Printf("  would add to %s: %s\n", mcpResult.ConfigPath, s)
				} else {
					_, _ = cyan.Printf("  added to %s: %s\n", mcpResult.ConfigPath, s)
				}
			}
			for _, s := range mcpResult.Merge.Skipped {
				_, _ = yellow.Printf("  skipped (identical): %s in %s\n", s, mcpResult.ConfigPath)
			}
			for _, s := range mcpResult.Merge.Conflict {
				_, _ = yellow.Printf("  conflict: %s in %s (use --force to overwrite)\n", s, mcpResult.ConfigPath)
			}
			for _, w := range mcpResult.Merge.Warnings {
				_, _ = yellow.Printf("  warning: %s\n", w)
			}
		}
	}

	// Summary
	fmt.Println()
	summaryParts := []string{
		fmt.Sprintf("%d copied", len(result.Copied)),
		fmt.Sprintf("%d skipped", len(result.Skipped)),
		fmt.Sprintf("%d errors", len(result.Errors)),
	}
	if cleanup {
		summaryParts = append(summaryParts, fmt.Sprintf("%d cleaned", len(result.Cleaned)))
	}
	_, _ = bold.Println(strings.Join(summaryParts, ", "))

	if len(result.Errors) > 0 {
		return fmt.Errorf("transfer completed with %d errors", len(result.Errors))
	}

	return nil
}
