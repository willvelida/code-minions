package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/fatih/color"
	"github.com/willvelida/code-minions/internal/manifest"
)

// testContentFSForInit builds a minimal embedded filesystem for init tests.
func testContentFSForInit() fstest.MapFS {
	return fstest.MapFS{
		"packages/threat-modelling/package.yaml": &fstest.MapFile{
			Data: []byte("name: threat-modelling\nversion: 0.1.0\ndescription: STRIDE-based threat modelling\n"),
		},
		"packages/threat-modelling/agents/threat-modelling.agent.md": &fstest.MapFile{
			Data: []byte("# Threat Modelling Agent"),
		},
		"packages/git-workflow/package.yaml": &fstest.MapFile{
			Data: []byte("name: git-workflow\nversion: 0.1.0\ndescription: Git workflow conventions\n"),
		},
		"packages/git-workflow/agents/git-workflow.agent.md": &fstest.MapFile{
			Data: []byte("# Git Agent"),
		},
	}
}

func TestInitCreatesManifest(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	// Force non-interactive mode
	origInteractive := isInteractiveFunc
	isInteractiveFunc = func() bool { return false }
	t.Cleanup(func() { isInteractiveFunc = origInteractive })

	dir := t.TempDir()
	content := testContentFSForInit()

	var buf bytes.Buffer
	root := NewRootCommand(content)
	root.SetOut(&buf)
	root.SetArgs([]string{"init", "--target", dir, "--yes", "--assistant", "copilot"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify manifest was created
	manifestPath := manifest.DefaultPath(dir)
	exists, err := manifest.Exists(manifestPath)
	if err != nil {
		t.Fatalf("Exists check failed: %v", err)
	}
	if !exists {
		t.Fatal("expected code-minions.yml to exist after init")
	}

	// Verify content
	m, err := manifest.Load(manifestPath)
	if err != nil {
		t.Fatalf("failed to load manifest: %v", err)
	}
	if m.Assistant != "copilot" {
		t.Errorf("Assistant: got %q, want %q", m.Assistant, "copilot")
	}
	if len(m.Packages) == 0 {
		t.Error("expected packages to be populated")
	}
}

func TestInitRefusesOverwrite(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	origInteractive := isInteractiveFunc
	isInteractiveFunc = func() bool { return false }
	t.Cleanup(func() { isInteractiveFunc = origInteractive })

	dir := t.TempDir()

	// Create an existing manifest
	manifestPath := manifest.DefaultPath(dir)
	if err := os.WriteFile(manifestPath, []byte("name: existing\n"), 0644); err != nil {
		t.Fatal(err)
	}

	content := testContentFSForInit()

	var buf bytes.Buffer
	root := NewRootCommand(content)
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"init", "--target", dir, "--yes", "--assistant", "copilot"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when manifest already exists, got nil")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error should mention 'already exists', got: %v", err)
	}
}

func TestInitForceOverwrites(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	origInteractive := isInteractiveFunc
	isInteractiveFunc = func() bool { return false }
	t.Cleanup(func() { isInteractiveFunc = origInteractive })

	dir := t.TempDir()

	// Create an existing manifest
	manifestPath := manifest.DefaultPath(dir)
	if err := os.WriteFile(manifestPath, []byte("name: old\n"), 0644); err != nil {
		t.Fatal(err)
	}

	content := testContentFSForInit()

	var buf bytes.Buffer
	root := NewRootCommand(content)
	root.SetOut(&buf)
	root.SetArgs([]string{"init", "--target", dir, "--yes", "--force", "--assistant", "copilot"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error with --force: %v", err)
	}

	// Verify the manifest was overwritten
	m, err := manifest.Load(manifestPath)
	if err != nil {
		t.Fatalf("failed to load manifest: %v", err)
	}
	if m.Assistant != "copilot" {
		t.Errorf("Assistant: got %q, want %q (manifest not overwritten?)", m.Assistant, "copilot")
	}
}

func TestInitJSON(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	origInteractive := isInteractiveFunc
	isInteractiveFunc = func() bool { return false }
	t.Cleanup(func() { isInteractiveFunc = origInteractive })

	dir := t.TempDir()
	content := testContentFSForInit()

	var buf bytes.Buffer
	root := NewRootCommand(content)
	root.SetOut(&buf)
	root.SetArgs([]string{"init", "--target", dir, "--assistant", "copilot", "--packages", "threat-modelling", "--json"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result initResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v\nraw: %s", err, buf.String())
	}
	if result.ManifestPath != "code-minions.yml" {
		t.Errorf("ManifestPath: got %q, want %q", result.ManifestPath, "code-minions.yml")
	}
	if result.Assistant != "copilot" {
		t.Errorf("Assistant: got %q, want %q", result.Assistant, "copilot")
	}
	if len(result.Packages) != 1 || result.Packages[0] != "threat-modelling" {
		t.Errorf("Packages: got %v, want [threat-modelling]", result.Packages)
	}
}

func TestInitDetectsAssistant(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	origInteractive := isInteractiveFunc
	isInteractiveFunc = func() bool { return false }
	t.Cleanup(func() { isInteractiveFunc = origInteractive })

	dir := t.TempDir()

	// Create a Claude marker
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# Claude\n"), 0644); err != nil {
		t.Fatal(err)
	}

	content := testContentFSForInit()

	var buf bytes.Buffer
	root := NewRootCommand(content)
	root.SetOut(&buf)
	root.SetArgs([]string{"init", "--target", dir, "--yes"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The manifest should have claude as the assistant (auto-detected)
	m, err := manifest.Load(manifest.DefaultPath(dir))
	if err != nil {
		t.Fatalf("failed to load manifest: %v", err)
	}
	if m.Assistant != "claude" {
		t.Errorf("Assistant: got %q, want %q (should detect Claude)", m.Assistant, "claude")
	}
}

func TestInitPackageFlag(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	origInteractive := isInteractiveFunc
	isInteractiveFunc = func() bool { return false }
	t.Cleanup(func() { isInteractiveFunc = origInteractive })

	dir := t.TempDir()
	content := testContentFSForInit()

	var buf bytes.Buffer
	root := NewRootCommand(content)
	root.SetOut(&buf)
	root.SetArgs([]string{"init", "--target", dir, "--yes", "--assistant", "copilot", "--packages", "threat-modelling"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m, err := manifest.Load(manifest.DefaultPath(dir))
	if err != nil {
		t.Fatalf("failed to load manifest: %v", err)
	}
	if len(m.Packages) != 1 || m.Packages[0] != "threat-modelling" {
		t.Errorf("Packages: got %v, want [threat-modelling]", m.Packages)
	}
}

func TestInitInvalidAssistant(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	origInteractive := isInteractiveFunc
	isInteractiveFunc = func() bool { return false }
	t.Cleanup(func() { isInteractiveFunc = origInteractive })

	dir := t.TempDir()
	content := testContentFSForInit()

	var buf bytes.Buffer
	root := NewRootCommand(content)
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"init", "--target", dir, "--yes", "--assistant", "unknown"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for unknown assistant, got nil")
	}
	if !strings.Contains(err.Error(), "unknown assistant") {
		t.Errorf("error should mention 'unknown assistant', got: %v", err)
	}
}

func TestInitInvalidPackage(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	origInteractive := isInteractiveFunc
	isInteractiveFunc = func() bool { return false }
	t.Cleanup(func() { isInteractiveFunc = origInteractive })

	dir := t.TempDir()
	content := testContentFSForInit()

	var buf bytes.Buffer
	root := NewRootCommand(content)
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"init", "--target", dir, "--yes", "--assistant", "copilot", "--packages", "nonexistent"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for unknown package, got nil")
	}
	if !strings.Contains(err.Error(), "unknown package") {
		t.Errorf("error should mention 'unknown package', got: %v", err)
	}
}

func TestInitYesDefaultsToAllPackages(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	origInteractive := isInteractiveFunc
	isInteractiveFunc = func() bool { return false }
	t.Cleanup(func() { isInteractiveFunc = origInteractive })

	dir := t.TempDir()
	content := testContentFSForInit()

	var buf bytes.Buffer
	root := NewRootCommand(content)
	root.SetOut(&buf)
	root.SetArgs([]string{"init", "--target", dir, "--yes", "--assistant", "copilot"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m, err := manifest.Load(manifest.DefaultPath(dir))
	if err != nil {
		t.Fatalf("failed to load manifest: %v", err)
	}
	// Should include all packages from the test FS
	if len(m.Packages) != 2 {
		t.Errorf("expected 2 packages (all), got %d: %v", len(m.Packages), m.Packages)
	}
}

func TestInitManifestHasHeaderComment(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	origInteractive := isInteractiveFunc
	isInteractiveFunc = func() bool { return false }
	t.Cleanup(func() { isInteractiveFunc = origInteractive })

	dir := t.TempDir()
	content := testContentFSForInit()

	var buf bytes.Buffer
	root := NewRootCommand(content)
	root.SetOut(&buf)
	root.SetArgs([]string{"init", "--target", dir, "--yes", "--assistant", "copilot"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(manifest.DefaultPath(dir))
	if err != nil {
		t.Fatalf("failed to read manifest: %v", err)
	}
	if !strings.HasPrefix(string(data), "# Generated by code-minions init\n") {
		t.Errorf("manifest should start with header comment, got:\n%s", string(data))
	}
}

func TestInitProjectNameFromDirectory(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	origInteractive := isInteractiveFunc
	isInteractiveFunc = func() bool { return false }
	t.Cleanup(func() { isInteractiveFunc = origInteractive })

	dir := t.TempDir()
	content := testContentFSForInit()

	var buf bytes.Buffer
	root := NewRootCommand(content)
	root.SetOut(&buf)
	root.SetArgs([]string{"init", "--target", dir, "--yes", "--assistant", "copilot"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m, err := manifest.Load(manifest.DefaultPath(dir))
	if err != nil {
		t.Fatalf("failed to load manifest: %v", err)
	}

	expectedName := filepath.Base(dir)
	if m.Name != expectedName {
		t.Errorf("Name: got %q, want %q (should match directory name)", m.Name, expectedName)
	}
}
