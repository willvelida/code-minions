package installer

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestInstructionsFileOnInstallCreatesNew(t *testing.T) {
	target := t.TempDir()

	handler := &InstructionsFileHandler{
		Target:   target,
		DryRun:   false,
		FileName: "CLAUDE.md",
		Stdin:    bytes.NewBufferString(""),
		Stdout:   &bytes.Buffer{},
	}

	data := []byte("# Project Instructions\n")
	action, err := handler.OnInstall(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != "created" {
		t.Errorf("action = %q, want %q", action, "created")
	}

	got, err := os.ReadFile(filepath.Join(target, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(got) != "# Project Instructions\n" {
		t.Errorf("file content = %q, want %q", string(got), "# Project Instructions\n")
	}
}

func TestInstructionsFileOnInstallSkipsExisting(t *testing.T) {
	target := t.TempDir()

	original := []byte("# My Custom CLAUDE.md\n")
	if err := os.WriteFile(filepath.Join(target, "CLAUDE.md"), original, 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	handler := &InstructionsFileHandler{
		Target:   target,
		DryRun:   false,
		FileName: "CLAUDE.md",
		Stdin:    bytes.NewBufferString(""),
		Stdout:   &bytes.Buffer{},
	}

	action, err := handler.OnInstall([]byte("# Replacement\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != "skipped" {
		t.Errorf("action = %q, want %q", action, "skipped")
	}

	got, err := os.ReadFile(filepath.Join(target, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(got) != "# My Custom CLAUDE.md\n" {
		t.Errorf("file was overwritten: got %q, want %q", string(got), "# My Custom CLAUDE.md\n")
	}
}

func TestInstructionsFileOnInstallDryRun(t *testing.T) {
	target := t.TempDir()

	handler := &InstructionsFileHandler{
		Target:   target,
		DryRun:   true,
		FileName: "CLAUDE.md",
		Stdin:    bytes.NewBufferString(""),
		Stdout:   &bytes.Buffer{},
	}

	action, err := handler.OnInstall([]byte("# Agents\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != "created" {
		t.Errorf("action = %q, want %q", action, "created")
	}

	if _, err := os.Stat(filepath.Join(target, "CLAUDE.md")); !os.IsNotExist(err) {
		t.Error("expected CLAUDE.md to NOT exist in dry-run mode")
	}
}

func TestInstructionsFileOnInstallCreatesSubdir(t *testing.T) {
	target := t.TempDir()

	handler := &InstructionsFileHandler{
		Target:   target,
		DryRun:   false,
		FileName: ".cursor/rules/instructions.mdc",
		Stdin:    bytes.NewBufferString(""),
		Stdout:   &bytes.Buffer{},
	}

	data := []byte("# Cursor Instructions\n")
	action, err := handler.OnInstall(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != "created" {
		t.Errorf("action = %q, want %q", action, "created")
	}

	got, err := os.ReadFile(filepath.Join(target, ".cursor", "rules", "instructions.mdc"))
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(got) != "# Cursor Instructions\n" {
		t.Errorf("file content = %q, want %q", string(got), "# Cursor Instructions\n")
	}
}

func TestInstructionsFileOnUninstallConfirmRemove(t *testing.T) {
	target := t.TempDir()

	if err := os.WriteFile(filepath.Join(target, "CLAUDE.md"), []byte("# Agents\n"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	handler := &InstructionsFileHandler{
		Target:   target,
		DryRun:   false,
		FileName: "CLAUDE.md",
		Stdin:    bytes.NewBufferString("y\n"),
		Stdout:   &bytes.Buffer{},
	}

	action, err := handler.OnUninstall()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != "removed" {
		t.Errorf("action = %q, want %q", action, "removed")
	}

	if _, err := os.Stat(filepath.Join(target, "CLAUDE.md")); !os.IsNotExist(err) {
		t.Error("expected CLAUDE.md to be removed")
	}
}

func TestInstructionsFileOnUninstallDeclineRemove(t *testing.T) {
	target := t.TempDir()

	if err := os.WriteFile(filepath.Join(target, "CLAUDE.md"), []byte("# Agents\n"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	handler := &InstructionsFileHandler{
		Target:   target,
		DryRun:   false,
		FileName: "CLAUDE.md",
		Stdin:    bytes.NewBufferString("n\n"),
		Stdout:   &bytes.Buffer{},
	}

	action, err := handler.OnUninstall()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != "kept" {
		t.Errorf("action = %q, want %q", action, "kept")
	}

	if _, err := os.Stat(filepath.Join(target, "CLAUDE.md")); os.IsNotExist(err) {
		t.Error("expected CLAUDE.md to still exist")
	}
}

func TestInstructionsFileOnUninstallNotExists(t *testing.T) {
	target := t.TempDir()

	handler := &InstructionsFileHandler{
		Target:   target,
		DryRun:   false,
		FileName: "CLAUDE.md",
		Stdin:    bytes.NewBufferString(""),
		Stdout:   &bytes.Buffer{},
	}

	action, err := handler.OnUninstall()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != "kept" {
		t.Errorf("action = %q, want %q", action, "kept")
	}
}

func TestInstructionsFileOnUninstallDryRun(t *testing.T) {
	target := t.TempDir()

	if err := os.WriteFile(filepath.Join(target, "CLAUDE.md"), []byte("# Agents\n"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	handler := &InstructionsFileHandler{
		Target:   target,
		DryRun:   true,
		FileName: "CLAUDE.md",
		Stdin:    bytes.NewBufferString(""),
		Stdout:   &bytes.Buffer{},
	}

	action, err := handler.OnUninstall()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != "kept" {
		t.Errorf("action = %q, want %q", action, "kept")
	}

	// File should still exist in dry-run
	if _, err := os.Stat(filepath.Join(target, "CLAUDE.md")); os.IsNotExist(err) {
		t.Error("expected CLAUDE.md to still exist in dry-run mode")
	}
}

func TestInstructionsFileOnUninstallPromptOutput(t *testing.T) {
	target := t.TempDir()

	if err := os.WriteFile(filepath.Join(target, "CLAUDE.md"), []byte("# Agents\n"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	stdout := &bytes.Buffer{}
	handler := &InstructionsFileHandler{
		Target:   target,
		DryRun:   false,
		FileName: "CLAUDE.md",
		Stdin:    bytes.NewBufferString("n\n"),
		Stdout:   stdout,
	}

	_, _ = handler.OnUninstall()

	output := stdout.String()
	if !bytes.Contains([]byte(output), []byte("CLAUDE.md")) {
		t.Errorf("prompt should mention the filename, got: %q", output)
	}
	if !bytes.Contains([]byte(output), []byte("remove")) {
		t.Errorf("prompt should ask about removal, got: %q", output)
	}
}
