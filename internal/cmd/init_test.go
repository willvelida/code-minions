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

func TestInitYesDefaultsToStandardTemplate(t *testing.T) {
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
	// --yes defaults to "standard" template: git-workflow, developer-mentor, raise-pull-requests
	if m.Template != "standard" {
		t.Errorf("Template: got %q, want %q", m.Template, "standard")
	}
	if len(m.Packages) != 3 {
		t.Errorf("expected 3 packages (standard template), got %d: %v", len(m.Packages), m.Packages)
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

func TestInitWithTemplateStandard(t *testing.T) {
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
	root.SetArgs([]string{"init", "--target", dir, "--template", "standard", "--assistant", "copilot", "--yes"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m, err := manifest.Load(manifest.DefaultPath(dir))
	if err != nil {
		t.Fatalf("failed to load manifest: %v", err)
	}
	if m.Template != "standard" {
		t.Errorf("Template: got %q, want %q", m.Template, "standard")
	}
	// standard template: git-workflow, developer-mentor, raise-pull-requests
	expected := []string{"git-workflow", "developer-mentor", "raise-pull-requests"}
	if len(m.Packages) != len(expected) {
		t.Fatalf("Packages: got %v, want %v", m.Packages, expected)
	}
	for i, pkg := range expected {
		if m.Packages[i] != pkg {
			t.Errorf("Packages[%d]: got %q, want %q", i, m.Packages[i], pkg)
		}
	}
}

func TestInitWithTemplateMinimal(t *testing.T) {
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
	root.SetArgs([]string{"init", "--target", dir, "--template", "minimal", "--assistant", "copilot", "--yes"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m, err := manifest.Load(manifest.DefaultPath(dir))
	if err != nil {
		t.Fatalf("failed to load manifest: %v", err)
	}
	if m.Template != "minimal" {
		t.Errorf("Template: got %q, want %q", m.Template, "minimal")
	}
	if len(m.Packages) != 1 || m.Packages[0] != "git-workflow" {
		t.Errorf("Packages: got %v, want [git-workflow]", m.Packages)
	}
}

func TestInitWithTemplateFullstack(t *testing.T) {
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
	root.SetArgs([]string{"init", "--target", dir, "--template", "fullstack", "--assistant", "copilot", "--yes"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m, err := manifest.Load(manifest.DefaultPath(dir))
	if err != nil {
		t.Fatalf("failed to load manifest: %v", err)
	}
	if m.Template != "fullstack" {
		t.Errorf("Template: got %q, want %q", m.Template, "fullstack")
	}
	// fullstack should include all packages from the test FS
	if len(m.Packages) != 2 {
		t.Errorf("expected 2 packages (all in test FS), got %d: %v", len(m.Packages), m.Packages)
	}
}

func TestInitWithInvalidTemplate(t *testing.T) {
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
	root.SetArgs([]string{"init", "--target", dir, "--template", "nonexistent", "--yes"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for unknown template, got nil")
	}
	if !strings.Contains(err.Error(), "unknown template") {
		t.Errorf("error should mention 'unknown template', got: %v", err)
	}
}

func TestInitTemplateAndPackagesMutuallyExclusive(t *testing.T) {
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
	root.SetArgs([]string{"init", "--target", dir, "--template", "standard", "--packages", "git-workflow", "--yes"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when --template and --packages are both set, got nil")
	}
}

func TestInitListTemplates(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	content := testContentFSForInit()

	var buf bytes.Buffer
	root := NewRootCommand(content)
	root.SetOut(&buf)
	root.SetArgs([]string{"init", "--list-templates"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Available templates:") {
		t.Error("output should contain 'Available templates:'")
	}
	for _, name := range []string{"minimal", "standard", "security", "fullstack", "docs"} {
		if !strings.Contains(output, name) {
			t.Errorf("output should list template %q", name)
		}
	}
	// Should not create a manifest
	// (no --target was needed since no manifest is created)
}

func TestInitListTemplatesJSON(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	content := testContentFSForInit()

	var buf bytes.Buffer
	root := NewRootCommand(content)
	root.SetOut(&buf)
	root.SetArgs([]string{"init", "--list-templates", "--json"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var templates []struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Packages    []string `json:"packages"`
	}
	if err := json.Unmarshal(buf.Bytes(), &templates); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v\nraw: %s", err, buf.String())
	}
	if len(templates) != 5 {
		t.Errorf("expected 5 templates in JSON, got %d", len(templates))
	}
}

func TestInitWithTemplateJSON(t *testing.T) {
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
	root.SetArgs([]string{"init", "--target", dir, "--template", "security", "--assistant", "copilot", "--json"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result initResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v\nraw: %s", err, buf.String())
	}
	if result.Template != "security" {
		t.Errorf("Template: got %q, want %q", result.Template, "security")
	}
	if result.Assistant != "copilot" {
		t.Errorf("Assistant: got %q, want %q", result.Assistant, "copilot")
	}
}

func TestInitWithPackagesFlagNoTemplate(t *testing.T) {
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
	root.SetArgs([]string{"init", "--target", dir, "--packages", "threat-modelling", "--assistant", "copilot", "--yes"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m, err := manifest.Load(manifest.DefaultPath(dir))
	if err != nil {
		t.Fatalf("failed to load manifest: %v", err)
	}
	// --packages skips template entirely
	if m.Template != "" {
		t.Errorf("Template should be empty when --packages is used, got %q", m.Template)
	}
	if len(m.Packages) != 1 || m.Packages[0] != "threat-modelling" {
		t.Errorf("Packages: got %v, want [threat-modelling]", m.Packages)
	}
}

func TestInitTemplateOutputShowsTemplateName(t *testing.T) {
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
	root.SetArgs([]string{"init", "--target", dir, "--template", "minimal", "--assistant", "copilot", "--yes"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "minimal") {
		t.Error("output should mention the template name")
	}
	if !strings.Contains(output, "Template:") {
		t.Error("output should contain 'Template:' line")
	}
}
