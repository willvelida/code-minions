package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/willvelida/code-minions/internal/assistant"
	"github.com/willvelida/code-minions/internal/installer"
)

func newUpdateCommand(content fs.FS) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update installed packages and standards with the latest versions",
		Long: `Update overwrites previously installed files with the latest embedded
content. When run with no flags, only packages and standards that are
already installed in the target directory are updated.

AGENTS.md is not modified during updates.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			packageFlag, _ := cmd.Flags().GetString("package")
			standardsFlag, _ := cmd.Flags().GetString("standards")
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

			// When flags are provided, use buildDirList (same as install).
			// When no flags are provided, detect what's already installed.
			var dirs []string
			if packageFlag != "" || standardsFlag != "" {
				var err error
				dirs, err = buildDirList(content, packageFlag, standardsFlag)
				if err != nil {
					return err
				}
			} else {
				var err error
				dirs, err = detectInstalled(content, target, pathMapper)
				if err != nil {
					return err
				}
				if len(dirs) == 0 {
					fmt.Println("No installed packages or standards found. Use 'code-minions install' to install packages or standards first.")
					return nil
				}
			}

			if dryRun {
				color.New(color.FgYellow, color.Bold).Println("Dry run - no files will be written")
				fmt.Println()
			}

			// Separate package dirs from standard dirs
			var packageDirs []string
			var standardDirs []string
			for _, d := range dirs {
				if strings.HasPrefix(d, "packages/") {
					packageDirs = append(packageDirs, d)
				} else {
					standardDirs = append(standardDirs, d)
				}
			}

			// Update each package (with prefix stripping, force always true)
			combinedResult := &installer.Result{}
			for _, pkgDir := range packageDirs {
				inst := &installer.Installer{
					Content:     content,
					Target:      target,
					Force:       true, // Update always overwrites
					DryRun:      dryRun,
					StripPrefix: pkgDir,
					PathMapper:  pathMapper,
				}
				result, err := inst.Install([]string{pkgDir})
				if err != nil {
					return fmt.Errorf("update failed: %w", err)
				}
				combinedResult.Copied = append(combinedResult.Copied, result.Copied...)
				combinedResult.Skipped = append(combinedResult.Skipped, result.Skipped...)
				combinedResult.Errors = append(combinedResult.Errors, result.Errors...)
			}

			// Update standards (force always true)
			if len(standardDirs) > 0 {
				inst := &installer.Installer{
					Content:    content,
					Target:     target,
					Force:      true, // Update always overwrites
					DryRun:     dryRun,
					PathMapper: pathMapper,
				}
				result, err := inst.Install(standardDirs)
				if err != nil {
					return fmt.Errorf("update failed: %w", err)
				}
				combinedResult.Copied = append(combinedResult.Copied, result.Copied...)
				combinedResult.Skipped = append(combinedResult.Skipped, result.Skipped...)
				combinedResult.Errors = append(combinedResult.Errors, result.Errors...)
			}

			// No AGENTS.md handling — update leaves it untouched

			green := color.New(color.FgGreen)
			yellow := color.New(color.FgYellow)
			red := color.New(color.FgRed)
			bold := color.New(color.Bold)

			// Print results — say "updated" instead of "copied"
			for _, f := range combinedResult.Copied {
				if dryRun {
					yellow.Printf("  would update: %s\n", f)
				} else {
					green.Printf("  updated: %s\n", f)
				}
			}
			for _, e := range combinedResult.Errors {
				red.Fprintf(os.Stderr, "  error: %s\n", e)
			}

			// Summary
			fmt.Println()
			if dryRun {
				bold.Printf("%d would be updated, %d errors\n",
					len(combinedResult.Copied), len(combinedResult.Errors))
			} else {
				bold.Printf("%d updated, %d errors\n",
					len(combinedResult.Copied), len(combinedResult.Errors))
			}

			if len(combinedResult.Errors) > 0 {
				return fmt.Errorf("update completed with %d errors", len(combinedResult.Errors))
			}

			return nil
		},
	}

	cmd.Flags().String("target", ".", "Target directory for update")
	cmd.Flags().String("package", "", "Comma-separated list of packages to update (omit to auto-detect installed packages)")
	cmd.Flags().String("standards", "", "Comma-separated list of language standards to update (omit to auto-detect installed standards)")
	cmd.Flags().String("for", "", "Target coding assistant (copilot, claude, opencode)")
	cmd.Flags().Bool("dry-run", false, "Show what would be updated without writing files")

	return cmd
}

// detectInstalled scans the target directory to find which packages and
// standards are already installed. For packages, it checks for
// skills/<pkg>/SKILL.md (with path mapping applied). For standards, it
// checks for directories under standards/languages/.
func detectInstalled(content fs.FS, target string, pathMapper func(string) string) ([]string, error) {
	var dirs []string

	// Detect installed packages
	available, err := listSubDirs(content, "packages")
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}

	for _, pkg := range available {
		// Each package installs a SKILL.md — use it as the detection marker
		skillPath := fmt.Sprintf("skills/%s/SKILL.md", pkg)
		if pathMapper != nil {
			skillPath = pathMapper(skillPath)
		}
		fullPath := filepath.Join(target, skillPath)
		if _, err := os.Stat(fullPath); err == nil {
			dirs = append(dirs, "packages/"+pkg)
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to check package %q at %s: %w", pkg, fullPath, err)
		}
	}

	// Detect installed standards
	stdBase := "standards/languages"
	if pathMapper != nil {
		stdBase = pathMapper(stdBase)
	}
	stdFullPath := filepath.Join(target, stdBase)

	entries, err := os.ReadDir(stdFullPath)
	if err != nil {
		// If standards/languages/ doesn't exist, that's fine — just no standards installed
		if os.IsNotExist(err) {
			return dirs, nil
		}
		return nil, fmt.Errorf("failed to list installed standards in %s: %w", stdFullPath, err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, "standards/languages/"+entry.Name())
		}
	}

	return dirs, nil
}
