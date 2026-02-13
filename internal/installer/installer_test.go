package installer

import (
	"os"
	"path/filepath"
	"strings"
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

			// Verify directories were created on disk
			for _, d := range tt.dirs {
				dirPath := filepath.Join(target, d)
				info, err := os.Stat(dirPath)
				if os.IsNotExist(err) {
					t.Errorf("expected directory to exist: %s", dirPath)
				} else if err != nil {
					t.Errorf("unexpected error checking directory: %v", err)
				} else if !info.IsDir() {
					t.Errorf("expected %s to be a directory", dirPath)
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

func TestInstallStripPrefixRemovesPackageDir(t *testing.T) {
	target := t.TempDir()

	content := fstest.MapFS{
		"packages/my-pkg/agents/agent.md":        &fstest.MapFile{Data: []byte("# Agent")},
		"packages/my-pkg/skills/my-pkg/SKILL.md": &fstest.MapFile{Data: []byte("# Skill")},
	}

	inst := &Installer{
		Content:     content,
		Target:      target,
		Force:       false,
		DryRun:      false,
		StripPrefix: "packages/my-pkg",
	}

	result, err := inst.Install([]string{"packages/my-pkg"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Copied) != 2 {
		t.Errorf("copied count: got %d, want 2", len(result.Copied))
	}

	// The package wrapper directory should NOT exist in the output
	pkgDir := filepath.Join(target, "packages")
	if _, err := os.Stat(pkgDir); !os.IsNotExist(err) {
		t.Errorf("packages/ directory should not exist in target, but it does")
	}

	// The stripped paths SHOULD exist
	agentPath := filepath.Join(target, "agents", "agent.md")
	if _, err := os.Stat(agentPath); os.IsNotExist(err) {
		t.Errorf("expected file to exist: %s", agentPath)
	}

	skillPath := filepath.Join(target, "skills", "my-pkg", "SKILL.md")
	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		t.Errorf("expected file to exist: %s", skillPath)
	}
}

func TestInstallNilPathMapperPreservesBehaviour(t *testing.T) {
	target := t.TempDir()

	inst := &Installer{
		Content:    testFS(),
		Target:     target,
		Force:      false,
		DryRun:     false,
		PathMapper: nil, // Explicitly nil — should behave exactly like before
	}

	result, err := inst.Install([]string{"agents"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Copied) != 2 {
		t.Errorf("copied count: got %d, want 2", len(result.Copied))
	}

	// Files should land in their original paths (agents/)
	agentPath := filepath.Join(target, "agents", "my-agent.agent.md")
	if _, err := os.Stat(agentPath); os.IsNotExist(err) {
		t.Errorf("expected file to exist at original path: %s", agentPath)
	}
}

func TestInstallPathMapperRemapsPaths(t *testing.T) {
	target := t.TempDir()

	// A PathMapper that moves agents/ to .github/agents/
	// This simulates what the Copilot assistant config would do
	mapper := func(path string) string {
		if strings.HasPrefix(path, "agents/") {
			return ".github/agents/" + strings.TrimPrefix(path, "agents/")
		}
		return path
	}

	inst := &Installer{
		Content:    testFS(),
		Target:     target,
		Force:      false,
		DryRun:     false,
		PathMapper: mapper,
	}

	result, err := inst.Install([]string{"agents"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Copied) != 2 {
		t.Errorf("copied count: got %d, want 2", len(result.Copied))
	}

	// Files should land in the REMAPPED path (.github/agents/), not agents/
	remappedPath := filepath.Join(target, ".github", "agents", "my-agent.agent.md")
	if _, err := os.Stat(remappedPath); os.IsNotExist(err) {
		t.Errorf("expected file at remapped path: %s", remappedPath)
	}

	// The original agents/ path should NOT exist
	originalPath := filepath.Join(target, "agents", "my-agent.agent.md")
	if _, err := os.Stat(originalPath); !os.IsNotExist(err) {
		t.Errorf("file should NOT exist at original path: %s", originalPath)
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
