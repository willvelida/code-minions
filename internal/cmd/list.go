package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/willvelida/code-minions/internal/assistant"
	"github.com/willvelida/code-minions/internal/model"
	"github.com/willvelida/code-minions/internal/registry"
)

func newListCommand(content fs.FS) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available packages",
		Long: `Display all available packages in the built-in registry and the
supported coding assistants. Each package includes its name, version,
and a short description.

Use --detail to also see what each package contains (agents, skills,
actions, standards).

Use this command to discover what packages are available before running
install, or to check which assistants are supported by the --for flag.`,
		Example: `  # List all available packages and assistants
  code-minions list

  # List with content details
  code-minions list --detail`,
		RunE: func(cmd *cobra.Command, args []string) error {
			detail, _ := cmd.Flags().GetBool("detail")

			// Collect package data via the registry
			type packageEntry struct {
				Name        string                 `json:"name"`
				Version     string                 `json:"version,omitempty"`
				Description string                 `json:"description,omitempty"`
				Contents    *model.PackageContents `json:"contents,omitempty"`
			}
			type assistantEntry struct {
				Name        string `json:"name"`
				Description string `json:"description"`
			}

			src := registry.NewEmbeddedSource(content)
			pkgModels, err := src.ListPackages()
			if err != nil {
				return fmt.Errorf("failed to list packages: %w", err)
			}

			var pkgs []packageEntry
			for _, p := range pkgModels {
				entry := packageEntry{
					Name:        p.Name,
					Version:     p.Version,
					Description: p.Description,
				}
				if detail {
					c := p.Contents
					entry.Contents = &c
				}
				pkgs = append(pkgs, entry)
			}

			var assistants []assistantEntry
			for _, name := range assistant.List() {
				cfg, _ := assistant.Get(name)
				assistants = append(assistants, assistantEntry{Name: name, Description: cfg.Description})
			}

			// Output mode
			mode := getOutputMode(cmd)

			if mode == OutputJSON {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(struct {
					Packages   []packageEntry   `json:"packages"`
					Assistants []assistantEntry `json:"assistants"`
				}{Packages: pkgs, Assistants: assistants})
			}

			// --quiet on list is a no-op — warn and print anyway
			quietWarning(cmd, mode)

			// Verbose: show source info
			for _, p := range pkgs {
				verbosePrintf(cmd, mode, "source: embedded, package: %s, version: %s\n", p.Name, p.Version)
			}

			// truncateDesc shortens descriptions for terminal display
			truncateDesc := func(s string, max int) string {
				if len(s) > max {
					return s[:max-3] + "..."
				}
				return s
			}

			// Human-readable output
			w := cmd.OutOrStdout()
			bold := color.New(color.Bold)
			cyan := color.New(color.FgCyan)
			dim := color.New(color.Faint)
			green := color.New(color.FgGreen)

			_, _ = bold.Fprintln(w, "\nPackages")
			_, _ = dim.Fprintln(w, strings.Repeat("-", 60))
			for _, p := range pkgs {
				nameVersion := p.Name
				if p.Version != "" {
					nameVersion += " (" + p.Version + ")"
				}
				_, _ = cyan.Fprintf(w, "  %-40s", nameVersion)
				if p.Description != "" {
					_, _ = dim.Fprintf(w, "  %s", truncateDesc(p.Description, 80))
				}
				_, _ = fmt.Fprintln(w)
				if detail && p.Contents != nil {
					contentsSummary(w, green, dim, p.Contents)
				}
			}

			_, _ = bold.Fprintln(w, "\nAssistants (use with --for)")
			_, _ = dim.Fprintln(w, strings.Repeat("-", 60))
			for _, a := range assistants {
				_, _ = cyan.Fprintf(w, "  %-15s", a.Name)
				_, _ = dim.Fprintf(w, "  %s", a.Description)
				_, _ = fmt.Fprintln(w)
			}

			_, _ = fmt.Fprintln(w)
			return nil
		},
	}

	cmd.Flags().Bool("detail", false, "Show package contents (agents, skills, actions, standards)")

	return cmd
}

// contentsSummary prints a compact content listing under a package.
func contentsSummary(w io.Writer, accent, dim *color.Color, c *model.PackageContents) {
	parts := []struct {
		label string
		count int
	}{
		{"agents", len(c.Agents)},
		{"skills", len(c.Skills)},
		{"actions", len(c.Actions)},
		{"standards", len(c.Standards)},
		{"mcp", len(c.MCP)},
	}

	var nonZero []string
	for _, p := range parts {
		if p.count > 0 {
			nonZero = append(nonZero, fmt.Sprintf("%d %s", p.count, p.label))
		}
	}

	if len(nonZero) > 0 {
		_, _ = dim.Fprint(w, "    ")
		_, _ = accent.Fprintf(w, "→ %s\n", strings.Join(nonZero, ", "))
	}
}
