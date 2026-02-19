package assistant

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectEmptyDirectory(t *testing.T) {
	dir := t.TempDir()

	found := Detect(dir)
	if len(found) != 0 {
		t.Errorf("expected no assistants in empty dir, got %v", found)
	}
}

func TestDetectCopilotByInstructionsFile(t *testing.T) {
	dir := t.TempDir()

	// Create .github/copilot-instructions.md
	ghDir := filepath.Join(dir, ".github")
	if err := os.MkdirAll(ghDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ghDir, "copilot-instructions.md"), []byte("# Instructions\n"), 0644); err != nil {
		t.Fatal(err)
	}

	found := Detect(dir)
	if len(found) != 1 || found[0] != "copilot" {
		t.Errorf("expected [copilot], got %v", found)
	}
}

func TestDetectCopilotByAgentDir(t *testing.T) {
	dir := t.TempDir()

	// Create .github/agents/ directory only (no instructions file)
	agentDir := filepath.Join(dir, ".github", "agents")
	if err := os.MkdirAll(agentDir, 0755); err != nil {
		t.Fatal(err)
	}

	found := Detect(dir)
	if len(found) != 1 || found[0] != "copilot" {
		t.Errorf("expected [copilot], got %v", found)
	}
}

func TestDetectClaudeByInstructionsFile(t *testing.T) {
	dir := t.TempDir()

	// Create CLAUDE.md
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# Claude\n"), 0644); err != nil {
		t.Fatal(err)
	}

	found := Detect(dir)
	if len(found) != 1 || found[0] != "claude" {
		t.Errorf("expected [claude], got %v", found)
	}
}

func TestDetectMultipleAssistants(t *testing.T) {
	dir := t.TempDir()

	// Create markers for copilot and claude
	ghDir := filepath.Join(dir, ".github")
	if err := os.MkdirAll(ghDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ghDir, "copilot-instructions.md"), []byte("# Copilot\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# Claude\n"), 0644); err != nil {
		t.Fatal(err)
	}

	found := Detect(dir)
	if len(found) != 2 {
		t.Fatalf("expected 2 assistants, got %d: %v", len(found), found)
	}
	if found[0] != "claude" || found[1] != "copilot" {
		t.Errorf("expected [claude, copilot], got %v", found)
	}
}

func TestDetectByMCPConfigPath(t *testing.T) {
	dir := t.TempDir()

	// Create only the MCP config for opencode (opencode.json)
	if err := os.WriteFile(filepath.Join(dir, "opencode.json"), []byte("{}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	found := Detect(dir)
	if len(found) != 1 || found[0] != "opencode" {
		t.Errorf("expected [opencode], got %v", found)
	}
}

func TestDetectGeminiByInstructionsFile(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "GEMINI.md"), []byte("# Gemini\n"), 0644); err != nil {
		t.Fatal(err)
	}

	found := Detect(dir)
	if len(found) != 1 || found[0] != "gemini" {
		t.Errorf("expected [gemini], got %v", found)
	}
}

func TestDetectCursorByAgentDir(t *testing.T) {
	dir := t.TempDir()

	cursorAgents := filepath.Join(dir, ".cursor", "agents")
	if err := os.MkdirAll(cursorAgents, 0755); err != nil {
		t.Fatal(err)
	}

	found := Detect(dir)
	if len(found) != 1 || found[0] != "cursor" {
		t.Errorf("expected [cursor], got %v", found)
	}
}

func TestDetectResultIsSorted(t *testing.T) {
	dir := t.TempDir()

	// Create markers for gemini, copilot, and claude — result should be alphabetical
	if err := os.WriteFile(filepath.Join(dir, "GEMINI.md"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	ghDir := filepath.Join(dir, ".github")
	if err := os.MkdirAll(ghDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ghDir, "copilot-instructions.md"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	found := Detect(dir)
	if len(found) < 3 {
		t.Fatalf("expected at least 3 assistants, got %d: %v", len(found), found)
	}

	for i := 1; i < len(found); i++ {
		if found[i] < found[i-1] {
			t.Errorf("result not sorted: %v", found)
			break
		}
	}
}
