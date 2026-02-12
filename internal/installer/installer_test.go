package installer

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestInstallCopiesFiles(t *testing.T) {
	tests := []struct {
		name          string
		dirs          []string
		expectedCount int
	}{
		{"install agents", []string{"agents"}, 2},
		{"install skills", []string{"skills"}, 2},
		{"install all", []string{"agents", "skills"}, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := t.TempDir()

			inst := &Installer{
				Content: testFS(),
				Target:  target,
				Force:   false,
				DryRun:  false,
			}

			result, err := inst.Install(tt.dirs)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result.Copied) != tt.expectedCount {
				t.Errorf("copied count: got %d, want %d", len(result.Copied), tt.expectedCount)
			}

			// Verify files actually exist on disk
			for _, f := range result.Copied {
				path := filepath.Join(target, f)
				if _, err := os.Stat(path); os.IsNotExist(err) {
					t.Errorf("expected file to exist: %s", path)
				}
			}
		})
	}
}

func TestInstallSkipsExistingFiles(t *testing.T) {
	target := t.TempDir()

	// Pre-create a file that already exists
	agentsDir := filepath.Join(target, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("failed to create agents dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(agentsDir, "AGENTS.md"), []byte("original"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	inst := &Installer{
		Content: testFS(),
		Target:  target,
		Force:   false,
		DryRun:  false,
	}

	result, err := inst.Install([]string{"agents"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Skipped) != 1 {
		t.Errorf("skipped count: got %d, want 1", len(result.Skipped))
	}
	if len(result.Copied) != 1 {
		t.Errorf("copied count: got %d, want 1", len(result.Copied))
	}

	// Verify the original file was NOT overridden
	data, err := os.ReadFile(filepath.Join(agentsDir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(data) != "original" {
		t.Errorf("file was overwritten: got %q, want %q", string(data), "original")
	}
}

func TestInstallForceOverwrites(t *testing.T) {
	target := t.TempDir()

	// Pre-create a file
	agentsDir := filepath.Join(target, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("failed to create agents dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(agentsDir, "AGENTS.md"), []byte("original"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	inst := &Installer{
		Content: testFS(),
		Target:  target,
		Force:   true,
		DryRun:  false,
	}

	result, err := inst.Install([]string{"agents"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Copied) != 2 {
		t.Errorf("copied count: got %d, want 2", len(result.Copied))
	}
	if len(result.Skipped) != 0 {
		t.Errorf("skipped count: got %d, want 0", len(result.Skipped))
	}

	// Verify the file WAS overwritten
	data, err := os.ReadFile(filepath.Join(agentsDir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(data) != "# Agents" {
		t.Errorf("file was not overwritten: got %q, want %q", string(data), "# Agents")
	}
}

func TestInstallDryRunDoesNotWrite(t *testing.T) {
	target := t.TempDir()

	inst := &Installer{
		Content: testFS(),
		Target:  target,
		Force:   false,
		DryRun:  true,
	}

	result, err := inst.Install([]string{"agents", "skills"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Copied) != 4 {
		t.Errorf("copied count: got %d, want 4", len(result.Copied))
	}

	// Verify nothing was actually written to disk
	entries, err := os.ReadDir(target)
	if err != nil {
		t.Fatalf("failed to read target dir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty target dir, got %d entries", len(entries))
	}
}

func TestInstallInvalidDirReturnsError(t *testing.T) {
	target := t.TempDir()

	inst := &Installer{
		Content: testFS(),
		Target:  target,
		Force:   false,
		DryRun:  false,
	}

	// Install returns nil error because the walk callback captures errors
	// into result.Errors rather than propagating them (soft error approach).
	result, err := inst.Install([]string{"nonexistent"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Errors) == 0 {
		t.Fatal("expected errors in result for nonexistent directory")
	}
}

func testFS() fstest.MapFS {
	return fstest.MapFS{
		"agents/AGENTS.md":                  &fstest.MapFile{Data: []byte("# Agents")},
		"agents/my-agent.agent.md":          &fstest.MapFile{Data: []byte("# My Agent")},
		"skills/my-skill/SKILL.md":          &fstest.MapFile{Data: []byte("# My Skill")},
		"skills/my-skill/actions/create.md": &fstest.MapFile{Data: []byte("# Create")},
	}
}
