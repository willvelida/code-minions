package mcp

import (
	"encoding/json"
	"testing"
)

// testStdioConfig returns a canonical Config with a single stdio server
// for consistent use across translator tests.
func testStdioConfig() *Config {
	return &Config{
		Servers: map[string]Server{
			"github": {
				Description: "GitHub MCP server",
				Transport:   TransportStdio,
				Command:     "npx",
				Args:        []string{"-y", "@modelcontextprotocol/server-github"},
				Env: map[string]string{
					"GITHUB_PERSONAL_ACCESS_TOKEN": "",
				},
			},
		},
	}
}

// testHTTPConfig returns a canonical Config with a streamable-http server.
func testHTTPConfig() *Config {
	return &Config{
		Servers: map[string]Server{
			"remote-api": {
				Description: "Remote API docs",
				Transport:   TransportStreamableHTTP,
				URL:         "https://mcp.example.com/v1",
				Headers: map[string]string{
					"Authorization": "Bearer token",
				},
			},
		},
	}
}

// --- Copilot ---

func TestCopilotTranslator_Stdio(t *testing.T) {
	tr := &CopilotTranslator{}
	servers, warnings, err := tr.Translate(testStdioConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("unexpected warnings: %v", warnings)
	}

	github, ok := servers["github"]
	if !ok {
		t.Fatal("expected server 'github'")
	}
	m := github.(map[string]any)

	if m["command"] != "npx" {
		t.Errorf("command: got %v, want %q", m["command"], "npx")
	}
	args := m["args"].([]string)
	if len(args) != 2 || args[0] != "-y" {
		t.Errorf("args: got %v", args)
	}
	env := m["env"].(map[string]string)
	if _, ok := env["GITHUB_PERSONAL_ACCESS_TOKEN"]; !ok {
		t.Error("expected GITHUB_PERSONAL_ACCESS_TOKEN in env")
	}
}

func TestCopilotTranslator_HTTP(t *testing.T) {
	tr := &CopilotTranslator{}
	servers, _, err := tr.Translate(testHTTPConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	api, ok := servers["remote-api"]
	if !ok {
		t.Fatal("expected server 'remote-api'")
	}
	m := api.(map[string]any)

	if m["type"] != "http" {
		t.Errorf("type: got %v, want %q", m["type"], "http")
	}
	if m["url"] != "https://mcp.example.com/v1" {
		t.Errorf("url: got %v", m["url"])
	}
}

func TestCopilotTranslator_ConfigPath(t *testing.T) {
	tr := &CopilotTranslator{}
	if tr.ConfigPath() != ".vscode/mcp.json" {
		t.Errorf("ConfigPath: got %q", tr.ConfigPath())
	}
	if tr.ConfigKey() != "servers" {
		t.Errorf("ConfigKey: got %q", tr.ConfigKey())
	}
}

func TestCopilotTranslator_JSONSnapshot(t *testing.T) {
	tr := &CopilotTranslator{}
	servers, _, _ := tr.Translate(testStdioConfig())

	result := map[string]any{tr.ConfigKey(): servers}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	// Verify it round-trips as valid JSON
	var check map[string]any
	if err := json.Unmarshal(data, &check); err != nil {
		t.Fatalf("produced invalid JSON: %v\n%s", err, data)
	}

	// Verify structure
	serversKey, ok := check["servers"]
	if !ok {
		t.Fatal("expected top-level 'servers' key")
	}
	serversMap := serversKey.(map[string]any)
	if _, ok := serversMap["github"]; !ok {
		t.Error("expected 'github' server under 'servers'")
	}
}

// --- Claude ---

func TestClaudeTranslator_Stdio(t *testing.T) {
	tr := &ClaudeTranslator{}
	servers, warnings, err := tr.Translate(testStdioConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("unexpected warnings: %v", warnings)
	}

	github := servers["github"].(map[string]any)
	if github["command"] != "npx" {
		t.Errorf("command: got %v", github["command"])
	}
}

func TestClaudeTranslator_HTTP(t *testing.T) {
	tr := &ClaudeTranslator{}
	servers, _, err := tr.Translate(testHTTPConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	api := servers["remote-api"].(map[string]any)
	if api["url"] != "https://mcp.example.com/v1" {
		t.Errorf("url: got %v", api["url"])
	}
}

func TestClaudeTranslator_ConfigPath(t *testing.T) {
	tr := &ClaudeTranslator{}
	if tr.ConfigPath() != ".claude/settings.local.json" {
		t.Errorf("ConfigPath: got %q", tr.ConfigPath())
	}
	if tr.ConfigKey() != "mcpServers" {
		t.Errorf("ConfigKey: got %q", tr.ConfigKey())
	}
}

// --- OpenCode ---

func TestOpenCodeTranslator_Stdio(t *testing.T) {
	tr := &OpenCodeTranslator{}
	servers, warnings, err := tr.Translate(testStdioConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("unexpected warnings: %v", warnings)
	}

	github := servers["github"].(map[string]any)

	if github["type"] != "local" {
		t.Errorf("type: got %v, want %q", github["type"], "local")
	}
	if github["enabled"] != true {
		t.Errorf("enabled: got %v, want true", github["enabled"])
	}

	// OpenCode merges command + args into a single array
	cmdArray := github["command"].([]string)
	if len(cmdArray) != 3 || cmdArray[0] != "npx" || cmdArray[1] != "-y" {
		t.Errorf("command array: got %v, want [npx -y ...]", cmdArray)
	}

	// env → environment
	env := github["environment"].(map[string]string)
	if _, ok := env["GITHUB_PERSONAL_ACCESS_TOKEN"]; !ok {
		t.Error("expected GITHUB_PERSONAL_ACCESS_TOKEN in environment")
	}
}

func TestOpenCodeTranslator_HTTP(t *testing.T) {
	tr := &OpenCodeTranslator{}
	servers, _, err := tr.Translate(testHTTPConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	api := servers["remote-api"].(map[string]any)
	if api["type"] != "remote" {
		t.Errorf("type: got %v, want %q", api["type"], "remote")
	}
	if api["url"] != "https://mcp.example.com/v1" {
		t.Errorf("url: got %v", api["url"])
	}
	if api["enabled"] != true {
		t.Errorf("enabled: got %v", api["enabled"])
	}
}

func TestOpenCodeTranslator_ConfigPath(t *testing.T) {
	tr := &OpenCodeTranslator{}
	if tr.ConfigPath() != "opencode.json" {
		t.Errorf("ConfigPath: got %q", tr.ConfigPath())
	}
	if tr.ConfigKey() != "mcp" {
		t.Errorf("ConfigKey: got %q", tr.ConfigKey())
	}
}

// --- Factory ---

func TestNewTranslator(t *testing.T) {
	tests := []struct {
		name      string
		expectErr bool
	}{
		{"copilot", false},
		{"claude", false},
		{"opencode", false},
		{"unknown", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr, err := NewTranslator(tt.name)
			if tt.expectErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tr == nil {
				t.Error("expected non-nil translator")
			}
		})
	}
}
