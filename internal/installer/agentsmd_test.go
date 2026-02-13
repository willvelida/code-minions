package installer

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestAgentsMDShouldSkip(t *testing.T) {
	handler := &AgentsMDHandler{Stdout: &bytes.Buffer{}}

	tests := []struct {
		path   string
		expect bool
	}{
		{"AGENTS.md", true},
		{"agents/AGENTS.md", true},
		{".github/agents/AGENTS.md", true},
		{".claude/agents/AGENTS.md", true},
		{"agents/my-agent.agent.md", false},
		{"skills/my-skill/SKILL.md", false},
		{"AGENTS.md.bak", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := handler.ShouldSkip(tt.path)
			if got != tt.expect {
				t.Errorf("ShouldSkip(%q) = %v, want %v", tt.path, got, tt.expect)
			}
		})
	}
}

func TestAgentsMDOnInstallCreatesNew(t *testing.T) {
	target := t.TempDir()

	handler := &AgentsMDHandler{
		Target: target,
		DryRun: false,
		Stdin:  bytes.NewBufferString(""),
		Stdout: &bytes.Buffer{},
	}

	data := []byte("# My Agents\n")
	action, err := handler.OnInstall("AGENTS.md", data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != "created" {
		t.Errorf("action = %q, want %q", action, "created")
	}

	// Verify the file was written
	got, err := os.ReadFile(filepath.Join(target, "AGENTS.md"))
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(got) != "# My Agents\n" {
		t.Errorf("file content = %q, want %q", string(got), "# My Agents\n")
	}
}

func TestAgentsMDOnInstallSkipsExisting(t *testing.T) {
	target := t.TempDir()

	// Pre-create an AGENTS.md with custom content
	original := []byte("# My Custom Agents\n")
	if err := os.WriteFile(filepath.Join(target, "AGENTS.md"), original, 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	handler := &AgentsMDHandler{
		Target: target,
		DryRun: false,
		Stdin:  bytes.NewBufferString(""),
		Stdout: &bytes.Buffer{},
	}

	action, err := handler.OnInstall("AGENTS.md", []byte("# Replacement\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != "skipped" {
		t.Errorf("action = %q, want %q", action, "skipped")
	}

	// Verify the original content was preserved
	got, err := os.ReadFile(filepath.Join(target, "AGENTS.md"))
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(got) != "# My Custom Agents\n" {
		t.Errorf("file was overwritten: got %q, want %q", string(got), "# My Custom Agents\n")
	}
}

func TestAgentsMDOnInstallDryRun(t *testing.T) {
	target := t.TempDir()

	handler := &AgentsMDHandler{
		Target: target,
		DryRun: true,
		Stdin:  bytes.NewBufferString(""),
		Stdout: &bytes.Buffer{},
	}

	action, err := handler.OnInstall("AGENTS.md", []byte("# Agents\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != "created" {
		t.Errorf("action = %q, want %q", action, "created")
	}

	// Verify nothing was written to disk
	if _, err := os.Stat(filepath.Join(target, "AGENTS.md")); !os.IsNotExist(err) {
		t.Error("expected AGENTS.md to NOT exist in dry-run mode")
	}
}

func TestAgentsMDOnInstallWithPathMapper(t *testing.T) {
	target := t.TempDir()

	handler := &AgentsMDHandler{
		Target: target,
		DryRun: false,
		Stdin:  bytes.NewBufferString(""),
		Stdout: &bytes.Buffer{},
	}

	// Simulate Copilot layout: agents/AGENTS.md → .github/agents/AGENTS.md
	action, err := handler.OnInstall(".github/agents/AGENTS.md", []byte("# Copilot Agents\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != "created" {
		t.Errorf("action = %q, want %q", action, "created")
	}

	// Verify the file was created at the mapped path
	got, err := os.ReadFile(filepath.Join(target, ".github", "agents", "AGENTS.md"))
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(got) != "# Copilot Agents\n" {
		t.Errorf("file content = %q, want %q", string(got), "# Copilot Agents\n")
	}
}

func TestAgentsMDOnUninstallConfirmRemove(t *testing.T) {
	target := t.TempDir()

	// Create the file to be removed
	if err := os.WriteFile(filepath.Join(target, "AGENTS.md"), []byte("# Agents\n"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	handler := &AgentsMDHandler{
		Target: target,
		DryRun: false,
		Stdin:  bytes.NewBufferString("y\n"),
		Stdout: &bytes.Buffer{},
	}

	action, err := handler.OnUninstall("AGENTS.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != "removed" {
		t.Errorf("action = %q, want %q", action, "removed")
	}

	// Verify the file was deleted
	if _, err := os.Stat(filepath.Join(target, "AGENTS.md")); !os.IsNotExist(err) {
		t.Error("expected AGENTS.md to be removed")
	}
}

func TestAgentsMDOnUninstallDeclineRemove(t *testing.T) {
	target := t.TempDir()

	if err := os.WriteFile(filepath.Join(target, "AGENTS.md"), []byte("# Agents\n"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	handler := &AgentsMDHandler{
		Target: target,
		DryRun: false,
		Stdin:  bytes.NewBufferString("n\n"),
		Stdout: &bytes.Buffer{},
	}

	action, err := handler.OnUninstall("AGENTS.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != "kept" {
		t.Errorf("action = %q, want %q", action, "kept")
	}

	// Verify the file still exists
	if _, err := os.Stat(filepath.Join(target, "AGENTS.md")); os.IsNotExist(err) {
		t.Error("expected AGENTS.md to still exist")
	}
}

func TestAgentsMDOnUninstallDefaultNo(t *testing.T) {
	target := t.TempDir()

	if err := os.WriteFile(filepath.Join(target, "AGENTS.md"), []byte("# Agents\n"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	handler := &AgentsMDHandler{
		Target: target,
		DryRun: false,
		Stdin:  bytes.NewBufferString("\n"),
		Stdout: &bytes.Buffer{},
	}

	action, err := handler.OnUninstall("AGENTS.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != "kept" {
		t.Errorf("action = %q, want %q", action, "kept")
	}

	// Verify the file still exists
	if _, err := os.Stat(filepath.Join(target, "AGENTS.md")); os.IsNotExist(err) {
		t.Error("expected AGENTS.md to still exist")
	}
}

func TestAgentsMDOnUninstallNotFound(t *testing.T) {
	target := t.TempDir()

	handler := &AgentsMDHandler{
		Target: target,
		DryRun: false,
		Stdin:  bytes.NewBufferString(""),
		Stdout: &bytes.Buffer{},
	}

	// No AGENTS.md exists — should return "kept" with no error and no prompt
	action, err := handler.OnUninstall("AGENTS.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != "kept" {
		t.Errorf("action = %q, want %q", action, "kept")
	}
}

func TestAgentsMDOnUninstallDryRun(t *testing.T) {
	target := t.TempDir()

	if err := os.WriteFile(filepath.Join(target, "AGENTS.md"), []byte("# Agents\n"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	handler := &AgentsMDHandler{
		Target: target,
		DryRun: true,
		Stdin:  bytes.NewBufferString("y\n"),
		Stdout: &bytes.Buffer{},
	}

	action, err := handler.OnUninstall("AGENTS.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != "kept" {
		t.Errorf("action = %q, want %q", action, "kept")
	}

	// Verify the file still exists
	if _, err := os.Stat(filepath.Join(target, "AGENTS.md")); os.IsNotExist(err) {
		t.Error("expected AGENTS.md to still exist in dry-run mode")
	}
}
