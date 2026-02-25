package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/willvelida/code-minions/internal/installer"
)

func TestPrintDryRunInstallActions(t *testing.T) {
	actions := []installer.FileAction{
		{Path: "agents/my-agent.agent.md", Kind: installer.ActionCreate},
		{Path: "skills/my-skill/SKILL.md", Kind: installer.ActionCreate},
		{Path: "agents/AGENTS.md", Kind: installer.ActionModify, Annotation: "add reference to my-skill"},
		{Path: ".github/copilot-instructions.md", Kind: installer.ActionUnchanged, Annotation: "already up to date"},
		{Path: "skills/old-skill/SKILL.md", Kind: installer.ActionSkipped, Annotation: "exists (use --force to overwrite)"},
	}

	var buf bytes.Buffer
	printDryRunInstallActions(&buf, actions)
	out := buf.String()

	// Check section headers
	if !strings.Contains(out, "Would create:") {
		t.Error("missing 'Would create:' header")
	}
	if !strings.Contains(out, "Would modify:") {
		t.Error("missing 'Would modify:' header")
	}
	if !strings.Contains(out, "Would not change:") {
		t.Error("missing 'Would not change:' header")
	}
	if !strings.Contains(out, "Would skip (use --force to overwrite):") {
		t.Error("missing 'Would skip' header")
	}

	// Check prefixes
	if !strings.Contains(out, "+ agents/my-agent.agent.md") {
		t.Error("missing '+ agents/my-agent.agent.md'")
	}
	if !strings.Contains(out, "~ agents/AGENTS.md (add reference to my-skill)") {
		t.Error("missing '~ agents/AGENTS.md' with annotation")
	}
	if !strings.Contains(out, "= .github/copilot-instructions.md (already up to date)") {
		t.Error("missing '= .github/copilot-instructions.md' with annotation")
	}
	if !strings.Contains(out, "! skills/old-skill/SKILL.md (exists (use --force to overwrite))") {
		t.Error("missing '! skills/old-skill/SKILL.md' with annotation")
	}

	// Check summary
	if !strings.Contains(out, "Summary: 2 to create, 1 to modify, 1 skipped, 1 unchanged") {
		t.Errorf("unexpected summary in output:\n%s", out)
	}
}

func TestPrintDryRunInstallActionsEmptyCategories(t *testing.T) {
	// Only create actions — modify and unchanged headers should be absent
	actions := []installer.FileAction{
		{Path: "agents/my-agent.agent.md", Kind: installer.ActionCreate},
	}

	var buf bytes.Buffer
	printDryRunInstallActions(&buf, actions)
	out := buf.String()

	if !strings.Contains(out, "Would create:") {
		t.Error("missing 'Would create:' header")
	}
	if strings.Contains(out, "Would modify:") {
		t.Error("'Would modify:' should not appear when there are no modify actions")
	}
	if strings.Contains(out, "Would not change:") {
		t.Error("'Would not change:' should not appear when there are no unchanged actions")
	}
}

func TestPrintDryRunUninstallActions(t *testing.T) {
	actions := []installer.FileAction{
		{Path: "agents/my-agent.agent.md", Kind: installer.ActionRemove},
		{Path: "skills/my-skill/SKILL.md", Kind: installer.ActionRemove},
		{Path: "agents/nonexistent.agent.md", Kind: installer.ActionNotFound},
	}

	var buf bytes.Buffer
	printDryRunUninstallActions(&buf, actions)
	out := buf.String()

	if !strings.Contains(out, "Would remove:") {
		t.Error("missing 'Would remove:' header")
	}
	if !strings.Contains(out, "- agents/my-agent.agent.md") {
		t.Error("missing '- agents/my-agent.agent.md'")
	}
	if !strings.Contains(out, "? agents/nonexistent.agent.md") {
		t.Error("missing '? agents/nonexistent.agent.md'")
	}
	if !strings.Contains(out, "Summary: 2 to remove, 1 not found") {
		t.Errorf("unexpected summary in output:\n%s", out)
	}
}

func TestPrintDryRunUpdateActions(t *testing.T) {
	actions := []installer.FileAction{
		{Path: "skills/my-skill/SKILL.md", Kind: installer.ActionModify},
		{Path: "agents/my-agent.agent.md", Kind: installer.ActionUnchanged},
	}

	var buf bytes.Buffer
	printDryRunUpdateActions(&buf, actions)
	out := buf.String()

	if !strings.Contains(out, "Would update:") {
		t.Error("missing 'Would update:' header")
	}
	if !strings.Contains(out, "Already up to date:") {
		t.Error("missing 'Already up to date:' header")
	}
	if !strings.Contains(out, "Summary: 1 to update, 1 unchanged") {
		t.Errorf("unexpected summary in output:\n%s", out)
	}
}

func TestActionsToJSON(t *testing.T) {
	actions := []installer.FileAction{
		{Path: "agents/a.md", Kind: installer.ActionCreate},
		{Path: "agents/b.md", Kind: installer.ActionModify, Annotation: "updated"},
	}

	result := actionsToJSON(actions)
	if len(result) != 2 {
		t.Fatalf("expected 2 actions, got %d", len(result))
	}
	if result[0].Action != "create" {
		t.Errorf("expected action 'create', got %q", result[0].Action)
	}
	if result[1].Annotation != "updated" {
		t.Errorf("expected annotation 'updated', got %q", result[1].Annotation)
	}
}

func TestActionsToJSONEmpty(t *testing.T) {
	result := actionsToJSON(nil)
	if result == nil {
		t.Error("actionsToJSON(nil) should return empty slice, not nil")
	}
	if len(result) != 0 {
		t.Errorf("expected 0 actions, got %d", len(result))
	}
}
