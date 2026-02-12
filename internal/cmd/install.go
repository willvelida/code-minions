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
			includeAgents, _ := cmd.Flags().GetBool("agents")
			skillsFlag, _ := cmd.Flags().GetString("skills")
			standardsFlag, _ := cmd.Flags().GetString("standards")

			// Build the list of directories to install
			dirs, err := buildDirList(content, includeAgents, skillsFlag, standardsFlag)
			if err != nil {
				return err
			}

			inst := &installer.Installer{
				Content: content,
				Target:  target,
				Force:   force,
				DryRun:  dryRun,
			}

			if dryRun {
				color.New(color.FgYellow, color.Bold).Println("Dry run - no files will be written")
				fmt.Println()
			}

			result, err := inst.Install(dirs)
			if err != nil {
				return fmt.Errorf("installation failed: %w", err)
			}

			green := color.New(color.FgGreen)
			yellow := color.New(color.FgYellow)
			red := color.New(color.FgRed)
			bold := color.New(color.Bold)

			// Print results
			for _, f := range result.Copied {
				if dryRun {
					yellow.Printf("  would copy: %s\n", f)
				} else {
					green.Printf("  copied: %s\n", f)
				}
			}
			for _, f := range result.Skipped {
				yellow.Printf("  skipped (exists): %s\n", f)
			}
			for _, e := range result.Errors {
				red.Fprintf(os.Stderr, "  error: %s\n", e)
			}

			// Summary
			fmt.Println()
			bold.Printf("%d copied, %d skipped, %d errors\n",
				len(result.Copied), len(result.Skipped), len(result.Errors))

			if len(result.Errors) > 0 {
				return fmt.Errorf("installation completed with %d errors", len(result.Errors))
			}

			return nil
		},
	}

	cmd.Flags().String("target", ".", "Target directory for installation")
	cmd.Flags().String("skills", "", "Comma-separated list of skills to install (default: all)")
	cmd.Flags().String("standards", "", "Comma-separated list of language standards to install (default: all)")
	cmd.Flags().Bool("agents", false, "Include agents in the installation")
	cmd.Flags().Bool("dry-run", false, "Show what would be installed without writing files")
	cmd.Flags().Bool("force", false, "Overwrite existing files")

	return cmd
}

func buildDirList(content fs.FS, includeAgents bool, skillsFlag string, standardsFlag string) ([]string, error) {
	// No flags = install everything
	if !includeAgents && skillsFlag == "" && standardsFlag == "" {
		return []string{"agents", "skills", "standards"}, nil
	}

	var dirs []string

	// Agents
	if includeAgents {
		dirs = append(dirs, "agents")
	}

	// Skills filtering
	if skillsFlag != "" {
		skills := strings.Split(skillsFlag, ",")
		for _, skill := range skills {
			skill = strings.TrimSpace(skill)
			if skill == "" {
				continue
			}
			dirPath := "skills/" + skill

			// Validate skill exists in embedded FS
			if _, err := fs.Stat(content, dirPath); err != nil {
				available, _ := listSubDirs(content, "skills")
				return nil, fmt.Errorf("skill %q not found\n\nAvailable skills:\n  %s",
					skill, strings.Join(available, "\n  "))
			}

			dirs = append(dirs, dirPath)
		}
	}

	// Standards filtering
	if standardsFlag != "" {
		standards := strings.Split(standardsFlag, ",")
		for _, std := range standards {
			std = strings.TrimSpace(std)
			if std == "" {
				continue
			}
			dirPath := "standards/languages/" + std

			// Validate standard exists in embedded FS
			if _, err := fs.Stat(content, dirPath); err != nil {
				available, _ := listSubDirs(content, "standards/languages")
				return nil, fmt.Errorf("standard %q not found\n\nAvailable standards:\n  %s",
					std, strings.Join(available, "\n  "))
			}

			dirs = append(dirs, dirPath)
		}

		// Always include the standards index when installing standards
		dirs = append(dirs, "standards/languages/standards.index.md")
	}

	return dirs, nil
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
