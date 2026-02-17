package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/fatih/color"
	"github.com/willvelida/code-minions/internal/installer"
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

// TestUninstallPersonaJSON verifies JSON output for persona uninstall.
func TestUninstallPersonaJSON(t *testing.T) {
	target := t.TempDir()
	content := testContentFSWithPersonas()

	// Install first
	installCmd := newInstallCommand(content)
	installCmd.SetArgs([]string{
		"--persona", "test-persona",
		"--for", "copilot",
		"--target", target,
	})
	if err := installCmd.Execute(); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Uninstall with --json
	root := NewRootCommand(content)
	root.SetArgs([]string{
		"uninstall",
		"--persona", "test-persona",
		"--for", "copilot",
		"--target", target,
		"--yes",
		"--json",
	})

	var buf strings.Builder
	root.SetOut(&buf)

	if err := root.Execute(); err != nil {
		t.Fatalf("uninstall --json failed: %v", err)
	}

	var result struct {
		Persona  string   `json:"persona"`
		Removed  []string `json:"removed"`
		NotFound []string `json:"not_found"`
		Warnings []string `json:"warnings"`
		Errors   []string `json:"errors"`
		Summary  struct {
			Removed  int `json:"removed"`
			NotFound int `json:"not_found"`
			Errors   int `json:"errors"`
		} `json:"summary"`
	}
	if err := json.Unmarshal([]byte(buf.String()), &result); err != nil {
		t.Fatalf("invalid JSON: %v\nraw: %s", err, buf.String())
	}
	if result.Persona != "test-persona" {
		t.Errorf("persona = %q, want test-persona", result.Persona)
	}
	if result.Summary.Removed == 0 {
		t.Error("expected at least 1 removed file")
	}
}

// TestUninstallPersonaDryRun verifies dry-run persona uninstall
// does not delete files.
func TestUninstallPersonaDryRun(t *testing.T) {
	target := t.TempDir()
	content := testContentFSWithPersonas()

	// Install first
	installCmd := newInstallCommand(content)
	installCmd.SetArgs([]string{
		"--persona", "test-persona",
		"--for", "copilot",
		"--target", target,
	})
	if err := installCmd.Execute(); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	agentFile := filepath.Join(target, ".github", "agents", "git-workflow.agent.md")
	if _, err := os.Stat(agentFile); os.IsNotExist(err) {
		t.Fatalf("agent file should exist after install: %s", agentFile)
	}

	// Dry-run uninstall
	uninstallCmd := newUninstallCommand(content)
	uninstallCmd.SetArgs([]string{
		"--persona", "test-persona",
		"--for", "copilot",
		"--target", target,
		"--dry-run",
	})
	if err := uninstallCmd.Execute(); err != nil {
		t.Fatalf("dry-run uninstall failed: %v", err)
	}

	// Files should still exist after dry-run
	if _, err := os.Stat(agentFile); os.IsNotExist(err) {
		t.Errorf("agent file should still exist after dry-run uninstall: %s", agentFile)
	}

	// Manifest should still list the persona
	manifest, err := installer.LoadManifest(target)
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if p := installer.FindInstalledPersona(manifest, "test-persona"); p == nil {
		t.Error("persona should still be in manifest after dry-run")
	}
}

// TestUninstallPersonaQuiet verifies quiet-mode persona uninstall
// produces no stdout on success.
func TestUninstallPersonaQuiet(t *testing.T) {
	target := t.TempDir()
	content := testContentFSWithPersonas()

	// Install first
	installCmd := newInstallCommand(content)
	installCmd.SetArgs([]string{
		"--persona", "test-persona",
		"--for", "copilot",
		"--target", target,
	})
	if err := installCmd.Execute(); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Uninstall with --quiet
	root := NewRootCommand(content)
	root.SetArgs([]string{
		"uninstall",
		"--persona", "test-persona",
		"--for", "copilot",
		"--target", target,
		"--yes",
		"--quiet",
	})

	var buf strings.Builder
	root.SetOut(&buf)

	if err := root.Execute(); err != nil {
		t.Fatalf("quiet uninstall failed: %v", err)
	}

	if buf.Len() != 0 {
		t.Errorf("expected no stdout in quiet mode, got: %q", buf.String())
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

// ---------------------------------------------------------------------------
// List --detail persona display
// ---------------------------------------------------------------------------

// TestListDetailShowsPersonaPackages verifies that `list --detail` shows
// the persona's packages in human-readable output (the "→ packages:" line).
func TestListDetailShowsPersonaPackages(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	content := testContentFSWithPersonas()

	var buf strings.Builder
	cmd := newListCommand(content)
	cmd.SetArgs([]string{"--detail"})
	cmd.SetOut(&buf)

	old := color.Output
	color.Output = &buf
	t.Cleanup(func() { color.Output = old })

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Should show the Personas section header
	if !strings.Contains(output, "Personas") {
		t.Errorf("output should contain Personas header, got:\n%s", output)
	}
	// Should show the persona name
	if !strings.Contains(output, "test-persona") {
		t.Errorf("output should contain persona name, got:\n%s", output)
	}
	// In --detail mode, should show the packages each persona references
	if !strings.Contains(output, "packages:") {
		t.Errorf("output should contain → packages: line, got:\n%s", output)
	}
	if !strings.Contains(output, "git-workflow") {
		t.Errorf("output should contain git-workflow in packages, got:\n%s", output)
	}
	if !strings.Contains(output, "developer-mentor") {
		t.Errorf("output should contain developer-mentor in packages, got:\n%s", output)
	}
}

// TestListDetailShowsPersonaDescription verifies that persona
// descriptions appear in the human-readable list output.
func TestListDetailShowsPersonaDescription(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	content := testContentFSWithPersonas()

	var buf strings.Builder
	cmd := newListCommand(content)
	cmd.SetArgs([]string{"--detail"})
	cmd.SetOut(&buf)

	old := color.Output
	color.Output = &buf
	t.Cleanup(func() { color.Output = old })

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Persona description should be displayed
	if !strings.Contains(output, "A test persona for CLI testing") {
		t.Errorf("output should show persona description, got:\n%s", output)
	}
}

// TestListPersonaDescriptionTruncation verifies that long descriptions
// get truncated in the list output (exercises truncateDesc).
func TestListPersonaDescriptionTruncation(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	// Build an FS with a very long description
	longDesc := strings.Repeat("a", 200)
	testFS := fstest.MapFS{
		"packages/test-pkg/package.yaml": &fstest.MapFile{
			Data: []byte("name: test-pkg\nversion: \"0.1.0\"\ndescription: '" + longDesc + "'\n"),
		},
		"packages/test-pkg/agents/test-pkg.agent.md": &fstest.MapFile{
			Data: []byte("# Test Agent"),
		},
	}

	var buf strings.Builder
	cmd := newListCommand(testFS)
	cmd.SetArgs([]string{})
	cmd.SetOut(&buf)

	old := color.Output
	color.Output = &buf
	t.Cleanup(func() { color.Output = old })

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	// Truncated descriptions end with "..."
	if !strings.Contains(output, "...") {
		t.Errorf("long description should be truncated with '...', got:\n%s", output)
	}
}

// ---------------------------------------------------------------------------
// Persona uninstall confirmation prompt tests
// ---------------------------------------------------------------------------

// testContentFSWithTwoPersonas returns a test FS with two personas that
// share a package (developer-mentor). This lets us test shared-package
// detection during uninstall.
func testContentFSWithTwoPersonas() fstest.MapFS {
	persona1YAML := `name: persona-a
description: First persona
packages:
  - name: git-workflow
  - name: developer-mentor
`
	persona2YAML := `name: persona-b
description: Second persona
packages:
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

		// Personas
		"personas/persona-a/persona.yaml": &fstest.MapFile{Data: []byte(persona1YAML)},
		"personas/persona-b/persona.yaml": &fstest.MapFile{Data: []byte(persona2YAML)},
	}
}

// TestUninstallPersonaConfirmationAccepted verifies that the persona
// uninstall path shows a confirmation prompt and proceeds when "y" is entered.
func TestUninstallPersonaConfirmationAccepted(t *testing.T) {
	target := t.TempDir()
	content := testContentFSWithPersonas()

	// Install first
	installCmd := newInstallCommand(content)
	installCmd.SetArgs([]string{
		"--persona", "test-persona",
		"--for", "copilot",
		"--target", target,
	})
	if err := installCmd.Execute(); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Override TTY check to simulate interactive terminal
	origIsInteractive := isInteractiveFunc
	isInteractiveFunc = func() bool { return true }
	t.Cleanup(func() { isInteractiveFunc = origIsInteractive })

	// Pipe "y" into stdin
	origStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdin = r
	t.Cleanup(func() {
		os.Stdin = origStdin
		_ = r.Close()
	})
	_, _ = w.WriteString("y\n")
	_ = w.Close()

	uninstallCmd := newUninstallCommand(content)
	uninstallCmd.SetArgs([]string{
		"--persona", "test-persona",
		"--for", "copilot",
		"--target", target,
	})
	if err := uninstallCmd.Execute(); err != nil {
		t.Fatalf("uninstall with confirmation failed: %v", err)
	}

	// Files should be removed
	agentFile := filepath.Join(target, ".github", "agents", "git-workflow.agent.md")
	if _, err := os.Stat(agentFile); !os.IsNotExist(err) {
		t.Errorf("agent file should be removed after confirmed uninstall: %s", agentFile)
	}
}

// TestUninstallPersonaConfirmationDeclined verifies that declining the
// confirmation prompt aborts the persona uninstall cleanly.
func TestUninstallPersonaConfirmationDeclined(t *testing.T) {
	target := t.TempDir()
	content := testContentFSWithPersonas()

	// Install first
	installCmd := newInstallCommand(content)
	installCmd.SetArgs([]string{
		"--persona", "test-persona",
		"--for", "copilot",
		"--target", target,
	})
	if err := installCmd.Execute(); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Override TTY check to simulate interactive terminal
	origIsInteractive := isInteractiveFunc
	isInteractiveFunc = func() bool { return true }
	t.Cleanup(func() { isInteractiveFunc = origIsInteractive })

	// Pipe "n" into stdin
	origStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdin = r
	t.Cleanup(func() {
		os.Stdin = origStdin
		_ = r.Close()
	})
	_, _ = w.WriteString("n\n")
	_ = w.Close()

	var buf strings.Builder
	uninstallCmd := newUninstallCommand(content)
	uninstallCmd.SetArgs([]string{
		"--persona", "test-persona",
		"--for", "copilot",
		"--target", target,
	})
	uninstallCmd.SetOut(&buf)
	if err := uninstallCmd.Execute(); err != nil {
		t.Fatalf("expected clean exit on decline, got: %v", err)
	}

	// Output should contain "Aborted"
	if !strings.Contains(buf.String(), "Aborted") {
		t.Errorf("expected 'Aborted' in output, got: %q", buf.String())
	}

	// Files should still exist
	agentFile := filepath.Join(target, ".github", "agents", "git-workflow.agent.md")
	if _, err := os.Stat(agentFile); os.IsNotExist(err) {
		t.Errorf("agent file should still exist after decline: %s", agentFile)
	}
}

// TestUninstallPersonaNonInteractiveAbortsWithoutYes verifies that in
// non-interactive mode (no TTY), persona uninstall fails without --yes.
func TestUninstallPersonaNonInteractiveAbortsWithoutYes(t *testing.T) {
	target := t.TempDir()
	content := testContentFSWithPersonas()

	// Install first
	installCmd := newInstallCommand(content)
	installCmd.SetArgs([]string{
		"--persona", "test-persona",
		"--for", "copilot",
		"--target", target,
	})
	if err := installCmd.Execute(); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Override TTY check to simulate non-interactive
	origIsInteractive := isInteractiveFunc
	isInteractiveFunc = func() bool { return false }
	t.Cleanup(func() { isInteractiveFunc = origIsInteractive })

	uninstallCmd := newUninstallCommand(content)
	uninstallCmd.SetArgs([]string{
		"--persona", "test-persona",
		"--for", "copilot",
		"--target", target,
	})
	uninstallCmd.SilenceErrors = true
	uninstallCmd.SilenceUsage = true

	err := uninstallCmd.Execute()
	if err == nil {
		t.Fatal("expected error for non-interactive without --yes")
	}
	if !strings.Contains(err.Error(), "--yes") {
		t.Errorf("error should mention --yes, got: %q", err.Error())
	}

	// Files should still exist
	agentFile := filepath.Join(target, ".github", "agents", "git-workflow.agent.md")
	if _, err := os.Stat(agentFile); os.IsNotExist(err) {
		t.Errorf("agent file should still exist: %s", agentFile)
	}
}

// TestUninstallPersonaWithSharedPackages verifies that when two personas
// share a package, uninstalling one keeps the shared package's files and
// produces a warning.
func TestUninstallPersonaWithSharedPackages(t *testing.T) {
	target := t.TempDir()
	content := testContentFSWithTwoPersonas()

	// Install both personas. Use --force on the second so that shared
	// package files are re-copied and properly tracked in the manifest
	// (without --force, skipped files don't appear in the persona entry).
	installA := newInstallCommand(content)
	installA.SetArgs([]string{
		"--persona", "persona-a",
		"--for", "copilot",
		"--target", target,
	})
	if err := installA.Execute(); err != nil {
		t.Fatalf("install persona-a failed: %v", err)
	}

	installB := newInstallCommand(content)
	installB.SetArgs([]string{
		"--persona", "persona-b",
		"--for", "copilot",
		"--target", target,
		"--force",
	})
	if err := installB.Execute(); err != nil {
		t.Fatalf("install persona-b failed: %v", err)
	}

	// developer-mentor agent should exist
	sharedAgent := filepath.Join(target, ".github", "agents", "developer-mentor.agent.md")
	if _, err := os.Stat(sharedAgent); os.IsNotExist(err) {
		t.Fatalf("shared agent should exist after install: %s", sharedAgent)
	}

	// Uninstall persona-a (shares developer-mentor with persona-b)
	var buf strings.Builder
	uninstallCmd := newUninstallCommand(content)
	uninstallCmd.SetArgs([]string{
		"--persona", "persona-a",
		"--for", "copilot",
		"--target", target,
		"--yes",
	})
	uninstallCmd.SetOut(&buf)
	if err := uninstallCmd.Execute(); err != nil {
		t.Fatalf("uninstall persona-a failed: %v", err)
	}

	output := buf.String()

	// git-workflow files should be removed (exclusive to persona-a)
	gitAgent := filepath.Join(target, ".github", "agents", "git-workflow.agent.md")
	if _, err := os.Stat(gitAgent); !os.IsNotExist(err) {
		t.Errorf("exclusive package file should be removed: %s", gitAgent)
	}

	// developer-mentor files should be KEPT (shared with persona-b)
	if _, err := os.Stat(sharedAgent); os.IsNotExist(err) {
		t.Errorf("shared package file should be kept: %s", sharedAgent)
	}

	// Output should mention that the shared package is kept
	if !strings.Contains(output, "shared") {
		t.Errorf("output should warn about shared package, got:\n%s", output)
	}
}

// TestUninstallPersonaNormalOutputShowsWarningsAndNotFound verifies
// the normal-mode display of warnings and not-found entries in
// formatPersonaUninstallResult.
func TestUninstallPersonaNormalOutputShowsWarningsAndNotFound(t *testing.T) {
	target := t.TempDir()
	content := testContentFSWithPersonas()

	// Install the persona
	installCmd := newInstallCommand(content)
	installCmd.SetArgs([]string{
		"--persona", "test-persona",
		"--for", "copilot",
		"--target", target,
	})
	if err := installCmd.Execute(); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Manually delete one file so it shows up as "not found" during uninstall
	agentFile := filepath.Join(target, ".github", "agents", "git-workflow.agent.md")
	if err := os.Remove(agentFile); err != nil {
		t.Fatalf("failed to pre-remove file: %v", err)
	}

	var buf strings.Builder
	uninstallCmd := newUninstallCommand(content)
	uninstallCmd.SetArgs([]string{
		"--persona", "test-persona",
		"--for", "copilot",
		"--target", target,
		"--yes",
	})
	uninstallCmd.SetOut(&buf)
	if err := uninstallCmd.Execute(); err != nil {
		t.Fatalf("uninstall failed: %v", err)
	}

	output := buf.String()

	// Should show "not found" for the manually deleted file
	if !strings.Contains(output, "not found") {
		t.Errorf("output should show 'not found' for pre-deleted file, got:\n%s", output)
	}
	// Should show a summary line
	if !strings.Contains(output, "removed") {
		t.Errorf("output should show summary with 'removed', got:\n%s", output)
	}
}

// TestInstallPersonaQuietShowsNoOutput verifies quiet-mode persona install
// produces no stdout on success (exercises formatPersonaResult quiet path).
func TestInstallPersonaQuietShowsNoOutput(t *testing.T) {
	target := t.TempDir()
	content := testContentFSWithPersonas()

	root := NewRootCommand(content)
	root.SetArgs([]string{
		"install",
		"--persona", "test-persona",
		"--for", "copilot",
		"--target", target,
		"--quiet",
	})

	var buf strings.Builder
	root.SetOut(&buf)

	if err := root.Execute(); err != nil {
		t.Fatalf("quiet install failed: %v", err)
	}

	if buf.Len() != 0 {
		t.Errorf("expected no stdout in quiet mode, got: %q", buf.String())
	}
}

// TestInstallPersonaMCPConflictDifferentConfigs verifies that installing
// a persona where two packages declare the same MCP server with different
// configs results in a conflict warning.
func TestInstallPersonaMCPConflictDifferentConfigs(t *testing.T) {
	// Build FS where two packages have same server name but different configs
	personaYAML := `name: conflict-persona
description: Persona with conflicting MCP
packages:
  - name: pkg-a
  - name: pkg-b
`
	testFS := fstest.MapFS{
		"packages/pkg-a/package.yaml":            &fstest.MapFile{Data: []byte("name: pkg-a\nversion: \"0.1.0\"\n")},
		"packages/pkg-a/agents/pkg-a.agent.md":   &fstest.MapFile{Data: []byte("# A")},
		"packages/pkg-b/package.yaml":            &fstest.MapFile{Data: []byte("name: pkg-b\nversion: \"0.1.0\"\n")},
		"packages/pkg-b/agents/pkg-b.agent.md":   &fstest.MapFile{Data: []byte("# B")},
		"personas/conflict-persona/persona.yaml": &fstest.MapFile{Data: []byte(personaYAML)},

		// Both packages declare "github" but with different commands
		"packages/pkg-a/mcp.yaml": &fstest.MapFile{Data: []byte(`servers:
  github:
    transport: stdio
    command: npx
    args: ["-y", "@modelcontextprotocol/server-github"]
    env:
      GITHUB_TOKEN: ""
`)},
		"packages/pkg-b/mcp.yaml": &fstest.MapFile{Data: []byte(`servers:
  github:
    transport: stdio
    command: docker
    args: ["run", "mcp-github"]
    env:
      GITHUB_TOKEN: ""
`)},
	}

	target := t.TempDir()
	root := NewRootCommand(testFS)
	root.SetArgs([]string{
		"install",
		"--persona", "conflict-persona",
		"--for", "copilot",
		"--target", target,
		"--json",
	})
	root.SilenceErrors = true
	root.SilenceUsage = true

	var buf strings.Builder
	root.SetOut(&buf)

	// The install may return an error because of the conflict count,
	// but the JSON output should still be written.
	_ = root.Execute()

	var result struct {
		Errors []string `json:"errors"`
	}
	if err := json.Unmarshal([]byte(buf.String()), &result); err != nil {
		t.Fatalf("invalid JSON: %v\nraw: %s", err, buf.String())
	}

	// Look for conflict warning
	found := false
	for _, e := range result.Errors {
		if strings.Contains(e, "conflict") || strings.Contains(e, "MCP conflict") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected MCP conflict warning in errors, got: %v", result.Errors)
	}
}
