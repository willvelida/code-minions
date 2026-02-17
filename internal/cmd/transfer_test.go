package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// seedTransferSource creates source files in the given target dir using
// the named assistant's layout. Returns the list of files created (relative).
func seedTransferSource(t *testing.T, dir, assistantName string) []string {
	t.Helper()

	layouts := map[string][]struct {
		path    string
		content string
	}{
		"copilot": {
			{path: ".github/agents/my-agent.agent.md", content: "# Copilot Agent"},
			{path: "skills/my-skill/SKILL.md", content: "# Copilot Skill"},
		},
		"claude": {
			{path: ".claude/agents/my-agent.agent.md", content: "# Claude Agent"},
			{path: ".claude/skills/my-skill/SKILL.md", content: "# Claude Skill"},
		},
		"opencode": {
			{path: ".opencode/agents/my-agent.agent.md", content: "# Opencode Agent"},
			{path: ".opencode/skills/my-skill/SKILL.md", content: "# Opencode Skill"},
		},
	}

	files, ok := layouts[assistantName]
	if !ok {
		t.Fatalf("unknown assistant %q", assistantName)
	}

	var created []string
	for _, f := range files {
		full := filepath.Join(dir, f.path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(f.content), 0o644); err != nil {
			t.Fatal(err)
		}
		created = append(created, f.path)
	}
	return created
}

// TestTransferFlagValidation checks that missing --from or --to produces errors.
func TestTransferFlagValidation(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		errMsg string
	}{
		{
			name:   "missing --from",
			args:   []string{"transfer", "--to", "claude"},
			errMsg: "--from is required",
		},
		{
			name:   "missing --to",
			args:   []string{"transfer", "--from", "copilot"},
			errMsg: "--to is required",
		},
		{
			name:   "same --from and --to",
			args:   []string{"transfer", "--from", "copilot", "--to", "copilot", "--target", t.TempDir()},
			errMsg: "cannot be the same",
		},
		{
			name:   "unknown --from",
			args:   []string{"transfer", "--from", "vscode", "--to", "claude", "--target", t.TempDir()},
			errMsg: "unknown assistant",
		},
		{
			name:   "unknown --to",
			args:   []string{"transfer", "--from", "copilot", "--to", "vscode", "--target", t.TempDir()},
			errMsg: "unknown assistant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := testContentFS()
			root := NewRootCommand(content)
			root.SilenceErrors = true
			root.SilenceUsage = true
			root.SetArgs(tt.args)

			err := root.Execute()
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("error = %q, want substring %q", err.Error(), tt.errMsg)
			}
		})
	}
}

// TestTransferCopilotToClaude verifies files are transferred from Copilot to Claude layout.
func TestTransferCopilotToClaude(t *testing.T) {
	target := t.TempDir()
	seedTransferSource(t, target, "copilot")

	content := testContentFS()
	root := NewRootCommand(content)
	root.SilenceErrors = true
	root.SilenceUsage = true
	root.SetArgs([]string{
		"transfer",
		"--from", "copilot",
		"--to", "claude",
		"--target", target,
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Target files should exist in Claude layout
	claudeAgent := filepath.Join(target, ".claude", "agents", "my-agent.agent.md")
	if _, err := os.Stat(claudeAgent); os.IsNotExist(err) {
		t.Errorf("expected agent at Claude path: %s", claudeAgent)
	}
	claudeSkill := filepath.Join(target, ".claude", "skills", "my-skill", "SKILL.md")
	if _, err := os.Stat(claudeSkill); os.IsNotExist(err) {
		t.Errorf("expected skill at Claude path: %s", claudeSkill)
	}

	// Source files should still exist (copy, not move)
	copilotAgent := filepath.Join(target, ".github", "agents", "my-agent.agent.md")
	if _, err := os.Stat(copilotAgent); os.IsNotExist(err) {
		t.Errorf("source agent should still exist after transfer: %s", copilotAgent)
	}
}

// TestTransferClaudeToOpencode verifies Claude → Opencode transfer works.
func TestTransferClaudeToOpencode(t *testing.T) {
	target := t.TempDir()
	seedTransferSource(t, target, "claude")

	content := testContentFS()
	root := NewRootCommand(content)
	root.SilenceErrors = true
	root.SilenceUsage = true
	root.SetArgs([]string{
		"transfer",
		"--from", "claude",
		"--to", "opencode",
		"--target", target,
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	opencodeAgent := filepath.Join(target, ".opencode", "agents", "my-agent.agent.md")
	if _, err := os.Stat(opencodeAgent); os.IsNotExist(err) {
		t.Errorf("expected agent at Opencode path: %s", opencodeAgent)
	}
	opencodeSkill := filepath.Join(target, ".opencode", "skills", "my-skill", "SKILL.md")
	if _, err := os.Stat(opencodeSkill); os.IsNotExist(err) {
		t.Errorf("expected skill at Opencode path: %s", opencodeSkill)
	}
}

// TestTransferDryRunDoesNotCopy ensures --dry-run previews without writing.
func TestTransferDryRunDoesNotCopy(t *testing.T) {
	target := t.TempDir()
	seedTransferSource(t, target, "copilot")

	content := testContentFS()
	root := NewRootCommand(content)
	root.SilenceErrors = true
	root.SilenceUsage = true
	root.SetArgs([]string{
		"transfer",
		"--from", "copilot",
		"--to", "claude",
		"--target", target,
		"--dry-run",
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Target should NOT have been written
	claudeAgent := filepath.Join(target, ".claude", "agents", "my-agent.agent.md")
	if _, err := os.Stat(claudeAgent); !os.IsNotExist(err) {
		t.Errorf("dry-run should not create files, but found: %s", claudeAgent)
	}
}

// TestTransferCleanup verifies --cleanup removes source files after transfer.
func TestTransferCleanup(t *testing.T) {
	target := t.TempDir()
	seedTransferSource(t, target, "copilot")

	content := testContentFS()
	root := NewRootCommand(content)
	root.SilenceErrors = true
	root.SilenceUsage = true
	root.SetArgs([]string{
		"transfer",
		"--from", "copilot",
		"--to", "claude",
		"--target", target,
		"--cleanup",
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Target files exist
	claudeAgent := filepath.Join(target, ".claude", "agents", "my-agent.agent.md")
	if _, err := os.Stat(claudeAgent); os.IsNotExist(err) {
		t.Errorf("expected agent at Claude path: %s", claudeAgent)
	}

	// Source files should be gone
	copilotAgent := filepath.Join(target, ".github", "agents", "my-agent.agent.md")
	if _, err := os.Stat(copilotAgent); !os.IsNotExist(err) {
		t.Errorf("source agent should be removed after --cleanup: %s", copilotAgent)
	}
}

// TestTransferForceOverwrite verifies --force overwrites existing destination files.
func TestTransferForceOverwrite(t *testing.T) {
	target := t.TempDir()

	// Seed source (copilot)
	seedTransferSource(t, target, "copilot")

	// Pre-create a destination file with different content
	claudeAgent := filepath.Join(target, ".claude", "agents", "my-agent.agent.md")
	if err := os.MkdirAll(filepath.Dir(claudeAgent), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(claudeAgent, []byte("old content"), 0o644); err != nil {
		t.Fatal(err)
	}

	content := testContentFS()
	root := NewRootCommand(content)
	root.SilenceErrors = true
	root.SilenceUsage = true
	root.SetArgs([]string{
		"transfer",
		"--from", "copilot",
		"--to", "claude",
		"--target", target,
		"--force",
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// File should be overwritten with source content
	data, err := os.ReadFile(claudeAgent)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "# Copilot Agent" {
		t.Errorf("agent content = %q, want %q", string(data), "# Copilot Agent")
	}
}

// TestTransferSkipExisting verifies that without --force, existing files are skipped.
func TestTransferSkipExisting(t *testing.T) {
	target := t.TempDir()

	// Seed source
	seedTransferSource(t, target, "copilot")

	// Pre-create a destination file
	claudeAgent := filepath.Join(target, ".claude", "agents", "my-agent.agent.md")
	if err := os.MkdirAll(filepath.Dir(claudeAgent), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(claudeAgent, []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}

	content := testContentFS()
	root := NewRootCommand(content)
	root.SilenceErrors = true
	root.SilenceUsage = true
	root.SetArgs([]string{
		"transfer",
		"--from", "copilot",
		"--to", "claude",
		"--target", target,
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Original content should be preserved
	data, err := os.ReadFile(claudeAgent)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "existing" {
		t.Errorf("agent content = %q, want %q (should not be overwritten)", string(data), "existing")
	}
}

// TestTransferJSONOutput verifies the --json flag produces structured output.
func TestTransferJSONOutput(t *testing.T) {
	target := t.TempDir()
	seedTransferSource(t, target, "copilot")

	content := testContentFS()
	root := NewRootCommand(content)
	root.SilenceErrors = true
	root.SilenceUsage = true

	var buf strings.Builder
	root.SetOut(&buf)
	root.SetArgs([]string{
		"transfer",
		"--from", "copilot",
		"--to", "claude",
		"--target", target,
		"--json",
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		From    string `json:"from"`
		To      string `json:"to"`
		DryRun  bool   `json:"dry_run"`
		Cleanup bool   `json:"cleanup"`
		Files   struct {
			Copied   []string `json:"copied"`
			Skipped  []string `json:"skipped"`
			Errors   []string `json:"errors"`
			Cleaned  []string `json:"cleaned"`
			Warnings []string `json:"warnings"`
		} `json:"files"`
	}
	if err := json.Unmarshal([]byte(buf.String()), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, buf.String())
	}

	if result.From != "copilot" {
		t.Errorf("from = %q, want %q", result.From, "copilot")
	}
	if result.To != "claude" {
		t.Errorf("to = %q, want %q", result.To, "claude")
	}
	if len(result.Files.Copied) == 0 {
		t.Error("expected at least one copied file")
	}
	if result.DryRun {
		t.Error("dry_run should be false")
	}
}

// TestTransferJSONDryRun verifies JSON output has dry_run: true when --dry-run is set.
func TestTransferJSONDryRun(t *testing.T) {
	target := t.TempDir()
	seedTransferSource(t, target, "copilot")

	content := testContentFS()
	root := NewRootCommand(content)
	root.SilenceErrors = true
	root.SilenceUsage = true

	var buf strings.Builder
	root.SetOut(&buf)
	root.SetArgs([]string{
		"transfer",
		"--from", "copilot",
		"--to", "claude",
		"--target", target,
		"--json",
		"--dry-run",
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		DryRun bool `json:"dry_run"`
		Files  struct {
			Copied []string `json:"copied"`
		} `json:"files"`
	}
	if err := json.Unmarshal([]byte(buf.String()), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, buf.String())
	}
	if !result.DryRun {
		t.Error("dry_run should be true")
	}
	if len(result.Files.Copied) == 0 {
		t.Error("expected at least one file in copied list (dry-run tracks what would be copied)")
	}
}

// TestTransferQuietMode verifies quiet mode suppresses normal output but reports errors.
func TestTransferQuietMode(t *testing.T) {
	target := t.TempDir()
	seedTransferSource(t, target, "copilot")

	content := testContentFS()
	root := NewRootCommand(content)
	root.SilenceErrors = true
	root.SilenceUsage = true
	root.SetArgs([]string{
		"transfer",
		"--from", "copilot",
		"--to", "claude",
		"--target", target,
		"--quiet",
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the files were actually transferred despite quiet mode
	claudeAgent := filepath.Join(target, ".claude", "agents", "my-agent.agent.md")
	if _, err := os.Stat(claudeAgent); os.IsNotExist(err) {
		t.Errorf("expected agent at Claude path: %s", claudeAgent)
	}
}

// TestTransferMissingSourceDirsErrors verifies that missing source dirs produce an error.
func TestTransferMissingSourceDirsErrors(t *testing.T) {
	target := t.TempDir()
	// Don't seed any files — both source dirs are missing

	content := testContentFS()
	root := NewRootCommand(content)
	root.SilenceErrors = true
	root.SilenceUsage = true
	root.SetArgs([]string{
		"transfer",
		"--from", "copilot",
		"--to", "claude",
		"--target", target,
	})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when both source dirs are missing")
	}
	if !strings.Contains(err.Error(), "no agent or skill files found") {
		t.Errorf("error = %q, want substring 'no agent or skill files found'", err.Error())
	}
}

// TestTransferVerboseMode exercises the verbose output rendering path.
func TestTransferVerboseMode(t *testing.T) {
	target := t.TempDir()
	seedTransferSource(t, target, "copilot")

	// Pre-create one destination file so we get a skip + verbose hint
	claudeAgent := filepath.Join(target, ".claude", "agents", "my-agent.agent.md")
	if err := os.MkdirAll(filepath.Dir(claudeAgent), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(claudeAgent, []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}

	content := testContentFS()
	root := NewRootCommand(content)
	root.SilenceErrors = true
	root.SilenceUsage = true
	root.SetArgs([]string{
		"transfer",
		"--from", "copilot",
		"--to", "claude",
		"--target", target,
		"--verbose",
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify transfer still works — skill should be copied
	claudeSkill := filepath.Join(target, ".claude", "skills", "my-skill", "SKILL.md")
	if _, err := os.Stat(claudeSkill); os.IsNotExist(err) {
		t.Errorf("expected skill at Claude path: %s", claudeSkill)
	}
}

// TestTransferJSONWithCleanup verifies JSON output includes cleaned files.
func TestTransferJSONWithCleanup(t *testing.T) {
	target := t.TempDir()
	seedTransferSource(t, target, "copilot")

	content := testContentFS()
	root := NewRootCommand(content)
	root.SilenceErrors = true
	root.SilenceUsage = true

	var buf strings.Builder
	root.SetOut(&buf)
	root.SetArgs([]string{
		"transfer",
		"--from", "copilot",
		"--to", "claude",
		"--target", target,
		"--json",
		"--cleanup",
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Cleanup bool `json:"cleanup"`
		Files   struct {
			Copied  []string `json:"copied"`
			Cleaned []string `json:"cleaned"`
		} `json:"files"`
	}
	if err := json.Unmarshal([]byte(buf.String()), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, buf.String())
	}
	if !result.Cleanup {
		t.Error("cleanup should be true")
	}
	if len(result.Files.Cleaned) == 0 {
		t.Error("expected at least one cleaned file")
	}
	if len(result.Files.Copied) == 0 {
		t.Error("expected at least one copied file")
	}
}

// TestTransferCleanupDryRun verifies --cleanup with --dry-run previews deletions.
func TestTransferCleanupDryRun(t *testing.T) {
	target := t.TempDir()
	seedTransferSource(t, target, "copilot")

	content := testContentFS()
	root := NewRootCommand(content)
	root.SilenceErrors = true
	root.SilenceUsage = true
	root.SetArgs([]string{
		"transfer",
		"--from", "copilot",
		"--to", "claude",
		"--target", target,
		"--cleanup",
		"--dry-run",
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Source files should still exist (dry-run)
	copilotAgent := filepath.Join(target, ".github", "agents", "my-agent.agent.md")
	if _, err := os.Stat(copilotAgent); os.IsNotExist(err) {
		t.Errorf("source should still exist after --cleanup --dry-run: %s", copilotAgent)
	}

	// Destination should NOT exist (dry-run)
	claudeAgent := filepath.Join(target, ".claude", "agents", "my-agent.agent.md")
	if _, err := os.Stat(claudeAgent); !os.IsNotExist(err) {
		t.Errorf("destination should not exist after --dry-run: %s", claudeAgent)
	}
}

// TestTransferJSONSkippedFiles verifies JSON output includes skipped files.
func TestTransferJSONSkippedFiles(t *testing.T) {
	target := t.TempDir()
	seedTransferSource(t, target, "copilot")

	// Pre-create destination agent to trigger a skip
	claudeAgent := filepath.Join(target, ".claude", "agents", "my-agent.agent.md")
	if err := os.MkdirAll(filepath.Dir(claudeAgent), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(claudeAgent, []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}

	content := testContentFS()
	root := NewRootCommand(content)
	root.SilenceErrors = true
	root.SilenceUsage = true

	var buf strings.Builder
	root.SetOut(&buf)
	root.SetArgs([]string{
		"transfer",
		"--from", "copilot",
		"--to", "claude",
		"--target", target,
		"--json",
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Files struct {
			Copied  []string `json:"copied"`
			Skipped []string `json:"skipped"`
		} `json:"files"`
	}
	if err := json.Unmarshal([]byte(buf.String()), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, buf.String())
	}
	if len(result.Files.Skipped) == 0 {
		t.Error("expected at least one skipped file")
	}
}

// TestTransferOpencodeTocopilot exercises a third permutation at the CLI level.
func TestTransferOpencodeToCopilot(t *testing.T) {
	target := t.TempDir()
	seedTransferSource(t, target, "opencode")

	content := testContentFS()
	root := NewRootCommand(content)
	root.SilenceErrors = true
	root.SilenceUsage = true
	root.SetArgs([]string{
		"transfer",
		"--from", "opencode",
		"--to", "copilot",
		"--target", target,
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	copilotAgent := filepath.Join(target, ".github", "agents", "my-agent.agent.md")
	if _, err := os.Stat(copilotAgent); os.IsNotExist(err) {
		t.Errorf("expected agent at Copilot path: %s", copilotAgent)
	}
	copilotSkill := filepath.Join(target, "skills", "my-skill", "SKILL.md")
	if _, err := os.Stat(copilotSkill); os.IsNotExist(err) {
		t.Errorf("expected skill at Copilot path: %s", copilotSkill)
	}
}

// TestTransferRegeneratesAgentsMD verifies that an existing stale AGENTS.md
// in the target layout is overwritten with fresh content during transfer.
func TestTransferRegeneratesAgentsMD(t *testing.T) {
	target := t.TempDir()
	seedTransferSource(t, target, "copilot")

	// Place a stale AGENTS.md in the target Claude layout
	claudeAgentsDir := filepath.Join(target, ".claude", "agents")
	if err := os.MkdirAll(claudeAgentsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	staleContent := "# Stale old content that should be replaced"
	if err := os.WriteFile(filepath.Join(claudeAgentsDir, "AGENTS.md"), []byte(staleContent), 0o644); err != nil {
		t.Fatal(err)
	}

	content := testContentFS()
	root := NewRootCommand(content)
	root.SilenceErrors = true
	root.SilenceUsage = true
	root.SetArgs([]string{
		"transfer",
		"--from", "copilot",
		"--to", "claude",
		"--target", target,
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// AGENTS.md should have been regenerated, not kept stale
	agentsMDPath := filepath.Join(claudeAgentsDir, "AGENTS.md")
	data, err := os.ReadFile(agentsMDPath)
	if err != nil {
		t.Fatalf("expected AGENTS.md to exist: %v", err)
	}
	if string(data) == staleContent {
		t.Error("AGENTS.md should have been regenerated, but still contains stale content")
	}
	if len(data) == 0 {
		t.Error("AGENTS.md should not be empty after regeneration")
	}
}
