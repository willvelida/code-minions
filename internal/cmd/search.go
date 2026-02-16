package cmd

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/willvelida/code-minions/internal/model"
	"github.com/willvelida/code-minions/internal/registry"
)

func newSearchCommand(content fs.FS) *cobra.Command {
	return &cobra.Command{
		Use:   "search <query>",
		Short: "Search for packages by name or description",
		Long: `Search performs a case-insensitive text search across all package
names and descriptions. Results show the package name, version, and
a matching description excerpt.

Use 'code-minions show <package>' to see full details for a result.`,
		Example: `  # Search for packages related to git
  code-minions search git

  # Search for documentation-related packages
  code-minions search documentation

  # Search with JSON output
  code-minions search agent --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]

			src := registry.NewEmbeddedSource(content)
			results, err := src.Search(query)
			if err != nil {
				return fmt.Errorf("search failed: %w", err)
			}

			mode := getOutputMode(cmd)

			if mode == OutputJSON {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(struct {
					Query   string               `json:"query"`
					Results []searchResultOutput `json:"results"`
					Count   int                  `json:"count"`
				}{
					Query:   query,
					Results: toSearchOutput(results),
					Count:   len(results),
				})
			}

			quietWarning(cmd, mode)

			if len(results) == 0 {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "No results found for %q\n", query)
				return nil
			}

			w := cmd.OutOrStdout()

			bold := color.New(color.Bold)
			cyan := color.New(color.FgCyan)
			dim := color.New(color.Faint)

			_, _ = bold.Fprintf(w, "\nSearch results for %q (%d found)\n", query, len(results))
			_, _ = dim.Fprintln(w, strings.Repeat("-", 60))

			for _, r := range results {
				nameVersion := r.Name
				if r.Version != "" {
					nameVersion += " (" + r.Version + ")"
				}
				_, _ = cyan.Fprintf(w, "  %-40s", nameVersion)
				if r.Description != "" {
					desc := r.Description
					if len(desc) > 80 {
						desc = desc[:77] + "..."
					}
					_, _ = dim.Fprintf(w, "  %s", desc)
				}
				_, _ = fmt.Fprintln(w)

				verbosePrintf(cmd, mode, "    kind: %s, source: %s\n", r.Kind, r.Source)
			}

			_, _ = fmt.Fprintln(w)
			return nil
		},
	}
}

// searchResultOutput is the JSON-serialisable search result.
type searchResultOutput struct {
	Kind        string `json:"kind"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Source      string `json:"source"`
}

func toSearchOutput(results []model.SearchResult) []searchResultOutput {
	if len(results) == 0 {
		return []searchResultOutput{}
	}
	out := make([]searchResultOutput, len(results))
	for i, r := range results {
		out[i] = searchResultOutput{
			Kind:        r.Kind,
			Name:        r.Name,
			Description: r.Description,
			Version:     r.Version,
			Source:      r.Source,
		}
	}
	return out
}
