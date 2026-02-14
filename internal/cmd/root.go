package cmd

import (
	"io/fs"

	"github.com/spf13/cobra"
)

func NewRootCommand(content fs.FS) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "code-minions",
		Short: "Install code-minions agents and skills into your repository",
		Long: `code-minions is a CLI tool that installs reusable AI agent skills
and configurations into your project.

Available components:
  - agents:    AI agent definitions (.agent.md files)
  - skills:    Step-by-step procedures for common workflows`,
	}

	cmd.AddCommand(newInstallCommand(content))
	cmd.AddCommand(newUninstallCommand(content))
	cmd.AddCommand(newUpdateCommand(content))
	cmd.AddCommand(newListCommand(content))
	cmd.AddCommand(newVersionCommand())
	cmd.AddCommand(newCompletionCommand())

	cmd.PersistentFlags().Bool("json", false, "Output results as JSON")

	return cmd
}
