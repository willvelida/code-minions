package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/BurntSushi/toml"
)

func TestUninstall_RemovesServers(t *testing.T) {
	target := t.TempDir()
	translator := &CopilotTranslator{}

	// Create existing config with two servers
	configDir := filepath.Join(target, ".vscode")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	existing := `{
  "servers": {
    "github": {
      "command": "npx"
    },
    "other": {
      "command": "other-cmd"
    }
  }
}
`
	if err := os.WriteFile(filepath.Join(configDir, "mcp.json"), []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Uninstall(target, translator, []string{"github"}, false)
	if err != nil {
		t.Fatalf("Uninstall() error = %v", err)
	}
	if len(result.Removed) != 1 || result.Removed[0] != "github" {
		t.Errorf("Removed = %v, want [github]", result.Removed)
	}

	// Verify file was updated
	data, err := os.ReadFile(filepath.Join(configDir, "mcp.json"))
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	servers := doc["servers"].(map[string]any)
	if _, ok := servers["github"]; ok {
		t.Error("github server should have been removed")
	}
	if _, ok := servers["other"]; !ok {
		t.Error("other server should still be present")
	}
}

func TestUninstall_RemovesKeyWhenEmpty(t *testing.T) {
	target := t.TempDir()
	translator := &CopilotTranslator{}

	configDir := filepath.Join(target, ".vscode")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	existing := `{
  "servers": {
    "github": {
      "command": "npx"
    }
  },
  "other-key": "preserved"
}
`
	if err := os.WriteFile(filepath.Join(configDir, "mcp.json"), []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Uninstall(target, translator, []string{"github"}, false)
	if err != nil {
		t.Fatalf("Uninstall() error = %v", err)
	}
	if len(result.Removed) != 1 {
		t.Errorf("Removed = %v, want [github]", result.Removed)
	}

	// Verify the servers key was removed but other keys preserved
	data, err := os.ReadFile(filepath.Join(configDir, "mcp.json"))
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	if _, ok := doc["servers"]; ok {
		t.Error("empty servers key should have been removed")
	}
	if _, ok := doc["other-key"]; !ok {
		t.Error("other-key should be preserved")
	}
}

func TestUninstall_NoConfigFile(t *testing.T) {
	target := t.TempDir()
	translator := &CopilotTranslator{}

	result, err := Uninstall(target, translator, []string{"github"}, false)
	if err != nil {
		t.Fatalf("Uninstall() error = %v", err)
	}
	if result != nil {
		t.Errorf("Uninstall() = %v, want nil (no config file)", result)
	}
}

func TestUninstall_ServerNotFound(t *testing.T) {
	target := t.TempDir()
	translator := &CopilotTranslator{}

	configDir := filepath.Join(target, ".vscode")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	existing := `{
  "servers": {
    "other": {
      "command": "cmd"
    }
  }
}
`
	if err := os.WriteFile(filepath.Join(configDir, "mcp.json"), []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Uninstall(target, translator, []string{"github"}, false)
	if err != nil {
		t.Fatalf("Uninstall() error = %v", err)
	}
	if len(result.NotFound) != 1 || result.NotFound[0] != "github" {
		t.Errorf("NotFound = %v, want [github]", result.NotFound)
	}
	if len(result.Removed) != 0 {
		t.Errorf("Removed = %v, want empty", result.Removed)
	}
}

func TestUninstall_DryRunDoesNotWrite(t *testing.T) {
	target := t.TempDir()
	translator := &CopilotTranslator{}

	configDir := filepath.Join(target, ".vscode")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	existing := `{
  "servers": {
    "github": {
      "command": "npx"
    }
  }
}
`
	if err := os.WriteFile(filepath.Join(configDir, "mcp.json"), []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Uninstall(target, translator, []string{"github"}, true)
	if err != nil {
		t.Fatalf("Uninstall() error = %v", err)
	}
	if len(result.Removed) != 1 {
		t.Errorf("Removed = %v, want [github]", result.Removed)
	}

	// Verify file was NOT modified
	data, err := os.ReadFile(filepath.Join(configDir, "mcp.json"))
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	servers := doc["servers"].(map[string]any)
	if _, ok := servers["github"]; !ok {
		t.Error("dry-run should not have removed the server")
	}
}

func TestUninstall_EmptyServerNames(t *testing.T) {
	target := t.TempDir()
	translator := &CopilotTranslator{}

	result, err := Uninstall(target, translator, nil, false)
	if err != nil {
		t.Fatalf("Uninstall() error = %v", err)
	}
	if result != nil {
		t.Errorf("Uninstall() = %v, want nil (no server names)", result)
	}
}

func TestUninstall_ClaudeTranslator(t *testing.T) {
	target := t.TempDir()
	translator := &ClaudeTranslator{}

	configDir := filepath.Join(target, ".claude")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	existing := `{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"]
    }
  }
}
`
	if err := os.WriteFile(filepath.Join(configDir, "settings.local.json"), []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Uninstall(target, translator, []string{"github"}, false)
	if err != nil {
		t.Fatalf("Uninstall() error = %v", err)
	}
	if len(result.Removed) != 1 || result.Removed[0] != "github" {
		t.Errorf("Removed = %v, want [github]", result.Removed)
	}
	if result.ConfigPath != ".claude/settings.local.json" {
		t.Errorf("ConfigPath = %q, want .claude/settings.local.json", result.ConfigPath)
	}
}

// TestUninstall_CleansEmptyParentDir verifies that when the config file is
// removed (doc becomes empty), the parent directory is also cleaned up if empty.
func TestUninstall_CleansEmptyParentDir(t *testing.T) {
	target := t.TempDir()
	translator := &CopilotTranslator{}

	// Create .vscode/mcp.json with only MCP content (no other keys)
	configDir := filepath.Join(target, ".vscode")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	existing := `{
  "servers": {
    "github": {
      "command": "npx"
    }
  }
}
`
	if err := os.WriteFile(filepath.Join(configDir, "mcp.json"), []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Uninstall(target, translator, []string{"github"}, false)
	if err != nil {
		t.Fatalf("Uninstall() error = %v", err)
	}
	if len(result.Removed) != 1 {
		t.Errorf("Removed = %v, want [github]", result.Removed)
	}

	// The config file should be removed (doc was empty)
	configPath := filepath.Join(configDir, "mcp.json")
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Error("config file should have been removed when doc is empty")
	}

	// The parent dir (.vscode) should also be cleaned up
	if _, err := os.Stat(configDir); !os.IsNotExist(err) {
		t.Error("empty parent directory should have been cleaned up")
	}
}

// ---------- TOML uninstall tests ----------

func TestUninstall_CodexTranslator_RemovesServer(t *testing.T) {
	target := t.TempDir()
	translator := &CodexTranslator{}

	configDir := filepath.Join(target, ".codex")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	existing := `[mcp_servers.github]
type = "stdio"
command = "npx"
args = ["-y", "@modelcontextprotocol/server-github"]

[mcp_servers.linear]
type = "stdio"
command = "npx"
args = ["-y", "@modelcontextprotocol/server-linear"]
`
	if err := os.WriteFile(filepath.Join(configDir, "config.toml"), []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Uninstall(target, translator, []string{"github"}, false)
	if err != nil {
		t.Fatalf("Uninstall() error = %v", err)
	}
	if len(result.Removed) != 1 || result.Removed[0] != "github" {
		t.Errorf("Removed = %v, want [github]", result.Removed)
	}
	if result.ConfigPath != ".codex/config.toml" {
		t.Errorf("ConfigPath = %q, want .codex/config.toml", result.ConfigPath)
	}

	// Verify file was updated with valid TOML
	data, err := os.ReadFile(filepath.Join(configDir, "config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := toml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("output is not valid TOML: %v", err)
	}
	servers := doc["mcp_servers"].(map[string]any)
	if _, ok := servers["github"]; ok {
		t.Error("github server should have been removed")
	}
	if _, ok := servers["linear"]; !ok {
		t.Error("linear server should still be present")
	}
}

func TestUninstall_CodexTranslator_RemovesKeyWhenEmpty(t *testing.T) {
	target := t.TempDir()
	translator := &CodexTranslator{}

	configDir := filepath.Join(target, ".codex")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	existing := `[mcp_servers.github]
type = "stdio"
command = "npx"
`
	if err := os.WriteFile(filepath.Join(configDir, "config.toml"), []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Uninstall(target, translator, []string{"github"}, false)
	if err != nil {
		t.Fatalf("Uninstall() error = %v", err)
	}
	if len(result.Removed) != 1 {
		t.Errorf("Removed = %v, want [github]", result.Removed)
	}

	// File should be removed when doc is empty
	configPath := filepath.Join(configDir, "config.toml")
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Error("config file should have been removed when doc is empty")
	}

	// Parent dir should be cleaned up
	if _, err := os.Stat(configDir); !os.IsNotExist(err) {
		t.Error("empty parent directory should have been cleaned up")
	}
}

func TestUninstall_CodexTranslator_PreservesOtherKeys(t *testing.T) {
	target := t.TempDir()
	translator := &CodexTranslator{}

	configDir := filepath.Join(target, ".codex")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	existing := `title = "my config"

[mcp_servers.github]
type = "stdio"
command = "npx"
`
	if err := os.WriteFile(filepath.Join(configDir, "config.toml"), []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Uninstall(target, translator, []string{"github"}, false)
	if err != nil {
		t.Fatalf("Uninstall() error = %v", err)
	}
	if len(result.Removed) != 1 {
		t.Errorf("Removed = %v, want [github]", result.Removed)
	}

	// File should still exist because non-MCP keys remain
	data, err := os.ReadFile(filepath.Join(configDir, "config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := toml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("output is not valid TOML: %v", err)
	}
	if doc["title"] != "my config" {
		t.Errorf("title not preserved: got %v", doc["title"])
	}
	// mcp_servers key should be removed entirely (empty)
	if _, ok := doc["mcp_servers"]; ok {
		t.Error("mcp_servers key should be removed when empty")
	}
}
