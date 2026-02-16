package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestInstall_NewFile(t *testing.T) {
	content := fstest.MapFS{
		"packages/test-pkg/mcp.yaml": &fstest.MapFile{
			Data: []byte(`servers:
  github:
    transport: stdio
    command: npx
    args: ["-y", "@modelcontextprotocol/server-github"]
    env:
      GITHUB_TOKEN: "my-token"
`),
		},
	}

	target := t.TempDir()
	translator := &CopilotTranslator{}

	result, err := Install(content, "packages/test-pkg", target, translator, false, false)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if result == nil {
		t.Fatal("Install() returned nil result")
	}
	if result.ConfigPath != ".vscode/mcp.json" {
		t.Errorf("ConfigPath = %q, want %q", result.ConfigPath, ".vscode/mcp.json")
	}
	if len(result.Merge.Added) != 1 || result.Merge.Added[0] != "github" {
		t.Errorf("Added = %v, want [github]", result.Merge.Added)
	}
	if len(result.ServerNames) != 1 || result.ServerNames[0] != "github" {
		t.Errorf("ServerNames = %v, want [github]", result.ServerNames)
	}

	// Verify file was written
	data, err := os.ReadFile(filepath.Join(target, ".vscode", "mcp.json"))
	if err != nil {
		t.Fatalf("reading written config: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("parsing written config: %v", err)
	}
	servers, ok := doc["servers"].(map[string]any)
	if !ok {
		t.Fatal("written config missing 'servers' key")
	}
	if _, ok := servers["github"]; !ok {
		t.Error("written config missing 'github' server")
	}
}

func TestInstall_NoMCPYaml(t *testing.T) {
	content := fstest.MapFS{
		"packages/test-pkg/package.yaml": &fstest.MapFile{
			Data: []byte("name: test-pkg\n"),
		},
	}

	target := t.TempDir()
	translator := &CopilotTranslator{}

	result, err := Install(content, "packages/test-pkg", target, translator, false, false)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if result != nil {
		t.Errorf("Install() = %v, want nil (no mcp.yaml)", result)
	}
}

func TestInstall_DryRunDoesNotWrite(t *testing.T) {
	content := fstest.MapFS{
		"packages/test-pkg/mcp.yaml": &fstest.MapFile{
			Data: []byte(`servers:
  github:
    transport: stdio
    command: npx
`),
		},
	}

	target := t.TempDir()
	translator := &CopilotTranslator{}

	result, err := Install(content, "packages/test-pkg", target, translator, false, true)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if result == nil {
		t.Fatal("Install() returned nil result")
	}
	if !result.DryRun {
		t.Error("DryRun should be true")
	}

	// Verify file was NOT written
	configPath := filepath.Join(target, ".vscode", "mcp.json")
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Error("dry-run should not have created the config file")
	}
}

func TestInstall_MergesWithExisting(t *testing.T) {
	content := fstest.MapFS{
		"packages/test-pkg/mcp.yaml": &fstest.MapFile{
			Data: []byte(`servers:
  new-server:
    transport: stdio
    command: my-cmd
`),
		},
	}

	target := t.TempDir()
	translator := &CopilotTranslator{}

	// Create an existing config with a different server
	configDir := filepath.Join(target, ".vscode")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	existing := `{
  "servers": {
    "existing": {
      "command": "other"
    }
  }
}
`
	if err := os.WriteFile(filepath.Join(configDir, "mcp.json"), []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Install(content, "packages/test-pkg", target, translator, false, false)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if result == nil {
		t.Fatal("Install() returned nil")
	}
	if len(result.Merge.Added) != 1 || result.Merge.Added[0] != "new-server" {
		t.Errorf("Added = %v, want [new-server]", result.Merge.Added)
	}

	// Verify both servers exist in the written file
	data, err := os.ReadFile(filepath.Join(target, ".vscode", "mcp.json"))
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	servers := doc["servers"].(map[string]any)
	if _, ok := servers["existing"]; !ok {
		t.Error("existing server was lost during merge")
	}
	if _, ok := servers["new-server"]; !ok {
		t.Error("new server was not added")
	}
}

func TestInstall_ConflictNoForce(t *testing.T) {
	content := fstest.MapFS{
		"packages/test-pkg/mcp.yaml": &fstest.MapFile{
			Data: []byte(`servers:
  github:
    transport: stdio
    command: new-cmd
`),
		},
	}

	target := t.TempDir()
	translator := &CopilotTranslator{}

	// Create existing config with same server name but different config
	configDir := filepath.Join(target, ".vscode")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	existing := `{
  "servers": {
    "github": {
      "command": "old-cmd"
    }
  }
}
`
	if err := os.WriteFile(filepath.Join(configDir, "mcp.json"), []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Install(content, "packages/test-pkg", target, translator, false, false)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if len(result.Merge.Conflict) != 1 || result.Merge.Conflict[0] != "github" {
		t.Errorf("Conflict = %v, want [github]", result.Merge.Conflict)
	}
	if len(result.Merge.Added) != 0 {
		t.Errorf("Added = %v, want empty", result.Merge.Added)
	}
}

func TestInstall_ConflictWithForce(t *testing.T) {
	content := fstest.MapFS{
		"packages/test-pkg/mcp.yaml": &fstest.MapFile{
			Data: []byte(`servers:
  github:
    transport: stdio
    command: new-cmd
`),
		},
	}

	target := t.TempDir()
	translator := &CopilotTranslator{}

	// Create existing config with same server name but different config
	configDir := filepath.Join(target, ".vscode")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	existing := `{
  "servers": {
    "github": {
      "command": "old-cmd"
    }
  }
}
`
	if err := os.WriteFile(filepath.Join(configDir, "mcp.json"), []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Install(content, "packages/test-pkg", target, translator, true, false)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if len(result.Merge.Added) != 1 || result.Merge.Added[0] != "github" {
		t.Errorf("Added = %v, want [github]", result.Merge.Added)
	}

	// Verify the server was overwritten
	data, err := os.ReadFile(filepath.Join(target, ".vscode", "mcp.json"))
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	servers := doc["servers"].(map[string]any)
	gh := servers["github"].(map[string]any)
	if gh["command"] != "new-cmd" {
		t.Errorf("command = %v, want new-cmd", gh["command"])
	}
}

func TestInstall_ClaudeTranslator(t *testing.T) {
	content := fstest.MapFS{
		"packages/test-pkg/mcp.yaml": &fstest.MapFile{
			Data: []byte(`servers:
  github:
    transport: stdio
    command: npx
    args: ["-y", "@modelcontextprotocol/server-github"]
`),
		},
	}

	target := t.TempDir()
	translator := &ClaudeTranslator{}

	result, err := Install(content, "packages/test-pkg", target, translator, false, false)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if result.ConfigPath != ".claude/settings.local.json" {
		t.Errorf("ConfigPath = %q, want .claude/settings.local.json", result.ConfigPath)
	}

	// Verify file was written at Claude's path
	data, err := os.ReadFile(filepath.Join(target, ".claude", "settings.local.json"))
	if err != nil {
		t.Fatalf("reading Claude config: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	if _, ok := doc["mcpServers"]; !ok {
		t.Error("Claude config missing 'mcpServers' key")
	}
}
