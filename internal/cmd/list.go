package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/willvelida/code-minions/internal/assistant"
)

func newListCommand(content fs.FS) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available packages",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Collect package data
			type packageEntry struct {
				Name        string `json:"name"`
				Description string `json:"description,omitempty"`
			}
			type assistantEntry struct {
				Name        string `json:"name"`
				Description string `json:"description"`
			}

			var pkgs []packageEntry
			packages, err := fs.ReadDir(content, "packages")
			if err != nil {
				return fmt.Errorf("failed to read packages: %w", err)
			}
			for _, entry := range packages {
				if entry.IsDir() {
					desc := readSkillDescription(content, "packages/"+entry.Name()+"/skills/"+entry.Name())
					pkgs = append(pkgs, packageEntry{Name: entry.Name(), Description: desc})
				}
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

			// Verbose: show scan paths
			for _, p := range pkgs {
				verbosePrintf(cmd, mode, "scanned: packages/%s/skills/%s/SKILL.md\n", p.Name, p.Name)
			}

			// Human-readable output
			bold := color.New(color.Bold)
			cyan := color.New(color.FgCyan)
			dim := color.New(color.Faint)

			_, _ = bold.Println("\nPackages")
			_, _ = dim.Println(strings.Repeat("-", 40))
			for _, p := range pkgs {
				_, _ = cyan.Printf("  %-30s", p.Name)
				if p.Description != "" {
					_, _ = dim.Printf("  %s", p.Description)
				}
				fmt.Println()
			}

			_, _ = bold.Println("\nAssistants (use with --for)")
			_, _ = dim.Println(strings.Repeat("-", 40))
			for _, a := range assistants {
				_, _ = cyan.Printf("  %-15s", a.Name)
				_, _ = dim.Printf("  %s", a.Description)
				fmt.Println()
			}

			fmt.Println()
			return nil
		},
	}
}

// readSkillDescription reads the description from a skill's SKILL.md frontmatter.
func readSkillDescription(content fs.FS, skillDir string) string {
	data, err := fs.ReadFile(content, skillDir+"/SKILL.md")
	if err != nil {
		return ""
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	inFrontmatter := false

	for scanner.Scan() {
		line := scanner.Text()

		if line == "---" {
			if inFrontmatter {
				return "" // End of frontmatter, no description found
			}
			inFrontmatter = true
			continue
		}

		if inFrontmatter && strings.HasPrefix(line, "description:") {
			desc := strings.TrimPrefix(line, "description:")
			desc = strings.TrimSpace(desc)
			desc = strings.Trim(desc, "'\n")
			// Truncate long descriptions
			if len(desc) > 80 {
				desc = desc[:77] + "..."
			}
			return desc
		}
	}

	if err := scanner.Err(); err != nil {
		return ""
	}

	return ""
}
