package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestBuildPackageList(t *testing.T) {
	content := testContentFS()

	tests := []struct {
		name        string
		packageFlag string
		expectDirs  []string
		expectError bool
	}{
		{
			name:       "no flag installs everything",
			expectDirs: []string{"packages/developer-mentor", "packages/git-workflow"},
		},
		{
			name:        "single package",
			packageFlag: "git-workflow",
			expectDirs:  []string{"packages/git-workflow"},
		},
		{
			name:        "invalid package returns error",
			packageFlag: "nonexistent",
			expectError: true,
		},
		{
			name:        "whitespace around names is trimmed",
			packageFlag: " git-workflow , developer-mentor ",
			expectDirs:  []string{"packages/git-workflow", "packages/developer-mentor"},
		},
		{
			name:        "empty items from double commas are skipped",
			packageFlag: "git-workflow,,developer-mentor",
			expectDirs:  []string{"packages/git-workflow", "packages/developer-mentor"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dirs, err := buildPackageList(content, tt.packageFlag)

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
		"packages/git-workflow/agents/git-workflow.agent.md":         &fstest.MapFile{Data: []byte("# Git Agent")},
		"packages/git-workflow/skills/git-workflow/SKILL.md":         &fstest.MapFile{Data: []byte("# Git")},
		"packages/developer-mentor/agents/developer-mentor.agent.md": &fstest.MapFile{Data: []byte("# Mentor Agent")},
		"packages/developer-mentor/skills/developer-mentor/SKILL.md": &fstest.MapFile{Data: []byte("# Mentor")},
	}
}

// TestInstallForCopilotRemapsPaths verifies that --for copilot places
// agent files in .github/agents/ while keeping skills in their default
// locations.
func TestInstallForCopilotRemapsPaths(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	cmd := newInstallCommand(content)
	cmd.SetArgs([]string{
		"--package", "git-workflow",
		"--for", "copilot",
		"--target", target,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Agent file should be in .github/agents/ (remapped from agents/)
	copilotAgent := filepath.Join(target, ".github", "agents", "git-workflow.agent.md")
	if _, err := os.Stat(copilotAgent); os.IsNotExist(err) {
		t.Errorf("expected agent at Copilot path: %s", copilotAgent)
	}

	// Skills should stay at skills/ (Copilot uses project-root skills/)
	skillFile := filepath.Join(target, "skills", "git-workflow", "SKILL.md")
	if _, err := os.Stat(skillFile); os.IsNotExist(err) {
		t.Errorf("expected skill at default path: %s", skillFile)
	}

	// Agent should NOT be at the old agents/ path
	oldAgent := filepath.Join(target, "agents", "git-workflow.agent.md")
	if _, err := os.Stat(oldAgent); !os.IsNotExist(err) {
		t.Errorf("agent should NOT exist at old path: %s", oldAgent)
	}
}

// TestInstallForClaudeRemapsPaths verifies that --for claude places
// agent files in .claude/agents/ and skills in .claude/skills/.
func TestInstallForClaudeRemapsPaths(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	cmd := newInstallCommand(content)
	cmd.SetArgs([]string{
		"--package", "git-workflow",
		"--for", "claude",
		"--target", target,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Agent should be in .claude/agents/
	claudeAgent := filepath.Join(target, ".claude", "agents", "git-workflow.agent.md")
	if _, err := os.Stat(claudeAgent); os.IsNotExist(err) {
		t.Errorf("expected agent at Claude path: %s", claudeAgent)
	}

	// Skills should be in .claude/skills/
	claudeSkill := filepath.Join(target, ".claude", "skills", "git-workflow", "SKILL.md")
	if _, err := os.Stat(claudeSkill); os.IsNotExist(err) {
		t.Errorf("expected skill at Claude path: %s", claudeSkill)
	}
}

// TestInstallForUnknownAssistantReturnsError verifies that --for
// with an invalid assistant name returns a helpful error.
func TestInstallForUnknownAssistantReturnsError(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	cmd := newInstallCommand(content)
	cmd.SetArgs([]string{
		"--package", "git-workflow",
		"--for", "vscode",
		"--target", target,
	})

	// Silence Cobra's default error printing so it doesn't clutter test output
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for unknown assistant, got nil")
	}
}

// TestInstallWithoutForPreservesBehaviour verifies that omitting
// --for produces the same output as before (no path remapping).
func TestInstallWithoutForPreservesBehaviour(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	cmd := newInstallCommand(content)
	cmd.SetArgs([]string{
		"--package", "git-workflow",
		"--target", target,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Without --for, agent lands in agents/ (the generic location)
	genericAgent := filepath.Join(target, "agents", "git-workflow.agent.md")
	if _, err := os.Stat(genericAgent); os.IsNotExist(err) {
		t.Errorf("expected agent at generic path: %s", genericAgent)
	}

	// Skills in skills/
	genericSkill := filepath.Join(target, "skills", "git-workflow", "SKILL.md")
	if _, err := os.Stat(genericSkill); os.IsNotExist(err) {
		t.Errorf("expected skill at generic path: %s", genericSkill)
	}
}

// TestInstallCreatesAgentsMD verifies that installing a package
// creates AGENTS.md when it does not exist.
func TestInstallCreatesAgentsMD(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	cmd := newInstallCommand(content)
	cmd.SetArgs([]string{
		"--package", "git-workflow",
		"--target", target,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	agentsMD := filepath.Join(target, "AGENTS.md")
	data, err := os.ReadFile(agentsMD)
	if err != nil {
		t.Fatalf("expected AGENTS.md to be created: %v", err)
	}

	if !strings.Contains(string(data), "code-minions") {
		t.Errorf("AGENTS.md content missing expected text, got:\n%s", data)
	}
}

// TestInstallSkipsExistingAgentsMD verifies that installing a package
// does not overwrite an existing AGENTS.md.
func TestInstallSkipsExistingAgentsMD(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	// Pre-create AGENTS.md with custom content
	original := []byte("# My Custom Agents\n")
	if err := os.WriteFile(filepath.Join(target, "AGENTS.md"), original, 0644); err != nil {
		t.Fatal(err)
	}

	cmd := newInstallCommand(content)
	cmd.SetArgs([]string{
		"--package", "git-workflow",
		"--target", target,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(target, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != string(original) {
		t.Errorf("AGENTS.md was overwritten:\n  got:  %q\n  want: %q", got, original)
	}
}

// TestInstallForCopilotCreatesAgentsMDAtMappedPath verifies that
// --for copilot creates AGENTS.md at .github/agents/AGENTS.md.
func TestInstallForCopilotCreatesAgentsMDAtMappedPath(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	cmd := newInstallCommand(content)
	cmd.SetArgs([]string{
		"--package", "git-workflow",
		"--for", "copilot",
		"--target", target,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	agentsMD := filepath.Join(target, ".github", "agents", "AGENTS.md")
	if _, err := os.Stat(agentsMD); os.IsNotExist(err) {
		t.Errorf("expected AGENTS.md at Copilot path: %s", agentsMD)
	}

	// Should NOT exist at default path
	defaultPath := filepath.Join(target, "AGENTS.md")
	if _, err := os.Stat(defaultPath); !os.IsNotExist(err) {
		t.Errorf("AGENTS.md should NOT exist at default path when using --for copilot")
	}
}
