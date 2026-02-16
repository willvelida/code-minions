package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/willvelida/code-minions/internal/assistant"
	"github.com/willvelida/code-minions/internal/installer"
	"github.com/willvelida/code-minions/internal/mcp"
)

func newUninstallCommand(content fs.FS) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove installed agents and skills from your repository",
		Long: `Remove previously installed agent and skill files from your repository.
Only files at paths derived from the built-in package registry are
targeted for removal. This may delete locally modified versions of
those files, but files outside these known paths are not touched.

In interactive mode a confirmation prompt is shown before any files are
deleted. Non-interactive output modes (--json, --quiet) do not prompt
and require --yes to proceed. Use --dry-run to see what would be
removed without deleting anything.

The --for flag is required when uninstalling all packages, so that
files are found in the correct assistant-specific location.`,
		Example: `  # Remove all packages (interactive confirmation prompt)
  code-minions uninstall --for copilot

  # Skip the confirmation prompt (for CI/scripts)
  code-minions uninstall --for copilot --yes

  # Remove a specific package
  code-minions uninstall --package git-workflow --for copilot

  # Preview what would be removed
  code-minions uninstall --dry-run --for copilot`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			packageFlag, _ := cmd.Flags().GetString("package")
			forFlag, _ := cmd.Flags().GetString("for")
			yesFlag, _ := cmd.Flags().GetBool("yes")

			if packageFlag == "" && forFlag == "" {
				return fmt.Errorf("when uninstalling everything, --for is required to identify the correct file locations\n\nUsage: code-minions uninstall --for <assistant>\n\nAvailable assistants: %s",
					strings.Join(assistant.List(), ", "))
			}

			var pathMapper func(string) string
			if forFlag != "" {
				cfg, err := assistant.Get(forFlag)
				if err != nil {
					return err
				}
				pathMapper = cfg.NewPathMapper()
			}

			packageDirs, err := buildPackageList(content, packageFlag)
			if err != nil {
				return err
			}

			mode := getOutputMode(cmd)

			if dryRun && (mode == OutputNormal || mode == OutputVerbose) {
				_, _ = color.New(color.FgYellow, color.Bold).Println("Dry run - no files will be removed")
				fmt.Println()
			}

			// --- Confirmation gate (skip for dry-run) ---
			if !dryRun && !yesFlag {
				// Phase 1: pre-flight scan to count files that would be removed
				scanResult := &installer.UninstallResult{}
				for _, pkgDir := range packageDirs {
					inst := &installer.Installer{
						Content:     content,
						Target:      target,
						DryRun:      true, // always dry-run for the scan
						StripPrefix: pkgDir,
						PathMapper:  pathMapper,
					}
					result, err := inst.Uninstall([]string{pkgDir})
					if err != nil {
						return fmt.Errorf("pre-flight scan failed: %w", err)
					}
					scanResult.Removed = append(scanResult.Removed, result.Removed...)
				}

				fileCount := len(scanResult.Removed)
				if fileCount > 0 {
					switch mode {
					case OutputJSON:
						// JSON mode is non-interactive — require --yes
						errResult := struct {
							Error     string `json:"error"`
							FileCount int    `json:"file_count"`
							Hint      string `json:"hint"`
						}{
							Error:     "confirmation required",
							FileCount: fileCount,
							Hint:      "use --yes to skip",
						}
						if err := json.NewEncoder(cmd.OutOrStdout()).Encode(errResult); err != nil {
							return fmt.Errorf("failed to write JSON error response: %w", err)
						}
						// Prevent Cobra from printing a second, non-JSON error line.
						cmd.SilenceErrors = true
						cmd.SilenceUsage = true
						return fmt.Errorf("confirmation required (use --yes to skip)")

					case OutputQuiet:
						// Quiet mode is non-interactive — require --yes
						cmd.SilenceErrors = true
						cmd.SilenceUsage = true
						_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "error: confirmation required — this will remove %d files (use --yes to skip)\n", fileCount)
						return fmt.Errorf("confirmation required (use --yes to skip)")

					default:
						// Normal/verbose — interactive prompt
						if !isInteractiveFunc() {
							return fmt.Errorf("confirmation required (use --yes to skip)")
						}

						promptMsg := fmt.Sprintf("This will remove %d files. Continue? [y/N]: ", fileCount)
						confirmed, err := confirmPrompt(cmd.InOrStdin(), cmd.OutOrStdout(), promptMsg)
						if err != nil {
							return err
						}
						if !confirmed {
							_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
							return nil
						}
					}
				}
			}

			// --- Phase 2: actual removal ---
			combinedResult := &installer.UninstallResult{}
			for _, pkgDir := range packageDirs {
				inst := &installer.Installer{
					Content:     content,
					Target:      target,
					DryRun:      dryRun,
					StripPrefix: pkgDir,
					PathMapper:  pathMapper,
				}
				result, err := inst.Uninstall([]string{pkgDir})
				if err != nil {
					return fmt.Errorf("uninstallation failed: %w", err)
				}
				combinedResult.Removed = append(combinedResult.Removed, result.Removed...)
				combinedResult.NotFound = append(combinedResult.NotFound, result.NotFound...)
				combinedResult.Errors = append(combinedResult.Errors, result.Errors...)
				combinedResult.DirsCleaned = append(combinedResult.DirsCleaned, result.DirsCleaned...)
			}

			// --- MCP server removal ---
			// Look up installed MCP server names from the manifest and remove
			// them from the assistant's config file.
			var mcpUninstallResults []*mcp.UninstallResult
			if forFlag != "" {
				translator, err := mcp.NewTranslator(forFlag)
				if err != nil {
					combinedResult.Errors = append(combinedResult.Errors, fmt.Sprintf("MCP: %v", err))
				} else {
					manifest, err := installer.LoadManifest(target)
					if err != nil {
						combinedResult.Errors = append(combinedResult.Errors, fmt.Sprintf("MCP manifest load: %v", err))
					} else {
						for _, pkgDir := range packageDirs {
							pkgName := strings.TrimPrefix(pkgDir, "packages/")
							// Find the package in the manifest to get its MCPServers
							var serverNames []string
							for _, p := range manifest.Packages {
								if p.Name == pkgName {
									serverNames = p.MCPServers
									break
								}
							}
							if len(serverNames) > 0 {
								mcpResult, err := mcp.Uninstall(target, translator, serverNames, dryRun)
								if err != nil {
									combinedResult.Errors = append(combinedResult.Errors, fmt.Sprintf("MCP (%s): %v", pkgName, err))
									continue
								}
								if mcpResult != nil {
									mcpUninstallResults = append(mcpUninstallResults, mcpResult)
								}
							}
						}
					}
				}
			}

			// Update manifest to remove uninstalled packages (skip for dry-run)
			if !dryRun && len(combinedResult.Removed) > 0 {
				manifest, err := installer.LoadManifest(target)
				if err != nil {
					combinedResult.Errors = append(combinedResult.Errors, fmt.Sprintf("manifest load: %v", err))
				} else {
					for _, pkgDir := range packageDirs {
						pkgName := strings.TrimPrefix(pkgDir, "packages/")
						installer.RecordUninstall(manifest, pkgName)
					}
					if err := installer.SaveManifest(target, manifest); err != nil {
						combinedResult.Errors = append(combinedResult.Errors, fmt.Sprintf("manifest: %v", err))
					}
				}
			}

			// JSON output — handle AGENTS.md non-interactively, then marshal
			if mode == OutputJSON {
				if len(packageDirs) > 0 {
					agentsMDPath := "AGENTS.md"
					if pathMapper != nil {
						agentsMDPath = pathMapper("agents/AGENTS.md")
					}
					handler := &installer.AgentsMDHandler{
						Target: target,
						DryRun: dryRun,
						Stdin:  strings.NewReader("n\n"), // non-interactive: decline removal
						Stdout: io.Discard,
					}
					action, err := handler.OnUninstall(agentsMDPath)
					if err != nil {
						combinedResult.Errors = append(combinedResult.Errors, fmt.Sprintf("AGENTS.md: %v", err))
					} else if action == "removed" {
						combinedResult.Removed = append(combinedResult.Removed, agentsMDPath)
					}
				}

				removed := combinedResult.Removed
				if removed == nil {
					removed = []string{}
				}
				notFound := combinedResult.NotFound
				if notFound == nil {
					notFound = []string{}
				}
				errs := combinedResult.Errors
				if errs == nil {
					errs = []string{}
				}
				dirsCleaned := combinedResult.DirsCleaned
				if dirsCleaned == nil {
					dirsCleaned = []string{}
				}
				result := struct {
					Removed     []string `json:"removed"`
					NotFound    []string `json:"not_found"`
					Errors      []string `json:"errors"`
					DirsCleaned []string `json:"dirs_cleaned"`
					Summary     struct {
						Removed  int `json:"removed"`
						NotFound int `json:"not_found"`
						Errors   int `json:"errors"`
					} `json:"summary"`
				}{
					Removed:     removed,
					NotFound:    notFound,
					Errors:      errs,
					DirsCleaned: dirsCleaned,
				}
				result.Summary.Removed = len(combinedResult.Removed)
				result.Summary.NotFound = len(combinedResult.NotFound)
				result.Summary.Errors = len(combinedResult.Errors)
				if err := json.NewEncoder(cmd.OutOrStdout()).Encode(result); err != nil {
					return err
				}
				if len(combinedResult.Errors) > 0 {
					return fmt.Errorf("uninstallation completed with %d errors", len(combinedResult.Errors))
				}
				return nil
			}

			// Quiet mode — handle AGENTS.md non-interactively, only report errors
			if mode == OutputQuiet {
				if len(packageDirs) > 0 {
					agentsMDPath := "AGENTS.md"
					if pathMapper != nil {
						agentsMDPath = pathMapper("agents/AGENTS.md")
					}
					handler := &installer.AgentsMDHandler{
						Target: target,
						DryRun: dryRun,
						Stdin:  strings.NewReader("n\n"),
						Stdout: io.Discard,
					}
					action, err := handler.OnUninstall(agentsMDPath)
					if err != nil {
						combinedResult.Errors = append(combinedResult.Errors, fmt.Sprintf("AGENTS.md: %v", err))
					} else if action == "removed" {
						combinedResult.Removed = append(combinedResult.Removed, agentsMDPath)
					}
				}
				for _, e := range combinedResult.Errors {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "  error: %s\n", e)
				}
				if len(combinedResult.Errors) > 0 {
					return fmt.Errorf("uninstallation completed with %d errors", len(combinedResult.Errors))
				}
				return nil
			}

			green := color.New(color.FgGreen)
			yellow := color.New(color.FgYellow)
			red := color.New(color.FgRed)
			dim := color.New(color.Faint)
			bold := color.New(color.Bold)

			// Verbose: show package list
			verbosePrintf(cmd, mode, "packages: %v\n", packageDirs)

			for _, f := range combinedResult.Removed {
				if dryRun {
					_, _ = yellow.Printf("  would remove: %s\n", f)
				} else {
					_, _ = green.Printf("  removed: %s\n", f)
				}
			}
			for _, f := range combinedResult.NotFound {
				_, _ = dim.Printf("  not found: %s\n", f)
				verbosePrintf(cmd, mode, "    → file does not exist in target\n")
			}
			for _, d := range combinedResult.DirsCleaned {
				_, _ = dim.Printf("  cleaned dir: %s\n", d)
			}
			for _, e := range combinedResult.Errors {
				_, _ = red.Fprintf(os.Stderr, "  error: %s\n", e)
			}

			// MCP server removal results
			if len(mcpUninstallResults) > 0 {
				cyan := color.New(color.FgCyan)
				fmt.Println()
				_, _ = bold.Println("MCP servers:")
				for _, mr := range mcpUninstallResults {
					for _, s := range mr.Removed {
						if dryRun {
							_, _ = yellow.Printf("  would remove from %s: %s\n", mr.ConfigPath, s)
						} else {
							_, _ = cyan.Printf("  removed from %s: %s\n", mr.ConfigPath, s)
						}
					}
					for _, s := range mr.NotFound {
						_, _ = dim.Printf("  not found in %s: %s\n", mr.ConfigPath, s)
					}
				}
			}

			if len(packageDirs) > 0 {
				agentsMDPath := "AGENTS.md"
				if pathMapper != nil {
					agentsMDPath = pathMapper("agents/AGENTS.md")
				}

				handler := &installer.AgentsMDHandler{
					Target: target,
					DryRun: dryRun,
					Stdin:  os.Stdin,
					Stdout: os.Stdout,
				}

				action, err := handler.OnUninstall(agentsMDPath)
				if err != nil {
					combinedResult.Errors = append(combinedResult.Errors, fmt.Sprintf("AGENTS.md: %v", err))
					_, _ = red.Fprintf(os.Stderr, "  error: %s\n", err)
				} else if action == "removed" {
					_, _ = green.Printf("  removed: %s\n", agentsMDPath)
				}
			}

			fmt.Println()
			_, _ = bold.Printf("%d removed, %d not found, %d errors\n",
				len(combinedResult.Removed), len(combinedResult.NotFound), len(combinedResult.Errors))

			if len(combinedResult.Errors) > 0 {
				return fmt.Errorf("uninstallation completed with %d errors", len(combinedResult.Errors))
			}

			return nil
		},
	}

	cmd.Flags().String("target", ".", "Target directory to uninstall from")
	cmd.Flags().String("package", "", "Comma-separated list of packages to uninstall (omit to uninstall all)")
	cmd.Flags().String("for", "", "Target coding assistant (copilot, claude, opencode)")
	cmd.Flags().Bool("dry-run", false, "Show what would be removed without deleting files")
	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt and proceed with removal")

	return cmd
}
