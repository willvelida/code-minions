package cmd

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/willvelida/code-minions/internal/assistant"
	"github.com/willvelida/code-minions/internal/installer"
)

func newInstallCommand(content fs.FS) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install agents and skills into your repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			force, _ := cmd.Flags().GetBool("force")
			packageFlag, _ := cmd.Flags().GetString("package")
			forFlag, _ := cmd.Flags().GetString("for")

			// If --for is set, look up the assistant config and build a path mapper
			var pathMapper func(string) string
			if forFlag != "" {
				cfg, err := assistant.Get(forFlag)
				if err != nil {
					return err
				}
				pathMapper = cfg.NewPathMapper()
			}

			// Build the list of package directories to install
			packageDirs, err := buildPackageList(content, packageFlag)
			if err != nil {
				return err
			}

			if dryRun {
				color.New(color.FgYellow, color.Bold).Println("Dry run - no files will be written")
				fmt.Println()
			}

			// Install each package (with prefix stripping)
			combinedResult := &installer.Result{}
			for _, pkgDir := range packageDirs {
				inst := &installer.Installer{
					Content:     content,
					Target:      target,
					Force:       force,
					DryRun:      dryRun,
					StripPrefix: pkgDir,
					PathMapper:  pathMapper,
				}
				result, err := inst.Install([]string{pkgDir})
				if err != nil {
					return fmt.Errorf("installation failed: %w", err)
				}
				combinedResult.Copied = append(combinedResult.Copied, result.Copied...)
				combinedResult.Skipped = append(combinedResult.Skipped, result.Skipped...)
				combinedResult.Errors = append(combinedResult.Errors, result.Errors...)
			}

			// Create AGENTS.md if installing packages
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

				action, err := handler.OnInstall(agentsMDPath, []byte(installer.DefaultAgentsMDContent))
				if err != nil {
					combinedResult.Errors = append(combinedResult.Errors, fmt.Sprintf("AGENTS.md: %v", err))
				} else if action == "created" {
					combinedResult.Copied = append(combinedResult.Copied, agentsMDPath)
				} else {
					combinedResult.Skipped = append(combinedResult.Skipped, agentsMDPath)
				}
			}

			// JSON output
			jsonFlag, _ := cmd.Flags().GetBool("json")
			if jsonFlag {
				result := struct {
					Copied  []string `json:"copied"`
					Skipped []string `json:"skipped"`
					Errors  []string `json:"errors"`
					Summary struct {
						Copied  int `json:"copied"`
						Skipped int `json:"skipped"`
						Errors  int `json:"errors"`
					} `json:"summary"`
				}{
					Copied:  combinedResult.Copied,
					Skipped: combinedResult.Skipped,
					Errors:  combinedResult.Errors,
				}
				result.Summary.Copied = len(combinedResult.Copied)
				result.Summary.Skipped = len(combinedResult.Skipped)
				result.Summary.Errors = len(combinedResult.Errors)
				if err := json.NewEncoder(cmd.OutOrStdout()).Encode(result); err != nil {
					return err
				}
				if len(combinedResult.Errors) > 0 {
					return fmt.Errorf("installation completed with %d errors", len(combinedResult.Errors))
				}
				return nil
			}

			green := color.New(color.FgGreen)
			yellow := color.New(color.FgYellow)
			red := color.New(color.FgRed)
			bold := color.New(color.Bold)

			// Print results
			for _, f := range combinedResult.Copied {
				if dryRun {
					yellow.Printf("  would copy: %s\n", f)
				} else {
					green.Printf("  copied: %s\n", f)
				}
			}
			for _, f := range combinedResult.Skipped {
				yellow.Printf("  skipped (exists): %s\n", f)
			}
			for _, e := range combinedResult.Errors {
				red.Fprintf(os.Stderr, "  error: %s\n", e)
			}

			// Summary
			fmt.Println()
			bold.Printf("%d copied, %d skipped, %d errors\n",
				len(combinedResult.Copied), len(combinedResult.Skipped), len(combinedResult.Errors))

			if len(combinedResult.Errors) > 0 {
				return fmt.Errorf("installation completed with %d errors", len(combinedResult.Errors))
			}

			return nil
		},
	}

	cmd.Flags().String("target", ".", "Target directory for installation")
	cmd.Flags().String("package", "", "Comma-separated list of packages to install (omit to install all)")
	cmd.Flags().String("for", "", "Target coding assistant (copilot, claude, opencode)")
	cmd.Flags().Bool("dry-run", false, "Show what would be installed without writing files")
	cmd.Flags().Bool("force", false, "Overwrite existing files")

	return cmd
}

func buildPackageList(content fs.FS, packageFlag string) ([]string, error) {
	// No flag = install all packages
	if packageFlag == "" {
		var dirs []string
		packages, err := listSubDirs(content, "packages")
		if err != nil {
			return nil, fmt.Errorf("failed to list packages: %w", err)
		}
		for _, pkg := range packages {
			dirs = append(dirs, "packages/"+pkg)
		}
		return dirs, nil
	}

	var dirs []string
	packages := strings.Split(packageFlag, ",")
	for _, pkg := range packages {
		pkg = strings.TrimSpace(pkg)
		if pkg == "" {
			continue
		}
		dirPath := "packages/" + pkg

		if _, err := fs.Stat(content, dirPath); err != nil {
			available, err := listSubDirs(content, "packages")
			if err != nil {
				return nil, fmt.Errorf("package %q not found", pkg)
			}
			return nil, fmt.Errorf("package %q not found\n\nAvailable packages:\n  %s",
				pkg, strings.Join(available, "\n  "))
		}

		dirs = append(dirs, dirPath)
	}

	// De-duplicate dirs
	seen := make(map[string]bool)
	unique := dirs[:0]
	for _, d := range dirs {
		if !seen[d] {
			seen[d] = true
			unique = append(unique, d)
		}
	}

	return unique, nil
}

func listSubDirs(content fs.FS, dir string) ([]string, error) {
	entries, err := fs.ReadDir(content, dir)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	return names, nil
}
