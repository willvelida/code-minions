package mcp

import (
	"testing"
	"testing/fstest"
)

func TestLoadConfig_ValidPackage(t *testing.T) {
	fs := fstest.MapFS{
		"packages/git-workflow/mcp.yaml": &fstest.MapFile{
			Data: []byte(`
servers:
  github:
    transport: stdio
    command: npx
    args: ["-y", "@modelcontextprotocol/server-github"]
    env:
      GITHUB_PERSONAL_ACCESS_TOKEN: ""
`),
		},
	}

	cfg, err := LoadConfig(fs, "packages/git-workflow")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if _, ok := cfg.Servers["github"]; !ok {
		t.Error("expected server 'github'")
	}
}

func TestLoadConfig_NoMCPFile(t *testing.T) {
	fs := fstest.MapFS{
		"packages/basic/agents/basic.agent.md": &fstest.MapFile{Data: []byte("# Agent")},
	}

	cfg, err := LoadConfig(fs, "packages/basic")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg != nil {
		t.Error("expected nil config when no mcp.yaml exists")
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	fs := fstest.MapFS{
		"packages/bad/mcp.yaml": &fstest.MapFile{
			Data: []byte("not: [valid yaml"),
		},
	}

	_, err := LoadConfig(fs, "packages/bad")
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestLoadConfig_ValidationError(t *testing.T) {
	fs := fstest.MapFS{
		"packages/bad/mcp.yaml": &fstest.MapFile{
			Data: []byte(`
servers:
  broken:
    command: npx
`),
		},
	}

	_, err := LoadConfig(fs, "packages/bad")
	if err == nil {
		t.Fatal("expected validation error")
	}
}
