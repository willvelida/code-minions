package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/willvelida/code-minions/internal/registry"
)

func newSourceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "source",
		Short: "Manage external package sources",
		Long: `Manage external package sources. Sources are Git repositories
that contain code-minions packages in the standard packages/ directory
layout.

Sources are stored in a global config file and are available across
all projects on this machine.`,
	}

	cmd.AddCommand(newSourceAddCommand())
	cmd.AddCommand(newSourceListCommand())
	cmd.AddCommand(newSourceRemoveCommand())

	return cmd
}

func newSourceAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <name> <url>",
		Short: "Add a package source",
		Long: `Add a Git repository as a package source. The repository must
contain packages in the standard code-minions packages/ directory layout.

Authentication uses your existing Git credentials (SSH keys, credential
helpers, GITHUB_TOKEN, etc). No additional auth configuration is needed.`,
		Example: `  # Add a team's package repository
  code-minions source add my-team https://github.com/my-org/coding-agents.git

  # Add a community skills repository
  code-minions source add community https://github.com/community/code-minions-packages.git`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			url := args[1]
			typeFlag, _ := cmd.Flags().GetString("type")

			mode := getOutputMode(cmd)

			cfg, err := registry.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			src := registry.SourceConfig{
				Name: name,
				Type: typeFlag,
				URL:  url,
			}

			if err := cfg.AddSource(src); err != nil {
				return err
			}

			if err := registry.SaveConfig(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			if mode == OutputJSON {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(struct {
					Name string `json:"name"`
					Type string `json:"type"`
					URL  string `json:"url"`
				}{Name: name, Type: typeFlag, URL: url})
			}

			if mode != OutputQuiet {
				green := color.New(color.FgGreen)
				_, _ = green.Fprintf(cmd.OutOrStdout(), "Added source %q (%s)\n", name, url)
			}
			return nil
		},
	}

	cmd.Flags().String("type", "git", "Source type (git)")

	return cmd
}

func newSourceListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List configured package sources",
		Long: `Display all configured package sources, including the built-in
embedded source which is always available.`,
		Example: `  # List all sources
  code-minions source list

  # List as JSON
  code-minions source list --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			mode := getOutputMode(cmd)

			cfg, err := registry.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			type sourceEntry struct {
				Name    string `json:"name"`
				Type    string `json:"type"`
				URL     string `json:"url,omitempty"`
				BuiltIn bool   `json:"built_in,omitempty"`
			}

			var entries []sourceEntry

			// Always include embedded source
			entries = append(entries, sourceEntry{
				Name:    "embedded",
				Type:    "embedded",
				BuiltIn: true,
			})

			for _, src := range cfg.Sources {
				entries = append(entries, sourceEntry{
					Name: src.Name,
					Type: src.Type,
					URL:  src.URL,
				})
			}

			if mode == OutputJSON {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(struct {
					Sources []sourceEntry `json:"sources"`
					Count   int           `json:"count"`
				}{
					Sources: entries,
					Count:   len(entries),
				})
			}

			quietWarning(cmd, mode)

			w := cmd.OutOrStdout()
			bold := color.New(color.Bold)
			cyan := color.New(color.FgCyan)
			dim := color.New(color.Faint)

			_, _ = bold.Fprintln(w, "\nPackage Sources")
			_, _ = dim.Fprintln(w, strings.Repeat("-", 60))

			for _, entry := range entries {
				if entry.BuiltIn {
					_, _ = cyan.Fprintf(w, "  %-20s", entry.Name)
					_, _ = dim.Fprintln(w, "(built-in)")
				} else {
					_, _ = cyan.Fprintf(w, "  %-20s", entry.Name)
					_, _ = fmt.Fprintf(w, "%s ", entry.Type)
					_, _ = dim.Fprintln(w, entry.URL)
				}
			}

			_, _ = fmt.Fprintln(w)
			return nil
		},
	}
}

func newSourceRemoveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a package source",
		Long:  `Remove a configured package source by name. The built-in embedded source cannot be removed.`,
		Example: `  # Remove a source
  code-minions source remove my-team`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			mode := getOutputMode(cmd)

			if name == "embedded" {
				return fmt.Errorf("cannot remove the built-in embedded source")
			}

			cfg, err := registry.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			if err := cfg.RemoveSource(name); err != nil {
				return fmt.Errorf("source %q not found. Run 'code-minions source list' to see configured sources", name)
			}

			if err := registry.SaveConfig(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			if mode == OutputJSON {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(struct {
					Removed string `json:"removed"`
				}{Removed: name})
			}

			if mode != OutputQuiet {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Removed source %q\n", name)
			}
			return nil
		},
	}
}
