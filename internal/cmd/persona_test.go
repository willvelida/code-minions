package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

// testContentFSWithPersonas returns a test filesystem that includes
// both packages and a persona. This lets us test the --persona flag
// end-to-end through the CLI layer.
//
// The persona "test-persona" references:
//   - git-workflow (has agents/ and skills/)
//   - developer-mentor (has agents/ and skills/)
func testContentFSWithPersonas() fstest.MapFS {
	personaYAML := `name: test-persona
description: A test persona for CLI testing
packages:
  - name: git-workflow
  - name: developer-mentor
`
	return fstest.MapFS{
		// Packages (same as testContentFS)
		"packages/git-workflow/package.yaml":                         &fstest.MapFile{Data: []byte("name: git-workflow\nversion: \"0.1.0\"\ndescription: Git workflow helpers\n")},
		"packages/git-workflow/agents/git-workflow.agent.md":         &fstest.MapFile{Data: []byte("# Git Agent")},
		"packages/git-workflow/skills/git-workflow/SKILL.md":         &fstest.MapFile{Data: []byte("# Git")},
		"packages/developer-mentor/package.yaml":                     &fstest.MapFile{Data: []byte("name: developer-mentor\nversion: \"0.1.0\"\ndescription: Developer mentoring\n")},
		"packages/developer-mentor/agents/developer-mentor.agent.md": &fstest.MapFile{Data: []byte("# Mentor Agent")},
		"packages/developer-mentor/skills/developer-mentor/SKILL.md": &fstest.MapFile{Data: []byte("# Mentor")},

		// Persona
		"personas/test-persona/persona.yaml": &fstest.MapFile{Data: []byte(personaYAML)},
	}
}

// ---------------------------------------------------------------------------
// Install --persona tests
// ---------------------------------------------------------------------------

// TestInstallPersonaCopilot verifies that `install --persona X --for copilot`
// installs all the persona's packages with correct path mapping and
// generates a grouping artefact.
func TestInstallPersonaCopilot(t *testing.T) {
	target := t.TempDir()
	content := testContentFSWithPersonas()

	cmd := newInstallCommand(content)
	cmd.SetArgs([]string{
		"--persona", "test-persona",
		"--for", "copilot",
		"--target", target,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Agent files should be in .github/agents/ (Copilot path mapping)
	copilotAgent := filepath.Join(target, ".github", "agents", "git-workflow.agent.md")
	if _, err := os.Stat(copilotAgent); os.IsNotExist(err) {
		t.Errorf("expected agent at Copilot path: %s", copilotAgent)
	}

	mentorAgent := filepath.Join(target, ".github", "agents", "developer-mentor.agent.md")
	if _, err := os.Stat(mentorAgent); os.IsNotExist(err) {
		t.Errorf("expected mentor agent at Copilot path: %s", mentorAgent)
	}

	// Skills should be at skills/ (Copilot keeps skills in project root)
	skillFile := filepath.Join(target, "skills", "git-workflow", "SKILL.md")
	if _, err := os.Stat(skillFile); os.IsNotExist(err) {
		t.Errorf("expected skill at default path: %s", skillFile)
	}

	// Manifest should exist and contain a persona entry
	manifestPath := filepath.Join(target, ".code-minions", "installed.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("expected manifest at %s: %v", manifestPath, err)
	}

	var manifest struct {
		Personas []struct {
			Name      string   `json:"name"`
			Assistant string   `json:"assistant"`
			Packages  []string `json:"packages"`
		} `json:"personas"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("invalid JSON in manifest: %v", err)
	}

	if len(manifest.Personas) != 1 {
		t.Fatalf("expected 1 persona in manifest, got %d", len(manifest.Personas))
	}
	if manifest.Personas[0].Name != "test-persona" {
		t.Errorf("persona name = %q, want test-persona", manifest.Personas[0].Name)
	}
	if manifest.Personas[0].Assistant != "copilot" {
		t.Errorf("persona assistant = %q, want copilot", manifest.Personas[0].Assistant)
	}
}

// TestInstallPersonaClaude verifies persona install for Claude.
func TestInstallPersonaClaude(t *testing.T) {
	target := t.TempDir()
	content := testContentFSWithPersonas()

	cmd := newInstallCommand(content)
	cmd.SetArgs([]string{
		"--persona", "test-persona",
		"--for", "claude",
		"--target", target,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Claude places agents in .claude/agents/
	claudeAgent := filepath.Join(target, ".claude", "agents", "git-workflow.agent.md")
	if _, err := os.Stat(claudeAgent); os.IsNotExist(err) {
		t.Errorf("expected agent at Claude path: %s", claudeAgent)
	}

	// Claude places skills in .claude/skills/
	claudeSkill := filepath.Join(target, ".claude", "skills", "git-workflow", "SKILL.md")
	if _, err := os.Stat(claudeSkill); os.IsNotExist(err) {
		t.Errorf("expected skill at Claude path: %s", claudeSkill)
	}
}

// TestInstallPersonaDryRun verifies that --dry-run doesn't write files.
func TestInstallPersonaDryRun(t *testing.T) {
	target := t.TempDir()
	content := testContentFSWithPersonas()

	cmd := newInstallCommand(content)
	cmd.SetArgs([]string{
		"--persona", "test-persona",
		"--for", "copilot",
		"--target", target,
		"--dry-run",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// No files should be written
	entries, _ := os.ReadDir(target)
	if len(entries) > 0 {
		t.Errorf("dry-run should not create any files, found %d entries", len(entries))
	}
}

// TestInstallPersonaWithoutFor verifies that --persona without --for
// returns an error.
func TestInstallPersonaWithoutFor(t *testing.T) {
	content := testContentFSWithPersonas()

	cmd := newInstallCommand(content)
	cmd.SetArgs([]string{
		"--persona", "test-persona",
		"--target", t.TempDir(),
	})
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when --persona used without --for")
	}
	if want := "--for is required"; !strings.Contains(err.Error(), want) {
		t.Errorf("error = %q, want it to contain %q", err, want)
	}
}

// TestInstallPersonaWithPackageConflict verifies that --persona and
// --package together return an error.
func TestInstallPersonaWithPackageConflict(t *testing.T) {
	content := testContentFSWithPersonas()

	cmd := newInstallCommand(content)
	cmd.SetArgs([]string{
		"--persona", "test-persona",
		"--package", "git-workflow",
		"--for", "copilot",
		"--target", t.TempDir(),
	})
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when both --persona and --package are used")
	}
	if want := "cannot be used together"; !strings.Contains(err.Error(), want) {
		t.Errorf("error = %q, want it to contain %q", err, want)
	}
}

// TestInstallPersonaNotFound verifies error for unknown persona name.
func TestInstallPersonaNotFound(t *testing.T) {
	content := testContentFSWithPersonas()

	cmd := newInstallCommand(content)
	cmd.SetArgs([]string{
		"--persona", "nonexistent",
		"--for", "copilot",
		"--target", t.TempDir(),
	})
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for nonexistent persona")
	}
}

// TestInstallPersonaJSON verifies JSON output for persona install.
func TestInstallPersonaJSON(t *testing.T) {
	target := t.TempDir()
	content := testContentFSWithPersonas()

	// Use the root command so that --json (a persistent flag) is available.
	cmd := NewRootCommand(content)
	cmd.SetArgs([]string{
		"install",
		"--persona", "test-persona",
		"--for", "copilot",
		"--target", target,
		"--json",
	})

	var buf strings.Builder
	cmd.SetOut(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Persona   string `json:"persona"`
		Assistant string `json:"assistant"`
		Summary   struct {
			Copied    int `json:"copied"`
			Skipped   int `json:"skipped"`
			Errors    int `json:"errors"`
			Generated int `json:"generated"`
		} `json:"summary"`
	}
	if err := json.Unmarshal([]byte(buf.String()), &result); err != nil {
		t.Fatalf("invalid JSON output: %v\nraw: %s", err, buf.String())
	}

	if result.Persona != "test-persona" {
		t.Errorf("persona = %q, want test-persona", result.Persona)
	}
	if result.Assistant != "copilot" {
		t.Errorf("assistant = %q, want copilot", result.Assistant)
	}
	if result.Summary.Errors != 0 {
		t.Errorf("errors = %d, want 0", result.Summary.Errors)
	}
	if result.Summary.Copied == 0 {
		t.Error("expected at least some files copied")
	}
}

// ---------------------------------------------------------------------------
// Uninstall --persona tests
// ---------------------------------------------------------------------------

// TestUninstallPersona verifies the full install → uninstall cycle.
func TestUninstallPersona(t *testing.T) {
	target := t.TempDir()
	content := testContentFSWithPersonas()

	// Step 1: Install the persona
	installCmd := newInstallCommand(content)
	installCmd.SetArgs([]string{
		"--persona", "test-persona",
		"--for", "copilot",
		"--target", target,
	})
	if err := installCmd.Execute(); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Verify files exist
	agentFile := filepath.Join(target, ".github", "agents", "git-workflow.agent.md")
	if _, err := os.Stat(agentFile); os.IsNotExist(err) {
		t.Fatalf("agent file should exist after install: %s", agentFile)
	}

	// Step 2: Uninstall the persona
	uninstallCmd := newUninstallCommand(content)
	uninstallCmd.SetArgs([]string{
		"--persona", "test-persona",
		"--for", "copilot",
		"--target", target,
		"--yes",
	})
	if err := uninstallCmd.Execute(); err != nil {
		t.Fatalf("uninstall failed: %v", err)
	}

	// Verify files are removed
	if _, err := os.Stat(agentFile); !os.IsNotExist(err) {
		t.Errorf("agent file should be removed after uninstall: %s", agentFile)
	}

	// Verify manifest no longer has the persona
	manifestPath := filepath.Join(target, ".code-minions", "installed.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("manifest should still exist: %v", err)
	}
	var manifest struct {
		Personas []struct{ Name string } `json:"personas"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("invalid manifest: %v", err)
	}
	if len(manifest.Personas) != 0 {
		t.Errorf("expected 0 personas after uninstall, got %d", len(manifest.Personas))
	}
}

// TestUninstallPersonaNotInstalled verifies error for uninstalling a
// persona that was never installed.
func TestUninstallPersonaNotInstalled(t *testing.T) {
	target := t.TempDir()
	content := testContentFSWithPersonas()

	cmd := newUninstallCommand(content)
	cmd.SetArgs([]string{
		"--persona", "nonexistent",
		"--for", "copilot",
		"--target", target,
		"--yes",
	})
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for persona that's not installed")
	}
	if want := "not installed"; !strings.Contains(err.Error(), want) {
		t.Errorf("error = %q, want it to contain %q", err, want)
	}
}

// ---------------------------------------------------------------------------
// Update --persona tests
// ---------------------------------------------------------------------------

// TestUpdatePersonaMutuallyExclusive verifies that --persona and
// --package together return an error on the update command.
func TestUpdatePersonaMutuallyExclusive(t *testing.T) {
	content := testContentFSWithPersonas()

	cmd := newUpdateCommand(content)
	cmd.SetArgs([]string{
		"--persona", "test-persona",
		"--package", "git-workflow",
		"--for", "copilot",
		"--target", t.TempDir(),
	})
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when both --persona and --package are used")
	}
}

// TestUpdatePersonaWithoutFor verifies that --persona without --for
// returns an error on the update command.
func TestUpdatePersonaWithoutFor(t *testing.T) {
	content := testContentFSWithPersonas()

	cmd := newUpdateCommand(content)
	cmd.SetArgs([]string{
		"--persona", "test-persona",
		"--target", t.TempDir(),
	})
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when --persona used without --for")
	}
}

// ---------------------------------------------------------------------------
// List --personas tests
// ---------------------------------------------------------------------------

// TestListShowsPersonasJSON verifies that list includes personas in
// JSON output when personas exist.
func TestListShowsPersonasJSON(t *testing.T) {
	content := testContentFSWithPersonas()

	// Use the root command so that --json (a persistent flag) is available.
	cmd := NewRootCommand(content)
	cmd.SetArgs([]string{"list", "--json"})

	var buf strings.Builder
	cmd.SetOut(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Packages []struct{ Name string } `json:"packages"`
		Personas []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"personas"`
	}
	if err := json.Unmarshal([]byte(buf.String()), &result); err != nil {
		t.Fatalf("invalid JSON: %v\nraw: %s", err, buf.String())
	}

	if len(result.Personas) != 1 {
		t.Fatalf("expected 1 persona, got %d", len(result.Personas))
	}
	if result.Personas[0].Name != "test-persona" {
		t.Errorf("persona name = %q, want test-persona", result.Personas[0].Name)
	}
}

// ---------------------------------------------------------------------------
// MCP persona tests
// ---------------------------------------------------------------------------

// testContentFSWithMCPPersonas extends the persona test FS with
// mcp.yaml files. Both packages declare a "github" MCP server (shared),
// while git-workflow also declares "linear" (exclusive).
func testContentFSWithMCPPersonas() fstest.MapFS {
	personaYAML := `name: test-persona
description: A test persona for CLI testing
packages:
  - name: git-workflow
  - name: developer-mentor
`
	return fstest.MapFS{
		// Packages
		"packages/git-workflow/package.yaml":                         &fstest.MapFile{Data: []byte("name: git-workflow\nversion: \"0.1.0\"\ndescription: Git workflow helpers\n")},
		"packages/git-workflow/agents/git-workflow.agent.md":         &fstest.MapFile{Data: []byte("# Git Agent")},
		"packages/git-workflow/skills/git-workflow/SKILL.md":         &fstest.MapFile{Data: []byte("# Git")},
		"packages/developer-mentor/package.yaml":                     &fstest.MapFile{Data: []byte("name: developer-mentor\nversion: \"0.1.0\"\ndescription: Developer mentoring\n")},
		"packages/developer-mentor/agents/developer-mentor.agent.md": &fstest.MapFile{Data: []byte("# Mentor Agent")},
		"packages/developer-mentor/skills/developer-mentor/SKILL.md": &fstest.MapFile{Data: []byte("# Mentor")},

		// MCP configs — both packages claim "github", git-workflow also claims "linear"
		"packages/git-workflow/mcp.yaml": &fstest.MapFile{Data: []byte(`servers:
  github:
    transport: stdio
    command: npx
    args: ["-y", "@modelcontextprotocol/server-github"]
    env:
      GITHUB_TOKEN: ""
  linear:
    transport: stdio
    command: npx
    args: ["-y", "@modelcontextprotocol/server-linear"]
    env:
      LINEAR_TOKEN: ""
`)},
		"packages/developer-mentor/mcp.yaml": &fstest.MapFile{Data: []byte(`servers:
  github:
    transport: stdio
    command: npx
    args: ["-y", "@modelcontextprotocol/server-github"]
    env:
      GITHUB_TOKEN: ""
`)},

		// Persona
		"personas/test-persona/persona.yaml": &fstest.MapFile{Data: []byte(personaYAML)},
	}
}

// TestUninstallPersonaRemovesMCP verifies that uninstalling a persona
// removes its MCP servers from the assistant config file.
func TestUninstallPersonaRemovesMCP(t *testing.T) {
	target := t.TempDir()
	content := testContentFSWithMCPPersonas()

	// Step 1: Install the persona (which should also install MCP servers)
	installCmd := newInstallCommand(content)
	installCmd.SetArgs([]string{
		"--persona", "test-persona",
		"--for", "copilot",
		"--target", target,
	})
	if err := installCmd.Execute(); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Verify MCP config file was created with servers
	mcpPath := filepath.Join(target, ".vscode", "mcp.json")
	data, err := os.ReadFile(mcpPath)
	if err != nil {
		t.Fatalf("MCP config should exist after install: %v", err)
	}

	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("invalid MCP JSON: %v", err)
	}
	servers, ok := doc["servers"].(map[string]any)
	if !ok || len(servers) == 0 {
		t.Fatalf("expected servers in MCP config, got: %v", doc)
	}

	// Step 2: Uninstall the persona
	uninstallCmd := newUninstallCommand(content)
	uninstallCmd.SetArgs([]string{
		"--persona", "test-persona",
		"--for", "copilot",
		"--target", target,
		"--yes",
	})
	if err := uninstallCmd.Execute(); err != nil {
		t.Fatalf("uninstall failed: %v", err)
	}

	// Verify MCP servers were removed
	data, err = os.ReadFile(mcpPath)
	if err != nil {
		// File was completely removed — that's fine too.
		return
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("invalid MCP JSON after uninstall: %v", err)
	}
	servers, _ = doc["servers"].(map[string]any)
	if len(servers) > 0 {
		t.Errorf("expected all MCP servers removed, still have: %v", servers)
	}
}
