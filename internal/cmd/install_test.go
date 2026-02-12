package cmd

import (
	"testing"
	"testing/fstest"
)

func TestBuildDirList(t *testing.T) {
	content := testContentFS()

	tests := []struct {
		name          string
		includeAgents bool
		skillsFlag    string
		standardsFlag string
		expectDirs    []string
		expectError   bool
	}{
		{
			name:       "no flag installs everything",
			expectDirs: []string{"agents", "skills", "standards"},
		},
		{
			name:          "agents only",
			includeAgents: true,
			expectDirs:    []string{"agents"},
		},
		{
			name:       "single skill",
			skillsFlag: "git-workflow",
			expectDirs: []string{"skills/git-workflow"},
		},
		{
			name:          "single standard",
			standardsFlag: "python",
			expectDirs:    []string{"standards/languages/python", "standards/languages/standards.index.md"},
		},
		{
			name:        "invalid skill returns error",
			skillsFlag:  "nonexistent",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dirs, err := buildDirList(content, tt.includeAgents, tt.skillsFlag, tt.standardsFlag)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(dirs) != len(tt.expectDirs) {
				t.Fatalf("dir count: got %d, want %d\n  got:  %v\n  want: %v", len(dirs), len(tt.expectDirs), dirs, tt.expectDirs)
			}

			for i, dir := range dirs {
				if dir != tt.expectDirs[i] {
					t.Errorf("dirs[%d]: got %q, want %q", i, dir, tt.expectDirs[i])
				}
			}
		})
	}
}

func testContentFS() fstest.MapFS {
	return fstest.MapFS{
		"agents/AGENTS.md":                       &fstest.MapFile{Data: []byte("# Agents")},
		"skills/git-workflow/SKILL.md":           &fstest.MapFile{Data: []byte("# Git")},
		"skills/developer-mentor/SKILL.md":       &fstest.MapFile{Data: []byte("# Mentor")},
		"standards/languages/python/core.md":     &fstest.MapFile{Data: []byte("# Python")},
		"standards/languages/typescript/core.md": &fstest.MapFile{Data: []byte("# TS")},
	}
}
