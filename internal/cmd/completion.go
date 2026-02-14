package cmd

import (
	"github.com/spf13/cobra"
)

// newCompletionCommand returns the parent "completion" command with
// subcommands for each supported shell.
func newCompletionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [shell]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for code-minions.

To load completions:

  bash:
    source <(code-minions completion bash)

    # To load completions for each session, add to your profile:
    # Linux:
    echo 'source <(code-minions completion bash)' >> ~/.bashrc
    # macOS:
    echo 'source <(code-minions completion bash)' >> ~/.bash_profile

  zsh:
    # If shell completion is not already enabled, enable it:
    echo "autoload -U compinit; compinit" >> ~/.zshrc

    # Load completions for each session:
    echo 'source <(code-minions completion zsh)' >> ~/.zshrc

  fish:
    code-minions completion fish | source

    # To load completions for each session:
    code-minions completion fish > ~/.config/fish/completions/code-minions.fish

  PowerShell:
    code-minions completion powershell | Out-String | Invoke-Expression

    # To load completions for each session, add to your profile:
    echo 'code-minions completion powershell | Out-String | Invoke-Expression' >> $PROFILE
`,
	}

	cmd.AddCommand(newBashCompletionCmd())
	cmd.AddCommand(newZshCompletionCmd())
	cmd.AddCommand(newFishCompletionCmd())
	cmd.AddCommand(newPowershellCompletionCmd())

	return cmd
}

func newBashCompletionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "bash",
		Short: "Generate bash completion script",
		Long: `Generate the autocompletion script for bash.

To load completions in your current shell session:

  source <(code-minions completion bash)

To load completions for every new session, add to your profile:

  # Linux:
  echo 'source <(code-minions completion bash)' >> ~/.bashrc

  # macOS:
  echo 'source <(code-minions completion bash)' >> ~/.bash_profile
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenBashCompletionV2(cmd.OutOrStdout(), true)
		},
	}
}

func newZshCompletionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "zsh",
		Short: "Generate zsh completion script",
		Long: `Generate the autocompletion script for zsh.

If shell completion is not already enabled in your environment,
you will need to enable it:

  echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions in your current shell session:

  source <(code-minions completion zsh)

To load completions for every new session, add to your profile:

  echo 'source <(code-minions completion zsh)' >> ~/.zshrc
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenZshCompletion(cmd.OutOrStdout())
		},
	}
}

func newFishCompletionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "fish",
		Short: "Generate fish completion script",
		Long: `Generate the autocompletion script for fish.

To load completions in your current shell session:

  code-minions completion fish | source

To load completions for every new session:

  code-minions completion fish > ~/.config/fish/completions/code-minions.fish
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenFishCompletion(cmd.OutOrStdout(), true)
		},
	}
}

func newPowershellCompletionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "powershell",
		Short: "Generate PowerShell completion script",
		Long: `Generate the autocompletion script for PowerShell.

To load completions in your current shell session:

  code-minions completion powershell | Out-String | Invoke-Expression

To load completions for every new session, add to your PowerShell profile:

  echo 'code-minions completion powershell | Out-String | Invoke-Expression' >> $PROFILE
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
		},
	}
}
