package cmd

import (
	"bufio"
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
			bold := color.New(color.Bold)
			cyan := color.New(color.FgCyan)
			dim := color.New(color.Faint)

			// List packages
			bold.Println("\nPackages")
			dim.Println(strings.Repeat("-", 40))
			packages, err := fs.ReadDir(content, "packages")
			if err != nil {
				return fmt.Errorf("failed to read packages: %w", err)
			}
			for _, entry := range packages {
				if entry.IsDir() {
					desc := readSkillDescription(content, "packages/"+entry.Name()+"/skills/"+entry.Name())
					cyan.Printf("  %-30s", entry.Name())
					if desc != "" {
						dim.Printf("  %s", desc)
					}
					fmt.Println()
				}
			}

			// List assistants (for --for flag)
			bold.Println("\nAssistants (use with --for)")
			dim.Println(strings.Repeat("-", 40))
			for _, name := range assistant.List() {
				cfg, _ := assistant.Get(name)
				cyan.Printf("  %-15s", name)
				dim.Printf("  %s", cfg.Description)
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
