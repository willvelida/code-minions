package cmd

import (
	"fmt"
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

			if dryRun {
				color.New(color.FgYellow, color.Bold).Println("Dry run - no files will be removed")
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

			green := color.New(color.FgGreen)
			yellow := color.New(color.FgYellow)
			red := color.New(color.FgRed)
			dim := color.New(color.Faint)
			bold := color.New(color.Bold)

			for _, f := range combinedResult.Removed {
				if dryRun {
					yellow.Printf("  would remove: %s\n", f)
				} else {
					green.Printf("  removed: %s\n", f)
				}
			}
			for _, f := range combinedResult.NotFound {
				dim.Printf("  not found: %s\n", f)
			}
			for _, d := range combinedResult.DirsCleaned {
				dim.Printf("  cleaned dir: %s\n", d)
			}
			for _, e := range combinedResult.Errors {
				red.Fprintf(os.Stderr, "  error: %s\n", e)
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
					red.Fprintf(os.Stderr, "  error: %s\n", err)
				} else if action == "removed" {
					green.Printf("  removed: %s\n", agentsMDPath)
				}
			}

			fmt.Println()
			bold.Printf("%d removed, %d not found, %d errors\n",
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
