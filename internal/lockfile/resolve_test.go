package lockfile

import (
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/willvelida/code-minions/internal/model"
	"github.com/willvelida/code-minions/internal/registry"
)

// --- Mock source types for testing ---

// mockSource implements registry.Source for testing. It returns
// a fixed fs.FS from DownloadPackage.
type mockSource struct {
	name    string
	typ     string
	content fs.FS
}

func (m *mockSource) Name() string                           { return m.name }
func (m *mockSource) Type() string                           { return m.typ }
func (m *mockSource) ListPackages() ([]model.Package, error) { return nil, nil }
func (m *mockSource) GetPackage(_ string) (*model.Package, error) {
	return nil, nil
}
func (m *mockSource) DownloadPackage(_, _ string) (fs.FS, error) {
	return m.content, nil
}
func (m *mockSource) ListPersonas() ([]model.Persona, error)        { return nil, nil }
func (m *mockSource) GetPersona(_ string) (*model.Persona, error)   { return nil, nil }
func (m *mockSource) ListTeams() ([]model.Team, error)              { return nil, nil }
func (m *mockSource) GetTeam(_ string) (*model.Team, error)         { return nil, nil }
func (m *mockSource) Search(_ string) ([]model.SearchResult, error) { return nil, nil }

// mockPinnableSource is a git-like source that satisfies the pinnable
// interface used by FromResolvedTree.
type mockPinnableSource struct {
	mockSource
	sha string
	url string
}

func (m *mockPinnableSource) HeadSHA() string { return m.sha }
func (m *mockPinnableSource) URL() string     { return m.url }

// --- Helpers ---

func makeEmbeddedSource() *mockSource {
	return &mockSource{
		name: "embedded",
		typ:  "embedded",
	}
}

func makeGitSource(url, sha string, content fs.FS) *mockPinnableSource {
	return &mockPinnableSource{
		mockSource: mockSource{
			name:    "test-git",
			typ:     "git",
			content: content,
		},
		sha: sha,
		url: url,
	}
}

func testFS() fstest.MapFS {
	return fstest.MapFS{
		"README.md": {Data: []byte("# Hello")},
		"SKILL.md":  {Data: []byte("skill content")},
	}
}

// --- Tests ---

func TestFromResolvedTreeBuiltinOnly(t *testing.T) {
	src := makeEmbeddedSource()
	tree := &registry.ResolvedTree{
		Order: []registry.ResolvedPackage{
			{Package: model.Package{Name: "threat-modelling", Version: "0.5.0"}, Source: src},
			{Package: model.Package{Name: "git-workflow", Version: "0.5.0"}, Source: src},
		},
		Direct: map[string]bool{
			"threat-modelling": true,
			"git-workflow":     true,
		},
		Graph: map[string][]string{
			"threat-modelling": {},
			"git-workflow":     {},
		},
	}

	lf, err := FromResolvedTree(tree, "v0.5.0")
	if err != nil {
		t.Fatalf("FromResolvedTree() error: %v", err)
	}

	if lf.LockfileVersion != CurrentVersion {
		t.Errorf("LockfileVersion = %d, want %d", lf.LockfileVersion, CurrentVersion)
	}
	if lf.CLIVersion != "v0.5.0" {
		t.Errorf("CLIVersion = %q, want %q", lf.CLIVersion, "v0.5.0")
	}
	if len(lf.Packages) != 2 {
		t.Fatalf("len(Packages) = %d, want 2", len(lf.Packages))
	}

	pkg := lf.Packages["threat-modelling"]
	if pkg.Type != TypeBuiltin {
		t.Errorf("Type = %q, want %q", pkg.Type, TypeBuiltin)
	}
	if pkg.Version != "v0.5.0" {
		t.Errorf("Version = %q, want %q", pkg.Version, "v0.5.0")
	}
	if !pkg.Direct {
		t.Error("Direct = false, want true")
	}
	if pkg.Source != "" {
		t.Errorf("Source = %q, want empty for builtin", pkg.Source)
	}
	if pkg.Resolved != "" {
		t.Errorf("Resolved = %q, want empty for builtin", pkg.Resolved)
	}
	if pkg.Integrity != "" {
		t.Errorf("Integrity = %q, want empty for builtin", pkg.Integrity)
	}
}

func TestFromResolvedTreeRemotePackage(t *testing.T) {
	content := testFS()
	src := makeGitSource(
		"https://github.com/org/skills.git",
		"abc123def456789abc123def456789abc123def4",
		content,
	)

	tree := &registry.ResolvedTree{
		Order: []registry.ResolvedPackage{
			{Package: model.Package{Name: "my-skill", Version: "v1.2.0"}, Source: src},
		},
		Direct: map[string]bool{"my-skill": true},
		Graph:  map[string][]string{"my-skill": {}},
	}

	lf, err := FromResolvedTree(tree, "v0.5.0")
	if err != nil {
		t.Fatalf("FromResolvedTree() error: %v", err)
	}

	pkg := lf.Packages["my-skill"]
	if pkg.Type != TypeRemote {
		t.Errorf("Type = %q, want %q", pkg.Type, TypeRemote)
	}
	if pkg.Version != "v1.2.0" {
		t.Errorf("Version = %q, want %q", pkg.Version, "v1.2.0")
	}
	if pkg.Source != "https://github.com/org/skills.git" {
		t.Errorf("Source = %q, want URL", pkg.Source)
	}
	if pkg.Resolved != "abc123def456789abc123def456789abc123def4" {
		t.Errorf("Resolved = %q, want SHA", pkg.Resolved)
	}
	if !strings.HasPrefix(pkg.Integrity, "sha256:") {
		t.Errorf("Integrity = %q, want sha256: prefix", pkg.Integrity)
	}
	if !pkg.Direct {
		t.Error("Direct = false, want true")
	}
}

func TestFromResolvedTreeDirectFlag(t *testing.T) {
	src := makeEmbeddedSource()
	tree := &registry.ResolvedTree{
		Order: []registry.ResolvedPackage{
			{Package: model.Package{Name: "direct-pkg"}, Source: src},
			{Package: model.Package{Name: "transitive-pkg"}, Source: src},
		},
		Direct: map[string]bool{
			"direct-pkg": true,
			// transitive-pkg is NOT in the Direct map
		},
		Graph: map[string][]string{
			"direct-pkg":     {"transitive-pkg"},
			"transitive-pkg": {},
		},
	}

	lf, err := FromResolvedTree(tree, "v0.5.0")
	if err != nil {
		t.Fatalf("FromResolvedTree() error: %v", err)
	}

	if !lf.Packages["direct-pkg"].Direct {
		t.Error("direct-pkg.Direct = false, want true")
	}
	if lf.Packages["transitive-pkg"].Direct {
		t.Error("transitive-pkg.Direct = true, want false")
	}
}

func TestFromResolvedTreeDependencyOf(t *testing.T) {
	src := makeEmbeddedSource()
	tree := &registry.ResolvedTree{
		Order: []registry.ResolvedPackage{
			{Package: model.Package{Name: "base-lib"}, Source: src},
			{Package: model.Package{Name: "parent-a"}, Source: src},
			{Package: model.Package{Name: "parent-b"}, Source: src},
		},
		Direct: map[string]bool{
			"parent-a": true,
			"parent-b": true,
			// base-lib is transitive
		},
		Graph: map[string][]string{
			"parent-a": {"base-lib"},
			"parent-b": {"base-lib"},
			"base-lib": {},
		},
	}

	lf, err := FromResolvedTree(tree, "v0.5.0")
	if err != nil {
		t.Fatalf("FromResolvedTree() error: %v", err)
	}

	pkg := lf.Packages["base-lib"]
	if pkg.Direct {
		t.Error("base-lib.Direct = true, want false")
	}
	if len(pkg.DependencyOf) != 2 {
		t.Fatalf("len(DependencyOf) = %d, want 2", len(pkg.DependencyOf))
	}
	// Should be sorted
	if pkg.DependencyOf[0] != "parent-a" || pkg.DependencyOf[1] != "parent-b" {
		t.Errorf("DependencyOf = %v, want [parent-a, parent-b]", pkg.DependencyOf)
	}
}

func TestFromResolvedTreeDirectPackageNoDependencyOf(t *testing.T) {
	src := makeEmbeddedSource()
	tree := &registry.ResolvedTree{
		Order: []registry.ResolvedPackage{
			{Package: model.Package{Name: "dep"}, Source: src},
			{Package: model.Package{Name: "root"}, Source: src},
		},
		Direct: map[string]bool{
			"root": true,
			"dep":  true, // explicitly requested even though it's also a dependency
		},
		Graph: map[string][]string{
			"root": {"dep"},
			"dep":  {},
		},
	}

	lf, err := FromResolvedTree(tree, "v0.5.0")
	if err != nil {
		t.Fatalf("FromResolvedTree() error: %v", err)
	}

	// Direct packages should NOT have DependencyOf populated
	pkg := lf.Packages["dep"]
	if len(pkg.DependencyOf) != 0 {
		t.Errorf("Direct pkg DependencyOf = %v, want empty", pkg.DependencyOf)
	}
}

func TestFromResolvedTreeMixedSources(t *testing.T) {
	embeddedSrc := makeEmbeddedSource()
	gitSrc := makeGitSource(
		"https://github.com/team/repo.git",
		"deadbeef"+strings.Repeat("0", 32),
		testFS(),
	)

	tree := &registry.ResolvedTree{
		Order: []registry.ResolvedPackage{
			{Package: model.Package{Name: "builtin-pkg", Version: "0.5.0"}, Source: embeddedSrc},
			{Package: model.Package{Name: "remote-pkg", Version: "v2.0.0"}, Source: gitSrc},
		},
		Direct: map[string]bool{
			"builtin-pkg": true,
			"remote-pkg":  true,
		},
		Graph: map[string][]string{
			"builtin-pkg": {},
			"remote-pkg":  {},
		},
	}

	lf, err := FromResolvedTree(tree, "v0.5.0")
	if err != nil {
		t.Fatalf("FromResolvedTree() error: %v", err)
	}

	if len(lf.Packages) != 2 {
		t.Fatalf("len(Packages) = %d, want 2", len(lf.Packages))
	}
	if lf.Packages["builtin-pkg"].Type != TypeBuiltin {
		t.Errorf("builtin-pkg type = %q, want %q", lf.Packages["builtin-pkg"].Type, TypeBuiltin)
	}
	if lf.Packages["remote-pkg"].Type != TypeRemote {
		t.Errorf("remote-pkg type = %q, want %q", lf.Packages["remote-pkg"].Type, TypeRemote)
	}
}

func TestFromResolvedTreeNonPinnableGitSource(t *testing.T) {
	// A git source that does NOT implement pinnable — tests the fallback path.
	src := &mockSource{
		name:    "my-source",
		typ:     "git",
		content: testFS(),
	}

	tree := &registry.ResolvedTree{
		Order: []registry.ResolvedPackage{
			{Package: model.Package{Name: "pkg", Version: "v1.0.0"}, Source: src},
		},
		Direct: map[string]bool{"pkg": true},
		Graph:  map[string][]string{"pkg": {}},
	}

	lf, err := FromResolvedTree(tree, "v0.5.0")
	if err != nil {
		t.Fatalf("FromResolvedTree() error: %v", err)
	}

	pkg := lf.Packages["pkg"]
	if pkg.Type != TypeRemote {
		t.Errorf("Type = %q, want %q", pkg.Type, TypeRemote)
	}
	// Falls back to Source.Name() when pinnable isn't satisfied
	if pkg.Source != "my-source" {
		t.Errorf("Source = %q, want %q", pkg.Source, "my-source")
	}
	// No SHA available
	if pkg.Resolved != "" {
		t.Errorf("Resolved = %q, want empty (non-pinnable source)", pkg.Resolved)
	}
	// Integrity should still be computed
	if !strings.HasPrefix(pkg.Integrity, "sha256:") {
		t.Errorf("Integrity = %q, want sha256: prefix", pkg.Integrity)
	}
}

func TestFromResolvedTreeIntegrityConsistency(t *testing.T) {
	content := testFS()
	src := makeGitSource("https://example.com/repo.git", "abc123", content)

	tree := &registry.ResolvedTree{
		Order: []registry.ResolvedPackage{
			{Package: model.Package{Name: "pkg", Version: "v1.0.0"}, Source: src},
		},
		Direct: map[string]bool{"pkg": true},
		Graph:  map[string][]string{"pkg": {}},
	}

	lf1, err := FromResolvedTree(tree, "v0.5.0")
	if err != nil {
		t.Fatalf("first call error: %v", err)
	}

	lf2, err := FromResolvedTree(tree, "v0.5.0")
	if err != nil {
		t.Fatalf("second call error: %v", err)
	}

	if lf1.Packages["pkg"].Integrity != lf2.Packages["pkg"].Integrity {
		t.Error("integrity hashes differ between identical calls")
	}
}

func TestFromResolvedTreeEmptyTree(t *testing.T) {
	tree := &registry.ResolvedTree{
		Order:  nil,
		Direct: map[string]bool{},
		Graph:  map[string][]string{},
	}

	lf, err := FromResolvedTree(tree, "v0.5.0")
	if err != nil {
		t.Fatalf("FromResolvedTree() error: %v", err)
	}

	if len(lf.Packages) != 0 {
		t.Errorf("len(Packages) = %d, want 0", len(lf.Packages))
	}
	if lf.CLIVersion != "v0.5.0" {
		t.Errorf("CLIVersion = %q, want %q", lf.CLIVersion, "v0.5.0")
	}
}

func TestFromResolvedTreeUnknownSourceType(t *testing.T) {
	src := &mockSource{
		name: "future-src",
		typ:  "oci",
	}

	tree := &registry.ResolvedTree{
		Order: []registry.ResolvedPackage{
			{Package: model.Package{Name: "oci-pkg", Version: "v3.0.0"}, Source: src},
		},
		Direct: map[string]bool{"oci-pkg": true},
		Graph:  map[string][]string{"oci-pkg": {}},
	}

	lf, err := FromResolvedTree(tree, "v0.5.0")
	if err != nil {
		t.Fatalf("FromResolvedTree() error: %v", err)
	}

	pkg := lf.Packages["oci-pkg"]
	if pkg.Type != "oci" {
		t.Errorf("Type = %q, want %q", pkg.Type, "oci")
	}
	if pkg.Version != "v3.0.0" {
		t.Errorf("Version = %q, want %q", pkg.Version, "v3.0.0")
	}
}

// errorSource is a git-like source whose DownloadPackage always fails.
type errorSource struct {
	mockSource
}

func (e *errorSource) DownloadPackage(_, _ string) (fs.FS, error) {
	return nil, fs.ErrNotExist
}

func TestFromResolvedTreeDownloadError(t *testing.T) {
	src := &errorSource{
		mockSource: mockSource{name: "broken", typ: "git"},
	}

	tree := &registry.ResolvedTree{
		Order: []registry.ResolvedPackage{
			{Package: model.Package{Name: "broken", Version: "v1.0.0"}, Source: src},
		},
		Direct: map[string]bool{"broken": true},
		Graph:  map[string][]string{"broken": {}},
	}

	_, err := FromResolvedTree(tree, "v0.5.0")
	if err == nil {
		t.Fatal("expected error for failed download, got nil")
	}
	if !strings.Contains(err.Error(), "broken") {
		t.Errorf("error %q should mention package name", err)
	}
}

func TestBuildReverseDeps(t *testing.T) {
	graph := map[string][]string{
		"a": {"b", "c"},
		"b": {"c"},
		"c": {},
	}

	reverse := buildReverseDeps(graph)

	// c is depended on by both a and b
	if len(reverse["c"]) != 2 {
		t.Fatalf("reverse[c] has %d parents, want 2", len(reverse["c"]))
	}
	// b is depended on by a
	if len(reverse["b"]) != 1 || reverse["b"][0] != "a" {
		t.Errorf("reverse[b] = %v, want [a]", reverse["b"])
	}
	// a has no parents
	if len(reverse["a"]) != 0 {
		t.Errorf("reverse[a] = %v, want empty", reverse["a"])
	}
}

func TestBuildReverseDepsEmpty(t *testing.T) {
	reverse := buildReverseDeps(map[string][]string{})
	if len(reverse) != 0 {
		t.Errorf("len(reverse) = %d, want 0", len(reverse))
	}
}
