package cmd

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/fatih/color"
	"github.com/willvelida/code-minions/internal/registry"
)

// ---------- Integration tests for external package sources ----------
//
// These tests exercise the full end-to-end flow using local Git repos
// as external sources. They cover:
//
//  - source add → list → remove lifecycle
//  - list --from with a named source and Git URL
//  - search --from
//  - show --from
//  - install --from (validates files are actually written)
//  - Registry builder correctly merges sources

// setupIntegrationGitRepo creates a local Git repository with a valid
// package structure for use in integration tests. It also redirects
// the GitSource cache root to a temporary directory so that tests are
// hermetic and never collide with the real user cache.
func setupIntegrationGitRepo(t *testing.T) string {
	t.Helper()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	// Redirect git source cache to an isolated temp dir so tests
	// don't read from / write to the real user cache.
	cacheDir := t.TempDir()
	registry.SetCacheRoot(cacheDir)
	t.Cleanup(func() { registry.SetCacheRoot("") })

	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "packages", "external-skill")
	if err := os.MkdirAll(filepath.Join(pkgDir, "agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(pkgDir, "skills", "external-skill"), 0o755); err != nil {
		t.Fatal(err)
	}

	// package.yaml
	manifest := `name: external-skill
version: 2.0.0
description: An externally-sourced skill for testing
author: integration-test
`
	if err := os.WriteFile(filepath.Join(pkgDir, "package.yaml"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}

	// Agent file
	agentContent := "# External Agent\nThis agent is from an external source.\n"
	if err := os.WriteFile(filepath.Join(pkgDir, "agents", "external-skill.agent.md"), []byte(agentContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Skill file
	skillContent := "---\nname: external-skill\ndescription: 'External skill for integration tests'\n---\n# External Skill\n"
	if err := os.WriteFile(filepath.Join(pkgDir, "skills", "external-skill", "SKILL.md"), []byte(skillContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Initialize Git repo
	for _, args := range [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
		{"git", "add", "."},
		{"git", "commit", "-m", "initial"},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%s failed: %v\n%s", strings.Join(args, " "), err, out)
		}
	}

	return dir
}

// TestIntegration_SourceLifecycle tests add → list → remove lifecycle.
func TestIntegration_SourceLifecycle(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	repoDir := setupIntegrationGitRepo(t)

	// Use a temporary config dir
	cfgDir := t.TempDir()
	cfgPath := filepath.Join(cfgDir, "config.yaml")

	// Add a source by directly using config API
	cfg := &registry.Config{}
	if err := cfg.AddSource(registry.SourceConfig{Name: "test-external", Type: "git", URL: repoDir}); err != nil {
		t.Fatalf("add source: %v", err)
	}
	if err := registry.SaveConfigTo(cfgPath, cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}

	// Verify source was added
	loaded, err := registry.LoadConfigFrom(cfgPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if len(loaded.Sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(loaded.Sources))
	}
	if loaded.Sources[0].Name != "test-external" {
		t.Errorf("source name = %q, want test-external", loaded.Sources[0].Name)
	}

	// Remove the source
	if err := loaded.RemoveSource("test-external"); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if len(loaded.Sources) != 0 {
		t.Errorf("expected 0 sources after remove, got %d", len(loaded.Sources))
	}
}

// TestIntegration_ListFromNamedSource tests list --from with a named
// source configured in the registry.
func TestIntegration_ListFromNamedSource(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	repoDir := setupIntegrationGitRepo(t)

	// Build a registry with the test source
	cfg := &registry.Config{
		Sources: []registry.SourceConfig{
			{Name: "test-source", Type: "git", URL: repoDir},
		},
	}

	embeddedFS := fstest.MapFS{
		"packages/embedded-pkg/agents/embedded-pkg.agent.md": &fstest.MapFile{
			Data: []byte("# Embedded Agent"),
		},
	}

	reg, err := registry.BuildRegistryFromConfig(embeddedFS, cfg)
	if err != nil {
		t.Fatalf("build registry: %v", err)
	}

	pkgs, err := reg.ListPackages()
	if err != nil {
		t.Fatalf("list packages: %v", err)
	}

	names := map[string]bool{}
	for _, p := range pkgs {
		names[p.Name] = true
	}

	// Should contain both external and embedded packages
	if !names["external-skill"] {
		t.Error("expected external-skill from git source")
	}
	if !names["embedded-pkg"] {
		t.Error("expected embedded-pkg from embedded source")
	}
}

// TestIntegration_SearchFromNamedSource tests search across a named
// source and embedded source.
func TestIntegration_SearchFromNamedSource(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	repoDir := setupIntegrationGitRepo(t)

	cfg := &registry.Config{
		Sources: []registry.SourceConfig{
			{Name: "test-source", Type: "git", URL: repoDir},
		},
	}

	embeddedFS := fstest.MapFS{
		"packages/embedded-pkg/agents/embedded-pkg.agent.md": &fstest.MapFile{
			Data: []byte("# Embedded Agent"),
		},
	}

	reg, err := registry.BuildRegistryFromConfig(embeddedFS, cfg)
	if err != nil {
		t.Fatalf("build registry: %v", err)
	}

	results, err := reg.Search("external")
	if err != nil {
		t.Fatalf("search: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("expected search results for 'external'")
	}

	found := false
	for _, r := range results {
		if r.Name == "external-skill" {
			found = true
			if r.Source != "test-source" {
				t.Errorf("expected source=test-source, got %q", r.Source)
			}
		}
	}
	if !found {
		t.Error("expected external-skill in search results")
	}
}

// TestIntegration_ShowFromNamedSource tests showing package details
// from an external source.
func TestIntegration_ShowFromNamedSource(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	repoDir := setupIntegrationGitRepo(t)

	cfg := &registry.Config{
		Sources: []registry.SourceConfig{
			{Name: "test-source", Type: "git", URL: repoDir},
		},
	}

	reg, err := registry.BuildRegistryFromConfig(fstest.MapFS{}, cfg)
	if err != nil {
		t.Fatalf("build registry: %v", err)
	}

	pkg, src, err := reg.ResolvePackage("external-skill")
	if err != nil {
		t.Fatalf("resolve package: %v", err)
	}

	if pkg.Name != "external-skill" {
		t.Errorf("name = %q, want external-skill", pkg.Name)
	}
	if pkg.Version != "2.0.0" {
		t.Errorf("version = %q, want 2.0.0", pkg.Version)
	}
	if pkg.Author != "integration-test" {
		t.Errorf("author = %q, want integration-test", pkg.Author)
	}
	if src.Name() != "test-source" {
		t.Errorf("source = %q, want test-source", src.Name())
	}
}

// TestIntegration_RegistryPriorityOrder tests that configured sources
// take priority over embedded when packages have the same name.
func TestIntegration_RegistryPriorityOrder(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	repoDir := setupIntegrationGitRepo(t)

	// Create embedded FS with a package that has the same name
	// but a different version
	embeddedFS := fstest.MapFS{
		"packages/external-skill/package.yaml": &fstest.MapFile{
			Data: []byte("name: external-skill\nversion: 1.0.0\ndescription: embedded version\n"),
		},
		"packages/external-skill/agents/external-skill.agent.md": &fstest.MapFile{
			Data: []byte("# Embedded Version Agent"),
		},
	}

	cfg := &registry.Config{
		Sources: []registry.SourceConfig{
			{Name: "team", Type: "git", URL: repoDir},
		},
	}

	reg, err := registry.BuildRegistryFromConfig(embeddedFS, cfg)
	if err != nil {
		t.Fatalf("build registry: %v", err)
	}

	// ResolvePackage should find the git source version (2.0.0) first
	pkg, src, err := reg.ResolvePackage("external-skill")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if pkg.Version != "2.0.0" {
		t.Errorf("expected version 2.0.0 from git source, got %q", pkg.Version)
	}
	if src.Name() != "team" {
		t.Errorf("expected source=team, got %q", src.Name())
	}
}

// TestIntegration_ListCommandWithFromFlag tests the actual list
// command with --from pointing to a named source.
func TestIntegration_ListCommandWithFromFlag(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	repoDir := setupIntegrationGitRepo(t)

	// We can't easily set a config for the command, so test with
	// the registry API directly but verify the JSON structure
	cfg := &registry.Config{
		Sources: []registry.SourceConfig{
			{Name: "int-test", Type: "git", URL: repoDir},
		},
	}

	// Use a minimal embedded FS with a packages/ dir so EmbeddedSource
	// doesn't error when listing packages.
	minFS := fstest.MapFS{
		"packages/.gitkeep": &fstest.MapFile{Data: []byte("")},
	}

	reg, err := registry.BuildRegistryFromConfig(minFS, cfg)
	if err != nil {
		t.Fatalf("build registry: %v", err)
	}

	pkgs, err := reg.ListPackages()
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	// Verify packages can be serialised to JSON (like the list command does)
	data, err := json.Marshal(pkgs)
	if err != nil {
		t.Fatalf("json marshal: %v", err)
	}

	var decoded []struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}

	found := false
	for _, p := range decoded {
		if p.Name == "external-skill" && p.Version == "2.0.0" {
			found = true
		}
	}
	if !found {
		t.Error("expected external-skill 2.0.0 in JSON output")
	}
}

// TestIntegration_DownloadPackageFromGitSource tests that
// DownloadPackage returns a usable filesystem.
func TestIntegration_DownloadPackageFromGitSource(t *testing.T) {
	repoDir := setupIntegrationGitRepo(t)

	src := registry.NewGitSourceWithDir("test", repoDir, t.TempDir())

	pkgFS, err := src.DownloadPackage("external-skill", "2.0.0")
	if err != nil {
		t.Fatalf("download package: %v", err)
	}

	// Should be able to read agent file from the package FS
	agentFile := "agents/external-skill.agent.md"
	data, err := pkgFS.Open(agentFile)
	if err != nil {
		t.Fatalf("open %s: %v", agentFile, err)
	}
	_ = data.Close()

	// Should be able to read skill file
	skillFile := "skills/external-skill/SKILL.md"
	sdata, err := pkgFS.Open(skillFile)
	if err != nil {
		t.Fatalf("open %s: %v", skillFile, err)
	}
	_ = sdata.Close()
}

// TestIntegration_ErrorPaths verifies proper error handling.
func TestIntegration_ErrorPaths(t *testing.T) {
	t.Run("invalid git URL in config is skipped", func(t *testing.T) {
		cfg := &registry.Config{
			Sources: []registry.SourceConfig{
				{Name: "bad", Type: "git", URL: ""},
			},
		}
		// Invalid sources are warn-and-skipped — only embedded remains
		reg, err := registry.BuildRegistryFromConfig(fstest.MapFS{}, cfg)
		if err != nil {
			t.Fatalf("expected no error (warn-and-skip), got: %v", err)
		}
		sources := reg.Sources()
		if len(sources) != 1 || sources[0].Name() != "embedded" {
			t.Errorf("expected only embedded source, got %d sources", len(sources))
		}
	})

	t.Run("unknown source type is skipped", func(t *testing.T) {
		cfg := &registry.Config{
			Sources: []registry.SourceConfig{
				{Name: "bad", Type: "http-api", URL: "https://example.com"},
			},
		}
		// Unsupported types are warn-and-skipped
		reg, err := registry.BuildRegistryFromConfig(fstest.MapFS{}, cfg)
		if err != nil {
			t.Fatalf("expected no error (warn-and-skip), got: %v", err)
		}
		sources := reg.Sources()
		if len(sources) != 1 || sources[0].Name() != "embedded" {
			t.Errorf("expected only embedded source, got %d sources", len(sources))
		}
	})

	t.Run("resolve from non-existent source", func(t *testing.T) {
		_, err := registry.ResolveFrom("does-not-exist", &registry.Config{})
		if err == nil {
			t.Fatal("expected error for non-existent source")
		}
	})

	t.Run("config validation rejects invalid names", func(t *testing.T) {
		cfg := &registry.Config{}
		err := cfg.AddSource(registry.SourceConfig{Name: "INVALID_NAME!", Type: "git", URL: "https://example.com/repo.git"})
		if err == nil {
			t.Fatal("expected validation error for invalid source name")
		}
	})

	t.Run("duplicate source name", func(t *testing.T) {
		cfg := &registry.Config{}
		if err := cfg.AddSource(registry.SourceConfig{Name: "my-source", Type: "git", URL: "https://example.com/repo.git"}); err != nil {
			t.Fatalf("first add: %v", err)
		}
		err := cfg.AddSource(registry.SourceConfig{Name: "my-source", Type: "git", URL: "https://example.com/other.git"})
		if err == nil {
			t.Fatal("expected error for duplicate source name")
		}
	})
}

// TestIntegration_SourceCommandRoundTrip tests the source command's
// add/list/remove cycle.
func TestIntegration_SourceCommandRoundTrip(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	repoDir := setupIntegrationGitRepo(t)

	// Use a temp config file
	cfgDir := t.TempDir()
	cfgPath := filepath.Join(cfgDir, "config.yaml")

	// Add
	cfg := &registry.Config{}
	if err := cfg.AddSource(registry.SourceConfig{Name: "roundtrip-test", Type: "git", URL: repoDir}); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := registry.SaveConfigTo(cfgPath, cfg); err != nil {
		t.Fatalf("save: %v", err)
	}

	// Verify
	loaded, _ := registry.LoadConfigFrom(cfgPath)
	if loaded.FindSource("roundtrip-test") == nil {
		t.Fatal("source should exist after add")
	}

	// Build registry with the config
	reg, err := registry.BuildRegistryFromConfig(fstest.MapFS{}, loaded)
	if err != nil {
		t.Fatalf("build registry: %v", err)
	}

	// Should have 2 sources (git + embedded)
	if len(reg.Sources()) != 2 {
		t.Errorf("expected 2 sources, got %d", len(reg.Sources()))
	}

	// Remove
	if err := loaded.RemoveSource("roundtrip-test"); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if err := registry.SaveConfigTo(cfgPath, loaded); err != nil {
		t.Fatalf("save after remove: %v", err)
	}

	// Verify removed
	reloaded, _ := registry.LoadConfigFrom(cfgPath)
	if reloaded.FindSource("roundtrip-test") != nil {
		t.Fatal("source should not exist after remove")
	}
}

// TestIntegration_IsGitURL verifies URL detection edge cases.
func TestIntegration_IsGitURL(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{"https://github.com/org/repo.git", true},
		{"https://github.com/org/repo", true},
		{"git@github.com:org/repo.git", true},
		{"ssh://git@github.com/org/repo.git", true},
		{"http://gitlab.com/group/project", true},
		{"not-a-url", false},
		{"", false},
		{"embedded", false},
	}

	for _, tt := range tests {
		got := registry.IsGitURL(tt.url)
		if got != tt.want {
			t.Errorf("IsGitURL(%q) = %v, want %v", tt.url, got, tt.want)
		}
	}
}

// TestIntegration_MultipleSourcesMerge verifies that packages from
// multiple sources are correctly merged by the registry.
func TestIntegration_MultipleSourcesMerge(t *testing.T) {
	repo1 := setupIntegrationGitRepo(t) // has external-skill

	// Create a second repo with a different package
	dir2 := t.TempDir()
	pkgDir2 := filepath.Join(dir2, "packages", "team-utils")
	if err := os.MkdirAll(filepath.Join(pkgDir2, "agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir2, "package.yaml"),
		[]byte("name: team-utils\nversion: 1.0.0\ndescription: Team utilities\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir2, "agents", "team-utils.agent.md"),
		[]byte("# Team Utils Agent"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
		{"git", "add", "."},
		{"git", "commit", "-m", "initial"},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir2
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %s", err, out)
		}
	}

	embeddedFS := fstest.MapFS{
		"packages/built-in/agents/built-in.agent.md": &fstest.MapFile{
			Data: []byte("# Built-in Agent"),
		},
	}

	cfg := &registry.Config{
		Sources: []registry.SourceConfig{
			{Name: "primary", Type: "git", URL: repo1},
			{Name: "secondary", Type: "git", URL: dir2},
		},
	}

	reg, err := registry.BuildRegistryFromConfig(embeddedFS, cfg)
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	pkgs, err := reg.ListPackages()
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	names := map[string]bool{}
	for _, p := range pkgs {
		names[p.Name] = true
	}

	for _, want := range []string{"external-skill", "team-utils", "built-in"} {
		if !names[want] {
			t.Errorf("missing package %q from merged list", want)
		}
	}

	// Verify source count: primary + secondary + embedded = 3
	if len(reg.Sources()) != 3 {
		t.Errorf("expected 3 sources, got %d", len(reg.Sources()))
	}
}

// ---------- Integration tests for install --from ----------

// setupIntegrationGitRepoWithMCP creates a local Git repository with
// a package that includes an mcp.yaml file for MCP server testing.
func setupIntegrationGitRepoWithMCP(t *testing.T) string {
	t.Helper()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	cacheDir := t.TempDir()
	registry.SetCacheRoot(cacheDir)
	t.Cleanup(func() { registry.SetCacheRoot("") })

	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "packages", "mcp-skill")
	if err := os.MkdirAll(filepath.Join(pkgDir, "agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(pkgDir, "skills", "mcp-skill"), 0o755); err != nil {
		t.Fatal(err)
	}

	// package.yaml
	manifest := "name: mcp-skill\nversion: 1.0.0\ndescription: MCP test skill\n"
	if err := os.WriteFile(filepath.Join(pkgDir, "package.yaml"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "agents", "mcp-skill.agent.md"),
		[]byte("# MCP Agent\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "skills", "mcp-skill", "SKILL.md"),
		[]byte("# MCP Skill\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// mcp.yaml
	mcpYAML := `servers:
  test-server:
    transport: stdio
    command: node
    args: ["server.js"]
`
	if err := os.WriteFile(filepath.Join(pkgDir, "mcp.yaml"), []byte(mcpYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	for _, args := range [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
		{"git", "add", "."},
		{"git", "commit", "-m", "initial"},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%s failed: %v\n%s", strings.Join(args, " "), err, out)
		}
	}

	return dir
}

// TestIntegration_InstallFromGitURL tests installing a package from
// a local Git repository URL using --from.
func TestIntegration_InstallFromGitURL(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	repoDir := setupIntegrationGitRepo(t)
	target := t.TempDir()

	cmd := newInstallCommand(fstest.MapFS{})
	cmd.SetArgs([]string{
		"--package", "external-skill",
		"--from", repoDir,
		"--target", target,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("install --from failed: %v", err)
	}

	// Verify agent file was installed
	agentFile := filepath.Join(target, "agents", "external-skill.agent.md")
	if _, err := os.Stat(agentFile); os.IsNotExist(err) {
		t.Errorf("expected agent file at %s", agentFile)
	}

	// Verify skill file was installed
	skillFile := filepath.Join(target, "skills", "external-skill", "SKILL.md")
	if _, err := os.Stat(skillFile); os.IsNotExist(err) {
		t.Errorf("expected skill file at %s", skillFile)
	}

	// Verify AGENTS.md was created
	agentsMD := filepath.Join(target, "AGENTS.md")
	if _, err := os.Stat(agentsMD); os.IsNotExist(err) {
		t.Errorf("expected AGENTS.md at %s", agentsMD)
	}
}

// TestIntegration_InstallFromGitURLForCopilot tests that --from with
// --for copilot places files in the correct assistant directory.
func TestIntegration_InstallFromGitURLForCopilot(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	repoDir := setupIntegrationGitRepo(t)
	target := t.TempDir()

	cmd := newInstallCommand(fstest.MapFS{})
	cmd.SetArgs([]string{
		"--package", "external-skill",
		"--from", repoDir,
		"--for", "copilot",
		"--target", target,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("install --from --for copilot failed: %v", err)
	}

	// Agent should be in .github/agents/ (Copilot path mapping)
	copilotAgent := filepath.Join(target, ".github", "agents", "external-skill.agent.md")
	if _, err := os.Stat(copilotAgent); os.IsNotExist(err) {
		t.Errorf("expected agent at Copilot path: %s", copilotAgent)
	}

	// Skill should remain at skills/ (Copilot doesn't remap skills)
	skillFile := filepath.Join(target, "skills", "external-skill", "SKILL.md")
	if _, err := os.Stat(skillFile); os.IsNotExist(err) {
		t.Errorf("expected skill at default path: %s", skillFile)
	}
}

// TestIntegration_InstallFromMultiplePackages tests installing multiple
// comma-separated packages from an external source.
func TestIntegration_InstallFromMultiplePackages(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	// Create a repo with two packages
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	cacheDir := t.TempDir()
	registry.SetCacheRoot(cacheDir)
	t.Cleanup(func() { registry.SetCacheRoot("") })

	dir := t.TempDir()
	for _, pkgName := range []string{"skill-a", "skill-b"} {
		pkgDir := filepath.Join(dir, "packages", pkgName)
		if err := os.MkdirAll(filepath.Join(pkgDir, "agents"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(pkgDir, "package.yaml"),
			[]byte("name: "+pkgName+"\nversion: 1.0.0\ndescription: "+pkgName+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(pkgDir, "agents", pkgName+".agent.md"),
			[]byte("# "+pkgName+" Agent\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	for _, args := range [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
		{"git", "add", "."},
		{"git", "commit", "-m", "initial"},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %s", err, out)
		}
	}

	target := t.TempDir()

	cmd := newInstallCommand(fstest.MapFS{})
	cmd.SetArgs([]string{
		"--package", "skill-a,skill-b",
		"--from", dir,
		"--target", target,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Both packages should be installed
	for _, pkgName := range []string{"skill-a", "skill-b"} {
		agentFile := filepath.Join(target, "agents", pkgName+".agent.md")
		if _, err := os.Stat(agentFile); os.IsNotExist(err) {
			t.Errorf("expected %s agent file at %s", pkgName, agentFile)
		}
	}
}

// TestIntegration_InstallFromDryRun tests that --dry-run with --from
// does not write any files.
func TestIntegration_InstallFromDryRun(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	repoDir := setupIntegrationGitRepo(t)
	target := t.TempDir()

	cmd := newInstallCommand(fstest.MapFS{})
	cmd.SetArgs([]string{
		"--package", "external-skill",
		"--from", repoDir,
		"--target", target,
		"--dry-run",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("install --dry-run failed: %v", err)
	}

	// No files should be written
	agentFile := filepath.Join(target, "agents", "external-skill.agent.md")
	if _, err := os.Stat(agentFile); !os.IsNotExist(err) {
		t.Errorf("dry-run should not create files, but found: %s", agentFile)
	}
}

// TestIntegration_InstallFromForce tests that --force overwrites
// existing files when installing from an external source.
func TestIntegration_InstallFromForce(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	repoDir := setupIntegrationGitRepo(t)
	target := t.TempDir()

	// Pre-create the agent file with different content
	agentDir := filepath.Join(target, "agents")
	if err := os.MkdirAll(agentDir, 0o755); err != nil {
		t.Fatal(err)
	}
	agentFile := filepath.Join(agentDir, "external-skill.agent.md")
	if err := os.WriteFile(agentFile, []byte("old content"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Install without --force: should skip
	cmd1 := newInstallCommand(fstest.MapFS{})
	cmd1.SetArgs([]string{
		"--package", "external-skill",
		"--from", repoDir,
		"--target", target,
	})
	if err := cmd1.Execute(); err != nil {
		t.Fatalf("first install failed: %v", err)
	}

	// Content should still be old
	data, _ := os.ReadFile(agentFile)
	if string(data) != "old content" {
		t.Error("expected file to be skipped (old content preserved)")
	}

	// Install with --force: should overwrite
	cmd2 := newInstallCommand(fstest.MapFS{})
	cmd2.SetArgs([]string{
		"--package", "external-skill",
		"--from", repoDir,
		"--target", target,
		"--force",
	})
	if err := cmd2.Execute(); err != nil {
		t.Fatalf("force install failed: %v", err)
	}

	// Content should now be from external source
	data, _ = os.ReadFile(agentFile)
	if string(data) == "old content" {
		t.Error("expected --force to overwrite file content")
	}
}

// TestIntegration_InstallFromPackageNotFound tests the error when a
// package doesn't exist in the external source.
func TestIntegration_InstallFromPackageNotFound(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	repoDir := setupIntegrationGitRepo(t)

	cmd := newInstallCommand(fstest.MapFS{})
	cmd.SetArgs([]string{
		"--package", "nonexistent-package",
		"--from", repoDir,
		"--target", t.TempDir(),
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for nonexistent package")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "nonexistent-package") {
		t.Errorf("error should mention the package name, got: %s", errMsg)
	}
	// Should list available packages
	if !strings.Contains(errMsg, "external-skill") {
		t.Errorf("error should list available packages, got: %s", errMsg)
	}
}

// TestIntegration_InstallFromRecordsManifest tests that the install
// manifest correctly records the source name for remote installs.
func TestIntegration_InstallFromRecordsManifest(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	repoDir := setupIntegrationGitRepo(t)
	target := t.TempDir()

	cmd := newInstallCommand(fstest.MapFS{})
	cmd.SetArgs([]string{
		"--package", "external-skill",
		"--from", repoDir,
		"--target", target,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Read the install manifest
	manifestPath := filepath.Join(target, ".code-minions", "installed.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("failed to read manifest: %v", err)
	}

	// Parse and verify
	var m struct {
		Packages []struct {
			Name    string `json:"name"`
			Version string `json:"version"`
			Source  string `json:"source"`
		} `json:"packages"`
	}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("failed to parse manifest: %v", err)
	}

	if len(m.Packages) != 1 {
		t.Fatalf("expected 1 package in manifest, got %d", len(m.Packages))
	}

	pkg := m.Packages[0]
	if pkg.Name != "external-skill" {
		t.Errorf("manifest package name = %q, want external-skill", pkg.Name)
	}
	if pkg.Version != "2.0.0" {
		t.Errorf("manifest version = %q, want 2.0.0", pkg.Version)
	}
	// Source should NOT be "embedded"
	if pkg.Source == "embedded" || pkg.Source == "" {
		t.Errorf("manifest source should not be empty or 'embedded', got %q", pkg.Source)
	}
}

// TestIntegration_InstallFromUpdatesProjectManifest tests that the
// project manifest (code-minions.yml) is updated after remote install.
func TestIntegration_InstallFromUpdatesProjectManifest(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	repoDir := setupIntegrationGitRepo(t)
	target := t.TempDir()

	cmd := newInstallCommand(fstest.MapFS{})
	cmd.SetArgs([]string{
		"--package", "external-skill",
		"--from", repoDir,
		"--target", target,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Read code-minions.yml
	manifestPath := filepath.Join(target, "code-minions.yml")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("failed to read project manifest: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "external-skill") {
		t.Errorf("project manifest should contain 'external-skill', got:\n%s", content)
	}
}

// TestIntegration_InstallFromJSONOutput tests JSON output mode with
// remote package installation.
func TestIntegration_InstallFromJSONOutput(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	repoDir := setupIntegrationGitRepo(t)
	target := t.TempDir()

	// Use root command so the persistent --json flag is available
	cmd := NewRootCommand(fstest.MapFS{})
	var buf strings.Builder
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{
		"install",
		"--package", "external-skill",
		"--from", repoDir,
		"--target", target,
		"--json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("install --json failed: %v", err)
	}

	// Parse JSON output
	var result struct {
		Copied  []string `json:"copied"`
		Skipped []string `json:"skipped"`
		Errors  []string `json:"errors"`
		Summary struct {
			Copied  int `json:"copied"`
			Skipped int `json:"skipped"`
			Errors  int `json:"errors"`
		} `json:"summary"`
	}
	if err := json.Unmarshal([]byte(buf.String()), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, buf.String())
	}

	if result.Summary.Copied == 0 {
		t.Error("expected at least one copied file")
	}
	if result.Summary.Errors != 0 {
		t.Errorf("expected 0 errors, got %d: %v", result.Summary.Errors, result.Errors)
	}
}

// TestIntegration_InstallFromMCPProcessing tests that MCP servers in
// remote packages are processed correctly.
func TestIntegration_InstallFromMCPProcessing(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	repoDir := setupIntegrationGitRepoWithMCP(t)
	target := t.TempDir()

	cmd := newInstallCommand(fstest.MapFS{})
	cmd.SetArgs([]string{
		"--package", "mcp-skill",
		"--from", repoDir,
		"--for", "copilot",
		"--target", target,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("install with MCP failed: %v", err)
	}

	// Agent file should be installed
	agentFile := filepath.Join(target, ".github", "agents", "mcp-skill.agent.md")
	if _, err := os.Stat(agentFile); os.IsNotExist(err) {
		t.Errorf("expected agent file at %s", agentFile)
	}

	// MCP config should be written
	mcpConfig := filepath.Join(target, ".vscode", "mcp.json")
	if _, err := os.Stat(mcpConfig); os.IsNotExist(err) {
		t.Errorf("expected MCP config at %s", mcpConfig)
	}

	// MCP config should contain the test-server
	if data, err := os.ReadFile(mcpConfig); err == nil {
		if !strings.Contains(string(data), "test-server") {
			t.Errorf("MCP config should contain 'test-server', got:\n%s", string(data))
		}
	}
}

// TestIntegration_InstallFromForClaude tests that --from with
// --for claude places files in Claude directories and generates CLAUDE.md.
func TestIntegration_InstallFromForClaude(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	repoDir := setupIntegrationGitRepo(t)
	target := t.TempDir()

	cmd := newInstallCommand(fstest.MapFS{})
	cmd.SetArgs([]string{
		"--package", "external-skill",
		"--from", repoDir,
		"--for", "claude",
		"--target", target,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("install --from --for claude failed: %v", err)
	}

	// Agent should be in .claude/agents/ (Claude path mapping)
	claudeAgent := filepath.Join(target, ".claude", "agents", "external-skill.agent.md")
	if _, err := os.Stat(claudeAgent); os.IsNotExist(err) {
		t.Errorf("expected agent at Claude path: %s", claudeAgent)
	}

	// Skill should be in .claude/skills/ (Claude remaps skills)
	skillFile := filepath.Join(target, ".claude", "skills", "external-skill", "SKILL.md")
	if _, err := os.Stat(skillFile); os.IsNotExist(err) {
		t.Errorf("expected skill at Claude path: %s", skillFile)
	}

	// CLAUDE.md should be generated
	claudeMD := filepath.Join(target, "CLAUDE.md")
	if _, err := os.Stat(claudeMD); os.IsNotExist(err) {
		t.Errorf("expected CLAUDE.md at %s", claudeMD)
	} else {
		data, _ := os.ReadFile(claudeMD)
		content := string(data)
		if !strings.Contains(content, "external-skill") {
			t.Errorf("CLAUDE.md should reference 'external-skill', got:\n%s", content)
		}
	}
}

// TestIntegration_InstallFromUnreachableSource tests error handling when
// the --from source is an unreachable URL.
func TestIntegration_InstallFromUnreachableSource(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	// Redirect cache to avoid contamination
	cacheDir := t.TempDir()
	registry.SetCacheRoot(cacheDir)
	t.Cleanup(func() { registry.SetCacheRoot("") })

	cmd := newInstallCommand(fstest.MapFS{})
	cmd.SetArgs([]string{
		"--package", "some-pkg",
		"--from", "https://github.com/nonexistent-org-12345/nonexistent-repo-67890.git",
		"--target", t.TempDir(),
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for unreachable source")
	}
}

// TestIntegration_InstallFromWithoutPackage tests that --from without
// --package produces a clear error message.
func TestIntegration_InstallFromWithoutPackage(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	cmd := newInstallCommand(fstest.MapFS{})
	cmd.SetArgs([]string{
		"--from", "some-source",
		"--target", t.TempDir(),
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when --from is used without --package")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "--from requires --package") {
		t.Errorf("error should mention '--from requires --package', got: %s", errMsg)
	}
}

// TestIntegration_InstallFromNamedSource tests the full lifecycle of
// configuring a named source and then installing from it.
func TestIntegration_InstallFromNamedSource(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	repoDir := setupIntegrationGitRepo(t)

	// Build a registry with a named source config
	cfg := &registry.Config{
		Sources: []registry.SourceConfig{
			{Name: "team-source", Type: "git", URL: repoDir},
		},
	}

	// Use BuildRegistryWithFromAndConfig to resolve the named source
	reg, err := registry.BuildRegistryWithFromAndConfig(fstest.MapFS{}, "team-source", cfg)
	if err != nil {
		t.Fatalf("failed to resolve named source: %v", err)
	}

	// Verify the registry can find the package
	pkg, src, err := reg.ResolvePackage("external-skill")
	if err != nil {
		t.Fatalf("failed to resolve package: %v", err)
	}
	if pkg.Name != "external-skill" {
		t.Errorf("package name = %q, want external-skill", pkg.Name)
	}
	if src.Name() != "team-source" {
		t.Errorf("source name = %q, want team-source", src.Name())
	}

	// Download and verify the package
	pkgFS, err := src.DownloadPackage("external-skill", pkg.Version)
	if err != nil {
		t.Fatalf("failed to download package: %v", err)
	}

	// Verify the package FS contains expected files
	agentFile, err := pkgFS.Open("agents/external-skill.agent.md")
	if err != nil {
		t.Fatalf("package FS missing agent file: %v", err)
	}
	_ = agentFile.Close()

	// Now do a full install using the install command directly
	// This tests the complete path through installFromSource
	target2 := t.TempDir()
	cmd := newInstallCommand(fstest.MapFS{})
	cmd.SetArgs([]string{
		"--package", "external-skill",
		"--from", repoDir,
		"--target", target2,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("install --from named source failed: %v", err)
	}

	installed := filepath.Join(target2, "agents", "external-skill.agent.md")
	if _, err := os.Stat(installed); os.IsNotExist(err) {
		t.Errorf("expected installed agent at %s", installed)
	}
}

// setupIntegrationGitRepoWithPersona creates a local Git repository
// that includes both packages and a persona for testing persona --from.
func setupIntegrationGitRepoWithPersona(t *testing.T) string {
	t.Helper()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	cacheDir := t.TempDir()
	registry.SetCacheRoot(cacheDir)
	t.Cleanup(func() { registry.SetCacheRoot("") })

	dir := t.TempDir()

	// Create two packages
	for _, pkgName := range []string{"pkg-alpha", "pkg-beta"} {
		pkgDir := filepath.Join(dir, "packages", pkgName)
		if err := os.MkdirAll(filepath.Join(pkgDir, "agents"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(pkgDir, "skills", pkgName), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(pkgDir, "package.yaml"),
			[]byte("name: "+pkgName+"\nversion: 1.0.0\ndescription: "+pkgName+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(pkgDir, "agents", pkgName+".agent.md"),
			[]byte("# "+pkgName+" Agent\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(pkgDir, "skills", pkgName, "SKILL.md"),
			[]byte("# "+pkgName+" Skill\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// Create a persona
	personaDir := filepath.Join(dir, "personas", "test-remote-persona")
	if err := os.MkdirAll(personaDir, 0o755); err != nil {
		t.Fatal(err)
	}
	personaYAML := `name: test-remote-persona
description: A test persona from a remote source
packages:
  - name: pkg-alpha
  - name: pkg-beta
`
	if err := os.WriteFile(filepath.Join(personaDir, "persona.yaml"), []byte(personaYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	for _, args := range [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
		{"git", "add", "."},
		{"git", "commit", "-m", "initial"},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%s failed: %v\n%s", strings.Join(args, " "), err, out)
		}
	}

	return dir
}

// TestIntegration_PersonaInstallWithFrom tests installing a persona
// from an external source using --from.
func TestIntegration_PersonaInstallWithFrom(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	repoDir := setupIntegrationGitRepoWithPersona(t)
	target := t.TempDir()

	cmd := newInstallCommand(fstest.MapFS{})
	cmd.SetArgs([]string{
		"--persona", "test-remote-persona",
		"--from", repoDir,
		"--for", "copilot",
		"--target", target,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("persona install --from failed: %v", err)
	}

	// Both packages from the persona should be installed
	for _, pkgName := range []string{"pkg-alpha", "pkg-beta"} {
		agentFile := filepath.Join(target, ".github", "agents", pkgName+".agent.md")
		if _, err := os.Stat(agentFile); os.IsNotExist(err) {
			t.Errorf("expected persona package %s agent at %s", pkgName, agentFile)
		}
	}
}

// TestIntegration_InstallFromMinionsHub tests installing from the real
// MinionsHub repository. This test requires network access and is
// skipped when the INTEGRATION environment variable is not set.
func TestIntegration_InstallFromMinionsHub(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("skipping network integration test; set INTEGRATION=1 to run")
	}

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	cacheDir := t.TempDir()
	registry.SetCacheRoot(cacheDir)
	t.Cleanup(func() { registry.SetCacheRoot("") })

	target := t.TempDir()

	cmd := newInstallCommand(fstest.MapFS{})
	cmd.SetArgs([]string{
		"--package", "developer-mentor",
		"--from", "github.com/willvelida/MinionsHub",
		"--for", "copilot",
		"--target", target,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("MinionsHub install failed: %v", err)
	}

	// Verify at least the agent file is present
	agentFile := filepath.Join(target, ".github", "agents", "developer-mentor.agent.md")
	if _, err := os.Stat(agentFile); os.IsNotExist(err) {
		t.Errorf("expected developer-mentor agent at %s", agentFile)
	}
}

// TestIntegration_InstallFromQuietMode tests that --quiet with --from
// suppresses normal output but still reports errors.
func TestIntegration_InstallFromQuietMode(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	repoDir := setupIntegrationGitRepo(t)
	target := t.TempDir()

	// Use root command for the persistent --quiet flag
	cmd := NewRootCommand(fstest.MapFS{})
	var stdout strings.Builder
	var stderr strings.Builder
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{
		"install",
		"--package", "external-skill",
		"--from", repoDir,
		"--target", target,
		"--quiet",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("install --quiet failed: %v", err)
	}

	// Quiet mode should produce no stdout
	if stdout.String() != "" {
		t.Errorf("quiet mode should produce no stdout, got:\n%s", stdout.String())
	}

	// Files should still be installed
	agentFile := filepath.Join(target, "agents", "external-skill.agent.md")
	if _, err := os.Stat(agentFile); os.IsNotExist(err) {
		t.Errorf("expected agent file at %s even in quiet mode", agentFile)
	}
}

// TestIntegration_InstallFromVerboseMode tests that --verbose with --from
// produces additional output.
func TestIntegration_InstallFromVerboseMode(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	repoDir := setupIntegrationGitRepo(t)
	target := t.TempDir()

	// Use root command for the persistent --verbose flag
	cmd := NewRootCommand(fstest.MapFS{})
	var buf strings.Builder
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{
		"install",
		"--package", "external-skill",
		"--from", repoDir,
		"--target", target,
		"--verbose",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("install --verbose failed: %v", err)
	}

	output := buf.String()

	// Verbose mode should include verbose-specific output like "updated code-minions.yml"
	// Note: "copied" messages go through color.Printf (os.Stdout) not cmd.OutOrStdout(),
	// so we check for the verbose-specific output that goes through verbosePrintf.
	if !strings.Contains(output, "updated code-minions.yml") {
		t.Errorf("verbose output should contain 'updated code-minions.yml', got:\n%s", output)
	}

	// Files should be installed
	agentFile := filepath.Join(target, "agents", "external-skill.agent.md")
	if _, err := os.Stat(agentFile); os.IsNotExist(err) {
		t.Errorf("expected agent file at %s", agentFile)
	}
}
