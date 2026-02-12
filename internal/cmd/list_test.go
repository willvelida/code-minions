package cmd

import (
	"testing"
	"testing/fstest"
)

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
