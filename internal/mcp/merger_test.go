package mcp

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestMerge_NewFile(t *testing.T) {
	translated := map[string]any{
		"github": map[string]any{
			"command": "npx",
			"args":    []string{"-y", "@modelcontextprotocol/server-github"},
		},
	}

	out, result, err := Merge(nil, translated, "servers", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Added) != 1 || result.Added[0] != "github" {
		t.Errorf("Added: got %v, want [github]", result.Added)
	}
	if len(result.Skipped) != 0 {
		t.Errorf("Skipped should be empty: %v", result.Skipped)
	}

	// Verify JSON structure
	var doc map[string]any
	if err := json.Unmarshal(out, &doc); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	servers, ok := doc["servers"].(map[string]any)
	if !ok {
		t.Fatal("expected 'servers' key in output")
	}
	if _, ok := servers["github"]; !ok {
		t.Error("expected 'github' server in output")
	}
}

func TestMerge_AddToExisting(t *testing.T) {
	existing := []byte(`{
  "servers": {
    "postgres": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-postgres"]
    }
  }
}`)

	translated := map[string]any{
		"github": map[string]any{
			"command": "npx",
			"args":    []string{"-y", "@modelcontextprotocol/server-github"},
		},
	}

	out, result, err := Merge(existing, translated, "servers", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Added) != 1 || result.Added[0] != "github" {
		t.Errorf("Added: got %v, want [github]", result.Added)
	}

	// Verify both servers exist
	var doc map[string]any
	if err := json.Unmarshal(out, &doc); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	servers := doc["servers"].(map[string]any)
	if _, ok := servers["postgres"]; !ok {
		t.Error("existing 'postgres' server should be preserved")
	}
	if _, ok := servers["github"]; !ok {
		t.Error("new 'github' server should be added")
	}
}

func TestMerge_SkipIdentical(t *testing.T) {
	existing := []byte(`{
  "servers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"]
    }
  }
}`)

	translated := map[string]any{
		"github": map[string]any{
			"command": "npx",
			"args":    []any{"-y", "@modelcontextprotocol/server-github"},
		},
	}

	_, result, err := Merge(existing, translated, "servers", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Added) != 0 {
		t.Errorf("Added should be empty: %v", result.Added)
	}
	if len(result.Skipped) != 1 || result.Skipped[0] != "github" {
		t.Errorf("Skipped: got %v, want [github]", result.Skipped)
	}
}

func TestMerge_ConflictNoForce(t *testing.T) {
	existing := []byte(`{
  "servers": {
    "github": {
      "command": "different-command"
    }
  }
}`)

	translated := map[string]any{
		"github": map[string]any{
			"command": "npx",
		},
	}

	_, result, err := Merge(existing, translated, "servers", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Conflict) != 1 || result.Conflict[0] != "github" {
		t.Errorf("Conflict: got %v, want [github]", result.Conflict)
	}
	if len(result.Added) != 0 {
		t.Errorf("Added should be empty when conflict without force: %v", result.Added)
	}
}

func TestMerge_ConflictWithForce(t *testing.T) {
	existing := []byte(`{
  "servers": {
    "github": {
      "command": "different-command"
    }
  }
}`)

	translated := map[string]any{
		"github": map[string]any{
			"command": "npx",
		},
	}

	out, result, err := Merge(existing, translated, "servers", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Added) != 1 || result.Added[0] != "github" {
		t.Errorf("Added: got %v, want [github] (force should override)", result.Added)
	}
	if len(result.Conflict) != 0 {
		t.Errorf("Conflict should be empty with force: %v", result.Conflict)
	}

	// Verify the new value was written
	var doc map[string]any
	if err := json.Unmarshal(out, &doc); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	servers := doc["servers"].(map[string]any)
	github := servers["github"].(map[string]any)
	if github["command"] != "npx" {
		t.Errorf("command should be overwritten to 'npx', got %v", github["command"])
	}
}

func TestMerge_PreserveNonMCPKeys(t *testing.T) {
	// Simulate opencode.json which has MCP alongside other config
	existing := []byte(`{
  "theme": "dark",
  "editor": {"font_size": 14},
  "mcp": {
    "existing": {"type": "local", "command": ["echo"]}
  }
}`)

	translated := map[string]any{
		"github": map[string]any{
			"type":    "local",
			"command": []string{"npx", "-y", "@modelcontextprotocol/server-github"},
			"enabled": true,
		},
	}

	out, _, err := Merge(existing, translated, "mcp", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var doc map[string]any
	if err := json.Unmarshal(out, &doc); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Non-MCP keys should be preserved
	if doc["theme"] != "dark" {
		t.Errorf("theme should be preserved, got %v", doc["theme"])
	}
	editor, ok := doc["editor"].(map[string]any)
	if !ok {
		t.Fatal("editor key should be preserved")
	}
	if editor["font_size"] != float64(14) {
		t.Errorf("editor.font_size should be 14, got %v", editor["font_size"])
	}

	// MCP servers should include both old and new
	mcpServers := doc["mcp"].(map[string]any)
	if _, ok := mcpServers["existing"]; !ok {
		t.Error("existing MCP server should be preserved")
	}
	if _, ok := mcpServers["github"]; !ok {
		t.Error("new github MCP server should be added")
	}
}

func TestMerge_EmptyEnvVarWarnings(t *testing.T) {
	translated := map[string]any{
		"github": map[string]any{
			"command": "npx",
			"env": map[string]string{
				"GITHUB_TOKEN": "",
				"OTHER_VAR":    "set",
			},
		},
	}

	_, result, err := Merge(nil, translated, "servers", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d: %v", len(result.Warnings), result.Warnings)
	}
	if !strings.Contains(result.Warnings[0], "GITHUB_TOKEN") {
		t.Errorf("warning should mention GITHUB_TOKEN, got: %s", result.Warnings[0])
	}
}

func TestMerge_MalformedExistingJSON(t *testing.T) {
	_, _, err := Merge([]byte("not json"), map[string]any{"x": 1}, "servers", false)
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
	if !strings.Contains(err.Error(), "failed to parse existing config") {
		t.Errorf("error should mention parsing failure, got: %v", err)
	}
}

func TestMerge_EmptyExistingFile(t *testing.T) {
	translated := map[string]any{
		"github": map[string]any{"command": "npx"},
	}

	out, result, err := Merge([]byte{}, translated, "servers", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Added) != 1 {
		t.Errorf("Added: got %v, want [github]", result.Added)
	}

	var doc map[string]any
	if err := json.Unmarshal(out, &doc); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
}

func TestMerge_TwoSpaceIndent(t *testing.T) {
	translated := map[string]any{
		"github": map[string]any{"command": "npx"},
	}

	out, _, err := Merge(nil, translated, "servers", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check indentation uses 2 spaces
	if !strings.Contains(string(out), "  \"servers\"") {
		t.Errorf("expected 2-space indent, got:\n%s", out)
	}
	// Check trailing newline
	if out[len(out)-1] != '\n' {
		t.Error("expected trailing newline")
	}
}

func TestMerge_McpServersKey(t *testing.T) {
	// Test with Claude's "mcpServers" key
	translated := map[string]any{
		"github": map[string]any{"command": "npx"},
	}

	out, _, err := Merge(nil, translated, "mcpServers", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var doc map[string]any
	if err := json.Unmarshal(out, &doc); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := doc["mcpServers"]; !ok {
		t.Error("expected 'mcpServers' key in output")
	}
}

// TestMerge_EmptyEnvVarWarningsOnlyForAdded verifies that empty env var
// warnings are only emitted for servers that were actually added, not for
// servers that were skipped (identical) or conflicted.
func TestMerge_EmptyEnvVarWarningsOnlyForAdded(t *testing.T) {
	// Existing config has "existing-server" with a different command
	existing := []byte(`{
  "servers": {
    "existing-server": {
      "command": "old-cmd",
      "env": { "OLD_TOKEN": "set" }
    }
  }
}`)

	translated := map[string]any{
		// This server conflicts (different command) — should NOT warn
		"existing-server": map[string]any{
			"command": "new-cmd",
			"env": map[string]any{
				"EMPTY_VAR": "",
			},
		},
		// This server is new — SHOULD warn
		"new-server": map[string]any{
			"command": "npx",
			"env": map[string]any{
				"NEW_EMPTY": "",
			},
		},
	}

	_, result, err := Merge(existing, translated, "servers", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// existing-server should be in Conflict (not Added)
	if len(result.Conflict) != 1 || result.Conflict[0] != "existing-server" {
		t.Errorf("Conflict = %v, want [existing-server]", result.Conflict)
	}

	// Only new-server should have a warning
	if len(result.Warnings) != 1 {
		t.Fatalf("expected 1 warning (for new-server only), got %d: %v", len(result.Warnings), result.Warnings)
	}
	if !strings.Contains(result.Warnings[0], "new-server") {
		t.Errorf("warning should mention new-server, got: %s", result.Warnings[0])
	}
	if strings.Contains(result.Warnings[0], "existing-server") {
		t.Errorf("warning should NOT mention existing-server (conflicted), got: %s", result.Warnings[0])
	}
}

func TestMergeCanonical_NilInputs(t *testing.T) {
	result := MergeCanonical(nil, nil)
	if result != nil {
		t.Errorf("expected nil, got %+v", result)
	}
}

func TestMergeCanonical_SingleConfig(t *testing.T) {
	cfg := &Config{
		Servers: map[string]Server{
			"github": {Transport: TransportStdio, Command: "npx"},
		},
	}
	result := MergeCanonical(cfg)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Servers) != 1 {
		t.Errorf("expected 1 server, got %d", len(result.Servers))
	}
	if result.Servers["github"].Command != "npx" {
		t.Errorf("github.Command: got %q", result.Servers["github"].Command)
	}
}

func TestMergeCanonical_TeamOverridesPackage(t *testing.T) {
	packageCfg := &Config{
		Servers: map[string]Server{
			"github": {
				Transport: TransportStdio,
				Command:   "npx",
				Args:      []string{"-y", "@mcp/server-github"},
			},
			"filesystem": {
				Transport: TransportStdio,
				Command:   "npx",
				Args:      []string{"-y", "@mcp/server-fs"},
			},
		},
	}
	teamCfg := &Config{
		Servers: map[string]Server{
			"github": {
				Transport: TransportStdio,
				Command:   "npx",
				Args:      []string{"-y", "@mcp/server-github"},
				Env:       map[string]string{"GITHUB_TOKEN": ""},
			},
			"postgres": {
				Transport: TransportStdio,
				Command:   "npx",
				Args:      []string{"-y", "@mcp/server-postgres"},
			},
		},
	}

	// Package first, then team — team takes priority.
	result := MergeCanonical(packageCfg, teamCfg)
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Should have 3 unique servers
	if len(result.Servers) != 3 {
		t.Errorf("expected 3 servers, got %d", len(result.Servers))
	}

	// github: team version should win (has Env)
	gh := result.Servers["github"]
	if gh.Env == nil || gh.Env["GITHUB_TOKEN"] != "" {
		t.Errorf("github should have team-level env override, got %+v", gh)
	}

	// filesystem: from package only
	if _, ok := result.Servers["filesystem"]; !ok {
		t.Error("expected filesystem server from package")
	}

	// postgres: from team only
	if _, ok := result.Servers["postgres"]; !ok {
		t.Error("expected postgres server from team")
	}
}

func TestMergeCanonical_MultiplePackages(t *testing.T) {
	pkg1 := &Config{
		Servers: map[string]Server{
			"github": {Transport: TransportStdio, Command: "npx"},
		},
	}
	pkg2 := &Config{
		Servers: map[string]Server{
			"postgres": {Transport: TransportStdio, Command: "npx"},
		},
	}
	team := &Config{
		Servers: map[string]Server{
			"sentry": {Transport: TransportSSE, URL: "https://sentry.io/mcp"},
		},
	}

	result := MergeCanonical(pkg1, pkg2, team)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Servers) != 3 {
		t.Errorf("expected 3 servers, got %d", len(result.Servers))
	}
}

func TestMergeCanonical_EmptyConfigs(t *testing.T) {
	empty := &Config{Servers: map[string]Server{}}
	result := MergeCanonical(empty, nil, empty)
	if result != nil {
		t.Errorf("expected nil for all-empty configs, got %+v", result)
	}
}
