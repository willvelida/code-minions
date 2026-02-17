package installer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInjectTeamInstructions_NewFile(t *testing.T) {
	dir := t.TempDir()
	path, err := InjectTeamInstructions(dir, ".github/copilot-instructions.md", "platform-eng", "Follow coding standards.\nNo secrets.")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path == "" {
		t.Fatal("expected non-empty path")
	}

	data, err := os.ReadFile(filepath.Join(dir, ".github/copilot-instructions.md"))
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "<!-- code-minions:team:platform-eng:start -->") {
		t.Error("missing start marker")
	}
	if !strings.Contains(content, "<!-- code-minions:team:platform-eng:end -->") {
		t.Error("missing end marker")
	}
	if !strings.Contains(content, "Follow coding standards.") {
		t.Error("missing instructions content")
	}
	if !strings.HasSuffix(content, "\n") {
		t.Error("should end with newline")
	}
}

func TestInjectTeamInstructions_AppendToExisting(t *testing.T) {
	dir := t.TempDir()
	instrDir := filepath.Join(dir, ".github")
	if err := os.MkdirAll(instrDir, 0755); err != nil {
		t.Fatal(err)
	}

	existing := "# Project Instructions\n\nDo good work.\n"
	if err := os.WriteFile(filepath.Join(instrDir, "copilot-instructions.md"), []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := InjectTeamInstructions(dir, ".github/copilot-instructions.md", "my-team", "Team rule A.")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(instrDir, "copilot-instructions.md"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	// Original content preserved
	if !strings.Contains(content, "# Project Instructions") {
		t.Error("original content lost")
	}
	if !strings.Contains(content, "Do good work.") {
		t.Error("original content lost")
	}
	// Team block appended
	if !strings.Contains(content, "<!-- code-minions:team:my-team:start -->") {
		t.Error("missing start marker")
	}
	if !strings.Contains(content, "Team rule A.") {
		t.Error("missing team instructions")
	}
}

func TestInjectTeamInstructions_ReplaceExisting(t *testing.T) {
	dir := t.TempDir()

	// Inject first time
	_, err := InjectTeamInstructions(dir, "CLAUDE.md", "my-team", "Version 1")
	if err != nil {
		t.Fatal(err)
	}

	// Inject again — should replace
	_, err = InjectTeamInstructions(dir, "CLAUDE.md", "my-team", "Version 2")
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	if strings.Contains(content, "Version 1") {
		t.Error("old instructions should be replaced")
	}
	if !strings.Contains(content, "Version 2") {
		t.Error("new instructions should be present")
	}
	// Should have exactly one start marker
	if strings.Count(content, "<!-- code-minions:team:my-team:start -->") != 1 {
		t.Error("should have exactly one start marker")
	}
}

func TestInjectTeamInstructions_EmptyInstructions(t *testing.T) {
	dir := t.TempDir()
	path, err := InjectTeamInstructions(dir, "CLAUDE.md", "my-team", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "" {
		t.Errorf("expected empty path for empty instructions, got %q", path)
	}
}

func TestRemoveTeamInstructions_BlockPresent(t *testing.T) {
	dir := t.TempDir()

	// Set up a file with existing content + team block
	content := "# My Project\n\nExisting content.\n\n" +
		"<!-- code-minions:team:platform-eng:start -->\n" +
		"Team instructions here.\n" +
		"<!-- code-minions:team:platform-eng:end -->\n"
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	removed, err := RemoveTeamInstructions(dir, "CLAUDE.md", "platform-eng")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !removed {
		t.Fatal("expected removed=true")
	}

	data, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}
	result := string(data)

	if strings.Contains(result, "code-minions:team:platform-eng") {
		t.Error("markers should be removed")
	}
	if !strings.Contains(result, "# My Project") {
		t.Error("original content should be preserved")
	}
	if !strings.Contains(result, "Existing content.") {
		t.Error("original content should be preserved")
	}
}

func TestRemoveTeamInstructions_NotPresent(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# Hello\n"), 0644); err != nil {
		t.Fatal(err)
	}

	removed, err := RemoveTeamInstructions(dir, "CLAUDE.md", "nonexistent-team")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if removed {
		t.Error("expected removed=false for absent block")
	}
}

func TestRemoveTeamInstructions_FileNotExist(t *testing.T) {
	dir := t.TempDir()
	removed, err := RemoveTeamInstructions(dir, "CLAUDE.md", "my-team")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if removed {
		t.Error("expected removed=false for missing file")
	}
}

func TestRemoveTeamInstructions_OnlyTeamBlock_DeletesFile(t *testing.T) {
	dir := t.TempDir()

	// File contains only the team block
	content := "<!-- code-minions:team:my-team:start -->\nInstructions\n<!-- code-minions:team:my-team:end -->\n"
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	removed, err := RemoveTeamInstructions(dir, "CLAUDE.md", "my-team")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !removed {
		t.Fatal("expected removed=true")
	}

	// File should be deleted
	if _, err := os.Stat(filepath.Join(dir, "CLAUDE.md")); !os.IsNotExist(err) {
		t.Error("expected file to be deleted when only content was team block")
	}
}

func TestRemoveTeamInstructions_MultipleTeams(t *testing.T) {
	dir := t.TempDir()

	// Inject two teams
	_, err := InjectTeamInstructions(dir, "CLAUDE.md", "team-a", "Team A rules")
	if err != nil {
		t.Fatal(err)
	}
	_, err = InjectTeamInstructions(dir, "CLAUDE.md", "team-b", "Team B rules")
	if err != nil {
		t.Fatal(err)
	}

	// Remove team-a only
	removed, err := RemoveTeamInstructions(dir, "CLAUDE.md", "team-a")
	if err != nil {
		t.Fatal(err)
	}
	if !removed {
		t.Fatal("expected removed=true")
	}

	data, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	if strings.Contains(content, "team-a") {
		t.Error("team-a should be removed")
	}
	if !strings.Contains(content, "Team B rules") {
		t.Error("team-b should be preserved")
	}
}
