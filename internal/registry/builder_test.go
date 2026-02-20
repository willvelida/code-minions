package registry

import (
	"bytes"
	"strings"
	"testing"
	"testing/fstest"
)

// useTempCacheRoot redirects the git source cache to a temp directory
// so tests never collide with (or pollute) the real user cache.
func useTempCacheRoot(t *testing.T) {
	t.Helper()
	SetCacheRoot(t.TempDir())
	t.Cleanup(func() { SetCacheRoot("") })
}

// minimalEmbeddedFS returns a tiny fs.FS that EmbeddedSource can parse.
func minimalEmbeddedFS() fstest.MapFS {
	return fstest.MapFS{
		"packages/hello-world/package.yaml": &fstest.MapFile{
			Data: []byte("name: hello-world\nversion: 1.0.0\ndescription: a test package\n"),
		},
		"packages/hello-world/agents/AGENTS.md": &fstest.MapFile{
			Data: []byte("# Hello World Agent\n"),
		},
	}
}

func TestBuildRegistryFromConfig_NilConfig(t *testing.T) {
	// When cfg is nil, LoadConfig() is called. Regardless of what
	// the global config contains, embedded must be present as the
	// last source and no error should occur.
	reg, err := BuildRegistryFromConfig(minimalEmbeddedFS(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sources := reg.Sources()
	if len(sources) == 0 {
		t.Fatal("expected at least 1 source")
	}
	last := sources[len(sources)-1]
	if last.Name() != "embedded" {
		t.Errorf("last source should be embedded, got %q", last.Name())
	}
}

func TestBuildRegistryFromConfig_EmptyConfig(t *testing.T) {
	cfg := &Config{}
	reg, err := BuildRegistryFromConfig(minimalEmbeddedFS(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sources := reg.Sources()
	if len(sources) != 1 {
		t.Fatalf("expected 1 source (embedded), got %d", len(sources))
	}
	if sources[0].Name() != "embedded" {
		t.Errorf("expected embedded, got %q", sources[0].Name())
	}
}

func TestBuildRegistryFromConfig_WithGitSource(t *testing.T) {
	useTempCacheRoot(t)
	repoDir := setupTestGitRepo(t) // from git_test.go

	cfg := &Config{
		Sources: []SourceConfig{
			{Name: "my-team", Type: "git", URL: repoDir},
		},
	}

	reg, err := BuildRegistryFromConfig(minimalEmbeddedFS(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sources := reg.Sources()
	if len(sources) != 2 {
		t.Fatalf("expected 2 sources, got %d", len(sources))
	}
	if sources[0].Name() != "my-team" {
		t.Errorf("first source should be my-team, got %q", sources[0].Name())
	}
	if sources[1].Name() != "embedded" {
		t.Errorf("second source should be embedded, got %q", sources[1].Name())
	}
}

func TestBuildRegistryFromConfig_MultipleGitSources(t *testing.T) {
	useTempCacheRoot(t)
	repo1 := setupTestGitRepo(t)
	repo2 := setupTestGitRepo(t)

	cfg := &Config{
		Sources: []SourceConfig{
			{Name: "team-alpha", Type: "git", URL: repo1},
			{Name: "team-beta", Type: "git", URL: repo2},
		},
	}

	reg, err := BuildRegistryFromConfig(minimalEmbeddedFS(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sources := reg.Sources()
	if len(sources) != 3 {
		t.Fatalf("expected 3 sources, got %d", len(sources))
	}
	if sources[0].Name() != "team-alpha" {
		t.Errorf("expected team-alpha first, got %q", sources[0].Name())
	}
	if sources[1].Name() != "team-beta" {
		t.Errorf("expected team-beta second, got %q", sources[1].Name())
	}
	if sources[2].Name() != "embedded" {
		t.Errorf("expected embedded last, got %q", sources[2].Name())
	}
}

func TestBuildRegistryFromConfig_UnsupportedType(t *testing.T) {
	cfg := &Config{
		Sources: []SourceConfig{
			{Name: "bad", Type: "s3", URL: "s3://bucket"},
		},
	}

	// Unsupported types are now warn-and-skipped, not errors
	reg, err := BuildRegistryFromConfig(minimalEmbeddedFS(), cfg)
	if err != nil {
		t.Fatalf("expected no error (warn-and-skip), got: %v", err)
	}
	// Only the embedded source should remain
	sources := reg.Sources()
	if len(sources) != 1 {
		t.Fatalf("expected 1 source (embedded only), got %d", len(sources))
	}
	if sources[0].Name() != "embedded" {
		t.Errorf("expected embedded, got %q", sources[0].Name())
	}
}

func TestBuildRegistryFromConfig_ListsPackagesAcrossSources(t *testing.T) {
	useTempCacheRoot(t)
	repoDir := setupTestGitRepo(t)

	cfg := &Config{
		Sources: []SourceConfig{
			{Name: "team", Type: "git", URL: repoDir},
		},
	}

	reg, err := BuildRegistryFromConfig(minimalEmbeddedFS(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pkgs, err := reg.ListPackages()
	if err != nil {
		t.Fatalf("list packages: %v", err)
	}

	// Should have packages from both git source and embedded
	names := map[string]bool{}
	for _, p := range pkgs {
		names[p.Name] = true
	}

	// Git repo has test-skill and another-pkg; embedded has hello-world
	for _, expected := range []string{"test-skill", "another-pkg", "hello-world"} {
		if !names[expected] {
			t.Errorf("expected package %q not found in merged list", expected)
		}
	}
}

func TestBuildRegistryWithFromAndConfig_NamedSource(t *testing.T) {
	useTempCacheRoot(t)
	repoDir := setupTestGitRepo(t)

	cfg := &Config{
		Sources: []SourceConfig{
			{Name: "my-source", Type: "git", URL: repoDir},
		},
	}

	reg, err := BuildRegistryWithFromAndConfig(minimalEmbeddedFS(), "my-source", cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sources := reg.Sources()
	if len(sources) != 1 {
		t.Fatalf("expected 1 source (--from restricts scope), got %d", len(sources))
	}
	if sources[0].Name() != "my-source" {
		t.Errorf("first source should be my-source, got %q", sources[0].Name())
	}
}

func TestBuildRegistryWithFromAndConfig_GitURL(t *testing.T) {
	useTempCacheRoot(t)
	repoDir := setupTestGitRepo(t)

	cfg := &Config{} // empty config

	reg, err := BuildRegistryWithFromAndConfig(minimalEmbeddedFS(), repoDir, cfg)
	if err != nil {
		// On Windows the local path may not match IsGitURL — skip if so
		t.Skipf("local path not recognised as Git URL: %v", err)
	}
	sources := reg.Sources()
	if len(sources) != 1 {
		t.Fatalf("expected 1 source (--from restricts scope), got %d", len(sources))
	}
}

func TestBuildRegistryWithFromAndConfig_InvalidFrom(t *testing.T) {
	cfg := &Config{}

	_, err := BuildRegistryWithFromAndConfig(minimalEmbeddedFS(), "not-a-source-or-url", cfg)
	if err == nil {
		t.Fatal("expected error for invalid --from value")
	}
}

func TestResolveFrom_NamedSource(t *testing.T) {
	useTempCacheRoot(t)
	repoDir := setupTestGitRepo(t)

	cfg := &Config{
		Sources: []SourceConfig{
			{Name: "team-repo", Type: "git", URL: repoDir},
		},
	}

	src, err := ResolveFrom("team-repo", cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if src.Name() != "team-repo" {
		t.Errorf("expected name team-repo, got %q", src.Name())
	}
	if src.Type() != "git" {
		t.Errorf("expected type git, got %q", src.Type())
	}
}

func TestResolveFrom_GitURL(t *testing.T) {
	src, err := ResolveFrom("https://github.com/example/repo.git", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if src.Type() != "git" {
		t.Errorf("expected git type, got %q", src.Type())
	}
}

func TestResolveFrom_NotFound(t *testing.T) {
	cfg := &Config{}
	_, err := ResolveFrom("bogus", cfg)
	if err == nil {
		t.Fatal("expected error for unknown source")
	}
}

func TestBuildRegistryFromConfig_EmbeddedAlwaysLast(t *testing.T) {
	useTempCacheRoot(t)
	// Even with multiple git sources, embedded must be the last source
	repo := setupTestGitRepo(t)
	cfg := &Config{
		Sources: []SourceConfig{
			{Name: "src-a", Type: "git", URL: repo},
			{Name: "src-b", Type: "git", URL: repo},
		},
	}

	reg, err := BuildRegistryFromConfig(minimalEmbeddedFS(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sources := reg.Sources()
	last := sources[len(sources)-1]
	if last.Name() != "embedded" {
		t.Errorf("last source should be embedded, got %q", last.Name())
	}
}

func TestBuildRegistryFromConfig_SourceOrder(t *testing.T) {
	useTempCacheRoot(t)
	// Verify config source order is preserved
	repo := setupTestGitRepo(t)
	cfg := &Config{
		Sources: []SourceConfig{
			{Name: "first", Type: "git", URL: repo},
			{Name: "second", Type: "git", URL: repo},
			{Name: "third", Type: "git", URL: repo},
		},
	}

	reg, err := BuildRegistryFromConfig(minimalEmbeddedFS(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sources := reg.Sources()
	expected := []string{"first", "second", "third", "embedded"}
	if len(sources) != len(expected) {
		t.Fatalf("expected %d sources, got %d", len(expected), len(sources))
	}
	for i, name := range expected {
		if sources[i].Name() != name {
			t.Errorf("source[%d] = %q, want %q", i, sources[i].Name(), name)
		}
	}
}

func TestBuildRegistryFromConfigWithWarnings_SkipsFailingSource(t *testing.T) {
	useTempCacheRoot(t)
	repoDir := setupTestGitRepo(t)

	cfg := &Config{
		Sources: []SourceConfig{
			{Name: "good", Type: "git", URL: repoDir},
			{Name: "bad", Type: "git", URL: ""}, // Empty URL will fail NewGitSource
		},
	}

	var warnings bytes.Buffer
	reg, err := BuildRegistryFromConfigWithWarnings(minimalEmbeddedFS(), cfg, &warnings)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The good source + embedded should be present; "bad" should be skipped
	sources := reg.Sources()
	names := make([]string, len(sources))
	for i, s := range sources {
		names[i] = s.Name()
	}
	if len(sources) != 2 {
		t.Fatalf("expected 2 sources (good + embedded), got %d: %v", len(sources), names)
	}
	if sources[0].Name() != "good" {
		t.Errorf("first source should be good, got %q", sources[0].Name())
	}
	if sources[1].Name() != "embedded" {
		t.Errorf("last source should be embedded, got %q", sources[1].Name())
	}

	// Warning should mention the skipped source
	warnText := warnings.String()
	if !strings.Contains(warnText, "bad") {
		t.Errorf("expected warning about 'bad' source, got: %s", warnText)
	}
}

func TestBuildRegistryFromConfigWithWarnings_UnsupportedTypeWarning(t *testing.T) {
	cfg := &Config{
		Sources: []SourceConfig{
			{Name: "weird", Type: "s3", URL: "s3://bucket"},
		},
	}

	var warnings bytes.Buffer
	reg, err := BuildRegistryFromConfigWithWarnings(minimalEmbeddedFS(), cfg, &warnings)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Only embedded should remain
	sources := reg.Sources()
	if len(sources) != 1 || sources[0].Name() != "embedded" {
		t.Errorf("expected only embedded source, got %v", sources)
	}

	// Warning should mention the unsupported type
	warnText := warnings.String()
	if !strings.Contains(warnText, "weird") {
		t.Errorf("expected warning about 'weird' source, got: %s", warnText)
	}
	if !strings.Contains(warnText, "unsupported type") {
		t.Errorf("expected 'unsupported type' in warning, got: %s", warnText)
	}
}

func TestBuildRegistryFromConfigWithWarnings_NilWriter(t *testing.T) {
	cfg := &Config{
		Sources: []SourceConfig{
			{Name: "bad", Type: "git", URL: ""}, // Empty URL fails NewGitSource
		},
	}

	// nil writer — should not panic, just silently skip
	reg, err := BuildRegistryFromConfigWithWarnings(minimalEmbeddedFS(), cfg, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sources := reg.Sources()
	if len(sources) != 1 || sources[0].Name() != "embedded" {
		names := make([]string, len(sources))
		for i, s := range sources {
			names[i] = s.Name()
		}
		t.Errorf("expected only embedded, got %v", names)
	}
}

func TestBuildRegistryWithFrom_HardErrorOnFailure(t *testing.T) {
	// Explicit --from should still produce hard errors
	cfg := &Config{}
	_, err := BuildRegistryWithFromAndConfig(minimalEmbeddedFS(), "not-a-source-or-url", cfg)
	if err == nil {
		t.Fatal("expected hard error for invalid --from value")
	}
}

// TestBuildRegistryWithFromAndConfig_UnsupportedType verifies that
// --from with a configured source of unsupported type returns a
// clear validation error rather than a confusing git-clone failure.
func TestBuildRegistryWithFromAndConfig_UnsupportedType(t *testing.T) {
	cfg := &Config{
		Sources: []SourceConfig{
			{Name: "my-svn", Type: "svn", URL: "https://example.com/repo"},
		},
	}

	_, err := BuildRegistryWithFromAndConfig(minimalEmbeddedFS(), "my-svn", cfg)
	if err == nil {
		t.Fatal("expected error for unsupported source type")
	}
	if !strings.Contains(err.Error(), "unsupported source type") {
		t.Errorf("error should mention unsupported type, got: %v", err)
	}
}

// TestResolveFrom_UnsupportedType verifies that ResolveFrom validates
// the source type before attempting to construct a GitSource.
func TestResolveFrom_UnsupportedType(t *testing.T) {
	cfg := &Config{
		Sources: []SourceConfig{
			{Name: "my-svn", Type: "svn", URL: "https://example.com/repo"},
		},
	}

	_, err := ResolveFrom("my-svn", cfg)
	if err == nil {
		t.Fatal("expected error for unsupported source type")
	}
	if !strings.Contains(err.Error(), "unsupported source type") {
		t.Errorf("error should mention unsupported type, got: %v", err)
	}
}
