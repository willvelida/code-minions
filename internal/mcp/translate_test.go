package mcp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- End-to-end translation tests ---

func TestTranslate_CopilotToClaude(t *testing.T) {
	dir := t.TempDir()

	// Write source Copilot config
	copilotDir := filepath.Join(dir, ".vscode")
	if err := os.MkdirAll(copilotDir, 0755); err != nil {
		t.Fatal(err)
	}
	copilotConfig := `{
  "servers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": { "GITHUB_PERSONAL_ACCESS_TOKEN": "ghp_xxx" }
    }
  }
}`
	if err := os.WriteFile(filepath.Join(copilotDir, "mcp.json"), []byte(copilotConfig), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Translate(TranslateOptions{
		From:      "copilot",
		To:        "claude",
		TargetDir: dir,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Merge.Added) != 1 || result.Merge.Added[0] != "github" {
		t.Errorf("expected 1 added server 'github', got %v", result.Merge.Added)
	}
	if result.ConfigPath != ".claude/settings.local.json" {
		t.Errorf("config path: got %q", result.ConfigPath)
	}

	// Verify the target file was written
	targetData, err := os.ReadFile(filepath.Join(dir, ".claude", "settings.local.json"))
	if err != nil {
		t.Fatalf("target file not written: %v", err)
	}
	if !strings.Contains(string(targetData), "mcpServers") {
		t.Error("target file missing mcpServers key")
	}
	if !strings.Contains(string(targetData), "github") {
		t.Error("target file missing github server")
	}
}

func TestTranslate_ClaudeToOpenCode(t *testing.T) {
	dir := t.TempDir()

	// Write source Claude config
	claudeDir := filepath.Join(dir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}
	claudeConfig := `{
  "permissions": { "allow": ["read"] },
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": { "GITHUB_PERSONAL_ACCESS_TOKEN": "" }
    }
  }
}`
	if err := os.WriteFile(filepath.Join(claudeDir, "settings.local.json"), []byte(claudeConfig), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Translate(TranslateOptions{
		From:      "claude",
		To:        "opencode",
		TargetDir: dir,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Merge.Added) != 1 {
		t.Errorf("expected 1 added server, got %d", len(result.Merge.Added))
	}

	// Verify OpenCode format
	targetData, err := os.ReadFile(filepath.Join(dir, "opencode.json"))
	if err != nil {
		t.Fatalf("target file not written: %v", err)
	}
	content := string(targetData)
	if !strings.Contains(content, `"mcp"`) {
		t.Error("target file missing mcp key")
	}
	// OpenCode uses "environment" not "env"
	if !strings.Contains(content, `"environment"`) {
		t.Error("target file should use 'environment' key for OpenCode")
	}
}

func TestTranslate_OpenCodeToCopilot(t *testing.T) {
	dir := t.TempDir()

	// Write source OpenCode config
	openCodeConfig := `{
  "mcp": {
    "github": {
      "type": "local",
      "command": ["npx", "-y", "@modelcontextprotocol/server-github"],
      "environment": { "GITHUB_PERSONAL_ACCESS_TOKEN": "token" },
      "enabled": true
    }
  }
}`
	if err := os.WriteFile(filepath.Join(dir, "opencode.json"), []byte(openCodeConfig), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Translate(TranslateOptions{
		From:      "opencode",
		To:        "copilot",
		TargetDir: dir,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Merge.Added) != 1 {
		t.Errorf("expected 1 added server, got %d", len(result.Merge.Added))
	}

	// Verify Copilot format — command should be split from array
	targetData, err := os.ReadFile(filepath.Join(dir, ".vscode", "mcp.json"))
	if err != nil {
		t.Fatalf("target file not written: %v", err)
	}
	content := string(targetData)
	if !strings.Contains(content, `"servers"`) {
		t.Error("target file missing servers key")
	}
	if !strings.Contains(content, `"command"`) {
		t.Error("target file missing command key")
	}
}

// --- Server filter ---

func TestTranslate_ServerFilter(t *testing.T) {
	dir := t.TempDir()

	copilotDir := filepath.Join(dir, ".vscode")
	if err := os.MkdirAll(copilotDir, 0755); err != nil {
		t.Fatal(err)
	}
	copilotConfig := `{
  "servers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"]
    },
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem"]
    }
  }
}`
	if err := os.WriteFile(filepath.Join(copilotDir, "mcp.json"), []byte(copilotConfig), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Translate(TranslateOptions{
		From:      "copilot",
		To:        "claude",
		TargetDir: dir,
		Server:    "github",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Merge.Added) != 1 || result.Merge.Added[0] != "github" {
		t.Errorf("expected only 'github', got %v", result.Merge.Added)
	}
}

func TestTranslate_ServerFilterNotFound(t *testing.T) {
	dir := t.TempDir()

	copilotDir := filepath.Join(dir, ".vscode")
	if err := os.MkdirAll(copilotDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(copilotDir, "mcp.json"), []byte(`{"servers":{"github":{"command":"npx"}}}`), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Translate(TranslateOptions{
		From:      "copilot",
		To:        "claude",
		TargetDir: dir,
		Server:    "nonexistent",
	})
	if err == nil {
		t.Fatal("expected error for non-existent server")
	}
	if !strings.Contains(err.Error(), "nonexistent") {
		t.Errorf("error should mention server name: %v", err)
	}
}

// --- Force flag ---

func TestTranslate_ForceOverwrite(t *testing.T) {
	dir := t.TempDir()

	// Source
	copilotDir := filepath.Join(dir, ".vscode")
	if err := os.MkdirAll(copilotDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(copilotDir, "mcp.json"), []byte(`{
  "servers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github-NEW"]
    }
  }
}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Pre-existing target with different config for same server
	claudeDir := filepath.Join(dir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(claudeDir, "settings.local.json"), []byte(`{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github-OLD"]
    }
  }
}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Without force — should conflict
	result, err := Translate(TranslateOptions{
		From:      "copilot",
		To:        "claude",
		TargetDir: dir,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Merge.Conflict) != 1 {
		t.Errorf("expected 1 conflict, got %d", len(result.Merge.Conflict))
	}

	// With force — should overwrite
	result, err = Translate(TranslateOptions{
		From:      "copilot",
		To:        "claude",
		TargetDir: dir,
		Force:     true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Merge.Added) != 1 {
		t.Errorf("expected 1 added (force overwrite), got %d", len(result.Merge.Added))
	}
}

// --- Dry run ---

func TestTranslate_DryRun(t *testing.T) {
	dir := t.TempDir()

	copilotDir := filepath.Join(dir, ".vscode")
	if err := os.MkdirAll(copilotDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(copilotDir, "mcp.json"), []byte(`{
  "servers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"]
    }
  }
}`), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Translate(TranslateOptions{
		From:      "copilot",
		To:        "claude",
		TargetDir: dir,
		DryRun:    true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.DryRun {
		t.Error("expected DryRun to be true")
	}

	// Verify target was NOT written
	_, err = os.ReadFile(filepath.Join(dir, ".claude", "settings.local.json"))
	if err == nil {
		t.Error("target file should not exist in dry-run mode")
	}
}

// --- Error cases ---

func TestTranslate_SameSourceAndTarget(t *testing.T) {
	_, err := Translate(TranslateOptions{
		From: "copilot",
		To:   "copilot",
	})
	if err == nil {
		t.Fatal("expected error for same source and target")
	}
	if !strings.Contains(err.Error(), "cannot be the same") {
		t.Errorf("error message: %v", err)
	}
}

func TestTranslate_UnknownSourceAssistant(t *testing.T) {
	_, err := Translate(TranslateOptions{
		From: "unknown",
		To:   "copilot",
	})
	if err == nil {
		t.Fatal("expected error for unknown source")
	}
}

func TestTranslate_UnknownTargetAssistant(t *testing.T) {
	_, err := Translate(TranslateOptions{
		From: "copilot",
		To:   "unknown",
	})
	if err == nil {
		t.Fatal("expected error for unknown target")
	}
}

func TestTranslate_SourceConfigMissing(t *testing.T) {
	dir := t.TempDir()
	_, err := Translate(TranslateOptions{
		From:      "copilot",
		To:        "claude",
		TargetDir: dir,
	})
	if err == nil {
		t.Fatal("expected error for missing source config")
	}
	if !strings.Contains(err.Error(), "no MCP config found") {
		t.Errorf("error message: %v", err)
	}
}

func TestTranslate_SourceConfigNoServers(t *testing.T) {
	dir := t.TempDir()

	copilotDir := filepath.Join(dir, ".vscode")
	if err := os.MkdirAll(copilotDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(copilotDir, "mcp.json"), []byte(`{ "servers": {} }`), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Translate(TranslateOptions{
		From:      "copilot",
		To:        "claude",
		TargetDir: dir,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, w := range result.Warnings {
		if strings.Contains(w, "no MCP servers found") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected warning about no servers, got %v", result.Warnings)
	}
}

func TestTranslate_TargetConfigMissing_CreatesNew(t *testing.T) {
	dir := t.TempDir()

	copilotDir := filepath.Join(dir, ".vscode")
	if err := os.MkdirAll(copilotDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(copilotDir, "mcp.json"), []byte(`{
  "servers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"]
    }
  }
}`), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Translate(TranslateOptions{
		From:      "copilot",
		To:        "claude",
		TargetDir: dir,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// File should have been created
	if _, err := os.Stat(filepath.Join(dir, ".claude", "settings.local.json")); err != nil {
		t.Errorf("target file should have been created: %v", err)
	}
}

func TestTranslate_PreservesNonMCPContent(t *testing.T) {
	dir := t.TempDir()

	// Source
	copilotDir := filepath.Join(dir, ".vscode")
	if err := os.MkdirAll(copilotDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(copilotDir, "mcp.json"), []byte(`{
  "servers": {
    "github": { "command": "npx" }
  }
}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Pre-existing target with non-MCP content
	claudeDir := filepath.Join(dir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(claudeDir, "settings.local.json"), []byte(`{
  "permissions": { "allow": ["read", "write"] },
  "mcpServers": {}
}`), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Translate(TranslateOptions{
		From:      "copilot",
		To:        "claude",
		TargetDir: dir,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	targetData, err := os.ReadFile(filepath.Join(dir, ".claude", "settings.local.json"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(targetData)
	if !strings.Contains(content, "permissions") {
		t.Error("non-MCP content (permissions) was not preserved")
	}
	if !strings.Contains(content, "github") {
		t.Error("translated server was not added")
	}
}

// --- ListServers ---

func TestListServers_ValidConfig(t *testing.T) {
	dir := t.TempDir()

	copilotDir := filepath.Join(dir, ".vscode")
	if err := os.MkdirAll(copilotDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(copilotDir, "mcp.json"), []byte(`{
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
}`), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, _, err := ListServers("copilot", dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Servers) != 2 {
		t.Errorf("expected 2 servers, got %d", len(cfg.Servers))
	}
}

func TestListServers_NoConfigFile(t *testing.T) {
	dir := t.TempDir()

	cfg, _, err := ListServers("copilot", dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Servers) != 0 {
		t.Errorf("expected 0 servers, got %d", len(cfg.Servers))
	}
}

func TestListServers_UnknownAssistant(t *testing.T) {
	_, _, err := ListServers("unknown", ".")
	if err == nil {
		t.Fatal("expected error for unknown assistant")
	}
}
