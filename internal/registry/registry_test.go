package registry

import (
	"testing"
	"testing/fstest"
)

func TestRegistryResolvePackage(t *testing.T) {
	src := NewEmbeddedSource(testEmbeddedFS())
	reg := NewRegistry(src)

	pkg, source, err := reg.ResolvePackage("git-workflow")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pkg.Name != "git-workflow" {
		t.Errorf("Name: got %q", pkg.Name)
	}
	if source.Name() != "embedded" {
		t.Errorf("Source: got %q", source.Name())
	}
}

func TestRegistryResolvePackageNotFound(t *testing.T) {
	src := NewEmbeddedSource(testEmbeddedFS())
	reg := NewRegistry(src)

	_, _, err := reg.ResolvePackage("nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRegistryResolvePackagePriority(t *testing.T) {
	// Create two sources with the same package name.
	// First source should win.
	fs1 := fstest.MapFS{
		"packages/shared/package.yaml": &fstest.MapFile{
			Data: []byte("name: shared\nversion: 1.0.0\ndescription: From source 1\n"),
		},
	}
	fs2 := fstest.MapFS{
		"packages/shared/package.yaml": &fstest.MapFile{
			Data: []byte("name: shared\nversion: 2.0.0\ndescription: From source 2\n"),
		},
	}

	src1 := NewEmbeddedSource(fs1)
	src2 := NewEmbeddedSource(fs2)
	reg := NewRegistry(src1, src2)

	pkg, _, err := reg.ResolvePackage("shared")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pkg.Version != "1.0.0" {
		t.Errorf("Expected first source to win, got version %q", pkg.Version)
	}
}

func TestRegistryListPackagesDeduplicates(t *testing.T) {
	fs1 := fstest.MapFS{
		"packages/shared/package.yaml": &fstest.MapFile{
			Data: []byte("name: shared\nversion: 1.0.0\ndescription: From source 1\n"),
		},
	}
	fs2 := fstest.MapFS{
		"packages/shared/package.yaml": &fstest.MapFile{
			Data: []byte("name: shared\nversion: 2.0.0\ndescription: From source 2\n"),
		},
		"packages/unique/package.yaml": &fstest.MapFile{
			Data: []byte("name: unique\nversion: 0.1.0\ndescription: Only in source 2\n"),
		},
	}

	src1 := NewEmbeddedSource(fs1)
	src2 := NewEmbeddedSource(fs2)
	reg := NewRegistry(src1, src2)

	pkgs, err := reg.ListPackages()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages (deduplicated), got %d", len(pkgs))
	}

	// "shared" should be version 1.0.0 (from first source)
	for _, p := range pkgs {
		if p.Name == "shared" && p.Version != "1.0.0" {
			t.Errorf("shared should be from first source (v1.0.0), got %q", p.Version)
		}
	}
}

func TestRegistryResolvePersonaNotFound(t *testing.T) {
	src := NewEmbeddedSource(testEmbeddedFS())
	reg := NewRegistry(src)

	_, _, err := reg.ResolvePersona("nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRegistryResolveTeamNotFound(t *testing.T) {
	src := NewEmbeddedSource(testEmbeddedFS())
	reg := NewRegistry(src)

	_, _, err := reg.ResolveTeam("nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRegistrySearch(t *testing.T) {
	src := NewEmbeddedSource(testEmbeddedFS())
	reg := NewRegistry(src)

	results, err := reg.Search("workflow")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].Name != "git-workflow" {
		t.Errorf("Name: got %q", results[0].Name)
	}
}

func TestRegistrySources(t *testing.T) {
	src := NewEmbeddedSource(testEmbeddedFS())
	reg := NewRegistry(src)

	sources := reg.Sources()
	if len(sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(sources))
	}
	if sources[0].Name() != "embedded" {
		t.Errorf("Source name: got %q", sources[0].Name())
	}
}

func TestRegistryEmptySources(t *testing.T) {
	reg := NewRegistry()

	_, _, err := reg.ResolvePackage("anything")
	if err == nil {
		t.Fatal("expected error with no sources")
	}

	pkgs, err := reg.ListPackages()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 0 {
		t.Errorf("expected 0 packages, got %d", len(pkgs))
	}
}
