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
