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

Use --from to list packages from a specific source (a configured source
name or a Git URL).

Use this command to discover what packages are available before running
install, or to check which assistants are supported by the --for flag.`,
		Example: `  # List all available packages and assistants
  code-minions list

  # List with content details
  code-minions list --detail

  # List packages from a specific source
  code-minions list --from my-team

  # List packages from a Git URL
  code-minions list --from https://github.com/org/packages.git`,
		RunE: func(cmd *cobra.Command, args []string) error {
			detail, _ := cmd.Flags().GetBool("detail")
			fromFlag, _ := cmd.Flags().GetString("from")

			// Collect package data via the registry
			type packageEntry struct {
				Name        string                 `json:"name"`
				Version     string                 `json:"version,omitempty"`
				Description string                 `json:"description,omitempty"`
				Source      string                 `json:"source,omitempty"`
				Contents    *model.PackageContents `json:"contents,omitempty"`
			}
			type assistantEntry struct {
				Name        string `json:"name"`
				Description string `json:"description"`
			}

			// Build source(s) based on --from flag
			var reg *registry.Registry
			if fromFlag != "" {
				var err error
				reg, err = registry.BuildRegistryWithFrom(content, fromFlag)
				if err != nil {
					return fmt.Errorf("failed to build registry: %w", err)
				}
			} else {
				// Auto-load all configured sources alongside embedded.
				// Failing sources are warned-and-skipped.
				var err error
				reg, err = registry.BuildRegistryWithWarnings(content, cmd.ErrOrStderr())
				if err != nil {
					return fmt.Errorf("failed to build registry: %w", err)
				}
			}

			pkgModels, err := reg.ListPackages()
			if err != nil {
				return fmt.Errorf("failed to list packages: %w", err)
			}

			var pkgs []packageEntry
			for _, p := range pkgModels {
				entry := packageEntry{
					Name:        p.Name,
					Version:     p.Version,
					Description: p.Description,
					Source:      p.Source,
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

			// Collect persona data — personas are optional, so we
			// don't fail if none exist.
			type personaEntry struct {
				Name        string   `json:"name"`
				Description string   `json:"description,omitempty"`
				Packages    []string `json:"packages,omitempty"`
			}

			var personas []personaEntry
			seenPersonas := make(map[string]bool)
			// Collect personas from all registry sources (first source wins)
			for _, src := range reg.Sources() {
				personaModels, pErr := src.ListPersonas()
				if pErr != nil {
					continue
				}
				for _, p := range personaModels {
					if seenPersonas[p.Name] {
						continue
					}
					seenPersonas[p.Name] = true
					entry := personaEntry{
						Name:        p.Name,
						Description: p.Description,
					}
					if detail {
						for _, ref := range p.Packages {
							entry.Packages = append(entry.Packages, ref.Name)
						}
					}
					personas = append(personas, entry)
				}
			}

			// Output mode
			mode := getOutputMode(cmd)

			if mode == OutputJSON {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(struct {
					Packages   []packageEntry   `json:"packages"`
					Personas   []personaEntry   `json:"personas"`
					Assistants []assistantEntry `json:"assistants"`
				}{Packages: pkgs, Personas: personas, Assistants: assistants})
			}

			// --quiet on list is a no-op — warn and print anyway
			quietWarning(cmd, mode)

			// Verbose: show source info
			for _, p := range pkgs {
				verbosePrintf(cmd, mode, "source: %s, package: %s, version: %s\n", p.Source, p.Name, p.Version)
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

			// Personas section (only shown if any exist)
			if len(personas) > 0 {
				_, _ = bold.Fprintln(w, "\nPersonas (use with --persona)")
				_, _ = dim.Fprintln(w, strings.Repeat("-", 60))
				for _, p := range personas {
					_, _ = cyan.Fprintf(w, "  %-30s", p.Name)
					if p.Description != "" {
						_, _ = dim.Fprintf(w, "  %s", p.Description)
					}
					_, _ = fmt.Fprintln(w)
					if detail && len(p.Packages) > 0 {
						_, _ = dim.Fprint(w, "    ")
						_, _ = green.Fprintf(w, "→ packages: %s\n", strings.Join(p.Packages, ", "))
					}
				}
			}

			_, _ = fmt.Fprintln(w)
			return nil
		},
	}

	cmd.Flags().Bool("detail", false, "Show package contents (agents, skills, actions, standards)")
	cmd.Flags().String("from", "", "List packages from a specific source (name or Git URL)")

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
		{"instructions", len(c.Instructions)},
		{"prompts", len(c.Prompts)},
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
