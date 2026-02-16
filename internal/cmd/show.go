package cmd

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/willvelida/code-minions/internal/registry"
)

func newShowCommand(content fs.FS) *cobra.Command {
	return &cobra.Command{
		Use:   "show <package>",
		Short: "Show detailed information about a package",
		Long: `Display detailed metadata for a specific package, including its full
description, version, author, license, compatibility, contents, and
dependencies.

Use 'code-minions list' to see all available package names.`,
		Example: `  # Show details for the git-workflow package
  code-minions show git-workflow

  # Show details as JSON
  code-minions show git-workflow --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			src := registry.NewEmbeddedSource(content)
			pkg, err := src.GetPackage(name)
			if err != nil {
				// List available packages in the error message
				available, _ := src.ListPackages()
				var names []string
				for _, p := range available {
					names = append(names, p.Name)
				}
				return fmt.Errorf("package %q not found\n\nAvailable packages:\n  %s",
					name, strings.Join(names, "\n  "))
			}

			mode := getOutputMode(cmd)

			if mode == OutputJSON {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(pkg)
			}

			quietWarning(cmd, mode)

			w := cmd.OutOrStdout()

			bold := color.New(color.Bold)
			cyan := color.New(color.FgCyan)
			dim := color.New(color.Faint)
			green := color.New(color.FgGreen)

			_, _ = bold.Fprintf(w, "\n%s", pkg.Name)
			if pkg.Version != "" {
				_, _ = dim.Fprintf(w, " (%s)", pkg.Version)
			}
			_, _ = fmt.Fprintln(w)
			_, _ = dim.Fprintln(w, strings.Repeat("-", 60))

			if pkg.Description != "" {
				_, _ = fmt.Fprintf(w, "  %s\n\n", pkg.Description)
			}

			// Metadata
			if pkg.Author != "" {
				_, _ = dim.Fprint(w, "  Author:        ")
				_, _ = fmt.Fprintln(w, pkg.Author)
			}
			if pkg.License != "" {
				_, _ = dim.Fprint(w, "  License:       ")
				_, _ = fmt.Fprintln(w, pkg.License)
			}
			if len(pkg.Compatibility) > 0 {
				_, _ = dim.Fprint(w, "  Compatibility: ")
				_, _ = fmt.Fprintln(w, strings.Join(pkg.Compatibility, ", "))
			}

			// Contents
			showContents := func(label string, items []string) {
				if len(items) == 0 {
					return
				}
				_, _ = fmt.Fprintln(w)
				_, _ = cyan.Fprintf(w, "  %s (%d)\n", label, len(items))
				for _, item := range items {
					_, _ = green.Fprintf(w, "    %s\n", item)
				}
			}

			showContents("Agents", pkg.Contents.Agents)
			showContents("Skills", pkg.Contents.Skills)
			showContents("Actions", pkg.Contents.Actions)
			showContents("Standards", pkg.Contents.Standards)
			showContents("MCP", pkg.Contents.MCP)

			// Dependencies
			if len(pkg.Dependencies) > 0 {
				_, _ = fmt.Fprintln(w)
				_, _ = cyan.Fprintln(w, "  Dependencies")
				for _, dep := range pkg.Dependencies {
					depStr := dep.Name
					if dep.Version != "" {
						depStr += " " + dep.Version
					}
					if dep.Source != "" {
						depStr += " (from " + dep.Source + ")"
					}
					_, _ = dim.Fprintf(w, "    %s\n", depStr)
				}
			}

			// Verbose: show source info
			verbosePrintf(cmd, mode, "\nsource: embedded\n")
			verbosePrintf(cmd, mode, "manifest: packages/%s/package.yaml\n", name)

			_, _ = fmt.Fprintln(w)
			return nil
		},
	}
}
