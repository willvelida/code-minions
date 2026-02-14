package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUninstallPackage(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	// Install first
	installCmd := newInstallCommand(content)
	installCmd.SetArgs([]string{
		"--package", "git-workflow",
		"--target", target,
	})
	if err := installCmd.Execute(); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Verify files exist after install
	agentFile := filepath.Join(target, "agents", "git-workflow.agent.md")
	if _, err := os.Stat(agentFile); os.IsNotExist(err) {
		t.Fatalf("expected file to exist after install: %s", agentFile)
	}

	// Uninstall
	uninstallCmd := newUninstallCommand(content)
	uninstallCmd.SetArgs([]string{
		"--package", "git-workflow",
		"--target", target,
	})
	if err := uninstallCmd.Execute(); err != nil {
		t.Fatalf("uninstall failed: %v", err)
	}

	// Verify files are removed
	if _, err := os.Stat(agentFile); !os.IsNotExist(err) {
		t.Errorf("expected file to be removed: %s", agentFile)
	}

	skillFile := filepath.Join(target, "skills", "git-workflow", "SKILL.md")
	if _, err := os.Stat(skillFile); !os.IsNotExist(err) {
		t.Errorf("expected file to be removed: %s", skillFile)
	}
}

func TestUninstallForCopilot(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	// Install with --for copilot
	installCmd := newInstallCommand(content)
	installCmd.SetArgs([]string{
		"--package", "git-workflow",
		"--for", "copilot",
		"--target", target,
	})
	if err := installCmd.Execute(); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Agent should be at Copilot path
	copilotAgent := filepath.Join(target, ".github", "agents", "git-workflow.agent.md")
	if _, err := os.Stat(copilotAgent); os.IsNotExist(err) {
		t.Fatalf("expected file at Copilot path: %s", copilotAgent)
	}

	// Uninstall with --for copilot
	uninstallCmd := newUninstallCommand(content)
	uninstallCmd.SetArgs([]string{
		"--package", "git-workflow",
		"--for", "copilot",
		"--target", target,
	})
	if err := uninstallCmd.Execute(); err != nil {
		t.Fatalf("uninstall failed: %v", err)
	}

	if _, err := os.Stat(copilotAgent); !os.IsNotExist(err) {
		t.Errorf("expected file to be removed: %s", copilotAgent)
	}
}

func TestUninstallForClaude(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	// Install with --for claude
	installCmd := newInstallCommand(content)
	installCmd.SetArgs([]string{
		"--package", "git-workflow",
		"--for", "claude",
		"--target", target,
	})
	if err := installCmd.Execute(); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	claudeAgent := filepath.Join(target, ".claude", "agents", "git-workflow.agent.md")
	if _, err := os.Stat(claudeAgent); os.IsNotExist(err) {
		t.Fatalf("expected file at Claude path: %s", claudeAgent)
	}

	// Uninstall with --for claude
	uninstallCmd := newUninstallCommand(content)
	uninstallCmd.SetArgs([]string{
		"--package", "git-workflow",
		"--for", "claude",
		"--target", target,
	})
	if err := uninstallCmd.Execute(); err != nil {
		t.Fatalf("uninstall failed: %v", err)
	}

	if _, err := os.Stat(claudeAgent); !os.IsNotExist(err) {
		t.Errorf("expected file to be removed: %s", claudeAgent)
	}
}

func TestUninstallForUnknownAssistantReturnsError(t *testing.T) {
	content := testContentFS()

	cmd := newUninstallCommand(content)
	cmd.SetArgs([]string{
		"--package", "git-workflow",
		"--for", "vscode",
	})
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for unknown assistant, got nil")
	}

	msg := err.Error()
	if !strings.Contains(msg, "vscode") {
		t.Errorf("error should mention the invalid name: got %q", msg)
	}
	if !strings.Contains(msg, "copilot") || !strings.Contains(msg, "claude") {
		t.Errorf("error should list valid assistants: got %q", msg)
	}
}

func TestUninstallDryRun(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	// Install first
	installCmd := newInstallCommand(content)
	installCmd.SetArgs([]string{
		"--package", "git-workflow",
		"--target", target,
	})
	if err := installCmd.Execute(); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Uninstall with --dry-run
	uninstallCmd := newUninstallCommand(content)
	uninstallCmd.SetArgs([]string{
		"--package", "git-workflow",
		"--dry-run",
		"--target", target,
	})
	if err := uninstallCmd.Execute(); err != nil {
		t.Fatalf("uninstall failed: %v", err)
	}

	// Files should still exist
	agentFile := filepath.Join(target, "agents", "git-workflow.agent.md")
	if _, err := os.Stat(agentFile); os.IsNotExist(err) {
		t.Errorf("file should still exist in dry-run: %s", agentFile)
	}
}

func TestUninstallInvalidPackage(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	uninstallCmd := newUninstallCommand(content)
	uninstallCmd.SetArgs([]string{
		"--package", "nonexistent",
		"--target", target,
	})

	uninstallCmd.SilenceErrors = true
	uninstallCmd.SilenceUsage = true

	err := uninstallCmd.Execute()
	if err == nil {
		t.Fatal("expected error for nonexistent package, got nil")
	}
}

func TestUninstallAllRequiresFor(t *testing.T) {
	content := testContentFS()

	uninstallCmd := newUninstallCommand(content)
	uninstallCmd.SetArgs([]string{})

	uninstallCmd.SilenceErrors = true
	uninstallCmd.SilenceUsage = true

	err := uninstallCmd.Execute()
	if err == nil {
		t.Fatal("expected error when no --package or --for is given")
	}
}

func TestUninstallAllWithFor(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	// Install everything with --for copilot
	installCmd := newInstallCommand(content)
	installCmd.SetArgs([]string{
		"--for", "copilot",
		"--target", target,
	})
	if err := installCmd.Execute(); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Uninstall everything with --for copilot
	uninstallCmd := newUninstallCommand(content)
	uninstallCmd.SetArgs([]string{
		"--for", "copilot",
		"--target", target,
	})
	if err := uninstallCmd.Execute(); err != nil {
		t.Fatalf("uninstall failed: %v", err)
	}

	// Verify agent files are removed from Copilot paths
	copilotAgent := filepath.Join(target, ".github", "agents", "git-workflow.agent.md")
	if _, err := os.Stat(copilotAgent); !os.IsNotExist(err) {
		t.Errorf("expected file to be removed: %s", copilotAgent)
	}

	mentorAgent := filepath.Join(target, ".github", "agents", "developer-mentor.agent.md")
	if _, err := os.Stat(mentorAgent); !os.IsNotExist(err) {
		t.Errorf("expected file to be removed: %s", mentorAgent)
	}
}
