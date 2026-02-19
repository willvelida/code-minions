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

Get started with 'code-minions init' to scaffold a project manifest,
then run 'code-minions install' to install the declared packages.

Available components:
  - agents:    AI agent definitions (.agent.md files)
  - skills:    Step-by-step procedures for common workflows`,
	}

	cmd.AddCommand(newInstallCommand(content))
	cmd.AddCommand(newUninstallCommand(content))
	cmd.AddCommand(newUpdateCommand(content))
	cmd.AddCommand(newListCommand(content))
	cmd.AddCommand(newShowCommand(content))
	cmd.AddCommand(newSearchCommand(content))
	cmd.AddCommand(newTransferCommand(content))
	cmd.AddCommand(newInitCommand(content))
	cmd.AddCommand(newMCPCommand())
	cmd.AddCommand(newVersionCommand())
	cmd.AddCommand(newCompletionCommand())

	cmd.PersistentFlags().Bool("json", false, "Output results as JSON")
	cmd.PersistentFlags().BoolP("verbose", "v", false, "Show detailed output")
	cmd.PersistentFlags().BoolP("quiet", "q", false, "Suppress all output except errors")

	cmd.MarkFlagsMutuallyExclusive("verbose", "quiet")
	cmd.MarkFlagsMutuallyExclusive("verbose", "json")
	cmd.MarkFlagsMutuallyExclusive("quiet", "json")

	return cmd
}
