package cmd

import (
	"io/fs"

	"github.com/spf13/cobra"
)

func NewRootCommand(content fs.FS) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "code-minions",
		Short: "Install code-minions agents, skills, and standards into your repository",
		Long: `code-minions is a CLI tool that installs reusable AI agent skills,
standards, and configurations into your project.

Available components:
  - agents:    AI agent definitions (.agent.md files)
  - skills:    Step-by-step procedures for common workflows
  - standards: Language-specific development guidelines`,
	}

	cmd.AddCommand(newInstallCommand(content))
	cmd.AddCommand(newUninstallCommand(content))
	cmd.AddCommand(newListCommand(content))
	cmd.AddCommand(newVersionCommand())

	return cmd
}
