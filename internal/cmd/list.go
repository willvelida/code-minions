package cmd

import (
	"bufio"
	"fmt"
	"io/fs"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newListCommand(content fs.FS) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available agents, skills, and standards",
		RunE: func(cmd *cobra.Command, args []string) error {
			bold := color.New(color.Bold)
			cyan := color.New(color.FgCyan)
			dim := color.New(color.Faint)

			// List agents
			bold.Println("\nAgents")
			dim.Println(strings.Repeat("-", 40))
			agents, err := fs.ReadDir(content, "agents")
			if err != nil {
				return fmt.Errorf("failed to read agents: %w", err)
			}
			for _, entry := range agents {
				if !entry.IsDir() {
					cyan.Printf("  %s\n", entry.Name())
				}
			}

			// List skills
			bold.Println("\nSkills")
			dim.Println(strings.Repeat("-", 40))
			skills, err := fs.ReadDir(content, "skills")
			if err != nil {
				return fmt.Errorf("failed to read skills: %w", err)
			}
			for _, entry := range skills {
				if entry.IsDir() {
					desc := readSkillDescription(content, "skills/"+entry.Name())
					cyan.Printf("  %-30s", entry.Name())
					if desc != "" {
						dim.Printf("  %s", desc)
					}
					fmt.Println()
				}
			}

			// List standards
			bold.Println("\nStandards")
			dim.Println(strings.Repeat("-", 40))
			standards, err := fs.ReadDir(content, "standards/languages")
			if err != nil {
				return fmt.Errorf("failed to read standards: %w", err)
			}
			for _, entry := range standards {
				if entry.IsDir() {
					cyan.Printf("  %s\n", entry.Name())
				}
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

	return ""
}
