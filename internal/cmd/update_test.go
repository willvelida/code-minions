package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestUpdateOverwritesExistingFiles verifies that the update command
// overwrites files that already exist on disk (force install behaviour).
func TestUpdateOverwritesExistingFiles(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	// First, install the package so files exist on disk
	installCmd := newInstallCommand(content)
	installCmd.SetArgs([]string{
		"--package", "git-workflow",
		"--target", target,
	})
	if err := installCmd.Execute(); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Modify a file to simulate stale content
	agentFile := filepath.Join(target, "agents", "git-workflow.agent.md")
	if err := os.WriteFile(agentFile, []byte("old content"), 0644); err != nil {
		t.Fatalf("failed to write stale content: %v", err)
	}

	// Run update — should overwrite with the embedded content
	updateCmd := newUpdateCommand(content)
	updateCmd.SetArgs([]string{
		"--package", "git-workflow",
		"--target", target,
	})
	if err := updateCmd.Execute(); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	// Verify the file was overwritten with the embedded content
	got, err := os.ReadFile(agentFile)
	if err != nil {
		t.Fatalf("failed to read updated file: %v", err)
	}
	if string(got) == "old content" {
		t.Error("update did not overwrite the stale file")
	}
	if !strings.Contains(string(got), "# Git Agent") {
		t.Errorf("updated content should contain '# Git Agent', got %q", got)
	}
}

// TestUpdateDryRunDoesNotWriteFiles verifies that --dry-run shows
// what would be updated without actually modifying any files.
func TestUpdateDryRunDoesNotWriteFiles(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	// Install first so files exist
	installCmd := newInstallCommand(content)
	installCmd.SetArgs([]string{
		"--package", "git-workflow",
		"--target", target,
	})
	if err := installCmd.Execute(); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Modify a file
	agentFile := filepath.Join(target, "agents", "git-workflow.agent.md")
	if err := os.WriteFile(agentFile, []byte("old content"), 0644); err != nil {
		t.Fatalf("failed to write stale content: %v", err)
	}

	// Run update with --dry-run
	updateCmd := newUpdateCommand(content)
	updateCmd.SetArgs([]string{
		"--package", "git-workflow",
		"--target", target,
		"--dry-run",
	})
	if err := updateCmd.Execute(); err != nil {
		t.Fatalf("update dry-run failed: %v", err)
	}

	// Verify the file was NOT overwritten
	got, err := os.ReadFile(agentFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(got) != "old content" {
		t.Errorf("dry-run should not modify files: got %q, want %q", got, "old content")
	}
}

// TestUpdateInvalidPackageReturnsError verifies that updating a
// non-existent package returns a helpful error.
func TestUpdateInvalidPackageReturnsError(t *testing.T) {
	content := testContentFS()

	cmd := newUpdateCommand(content)
	cmd.SetArgs([]string{
		"--package", "nonexistent",
	})
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid package, got nil")
	}
}

// TestUpdateForCopilotRemapsPaths verifies that --for copilot places
// updated files in the correct Copilot-specific directories.
func TestUpdateForCopilotRemapsPaths(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	// Install with --for copilot first
	installCmd := newInstallCommand(content)
	installCmd.SetArgs([]string{
		"--package", "git-workflow",
		"--for", "copilot",
		"--target", target,
	})
	if err := installCmd.Execute(); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Run update with --for copilot
	updateCmd := newUpdateCommand(content)
	updateCmd.SetArgs([]string{
		"--package", "git-workflow",
		"--for", "copilot",
		"--target", target,
	})
	if err := updateCmd.Execute(); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	// Agent should be at Copilot path
	copilotAgent := filepath.Join(target, ".github", "agents", "git-workflow.agent.md")
	if _, err := os.Stat(copilotAgent); os.IsNotExist(err) {
		t.Errorf("expected agent at Copilot path: %s", copilotAgent)
	}

	// Skills should be at skills/
	skillFile := filepath.Join(target, "skills", "git-workflow", "SKILL.md")
	if _, err := os.Stat(skillFile); os.IsNotExist(err) {
		t.Errorf("expected skill at default path: %s", skillFile)
	}
}

// TestUpdateDoesNotModifyAgentsMD verifies that the update command
// does not create or modify AGENTS.md.
func TestUpdateDoesNotModifyAgentsMD(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	// Pre-create AGENTS.md with custom content
	original := []byte("# My Custom Agents\n")
	if err := os.WriteFile(filepath.Join(target, "AGENTS.md"), original, 0644); err != nil {
		t.Fatal(err)
	}

	// Run update
	updateCmd := newUpdateCommand(content)
	updateCmd.SetArgs([]string{
		"--package", "git-workflow",
		"--target", target,
	})
	if err := updateCmd.Execute(); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	// AGENTS.md should be unchanged
	got, err := os.ReadFile(filepath.Join(target, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(original) {
		t.Errorf("AGENTS.md was modified during update:\n  got:  %q\n  want: %q", got, original)
	}
}

// TestUpdateNoFlagsUpdatesOnlyInstalled verifies that running update
// with no flags only updates packages and standards that are already
// installed — it does not install new ones.
func TestUpdateNoFlagsUpdatesOnlyInstalled(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	// Install only git-workflow (not developer-mentor)
	installCmd := newInstallCommand(content)
	installCmd.SetArgs([]string{
		"--package", "git-workflow",
		"--target", target,
	})
	if err := installCmd.Execute(); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Run update with no flags
	updateCmd := newUpdateCommand(content)
	updateCmd.SetArgs([]string{
		"--target", target,
	})
	if err := updateCmd.Execute(); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	// git-workflow should still exist (it was installed and updated)
	gitAgent := filepath.Join(target, "agents", "git-workflow.agent.md")
	if _, err := os.Stat(gitAgent); os.IsNotExist(err) {
		t.Errorf("expected git-workflow agent to be updated: %s", gitAgent)
	}

	// developer-mentor should NOT exist (it was never installed)
	mentorAgent := filepath.Join(target, "agents", "developer-mentor.agent.md")
	if _, err := os.Stat(mentorAgent); !os.IsNotExist(err) {
		t.Errorf("developer-mentor should NOT be installed by update: %s", mentorAgent)
	}
}

// TestUpdateNothingInstalledSucceedsWithoutChanges verifies that running
// update against an empty directory succeeds without writing any files.
func TestUpdateNothingInstalledSucceedsWithoutChanges(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	updateCmd := newUpdateCommand(content)
	updateCmd.SetArgs([]string{
		"--target", target,
	})

	// Should succeed (not error)
	if err := updateCmd.Execute(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Target should remain empty (no files installed)
	entries, err := os.ReadDir(target)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty target dir, found %d entries", len(entries))
	}
}
