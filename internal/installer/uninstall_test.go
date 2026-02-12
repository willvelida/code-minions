package installer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestUninstallRemovesFiles(t *testing.T) {
	target := t.TempDir()

	// First, install the files
	inst := &Installer{
		Content: testFS(),
		Target:  target,
		Force:   false,
		DryRun:  false,
	}

	_, err := inst.Install([]string{"agents", "skills"})
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Now uninstall
	result, err := inst.Uninstall([]string{"agents", "skills"})
	if err != nil {
		t.Fatalf("uninstall failed: %v", err)
	}

	if len(result.Removed) != 4 {
		t.Errorf("removed count: got %d, want 4\n  removed: %v", len(result.Removed), result.Removed)
	}

	// Verify files no longer exist on disk
	for _, f := range result.Removed {
		path := filepath.Join(target, f)
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("expected file to be removed: %s", path)
		}
	}
}

func TestUninstallNotFoundFiles(t *testing.T) {
	target := t.TempDir()

	// Uninstall without installing first
	inst := &Installer{
		Content: testFS(),
		Target:  target,
		Force:   false,
		DryRun:  false,
	}

	result, err := inst.Uninstall([]string{"agents", "skills"})
	if err != nil {
		t.Fatalf("uninstall failed: %v", err)
	}

	if len(result.NotFound) != 4 {
		t.Errorf("not found count: got %d, want 4\n  not found: %v", len(result.NotFound), result.NotFound)
	}
	if len(result.Removed) != 0 {
		t.Errorf("removed count: got %d, want 0", len(result.Removed))
	}
}

func TestUninstallDryRunDoesNotDelete(t *testing.T) {
	target := t.TempDir()

	inst := &Installer{
		Content: testFS(),
		Target:  target,
		Force:   false,
		DryRun:  false,
	}

	// Install files
	_, err := inst.Install([]string{"agents", "skills"})
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Uninstall in dry-run mode
	dryInst := &Installer{
		Content: testFS(),
		Target:  target,
		Force:   false,
		DryRun:  true,
	}

	result, err := dryInst.Uninstall([]string{"agents", "skills"})
	if err != nil {
		t.Fatalf("uninstall failed: %v", err)
	}

	if len(result.Removed) != 4 {
		t.Errorf("removed count: got %d, want 4 (dry-run should report)", len(result.Removed))
	}

	// Verify files still exist on disk
	for _, f := range result.Removed {
		path := filepath.Join(target, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("file should still exist in dry-run: %s", path)
		}
	}
}

func TestUninstallCleansEmptyDirs(t *testing.T) {
	target := t.TempDir()

	inst := &Installer{
		Content: testFS(),
		Target:  target,
		Force:   false,
		DryRun:  false,
	}

	// Install, then uninstall
	_, err := inst.Install([]string{"agents", "skills"})
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}

	result, err := inst.Uninstall([]string{"agents", "skills"})
	if err != nil {
		t.Fatalf("uninstall failed: %v", err)
	}

	if len(result.DirsCleaned) == 0 {
		t.Error("expected at least one directory to be cleaned up")
	}

	// Verify the deepest directory (skills/my-skill/actions/) was removed
	actionsDir := filepath.Join(target, "skills", "my-skill", "actions")
	if _, err := os.Stat(actionsDir); !os.IsNotExist(err) {
		t.Errorf("expected empty dir to be removed: %s", actionsDir)
	}
}

func TestUninstallPreservesNonEmptyDirs(t *testing.T) {
	target := t.TempDir()

	// Create a filesystem with two "packages" sharing a parent dir
	content := fstest.MapFS{
		"skills/pkg-a/SKILL.md": &fstest.MapFile{Data: []byte("# A")},
		"skills/pkg-b/SKILL.md": &fstest.MapFile{Data: []byte("# B")},
	}

	inst := &Installer{
		Content: content,
		Target:  target,
		Force:   false,
		DryRun:  false,
	}

	// Install both
	_, err := inst.Install([]string{"skills"})
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Uninstall only pkg-a (by creating a content FS with only pkg-a)
	contentA := fstest.MapFS{
		"skills/pkg-a/SKILL.md": &fstest.MapFile{Data: []byte("# A")},
	}
	instA := &Installer{
		Content: contentA,
		Target:  target,
		Force:   false,
		DryRun:  false,
	}

	result, err := instA.Uninstall([]string{"skills"})
	if err != nil {
		t.Fatalf("uninstall failed: %v", err)
	}

	if len(result.Removed) != 1 {
		t.Errorf("removed count: got %d, want 1", len(result.Removed))
	}

	// skills/ should still exist because pkg-b is still there
	skillsDir := filepath.Join(target, "skills")
	if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
		t.Error("skills/ dir should NOT be removed — it still contains pkg-b")
	}

	// pkg-b should still exist
	pkgBFile := filepath.Join(target, "skills", "pkg-b", "SKILL.md")
	if _, err := os.Stat(pkgBFile); os.IsNotExist(err) {
		t.Error("pkg-b/SKILL.md should still exist")
	}
}

func TestUninstallWithStripPrefix(t *testing.T) {
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

	// Install first
	_, err := inst.Install([]string{"packages/my-pkg"})
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Uninstall
	result, err := inst.Uninstall([]string{"packages/my-pkg"})
	if err != nil {
		t.Fatalf("uninstall failed: %v", err)
	}

	if len(result.Removed) != 2 {
		t.Errorf("removed count: got %d, want 2\n  removed: %v", len(result.Removed), result.Removed)
	}

	// Files should be gone from the stripped paths (not packages/my-pkg/...)
	agentPath := filepath.Join(target, "agents", "agent.md")
	if _, err := os.Stat(agentPath); !os.IsNotExist(err) {
		t.Errorf("expected file to be removed: %s", agentPath)
	}
}

func TestUninstallWithPathMapper(t *testing.T) {
	target := t.TempDir()

	content := fstest.MapFS{
		"agents/my-agent.agent.md": &fstest.MapFile{Data: []byte("# Agent")},
	}

	// Mapper that remaps agents/ → .github/agents/ (like Copilot)
	mapper := func(p string) string {
		if strings.HasPrefix(p, "agents/") || p == "agents" {
			return strings.Replace(p, "agents", ".github/agents", 1)
		}
		return p
	}

	inst := &Installer{
		Content:    content,
		Target:     target,
		Force:      false,
		DryRun:     false,
		PathMapper: mapper,
	}

	// Install
	_, err := inst.Install([]string{"agents"})
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Verify it's at the mapped path
	mappedPath := filepath.Join(target, ".github", "agents", "my-agent.agent.md")
	if _, err := os.Stat(mappedPath); os.IsNotExist(err) {
		t.Fatalf("expected file at mapped path: %s", mappedPath)
	}

	// Uninstall
	result, err := inst.Uninstall([]string{"agents"})
	if err != nil {
		t.Fatalf("uninstall failed: %v", err)
	}

	if len(result.Removed) != 1 {
		t.Errorf("removed count: got %d, want 1", len(result.Removed))
	}

	// File should be gone from the mapped path
	if _, err := os.Stat(mappedPath); !os.IsNotExist(err) {
		t.Errorf("expected file to be removed: %s", mappedPath)
	}
}
