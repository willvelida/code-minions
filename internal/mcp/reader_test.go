package mcp

import (
	"encoding/json"
	"testing"
)

// --- Copilot Reader ---

func TestCopilotReader_StdioServer(t *testing.T) {
	input := `{
  "servers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": { "GITHUB_PERSONAL_ACCESS_TOKEN": "ghp_xxx" }
    }
  }
}`
	r := &CopilotReader{}
	cfg, warnings, err := r.Read([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("unexpected warnings: %v", warnings)
	}
	if len(cfg.Servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(cfg.Servers))
	}

	s := cfg.Servers["github"]
	if s.Transport != TransportStdio {
		t.Errorf("transport: got %q, want %q", s.Transport, TransportStdio)
	}
	if s.Command != "npx" {
		t.Errorf("command: got %q, want %q", s.Command, "npx")
	}
	if len(s.Args) != 2 || s.Args[0] != "-y" || s.Args[1] != "@modelcontextprotocol/server-github" {
		t.Errorf("args: got %v", s.Args)
	}
	if s.Env["GITHUB_PERSONAL_ACCESS_TOKEN"] != "ghp_xxx" {
		t.Errorf("env: got %v", s.Env)
	}
}

func TestCopilotReader_SSEServer(t *testing.T) {
	input := `{
  "servers": {
    "remote": {
      "url": "https://mcp.example.com/sse",
      "headers": { "Authorization": "Bearer token" }
    }
  }
}`
	r := &CopilotReader{}
	cfg, _, err := r.Read([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s := cfg.Servers["remote"]
	if s.Transport != TransportSSE {
		t.Errorf("transport: got %q, want %q", s.Transport, TransportSSE)
	}
	if s.URL != "https://mcp.example.com/sse" {
		t.Errorf("url: got %q", s.URL)
	}
	if s.Headers["Authorization"] != "Bearer token" {
		t.Errorf("headers: got %v", s.Headers)
	}
}

func TestCopilotReader_StreamableHTTPServer(t *testing.T) {
	input := `{
  "servers": {
    "api": {
      "type": "http",
      "url": "https://mcp.example.com/v1"
    }
  }
}`
	r := &CopilotReader{}
	cfg, _, err := r.Read([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s := cfg.Servers["api"]
	if s.Transport != TransportStreamableHTTP {
		t.Errorf("transport: got %q, want %q", s.Transport, TransportStreamableHTTP)
	}
	if s.URL != "https://mcp.example.com/v1" {
		t.Errorf("url: got %q", s.URL)
	}
}

func TestCopilotReader_MultipleServers(t *testing.T) {
	input := `{
  "servers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": { "GITHUB_PERSONAL_ACCESS_TOKEN": "" }
    },
    "remote": {
      "url": "https://mcp.example.com/sse"
    }
  }
}`
	r := &CopilotReader{}
	cfg, _, err := r.Read([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Servers) != 2 {
		t.Fatalf("expected 2 servers, got %d", len(cfg.Servers))
	}
	if _, ok := cfg.Servers["github"]; !ok {
		t.Error("expected server 'github'")
	}
	if _, ok := cfg.Servers["remote"]; !ok {
		t.Error("expected server 'remote'")
	}
}

func TestCopilotReader_NoServersKey(t *testing.T) {
	input := `{ "someOtherKey": true }`
	r := &CopilotReader{}
	cfg, _, err := r.Read([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Servers) != 0 {
		t.Errorf("expected 0 servers, got %d", len(cfg.Servers))
	}
}

func TestCopilotReader_EmptyServers(t *testing.T) {
	input := `{ "servers": {} }`
	r := &CopilotReader{}
	cfg, _, err := r.Read([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Servers) != 0 {
		t.Errorf("expected 0 servers, got %d", len(cfg.Servers))
	}
}

func TestCopilotReader_MalformedJSON(t *testing.T) {
	r := &CopilotReader{}
	_, _, err := r.Read([]byte("not json"))
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestCopilotReader_MissingCommand(t *testing.T) {
	input := `{
  "servers": {
    "bad": { "args": ["-y"] }
  }
}`
	r := &CopilotReader{}
	cfg, warnings, err := r.Read([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Server with missing command should be skipped with a warning
	if len(cfg.Servers) != 0 {
		t.Errorf("expected 0 servers (skipped), got %d", len(cfg.Servers))
	}
	if len(warnings) != 1 {
		t.Errorf("expected 1 warning, got %d: %v", len(warnings), warnings)
	}
}

// --- Claude Reader ---

func TestClaudeReader_StdioServer(t *testing.T) {
	input := `{
  "permissions": { "allow": ["read"] },
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": { "GITHUB_PERSONAL_ACCESS_TOKEN": "" }
    }
  }
}`
	r := &ClaudeReader{}
	cfg, _, err := r.Read([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s := cfg.Servers["github"]
	if s.Transport != TransportStdio {
		t.Errorf("transport: got %q, want %q", s.Transport, TransportStdio)
	}
	if s.Command != "npx" {
		t.Errorf("command: got %q, want %q", s.Command, "npx")
	}
	if len(s.Args) != 2 {
		t.Errorf("args: got %v", s.Args)
	}
}

func TestClaudeReader_HTTPServer(t *testing.T) {
	input := `{
  "mcpServers": {
    "api": {
      "url": "https://mcp.example.com/v1",
      "headers": { "Authorization": "Bearer token" },
      "env": { "API_KEY": "secret" }
    }
  }
}`
	r := &ClaudeReader{}
	cfg, _, err := r.Read([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s := cfg.Servers["api"]
	if s.Transport != TransportSSE {
		t.Errorf("transport: got %q, want %q", s.Transport, TransportSSE)
	}
	if s.URL != "https://mcp.example.com/v1" {
		t.Errorf("url: got %q", s.URL)
	}
	if s.Headers["Authorization"] != "Bearer token" {
		t.Errorf("headers: got %v", s.Headers)
	}
	if s.Env["API_KEY"] != "secret" {
		t.Errorf("env: got %v", s.Env)
	}
}

func TestClaudeReader_NoMCPServersKey(t *testing.T) {
	input := `{ "permissions": { "allow": ["read"] } }`
	r := &ClaudeReader{}
	cfg, _, err := r.Read([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Servers) != 0 {
		t.Errorf("expected 0 servers, got %d", len(cfg.Servers))
	}
}

func TestClaudeReader_MalformedJSON(t *testing.T) {
	r := &ClaudeReader{}
	_, _, err := r.Read([]byte("{bad"))
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

// --- OpenCode Reader ---

func TestOpenCodeReader_StdioServer(t *testing.T) {
	input := `{
  "mcp": {
    "github": {
      "type": "local",
      "command": ["npx", "-y", "@modelcontextprotocol/server-github"],
      "environment": { "GITHUB_PERSONAL_ACCESS_TOKEN": "ghp_xxx" },
      "enabled": true
    }
  }
}`
	r := &OpenCodeReader{}
	cfg, warnings, err := r.Read([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("unexpected warnings: %v", warnings)
	}

	s := cfg.Servers["github"]
	if s.Transport != TransportStdio {
		t.Errorf("transport: got %q, want %q", s.Transport, TransportStdio)
	}
	if s.Command != "npx" {
		t.Errorf("command: got %q, want %q", s.Command, "npx")
	}
	if len(s.Args) != 2 || s.Args[0] != "-y" || s.Args[1] != "@modelcontextprotocol/server-github" {
		t.Errorf("args: got %v", s.Args)
	}
	if s.Env["GITHUB_PERSONAL_ACCESS_TOKEN"] != "ghp_xxx" {
		t.Errorf("env: got %v", s.Env)
	}
}

func TestOpenCodeReader_RemoteServer(t *testing.T) {
	input := `{
  "mcp": {
    "api": {
      "type": "remote",
      "url": "https://mcp.example.com/v1",
      "headers": { "Authorization": "Bearer token" },
      "enabled": true
    }
  }
}`
	r := &OpenCodeReader{}
	cfg, _, err := r.Read([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s := cfg.Servers["api"]
	if s.Transport != TransportSSE {
		t.Errorf("transport: got %q, want %q", s.Transport, TransportSSE)
	}
	if s.URL != "https://mcp.example.com/v1" {
		t.Errorf("url: got %q", s.URL)
	}
	if s.Headers["Authorization"] != "Bearer token" {
		t.Errorf("headers: got %v", s.Headers)
	}
}

func TestOpenCodeReader_DisabledServer(t *testing.T) {
	input := `{
  "mcp": {
    "github": {
      "type": "local",
      "command": ["npx", "-y", "@modelcontextprotocol/server-github"],
      "enabled": false
    }
  }
}`
	r := &OpenCodeReader{}
	cfg, warnings, err := r.Read([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Servers) != 0 {
		t.Errorf("expected 0 servers (disabled), got %d", len(cfg.Servers))
	}
	if len(warnings) != 1 {
		t.Errorf("expected 1 warning, got %d: %v", len(warnings), warnings)
	}
}

func TestOpenCodeReader_CommandArraySplitting(t *testing.T) {
	input := `{
  "mcp": {
    "test": {
      "type": "local",
      "command": ["python", "-m", "mcp_server", "--port", "8080"],
      "enabled": true
    }
  }
}`
	r := &OpenCodeReader{}
	cfg, _, err := r.Read([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s := cfg.Servers["test"]
	if s.Command != "python" {
		t.Errorf("command: got %q, want %q", s.Command, "python")
	}
	expected := []string{"-m", "mcp_server", "--port", "8080"}
	if len(s.Args) != len(expected) {
		t.Fatalf("args length: got %d, want %d", len(s.Args), len(expected))
	}
	for i, want := range expected {
		if s.Args[i] != want {
			t.Errorf("args[%d]: got %q, want %q", i, s.Args[i], want)
		}
	}
}

func TestOpenCodeReader_SingleElementCommand(t *testing.T) {
	input := `{
  "mcp": {
    "simple": {
      "type": "local",
      "command": ["my-server"],
      "enabled": true
    }
  }
}`
	r := &OpenCodeReader{}
	cfg, _, err := r.Read([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s := cfg.Servers["simple"]
	if s.Command != "my-server" {
		t.Errorf("command: got %q, want %q", s.Command, "my-server")
	}
	if len(s.Args) != 0 {
		t.Errorf("expected no args, got %v", s.Args)
	}
}

func TestOpenCodeReader_EnvironmentToEnv(t *testing.T) {
	input := `{
  "mcp": {
    "test": {
      "type": "local",
      "command": ["server"],
      "environment": { "KEY1": "val1", "KEY2": "val2" },
      "enabled": true
    }
  }
}`
	r := &OpenCodeReader{}
	cfg, _, err := r.Read([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s := cfg.Servers["test"]
	if s.Env["KEY1"] != "val1" || s.Env["KEY2"] != "val2" {
		t.Errorf("env: got %v", s.Env)
	}
}

func TestOpenCodeReader_NoMCPKey(t *testing.T) {
	input := `{ "theme": "dark" }`
	r := &OpenCodeReader{}
	cfg, _, err := r.Read([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Servers) != 0 {
		t.Errorf("expected 0 servers, got %d", len(cfg.Servers))
	}
}

func TestOpenCodeReader_MalformedJSON(t *testing.T) {
	r := &OpenCodeReader{}
	_, _, err := r.Read([]byte("nope"))
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestOpenCodeReader_MissingCommand(t *testing.T) {
	input := `{
  "mcp": {
    "bad": {
      "type": "local",
      "enabled": true
    }
  }
}`
	r := &OpenCodeReader{}
	cfg, warnings, err := r.Read([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Servers) != 0 {
		t.Errorf("expected 0 servers, got %d", len(cfg.Servers))
	}
	if len(warnings) == 0 {
		t.Error("expected warning for missing command")
	}
}

func TestOpenCodeReader_EmptyCommandArray(t *testing.T) {
	input := `{
  "mcp": {
    "bad": {
      "type": "local",
      "command": [],
      "enabled": true
    }
  }
}`
	r := &OpenCodeReader{}
	cfg, warnings, err := r.Read([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Servers) != 0 {
		t.Errorf("expected 0 servers, got %d", len(cfg.Servers))
	}
	if len(warnings) == 0 {
		t.Error("expected warning for empty command array")
	}
}

func TestOpenCodeReader_RemoteMissingURL(t *testing.T) {
	input := `{
  "mcp": {
    "bad": {
      "type": "remote",
      "enabled": true
    }
  }
}`
	r := &OpenCodeReader{}
	cfg, warnings, err := r.Read([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Servers) != 0 {
		t.Errorf("expected 0 servers, got %d", len(cfg.Servers))
	}
	if len(warnings) == 0 {
		t.Error("expected warning for missing URL")
	}
}

// --- NewReader factory ---

func TestNewReader_ValidAssistants(t *testing.T) {
	tests := []struct {
		name       string
		configPath string
	}{
		{"copilot", ".vscode/mcp.json"},
		{"claude", ".claude/settings.local.json"},
		{"opencode", "opencode.json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := NewReader(tt.name)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if r.ConfigPath() != tt.configPath {
				t.Errorf("ConfigPath: got %q, want %q", r.ConfigPath(), tt.configPath)
			}
		})
	}
}

func TestNewReader_UnknownAssistant(t *testing.T) {
	_, err := NewReader("unknown")
	if err == nil {
		t.Fatal("expected error for unknown assistant")
	}
}

// --- Round-trip tests ---

// TestRoundTrip_CopilotReadWrite reads Copilot config, translates back to
// Copilot format, and verifies the output matches the canonical representation.
func TestRoundTrip_CopilotReadWrite(t *testing.T) {
	input := `{
  "servers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": { "GITHUB_PERSONAL_ACCESS_TOKEN": "" }
    }
  }
}`
	reader := &CopilotReader{}
	cfg, _, err := reader.Read([]byte(input))
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	writer := &CopilotTranslator{}
	servers, _, err := writer.Translate(cfg)
	if err != nil {
		t.Fatalf("translate: %v", err)
	}

	// Re-read the translated output
	translated, _ := json.Marshal(map[string]any{"servers": servers})
	cfg2, _, err := reader.Read(translated)
	if err != nil {
		t.Fatalf("re-read: %v", err)
	}

	assertConfigsEqual(t, cfg, cfg2)
}

// TestRoundTrip_ClaudeReadWrite reads Claude config and round-trips it.
func TestRoundTrip_ClaudeReadWrite(t *testing.T) {
	input := `{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": { "GITHUB_PERSONAL_ACCESS_TOKEN": "" }
    }
  }
}`
	reader := &ClaudeReader{}
	cfg, _, err := reader.Read([]byte(input))
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	writer := &ClaudeTranslator{}
	servers, _, err := writer.Translate(cfg)
	if err != nil {
		t.Fatalf("translate: %v", err)
	}

	translated, _ := json.Marshal(map[string]any{"mcpServers": servers})
	cfg2, _, err := reader.Read(translated)
	if err != nil {
		t.Fatalf("re-read: %v", err)
	}

	assertConfigsEqual(t, cfg, cfg2)
}

// TestRoundTrip_OpenCodeReadWrite reads OpenCode config and round-trips it.
func TestRoundTrip_OpenCodeReadWrite(t *testing.T) {
	input := `{
  "mcp": {
    "github": {
      "type": "local",
      "command": ["npx", "-y", "@modelcontextprotocol/server-github"],
      "environment": { "GITHUB_PERSONAL_ACCESS_TOKEN": "" },
      "enabled": true
    }
  }
}`
	reader := &OpenCodeReader{}
	cfg, _, err := reader.Read([]byte(input))
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	writer := &OpenCodeTranslator{}
	servers, _, err := writer.Translate(cfg)
	if err != nil {
		t.Fatalf("translate: %v", err)
	}

	translated, _ := json.Marshal(map[string]any{"mcp": servers})
	cfg2, _, err := reader.Read(translated)
	if err != nil {
		t.Fatalf("re-read: %v", err)
	}

	assertConfigsEqual(t, cfg, cfg2)
}

// TestRoundTrip_CopilotToClaude reads from Copilot, writes to Claude,
// reads back from Claude, and verifies the canonical config matches.
func TestRoundTrip_CopilotToClaude(t *testing.T) {
	input := `{
  "servers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": { "GITHUB_PERSONAL_ACCESS_TOKEN": "token123" }
    }
  }
}`
	copilotReader := &CopilotReader{}
	cfg1, _, err := copilotReader.Read([]byte(input))
	if err != nil {
		t.Fatalf("read copilot: %v", err)
	}

	// Translate to Claude format
	claudeWriter := &ClaudeTranslator{}
	servers, _, err := claudeWriter.Translate(cfg1)
	if err != nil {
		t.Fatalf("translate to claude: %v", err)
	}

	// Read back from Claude format
	claudeJSON, _ := json.Marshal(map[string]any{"mcpServers": servers})
	claudeReader := &ClaudeReader{}
	cfg2, _, err := claudeReader.Read(claudeJSON)
	if err != nil {
		t.Fatalf("read claude: %v", err)
	}

	assertConfigsEqual(t, cfg1, cfg2)
}

// TestRoundTrip_ClaudeToOpenCode exercises the field mapping differences.
func TestRoundTrip_ClaudeToOpenCode(t *testing.T) {
	input := `{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": { "GITHUB_PERSONAL_ACCESS_TOKEN": "" }
    }
  }
}`
	claudeReader := &ClaudeReader{}
	cfg1, _, err := claudeReader.Read([]byte(input))
	if err != nil {
		t.Fatalf("read claude: %v", err)
	}

	// Translate to OpenCode
	ocWriter := &OpenCodeTranslator{}
	servers, _, err := ocWriter.Translate(cfg1)
	if err != nil {
		t.Fatalf("translate to opencode: %v", err)
	}

	// Read back from OpenCode
	ocJSON, _ := json.Marshal(map[string]any{"mcp": servers})
	ocReader := &OpenCodeReader{}
	cfg2, _, err := ocReader.Read(ocJSON)
	if err != nil {
		t.Fatalf("read opencode: %v", err)
	}

	assertConfigsEqual(t, cfg1, cfg2)
}

// TestRoundTrip_OpenCodeToCopilot exercises command array splitting.
func TestRoundTrip_OpenCodeToCopilot(t *testing.T) {
	input := `{
  "mcp": {
    "github": {
      "type": "local",
      "command": ["npx", "-y", "@modelcontextprotocol/server-github"],
      "environment": { "GITHUB_PERSONAL_ACCESS_TOKEN": "token" },
      "enabled": true
    }
  }
}`
	ocReader := &OpenCodeReader{}
	cfg1, _, err := ocReader.Read([]byte(input))
	if err != nil {
		t.Fatalf("read opencode: %v", err)
	}

	// Translate to Copilot
	copilotWriter := &CopilotTranslator{}
	servers, _, err := copilotWriter.Translate(cfg1)
	if err != nil {
		t.Fatalf("translate to copilot: %v", err)
	}

	// Read back from Copilot
	copilotJSON, _ := json.Marshal(map[string]any{"servers": servers})
	copilotReader := &CopilotReader{}
	cfg2, _, err := copilotReader.Read(copilotJSON)
	if err != nil {
		t.Fatalf("read copilot: %v", err)
	}

	assertConfigsEqual(t, cfg1, cfg2)
}

// assertConfigsEqual compares two canonical configs for equivalence.
func assertConfigsEqual(t *testing.T, a, b *Config) {
	t.Helper()
	if len(a.Servers) != len(b.Servers) {
		t.Fatalf("server count: got %d, want %d", len(b.Servers), len(a.Servers))
	}
	for name, sa := range a.Servers {
		sb, ok := b.Servers[name]
		if !ok {
			t.Errorf("missing server %q in second config", name)
			continue
		}
		if sa.Transport != sb.Transport {
			t.Errorf("server %q transport: got %q, want %q", name, sb.Transport, sa.Transport)
		}
		if sa.Command != sb.Command {
			t.Errorf("server %q command: got %q, want %q", name, sb.Command, sa.Command)
		}
		if len(sa.Args) != len(sb.Args) {
			t.Errorf("server %q args length: got %d, want %d", name, len(sb.Args), len(sa.Args))
		} else {
			for i := range sa.Args {
				if sa.Args[i] != sb.Args[i] {
					t.Errorf("server %q args[%d]: got %q, want %q", name, i, sb.Args[i], sa.Args[i])
				}
			}
		}
		if sa.URL != sb.URL {
			t.Errorf("server %q url: got %q, want %q", name, sb.URL, sa.URL)
		}
		for k, v := range sa.Env {
			if sb.Env[k] != v {
				t.Errorf("server %q env[%s]: got %q, want %q", name, k, sb.Env[k], v)
			}
		}
		for k, v := range sa.Headers {
			if sb.Headers[k] != v {
				t.Errorf("server %q headers[%s]: got %q, want %q", name, k, sb.Headers[k], v)
			}
		}
	}
}
