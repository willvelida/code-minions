package mcp

import (
	"strings"
	"testing"
)

func TestParseConfig_StdioServer(t *testing.T) {
	yaml := `
servers:
  github:
    description: "GitHub MCP server"
    transport: stdio
    command: npx
    args:
      - "-y"
      - "@modelcontextprotocol/server-github"
    env:
      GITHUB_PERSONAL_ACCESS_TOKEN: ""
    required: false
`
	cfg, err := ParseConfig([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Servers) != 1 {
		t.Fatalf("server count: got %d, want 1", len(cfg.Servers))
	}

	s, ok := cfg.Servers["github"]
	if !ok {
		t.Fatal("expected server 'github' to exist")
	}
	if s.Transport != TransportStdio {
		t.Errorf("Transport: got %q, want %q", s.Transport, TransportStdio)
	}
	if s.Command != "npx" {
		t.Errorf("Command: got %q, want %q", s.Command, "npx")
	}
	if len(s.Args) != 2 || s.Args[0] != "-y" || s.Args[1] != "@modelcontextprotocol/server-github" {
		t.Errorf("Args: got %v, want [-y @modelcontextprotocol/server-github]", s.Args)
	}
	if s.Env["GITHUB_PERSONAL_ACCESS_TOKEN"] != "" {
		t.Errorf("Env token should be empty string, got %q", s.Env["GITHUB_PERSONAL_ACCESS_TOKEN"])
	}
	if s.Description != "GitHub MCP server" {
		t.Errorf("Description: got %q, want %q", s.Description, "GitHub MCP server")
	}
	if s.Required {
		t.Error("Required should be false")
	}
}

func TestParseConfig_HTTPServer(t *testing.T) {
	yaml := `
servers:
  remote-api:
    description: "Remote API docs"
    transport: streamable-http
    url: "https://mcp.example.com/v1"
    headers:
      Authorization: "Bearer token123"
    required: false
`
	cfg, err := ParseConfig([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s := cfg.Servers["remote-api"]
	if s.Transport != TransportStreamableHTTP {
		t.Errorf("Transport: got %q, want %q", s.Transport, TransportStreamableHTTP)
	}
	if s.URL != "https://mcp.example.com/v1" {
		t.Errorf("URL: got %q, want %q", s.URL, "https://mcp.example.com/v1")
	}
	if s.Headers["Authorization"] != "Bearer token123" {
		t.Errorf("Headers: got %v", s.Headers)
	}
}

func TestParseConfig_SSEServer(t *testing.T) {
	yaml := `
servers:
  events:
    transport: sse
    url: "https://sse.example.com/events"
`
	cfg, err := ParseConfig([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Servers["events"].Transport != TransportSSE {
		t.Errorf("Transport: got %q, want %q", cfg.Servers["events"].Transport, TransportSSE)
	}
}

func TestParseConfig_MultipleServers(t *testing.T) {
	yaml := `
servers:
  github:
    transport: stdio
    command: npx
    args: ["-y", "@modelcontextprotocol/server-github"]
  postgres:
    transport: stdio
    command: npx
    args: ["-y", "@modelcontextprotocol/server-postgres"]
    env:
      POSTGRES_CONNECTION_STRING: ""
`
	cfg, err := ParseConfig([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Servers) != 2 {
		t.Fatalf("server count: got %d, want 2", len(cfg.Servers))
	}
	if _, ok := cfg.Servers["github"]; !ok {
		t.Error("expected server 'github'")
	}
	if _, ok := cfg.Servers["postgres"]; !ok {
		t.Error("expected server 'postgres'")
	}
}

func TestParseConfig_MissingTransport(t *testing.T) {
	yaml := `
servers:
  github:
    command: npx
`
	_, err := ParseConfig([]byte(yaml))
	if err == nil {
		t.Fatal("expected error for missing transport")
	}
	if !strings.Contains(err.Error(), "missing required field: transport") {
		t.Errorf("error should mention transport, got: %v", err)
	}
}

func TestParseConfig_StdioWithoutCommand(t *testing.T) {
	yaml := `
servers:
  github:
    transport: stdio
`
	_, err := ParseConfig([]byte(yaml))
	if err == nil {
		t.Fatal("expected error for stdio without command")
	}
	if !strings.Contains(err.Error(), "requires a command") {
		t.Errorf("error should mention command, got: %v", err)
	}
}

func TestParseConfig_HTTPWithoutURL(t *testing.T) {
	yaml := `
servers:
  remote:
    transport: streamable-http
`
	_, err := ParseConfig([]byte(yaml))
	if err == nil {
		t.Fatal("expected error for HTTP transport without url")
	}
	if !strings.Contains(err.Error(), "requires a url") {
		t.Errorf("error should mention url, got: %v", err)
	}
}

func TestParseConfig_StdioWithURL(t *testing.T) {
	yaml := `
servers:
  bad:
    transport: stdio
    command: npx
    url: "https://should-not-be-here.com"
`
	_, err := ParseConfig([]byte(yaml))
	if err == nil {
		t.Fatal("expected error for stdio with url")
	}
	if !strings.Contains(err.Error(), "must not set url") {
		t.Errorf("error should mention url conflict, got: %v", err)
	}
}

func TestParseConfig_HTTPWithCommand(t *testing.T) {
	yaml := `
servers:
  bad:
    transport: sse
    url: "https://example.com"
    command: npx
`
	_, err := ParseConfig([]byte(yaml))
	if err == nil {
		t.Fatal("expected error for HTTP with command")
	}
	if !strings.Contains(err.Error(), "must not set command") {
		t.Errorf("error should mention command conflict, got: %v", err)
	}
}

func TestParseConfig_UnknownTransport(t *testing.T) {
	yaml := `
servers:
  bad:
    transport: websocket
    url: "wss://example.com"
`
	_, err := ParseConfig([]byte(yaml))
	if err == nil {
		t.Fatal("expected error for unknown transport")
	}
	if !strings.Contains(err.Error(), "unknown transport") {
		t.Errorf("error should mention unknown transport, got: %v", err)
	}
}

func TestParseConfig_NoServers(t *testing.T) {
	yaml := `
servers: {}
`
	_, err := ParseConfig([]byte(yaml))
	if err == nil {
		t.Fatal("expected error for empty servers")
	}
	if !strings.Contains(err.Error(), "no servers defined") {
		t.Errorf("error should mention no servers, got: %v", err)
	}
}

func TestParseConfig_InvalidYAML(t *testing.T) {
	_, err := ParseConfig([]byte("not: [valid yaml"))
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestEmptyEnvVars(t *testing.T) {
	cfg := &Config{
		Servers: map[string]Server{
			"github": {
				Transport: TransportStdio,
				Command:   "npx",
				Env: map[string]string{
					"GITHUB_TOKEN": "",
					"OTHER_VAR":    "set",
				},
			},
			"postgres": {
				Transport: TransportStdio,
				Command:   "npx",
				Env: map[string]string{
					"CONNECTION_STRING": "",
				},
			},
			"no-env": {
				Transport: TransportStdio,
				Command:   "echo",
			},
		},
	}

	empty := cfg.EmptyEnvVars()

	if len(empty) != 2 {
		t.Fatalf("expected 2 servers with empty vars, got %d", len(empty))
	}
	if len(empty["github"]) != 1 || empty["github"][0] != "GITHUB_TOKEN" {
		t.Errorf("github empty vars: got %v, want [GITHUB_TOKEN]", empty["github"])
	}
	if len(empty["postgres"]) != 1 || empty["postgres"][0] != "CONNECTION_STRING" {
		t.Errorf("postgres empty vars: got %v, want [CONNECTION_STRING]", empty["postgres"])
	}
}

func TestCollectEmptyEnvVars_MultipleConfigs(t *testing.T) {
	pkg1 := &Config{
		Servers: map[string]Server{
			"github": {
				Transport: TransportStdio,
				Command:   "npx",
				Env:       map[string]string{"GITHUB_TOKEN": ""},
			},
		},
	}
	pkg2 := &Config{
		Servers: map[string]Server{
			"postgres": {
				Transport: TransportStdio,
				Command:   "npx",
				Env:       map[string]string{"DATABASE_URL": ""},
			},
		},
	}
	team := &Config{
		Servers: map[string]Server{
			"sentry": {
				Transport: TransportSSE,
				URL:       "https://sentry.io/mcp",
				Env:       map[string]string{"SENTRY_TOKEN": ""},
			},
		},
	}

	empty := CollectEmptyEnvVars(pkg1, pkg2, team)
	if len(empty) != 3 {
		t.Fatalf("expected 3 servers, got %d: %v", len(empty), empty)
	}
	if empty["github"][0] != "GITHUB_TOKEN" {
		t.Error("missing GITHUB_TOKEN")
	}
	if empty["postgres"][0] != "DATABASE_URL" {
		t.Error("missing DATABASE_URL")
	}
	if empty["sentry"][0] != "SENTRY_TOKEN" {
		t.Error("missing SENTRY_TOKEN")
	}
}

func TestCollectEmptyEnvVars_NilConfigs(t *testing.T) {
	empty := CollectEmptyEnvVars(nil, nil)
	if empty != nil {
		t.Errorf("expected nil, got %v", empty)
	}
}

func TestCollectEmptyEnvVars_TeamOverridesPackageEnv(t *testing.T) {
	// Package has github with empty token
	pkg := &Config{
		Servers: map[string]Server{
			"github": {
				Transport: TransportStdio,
				Command:   "npx",
				Env:       map[string]string{"GITHUB_TOKEN": ""},
			},
		},
	}
	// Team overrides github with a filled token
	team := &Config{
		Servers: map[string]Server{
			"github": {
				Transport: TransportStdio,
				Command:   "npx",
				Env:       map[string]string{"GITHUB_TOKEN": "ghp_abc"},
			},
		},
	}

	empty := CollectEmptyEnvVars(pkg, team)
	// Team overrides package — no empty vars on github
	if _, ok := empty["github"]; ok {
		t.Error("github should have no empty env vars after team override")
	}
}
