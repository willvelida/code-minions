package installer

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"runtime"
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

// ---------- AgentsMDHandler error branch tests ----------

// TestAgentsMDOnInstallWriteFailure verifies that OnInstall returns an error
// when os.WriteFile fails. We make the target directory read-only so
// WriteFile can't create the file. This only works on Unix (directory
// permissions don't prevent file creation on Windows).
func TestAgentsMDOnInstallWriteFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("write-failure via read-only directory is not supported on Windows")
	}

	target := t.TempDir()

	// Make target read-only: Stat returns NotExist for the file,
	// MkdirAll is a no-op (target already exists), but WriteFile
	// cannot create a file in a read-only directory.
	if err := os.Chmod(target, 0555); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(target, 0755) })

	handler := &AgentsMDHandler{
		Target: target,
		Stdin:  &bytes.Buffer{},
		Stdout: &bytes.Buffer{},
	}

	_, err := handler.OnInstall("AGENTS.md", []byte("# Agents\n"))
	if err == nil {
		t.Fatal("expected write error, got nil")
	}
}

// TestAgentsMDOnInstallMkdirAllFailure verifies that OnInstall returns an error
// when the parent directory can't be created.
func TestAgentsMDOnInstallMkdirAllFailure(t *testing.T) {
	target := t.TempDir()

	// Create a file where MkdirAll expects to create a directory.
	if err := os.WriteFile(filepath.Join(target, "deep"), []byte("blocker"), 0644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	handler := &AgentsMDHandler{
		Target: target,
		Stdin:  &bytes.Buffer{},
		Stdout: &bytes.Buffer{},
	}

	_, err := handler.OnInstall("deep/nested/AGENTS.md", []byte("# Agents\n"))
	if err == nil {
		t.Fatal("expected MkdirAll error, got nil")
	}
}

// errReader is an io.Reader that always returns an error.
type errReader struct{ err error }

func (r *errReader) Read([]byte) (int, error) { return 0, r.err }

// TestAgentsMDOnUninstallScannerError verifies that OnUninstall returns an error
// when reading stdin fails.
func TestAgentsMDOnUninstallScannerError(t *testing.T) {
	target := t.TempDir()

	if err := os.WriteFile(filepath.Join(target, "AGENTS.md"), []byte("# Agents\n"), 0644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	handler := &AgentsMDHandler{
		Target: target,
		Stdin:  &errReader{err: errors.New("stdin broken")},
		Stdout: &bytes.Buffer{},
	}

	_, err := handler.OnUninstall("AGENTS.md")
	if err == nil {
		t.Fatal("expected scanner error, got nil")
	}
}

// TestAgentsMDOnUninstallRemoveFailure verifies that OnUninstall returns an
// error when os.Remove fails (e.g. path is a non-empty directory).
func TestAgentsMDOnUninstallRemoveFailure(t *testing.T) {
	target := t.TempDir()

	// Create a non-empty directory at the AGENTS.md path so os.Remove fails.
	agentsDir := filepath.Join(target, "AGENTS.md")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(agentsDir, "blocker"), []byte("x"), 0644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	handler := &AgentsMDHandler{
		Target: target,
		Stdin:  bytes.NewBufferString("y\n"),
		Stdout: &bytes.Buffer{},
	}

	_, err := handler.OnUninstall("AGENTS.md")
	if err == nil {
		t.Fatal("expected remove error, got nil")
	}
}
