package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fatih/color"
)

// --- helpers ---

func writeCopilotMCPConfig(t *testing.T, dir, content string) {
	t.Helper()
	d := filepath.Join(dir, ".vscode")
	if err := os.MkdirAll(d, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(d, "mcp.json"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func writeClaudeMCPConfig(t *testing.T, dir, content string) {
	t.Helper()
	d := filepath.Join(dir, ".claude")
	if err := os.MkdirAll(d, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(d, "settings.local.json"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

const testCopilotConfig = `{
  "servers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": { "GITHUB_PERSONAL_ACCESS_TOKEN": "" }
    }
  }
}`

// --- mcp translate tests ---

func TestMCPTranslateCommand_Basic(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	dir := t.TempDir()
	writeCopilotMCPConfig(t, dir, testCopilotConfig)

	var buf bytes.Buffer
	cmd := newMCPCommand()
	cmd.SetArgs([]string{"translate", "--from", "copilot", "--to", "claude", "--target", dir})
	cmd.SetOut(&buf)
	cmd.SilenceUsage = true

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "copilot") || !strings.Contains(output, "claude") {
		t.Errorf("output should mention source and target: %s", output)
	}
	if !strings.Contains(output, "github") {
		t.Errorf("output should mention added server: %s", output)
	}

	// Verify file was written
	if _, err := os.Stat(filepath.Join(dir, ".claude", "settings.local.json")); err != nil {
		t.Errorf("target file should have been created: %v", err)
	}
}

func TestMCPTranslateCommand_JSON(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	dir := t.TempDir()
	writeCopilotMCPConfig(t, dir, testCopilotConfig)

	var buf bytes.Buffer
	root := newMCPCommand()
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().BoolP("verbose", "v", false, "")
	root.PersistentFlags().BoolP("quiet", "q", false, "")
	root.SetArgs([]string{"translate", "--from", "copilot", "--to", "claude", "--target", dir, "--json"})
	root.SetOut(&buf)
	root.SilenceUsage = true

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output should be valid JSON: %v\noutput: %s", err, buf.String())
	}

	if result["sourceAssistant"] != "copilot" {
		t.Errorf("expected sourceAssistant=copilot, got %v", result["sourceAssistant"])
	}
}

func TestMCPTranslateCommand_DryRun(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	dir := t.TempDir()
	writeCopilotMCPConfig(t, dir, testCopilotConfig)

	var buf bytes.Buffer
	cmd := newMCPCommand()
	cmd.SetArgs([]string{"translate", "--from", "copilot", "--to", "claude", "--target", dir, "--dry-run"})
	cmd.SetOut(&buf)
	cmd.SilenceUsage = true

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "dry-run") {
		t.Errorf("output should indicate dry-run: %s", output)
	}

	// File should NOT have been created
	if _, err := os.Stat(filepath.Join(dir, ".claude", "settings.local.json")); err == nil {
		t.Error("target file should not exist in dry-run mode")
	}
}

func TestMCPTranslateCommand_MultipleTargets(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	dir := t.TempDir()
	writeCopilotMCPConfig(t, dir, testCopilotConfig)

	var buf bytes.Buffer
	cmd := newMCPCommand()
	cmd.SetArgs([]string{"translate", "--from", "copilot", "--to", "claude,opencode", "--target", dir})
	cmd.SetOut(&buf)
	cmd.SilenceUsage = true

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Both target files should exist
	if _, err := os.Stat(filepath.Join(dir, ".claude", "settings.local.json")); err != nil {
		t.Errorf("claude target file not created: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "opencode.json")); err != nil {
		t.Errorf("opencode target file not created: %v", err)
	}
}

func TestMCPTranslateCommand_MissingFromFlag(t *testing.T) {
	cmd := newMCPCommand()
	cmd.SetArgs([]string{"translate", "--to", "claude"})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when --from is missing")
	}
}

func TestMCPTranslateCommand_MissingToFlag(t *testing.T) {
	cmd := newMCPCommand()
	cmd.SetArgs([]string{"translate", "--from", "copilot"})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when --to is missing")
	}
}

func TestMCPTranslateCommand_SameSourceAndTarget(t *testing.T) {
	dir := t.TempDir()
	writeCopilotMCPConfig(t, dir, testCopilotConfig)

	cmd := newMCPCommand()
	cmd.SetArgs([]string{"translate", "--from", "copilot", "--to", "copilot", "--target", dir})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for same source and target")
	}
}

// --- mcp list tests ---

func TestMCPListCommand_Basic(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	dir := t.TempDir()
	writeCopilotMCPConfig(t, dir, testCopilotConfig)

	var buf bytes.Buffer
	cmd := newMCPCommand()
	cmd.SetArgs([]string{"list", "--for", "copilot", "--target", dir})
	cmd.SetOut(&buf)
	cmd.SilenceUsage = true

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "github") {
		t.Errorf("output should list 'github' server: %s", output)
	}
	if !strings.Contains(output, "GitHub Copilot") {
		t.Errorf("output should mention assistant name: %s", output)
	}
}

func TestMCPListCommand_JSON(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	dir := t.TempDir()
	writeCopilotMCPConfig(t, dir, testCopilotConfig)

	var buf bytes.Buffer
	root := newMCPCommand()
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().BoolP("verbose", "v", false, "")
	root.PersistentFlags().BoolP("quiet", "q", false, "")
	root.SetArgs([]string{"list", "--for", "copilot", "--target", dir, "--json"})
	root.SetOut(&buf)
	root.SilenceUsage = true

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output should be valid JSON: %v\noutput: %s", err, buf.String())
	}

	if result["assistant"] != "copilot" {
		t.Errorf("expected assistant=copilot, got %v", result["assistant"])
	}
}

func TestMCPListCommand_NoConfig(t *testing.T) {
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = false })

	dir := t.TempDir()

	var buf bytes.Buffer
	cmd := newMCPCommand()
	cmd.SetArgs([]string{"list", "--for", "copilot", "--target", dir})
	cmd.SetOut(&buf)
	cmd.SilenceUsage = true

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "no MCP servers configured") {
		t.Errorf("output should say no servers: %s", output)
	}
}

func TestMCPListCommand_MissingForFlag(t *testing.T) {
	cmd := newMCPCommand()
	cmd.SetArgs([]string{"list"})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when --for is missing")
	}
}
