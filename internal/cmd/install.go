package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/willvelida/code-minions/internal/installer"
)

func newInstallCommand(content fs.FS) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install agents, skills, and standards into your repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			force, _ := cmd.Flags().GetBool("force")
			packageFlag, _ := cmd.Flags().GetString("package")
			standardsFlag, _ := cmd.Flags().GetString("standards")

			// Build the list of directories to install
			dirs, err := buildDirList(content, packageFlag, standardsFlag)
			if err != nil {
				return err
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

			// Install each package (with prefix stripping)
			combinedResult := &installer.Result{}
			for _, pkgDir := range packageDirs {
				inst := &installer.Installer{
					Content:     content,
					Target:      target,
					Force:       force,
					DryRun:      dryRun,
					StripPrefix: pkgDir,
				}
				result, err := inst.Install([]string{pkgDir})
				if err != nil {
					return fmt.Errorf("installation failed: %w", err)
				}
				combinedResult.Copied = append(combinedResult.Copied, result.Copied...)
				combinedResult.Skipped = append(combinedResult.Skipped, result.Skipped...)
				combinedResult.Errors = append(combinedResult.Errors, result.Errors...)
			}

			// Install standards
			if len(standardDirs) > 0 {
				inst := &installer.Installer{
					Content: content,
					Target:  target,
					Force:   force,
					DryRun:  dryRun,
				}
				result, err := inst.Install(standardDirs)
				if err != nil {
					return fmt.Errorf("installation failed: %w", err)
				}
				combinedResult.Copied = append(combinedResult.Copied, result.Copied...)
				combinedResult.Skipped = append(combinedResult.Skipped, result.Skipped...)
				combinedResult.Errors = append(combinedResult.Errors, result.Errors...)
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
	cmd.Flags().String("standards", "", "Comma-separated list of language standards to install")
	cmd.Flags().Bool("dry-run", false, "Show what would be installed without writing files")
	cmd.Flags().Bool("force", false, "Overwrite existing files")

	return cmd
}

func buildDirList(content fs.FS, packageFlag string, standardsFlag string) ([]string, error) {
	// No flags = install all packages + all standards
	if packageFlag == "" && standardsFlag == "" {
		var dirs []string
		// Add all packages
		packages, err := listSubDirs(content, "packages")
		if err != nil {
			return nil, fmt.Errorf("failed to list packages: %w", err)
		}
		for _, pkg := range packages {
			dirs = append(dirs, "packages/"+pkg)
		}
		// Add standards
		dirs = append(dirs, "standards")
		return dirs, nil
	}

	var dirs []string

	// Package filtering
	if packageFlag != "" {
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
	}

	// Standards filtering
	if standardsFlag != "" {
		standardsAdded := 0
		standards := strings.Split(standardsFlag, ",")
		for _, std := range standards {
			std = strings.TrimSpace(std)
			if std == "" {
				continue
			}
			dirPath := "standards/languages/" + std

			// Validate standard exists in embedded FS
			if _, err := fs.Stat(content, dirPath); err != nil {
				available, err := listSubDirs(content, "standards/languages")
				if err != nil {
					return nil, fmt.Errorf("standard %q not found", std)
				}
				return nil, fmt.Errorf("standard %q not found\n\nAvailable standards:\n  %s",
					std, strings.Join(available, "\n  "))
			}

			dirs = append(dirs, dirPath)
			standardsAdded++
		}

		// Always include the standards index when installing standards
		if standardsAdded > 0 {
			dirs = append(dirs, "standards/languages/standards.index.md")
		}

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
