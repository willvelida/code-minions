package codeminions

import (
	"io/fs"
	"testing"
)

func TestEmbeddedContentContainsAgents(t *testing.T) {
	entries, err := fs.ReadDir(Content, "agents")
	if err != nil {
		t.Fatalf("failed to read agents directory: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("agents directory is empty")
	}

	// Check that AGENTS.md exists
	found := false
	for _, entry := range entries {
		if entry.Name() == "AGENTS.md" {
			found = true
			break
		}
	}

	if !found {
		t.Error("AGENTS.md not found in embedded agents directory")
	}
}

func TestEmbeddedContentContainsSkills(t *testing.T) {
	entries, err := fs.ReadDir(Content, "skills")
	if err != nil {
		t.Fatalf("failed to read skills directory: %v", err)
	}

	if len(entries) < 8 {
		t.Errorf("expected at least 8 skills, got %d", len(entries))
	}
}
