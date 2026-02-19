package mcp

import (
	"encoding/json"
	"testing"

	"github.com/willvelida/code-minions/internal/assistant"
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

// testSSEConfig returns a canonical Config with an SSE server.
func testSSEConfig() *Config {
	return &Config{
		Servers: map[string]Server{
			"sse-api": {
				Description: "SSE endpoint",
				Transport:   TransportSSE,
				URL:         "https://sse.example.com/events",
				Headers: map[string]string{
					"Authorization": "Bearer token",
				},
				Env: map[string]string{
					"API_KEY": "secret",
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
	// Verify headers are included
	headers := m["headers"].(map[string]string)
	if headers["Authorization"] != "Bearer token" {
		t.Errorf("headers: got %v", headers)
	}
}

func TestCopilotTranslator_SSE(t *testing.T) {
	tr := &CopilotTranslator{}
	servers, _, err := tr.Translate(testSSEConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	api := servers["sse-api"].(map[string]any)
	if api["url"] != "https://sse.example.com/events" {
		t.Errorf("url: got %v", api["url"])
	}
	// SSE should NOT have type field (unlike streamable-http)
	if _, hasType := api["type"]; hasType {
		t.Error("SSE should not have 'type' field in Copilot format")
	}
	// Verify headers and env
	headers := api["headers"].(map[string]string)
	if headers["Authorization"] != "Bearer token" {
		t.Errorf("headers: got %v", headers)
	}
	env := api["env"].(map[string]string)
	if env["API_KEY"] != "secret" {
		t.Errorf("env: got %v", env)
	}
}

func TestCopilotTranslator_HTTPWithEnv(t *testing.T) {
	cfg := &Config{
		Servers: map[string]Server{
			"api": {
				Transport: TransportStreamableHTTP,
				URL:       "https://api.example.com",
				Env:       map[string]string{"TOKEN": "val"},
			},
		},
	}
	tr := &CopilotTranslator{}
	servers, _, err := tr.Translate(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := servers["api"].(map[string]any)
	if m["env"] == nil {
		t.Error("expected env field on HTTP server")
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

func TestClaudeTranslator_SSE(t *testing.T) {
	tr := &ClaudeTranslator{}
	servers, _, err := tr.Translate(testSSEConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	api := servers["sse-api"].(map[string]any)
	if api["url"] != "https://sse.example.com/events" {
		t.Errorf("url: got %v", api["url"])
	}
	// Claude should include env for SSE
	env := api["env"].(map[string]string)
	if env["API_KEY"] != "secret" {
		t.Errorf("env: got %v", env)
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
	// Verify headers
	headers := api["headers"].(map[string]string)
	if headers["Authorization"] != "Bearer token" {
		t.Errorf("headers: got %v", headers)
	}
}

func TestOpenCodeTranslator_SSE(t *testing.T) {
	tr := &OpenCodeTranslator{}
	servers, _, err := tr.Translate(testSSEConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	api := servers["sse-api"].(map[string]any)
	if api["type"] != "remote" {
		t.Errorf("type: got %v, want %q", api["type"], "remote")
	}
	if api["url"] != "https://sse.example.com/events" {
		t.Errorf("url: got %v", api["url"])
	}
	// env → environment
	env := api["environment"].(map[string]string)
	if env["API_KEY"] != "secret" {
		t.Errorf("environment: got %v", env)
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

// --- Cursor ---

func TestCursorTranslator_Stdio(t *testing.T) {
	tr := &CursorTranslator{}
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

func TestCursorTranslator_HTTP(t *testing.T) {
	tr := &CursorTranslator{}
	servers, _, err := tr.Translate(testHTTPConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	api := servers["remote-api"].(map[string]any)
	if api["url"] != "https://mcp.example.com/v1" {
		t.Errorf("url: got %v", api["url"])
	}
}

func TestCursorTranslator_SSE(t *testing.T) {
	tr := &CursorTranslator{}
	servers, _, err := tr.Translate(testSSEConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	api := servers["sse-api"].(map[string]any)
	if api["url"] != "https://sse.example.com/events" {
		t.Errorf("url: got %v", api["url"])
	}
	// Cursor should include env for SSE
	env := api["env"].(map[string]string)
	if env["API_KEY"] != "secret" {
		t.Errorf("env: got %v", env)
	}
}

func TestCursorTranslator_ConfigPath(t *testing.T) {
	tr := &CursorTranslator{}
	if tr.ConfigPath() != ".cursor/mcp.json" {
		t.Errorf("ConfigPath: got %q", tr.ConfigPath())
	}
	if tr.ConfigKey() != "mcpServers" {
		t.Errorf("ConfigKey: got %q", tr.ConfigKey())
	}
}

// --- Gemini ---

func TestGeminiTranslator_Stdio(t *testing.T) {
	tr := &GeminiTranslator{}
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

func TestGeminiTranslator_HTTP(t *testing.T) {
	tr := &GeminiTranslator{}
	servers, _, err := tr.Translate(testHTTPConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	api := servers["remote-api"].(map[string]any)
	if api["url"] != "https://mcp.example.com/v1" {
		t.Errorf("url: got %v", api["url"])
	}
}

func TestGeminiTranslator_SSE(t *testing.T) {
	tr := &GeminiTranslator{}
	servers, _, err := tr.Translate(testSSEConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	api := servers["sse-api"].(map[string]any)
	if api["url"] != "https://sse.example.com/events" {
		t.Errorf("url: got %v", api["url"])
	}
	// Gemini should include env for SSE
	env := api["env"].(map[string]string)
	if env["API_KEY"] != "secret" {
		t.Errorf("env: got %v", env)
	}
}

func TestGeminiTranslator_ConfigPath(t *testing.T) {
	tr := &GeminiTranslator{}
	if tr.ConfigPath() != ".gemini/settings.json" {
		t.Errorf("ConfigPath: got %q", tr.ConfigPath())
	}
	if tr.ConfigKey() != "mcpServers" {
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
		{"cursor", false},
		{"gemini", false},
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

// TestTranslatorConfigMatchesAssistant verifies that each translator's
// ConfigPath() and ConfigKey() return the same values as the corresponding
// assistant.Config. This prevents drift between the two sources.
func TestTranslatorConfigMatchesAssistant(t *testing.T) {
	names := []string{"copilot", "claude", "opencode", "cursor", "gemini"}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			tr, err := NewTranslator(name)
			if err != nil {
				t.Fatalf("NewTranslator(%q): %v", name, err)
			}
			cfg, err := assistant.Get(name)
			if err != nil {
				t.Fatalf("assistant.Get(%q): %v", name, err)
			}
			if tr.ConfigPath() != cfg.MCPConfigPath {
				t.Errorf("ConfigPath mismatch: translator=%q, assistant=%q", tr.ConfigPath(), cfg.MCPConfigPath)
			}
			if tr.ConfigKey() != cfg.MCPConfigKey {
				t.Errorf("ConfigKey mismatch: translator=%q, assistant=%q", tr.ConfigKey(), cfg.MCPConfigKey)
			}
		})
	}
}
