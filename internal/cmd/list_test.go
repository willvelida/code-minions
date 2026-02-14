package cmd

import (
	"bytes"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/fatih/color"
)

// testContentFSWithDescriptions returns a test filesystem where SKILL.md files
// include YAML frontmatter with description fields. This is separate from the
// shared testContentFS() so we don't break other tests that rely on that fixture.
func testContentFSWithDescriptions() fstest.MapFS {
	return fstest.MapFS{
		"packages/git-workflow/agents/git-workflow.agent.md": &fstest.MapFile{
			Data: []byte("# Git Agent"),
		},
		"packages/git-workflow/skills/git-workflow/SKILL.md": &fstest.MapFile{
			Data: []byte("---\nname: git-workflow\ndescription: 'Conventional commit workflows'\n---\n# Git"),
		},
		"packages/developer-mentor/agents/developer-mentor.agent.md": &fstest.MapFile{
			Data: []byte("# Mentor Agent"),
		},
		"packages/developer-mentor/skills/developer-mentor/SKILL.md": &fstest.MapFile{
			Data: []byte("---\nname: developer-mentor\ndescription: 'Socratic development mentoring'\n---\n# Mentor"),
		},
	}
}

// ---------- readSkillDescription unit tests ----------

func TestReadSkillDescription(t *testing.T) {
	tests := []struct {
		name     string
		files    fstest.MapFS
		skillDir string
		want     string
	}{
		{
			name: "returns description from frontmatter",
			files: fstest.MapFS{
				"packages/test-skill/skills/test-skill/SKILL.md": &fstest.MapFile{
					Data: []byte("---\nname: test\ndescription: 'A test skill'\n---\n# Test"),
				},
			},
			skillDir: "packages/test-skill/skills/test-skill",
			want:     "A test skill",
		},
		{
			name: "returns empty string when no frontmatter",
			files: fstest.MapFS{
				"packages/test-skill/skills/test-skill/SKILL.md": &fstest.MapFile{
					Data: []byte("# Just a heading\nSome content"),
				},
			},
			skillDir: "packages/test-skill/skills/test-skill",
			want:     "",
		},
		{
			name: "truncates long descriptions",
			files: fstest.MapFS{
				"packages/test-skill/skills/test-skill/SKILL.md": &fstest.MapFile{
					Data: []byte("---\ndescription: 'This is a very long description that exceeds eighty characters and should be truncated by the function'\n---"),
				},
			},
			skillDir: "packages/test-skill/skills/test-skill",
			want:     "This is a very long description that exceeds eighty characters and should be ...",
		},
		{
			name:     "returns empty string when file does not exist",
			files:    fstest.MapFS{},
			skillDir: "packages/nonexistent/skills/nonexistent",
			want:     "",
		},
		{
			name: "returns empty string when frontmatter has no description field",
			files: fstest.MapFS{
				"packages/test-skill/skills/test-skill/SKILL.md": &fstest.MapFile{
					Data: []byte("---\nname: test\nversion: 1.0\n---\n# Test"),
				},
			},
			skillDir: "packages/test-skill/skills/test-skill",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := readSkillDescription(tt.files, tt.skillDir)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// ---------- newListCommand (Cobra command-level) tests ----------

// TestListCommandReturnsNilError verifies that the list command executes
// without error when given a valid filesystem.
func TestListCommandReturnsNilError(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	cmd := newListCommand(testContentFS())
	cmd.SetArgs([]string{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SilenceUsage = true

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
}

// TestListCommandOutputContainsPackages verifies that executing the list
// command produces output containing the package names from the test FS.
func TestListCommandOutputContainsPackages(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	var buf bytes.Buffer
	cmd := newListCommand(testContentFS())
	cmd.SetArgs([]string{})
	// Redirect both Cobra output and color.Output to our buffer.
	cmd.SetOut(&buf)
	old := color.Output
	color.Output = &buf
	t.Cleanup(func() { color.Output = old })

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	expected := []string{
		"Packages",
		"git-workflow",
		"developer-mentor",
	}
	for _, want := range expected {
		if !strings.Contains(output, want) {
			t.Errorf("output should contain %q, got:\n%s", want, output)
		}
	}
}

// TestListCommandOutputContainsAssistants verifies that the output includes
// the Assistants section with all registered assistant names and descriptions.
func TestListCommandOutputContainsAssistants(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	var buf bytes.Buffer
	cmd := newListCommand(testContentFS())
	cmd.SetArgs([]string{})
	cmd.SetOut(&buf)

	old := color.Output
	color.Output = &buf
	t.Cleanup(func() { color.Output = old })

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	expected := []string{
		"Assistants",
		"copilot",
		"claude",
		"opencode",
		"GitHub Copilot",
		"Claude Code",
		"OpenCode",
	}
	for _, want := range expected {
		if !strings.Contains(output, want) {
			t.Errorf("output should contain %q, got:\n%s", want, output)
		}
	}
}

// TestListCommandShowsDescriptions verifies that when a package's SKILL.md
// has a frontmatter description, it appears in the list output.
func TestListCommandShowsDescriptions(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	var buf bytes.Buffer
	cmd := newListCommand(testContentFSWithDescriptions())
	cmd.SetArgs([]string{})
	cmd.SetOut(&buf)

	old := color.Output
	color.Output = &buf
	t.Cleanup(func() { color.Output = old })

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	expected := []string{
		"Conventional commit workflows",
		"Socratic development mentoring",
	}
	for _, want := range expected {
		if !strings.Contains(output, want) {
			t.Errorf("output should contain description %q, got:\n%s", want, output)
		}
	}
}
