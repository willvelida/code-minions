package registry

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/willvelida/code-minions/internal/model"
)

// ---------- helpers ----------

// testDepFS builds a minimal MapFS containing the packages needed for
// dependency resolution tests. Each package has a package.yaml with
// name, version, and optional dependencies.
func testDepFS() fstest.MapFS {
	return fstest.MapFS{
		"packages/pkg-a/package.yaml": &fstest.MapFile{
			Data: []byte(`name: pkg-a
version: 1.0.0
description: Test package A
dependencies:
  - name: pkg-b
    version: ">=1.0.0"
  - name: pkg-c
`),
		},
		"packages/pkg-b/package.yaml": &fstest.MapFile{
			Data: []byte(`name: pkg-b
version: 1.2.0
description: Test package B
dependencies:
  - name: pkg-c
    version: ">=0.2.0"
`),
		},
		"packages/pkg-c/package.yaml": &fstest.MapFile{
			Data: []byte(`name: pkg-c
version: 0.3.0
description: Test package C
dependencies: []
`),
		},
		"packages/standalone/package.yaml": &fstest.MapFile{
			Data: []byte(`name: standalone
version: 1.0.0
description: Package with no dependencies
`),
		},
		"packages/cycle-a/package.yaml": &fstest.MapFile{
			Data: []byte(`name: cycle-a
version: 1.0.0
description: Cycle A
dependencies:
  - name: cycle-b
`),
		},
		"packages/cycle-b/package.yaml": &fstest.MapFile{
			Data: []byte(`name: cycle-b
version: 1.0.0
description: Cycle B
dependencies:
  - name: cycle-a
`),
		},
		"packages/conflict-a/package.yaml": &fstest.MapFile{
			Data: []byte(`name: conflict-a
version: 1.0.0
description: Requires shared >=2.0.0
dependencies:
  - name: shared
    version: ">=2.0.0"
`),
		},
		"packages/conflict-b/package.yaml": &fstest.MapFile{
			Data: []byte(`name: conflict-b
version: 1.0.0
description: Requires shared <2.0.0
dependencies:
  - name: shared
    version: "<2.0.0"
`),
		},
		"packages/shared/package.yaml": &fstest.MapFile{
			Data: []byte(`name: shared
version: 1.5.0
description: Shared package at version 1.5.0
dependencies: []
`),
		},
		// deep chain: deep-1 → deep-2 → deep-3 → deep-4 → deep-5
		"packages/deep-1/package.yaml": &fstest.MapFile{
			Data: []byte(`name: deep-1
version: 1.0.0
dependencies:
  - name: deep-2
`),
		},
		"packages/deep-2/package.yaml": &fstest.MapFile{
			Data: []byte(`name: deep-2
version: 1.0.0
dependencies:
  - name: deep-3
`),
		},
		"packages/deep-3/package.yaml": &fstest.MapFile{
			Data: []byte(`name: deep-3
version: 1.0.0
dependencies:
  - name: deep-4
`),
		},
		"packages/deep-4/package.yaml": &fstest.MapFile{
			Data: []byte(`name: deep-4
version: 1.0.0
dependencies:
  - name: deep-5
`),
		},
		"packages/deep-5/package.yaml": &fstest.MapFile{
			Data: []byte(`name: deep-5
version: 1.0.0
dependencies: []
`),
		},
		// indirect cycle: ic-a → ic-b → ic-c → ic-a
		"packages/ic-a/package.yaml": &fstest.MapFile{
			Data: []byte(`name: ic-a
version: 1.0.0
dependencies:
  - name: ic-b
`),
		},
		"packages/ic-b/package.yaml": &fstest.MapFile{
			Data: []byte(`name: ic-b
version: 1.0.0
dependencies:
  - name: ic-c
`),
		},
		"packages/ic-c/package.yaml": &fstest.MapFile{
			Data: []byte(`name: ic-c
version: 1.0.0
dependencies:
  - name: ic-a
`),
		},
		// compatible constraints: compat-a and compat-b both need shared
		// with constraints that shared 1.5.0 satisfies
		"packages/compat-a/package.yaml": &fstest.MapFile{
			Data: []byte(`name: compat-a
version: 1.0.0
dependencies:
  - name: shared
    version: ">=1.0.0"
`),
		},
		"packages/compat-b/package.yaml": &fstest.MapFile{
			Data: []byte(`name: compat-b
version: 1.0.0
dependencies:
  - name: shared
    version: "<2.0.0"
`),
		},
	}
}

// resolveRoots resolves package names to ResolvedPackage using the registry.
func resolveRoots(t *testing.T, reg *Registry, names ...string) []ResolvedPackage {
	t.Helper()
	var roots []ResolvedPackage
	for _, name := range names {
		pkg, src, err := reg.ResolvePackage(name)
		if err != nil {
			t.Fatalf("failed to resolve root %q: %v", name, err)
		}
		roots = append(roots, ResolvedPackage{Package: *pkg, Source: src})
	}
	return roots
}

// indexOf returns the position of name in the order slice, or -1.
func indexOf(order []ResolvedPackage, name string) int {
	for i, rp := range order {
		if rp.Package.Name == name {
			return i
		}
	}
	return -1
}

// ---------- Tests ----------

func TestDependencyResolver_NoDependencies(t *testing.T) {
	src := NewEmbeddedSource(testDepFS())
	reg := NewRegistry(src)
	resolver := NewDependencyResolver(reg)

	roots := resolveRoots(t, reg, "standalone")
	tree, err := resolver.Resolve(roots)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tree.Order) != 1 {
		t.Fatalf("expected 1 package, got %d", len(tree.Order))
	}
	if tree.Order[0].Package.Name != "standalone" {
		t.Errorf("expected standalone, got %q", tree.Order[0].Package.Name)
	}
	if !tree.Direct["standalone"] {
		t.Error("standalone should be marked as direct")
	}
	if deps := tree.Graph["standalone"]; len(deps) != 0 {
		t.Errorf("standalone should have no deps, got %v", deps)
	}
}

func TestDependencyResolver_LinearChain(t *testing.T) {
	// pkg-b → pkg-c (linear chain from a single root)
	src := NewEmbeddedSource(testDepFS())
	reg := NewRegistry(src)
	resolver := NewDependencyResolver(reg)

	roots := resolveRoots(t, reg, "pkg-b")
	tree, err := resolver.Resolve(roots)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tree.Order) != 2 {
		t.Fatalf("expected 2 packages, got %d: %v", len(tree.Order), names(tree.Order))
	}

	// pkg-c must come before pkg-b (dependency first)
	cIdx := indexOf(tree.Order, "pkg-c")
	bIdx := indexOf(tree.Order, "pkg-b")
	if cIdx == -1 || bIdx == -1 {
		t.Fatalf("missing packages in order: %v", names(tree.Order))
	}
	if cIdx >= bIdx {
		t.Errorf("pkg-c (idx=%d) should come before pkg-b (idx=%d)", cIdx, bIdx)
	}

	// Only pkg-b is direct
	if !tree.Direct["pkg-b"] {
		t.Error("pkg-b should be marked as direct")
	}
	if tree.Direct["pkg-c"] {
		t.Error("pkg-c should NOT be marked as direct")
	}
}

func TestDependencyResolver_Diamond(t *testing.T) {
	// pkg-a → pkg-b, pkg-c; pkg-b → pkg-c (diamond)
	// pkg-c should appear once, before both pkg-a and pkg-b
	src := NewEmbeddedSource(testDepFS())
	reg := NewRegistry(src)
	resolver := NewDependencyResolver(reg)

	roots := resolveRoots(t, reg, "pkg-a")
	tree, err := resolver.Resolve(roots)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tree.Order) != 3 {
		t.Fatalf("expected 3 packages (deduped), got %d: %v", len(tree.Order), names(tree.Order))
	}

	cIdx := indexOf(tree.Order, "pkg-c")
	bIdx := indexOf(tree.Order, "pkg-b")
	aIdx := indexOf(tree.Order, "pkg-a")

	if cIdx == -1 || bIdx == -1 || aIdx == -1 {
		t.Fatalf("missing packages: %v", names(tree.Order))
	}

	// pkg-c before pkg-b, pkg-b before pkg-a
	if cIdx >= bIdx {
		t.Errorf("pkg-c (idx=%d) should come before pkg-b (idx=%d)", cIdx, bIdx)
	}
	if bIdx >= aIdx {
		t.Errorf("pkg-b (idx=%d) should come before pkg-a (idx=%d)", bIdx, aIdx)
	}

	// Graph structure
	if deps := tree.Graph["pkg-a"]; len(deps) != 2 {
		t.Errorf("pkg-a should have 2 deps, got %v", deps)
	}
	if deps := tree.Graph["pkg-b"]; len(deps) != 1 || deps[0] != "pkg-c" {
		t.Errorf("pkg-b deps: got %v, want [pkg-c]", deps)
	}
	if deps := tree.Graph["pkg-c"]; len(deps) != 0 {
		t.Errorf("pkg-c should have no deps, got %v", deps)
	}
}

func TestDependencyResolver_DirectCycle(t *testing.T) {
	// cycle-a → cycle-b → cycle-a
	src := NewEmbeddedSource(testDepFS())
	reg := NewRegistry(src)
	resolver := NewDependencyResolver(reg)

	roots := resolveRoots(t, reg, "cycle-a")
	_, err := resolver.Resolve(roots)
	if err == nil {
		t.Fatal("expected cycle error, got nil")
	}
	if !strings.Contains(err.Error(), "circular dependency") {
		t.Errorf("error should mention 'circular dependency': %v", err)
	}
	// Error should include the cycle path
	if !strings.Contains(err.Error(), "cycle-a") || !strings.Contains(err.Error(), "cycle-b") {
		t.Errorf("error should mention both cycle-a and cycle-b: %v", err)
	}
}

func TestDependencyResolver_IndirectCycle(t *testing.T) {
	// ic-a → ic-b → ic-c → ic-a (3-node cycle)
	src := NewEmbeddedSource(testDepFS())
	reg := NewRegistry(src)
	resolver := NewDependencyResolver(reg)

	roots := resolveRoots(t, reg, "ic-a")
	_, err := resolver.Resolve(roots)
	if err == nil {
		t.Fatal("expected cycle error, got nil")
	}
	if !strings.Contains(err.Error(), "circular dependency") {
		t.Errorf("error should mention 'circular dependency': %v", err)
	}
	if !strings.Contains(err.Error(), "ic-a") {
		t.Errorf("error should include the cycle start node ic-a: %v", err)
	}
}

func TestDependencyResolver_MissingPackage(t *testing.T) {
	// Create a package that depends on something that doesn't exist
	fs := fstest.MapFS{
		"packages/orphan/package.yaml": &fstest.MapFile{
			Data: []byte(`name: orphan
version: 1.0.0
dependencies:
  - name: does-not-exist
`),
		},
	}

	src := NewEmbeddedSource(fs)
	reg := NewRegistry(src)
	resolver := NewDependencyResolver(reg)

	roots := resolveRoots(t, reg, "orphan")
	_, err := resolver.Resolve(roots)
	if err == nil {
		t.Fatal("expected error for missing dependency, got nil")
	}
	if !strings.Contains(err.Error(), "does-not-exist") {
		t.Errorf("error should mention the missing package: %v", err)
	}
	if !strings.Contains(err.Error(), "orphan") {
		t.Errorf("error should mention the requiring package: %v", err)
	}
}

func TestDependencyResolver_DeepTree(t *testing.T) {
	// deep-1 → deep-2 → deep-3 → deep-4 → deep-5 (5 levels)
	src := NewEmbeddedSource(testDepFS())
	reg := NewRegistry(src)
	resolver := NewDependencyResolver(reg)

	roots := resolveRoots(t, reg, "deep-1")
	tree, err := resolver.Resolve(roots)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tree.Order) != 5 {
		t.Fatalf("expected 5 packages, got %d: %v", len(tree.Order), names(tree.Order))
	}

	// Verify strict order: deep-5 first, deep-1 last
	for i := 0; i < 4; i++ {
		cur := indexOf(tree.Order, "deep-"+string(rune('5'-i)))
		next := indexOf(tree.Order, "deep-"+string(rune('4'-i)))
		if cur == -1 || next == -1 {
			t.Fatalf("missing deep package in order: %v", names(tree.Order))
		}
		if cur >= next {
			t.Errorf("deep-%d should come before deep-%d", 5-i, 4-i)
		}
	}

	// Only deep-1 is direct
	if !tree.Direct["deep-1"] {
		t.Error("deep-1 should be marked as direct")
	}
	for i := 2; i <= 5; i++ {
		name := "deep-" + string(rune('0'+i))
		if tree.Direct[name] {
			t.Errorf("%s should NOT be marked as direct", name)
		}
	}
}

func TestDependencyResolver_SemverConstraintSatisfied(t *testing.T) {
	// pkg-b requires pkg-c >=0.2.0, pkg-c is at 0.3.0 → should pass
	src := NewEmbeddedSource(testDepFS())
	reg := NewRegistry(src)
	resolver := NewDependencyResolver(reg)

	roots := resolveRoots(t, reg, "pkg-b")
	tree, err := resolver.Resolve(roots)
	if err != nil {
		t.Fatalf("expected constraint to be satisfied: %v", err)
	}
	if len(tree.Order) != 2 {
		t.Errorf("expected 2 packages, got %d", len(tree.Order))
	}
}

func TestDependencyResolver_SemverConstraintNotSatisfied(t *testing.T) {
	// conflict-a requires shared >=2.0.0 but shared is 1.5.0
	src := NewEmbeddedSource(testDepFS())
	reg := NewRegistry(src)
	resolver := NewDependencyResolver(reg)

	roots := resolveRoots(t, reg, "conflict-a")
	_, err := resolver.Resolve(roots)
	if err == nil {
		t.Fatal("expected version conflict error, got nil")
	}
	if !strings.Contains(err.Error(), "version conflict") {
		t.Errorf("error should mention 'version conflict': %v", err)
	}
	if !strings.Contains(err.Error(), ">=2.0.0") {
		t.Errorf("error should include the constraint: %v", err)
	}
	if !strings.Contains(err.Error(), "1.5.0") {
		t.Errorf("error should include the available version: %v", err)
	}
}

func TestDependencyResolver_CompatibleMultipleConstraints(t *testing.T) {
	// compat-a requires shared >=1.0.0, compat-b requires shared <2.0.0
	// shared is 1.5.0 → satisfies both
	src := NewEmbeddedSource(testDepFS())
	reg := NewRegistry(src)
	resolver := NewDependencyResolver(reg)

	roots := resolveRoots(t, reg, "compat-a", "compat-b")
	tree, err := resolver.Resolve(roots)
	if err != nil {
		t.Fatalf("expected compatible constraints to succeed: %v", err)
	}

	// Should have 3 packages: compat-a, compat-b, shared
	if len(tree.Order) != 3 {
		t.Fatalf("expected 3 packages, got %d: %v", len(tree.Order), names(tree.Order))
	}

	// shared should come before both compat-a and compat-b
	sharedIdx := indexOf(tree.Order, "shared")
	if sharedIdx == -1 {
		t.Fatal("shared not found in order")
	}
	for _, name := range []string{"compat-a", "compat-b"} {
		idx := indexOf(tree.Order, name)
		if idx == -1 {
			t.Fatalf("%s not found in order", name)
		}
		if sharedIdx >= idx {
			t.Errorf("shared (idx=%d) should come before %s (idx=%d)", sharedIdx, name, idx)
		}
	}
}

func TestDependencyResolver_IncompatibleMultipleConstraints(t *testing.T) {
	// conflict-a requires shared >=2.0.0, conflict-b requires shared <2.0.0
	// shared is 1.5.0 → fails conflict-a's constraint
	src := NewEmbeddedSource(testDepFS())
	reg := NewRegistry(src)
	resolver := NewDependencyResolver(reg)

	roots := resolveRoots(t, reg, "conflict-a", "conflict-b")
	_, err := resolver.Resolve(roots)
	if err == nil {
		t.Fatal("expected version conflict error, got nil")
	}
	if !strings.Contains(err.Error(), "version conflict") || !strings.Contains(err.Error(), "conflict-a") {
		t.Errorf("error should mention the conflicting package and constraint: %v", err)
	}
}

func TestDependencyResolver_EmptyVersionConstraint(t *testing.T) {
	// pkg-a depends on pkg-c with no version constraint → any version accepted
	fs := fstest.MapFS{
		"packages/no-ver/package.yaml": &fstest.MapFile{
			Data: []byte(`name: no-ver
version: 1.0.0
dependencies:
  - name: leaf
`),
		},
		"packages/leaf/package.yaml": &fstest.MapFile{
			Data: []byte(`name: leaf
version: 0.0.1
`),
		},
	}

	src := NewEmbeddedSource(fs)
	reg := NewRegistry(src)
	resolver := NewDependencyResolver(reg)

	roots := resolveRoots(t, reg, "no-ver")
	tree, err := resolver.Resolve(roots)
	if err != nil {
		t.Fatalf("empty constraint should accept any version: %v", err)
	}
	if len(tree.Order) != 2 {
		t.Errorf("expected 2 packages, got %d", len(tree.Order))
	}
}

func TestDependencyResolver_MultipleRoots(t *testing.T) {
	// Two roots: pkg-b and standalone (no shared deps)
	src := NewEmbeddedSource(testDepFS())
	reg := NewRegistry(src)
	resolver := NewDependencyResolver(reg)

	roots := resolveRoots(t, reg, "pkg-b", "standalone")
	tree, err := resolver.Resolve(roots)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tree.Order) != 3 {
		t.Fatalf("expected 3 packages (pkg-b, pkg-c, standalone), got %d: %v",
			len(tree.Order), names(tree.Order))
	}

	if !tree.Direct["pkg-b"] || !tree.Direct["standalone"] {
		t.Error("both roots should be marked as direct")
	}
	if tree.Direct["pkg-c"] {
		t.Error("pkg-c should NOT be marked as direct")
	}
}

func TestDependencyResolver_EmptyRoots(t *testing.T) {
	src := NewEmbeddedSource(testDepFS())
	reg := NewRegistry(src)
	resolver := NewDependencyResolver(reg)

	tree, err := resolver.Resolve(nil)
	if err != nil {
		t.Fatalf("empty roots should succeed: %v", err)
	}
	if len(tree.Order) != 0 {
		t.Errorf("expected 0 packages, got %d", len(tree.Order))
	}
}

func TestDependencyResolver_PackageWithNoVersion(t *testing.T) {
	// A package with no version field — constraints should be skipped
	fs := fstest.MapFS{
		"packages/parent/package.yaml": &fstest.MapFile{
			Data: []byte(`name: parent
version: 1.0.0
dependencies:
  - name: no-version-pkg
    version: ">=1.0.0"
`),
		},
		"packages/no-version-pkg/package.yaml": &fstest.MapFile{
			Data: []byte(`name: no-version-pkg
description: A package with no version field
`),
		},
	}

	src := NewEmbeddedSource(fs)
	reg := NewRegistry(src)
	resolver := NewDependencyResolver(reg)

	roots := resolveRoots(t, reg, "parent")
	tree, err := resolver.Resolve(roots)
	if err != nil {
		t.Fatalf("package without version should be accepted (constraint skipped): %v", err)
	}
	if len(tree.Order) != 2 {
		t.Errorf("expected 2 packages, got %d", len(tree.Order))
	}
}

func TestDependencyResolver_DirectMap_MultipleRoots(t *testing.T) {
	// Verify Direct map is correctly populated with multiple roots
	src := NewEmbeddedSource(testDepFS())
	reg := NewRegistry(src)
	resolver := NewDependencyResolver(reg)

	roots := resolveRoots(t, reg, "pkg-a", "pkg-c")
	tree, err := resolver.Resolve(roots)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Both pkg-a and pkg-c were requested directly
	if !tree.Direct["pkg-a"] {
		t.Error("pkg-a should be direct")
	}
	if !tree.Direct["pkg-c"] {
		t.Error("pkg-c should be direct (explicitly requested)")
	}
	// pkg-b is transitive from pkg-a
	if tree.Direct["pkg-b"] {
		t.Error("pkg-b should NOT be direct (only transitive)")
	}
}

func TestDependencyResolver_GraphStructure(t *testing.T) {
	// Verify the Graph field is correctly populated
	src := NewEmbeddedSource(testDepFS())
	reg := NewRegistry(src)
	resolver := NewDependencyResolver(reg)

	roots := resolveRoots(t, reg, "pkg-a")
	tree, err := resolver.Resolve(roots)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// pkg-a depends on pkg-b and pkg-c
	aDeps := tree.Graph["pkg-a"]
	if len(aDeps) != 2 {
		t.Fatalf("pkg-a should have 2 deps, got %v", aDeps)
	}
	foundB, foundC := false, false
	for _, d := range aDeps {
		if d == "pkg-b" {
			foundB = true
		}
		if d == "pkg-c" {
			foundC = true
		}
	}
	if !foundB || !foundC {
		t.Errorf("pkg-a deps should include pkg-b and pkg-c, got %v", aDeps)
	}

	// pkg-b depends on pkg-c
	bDeps := tree.Graph["pkg-b"]
	if len(bDeps) != 1 || bDeps[0] != "pkg-c" {
		t.Errorf("pkg-b deps: got %v, want [pkg-c]", bDeps)
	}

	// pkg-c has no deps
	cDeps := tree.Graph["pkg-c"]
	if len(cDeps) != 0 {
		t.Errorf("pkg-c should have no deps, got %v", cDeps)
	}
}

func TestNewDependencyResolver(t *testing.T) {
	src := NewEmbeddedSource(testDepFS())
	reg := NewRegistry(src)
	resolver := NewDependencyResolver(reg)

	if resolver == nil {
		t.Fatal("NewDependencyResolver returned nil")
	}
	if resolver.registry != reg {
		t.Error("resolver should reference the provided registry")
	}
}

// names extracts package names from a slice of ResolvedPackage for
// diagnostics in test failure messages.
func names(order []ResolvedPackage) []string {
	out := make([]string, len(order))
	for i, rp := range order {
		out[i] = rp.Package.Name
	}
	return out
}

func TestDependencyResolver_InvalidConstraint(t *testing.T) {
	fs := fstest.MapFS{
		"packages/bad-constraint/package.yaml": &fstest.MapFile{
			Data: []byte(`name: bad-constraint
version: 1.0.0
dependencies:
  - name: target
    version: "not-a-valid-constraint!!!"
`),
		},
		"packages/target/package.yaml": &fstest.MapFile{
			Data: []byte(`name: target
version: 1.0.0
`),
		},
	}

	src := NewEmbeddedSource(fs)
	reg := NewRegistry(src)
	resolver := NewDependencyResolver(reg)

	roots := resolveRoots(t, reg, "bad-constraint")
	_, err := resolver.Resolve(roots)
	if err == nil {
		t.Fatal("expected error for invalid constraint, got nil")
	}
	if !strings.Contains(err.Error(), "invalid version constraint") {
		t.Errorf("error should mention 'invalid version constraint': %v", err)
	}
}

func TestDependencyResolver_InvalidVersion(t *testing.T) {
	fs := fstest.MapFS{
		"packages/parent/package.yaml": &fstest.MapFile{
			Data: []byte(`name: parent
version: 1.0.0
dependencies:
  - name: bad-version
    version: ">=1.0.0"
`),
		},
		"packages/bad-version/package.yaml": &fstest.MapFile{
			Data: []byte(`name: bad-version
version: not-semver
`),
		},
	}

	src := NewEmbeddedSource(fs)
	reg := NewRegistry(src)
	resolver := NewDependencyResolver(reg)

	roots := resolveRoots(t, reg, "parent")
	_, err := resolver.Resolve(roots)
	if err == nil {
		t.Fatal("expected error for invalid version, got nil")
	}
	if !strings.Contains(err.Error(), "invalid version") {
		t.Errorf("error should mention 'invalid version': %v", err)
	}
}

// Verify that the existing DependencyRef model type is compatible
// with the resolver (no schema mismatch).
func TestDependencyRef_ResolverCompatibility(t *testing.T) {
	ref := model.DependencyRef{
		Name:    "test-pkg",
		Version: ">=1.0.0",
		Source:  "my-hub",
	}

	if ref.Name == "" || ref.Version == "" || ref.Source == "" {
		t.Fatal("DependencyRef fields should be populated")
	}
}

// Verify ResolvedTree zero-value is usable
func TestResolvedTree_ZeroValue(t *testing.T) {
	var tree ResolvedTree
	if tree.Order != nil {
		t.Error("zero-value Order should be nil")
	}
	if tree.Direct != nil {
		t.Error("zero-value Direct should be nil")
	}
	if tree.Graph != nil {
		t.Error("zero-value Graph should be nil")
	}
}

func TestDependencyResolver_SelfDependency(t *testing.T) {
	fs := fstest.MapFS{
		"packages/self-dep/package.yaml": &fstest.MapFile{
			Data: []byte(`name: self-dep
version: 1.0.0
dependencies:
  - name: self-dep
`),
		},
	}

	src := NewEmbeddedSource(fs)
	reg := NewRegistry(src)
	resolver := NewDependencyResolver(reg)

	roots := resolveRoots(t, reg, "self-dep")
	_, err := resolver.Resolve(roots)
	if err == nil {
		t.Fatal("expected cycle error for self-dependency, got nil")
	}
	if !strings.Contains(err.Error(), "circular dependency") {
		t.Errorf("error should mention 'circular dependency': %v", err)
	}
}
