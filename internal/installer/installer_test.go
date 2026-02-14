package installer

import (
	"errors"
	"io/fs"
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

// ---------- Install error branch tests ----------

// TestInstallPathEscapeDetected verifies that a PathMapper returning a path
// outside the target directory is caught and reported as an error.
func TestInstallPathEscapeDetected(t *testing.T) {
	target := t.TempDir()

	inst := &Installer{
		Content:    testFS(),
		Target:     target,
		Force:      true,
		PathMapper: func(p string) string { return "../../" + p },
	}

	result, err := inst.Install([]string{"agents"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Errors) == 0 {
		t.Fatal("expected path escape errors, got none")
	}

	found := false
	for _, e := range result.Errors {
		if strings.Contains(e, "escapes target directory") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'escapes target directory' error, got: %v", result.Errors)
	}
}

// TestInstallReadFileFailure verifies that a ReadFile error from the content
// filesystem is captured in result.Errors without stopping the walk.
func TestInstallReadFileFailure(t *testing.T) {
	target := t.TempDir()

	content := &errFS{
		inner:        testFS(),
		readFileErrs: map[string]error{"agents/my-agent.agent.md": errors.New("disk read error")},
	}

	inst := &Installer{
		Content: content,
		Target:  target,
	}

	result, err := inst.Install([]string{"agents"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, e := range result.Errors {
		if strings.Contains(e, "failed to read") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'failed to read' error, got: %v", result.Errors)
	}
}

// TestInstallWriteFileFailure verifies that an os.WriteFile error is captured.
// Triggered by creating a directory where a file is expected to be written.
func TestInstallWriteFileFailure(t *testing.T) {
	target := t.TempDir()

	// Create a directory at the path where WriteFile expects to write a file.
	// os.WriteFile fails with "is a directory" on this path.
	conflictPath := filepath.Join(target, "agents", "my-agent.agent.md")
	if err := os.MkdirAll(conflictPath, 0755); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	inst := &Installer{
		Content: testFS(),
		Target:  target,
		Force:   true, // Skip "file exists" check — go straight to write
	}

	result, err := inst.Install([]string{"agents"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, e := range result.Errors {
		if strings.Contains(e, "failed to write") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'failed to write' error, got: %v", result.Errors)
	}
}

// TestInstallMkdirAllFailure verifies that an os.MkdirAll error for a directory
// is captured. Triggered by placing a file where a directory is expected.
func TestInstallMkdirAllFailure(t *testing.T) {
	target := t.TempDir()

	// Create a regular file named "agents" — MkdirAll fails because it can't
	// create a directory when a file already occupies that path.
	if err := os.WriteFile(filepath.Join(target, "agents"), []byte("blocker"), 0644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	inst := &Installer{
		Content: testFS(),
		Target:  target,
	}

	result, err := inst.Install([]string{"agents"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Errors) == 0 {
		t.Fatal("expected MkdirAll errors, got none")
	}

	found := false
	for _, e := range result.Errors {
		if strings.Contains(e, "failed to create directory") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'failed to create directory' error, got: %v", result.Errors)
	}
}

// TestInstallWalkDirCallbackError verifies that a ReadDir failure during
// WalkDir is captured in result.Errors via the callback's err parameter.
func TestInstallWalkDirCallbackError(t *testing.T) {
	target := t.TempDir()

	content := &errFS{
		inner:       testFS(),
		readDirErrs: map[string]error{"agents": errors.New("permission denied")},
	}

	inst := &Installer{
		Content: content,
		Target:  target,
	}

	result, err := inst.Install([]string{"agents"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, e := range result.Errors {
		if strings.Contains(e, "error reading") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'error reading' error, got: %v", result.Errors)
	}
}

// ---------- errFS: a test helper that injects errors into an fs.FS ----------

// errFS wraps an fs.FS and returns configured errors for specific paths.
// This lets us test error branches that fstest.MapFS can't trigger (e.g.
// ReadFile failures, ReadDir failures during WalkDir).
type errFS struct {
	inner        fs.FS
	readFileErrs map[string]error // ReadFile returns this error for matching paths
	readDirErrs  map[string]error // ReadDir returns this error for matching paths
}

// Open delegates to the inner filesystem. Required to implement fs.FS.
func (e *errFS) Open(name string) (fs.File, error) {
	return e.inner.Open(name)
}

// ReadFile returns a configured error or delegates to the inner filesystem.
// fs.ReadFile checks for this interface before falling back to Open+Read.
func (e *errFS) ReadFile(name string) ([]byte, error) {
	if e.readFileErrs != nil {
		if err, ok := e.readFileErrs[name]; ok {
			return nil, err
		}
	}
	return fs.ReadFile(e.inner, name)
}

// ReadDir returns a configured error or delegates to the inner filesystem.
// fs.WalkDir calls fs.ReadDir internally, so this triggers the WalkDir
// callback's err parameter when a subdirectory read fails.
func (e *errFS) ReadDir(name string) ([]fs.DirEntry, error) {
	if e.readDirErrs != nil {
		if err, ok := e.readDirErrs[name]; ok {
			return nil, err
		}
	}
	return fs.ReadDir(e.inner, name)
}

// ---------- test filesystem ----------

func testFS() fstest.MapFS {
	return fstest.MapFS{
		"agents/AGENTS.md":                  &fstest.MapFile{Data: []byte("# Agents")},
		"agents/my-agent.agent.md":          &fstest.MapFile{Data: []byte("# My Agent")},
		"skills/my-skill/SKILL.md":          &fstest.MapFile{Data: []byte("# My Skill")},
		"skills/my-skill/actions/create.md": &fstest.MapFile{Data: []byte("# Create")},
	}
}
