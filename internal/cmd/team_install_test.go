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
	"github.com/willvelida/code-minions/internal/installer"
)

// testContentFSForTeamInstall builds a minimal embedded filesystem
// suitable for team install tests. It includes packages with agents
// and skills plus package.yaml files so the registry can resolve them.
func testContentFSForTeamInstall() fstest.MapFS {
	return fstest.MapFS{
		"packages/git-workflow/package.yaml": &fstest.MapFile{
			Data: []byte("name: git-workflow\nversion: \"0.1.0\"\ndescription: Git workflow conventions\n"),
		},
		"packages/git-workflow/agents/git-workflow.agent.md": &fstest.MapFile{
			Data: []byte("# Git Agent"),
		},
		"packages/git-workflow/skills/git-workflow/SKILL.md": &fstest.MapFile{
			Data: []byte("# Git"),
		},
		"packages/developer-mentor/package.yaml": &fstest.MapFile{
			Data: []byte("name: developer-mentor\nversion: \"0.1.0\"\ndescription: Developer mentoring skill\n"),
		},
		"packages/developer-mentor/agents/developer-mentor.agent.md": &fstest.MapFile{
			Data: []byte("# Mentor Agent"),
		},
		"packages/developer-mentor/skills/developer-mentor/SKILL.md": &fstest.MapFile{
			Data: []byte("# Mentor"),
		},
		"packages/threat-modelling/package.yaml": &fstest.MapFile{
			Data: []byte("name: threat-modelling\nversion: \"0.1.0\"\ndescription: STRIDE-based threat modelling\n"),
		},
		"packages/threat-modelling/agents/threat-modelling.agent.md": &fstest.MapFile{
			Data: []byte("# Threat Agent"),
		},
	}
}

// writeTeamFile writes a team.yaml file at the given path with the
// specified content.
func writeTeamFile(t *testing.T, dir, content string) string {
	t.Helper()
	path := filepath.Join(dir, "team.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write team.yaml: %v", err)
	}
	return path
}

// minimalTeamYAML returns a valid team.yaml string with one persona
// referencing git-workflow and developer-mentor.
const minimalTeamYAML = `name: test-team
description: Test team configuration
personas:
  - name: default
    packages:
      - name: git-workflow
      - name: developer-mentor
config:
  default_assistant: copilot
`

// multiPersonaTeamYAML has two personas referencing different packages.
const multiPersonaTeamYAML = `name: multi-team
description: Multi persona team
personas:
  - name: front-end
    packages:
      - name: git-workflow
  - name: back-end
    packages:
      - name: developer-mentor
config:
  default_assistant: copilot
`

// ---------------------------------------------------------------------------
// Basic team install
// ---------------------------------------------------------------------------

func TestTeamInstallBasic(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	dir := t.TempDir()
	writeTeamFile(t, dir, minimalTeamYAML)

	content := testContentFSForTeamInstall()

	var buf bytes.Buffer
	root := NewRootCommand(content)
	root.SetOut(&buf)
	root.SetArgs([]string{
		"team", "install",
		"--for", "copilot",
		"--file", filepath.Join(dir, "team.yaml"),
		"--target", dir,
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Should contain success message
	if !strings.Contains(output, "Installed team") {
		t.Errorf("expected success message, got:\n%s", output)
	}
	if !strings.Contains(output, "test-team") {
		t.Errorf("expected team name in output, got:\n%s", output)
	}

	// Agent files should exist (Copilot → .github/agents/)
	agentFile := filepath.Join(dir, ".github", "agents", "git-workflow.agent.md")
	if _, err := os.Stat(agentFile); os.IsNotExist(err) {
		t.Errorf("expected agent file at %s", agentFile)
	}

	mentorAgent := filepath.Join(dir, ".github", "agents", "developer-mentor.agent.md")
	if _, err := os.Stat(mentorAgent); os.IsNotExist(err) {
		t.Errorf("expected mentor agent at %s", mentorAgent)
	}

	// Skills should exist
	skillFile := filepath.Join(dir, "skills", "git-workflow", "SKILL.md")
	if _, err := os.Stat(skillFile); os.IsNotExist(err) {
		t.Errorf("expected skill at %s", skillFile)
	}
}

// ---------------------------------------------------------------------------
// Default assistant from team config
// ---------------------------------------------------------------------------

func TestTeamInstallUsesDefaultAssistant(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	dir := t.TempDir()
	writeTeamFile(t, dir, minimalTeamYAML) // has default_assistant: copilot

	content := testContentFSForTeamInstall()

	var buf bytes.Buffer
	root := NewRootCommand(content)
	root.SetOut(&buf)
	// No --for flag: should use default_assistant from team config
	root.SetArgs([]string{
		"team", "install",
		"--file", filepath.Join(dir, "team.yaml"),
		"--target", dir,
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should succeed and mention copilot in output
	output := buf.String()
	if !strings.Contains(output, "copilot") {
		t.Errorf("expected 'copilot' in output, got:\n%s", output)
	}
}

// ---------------------------------------------------------------------------
// --for required when no default
// ---------------------------------------------------------------------------

func TestTeamInstallRequiresFor(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	dir := t.TempDir()
	teamYAML := `name: no-default
description: Team without default assistant
personas:
  - name: default
    packages:
      - name: git-workflow
`
	writeTeamFile(t, dir, teamYAML)

	content := testContentFSForTeamInstall()

	var buf bytes.Buffer
	root := NewRootCommand(content)
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{
		"team", "install",
		"--file", filepath.Join(dir, "team.yaml"),
		"--target", dir,
	})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when --for is missing and no default_assistant")
	}
	if !strings.Contains(err.Error(), "--for is required") {
		t.Errorf("error %q should mention '--for is required'", err)
	}
}

// ---------------------------------------------------------------------------
// Missing team file
// ---------------------------------------------------------------------------

func TestTeamInstallMissingFile(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	dir := t.TempDir()
	content := testContentFSForTeamInstall()

	var buf bytes.Buffer
	root := NewRootCommand(content)
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{
		"team", "install",
		"--for", "copilot",
		"--file", filepath.Join(dir, "team.yaml"),
		"--target", dir,
	})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for missing team.yaml")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error %q should mention 'not found'", err)
	}
	if !strings.Contains(err.Error(), "team init") {
		t.Errorf("error %q should suggest 'team init'", err)
	}
}

// ---------------------------------------------------------------------------
// Invalid team configuration
// ---------------------------------------------------------------------------

func TestTeamInstallInvalidTeam(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	dir := t.TempDir()
	// Missing required description
	teamYAML := `name: bad-team
personas:
  - name: default
    packages:
      - name: git-workflow
`
	writeTeamFile(t, dir, teamYAML)

	content := testContentFSForTeamInstall()

	var buf bytes.Buffer
	root := NewRootCommand(content)
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{
		"team", "install",
		"--for", "copilot",
		"--file", filepath.Join(dir, "team.yaml"),
		"--target", dir,
	})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for invalid team configuration")
	}
	if !strings.Contains(err.Error(), "invalid team") {
		t.Errorf("error %q should mention 'invalid team'", err)
	}
}

// ---------------------------------------------------------------------------
// Empty personas
// ---------------------------------------------------------------------------

func TestTeamInstallEmptyPersonas(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	dir := t.TempDir()
	teamYAML := `name: empty-team
description: Team with no personas
personas: []
`
	writeTeamFile(t, dir, teamYAML)

	content := testContentFSForTeamInstall()

	var buf bytes.Buffer
	root := NewRootCommand(content)
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{
		"team", "install",
		"--for", "copilot",
		"--file", filepath.Join(dir, "team.yaml"),
		"--target", dir,
	})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for team with no personas")
	}
	// ValidateTeam should reject empty personas
	if !strings.Contains(err.Error(), "at least one persona") {
		t.Errorf("error %q should mention 'at least one persona'", err)
	}
}

// ---------------------------------------------------------------------------
// Dry run
// ---------------------------------------------------------------------------

func TestTeamInstallDryRun(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	dir := t.TempDir()
	writeTeamFile(t, dir, minimalTeamYAML)

	content := testContentFSForTeamInstall()

	var buf bytes.Buffer
	root := NewRootCommand(content)
	root.SetOut(&buf)
	root.SetArgs([]string{
		"team", "install",
		"--for", "copilot",
		"--file", filepath.Join(dir, "team.yaml"),
		"--target", dir,
		"--dry-run",
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Should show dry-run banner
	if !strings.Contains(output, "Dry run") {
		t.Errorf("expected 'Dry run' banner, got:\n%s", output)
	}

	// Should say "would copy" not "copied"
	if !strings.Contains(output, "would copy") {
		t.Errorf("expected 'would copy' in dry-run output, got:\n%s", output)
	}

	// No agent files should be written
	agentDir := filepath.Join(dir, ".github", "agents")
	if _, err := os.Stat(agentDir); !os.IsNotExist(err) {
		t.Errorf("dry-run should not create agent directory %s", agentDir)
	}

	// No manifest should be written
	manifestPath := filepath.Join(dir, ".code-minions", "installed.json")
	if _, err := os.Stat(manifestPath); !os.IsNotExist(err) {
		t.Errorf("dry-run should not create manifest at %s", manifestPath)
	}
}

// ---------------------------------------------------------------------------
// Force overwrite
// ---------------------------------------------------------------------------

func TestTeamInstallForce(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	dir := t.TempDir()
	writeTeamFile(t, dir, minimalTeamYAML)

	content := testContentFSForTeamInstall()

	// First install
	root1 := NewRootCommand(content)
	root1.SetOut(&bytes.Buffer{})
	root1.SetArgs([]string{
		"team", "install",
		"--for", "copilot",
		"--file", filepath.Join(dir, "team.yaml"),
		"--target", dir,
	})
	if err := root1.Execute(); err != nil {
		t.Fatalf("first install failed: %v", err)
	}

	// Second install with --force
	var buf bytes.Buffer
	root2 := NewRootCommand(content)
	root2.SetOut(&buf)
	root2.SetArgs([]string{
		"team", "install",
		"--for", "copilot",
		"--file", filepath.Join(dir, "team.yaml"),
		"--target", dir,
		"--force",
	})
	if err := root2.Execute(); err != nil {
		t.Fatalf("force install failed: %v", err)
	}

	output := buf.String()
	// Should succeed and show copied (not skipped) because --force
	if !strings.Contains(output, "copied") {
		t.Errorf("expected 'copied' in force output, got:\n%s", output)
	}
}

// ---------------------------------------------------------------------------
// Persona filter
// ---------------------------------------------------------------------------

func TestTeamInstallPersonaFilter(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	dir := t.TempDir()
	writeTeamFile(t, dir, multiPersonaTeamYAML)

	content := testContentFSForTeamInstall()

	var buf bytes.Buffer
	root := NewRootCommand(content)
	root.SetOut(&buf)
	root.SetArgs([]string{
		"team", "install",
		"--for", "copilot",
		"--file", filepath.Join(dir, "team.yaml"),
		"--target", dir,
		"--persona", "front-end",
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Should install git-workflow (from front-end persona)
	if !strings.Contains(output, "git-workflow") {
		t.Errorf("expected 'git-workflow' in output, got:\n%s", output)
	}

	// Should NOT mention the back-end persona packages
	if strings.Contains(output, "developer-mentor") {
		t.Errorf("expected NO 'developer-mentor' when filtering to front-end, got:\n%s", output)
	}

	// Agent file for git-workflow should exist
	agentFile := filepath.Join(dir, ".github", "agents", "git-workflow.agent.md")
	if _, err := os.Stat(agentFile); os.IsNotExist(err) {
		t.Errorf("expected git-workflow agent at %s", agentFile)
	}

	// Mentor agent should NOT exist
	mentorAgent := filepath.Join(dir, ".github", "agents", "developer-mentor.agent.md")
	if _, err := os.Stat(mentorAgent); !os.IsNotExist(err) {
		t.Errorf("mentor agent should NOT exist when filtering to front-end: %s", mentorAgent)
	}
}

// ---------------------------------------------------------------------------
// Persona filter not found
// ---------------------------------------------------------------------------

func TestTeamInstallPersonaFilterNotFound(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	dir := t.TempDir()
	writeTeamFile(t, dir, multiPersonaTeamYAML)

	content := testContentFSForTeamInstall()

	var buf bytes.Buffer
	root := NewRootCommand(content)
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{
		"team", "install",
		"--for", "copilot",
		"--file", filepath.Join(dir, "team.yaml"),
		"--target", dir,
		"--persona", "nonexistent",
	})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for nonexistent persona filter")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error %q should mention 'not found'", err)
	}
	// Should list available personas
	if !strings.Contains(err.Error(), "front-end") {
		t.Errorf("error %q should list available persona names", err)
	}
}

// ---------------------------------------------------------------------------
// Multiple personas
// ---------------------------------------------------------------------------

func TestTeamInstallMultiplePersonas(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	dir := t.TempDir()
	writeTeamFile(t, dir, multiPersonaTeamYAML)

	content := testContentFSForTeamInstall()

	var buf bytes.Buffer
	root := NewRootCommand(content)
	root.SetOut(&buf)
	root.SetArgs([]string{
		"team", "install",
		"--for", "copilot",
		"--file", filepath.Join(dir, "team.yaml"),
		"--target", dir,
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Both personas should appear
	if !strings.Contains(output, "front-end") {
		t.Errorf("expected 'front-end' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "back-end") {
		t.Errorf("expected 'back-end' in output, got:\n%s", output)
	}

	// Both packages should be installed
	gitAgent := filepath.Join(dir, ".github", "agents", "git-workflow.agent.md")
	if _, err := os.Stat(gitAgent); os.IsNotExist(err) {
		t.Errorf("expected git-workflow agent at %s", gitAgent)
	}
	mentorAgent := filepath.Join(dir, ".github", "agents", "developer-mentor.agent.md")
	if _, err := os.Stat(mentorAgent); os.IsNotExist(err) {
		t.Errorf("expected developer-mentor agent at %s", mentorAgent)
	}
}

// ---------------------------------------------------------------------------
// Custom file path
// ---------------------------------------------------------------------------

func TestTeamInstallCustomFile(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	dir := t.TempDir()
	customPath := filepath.Join(dir, "config", "my-team.yaml")
	if err := os.MkdirAll(filepath.Dir(customPath), 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	if err := os.WriteFile(customPath, []byte(minimalTeamYAML), 0644); err != nil {
		t.Fatalf("failed to write team file: %v", err)
	}

	content := testContentFSForTeamInstall()

	var buf bytes.Buffer
	root := NewRootCommand(content)
	root.SetOut(&buf)
	root.SetArgs([]string{
		"team", "install",
		"--for", "copilot",
		"--file", customPath,
		"--target", dir,
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "Installed team") {
		t.Errorf("expected success message, got:\n%s", buf.String())
	}
}

// ---------------------------------------------------------------------------
// JSON output
// ---------------------------------------------------------------------------

func TestTeamInstallJSON(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	dir := t.TempDir()
	writeTeamFile(t, dir, minimalTeamYAML)

	content := testContentFSForTeamInstall()

	var buf bytes.Buffer
	root := NewRootCommand(content)
	root.SetOut(&buf)
	root.SetArgs([]string{
		"team", "install",
		"--for", "copilot",
		"--file", filepath.Join(dir, "team.yaml"),
		"--target", dir,
		"--json",
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result teamInstallResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v\nOutput: %s", err, buf.String())
	}

	if result.Team != "test-team" {
		t.Errorf("team = %q, want test-team", result.Team)
	}
	if result.Assistant != "copilot" {
		t.Errorf("assistant = %q, want copilot", result.Assistant)
	}
	if len(result.Personas) != 1 {
		t.Fatalf("personas = %d, want 1", len(result.Personas))
	}
	if result.Personas[0].Name != "default" {
		t.Errorf("persona[0].name = %q, want default", result.Personas[0].Name)
	}
	if len(result.Personas[0].Packages) != 2 {
		t.Errorf("persona[0].packages = %d, want 2", len(result.Personas[0].Packages))
	}
	if result.Summary.Errors != 0 {
		t.Errorf("summary.errors = %d, want 0", result.Summary.Errors)
	}
	if result.Summary.Copied == 0 {
		t.Error("expected at least some files copied")
	}
	if result.Summary.Personas != 1 {
		t.Errorf("summary.personas = %d, want 1", result.Summary.Personas)
	}
	if result.Summary.Packages != 2 {
		t.Errorf("summary.packages = %d, want 2", result.Summary.Packages)
	}
}

// ---------------------------------------------------------------------------
// JSON output with default assistant
// ---------------------------------------------------------------------------

func TestTeamInstallJSONDefaultAssistant(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	dir := t.TempDir()
	writeTeamFile(t, dir, minimalTeamYAML)

	content := testContentFSForTeamInstall()

	var buf bytes.Buffer
	root := NewRootCommand(content)
	root.SetOut(&buf)
	root.SetArgs([]string{
		"team", "install",
		"--file", filepath.Join(dir, "team.yaml"),
		"--target", dir,
		"--json",
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result teamInstallResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v\nOutput: %s", err, buf.String())
	}

	// When --for is omitted, the JSON should still show the assistant
	if result.Assistant != "copilot" {
		t.Errorf("assistant = %q, want copilot (from default_assistant)", result.Assistant)
	}
}

// ---------------------------------------------------------------------------
// Manifest recording
// ---------------------------------------------------------------------------

func TestTeamInstallManifest(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	dir := t.TempDir()
	writeTeamFile(t, dir, minimalTeamYAML)

	content := testContentFSForTeamInstall()

	var buf bytes.Buffer
	root := NewRootCommand(content)
	root.SetOut(&buf)
	root.SetArgs([]string{
		"team", "install",
		"--for", "copilot",
		"--file", filepath.Join(dir, "team.yaml"),
		"--target", dir,
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Manifest should exist
	manifestPath := filepath.Join(dir, ".code-minions", "installed.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("expected manifest at %s: %v", manifestPath, err)
	}

	// Verify manifest structure
	var manifest struct {
		Assistant string `json:"assistant"`
		Teams     []struct {
			Name     string   `json:"name"`
			Source   string   `json:"source"`
			Personas []string `json:"personas"`
		} `json:"teams"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("invalid manifest JSON: %v", err)
	}

	// assistant is stored at the manifest level, not per-team
	if manifest.Assistant != "copilot" {
		t.Errorf("manifest assistant = %q, want copilot", manifest.Assistant)
	}
	if len(manifest.Teams) != 1 {
		t.Fatalf("expected 1 team in manifest, got %d", len(manifest.Teams))
	}
	if manifest.Teams[0].Name != "test-team" {
		t.Errorf("team name = %q, want test-team", manifest.Teams[0].Name)
	}
}

// ---------------------------------------------------------------------------
// Quiet output
// ---------------------------------------------------------------------------

func TestTeamInstallQuiet(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	dir := t.TempDir()
	writeTeamFile(t, dir, minimalTeamYAML)

	content := testContentFSForTeamInstall()

	var buf bytes.Buffer
	root := NewRootCommand(content)
	root.SetOut(&buf)
	root.SetArgs([]string{
		"team", "install",
		"--for", "copilot",
		"--file", filepath.Join(dir, "team.yaml"),
		"--target", dir,
		"--quiet",
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Quiet mode should produce no stdout output
	if buf.String() != "" {
		t.Errorf("quiet mode should produce no output, got:\n%s", buf.String())
	}

	// Files should still be installed
	agentFile := filepath.Join(dir, ".github", "agents", "git-workflow.agent.md")
	if _, err := os.Stat(agentFile); os.IsNotExist(err) {
		t.Errorf("expected agent file at %s", agentFile)
	}
}

// ---------------------------------------------------------------------------
// Verbose output
// ---------------------------------------------------------------------------

func TestTeamInstallVerbose(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	dir := t.TempDir()
	writeTeamFile(t, dir, minimalTeamYAML)

	content := testContentFSForTeamInstall()

	var buf bytes.Buffer
	root := NewRootCommand(content)
	root.SetOut(&buf)
	root.SetArgs([]string{
		"team", "install",
		"--for", "copilot",
		"--file", filepath.Join(dir, "team.yaml"),
		"--target", dir,
		"--verbose",
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Verbose should include loaded and target assistant messages
	if !strings.Contains(output, "loaded team") {
		t.Errorf("expected verbose 'loaded team' message, got:\n%s", output)
	}
	if !strings.Contains(output, "target assistant") {
		t.Errorf("expected verbose 'target assistant' message, got:\n%s", output)
	}
}

// ---------------------------------------------------------------------------
// Install for Claude
// ---------------------------------------------------------------------------

func TestTeamInstallClaude(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	dir := t.TempDir()
	writeTeamFile(t, dir, minimalTeamYAML)

	content := testContentFSForTeamInstall()

	var buf bytes.Buffer
	root := NewRootCommand(content)
	root.SetOut(&buf)
	root.SetArgs([]string{
		"team", "install",
		"--for", "claude",
		"--file", filepath.Join(dir, "team.yaml"),
		"--target", dir,
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Claude places agents in .claude/agents/
	claudeAgent := filepath.Join(dir, ".claude", "agents", "git-workflow.agent.md")
	if _, err := os.Stat(claudeAgent); os.IsNotExist(err) {
		t.Errorf("expected agent at Claude path: %s", claudeAgent)
	}
}

// ---------------------------------------------------------------------------
// Second install skips existing files
// ---------------------------------------------------------------------------

func TestTeamInstallSkipsExisting(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	dir := t.TempDir()
	writeTeamFile(t, dir, minimalTeamYAML)

	content := testContentFSForTeamInstall()

	// First install
	root1 := NewRootCommand(content)
	root1.SetOut(&bytes.Buffer{})
	root1.SetArgs([]string{
		"team", "install",
		"--for", "copilot",
		"--file", filepath.Join(dir, "team.yaml"),
		"--target", dir,
	})
	if err := root1.Execute(); err != nil {
		t.Fatalf("first install failed: %v", err)
	}

	// Second install without --force
	// The grouping artefact (e.g. default.agent.md) already exists
	// and will cause a persona-level error; package files should be skipped.
	var buf bytes.Buffer
	root2 := NewRootCommand(content)
	root2.SetOut(&buf)
	root2.SetErr(&buf)
	root2.SetArgs([]string{
		"team", "install",
		"--for", "copilot",
		"--file", filepath.Join(dir, "team.yaml"),
		"--target", dir,
	})
	_ = root2.Execute() // may return error from grouping artefact conflict

	output := buf.String()
	if !strings.Contains(output, "skipped") && !strings.Contains(output, "already exists") {
		t.Errorf("expected 'skipped' or 'already exists' in second install output, got:\n%s", output)
	}
}

// ---------------------------------------------------------------------------
// Team with instructions
// ---------------------------------------------------------------------------

func TestTeamInstallWithInstructions(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	dir := t.TempDir()
	teamYAML := `name: instr-team
description: Team with instructions
personas:
  - name: default
    packages:
      - name: git-workflow
config:
  default_assistant: copilot
instructions: |
  Always follow the team coding standards.
  Use conventional commits.
`
	writeTeamFile(t, dir, teamYAML)

	content := testContentFSForTeamInstall()

	var buf bytes.Buffer
	root := NewRootCommand(content)
	root.SetOut(&buf)
	root.SetArgs([]string{
		"team", "install",
		"--for", "copilot",
		"--file", filepath.Join(dir, "team.yaml"),
		"--target", dir,
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that the instructions file was created/modified
	// Copilot instructions path is typically .github/copilot-instructions.md
	instrPath := filepath.Join(dir, ".github", "copilot-instructions.md")
	data, err := os.ReadFile(instrPath)
	if err != nil {
		t.Fatalf("expected instructions file at %s: %v", instrPath, err)
	}
	if !strings.Contains(string(data), "Always follow the team coding standards") {
		t.Errorf("expected team instructions in file, got:\n%s", string(data))
	}
}

// ---------------------------------------------------------------------------
// Invalid assistant name
// ---------------------------------------------------------------------------

func TestTeamInstallInvalidAssistant(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	dir := t.TempDir()
	writeTeamFile(t, dir, minimalTeamYAML)

	content := testContentFSForTeamInstall()

	var buf bytes.Buffer
	root := NewRootCommand(content)
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{
		"team", "install",
		"--for", "invalid-assistant",
		"--file", filepath.Join(dir, "team.yaml"),
		"--target", dir,
	})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for invalid assistant name")
	}
}

// ---------------------------------------------------------------------------
// formatInstalledPersonaSummary
// ---------------------------------------------------------------------------

func TestFormatInstalledPersonaSummary(t *testing.T) {
	r1 := dummyPersonaResult(1)
	r3 := dummyPersonaResult(3)
	entries := []personaInstallEntry{
		{
			name:   "front-end",
			result: &r1,
		},
		{
			name:   "back-end",
			result: &r3,
		},
	}

	got := formatInstalledPersonaSummary(entries)
	if !strings.Contains(got, "front-end (1 package)") {
		t.Errorf("expected 'front-end (1 package)' in %q", got)
	}
	if !strings.Contains(got, "back-end (3 packages)") {
		t.Errorf("expected 'back-end (3 packages)' in %q", got)
	}
}

// dummyPersonaResult creates a minimal PersonaResult with n packages.
func dummyPersonaResult(n int) installer.PersonaResult {
	pkgs := make(map[string]*installer.Result)
	for i := 0; i < n; i++ {
		pkgs[strings.Repeat("p", i+1)] = &installer.Result{}
	}
	return installer.PersonaResult{
		PackageResults: pkgs,
	}
}
