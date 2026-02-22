package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/willvelida/code-minions/internal/installer"
	"github.com/willvelida/code-minions/internal/model"
	"github.com/willvelida/code-minions/internal/registry"
)

func newSearchCommand(content fs.FS) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search for packages by name, description, or tag",
		Long: `Search performs a case-insensitive text search across all package
names, descriptions, and tags. Results show the package name, version,
a matching description excerpt, and tags.

Use --tag to filter results by a specific tag.
Use --info to display detailed metadata for a specific package.
Use --from to search packages from a specific source (a configured
source name or a Git URL).

Use 'code-minions show <package>' to see full details for a result.`,
		Example: `  # Search for packages related to git
  code-minions search git

  # Search for documentation-related packages
  code-minions search documentation

  # Search with JSON output
  code-minions search agent --json

  # Search a specific source
  code-minions search mentor --from my-team

  # Filter by tag
  code-minions search --tag security

  # Show detailed info for a package
  code-minions search --info threat-modelling`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fromFlag, _ := cmd.Flags().GetString("from")
			tagFlag, _ := cmd.Flags().GetString("tag")
			infoFlag, _ := cmd.Flags().GetString("info")

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

			mode := getOutputMode(cmd)

			// --info: show detailed package info (reuses show logic)
			if infoFlag != "" {
				return runSearchInfo(cmd, reg, infoFlag, mode)
			}

			// Require either a positional query or --tag
			query := ""
			if len(args) > 0 {
				query = args[0]
			}
			if query == "" && tagFlag == "" {
				return fmt.Errorf("provide a search query or use --tag to filter by tag")
			}

			results, err := reg.Search(query)
			if err != nil {
				return fmt.Errorf("search failed: %w", err)
			}

			// Filter by tag when --tag is provided
			if tagFlag != "" {
				results = filterByTag(results, tagFlag)
			}

			// Check installed status
			results = checkInstalled(results)

			if mode == OutputJSON {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(struct {
					Query   string               `json:"query"`
					Tag     string               `json:"tag,omitempty"`
					Results []searchResultOutput `json:"results"`
					Count   int                  `json:"count"`
				}{
					Query:   query,
					Tag:     tagFlag,
					Results: toSearchOutput(results),
					Count:   len(results),
				})
			}

			quietWarning(cmd, mode)

			if len(results) == 0 {
				msg := fmt.Sprintf("No results found for %q", query)
				if tagFlag != "" {
					msg = fmt.Sprintf("No results found for tag %q", tagFlag)
					if query != "" {
						msg = fmt.Sprintf("No results found for %q with tag %q", query, tagFlag)
					}
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), msg)
				return nil
			}

			w := cmd.OutOrStdout()

			bold := color.New(color.Bold)
			cyan := color.New(color.FgCyan)
			dim := color.New(color.Faint)
			green := color.New(color.FgGreen)

			header := fmt.Sprintf("Search results for %q (%d found)", query, len(results))
			if query == "" && tagFlag != "" {
				header = fmt.Sprintf("Packages tagged %q (%d found)", tagFlag, len(results))
			} else if tagFlag != "" {
				header = fmt.Sprintf("Search results for %q tagged %q (%d found)", query, tagFlag, len(results))
			}
			_, _ = bold.Fprintf(w, "\n%s\n", header)
			_, _ = dim.Fprintln(w, strings.Repeat("-", 60))

			for _, r := range results {
				nameVersion := r.Name
				if r.Version != "" {
					nameVersion += " (" + r.Version + ")"
				}
				_, _ = cyan.Fprintf(w, "  %-40s", nameVersion)
				if r.Description != "" {
					desc := r.Description
					if len(desc) > 60 {
						desc = desc[:57] + "..."
					}
					_, _ = dim.Fprintf(w, "  %s", desc)
				}
				if len(r.Tags) > 0 {
					_, _ = dim.Fprintf(w, "  [%s]", strings.Join(r.Tags, ", "))
				}
				if r.Installed {
					_, _ = green.Fprint(w, " ✓")
				}
				_, _ = fmt.Fprintln(w)

				verbosePrintf(cmd, mode, "    kind: %s, source: %s\n", r.Kind, r.Source)
			}

			_, _ = fmt.Fprintln(w)
			return nil
		},
	}

	cmd.Flags().String("from", "", "Search packages from a specific source (name or Git URL)")
	cmd.Flags().String("tag", "", "Filter results by tag")
	cmd.Flags().String("info", "", "Show detailed information for a specific package")

	return cmd
}

// runSearchInfo handles the --info flag by resolving and displaying a package.
func runSearchInfo(cmd *cobra.Command, reg *registry.Registry, name string, mode OutputMode) error {
	pkg, src, err := reg.ResolvePackage(name)
	if err != nil {
		if !errors.Is(err, registry.ErrNotFound) {
			return err
		}
		return fmt.Errorf("package %q not found", name)
	}
	pkg.Source = src.Name()

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
	if len(pkg.Tags) > 0 {
		_, _ = dim.Fprint(w, "  Tags:          ")
		_, _ = fmt.Fprintln(w, strings.Join(pkg.Tags, ", "))
	}
	if len(pkg.Compatibility) > 0 {
		_, _ = dim.Fprint(w, "  Compatibility: ")
		_, _ = fmt.Fprintln(w, strings.Join(pkg.Compatibility, ", "))
	}
	_, _ = dim.Fprint(w, "  Source:         ")
	_, _ = fmt.Fprintln(w, pkg.Source)

	// Installed status
	installed := isPackageInstalled(pkg.Name)
	_, _ = dim.Fprint(w, "  Installed:      ")
	if installed {
		_, _ = green.Fprintln(w, "yes")
	} else {
		_, _ = fmt.Fprintln(w, "no")
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
	showContents("Instructions", pkg.Contents.Instructions)
	showContents("Prompts", pkg.Contents.Prompts)
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

	verbosePrintf(cmd, mode, "\nsource: %s\n", pkg.Source)

	_, _ = fmt.Fprintln(w)
	return nil
}

// filterByTag returns only results that have a matching tag (case-insensitive).
func filterByTag(results []model.SearchResult, tag string) []model.SearchResult {
	var filtered []model.SearchResult
	for _, r := range results {
		for _, t := range r.Tags {
			if strings.EqualFold(t, tag) {
				filtered = append(filtered, r)
				break
			}
		}
	}
	return filtered
}

// checkInstalled marks each result with whether its package appears to be
// installed in the current working directory by consulting the install manifest.
func checkInstalled(results []model.SearchResult) []model.SearchResult {
	manifest, err := loadManifest(".")
	if err != nil {
		return results
	}
	for i, r := range results {
		if r.Kind == "package" {
			results[i].Installed = findInstalled(manifest, r.Name) != nil
		}
	}
	return results
}

// isPackageInstalled returns true if the named package is recorded in the
// install manifest (.code-minions/installed.json). This works regardless of
// which coding assistant layout was used during installation.
var isPackageInstalled = func(name string) bool {
	manifest, err := loadManifest(".")
	if err != nil {
		return false
	}
	return findInstalled(manifest, name) != nil
}

// loadManifest and findInstalled are thin wrappers so tests can replace them.
var loadManifest = installer.LoadManifest
var findInstalled = installer.FindInstalled

// searchResultOutput is the JSON-serialisable search result.
type searchResultOutput struct {
	Kind        string   `json:"kind"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Source      string   `json:"source"`
	Tags        []string `json:"tags"`
	Installed   bool     `json:"installed"`
}

func toSearchOutput(results []model.SearchResult) []searchResultOutput {
	if len(results) == 0 {
		return []searchResultOutput{}
	}
	out := make([]searchResultOutput, len(results))
	for i, r := range results {
		tags := r.Tags
		if tags == nil {
			tags = []string{}
		}
		out[i] = searchResultOutput{
			Kind:        r.Kind,
			Name:        r.Name,
			Description: r.Description,
			Version:     r.Version,
			Source:      r.Source,
			Tags:        tags,
			Installed:   r.Installed,
		}
	}
	return out
}
