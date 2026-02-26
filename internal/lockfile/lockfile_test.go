package lockfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewCreatesLockFileWithDefaults(t *testing.T) {
	lf := New("v1.2.3")

	if lf.LockfileVersion != CurrentVersion {
		t.Errorf("LockfileVersion: got %d, want %d", lf.LockfileVersion, CurrentVersion)
	}
	if lf.CLIVersion != "v1.2.3" {
		t.Errorf("CLIVersion: got %q, want %q", lf.CLIVersion, "v1.2.3")
	}
	if lf.GeneratedAt == "" {
		t.Error("GeneratedAt should not be empty")
	}
	if lf.Packages == nil {
		t.Error("Packages map should be initialized")
	}
	if len(lf.Packages) != 0 {
		t.Errorf("Packages: got %d entries, want 0", len(lf.Packages))
	}
}

func TestDefaultPath(t *testing.T) {
	got := DefaultPath("/some/dir")
	want := filepath.Join("/some/dir", FileName)
	if got != want {
		t.Errorf("DefaultPath: got %q, want %q", got, want)
	}
}

func TestExistsReturnsFalseForMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.lock")
	exists, err := Exists(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected Exists to return false for missing file")
	}
}

func TestExistsReturnsTrueForExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, FileName)
	if err := os.WriteFile(path, []byte("lockfile_version: 1\n"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	exists, err := Exists(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected Exists to return true for existing file")
	}
}

func TestLoadReturnsNilForMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.lock")
	lf, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lf != nil {
		t.Error("expected Load to return nil for missing file")
	}
}

func TestLoadParsesValidLockfile(t *testing.T) {
	content := `lockfile_version: 1
generated_at: "2026-02-26T10:00:00Z"
cli_version: "v0.5.0"
packages:
  threat-modelling:
    type: builtin
    version: "v0.5.0"
    direct: true
  github.com/org/skills:
    type: remote
    source: "github.com/org/skills"
    version: "v1.0.0"
    resolved: "abc123def456"
    integrity: "sha256:deadbeef"
    direct: false
    dependency_of:
      - some-parent
`
	dir := t.TempDir()
	path := filepath.Join(dir, FileName)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test lockfile: %v", err)
	}

	lf, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if lf == nil {
		t.Fatal("expected non-nil lockfile")
	}

	if lf.LockfileVersion != 1 {
		t.Errorf("LockfileVersion: got %d, want 1", lf.LockfileVersion)
	}
	if lf.CLIVersion != "v0.5.0" {
		t.Errorf("CLIVersion: got %q, want %q", lf.CLIVersion, "v0.5.0")
	}
	if lf.GeneratedAt != "2026-02-26T10:00:00Z" {
		t.Errorf("GeneratedAt: got %q, want %q", lf.GeneratedAt, "2026-02-26T10:00:00Z")
	}
	if len(lf.Packages) != 2 {
		t.Fatalf("Packages: got %d, want 2", len(lf.Packages))
	}

	// Check builtin package
	builtin, ok := lf.Packages["threat-modelling"]
	if !ok {
		t.Fatal("missing threat-modelling package")
	}
	if builtin.Type != TypeBuiltin {
		t.Errorf("Type: got %q, want %q", builtin.Type, TypeBuiltin)
	}
	if !builtin.Direct {
		t.Error("expected Direct to be true")
	}

	// Check remote package
	remote, ok := lf.Packages["github.com/org/skills"]
	if !ok {
		t.Fatal("missing github.com/org/skills package")
	}
	if remote.Type != TypeRemote {
		t.Errorf("Type: got %q, want %q", remote.Type, TypeRemote)
	}
	if remote.Resolved != "abc123def456" {
		t.Errorf("Resolved: got %q, want %q", remote.Resolved, "abc123def456")
	}
	if remote.Integrity != "sha256:deadbeef" {
		t.Errorf("Integrity: got %q, want %q", remote.Integrity, "sha256:deadbeef")
	}
	if remote.Direct {
		t.Error("expected Direct to be false")
	}
	if len(remote.DependencyOf) != 1 || remote.DependencyOf[0] != "some-parent" {
		t.Errorf("DependencyOf: got %v, want [some-parent]", remote.DependencyOf)
	}
}

func TestLoadInitializesNilPackagesMap(t *testing.T) {
	content := `lockfile_version: 1
generated_at: "2026-01-01T00:00:00Z"
cli_version: "v0.1.0"
`
	dir := t.TempDir()
	path := filepath.Join(dir, FileName)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test lockfile: %v", err)
	}

	lf, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if lf.Packages == nil {
		t.Error("Packages map should be initialized even when absent from YAML")
	}
}

func TestLoadReturnsErrorForInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, FileName)
	if err := os.WriteFile(path, []byte("{{invalid yaml"), 0644); err != nil {
		t.Fatalf("failed to write test lockfile: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, FileName)

	original := New("v1.0.0")
	original.Packages["threat-modelling"] = LockedPackage{
		Type:    TypeBuiltin,
		Version: "v1.0.0",
		Direct:  true,
	}
	original.Packages["github.com/org/skills"] = LockedPackage{
		Type:         TypeRemote,
		Version:      "v2.0.0",
		Source:       "github.com/org/skills",
		Resolved:     "abc123def456789abc123def456789abc123def4",
		Integrity:    "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		Direct:       false,
		DependencyOf: []string{"parent-pkg"},
	}

	if err := Save(path, original); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.LockfileVersion != original.LockfileVersion {
		t.Errorf("LockfileVersion: got %d, want %d", loaded.LockfileVersion, original.LockfileVersion)
	}
	if loaded.CLIVersion != original.CLIVersion {
		t.Errorf("CLIVersion: got %q, want %q", loaded.CLIVersion, original.CLIVersion)
	}
	if loaded.GeneratedAt != original.GeneratedAt {
		t.Errorf("GeneratedAt: got %q, want %q", loaded.GeneratedAt, original.GeneratedAt)
	}
	if len(loaded.Packages) != len(original.Packages) {
		t.Fatalf("Packages: got %d, want %d", len(loaded.Packages), len(original.Packages))
	}

	// Verify builtin
	builtin := loaded.Packages["threat-modelling"]
	if builtin.Type != TypeBuiltin {
		t.Errorf("builtin Type: got %q, want %q", builtin.Type, TypeBuiltin)
	}
	if !builtin.Direct {
		t.Error("builtin should be direct")
	}

	// Verify remote
	remote := loaded.Packages["github.com/org/skills"]
	if remote.Type != TypeRemote {
		t.Errorf("remote Type: got %q, want %q", remote.Type, TypeRemote)
	}
	if remote.Resolved != original.Packages["github.com/org/skills"].Resolved {
		t.Errorf("remote Resolved mismatch")
	}
	if remote.Integrity != original.Packages["github.com/org/skills"].Integrity {
		t.Errorf("remote Integrity mismatch")
	}
	if remote.Direct {
		t.Error("remote should not be direct")
	}
	if len(remote.DependencyOf) != 1 || remote.DependencyOf[0] != "parent-pkg" {
		t.Errorf("remote DependencyOf: got %v, want [parent-pkg]", remote.DependencyOf)
	}
}

func TestSaveIncludesHeaderComment(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, FileName)

	lf := New("v1.0.0")
	if err := Save(path, lf); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read lockfile: %v", err)
	}

	content := string(data)
	if !strings.HasPrefix(content, "# code-minions.lock") {
		t.Errorf("expected header comment, got:\n%s", content[:min(len(content), 100)])
	}
	if !strings.Contains(content, "DO NOT EDIT MANUALLY") {
		t.Error("expected 'DO NOT EDIT MANUALLY' in header")
	}
}

func TestSaveOverwritesExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, FileName)

	// Write initial lockfile
	lf1 := New("v1.0.0")
	lf1.Packages["pkg-a"] = LockedPackage{Type: TypeBuiltin, Version: "v1.0.0", Direct: true}
	if err := Save(path, lf1); err != nil {
		t.Fatalf("first Save failed: %v", err)
	}

	// Overwrite with a different lockfile
	lf2 := New("v2.0.0")
	lf2.Packages["pkg-b"] = LockedPackage{Type: TypeBuiltin, Version: "v2.0.0", Direct: true}
	if err := Save(path, lf2); err != nil {
		t.Fatalf("second Save failed: %v", err)
	}

	// Verify the second write took effect
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.CLIVersion != "v2.0.0" {
		t.Errorf("CLIVersion: got %q, want %q", loaded.CLIVersion, "v2.0.0")
	}
	if _, ok := loaded.Packages["pkg-a"]; ok {
		t.Error("pkg-a should not be present after overwrite")
	}
	if _, ok := loaded.Packages["pkg-b"]; !ok {
		t.Error("pkg-b should be present after overwrite")
	}
}

func TestSaveDeterministicOutput(t *testing.T) {
	dir := t.TempDir()

	lf := New("v1.0.0")
	lf.GeneratedAt = "2026-01-01T00:00:00Z" // Fix timestamp for determinism
	lf.Packages["zzz-last"] = LockedPackage{Type: TypeBuiltin, Version: "v1.0.0", Direct: true}
	lf.Packages["aaa-first"] = LockedPackage{Type: TypeBuiltin, Version: "v1.0.0", Direct: true}
	lf.Packages["mmm-middle"] = LockedPackage{
		Type:         TypeRemote,
		Version:      "v1.0.0",
		Direct:       false,
		DependencyOf: []string{"zzz-last", "aaa-first"},
	}

	// Save twice and compare
	path1 := filepath.Join(dir, "lock1")
	path2 := filepath.Join(dir, "lock2")

	if err := Save(path1, lf); err != nil {
		t.Fatalf("first Save failed: %v", err)
	}
	if err := Save(path2, lf); err != nil {
		t.Fatalf("second Save failed: %v", err)
	}

	data1, _ := os.ReadFile(path1)
	data2, _ := os.ReadFile(path2)

	if string(data1) != string(data2) {
		t.Errorf("two saves of the same lockfile produced different output:\n--- save 1 ---\n%s\n--- save 2 ---\n%s", data1, data2)
	}

	// Verify packages are in sorted order
	content := string(data1)
	aIdx := strings.Index(content, "aaa-first")
	mIdx := strings.Index(content, "mmm-middle")
	zIdx := strings.Index(content, "zzz-last")
	if aIdx >= mIdx || mIdx >= zIdx {
		t.Errorf("packages not in sorted order: aaa=%d, mmm=%d, zzz=%d", aIdx, mIdx, zIdx)
	}

	// Verify dependency_of is also sorted
	depOfIdx1 := strings.Index(content, "- aaa-first")
	depOfIdx2 := strings.Index(content, "- zzz-last")
	if depOfIdx1 >= depOfIdx2 {
		t.Error("dependency_of entries not in sorted order")
	}
}

func TestMergeAddsAndUpdatesEntries(t *testing.T) {
	lf := New("v1.0.0")
	lf.Packages["existing"] = LockedPackage{Type: TypeBuiltin, Version: "v1.0.0", Direct: true}

	updates := map[string]LockedPackage{
		"new-pkg":  {Type: TypeRemote, Version: "v2.0.0", Direct: true},
		"existing": {Type: TypeBuiltin, Version: "v2.0.0", Direct: true}, // update
	}

	lf.Merge(updates)

	if len(lf.Packages) != 2 {
		t.Fatalf("Packages: got %d, want 2", len(lf.Packages))
	}
	if lf.Packages["existing"].Version != "v2.0.0" {
		t.Errorf("existing version: got %q, want %q", lf.Packages["existing"].Version, "v2.0.0")
	}
	if _, ok := lf.Packages["new-pkg"]; !ok {
		t.Error("new-pkg should be present after merge")
	}
}

func TestPruneRemovesUnlistedEntries(t *testing.T) {
	lf := New("v1.0.0")
	lf.Packages["keep-me"] = LockedPackage{Type: TypeBuiltin, Version: "v1.0.0", Direct: true}
	lf.Packages["remove-me"] = LockedPackage{Type: TypeBuiltin, Version: "v1.0.0", Direct: true}
	lf.Packages["also-keep"] = LockedPackage{Type: TypeRemote, Version: "v1.0.0", Direct: false}

	keep := map[string]bool{
		"keep-me":   true,
		"also-keep": true,
	}

	lf.Prune(keep)

	if len(lf.Packages) != 2 {
		t.Fatalf("Packages: got %d, want 2", len(lf.Packages))
	}
	if _, ok := lf.Packages["remove-me"]; ok {
		t.Error("remove-me should have been pruned")
	}
	if _, ok := lf.Packages["keep-me"]; !ok {
		t.Error("keep-me should still be present")
	}
	if _, ok := lf.Packages["also-keep"]; !ok {
		t.Error("also-keep should still be present")
	}
}

func TestPackageNamesReturnsSorted(t *testing.T) {
	lf := New("v1.0.0")
	lf.Packages["charlie"] = LockedPackage{}
	lf.Packages["alpha"] = LockedPackage{}
	lf.Packages["bravo"] = LockedPackage{}

	names := lf.PackageNames()
	if len(names) != 3 {
		t.Fatalf("got %d names, want 3", len(names))
	}
	if names[0] != "alpha" || names[1] != "bravo" || names[2] != "charlie" {
		t.Errorf("expected [alpha bravo charlie], got %v", names)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
