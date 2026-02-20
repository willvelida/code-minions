package registry

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// setupTestGitRepo creates a local Git repo with the code-minions
// packages/ directory layout. Returns the repo path.
func setupTestGitRepo(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	// Create packages directory structure
	pkgDir := filepath.Join(dir, "packages", "test-skill")
	if err := os.MkdirAll(filepath.Join(pkgDir, "agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(pkgDir, "skills", "test-skill"), 0o755); err != nil {
		t.Fatal(err)
	}

	// Write package.yaml
	packageYAML := `name: test-skill
version: 1.0.0
description: A test skill for testing
author: test-author
contents:
  agents:
    - agents/test-skill.agent.md
  skills:
    - skills/test-skill/SKILL.md
`
	if err := os.WriteFile(filepath.Join(pkgDir, "package.yaml"), []byte(packageYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	// Write agent file
	if err := os.WriteFile(filepath.Join(pkgDir, "agents", "test-skill.agent.md"), []byte("# Test Agent"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Write skill file
	skillContent := "---\nname: test-skill\ndescription: A test skill\n---\n# Test Skill"
	if err := os.WriteFile(filepath.Join(pkgDir, "skills", "test-skill", "SKILL.md"), []byte(skillContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create a second package
	pkg2Dir := filepath.Join(dir, "packages", "another-pkg")
	if err := os.MkdirAll(filepath.Join(pkg2Dir, "agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(pkg2Dir, "skills", "another-pkg"), 0o755); err != nil {
		t.Fatal(err)
	}

	pkg2YAML := `name: another-pkg
version: 0.2.0
description: Another package for testing
`
	if err := os.WriteFile(filepath.Join(pkg2Dir, "package.yaml"), []byte(pkg2YAML), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkg2Dir, "agents", "another-pkg.agent.md"), []byte("# Another Agent"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkg2Dir, "skills", "another-pkg", "SKILL.md"), []byte("# Another Skill"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Git init and commit
	runGit(t, dir, "init")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "initial commit")

	return dir
}

// setupTestGitRepoNoPackages creates a Git repo without a packages/ directory.
func setupTestGitRepoNoPackages(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# No packages"), 0o644); err != nil {
		t.Fatal(err)
	}

	runGit(t, dir, "init")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "initial")

	return dir
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test",
		"GIT_AUTHOR_EMAIL=test@test.com",
		"GIT_COMMITTER_NAME=test",
		"GIT_COMMITTER_EMAIL=test@test.com",
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %s\n%s", args, err, output)
	}
}

func TestGitSourceNameAndType(t *testing.T) {
	src := NewGitSourceWithDir("my-team", "https://example.com/repo.git", t.TempDir())

	if src.Name() != "my-team" {
		t.Errorf("Name: got %q, want %q", src.Name(), "my-team")
	}
	if src.Type() != "git" {
		t.Errorf("Type: got %q, want %q", src.Type(), "git")
	}
}

func TestGitSourceImplementsSource(t *testing.T) {
	// Compile-time check that GitSource implements Source
	var _ Source = (*GitSource)(nil)
}

func TestGitSourceWithLocalRepo(t *testing.T) {
	repoDir := setupTestGitRepo(t)

	// Use the repo directory directly (no clone needed)
	src := NewGitSourceWithDir("test-source", repoDir, repoDir)

	// Manually initialise the embedded source
	src.embedded = NewEmbeddedSource(os.DirFS(repoDir))

	// Test ListPackages
	pkgs, err := src.ListPackages()
	if err != nil {
		t.Fatalf("ListPackages: %v", err)
	}

	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}

	// Find test-skill
	var found bool
	for _, pkg := range pkgs {
		if pkg.Name == "test-skill" {
			found = true
			if pkg.Version != "1.0.0" {
				t.Errorf("version: got %q, want %q", pkg.Version, "1.0.0")
			}
			if pkg.Description != "A test skill for testing" {
				t.Errorf("description: got %q", pkg.Description)
			}
		}
	}
	if !found {
		t.Error("test-skill package not found")
	}
}

func TestGitSourceGetPackage(t *testing.T) {
	repoDir := setupTestGitRepo(t)
	src := NewGitSourceWithDir("test-source", repoDir, repoDir)
	src.embedded = NewEmbeddedSource(os.DirFS(repoDir))

	pkg, err := src.GetPackage("test-skill")
	if err != nil {
		t.Fatalf("GetPackage: %v", err)
	}
	if pkg.Name != "test-skill" {
		t.Errorf("Name: got %q", pkg.Name)
	}
	if pkg.Version != "1.0.0" {
		t.Errorf("Version: got %q", pkg.Version)
	}
}

func TestGitSourceGetPackageNotFound(t *testing.T) {
	repoDir := setupTestGitRepo(t)
	src := NewGitSourceWithDir("test-source", repoDir, repoDir)
	src.embedded = NewEmbeddedSource(os.DirFS(repoDir))

	_, err := src.GetPackage("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent package")
	}
}

func TestGitSourceDownloadPackage(t *testing.T) {
	repoDir := setupTestGitRepo(t)
	src := NewGitSourceWithDir("test-source", repoDir, repoDir)
	src.embedded = NewEmbeddedSource(os.DirFS(repoDir))

	pkgFS, err := src.DownloadPackage("test-skill", "1.0.0")
	if err != nil {
		t.Fatalf("DownloadPackage: %v", err)
	}

	// The returned FS should be rooted at the package directory
	// and contain agents/ and skills/
	agentData, err := fs.ReadFile(pkgFS, "agents/test-skill.agent.md")
	if err != nil {
		t.Fatalf("reading agent from downloaded FS: %v", err)
	}
	if string(agentData) != "# Test Agent" {
		t.Errorf("agent content: got %q", string(agentData))
	}

	skillData, err := fs.ReadFile(pkgFS, "skills/test-skill/SKILL.md")
	if err != nil {
		t.Fatalf("reading skill from downloaded FS: %v", err)
	}
	if len(skillData) == 0 {
		t.Error("skill file is empty")
	}
}

func TestGitSourceSearch(t *testing.T) {
	repoDir := setupTestGitRepo(t)
	src := NewGitSourceWithDir("my-team", repoDir, repoDir)
	src.embedded = NewEmbeddedSource(os.DirFS(repoDir))

	results, err := src.Search("test")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("expected at least one search result")
	}

	// Results should have the git source name, not "embedded"
	for _, r := range results {
		if r.Source != "my-team" {
			t.Errorf("search result source: got %q, want %q", r.Source, "my-team")
		}
	}
}

func TestGitSourceEnsureWithClone(t *testing.T) {
	repoDir := setupTestGitRepo(t)

	// Create a target cache directory that doesn't exist yet
	cacheDir := filepath.Join(t.TempDir(), "cache", "test-source")

	src := &GitSource{
		name:     "test-source",
		url:      repoDir, // Clone from local repo
		cacheDir: cacheDir,
	}

	// Ensure should clone the repo
	if err := src.Ensure(); err != nil {
		t.Fatalf("Ensure (clone): %v", err)
	}

	// Should now be able to list packages
	pkgs, err := src.ListPackages()
	if err != nil {
		t.Fatalf("ListPackages after clone: %v", err)
	}
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages after clone, got %d", len(pkgs))
	}
}

func TestGitSourceEnsureIdempotent(t *testing.T) {
	repoDir := setupTestGitRepo(t)
	cacheDir := filepath.Join(t.TempDir(), "cache", "test-source")

	src := &GitSource{
		name:     "test-source",
		url:      repoDir,
		cacheDir: cacheDir,
	}

	// First Ensure — clone
	if err := src.Ensure(); err != nil {
		t.Fatalf("first Ensure: %v", err)
	}

	// Second Ensure — should be a no-op (already initialised)
	if err := src.Ensure(); err != nil {
		t.Fatalf("second Ensure: %v", err)
	}

	pkgs, err := src.ListPackages()
	if err != nil {
		t.Fatalf("ListPackages: %v", err)
	}
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
}

func TestGitSourceEnsureNoPackagesDir(t *testing.T) {
	repoDir := setupTestGitRepoNoPackages(t)
	cacheDir := filepath.Join(t.TempDir(), "cache", "test-source")

	src := &GitSource{
		name:     "test-source",
		url:      repoDir,
		cacheDir: cacheDir,
	}

	err := src.Ensure()
	if err == nil {
		t.Fatal("expected error when repo has no packages/ directory")
	}
	if !strings.Contains(err.Error(), "does not contain a packages/ directory") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestGitSourceEnsurePullAfterClone(t *testing.T) {
	repoDir := setupTestGitRepo(t)
	cacheDir := filepath.Join(t.TempDir(), "cache", "test-source")

	// First clone
	src := &GitSource{
		name:     "test-source",
		url:      repoDir,
		cacheDir: cacheDir,
	}
	if err := src.Ensure(); err != nil {
		t.Fatalf("first Ensure: %v", err)
	}

	// Reset embedded so Ensure tries again
	src.embedded = nil

	// Should pull (no new commits, but should succeed)
	if err := src.Ensure(); err != nil {
		t.Fatalf("second Ensure (pull): %v", err)
	}
}

func TestNewGitSource(t *testing.T) {
	src, err := NewGitSource(SourceConfig{
		Name: "my-team",
		Type: "git",
		URL:  "https://github.com/org/repo.git",
	})
	if err != nil {
		t.Fatalf("NewGitSource: %v", err)
	}

	if src.Name() != "my-team" {
		t.Errorf("Name: got %q", src.Name())
	}
	if src.url != "https://github.com/org/repo.git" {
		t.Errorf("URL: got %q", src.url)
	}
	if src.cacheDir == "" {
		t.Error("cacheDir should not be empty")
	}
}

func TestNewGitSourceEmptyURL(t *testing.T) {
	_, err := NewGitSource(SourceConfig{
		Name: "my-source",
		Type: "git",
		URL:  "",
	})
	if err == nil {
		t.Fatal("expected error for empty URL")
	}
}

func TestNewGitSourceFromURL(t *testing.T) {
	src, err := NewGitSourceFromURL("https://github.com/example/skills.git")
	if err != nil {
		t.Fatalf("NewGitSourceFromURL: %v", err)
	}

	if src.Type() != "git" {
		t.Errorf("Type: got %q", src.Type())
	}
	// Name should be auto-generated
	if src.Name() == "" {
		t.Error("Name should not be empty")
	}
}

func TestNewGitSourceFromURLEmpty(t *testing.T) {
	_, err := NewGitSourceFromURL("")
	if err == nil {
		t.Fatal("expected error for empty URL")
	}
}

// TestNewGitSourceFromURL_NormalisesBareDomain verifies that bare
// domain/path URLs (e.g. github.com/org/repo) are normalised to
// https:// so that git clone will work reliably.
func TestNewGitSourceFromURL_NormalisesBareDomain(t *testing.T) {
	src, err := NewGitSourceFromURL("github.com/example/skills")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(src.url, "https://") {
		t.Errorf("expected url to have https:// prefix, got %q", src.url)
	}
}

// TestNewGitSourceFromURL_PreservesHTTPS verifies that URLs that
// already have a scheme are left unchanged.
func TestNewGitSourceFromURL_PreservesHTTPS(t *testing.T) {
	url := "https://github.com/example/skills.git"
	src, err := NewGitSourceFromURL(url)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if src.url != url {
		t.Errorf("url should be unchanged: got %q, want %q", src.url, url)
	}
}

// TestNewGitSourceFromURL_PreservesSSH verifies SSH URLs are unchanged.
func TestNewGitSourceFromURL_PreservesSSH(t *testing.T) {
	url := "git@github.com:example/skills.git"
	src, err := NewGitSourceFromURL(url)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if src.url != url {
		t.Errorf("url should be unchanged: got %q, want %q", src.url, url)
	}
}

func TestIsGitURL(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"https://github.com/org/repo.git", true},
		{"http://github.com/org/repo.git", true},
		{"git@github.com:org/repo.git", true},
		{"ssh://git@github.com/org/repo.git", true},
		{"github.com/org/repo", true},
		{"my-team", false},
		{"embedded", false},
		{"just-a-name", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := IsGitURL(tt.input)
			if got != tt.want {
				t.Errorf("IsGitURL(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestCheckGitAvailable(t *testing.T) {
	// Git should be available in CI and development environments
	err := checkGitAvailable()
	if err != nil {
		t.Skipf("git not available: %v", err)
	}
}

func TestGitSourceListPersonasEmpty(t *testing.T) {
	repoDir := setupTestGitRepo(t)
	src := NewGitSourceWithDir("test-source", repoDir, repoDir)
	src.embedded = NewEmbeddedSource(os.DirFS(repoDir))

	personas, err := src.ListPersonas()
	if err != nil {
		t.Fatalf("ListPersonas: %v", err)
	}
	// The test repo has no personas/ directory, should return empty
	if len(personas) != 0 {
		t.Errorf("expected 0 personas, got %d", len(personas))
	}
}

func TestGitSourceListTeamsEmpty(t *testing.T) {
	repoDir := setupTestGitRepo(t)
	src := NewGitSourceWithDir("test-source", repoDir, repoDir)
	src.embedded = NewEmbeddedSource(os.DirFS(repoDir))

	teams, err := src.ListTeams()
	if err != nil {
		t.Fatalf("ListTeams: %v", err)
	}
	if len(teams) != 0 {
		t.Errorf("expected 0 teams, got %d", len(teams))
	}
}
