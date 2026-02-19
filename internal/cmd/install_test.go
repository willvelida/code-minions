package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestBuildPackageList(t *testing.T) {
	content := testContentFS()

	tests := []struct {
		name        string
		packageFlag string
		target      string
		expectDirs  []string
		expectError bool
	}{
		{
			name:       "no flag installs everything",
			target:     t.TempDir(),
			expectDirs: []string{"packages/developer-mentor", "packages/git-workflow"},
		},
		{
			name:        "single package",
			packageFlag: "git-workflow",
			target:      t.TempDir(),
			expectDirs:  []string{"packages/git-workflow"},
		},
		{
			name:        "invalid package returns error",
			packageFlag: "nonexistent",
			target:      t.TempDir(),
			expectError: true,
		},
		{
			name:        "whitespace around names is trimmed",
			packageFlag: " git-workflow , developer-mentor ",
			target:      t.TempDir(),
			expectDirs:  []string{"packages/git-workflow", "packages/developer-mentor"},
		},
		{
			name:        "empty items from double commas are skipped",
			packageFlag: "git-workflow,,developer-mentor",
			target:      t.TempDir(),
			expectDirs:  []string{"packages/git-workflow", "packages/developer-mentor"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dirs, err := buildPackageList(content, tt.packageFlag, tt.target)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(dirs) != len(tt.expectDirs) {
				t.Fatalf("dir count: got %d, want %d\n  got:  %v\n  want: %v", len(dirs), len(tt.expectDirs), dirs, tt.expectDirs)
			}

			for i, dir := range dirs {
				if dir != tt.expectDirs[i] {
					t.Errorf("dirs[%d]: got %q, want %q", i, dir, tt.expectDirs[i])
				}
			}
		})
	}
}

func TestBuildPackageListReadsManifest(t *testing.T) {
	content := testContentFS()
	target := t.TempDir()

	// Write a manifest with only git-workflow
	manifestPath := filepath.Join(target, "code-minions.yml")
	if err := os.WriteFile(manifestPath, []byte("name: test\npackages:\n  - git-workflow\n"), 0644); err != nil {
		t.Fatal(err)
	}

	dirs, err := buildPackageList(content, "", target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dirs) != 1 || dirs[0] != "packages/git-workflow" {
		t.Errorf("expected [packages/git-workflow], got %v", dirs)
	}
}

func TestBuildPackageListManifestEmptyPackagesFallsBack(t *testing.T) {
	content := testContentFS()
	target := t.TempDir()

	// Manifest exists but packages list is empty → install all
	manifestPath := filepath.Join(target, "code-minions.yml")
	if err := os.WriteFile(manifestPath, []byte("name: test\n"), 0644); err != nil {
		t.Fatal(err)
	}

	dirs, err := buildPackageList(content, "", target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should fall back to all packages
	if len(dirs) != 2 {
		t.Errorf("expected 2 packages (all), got %d: %v", len(dirs), dirs)
	}
}

func TestBuildPackageListPackageFlagOverridesManifest(t *testing.T) {
	content := testContentFS()
	target := t.TempDir()

	// Manifest says git-workflow, but --package says developer-mentor
	manifestPath := filepath.Join(target, "code-minions.yml")
	if err := os.WriteFile(manifestPath, []byte("name: test\npackages:\n  - git-workflow\n"), 0644); err != nil {
		t.Fatal(err)
	}

	dirs, err := buildPackageList(content, "developer-mentor", target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dirs) != 1 || dirs[0] != "packages/developer-mentor" {
		t.Errorf("expected [packages/developer-mentor], got %v", dirs)
	}
}

func TestInstallPackageFlagCreatesManifest(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	cmd := newInstallCommand(content)
	cmd.SetArgs([]string{
		"--package", "git-workflow",
		"--target", target,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Manifest should have been created
	manifestPath := filepath.Join(target, "code-minions.yml")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("manifest not created: %v", err)
	}
	if !strings.Contains(string(data), "git-workflow") {
		t.Errorf("manifest should contain git-workflow, got:\n%s", data)
	}
}

func TestInstallPackageFlagAddsToExistingManifest(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	// Pre-create a manifest with developer-mentor
	manifestPath := filepath.Join(target, "code-minions.yml")
	if err := os.WriteFile(manifestPath, []byte("name: test\npackages:\n    - developer-mentor\n"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := newInstallCommand(content)
	cmd.SetArgs([]string{
		"--package", "git-workflow",
		"--target", target,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	content2 := string(data)
	if !strings.Contains(content2, "developer-mentor") {
		t.Error("manifest should still contain developer-mentor")
	}
	if !strings.Contains(content2, "git-workflow") {
		t.Error("manifest should contain newly added git-workflow")
	}
}

func TestInstallDryRunDoesNotCreateManifest(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	cmd := newInstallCommand(content)
	cmd.SetArgs([]string{
		"--package", "git-workflow",
		"--target", target,
		"--dry-run",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	manifestPath := filepath.Join(target, "code-minions.yml")
	if _, err := os.Stat(manifestPath); !os.IsNotExist(err) {
		t.Error("manifest should not be created during dry-run")
	}
}

func TestInstallUsesManifestAssistantAsDefault(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	// Write a manifest with assistant: copilot
	manifestPath := filepath.Join(target, "code-minions.yml")
	if err := os.WriteFile(manifestPath, []byte("name: test\nassistant: copilot\npackages:\n    - git-workflow\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Install without --for — should use manifest's assistant
	cmd := newInstallCommand(content)
	cmd.SetArgs([]string{
		"--target", target,
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Agent file should be at Copilot path (.github/agents/)
	copilotAgent := filepath.Join(target, ".github", "agents", "git-workflow.agent.md")
	if _, err := os.Stat(copilotAgent); os.IsNotExist(err) {
		t.Errorf("expected agent at Copilot path: %s (manifest assistant should be used as default --for)", copilotAgent)
	}
}

func TestInstallForFlagOverridesManifestAssistant(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	// Write a manifest with assistant: copilot
	manifestPath := filepath.Join(target, "code-minions.yml")
	if err := os.WriteFile(manifestPath, []byte("name: test\nassistant: copilot\npackages:\n    - git-workflow\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Override isInteractiveFunc to avoid TTY prompt
	original := isInteractiveFunc
	isInteractiveFunc = func() bool { return false }
	defer func() { isInteractiveFunc = original }()

	// Install with --for claude — should override manifest
	cmd := newInstallCommand(content)
	cmd.SetArgs([]string{
		"--target", target,
		"--for", "claude",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Agent file should be at Claude path
	claudeAgent := filepath.Join(target, ".claude", "agents", "git-workflow.agent.md")
	if _, err := os.Stat(claudeAgent); os.IsNotExist(err) {
		t.Errorf("expected agent at Claude path: %s (--for should override manifest assistant)", claudeAgent)
	}
}

func testContentFS() fstest.MapFS {
	return fstest.MapFS{
		"packages/git-workflow/agents/git-workflow.agent.md":         &fstest.MapFile{Data: []byte("# Git Agent")},
		"packages/git-workflow/skills/git-workflow/SKILL.md":         &fstest.MapFile{Data: []byte("# Git")},
		"packages/developer-mentor/agents/developer-mentor.agent.md": &fstest.MapFile{Data: []byte("# Mentor Agent")},
		"packages/developer-mentor/skills/developer-mentor/SKILL.md": &fstest.MapFile{Data: []byte("# Mentor")},
	}
}

// TestInstallForCopilotRemapsPaths verifies that --for copilot places
// agent files in .github/agents/ while keeping skills in their default
// locations.
func TestInstallForCopilotRemapsPaths(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	cmd := newInstallCommand(content)
	cmd.SetArgs([]string{
		"--package", "git-workflow",
		"--for", "copilot",
		"--target", target,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Agent file should be in .github/agents/ (remapped from agents/)
	copilotAgent := filepath.Join(target, ".github", "agents", "git-workflow.agent.md")
	if _, err := os.Stat(copilotAgent); os.IsNotExist(err) {
		t.Errorf("expected agent at Copilot path: %s", copilotAgent)
	}

	// Skills should stay at skills/ (Copilot uses project-root skills/)
	skillFile := filepath.Join(target, "skills", "git-workflow", "SKILL.md")
	if _, err := os.Stat(skillFile); os.IsNotExist(err) {
		t.Errorf("expected skill at default path: %s", skillFile)
	}

	// Agent should NOT be at the old agents/ path
	oldAgent := filepath.Join(target, "agents", "git-workflow.agent.md")
	if _, err := os.Stat(oldAgent); !os.IsNotExist(err) {
		t.Errorf("agent should NOT exist at old path: %s", oldAgent)
	}
}

// TestInstallForClaudeRemapsPaths verifies that --for claude places
// agent files in .claude/agents/ and skills in .claude/skills/.
func TestInstallForClaudeRemapsPaths(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	cmd := newInstallCommand(content)
	cmd.SetArgs([]string{
		"--package", "git-workflow",
		"--for", "claude",
		"--target", target,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Agent should be in .claude/agents/
	claudeAgent := filepath.Join(target, ".claude", "agents", "git-workflow.agent.md")
	if _, err := os.Stat(claudeAgent); os.IsNotExist(err) {
		t.Errorf("expected agent at Claude path: %s", claudeAgent)
	}

	// Skills should be in .claude/skills/
	claudeSkill := filepath.Join(target, ".claude", "skills", "git-workflow", "SKILL.md")
	if _, err := os.Stat(claudeSkill); os.IsNotExist(err) {
		t.Errorf("expected skill at Claude path: %s", claudeSkill)
	}
}

// TestInstallForUnknownAssistantReturnsError verifies that --for
// with an invalid assistant name returns a helpful error.
func TestInstallForUnknownAssistantReturnsError(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	cmd := newInstallCommand(content)
	cmd.SetArgs([]string{
		"--package", "git-workflow",
		"--for", "vscode",
		"--target", target,
	})

	// Silence Cobra's default error printing so it doesn't clutter test output
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for unknown assistant, got nil")
	}
}

// TestInstallWithoutForPreservesBehaviour verifies that omitting
// --for produces the same output as before (no path remapping).
func TestInstallWithoutForPreservesBehaviour(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	cmd := newInstallCommand(content)
	cmd.SetArgs([]string{
		"--package", "git-workflow",
		"--target", target,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Without --for, agent lands in agents/ (the generic location)
	genericAgent := filepath.Join(target, "agents", "git-workflow.agent.md")
	if _, err := os.Stat(genericAgent); os.IsNotExist(err) {
		t.Errorf("expected agent at generic path: %s", genericAgent)
	}

	// Skills in skills/
	genericSkill := filepath.Join(target, "skills", "git-workflow", "SKILL.md")
	if _, err := os.Stat(genericSkill); os.IsNotExist(err) {
		t.Errorf("expected skill at generic path: %s", genericSkill)
	}
}

// TestInstallCreatesAgentsMD verifies that installing a package
// creates AGENTS.md when it does not exist.
func TestInstallCreatesAgentsMD(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	cmd := newInstallCommand(content)
	cmd.SetArgs([]string{
		"--package", "git-workflow",
		"--target", target,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	agentsMD := filepath.Join(target, "AGENTS.md")
	data, err := os.ReadFile(agentsMD)
	if err != nil {
		t.Fatalf("expected AGENTS.md to be created: %v", err)
	}

	if !strings.Contains(string(data), "code-minions") {
		t.Errorf("AGENTS.md content missing expected text, got:\n%s", data)
	}
}

// TestInstallSkipsExistingAgentsMD verifies that installing a package
// does not overwrite an existing AGENTS.md.
func TestInstallSkipsExistingAgentsMD(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	// Pre-create AGENTS.md with custom content
	original := []byte("# My Custom Agents\n")
	if err := os.WriteFile(filepath.Join(target, "AGENTS.md"), original, 0644); err != nil {
		t.Fatal(err)
	}

	cmd := newInstallCommand(content)
	cmd.SetArgs([]string{
		"--package", "git-workflow",
		"--target", target,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(target, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != string(original) {
		t.Errorf("AGENTS.md was overwritten:\n  got:  %q\n  want: %q", got, original)
	}
}

// TestInstallForCopilotCreatesAgentsMDAtMappedPath verifies that
// --for copilot creates AGENTS.md at .github/agents/AGENTS.md.
func TestInstallForCopilotCreatesAgentsMDAtMappedPath(t *testing.T) {
	target := t.TempDir()
	content := testContentFS()

	cmd := newInstallCommand(content)
	cmd.SetArgs([]string{
		"--package", "git-workflow",
		"--for", "copilot",
		"--target", target,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	agentsMD := filepath.Join(target, ".github", "agents", "AGENTS.md")
	if _, err := os.Stat(agentsMD); os.IsNotExist(err) {
		t.Errorf("expected AGENTS.md at Copilot path: %s", agentsMD)
	}

	// Should NOT exist at default path
	defaultPath := filepath.Join(target, "AGENTS.md")
	if _, err := os.Stat(defaultPath); !os.IsNotExist(err) {
		t.Errorf("AGENTS.md should NOT exist at default path when using --for copilot")
	}
}

// testContentFSWithMCP returns a content FS that includes an mcp.yaml for the
// git-workflow package, enabling cmd-level MCP integration tests.
func testContentFSWithMCP() fstest.MapFS {
	mcpYAML := `servers:
  github:
    description: GitHub API access via MCP
    transport: stdio
    command: npx
    args:
      - "-y"
      - "@modelcontextprotocol/server-github"
    env:
      GITHUB_PERSONAL_ACCESS_TOKEN: ""
    required: false
`
	return fstest.MapFS{
		"packages/git-workflow/agents/git-workflow.agent.md":         &fstest.MapFile{Data: []byte("# Git Agent")},
		"packages/git-workflow/skills/git-workflow/SKILL.md":         &fstest.MapFile{Data: []byte("# Git")},
		"packages/git-workflow/mcp.yaml":                             &fstest.MapFile{Data: []byte(mcpYAML)},
		"packages/developer-mentor/agents/developer-mentor.agent.md": &fstest.MapFile{Data: []byte("# Mentor Agent")},
		"packages/developer-mentor/skills/developer-mentor/SKILL.md": &fstest.MapFile{Data: []byte("# Mentor")},
	}
}

// TestInstallForCopilotWithMCP verifies that --for copilot processes
// mcp.yaml and writes the Copilot MCP config file.
func TestInstallForCopilotWithMCP(t *testing.T) {
	target := t.TempDir()
	content := testContentFSWithMCP()

	cmd := newInstallCommand(content)
	cmd.SetArgs([]string{
		"--package", "git-workflow",
		"--for", "copilot",
		"--target", target,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// MCP config should be written at .vscode/mcp.json
	mcpConfig := filepath.Join(target, ".vscode", "mcp.json")
	data, err := os.ReadFile(mcpConfig)
	if err != nil {
		t.Fatalf("expected MCP config at %s: %v", mcpConfig, err)
	}

	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("invalid JSON in MCP config: %v", err)
	}

	servers, ok := doc["servers"].(map[string]any)
	if !ok {
		t.Fatal("expected 'servers' key in MCP config")
	}
	if _, ok := servers["github"]; !ok {
		t.Error("expected 'github' server in MCP config")
	}
}

// TestInstallForCopilotWithMCPJSON verifies that --json output includes
// the MCP section when installing a package with mcp.yaml.
func TestInstallForCopilotWithMCPJSON(t *testing.T) {
	target := t.TempDir()
	content := testContentFSWithMCP()

	root := NewRootCommand(content)
	var buf strings.Builder
	root.SetOut(&buf)
	root.SetArgs([]string{
		"install",
		"--package", "git-workflow",
		"--for", "copilot",
		"--target", target,
		"--json",
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Copied  []string `json:"copied"`
		Skipped []string `json:"skipped"`
		Errors  []string `json:"errors"`
		MCP     []struct {
			Package    string   `json:"package"`
			ConfigPath string   `json:"config_path"`
			Added      []string `json:"added"`
			Skipped    []string `json:"skipped"`
			Conflicts  []string `json:"conflicts"`
			Warnings   []string `json:"warnings"`
		} `json:"mcp"`
		Summary struct {
			Copied  int `json:"copied"`
			Skipped int `json:"skipped"`
			Errors  int `json:"errors"`
		} `json:"summary"`
	}
	if err := json.Unmarshal([]byte(buf.String()), &result); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v\nraw output: %s", err, buf.String())
	}

	if len(result.MCP) != 1 {
		t.Fatalf("expected 1 MCP entry, got %d", len(result.MCP))
	}
	if result.MCP[0].Package != "git-workflow" {
		t.Errorf("MCP package = %q, want %q", result.MCP[0].Package, "git-workflow")
	}
	if len(result.MCP[0].Added) != 1 || result.MCP[0].Added[0] != "github" {
		t.Errorf("MCP added = %v, want [github]", result.MCP[0].Added)
	}
	if result.MCP[0].ConfigPath != ".vscode/mcp.json" {
		t.Errorf("MCP config_path = %q, want '.vscode/mcp.json'", result.MCP[0].ConfigPath)
	}
}

// TestInstallForClaudeWithMCPJSON verifies MCP JSON output for Claude.
func TestInstallForClaudeWithMCPJSON(t *testing.T) {
	target := t.TempDir()
	content := testContentFSWithMCP()

	root := NewRootCommand(content)
	var buf strings.Builder
	root.SetOut(&buf)
	root.SetArgs([]string{
		"install",
		"--package", "git-workflow",
		"--for", "claude",
		"--target", target,
		"--json",
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		MCP []struct {
			Package    string   `json:"package"`
			ConfigPath string   `json:"config_path"`
			Added      []string `json:"added"`
		} `json:"mcp"`
	}
	if err := json.Unmarshal([]byte(buf.String()), &result); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v\nraw output: %s", err, buf.String())
	}

	if len(result.MCP) != 1 {
		t.Fatalf("expected 1 MCP entry, got %d", len(result.MCP))
	}
	if result.MCP[0].ConfigPath != ".claude/settings.local.json" {
		t.Errorf("MCP config_path = %q, want '.claude/settings.local.json'", result.MCP[0].ConfigPath)
	}
}

// TestInstallWithoutForSkipsMCP verifies that omitting --for does not
// process mcp.yaml or create MCP config files.
func TestInstallWithoutForSkipsMCP(t *testing.T) {
	target := t.TempDir()
	content := testContentFSWithMCP()

	cmd := newInstallCommand(content)
	cmd.SetArgs([]string{
		"--package", "git-workflow",
		"--target", target,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// No MCP config should be created
	mcpConfig := filepath.Join(target, ".vscode", "mcp.json")
	if _, err := os.Stat(mcpConfig); !os.IsNotExist(err) {
		t.Errorf("MCP config should NOT be created without --for: %s", mcpConfig)
	}
}

// TestInstallMCPDryRun verifies that --dry-run does not write MCP config.
func TestInstallMCPDryRun(t *testing.T) {
	target := t.TempDir()
	content := testContentFSWithMCP()

	cmd := newInstallCommand(content)
	cmd.SetArgs([]string{
		"--package", "git-workflow",
		"--for", "copilot",
		"--target", target,
		"--dry-run",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// MCP config should NOT exist
	mcpConfig := filepath.Join(target, ".vscode", "mcp.json")
	if _, err := os.Stat(mcpConfig); !os.IsNotExist(err) {
		t.Errorf("MCP config should NOT be written in dry-run: %s", mcpConfig)
	}
}

// TestInstallMCPManifestRecordsServers verifies that the manifest records
// MCP server names after install.
func TestInstallMCPManifestRecordsServers(t *testing.T) {
	target := t.TempDir()
	content := testContentFSWithMCP()

	cmd := newInstallCommand(content)
	cmd.SetArgs([]string{
		"--package", "git-workflow",
		"--for", "copilot",
		"--target", target,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Read manifest and check MCPServers
	manifestPath := filepath.Join(target, ".code-minions", "installed.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("expected manifest at %s: %v", manifestPath, err)
	}

	var manifest struct {
		Packages []struct {
			Name       string   `json:"name"`
			MCPServers []string `json:"mcp_servers"`
		} `json:"packages"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("invalid JSON in manifest: %v", err)
	}

	found := false
	for _, pkg := range manifest.Packages {
		if pkg.Name == "git-workflow" {
			found = true
			if len(pkg.MCPServers) != 1 || pkg.MCPServers[0] != "github" {
				t.Errorf("MCPServers = %v, want [github]", pkg.MCPServers)
			}
			break
		}
	}
	if !found {
		t.Error("expected git-workflow in manifest packages")
	}
}

// testContentFSMCPOnly returns a content FS with a package that contains
// only an mcp.yaml (no agent/skill files). Used to verify that MCP-only
// packages are still tracked in the manifest.
func testContentFSMCPOnly() fstest.MapFS {
	mcpYAML := `servers:
  github:
    description: GitHub API access via MCP
    transport: stdio
    command: npx
    args:
      - "-y"
      - "@modelcontextprotocol/server-github"
    env:
      GITHUB_PERSONAL_ACCESS_TOKEN: ""
    required: false
`
	return fstest.MapFS{
		"packages/mcp-only/mcp.yaml": &fstest.MapFile{Data: []byte(mcpYAML)},
	}
}

// TestInstallMCPOnlyPackageManifest verifies that a package with only an
// mcp.yaml (no agent/skill files) is still recorded in the manifest.
func TestInstallMCPOnlyPackageManifest(t *testing.T) {
	target := t.TempDir()
	content := testContentFSMCPOnly()

	cmd := newInstallCommand(content)
	cmd.SetArgs([]string{
		"--package", "mcp-only",
		"--for", "copilot",
		"--target", target,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// MCP config should be written
	mcpConfig := filepath.Join(target, ".vscode", "mcp.json")
	if _, err := os.ReadFile(mcpConfig); err != nil {
		t.Fatalf("expected MCP config at %s: %v", mcpConfig, err)
	}

	// Manifest should track the package
	manifestPath := filepath.Join(target, ".code-minions", "installed.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("expected manifest at %s: %v", manifestPath, err)
	}

	var manifest struct {
		Packages []struct {
			Name       string   `json:"name"`
			MCPServers []string `json:"mcp_servers"`
		} `json:"packages"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("invalid JSON in manifest: %v", err)
	}

	found := false
	for _, pkg := range manifest.Packages {
		if pkg.Name == "mcp-only" {
			found = true
			if len(pkg.MCPServers) != 1 || pkg.MCPServers[0] != "github" {
				t.Errorf("MCPServers = %v, want [github]", pkg.MCPServers)
			}
			break
		}
	}
	if !found {
		t.Error("expected mcp-only package in manifest despite no agent/skill files")
	}
}
