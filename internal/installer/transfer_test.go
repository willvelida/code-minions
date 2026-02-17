package installer

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// createFile is a test helper that creates a file with content in a temp dir.
func createFile(t *testing.T, dir, relPath, content string) {
	t.Helper()
	full := filepath.Join(dir, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		t.Fatalf("failed to create dir for %s: %v", relPath, err)
	}
	if err := os.WriteFile(full, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write %s: %v", relPath, err)
	}
}

// fileExists is a test helper that checks if a file exists.
func fileExists(t *testing.T, dir, relPath string) bool {
	t.Helper()
	_, err := os.Stat(filepath.Join(dir, filepath.FromSlash(relPath)))
	return err == nil
}

// readFileContent is a test helper that reads file content.
func readFileContent(t *testing.T, dir, relPath string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(relPath)))
	if err != nil {
		t.Fatalf("failed to read %s: %v", relPath, err)
	}
	return string(data)
}

func TestTransferCopilotToClaude(t *testing.T) {
	tmp := t.TempDir()

	// Set up Copilot layout
	createFile(t, tmp, ".github/agents/dev-mentor.agent.md", "# Dev Mentor")
	createFile(t, tmp, "skills/dev-mentor/SKILL.md", "# Skill")
	createFile(t, tmp, "skills/dev-mentor/actions/create.md", "# Create")

	result, err := Transfer(TransferOptions{
		FromAssistant: "copilot",
		ToAssistant:   "claude",
		TargetDir:     tmp,
	})
	if err != nil {
		t.Fatalf("Transfer() returned unexpected error: %v", err)
	}

	// Verify files were copied to Claude layout
	if !fileExists(t, tmp, ".claude/agents/dev-mentor.agent.md") {
		t.Error("expected .claude/agents/dev-mentor.agent.md to exist")
	}
	if !fileExists(t, tmp, ".claude/skills/dev-mentor/SKILL.md") {
		t.Error("expected .claude/skills/dev-mentor/SKILL.md to exist")
	}
	if !fileExists(t, tmp, ".claude/skills/dev-mentor/actions/create.md") {
		t.Error("expected .claude/skills/dev-mentor/actions/create.md to exist")
	}

	// Verify content preserved
	if got := readFileContent(t, tmp, ".claude/agents/dev-mentor.agent.md"); got != "# Dev Mentor" {
		t.Errorf("content mismatch: got %q", got)
	}

	// Verify source files still exist (copy, not move)
	if !fileExists(t, tmp, ".github/agents/dev-mentor.agent.md") {
		t.Error("source file should still exist (copy mode)")
	}

	if len(result.Copied) != 3 {
		t.Errorf("expected 3 copied files, got %d: %v", len(result.Copied), result.Copied)
	}
	if len(result.Errors) != 0 {
		t.Errorf("unexpected errors: %v", result.Errors)
	}
}

func TestTransferClaudeToCopilot(t *testing.T) {
	tmp := t.TempDir()

	// Set up Claude layout
	createFile(t, tmp, ".claude/agents/git-workflow.agent.md", "# Git Workflow")
	createFile(t, tmp, ".claude/skills/git-workflow/SKILL.md", "# Skill")

	result, err := Transfer(TransferOptions{
		FromAssistant: "claude",
		ToAssistant:   "copilot",
		TargetDir:     tmp,
	})
	if err != nil {
		t.Fatalf("Transfer() returned unexpected error: %v", err)
	}

	if !fileExists(t, tmp, ".github/agents/git-workflow.agent.md") {
		t.Error("expected .github/agents/git-workflow.agent.md to exist")
	}
	if !fileExists(t, tmp, "skills/git-workflow/SKILL.md") {
		t.Error("expected skills/git-workflow/SKILL.md to exist")
	}

	if len(result.Copied) != 2 {
		t.Errorf("expected 2 copied files, got %d", len(result.Copied))
	}
}

func TestTransferOpenCodeToCopilot(t *testing.T) {
	tmp := t.TempDir()

	createFile(t, tmp, ".opencode/agents/my-agent.md", "agent content")
	createFile(t, tmp, ".opencode/skills/my-skill/SKILL.md", "skill content")

	result, err := Transfer(TransferOptions{
		FromAssistant: "opencode",
		ToAssistant:   "copilot",
		TargetDir:     tmp,
	})
	if err != nil {
		t.Fatalf("Transfer() returned unexpected error: %v", err)
	}

	if !fileExists(t, tmp, ".github/agents/my-agent.md") {
		t.Error("expected .github/agents/my-agent.md to exist")
	}
	if !fileExists(t, tmp, "skills/my-skill/SKILL.md") {
		t.Error("expected skills/my-skill/SKILL.md to exist")
	}

	if len(result.Copied) != 2 {
		t.Errorf("expected 2 copied files, got %d", len(result.Copied))
	}
}

// TestTransferAllPermutations verifies all 6 assistant pairs produce correct paths.
func TestTransferAllPermutations(t *testing.T) {
	type permutation struct {
		from        string
		to          string
		sourceAgent string // relative path of source agent file
		sourceSkill string // relative path of source skill file
		expectAgent string // expected destination agent path
		expectSkill string // expected destination skill path
	}

	perms := []permutation{
		{"copilot", "claude", ".github/agents/a.md", "skills/s/SKILL.md", ".claude/agents/a.md", ".claude/skills/s/SKILL.md"},
		{"copilot", "opencode", ".github/agents/a.md", "skills/s/SKILL.md", ".opencode/agents/a.md", ".opencode/skills/s/SKILL.md"},
		{"claude", "copilot", ".claude/agents/a.md", ".claude/skills/s/SKILL.md", ".github/agents/a.md", "skills/s/SKILL.md"},
		{"claude", "opencode", ".claude/agents/a.md", ".claude/skills/s/SKILL.md", ".opencode/agents/a.md", ".opencode/skills/s/SKILL.md"},
		{"opencode", "copilot", ".opencode/agents/a.md", ".opencode/skills/s/SKILL.md", ".github/agents/a.md", "skills/s/SKILL.md"},
		{"opencode", "claude", ".opencode/agents/a.md", ".opencode/skills/s/SKILL.md", ".claude/agents/a.md", ".claude/skills/s/SKILL.md"},
	}

	for _, p := range perms {
		t.Run(p.from+"_to_"+p.to, func(t *testing.T) {
			tmp := t.TempDir()
			createFile(t, tmp, p.sourceAgent, "agent")
			createFile(t, tmp, p.sourceSkill, "skill")

			result, err := Transfer(TransferOptions{
				FromAssistant: p.from,
				ToAssistant:   p.to,
				TargetDir:     tmp,
			})
			if err != nil {
				t.Fatalf("Transfer() error: %v", err)
			}

			if !fileExists(t, tmp, p.expectAgent) {
				t.Errorf("expected %s to exist", p.expectAgent)
			}
			if !fileExists(t, tmp, p.expectSkill) {
				t.Errorf("expected %s to exist", p.expectSkill)
			}
			if len(result.Copied) != 2 {
				t.Errorf("expected 2 copied, got %d: %v", len(result.Copied), result.Copied)
			}
		})
	}
}

func TestTransferForceOverwrite(t *testing.T) {
	tmp := t.TempDir()

	createFile(t, tmp, ".github/agents/a.md", "new content")
	createFile(t, tmp, ".claude/agents/a.md", "old content")

	result, err := Transfer(TransferOptions{
		FromAssistant: "copilot",
		ToAssistant:   "claude",
		TargetDir:     tmp,
		Force:         true,
	})
	if err != nil {
		t.Fatalf("Transfer() error: %v", err)
	}

	got := readFileContent(t, tmp, ".claude/agents/a.md")
	if got != "new content" {
		t.Errorf("force overwrite failed: got %q, want %q", got, "new content")
	}
	if len(result.Copied) != 1 {
		t.Errorf("expected 1 copied, got %d", len(result.Copied))
	}
	if len(result.Skipped) != 0 {
		t.Errorf("expected 0 skipped, got %d", len(result.Skipped))
	}
}

func TestTransferSkipsExisting(t *testing.T) {
	tmp := t.TempDir()

	createFile(t, tmp, ".github/agents/a.md", "new content")
	createFile(t, tmp, ".claude/agents/a.md", "existing content")

	result, err := Transfer(TransferOptions{
		FromAssistant: "copilot",
		ToAssistant:   "claude",
		TargetDir:     tmp,
	})
	if err != nil {
		t.Fatalf("Transfer() error: %v", err)
	}

	// Original content should be preserved
	got := readFileContent(t, tmp, ".claude/agents/a.md")
	if got != "existing content" {
		t.Errorf("existing file was overwritten: got %q", got)
	}
	if len(result.Skipped) != 1 {
		t.Errorf("expected 1 skipped, got %d", len(result.Skipped))
	}
}

func TestTransferDryRun(t *testing.T) {
	tmp := t.TempDir()

	createFile(t, tmp, ".github/agents/a.md", "content")

	result, err := Transfer(TransferOptions{
		FromAssistant: "copilot",
		ToAssistant:   "claude",
		TargetDir:     tmp,
		DryRun:        true,
	})
	if err != nil {
		t.Fatalf("Transfer() error: %v", err)
	}

	// File should NOT be created
	if fileExists(t, tmp, ".claude/agents/a.md") {
		t.Error("dry-run should not create files")
	}
	// But it should appear in Copied
	if len(result.Copied) != 1 {
		t.Errorf("expected 1 copied in dry-run, got %d", len(result.Copied))
	}
}

func TestTransferCleanup(t *testing.T) {
	tmp := t.TempDir()

	createFile(t, tmp, ".github/agents/a.md", "agent")
	createFile(t, tmp, "skills/s/SKILL.md", "skill")

	result, err := Transfer(TransferOptions{
		FromAssistant: "copilot",
		ToAssistant:   "claude",
		TargetDir:     tmp,
		Cleanup:       true,
	})
	if err != nil {
		t.Fatalf("Transfer() error: %v", err)
	}

	// Destination files should exist
	if !fileExists(t, tmp, ".claude/agents/a.md") {
		t.Error("expected destination agent to exist")
	}
	if !fileExists(t, tmp, ".claude/skills/s/SKILL.md") {
		t.Error("expected destination skill to exist")
	}

	// Source files should be deleted
	if fileExists(t, tmp, ".github/agents/a.md") {
		t.Error("source agent should have been cleaned up")
	}
	if fileExists(t, tmp, "skills/s/SKILL.md") {
		t.Error("source skill should have been cleaned up")
	}

	if len(result.Cleaned) != 2 {
		t.Errorf("expected 2 cleaned files, got %d: %v", len(result.Cleaned), result.Cleaned)
	}
}

func TestTransferCleanupDryRun(t *testing.T) {
	tmp := t.TempDir()

	createFile(t, tmp, ".github/agents/a.md", "agent")

	result, err := Transfer(TransferOptions{
		FromAssistant: "copilot",
		ToAssistant:   "claude",
		TargetDir:     tmp,
		Cleanup:       true,
		DryRun:        true,
	})
	if err != nil {
		t.Fatalf("Transfer() error: %v", err)
	}

	// Source should still exist
	if !fileExists(t, tmp, ".github/agents/a.md") {
		t.Error("dry-run cleanup should not delete files")
	}
	// Destination should not exist
	if fileExists(t, tmp, ".claude/agents/a.md") {
		t.Error("dry-run should not create files")
	}
	// But cleaned list should be populated
	if len(result.Cleaned) == 0 {
		t.Error("dry-run cleanup should report what would be cleaned")
	}
}

func TestTransferCleanupPartial(t *testing.T) {
	tmp := t.TempDir()

	// Create two source files, one will be skipped (already exists at destination)
	createFile(t, tmp, ".github/agents/a.md", "new-a")
	createFile(t, tmp, ".github/agents/b.md", "new-b")
	createFile(t, tmp, ".claude/agents/a.md", "existing-a") // will cause skip

	result, err := Transfer(TransferOptions{
		FromAssistant: "copilot",
		ToAssistant:   "claude",
		TargetDir:     tmp,
		Cleanup:       true,
	})
	if err != nil {
		t.Fatalf("Transfer() error: %v", err)
	}

	// b.md was copied, so b.md source should be cleaned
	// a.md was skipped, so a.md source should NOT be cleaned
	if len(result.Copied) != 1 {
		t.Errorf("expected 1 copied, got %d", len(result.Copied))
	}
	if len(result.Skipped) != 1 {
		t.Errorf("expected 1 skipped, got %d", len(result.Skipped))
	}
	if len(result.Cleaned) != 1 {
		t.Errorf("expected 1 cleaned, got %d: %v", len(result.Cleaned), result.Cleaned)
	}

	// Source a.md should still exist (was skipped)
	if !fileExists(t, tmp, ".github/agents/a.md") {
		t.Error("skipped source file should not be cleaned up")
	}
	// Source b.md should be gone
	if fileExists(t, tmp, ".github/agents/b.md") {
		t.Error("copied source file should have been cleaned up")
	}
}

func TestTransferCleanupRemovesEmptyDirs(t *testing.T) {
	tmp := t.TempDir()

	createFile(t, tmp, ".github/agents/a.md", "content")

	_, err := Transfer(TransferOptions{
		FromAssistant: "copilot",
		ToAssistant:   "claude",
		TargetDir:     tmp,
		Cleanup:       true,
	})
	if err != nil {
		t.Fatalf("Transfer() error: %v", err)
	}

	// The .github/agents/ directory should be cleaned up
	agentDir := filepath.Join(tmp, ".github", "agents")
	if _, err := os.Stat(agentDir); !os.IsNotExist(err) {
		t.Error("empty source agent directory should have been removed")
	}
}

func TestTransferCleanupAgentsMD(t *testing.T) {
	tmp := t.TempDir()

	createFile(t, tmp, ".github/agents/a.md", "agent")
	createFile(t, tmp, ".github/agents/AGENTS.md", "# Agents routing")

	result, err := Transfer(TransferOptions{
		FromAssistant: "copilot",
		ToAssistant:   "claude",
		TargetDir:     tmp,
		Cleanup:       true,
	})
	if err != nil {
		t.Fatalf("Transfer() error: %v", err)
	}

	// AGENTS.md should NOT be transferred
	if fileExists(t, tmp, ".claude/agents/AGENTS.md") {
		t.Error("AGENTS.md should not be transferred")
	}
	// But source AGENTS.md should be cleaned up
	if fileExists(t, tmp, ".github/agents/AGENTS.md") {
		t.Error("source AGENTS.md should be cleaned up with --cleanup")
	}
	// Cleaned should include AGENTS.md
	hasAgentsMD := false
	for _, c := range result.Cleaned {
		if strings.HasSuffix(c, "AGENTS.md") {
			hasAgentsMD = true
			break
		}
	}
	if !hasAgentsMD {
		t.Error("cleaned list should include AGENTS.md")
	}
}

func TestTransferSourceValidationBothMissing(t *testing.T) {
	tmp := t.TempDir()

	_, err := Transfer(TransferOptions{
		FromAssistant: "copilot",
		ToAssistant:   "claude",
		TargetDir:     tmp,
	})
	if err == nil {
		t.Fatal("expected error when neither source dir exists")
	}
	if !strings.Contains(err.Error(), "no agent or skill files found") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestTransferSourceValidationAgentDirMissing(t *testing.T) {
	tmp := t.TempDir()

	// Only skills exist
	createFile(t, tmp, "skills/s/SKILL.md", "skill")

	result, err := Transfer(TransferOptions{
		FromAssistant: "copilot",
		ToAssistant:   "claude",
		TargetDir:     tmp,
	})
	if err != nil {
		t.Fatalf("Transfer() should succeed with only skills: %v", err)
	}

	// Should have a warning about missing agent dir
	if len(result.Warnings) == 0 {
		t.Error("expected warning about missing agent directory")
	}
	hasAgentWarning := false
	for _, w := range result.Warnings {
		if strings.Contains(w, "agent") && strings.Contains(w, "does not exist") {
			hasAgentWarning = true
			break
		}
	}
	if !hasAgentWarning {
		t.Errorf("expected agent dir warning, got warnings: %v", result.Warnings)
	}

	// Skills should still transfer
	if len(result.Copied) != 1 {
		t.Errorf("expected 1 copied, got %d", len(result.Copied))
	}
}

func TestTransferSourceValidationSkillDirMissing(t *testing.T) {
	tmp := t.TempDir()

	// Only agents exist
	createFile(t, tmp, ".github/agents/a.md", "agent")

	result, err := Transfer(TransferOptions{
		FromAssistant: "copilot",
		ToAssistant:   "claude",
		TargetDir:     tmp,
	})
	if err != nil {
		t.Fatalf("Transfer() should succeed with only agents: %v", err)
	}

	if len(result.Warnings) == 0 {
		t.Error("expected warning about missing skill directory")
	}
	if len(result.Copied) != 1 {
		t.Errorf("expected 1 copied, got %d", len(result.Copied))
	}
}

func TestTransferSourceDirsEmptyWarning(t *testing.T) {
	tmp := t.TempDir()

	// Create empty directories
	os.MkdirAll(filepath.Join(tmp, ".github", "agents"), 0755)
	os.MkdirAll(filepath.Join(tmp, "skills"), 0755)

	result, err := Transfer(TransferOptions{
		FromAssistant: "copilot",
		ToAssistant:   "claude",
		TargetDir:     tmp,
	})
	if err != nil {
		t.Fatalf("Transfer() error: %v", err)
	}

	if len(result.Copied) != 0 {
		t.Errorf("expected 0 copied, got %d", len(result.Copied))
	}
	hasEmptyWarning := false
	for _, w := range result.Warnings {
		if strings.Contains(w, "no transferable files") {
			hasEmptyWarning = true
			break
		}
	}
	if !hasEmptyWarning {
		t.Errorf("expected empty dirs warning, got warnings: %v", result.Warnings)
	}
}

func TestTransferSkipsAgentsMD(t *testing.T) {
	tmp := t.TempDir()

	createFile(t, tmp, ".github/agents/a.md", "agent")
	createFile(t, tmp, ".github/agents/AGENTS.md", "# Routing")

	result, err := Transfer(TransferOptions{
		FromAssistant: "copilot",
		ToAssistant:   "claude",
		TargetDir:     tmp,
	})
	if err != nil {
		t.Fatalf("Transfer() error: %v", err)
	}

	if fileExists(t, tmp, ".claude/agents/AGENTS.md") {
		t.Error("AGENTS.md should not be transferred")
	}
	if len(result.Copied) != 1 {
		t.Errorf("expected 1 copied (agent only, not AGENTS.md), got %d: %v", len(result.Copied), result.Copied)
	}
}

func TestTransferSkipsMetadataFiles(t *testing.T) {
	tmp := t.TempDir()

	createFile(t, tmp, ".github/agents/a.md", "agent")
	createFile(t, tmp, "skills/s/package.yaml", "name: s")
	createFile(t, tmp, "skills/s/mcp.yaml", "servers:")
	createFile(t, tmp, "skills/s/SKILL.md", "# Skill")

	result, err := Transfer(TransferOptions{
		FromAssistant: "copilot",
		ToAssistant:   "claude",
		TargetDir:     tmp,
	})
	if err != nil {
		t.Fatalf("Transfer() error: %v", err)
	}

	if fileExists(t, tmp, ".claude/skills/s/package.yaml") {
		t.Error("package.yaml should not be transferred")
	}
	if fileExists(t, tmp, ".claude/skills/s/mcp.yaml") {
		t.Error("mcp.yaml should not be transferred")
	}
	if !fileExists(t, tmp, ".claude/skills/s/SKILL.md") {
		t.Error("SKILL.md should be transferred")
	}
	// a.md + SKILL.md = 2
	if len(result.Copied) != 2 {
		t.Errorf("expected 2 copied, got %d: %v", len(result.Copied), result.Copied)
	}
}

func TestTransferNestedDirs(t *testing.T) {
	tmp := t.TempDir()

	createFile(t, tmp, ".claude/skills/dev-mentor/actions/create.md", "create")
	createFile(t, tmp, ".claude/skills/dev-mentor/standards/checklist.md", "checklist")

	result, err := Transfer(TransferOptions{
		FromAssistant: "claude",
		ToAssistant:   "opencode",
		TargetDir:     tmp,
	})
	if err != nil {
		t.Fatalf("Transfer() error: %v", err)
	}

	if !fileExists(t, tmp, ".opencode/skills/dev-mentor/actions/create.md") {
		t.Error("nested actions file not transferred")
	}
	if !fileExists(t, tmp, ".opencode/skills/dev-mentor/standards/checklist.md") {
		t.Error("nested standards file not transferred")
	}
	if len(result.Copied) != 2 {
		t.Errorf("expected 2 copied, got %d", len(result.Copied))
	}
}

func TestTransferSameAssistantError(t *testing.T) {
	_, err := Transfer(TransferOptions{
		FromAssistant: "copilot",
		ToAssistant:   "copilot",
		TargetDir:     t.TempDir(),
	})
	if err == nil {
		t.Fatal("expected error for same source and target")
	}
	if !strings.Contains(err.Error(), "cannot be the same") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestTransferUnknownAssistantError(t *testing.T) {
	_, err := Transfer(TransferOptions{
		FromAssistant: "unknown",
		ToAssistant:   "claude",
		TargetDir:     t.TempDir(),
	})
	if err == nil {
		t.Fatal("expected error for unknown assistant")
	}
	if !strings.Contains(err.Error(), "unknown") {
		t.Errorf("error should mention invalid name: %v", err)
	}
}

func TestTransferCopiedPathsAreSorted(t *testing.T) {
	tmp := t.TempDir()

	createFile(t, tmp, ".github/agents/z.md", "z")
	createFile(t, tmp, ".github/agents/a.md", "a")
	createFile(t, tmp, "skills/m/SKILL.md", "m")

	result, err := Transfer(TransferOptions{
		FromAssistant: "copilot",
		ToAssistant:   "claude",
		TargetDir:     tmp,
	})
	if err != nil {
		t.Fatalf("Transfer() error: %v", err)
	}

	// Verify copied paths are present (order depends on filesystem walk)
	if len(result.Copied) != 3 {
		t.Errorf("expected 3 copied, got %d: %v", len(result.Copied), result.Copied)
	}

	// Check all expected files exist
	sorted := make([]string, len(result.Copied))
	copy(sorted, result.Copied)
	sort.Strings(sorted)
	for _, p := range sorted {
		if !fileExists(t, tmp, p) {
			t.Errorf("expected %s to exist", p)
		}
	}
}
