package manifest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestProjectManifestYAMLRoundTrip(t *testing.T) {
	input := `name: my-project
version: 1.0.0
assistant: copilot
packages:
    - threat-modelling
    - git-workflow
`

	var m ProjectManifest
	if err := yaml.Unmarshal([]byte(input), &m); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if m.Name != "my-project" {
		t.Errorf("Name: got %q, want %q", m.Name, "my-project")
	}
	if m.Version != "1.0.0" {
		t.Errorf("Version: got %q, want %q", m.Version, "1.0.0")
	}
	if m.Assistant != "copilot" {
		t.Errorf("Assistant: got %q, want %q", m.Assistant, "copilot")
	}
	if len(m.Packages) != 2 {
		t.Fatalf("Packages: got %d, want 2", len(m.Packages))
	}
	if m.Packages[0] != "threat-modelling" {
		t.Errorf("Packages[0]: got %q, want %q", m.Packages[0], "threat-modelling")
	}
	if m.Packages[1] != "git-workflow" {
		t.Errorf("Packages[1]: got %q, want %q", m.Packages[1], "git-workflow")
	}

	// Round trip: marshal back and verify it contains the expected fields
	out, err := yaml.Marshal(&m)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	outStr := string(out)
	for _, want := range []string{"name: my-project", "version: 1.0.0", "assistant: copilot", "threat-modelling", "git-workflow"} {
		if !strings.Contains(outStr, want) {
			t.Errorf("marshalled YAML missing %q:\n%s", want, outStr)
		}
	}
}

func TestProjectManifestOmitsEmptyFields(t *testing.T) {
	m := ProjectManifest{Name: "bare-project"}

	out, err := yaml.Marshal(&m)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	outStr := string(out)
	if strings.Contains(outStr, "version:") {
		t.Errorf("expected version to be omitted, got:\n%s", outStr)
	}
	if strings.Contains(outStr, "assistant:") {
		t.Errorf("expected assistant to be omitted, got:\n%s", outStr)
	}
	if strings.Contains(outStr, "packages:") {
		t.Errorf("expected packages to be omitted, got:\n%s", outStr)
	}
}

func TestDefaultPath(t *testing.T) {
	got := DefaultPath("/some/dir")
	want := filepath.Join("/some/dir", "code-minions.yml")
	if got != want {
		t.Errorf("DefaultPath: got %q, want %q", got, want)
	}
}

func TestExistsReturnsFalseForMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.yml")

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

	if err := os.WriteFile(path, []byte("name: test\n"), 0644); err != nil {
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

func TestLoadReturnZeroValueForMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.yml")

	m, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Name != "" {
		t.Errorf("expected empty Name, got %q", m.Name)
	}
	if len(m.Packages) != 0 {
		t.Errorf("expected empty Packages, got %v", m.Packages)
	}
}

func TestLoadParsesExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, FileName)

	content := []byte("name: loaded-project\nassistant: claude\npackages:\n  - git-workflow\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	m, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Name != "loaded-project" {
		t.Errorf("Name: got %q, want %q", m.Name, "loaded-project")
	}
	if m.Assistant != "claude" {
		t.Errorf("Assistant: got %q, want %q", m.Assistant, "claude")
	}
	if len(m.Packages) != 1 || m.Packages[0] != "git-workflow" {
		t.Errorf("Packages: got %v, want [git-workflow]", m.Packages)
	}
}

func TestLoadReturnsErrorForInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, FileName)

	if err := os.WriteFile(path, []byte(":\ninvalid: [yaml\n"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse manifest") {
		t.Errorf("error message should mention parsing, got: %v", err)
	}
}

func TestSaveCreatesFileWithHeaderComment(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, FileName)

	m := &ProjectManifest{
		Name:      "test-project",
		Version:   "0.1.0",
		Assistant: "copilot",
		Packages:  []string{"threat-modelling", "git-workflow"},
	}

	if err := Save(path, m); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}

	content := string(data)

	// Verify header comment is present and is the first line
	if !strings.HasPrefix(content, "# Generated by code-minions init\n") {
		t.Errorf("expected header comment at start, got:\n%s", content)
	}

	// Verify YAML content
	for _, want := range []string{"name: test-project", "version: 0.1.0", "assistant: copilot", "threat-modelling", "git-workflow"} {
		if !strings.Contains(content, want) {
			t.Errorf("saved file missing %q:\n%s", want, content)
		}
	}
}

func TestSaveCreatesParentDirectories(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "sub", "dir")
	path := filepath.Join(nested, FileName)

	m := &ProjectManifest{Name: "nested-project"}

	if err := Save(path, m); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify the file actually exists
	exists, err := Exists(path)
	if err != nil {
		t.Fatalf("Exists check failed: %v", err)
	}
	if !exists {
		t.Error("expected file to exist after Save with nested dirs")
	}

	// Verify round-trip through Load
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.Name != "nested-project" {
		t.Errorf("Name: got %q, want %q", loaded.Name, "nested-project")
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, FileName)

	original := &ProjectManifest{
		Name:      "round-trip",
		Version:   "2.0.0",
		Assistant: "claude",
		Packages:  []string{"developer-mentor", "raise-pull-requests"},
	}

	if err := Save(path, original); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Name != original.Name {
		t.Errorf("Name: got %q, want %q", loaded.Name, original.Name)
	}
	if loaded.Version != original.Version {
		t.Errorf("Version: got %q, want %q", loaded.Version, original.Version)
	}
	if loaded.Assistant != original.Assistant {
		t.Errorf("Assistant: got %q, want %q", loaded.Assistant, original.Assistant)
	}
	if len(loaded.Packages) != len(original.Packages) {
		t.Fatalf("Packages length: got %d, want %d", len(loaded.Packages), len(original.Packages))
	}
	for i, pkg := range original.Packages {
		if loaded.Packages[i] != pkg {
			t.Errorf("Packages[%d]: got %q, want %q", i, loaded.Packages[i], pkg)
		}
	}
}
