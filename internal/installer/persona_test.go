package installer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/willvelida/code-minions/internal/assistant"
	"github.com/willvelida/code-minions/internal/model"
	"github.com/willvelida/code-minions/internal/registry"
)

// ---------- Test helpers ----------

// testResolvedPersona creates a ResolvedPersona for testing.
// It builds a fake filesystem with 2 packages and returns both
// the resolved persona and the embedded source.
//
// This simulates what the PersonaResolver produces after looking
// up a persona and resolving all its packages.
func testResolvedPersona(t *testing.T) (*registry.ResolvedPersona, *registry.EmbeddedSource) {
	t.Helper()

	// Create a fake filesystem with 2 packages.
	// Each package has an agent file and a skill file.
	testFS := fstest.MapFS{
		"packages/git-workflow/package.yaml": &fstest.MapFile{
			Data: []byte("name: git-workflow\nversion: 0.1.0\ndescription: Git workflow skill\n"),
		},
		"packages/git-workflow/agents/git-workflow.agent.md": &fstest.MapFile{
			Data: []byte("# Git Workflow Agent\n\nHelps with git branching and commits."),
		},
		"packages/git-workflow/skills/git-workflow/SKILL.md": &fstest.MapFile{
			Data: []byte("---\nname: git-workflow\ndescription: Git workflow skill\n---\n# Git Workflow"),
		},
		"packages/threat-modelling/package.yaml": &fstest.MapFile{
			Data: []byte("name: threat-modelling\nversion: 0.2.0\ndescription: Threat modelling skill\n"),
		},
		"packages/threat-modelling/agents/threat-modelling.agent.md": &fstest.MapFile{
			Data: []byte("# Threat Modelling Agent\n\nPerforms STRIDE analysis."),
		},
		"packages/threat-modelling/skills/threat-modelling/SKILL.md": &fstest.MapFile{
			Data: []byte("---\nname: threat-modelling\ndescription: Threat modelling skill\n---\n# Threat Modelling"),
		},
	}

	src := registry.NewEmbeddedSource(testFS)

	// Build the resolved persona manually.
	// In production, the PersonaResolver does this automatically.
	pkg1, _ := src.GetPackage("git-workflow")
	pkg2, _ := src.GetPackage("threat-modelling")

	resolved := &registry.ResolvedPersona{
		Persona: model.Persona{
			Name:        "senior-dev",
			Description: "A senior developer persona",
			Packages: []model.PackageRef{
				{Name: "git-workflow"},
				{Name: "threat-modelling"},
			},
			Instructions: "Prioritise code quality and security.",
		},
		Packages: []registry.ResolvedPackage{
			{Package: *pkg1, Source: src},
			{Package: *pkg2, Source: src},
		},
	}

	return resolved, src
}

// ---------- PersonaInstaller tests ----------

func TestPersonaInstallerInstallCopilot(t *testing.T) {
	// This test verifies that installing a persona for Copilot:
	// 1. Copies all package files to the right locations
	// 2. Generates a .agent.md persona file with frontmatter
	resolved, _ := testResolvedPersona(t)

	// Create a temporary directory to install into
	target := t.TempDir()

	pi := &PersonaInstaller{
		Resolved:      resolved,
		AssistantName: "copilot",
		Target:        target,
		Force:         false,
		DryRun:        false,
	}

	result, err := pi.Install()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that we got results for both packages
	if len(result.PackageResults) != 2 {
		t.Fatalf("expected 2 package results, got %d", len(result.PackageResults))
	}

	// Verify package files were copied
	// Copilot maps: agents/ → .github/agents/, skills/ stays as skills/
	expectedFiles := []string{
		".github/agents/git-workflow.agent.md",
		"skills/git-workflow/SKILL.md",
		".github/agents/threat-modelling.agent.md",
		"skills/threat-modelling/SKILL.md",
	}

	for _, f := range expectedFiles {
		fullPath := filepath.Join(target, f)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("expected file to exist: %s", f)
		}
	}

	// Verify persona .agent.md was generated
	if len(result.GeneratedFiles) == 0 {
		t.Fatal("expected generated files, got none")
	}

	personaAgentPath := filepath.Join(target, ".github/agents/senior-dev.agent.md")
	data, err := os.ReadFile(personaAgentPath)
	if err != nil {
		t.Fatalf("failed to read persona agent file: %v", err)
	}

	content := string(data)

	// Check YAML frontmatter
	if !strings.Contains(content, "---") {
		t.Error("persona agent missing YAML frontmatter")
	}
	if !strings.Contains(content, "description:") {
		t.Error("persona agent missing description in frontmatter")
	}
	// Check body references the persona and packages
	if !strings.Contains(content, "senior-dev") {
		t.Error("persona agent missing persona name")
	}
	if !strings.Contains(content, "git-workflow") {
		t.Error("persona agent missing git-workflow reference")
	}
	if !strings.Contains(content, "threat-modelling") {
		t.Error("persona agent missing threat-modelling reference")
	}

	// Check no errors
	if result.TotalErrors() > 0 {
		t.Errorf("expected 0 errors, got %d", result.TotalErrors())
		for _, e := range result.Errors {
			t.Logf("  error: %s", e)
		}
	}
}

func TestPersonaInstallerInstallClaude(t *testing.T) {
	// This test verifies that installing a persona for Claude:
	// 1. Copies files to .claude/agents/ and .claude/skills/
	// 2. Generates a persona subagent at .claude/agents/senior-dev.agent.md
	resolved, _ := testResolvedPersona(t)
	target := t.TempDir()

	pi := &PersonaInstaller{
		Resolved:      resolved,
		AssistantName: "claude",
		Target:        target,
		Force:         false,
		DryRun:        false,
	}

	result, err := pi.Install()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify Claude-specific file locations
	expectedFiles := []string{
		".claude/agents/git-workflow.agent.md",
		".claude/skills/git-workflow/SKILL.md",
		".claude/agents/threat-modelling.agent.md",
		".claude/skills/threat-modelling/SKILL.md",
	}

	for _, f := range expectedFiles {
		fullPath := filepath.Join(target, f)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("expected file to exist: %s", f)
		}
	}

	// Verify persona subagent was generated
	personaAgentPath := filepath.Join(target, ".claude/agents/senior-dev.agent.md")
	data, err := os.ReadFile(personaAgentPath)
	if err != nil {
		t.Fatalf("failed to read persona agent: %v", err)
	}

	content := string(data)

	// Check frontmatter
	if !strings.Contains(content, "description:") {
		t.Error("persona agent missing frontmatter description")
	}
	// Check body
	if !strings.Contains(content, "senior-dev") {
		t.Error("persona agent missing persona name in body")
	}
	if !strings.Contains(content, "Git Workflow") {
		t.Error("persona agent missing package reference")
	}

	if result.TotalErrors() > 0 {
		t.Errorf("expected 0 errors, got %d", result.TotalErrors())
	}
}

func TestPersonaInstallerInstallOpenCode(t *testing.T) {
	// This test verifies that installing a persona for OpenCode:
	// 1. Copies files to .opencode/agents/ and .opencode/skills/
	// 2. Generates a persona agent at .opencode/agents/senior-dev.md
	resolved, _ := testResolvedPersona(t)
	target := t.TempDir()

	pi := &PersonaInstaller{
		Resolved:      resolved,
		AssistantName: "opencode",
		Target:        target,
		Force:         false,
		DryRun:        false,
	}

	result, err := pi.Install()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify OpenCode-specific file locations
	expectedFiles := []string{
		".opencode/agents/git-workflow.agent.md",
		".opencode/skills/git-workflow/SKILL.md",
		".opencode/agents/threat-modelling.agent.md",
		".opencode/skills/threat-modelling/SKILL.md",
	}

	for _, f := range expectedFiles {
		fullPath := filepath.Join(target, f)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("expected file to exist: %s", f)
		}
	}

	// Verify persona agent was generated (note: .md not .agent.md)
	personaAgentPath := filepath.Join(target, ".opencode/agents/senior-dev.md")
	data, err := os.ReadFile(personaAgentPath)
	if err != nil {
		t.Fatalf("failed to read persona agent: %v", err)
	}

	content := string(data)

	// Check OpenCode-specific frontmatter
	if !strings.Contains(content, "mode: primary") {
		t.Error("persona agent missing OpenCode 'mode: primary' frontmatter")
	}
	if !strings.Contains(content, "skill: true") {
		t.Error("persona agent missing skill tool reference")
	}

	if result.TotalErrors() > 0 {
		t.Errorf("expected 0 errors, got %d", result.TotalErrors())
	}
}

func TestPersonaInstallerDryRun(t *testing.T) {
	// Dry run should NOT create any files on disk,
	// but should still report what WOULD be created.
	resolved, _ := testResolvedPersona(t)
	target := t.TempDir()

	pi := &PersonaInstaller{
		Resolved:      resolved,
		AssistantName: "copilot",
		Target:        target,
		Force:         false,
		DryRun:        true,
	}

	result, err := pi.Install()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have package results showing what would be copied
	if result.TotalCopied() == 0 {
		t.Error("expected dry-run to report files that would be copied")
	}

	// But no files should actually exist on disk
	agentPath := filepath.Join(target, ".github/agents/git-workflow.agent.md")
	if _, err := os.Stat(agentPath); err == nil {
		t.Error("dry-run should not create files on disk")
	}
}

func TestPersonaInstallerUnknownAssistant(t *testing.T) {
	// Installing for an unknown assistant should fail with a
	// clear error message.
	resolved, _ := testResolvedPersona(t)
	target := t.TempDir()

	pi := &PersonaInstaller{
		Resolved:      resolved,
		AssistantName: "nonexistent",
		Target:        target,
	}

	_, err := pi.Install()
	if err == nil {
		t.Fatal("expected error for unknown assistant")
	}
}

// ---------- Grouping generator tests ----------

func TestToTitleCase(t *testing.T) {
	// toTitleCase converts hyphenated names to Title Case.
	// This is used in all grouping generators for display names.
	tests := []struct {
		input string
		want  string
	}{
		{"git-workflow", "Git Workflow"},
		{"raise-pull-requests", "Raise Pull Requests"},
		{"threat-modelling", "Threat Modelling"},
		{"simple", "Simple"},
		{"", ""},
	}

	for _, tt := range tests {
		got := toTitleCase(tt.input)
		if got != tt.want {
			t.Errorf("toTitleCase(%q): got %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestBuildPackageSummary(t *testing.T) {
	// buildPackageSummary generates a Markdown list of packages.
	resolved, _ := testResolvedPersona(t)

	summary := buildPackageSummary(resolved)

	if !strings.Contains(summary, "git-workflow") {
		t.Error("summary missing git-workflow")
	}
	if !strings.Contains(summary, "threat-modelling") {
		t.Error("summary missing threat-modelling")
	}
	if !strings.Contains(summary, "**Git Workflow**") {
		t.Error("summary missing formatted display name")
	}
}

func TestPersonaResultCounters(t *testing.T) {
	// Test the TotalCopied/TotalSkipped/TotalErrors helper methods.
	result := &PersonaResult{
		PackageResults: map[string]*Result{
			"pkg-a": {
				Copied:  []string{"file1", "file2"},
				Skipped: []string{"file3"},
				Errors:  []string{"error1"},
			},
			"pkg-b": {
				Copied:  []string{"file4"},
				Skipped: []string{},
				Errors:  []string{},
			},
		},
		GeneratedFiles: []string{".github/agents/senior-dev.agent.md"},
		Errors:         []string{"persona-error"},
	}

	// TotalCopied = 3 package files + 1 generated file = 4
	if result.TotalCopied() != 4 {
		t.Errorf("TotalCopied: got %d, want 4", result.TotalCopied())
	}

	// TotalSkipped = 1
	if result.TotalSkipped() != 1 {
		t.Errorf("TotalSkipped: got %d, want 1", result.TotalSkipped())
	}

	// TotalErrors = 1 package error + 1 persona error = 2
	if result.TotalErrors() != 2 {
		t.Errorf("TotalErrors: got %d, want 2", result.TotalErrors())
	}
}

// ---------- Exported type assertions ----------
// These compile-time checks verify that our generators implement
// the GroupingGenerator interface. If they don't, the build fails.

var _ GroupingGenerator = (*CopilotGrouping)(nil)
var _ GroupingGenerator = (*ClaudeGrouping)(nil)
var _ GroupingGenerator = (*OpenCodeGrouping)(nil)
var _ GroupingGenerator = (*NoopGrouping)(nil)

// ---------- collectUniqueServers tests ----------

func TestCollectUniqueServers(t *testing.T) {
	// Two packages share "github", git-workflow also has "linear".
	// collectUniqueServers should return a sorted, deduplicated list.
	serversByPkg := map[string][]string{
		"git-workflow":        {"github", "linear"},
		"raise-pull-requests": {"github"},
	}

	got := collectUniqueServers(serversByPkg)

	if len(got) != 2 {
		t.Fatalf("expected 2 unique servers, got %d: %v", len(got), got)
	}
	if got[0] != "github" || got[1] != "linear" {
		t.Errorf("expected [github, linear], got %v", got)
	}
}

func TestCollectUniqueServersEmpty(t *testing.T) {
	got := collectUniqueServers(map[string][]string{})
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %v", got)
	}
}

// ---------- MCP persona tests ----------

// testResolvedPersonaWithMCP creates a ResolvedPersona where both
// packages have mcp.yaml files. The "github" server is declared by
// both packages (shared), while "linear" is only in git-workflow
// (exclusive). This lets us test deduplication and reference counting.
func testResolvedPersonaWithMCP(t *testing.T) (*registry.ResolvedPersona, fstest.MapFS) {
	t.Helper()

	testFS := fstest.MapFS{
		"packages/git-workflow/package.yaml": &fstest.MapFile{
			Data: []byte("name: git-workflow\nversion: 0.1.0\ndescription: Git workflow\n"),
		},
		"packages/git-workflow/agents/git-workflow.agent.md": &fstest.MapFile{
			Data: []byte("# Git Agent"),
		},
		"packages/git-workflow/skills/git-workflow/SKILL.md": &fstest.MapFile{
			Data: []byte("# Git"),
		},
		// git-workflow declares TWO MCP servers: github and linear
		"packages/git-workflow/mcp.yaml": &fstest.MapFile{
			Data: []byte(`servers:
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
`),
		},
		"packages/raise-pull-requests/package.yaml": &fstest.MapFile{
			Data: []byte("name: raise-pull-requests\nversion: 0.1.0\ndescription: PR helpers\n"),
		},
		"packages/raise-pull-requests/agents/raise-pull-requests.agent.md": &fstest.MapFile{
			Data: []byte("# PR Agent"),
		},
		"packages/raise-pull-requests/skills/raise-pull-requests/SKILL.md": &fstest.MapFile{
			Data: []byte("# PRs"),
		},
		// raise-pull-requests declares ONE MCP server: github (shared!)
		"packages/raise-pull-requests/mcp.yaml": &fstest.MapFile{
			Data: []byte(`servers:
  github:
    transport: stdio
    command: npx
    args: ["-y", "@modelcontextprotocol/server-github"]
    env:
      GITHUB_TOKEN: ""
`),
		},
	}

	src := registry.NewEmbeddedSource(testFS)
	pkg1, _ := src.GetPackage("git-workflow")
	pkg2, _ := src.GetPackage("raise-pull-requests")

	resolved := &registry.ResolvedPersona{
		Persona: model.Persona{
			Name:        "pr-dev",
			Description: "PR-focused developer",
			Packages: []model.PackageRef{
				{Name: "git-workflow"},
				{Name: "raise-pull-requests"},
			},
		},
		Packages: []registry.ResolvedPackage{
			{Package: *pkg1, Source: src},
			{Package: *pkg2, Source: src},
		},
	}

	return resolved, testFS
}

// TestPersonaInstallMCPMergesCopilot verifies that persona install
// collects mcp.yaml from multiple packages and merges the servers
// into the Copilot MCP config file (.vscode/mcp.json).
func TestPersonaInstallMCPMergesCopilot(t *testing.T) {
	resolved, contentFS := testResolvedPersonaWithMCP(t)
	target := t.TempDir()

	pi := &PersonaInstaller{
		Resolved:      resolved,
		Content:       contentFS,
		AssistantName: "copilot",
		Target:        target,
	}

	result, err := pi.Install()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The MCP result should exist.
	if result.MCPResult == nil {
		t.Fatal("expected MCPResult to be non-nil")
	}

	// Both "github" and "linear" should be added to .vscode/mcp.json.
	if len(result.MCPResult.Merge.Added) < 2 {
		t.Errorf("Added = %v, want at least [github, linear]", result.MCPResult.Merge.Added)
	}

	// Verify the config file was written.
	mcpPath := filepath.Join(target, ".vscode", "mcp.json")
	data, err := os.ReadFile(mcpPath)
	if err != nil {
		t.Fatalf("expected MCP config at %s: %v", mcpPath, err)
	}

	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("invalid MCP JSON: %v", err)
	}

	servers, ok := doc["servers"].(map[string]any)
	if !ok {
		t.Fatal("missing 'servers' key in MCP config")
	}
	if _, ok := servers["github"]; !ok {
		t.Error("expected 'github' server in config")
	}
	if _, ok := servers["linear"]; !ok {
		t.Error("expected 'linear' server in config")
	}
}

// TestPersonaInstallMCPServersByPackage verifies that the
// MCPServersByPackage map correctly tracks ownership.
func TestPersonaInstallMCPServersByPackage(t *testing.T) {
	resolved, contentFS := testResolvedPersonaWithMCP(t)
	target := t.TempDir()

	pi := &PersonaInstaller{
		Resolved:      resolved,
		Content:       contentFS,
		AssistantName: "copilot",
		Target:        target,
	}

	result, err := pi.Install()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// git-workflow should own both "github" and "linear".
	gwServers := result.MCPServersByPackage["git-workflow"]
	if len(gwServers) != 2 {
		t.Errorf("git-workflow servers = %v, want 2 entries", gwServers)
	}

	// raise-pull-requests should own "github".
	prServers := result.MCPServersByPackage["raise-pull-requests"]
	if len(prServers) == 0 {
		t.Error("raise-pull-requests should own at least 'github'")
	}
	hasGithub := false
	for _, s := range prServers {
		if s == "github" {
			hasGithub = true
		}
	}
	if !hasGithub {
		t.Errorf("raise-pull-requests servers = %v, want 'github'", prServers)
	}
}

// TestPersonaInstallMCPManifestTracksServers verifies that after
// persona installation, the manifest correctly records which MCP
// servers each package owns.
func TestPersonaInstallMCPManifestTracksServers(t *testing.T) {
	resolved, contentFS := testResolvedPersonaWithMCP(t)
	target := t.TempDir()

	pi := &PersonaInstaller{
		Resolved:      resolved,
		Content:       contentFS,
		AssistantName: "copilot",
		Target:        target,
	}

	_, err := pi.Install()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Load the manifest and check MCPServers per package.
	manifest, err := LoadManifest(target)
	if err != nil {
		t.Fatalf("LoadManifest failed: %v", err)
	}

	gwEntry := FindInstalled(manifest, "git-workflow")
	if gwEntry == nil {
		t.Fatal("git-workflow not in manifest")
	}
	if len(gwEntry.MCPServers) != 2 {
		t.Errorf("git-workflow MCPServers = %v, want 2 entries", gwEntry.MCPServers)
	}

	prEntry := FindInstalled(manifest, "raise-pull-requests")
	if prEntry == nil {
		t.Fatal("raise-pull-requests not in manifest")
	}
	hasGithub := false
	for _, s := range prEntry.MCPServers {
		if s == "github" {
			hasGithub = true
		}
	}
	if !hasGithub {
		t.Errorf("raise-pull-requests MCPServers = %v, want 'github'", prEntry.MCPServers)
	}

	// "github" is shared — both packages claim it.
	if !IsMCPServerShared(manifest, "github", map[string]bool{"git-workflow": true}) {
		t.Error("github should be shared between the two packages")
	}

	// "linear" is exclusive to git-workflow.
	if IsMCPServerShared(manifest, "linear", map[string]bool{"git-workflow": true}) {
		t.Error("linear should be exclusive to git-workflow")
	}
}

// TestPersonaInstallMCPDryRunNoFile verifies that dry-run does not
// write the MCP config file.
func TestPersonaInstallMCPDryRunNoFile(t *testing.T) {
	resolved, contentFS := testResolvedPersonaWithMCP(t)
	target := t.TempDir()

	pi := &PersonaInstaller{
		Resolved:      resolved,
		Content:       contentFS,
		AssistantName: "copilot",
		Target:        target,
		DryRun:        true,
	}

	result, err := pi.Install()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// MCP should still be collected and reported...
	if result.MCPResult == nil {
		t.Fatal("expected MCPResult even for dry-run")
	}

	// ...but the file should not exist.
	mcpPath := filepath.Join(target, ".vscode", "mcp.json")
	if _, err := os.Stat(mcpPath); !os.IsNotExist(err) {
		t.Error("MCP config file should not be written during dry-run")
	}
}

// TestPersonaInstallMCPNoContentFS verifies graceful behaviour when
// Content is nil (backwards compatibility with existing tests that
// don't set it).
func TestPersonaInstallMCPNoContentFS(t *testing.T) {
	resolved, _ := testResolvedPersona(t)
	target := t.TempDir()

	pi := &PersonaInstaller{
		Resolved:      resolved,
		Content:       nil, // no content FS — skip MCP processing
		AssistantName: "copilot",
		Target:        target,
	}

	result, err := pi.Install()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.MCPResult != nil {
		t.Error("expected nil MCPResult when Content is nil")
	}
}

// TestPersonaGroupingIncludesMCPServers verifies that the generated
// persona agent files include MCP server references in their
// frontmatter. Each assistant uses a different format:
//   - Copilot: tools: ['server/*']
//   - Claude:  mcpServers: ['server']
//   - OpenCode: tools: { server_*: true }
func TestPersonaGroupingIncludesMCPServers(t *testing.T) {
	tests := []struct {
		name      string
		assistant string
		wantFile  string   // expected generated file path
		wantMCP   []string // strings that should be in the file
	}{
		{
			name:      "copilot includes MCP tools",
			assistant: "copilot",
			wantFile:  ".github/agents/pr-dev.agent.md",
			wantMCP:   []string{"tools:", "github/*", "linear/*"},
		},
		{
			name:      "claude includes MCP servers",
			assistant: "claude",
			wantFile:  ".claude/agents/pr-dev.agent.md",
			wantMCP:   []string{"mcpServers:", "github", "linear"},
		},
		{
			name:      "opencode includes MCP tools",
			assistant: "opencode",
			wantFile:  ".opencode/agents/pr-dev.md",
			wantMCP:   []string{"github_*: true", "linear_*: true"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved, contentFS := testResolvedPersonaWithMCP(t)
			target := t.TempDir()

			pi := &PersonaInstaller{
				Resolved:      resolved,
				Content:       contentFS,
				AssistantName: tt.assistant,
				Target:        target,
				Force:         true,
			}

			result, err := pi.Install()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.TotalErrors() > 0 {
				for _, e := range result.Errors {
					t.Logf("  error: %s", e)
				}
				t.Fatalf("expected 0 errors, got %d", result.TotalErrors())
			}

			// Read the generated grouping file
			filePath := filepath.Join(target, tt.wantFile)
			data, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("failed to read grouping file %s: %v", tt.wantFile, err)
			}
			content := string(data)

			for _, want := range tt.wantMCP {
				if !strings.Contains(content, want) {
					t.Errorf("grouping file missing %q\ngot:\n%s", want, content)
				}
			}
		})
	}
}

// ---------- escapeYAMLValue tests ----------

func TestEscapeYAMLValue(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple", "A simple description", "A simple description"},
		{"with colon", "Step 1: do something", `"Step 1: do something"`},
		{"with quotes", `She said "hello"`, `"She said \"hello\""`},
		{"with newline", "line1\nline2", `"line1\nline2"`},
		{"empty", "", `""`},
		{"with hash", "# not a comment", `"# not a comment"`},
		{"with ampersand", "a & b", `"a & b"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeYAMLValue(tt.input)
			if got != tt.want {
				t.Errorf("escapeYAMLValue(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ---------- NoopGrouping tests ----------

func TestNoopGroupingGenerate(t *testing.T) {
	noop := &NoopGrouping{}
	files, err := noop.Generate()
	if err != nil {
		t.Fatalf("NoopGrouping.Generate() returned unexpected error: %v", err)
	}
	if files != nil {
		t.Errorf("NoopGrouping.Generate() = %v, want nil", files)
	}
}

func TestNoopGroupingSetMCPServers(t *testing.T) {
	noop := &NoopGrouping{}
	// Should not panic or error
	noop.SetMCPServers([]string{"github", "linear"})
	noop.SetMCPServers(nil)
}

// ---------- NewGroupingGenerator tests ----------

func TestNewGroupingGeneratorUnknownAssistant(t *testing.T) {
	// When given an unknown assistant name (via a custom Config),
	// the factory should return a NoopGrouping.
	cfg := &assistant.Config{Name: "unknown-assistant"}
	resolved, _ := testResolvedPersona(t)

	gen, err := NewGroupingGenerator(cfg, resolved, t.TempDir(), false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it's a NoopGrouping by calling Generate
	files, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}
	if files != nil {
		t.Errorf("expected nil files from NoopGrouping, got: %v", files)
	}
}

func TestNewGroupingGeneratorCopilot(t *testing.T) {
	cfg, _ := assistant.Get("copilot")
	resolved, _ := testResolvedPersona(t)

	gen, err := NewGroupingGenerator(cfg, resolved, t.TempDir(), false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := gen.(*CopilotGrouping); !ok {
		t.Errorf("expected *CopilotGrouping, got %T", gen)
	}
}

func TestNewGroupingGeneratorClaude(t *testing.T) {
	cfg, _ := assistant.Get("claude")
	resolved, _ := testResolvedPersona(t)

	gen, err := NewGroupingGenerator(cfg, resolved, t.TempDir(), false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := gen.(*ClaudeGrouping); !ok {
		t.Errorf("expected *ClaudeGrouping, got %T", gen)
	}
}

func TestNewGroupingGeneratorOpenCode(t *testing.T) {
	cfg, _ := assistant.Get("opencode")
	resolved, _ := testResolvedPersona(t)

	gen, err := NewGroupingGenerator(cfg, resolved, t.TempDir(), false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := gen.(*OpenCodeGrouping); !ok {
		t.Errorf("expected *OpenCodeGrouping, got %T", gen)
	}
}

// ---------- writeGeneratedFile tests ----------

func TestWriteGeneratedFileExistsNoForce(t *testing.T) {
	target := t.TempDir()
	relPath := filepath.Join("agents", "existing.md")

	// Pre-create the file
	fullPath := filepath.Join(target, relPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte("original"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	err := writeGeneratedFile(target, relPath, []byte("new content"), false, false)
	if err == nil {
		t.Fatal("expected error for existing file without force")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error = %q, want 'already exists'", err)
	}

	// Original content should be unchanged
	data, _ := os.ReadFile(fullPath)
	if string(data) != "original" {
		t.Errorf("file content = %q, want 'original'", data)
	}
}

func TestWriteGeneratedFileExistsWithForce(t *testing.T) {
	target := t.TempDir()
	relPath := filepath.Join("agents", "existing.md")

	// Pre-create the file
	fullPath := filepath.Join(target, relPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte("original"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	err := writeGeneratedFile(target, relPath, []byte("new content"), false, true)
	if err != nil {
		t.Fatalf("unexpected error with force=true: %v", err)
	}

	// Content should be overwritten
	data, _ := os.ReadFile(fullPath)
	if string(data) != "new content" {
		t.Errorf("file content = %q, want 'new content'", data)
	}
}

func TestWriteGeneratedFileDryRun(t *testing.T) {
	target := t.TempDir()
	relPath := filepath.Join("agents", "test.md")

	err := writeGeneratedFile(target, relPath, []byte("content"), true, false)
	if err != nil {
		t.Fatalf("unexpected error in dry-run: %v", err)
	}

	// File should NOT exist in dry-run
	fullPath := filepath.Join(target, relPath)
	if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
		t.Errorf("file should not exist in dry-run mode")
	}
}

func TestWriteGeneratedFileCreatesDirectories(t *testing.T) {
	target := t.TempDir()
	relPath := filepath.Join("deep", "nested", "dir", "file.md")

	err := writeGeneratedFile(target, relPath, []byte("content"), false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fullPath := filepath.Join(target, relPath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("file should exist: %v", err)
	}
	if string(data) != "content" {
		t.Errorf("file content = %q, want 'content'", data)
	}
}
