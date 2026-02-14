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
)

func newUninstallCommand(content fs.FS) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove installed agents and skills from your repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			packageFlag, _ := cmd.Flags().GetString("package")
			forFlag, _ := cmd.Flags().GetString("for")

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

			jsonFlag, _ := cmd.Flags().GetBool("json")

			if dryRun && !jsonFlag {
				_, _ = color.New(color.FgYellow, color.Bold).Println("Dry run - no files will be removed")
				fmt.Println()
			}

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

			// JSON output — handle AGENTS.md non-interactively, then marshal
			if jsonFlag {
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

			green := color.New(color.FgGreen)
			yellow := color.New(color.FgYellow)
			red := color.New(color.FgRed)
			dim := color.New(color.Faint)
			bold := color.New(color.Bold)

			for _, f := range combinedResult.Removed {
				if dryRun {
					_, _ = yellow.Printf("  would remove: %s\n", f)
				} else {
					_, _ = green.Printf("  removed: %s\n", f)
				}
			}
			for _, f := range combinedResult.NotFound {
				_, _ = dim.Printf("  not found: %s\n", f)
			}
			for _, d := range combinedResult.DirsCleaned {
				_, _ = dim.Printf("  cleaned dir: %s\n", d)
			}
			for _, e := range combinedResult.Errors {
				_, _ = red.Fprintf(os.Stderr, "  error: %s\n", e)
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

	return cmd
}
