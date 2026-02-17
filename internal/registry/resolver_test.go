package registry

import (
	"errors"
	"testing"
	"testing/fstest"

	"github.com/willvelida/code-minions/internal/model"
)

// ---------- Test helper ----------

// testEmbeddedFSWithPersonas creates a fake filesystem that has both
// packages AND personas. This is like setting up a miniature version
// of the real embedded filesystem for testing purposes.
//
// The layout:
//
//	packages/
//	├── git-workflow/
//	│   └── package.yaml
//	├── raise-pull-requests/
//	│   └── package.yaml
//	└── threat-modelling/
//	    └── package.yaml
//	personas/
//	├── senior-dev/
//	│   └── persona.yaml     ← references all 3 packages
//	└── empty-persona/
//	    └── persona.yaml     ← references 0 packages (invalid)
func testEmbeddedFSWithPersonas() fstest.MapFS {
	return fstest.MapFS{
		// --- Packages ---
		"packages/git-workflow/package.yaml": &fstest.MapFile{
			Data: []byte(`name: git-workflow
version: 0.1.0
description: Git workflow skill
`),
		},
		"packages/git-workflow/agents/git-workflow.agent.md": &fstest.MapFile{
			Data: []byte("# Git Agent"),
		},
		"packages/raise-pull-requests/package.yaml": &fstest.MapFile{
			Data: []byte(`name: raise-pull-requests
version: 0.1.0
description: Pull request skill
`),
		},
		"packages/raise-pull-requests/agents/raise-pull-requests.agent.md": &fstest.MapFile{
			Data: []byte("# PR Agent"),
		},
		"packages/threat-modelling/package.yaml": &fstest.MapFile{
			Data: []byte(`name: threat-modelling
version: 0.2.0
description: Threat modelling skill
`),
		},
		"packages/threat-modelling/agents/threat-modelling.agent.md": &fstest.MapFile{
			Data: []byte("# Threat Agent"),
		},

		// --- Personas ---
		// "senior-dev" references 3 packages
		"personas/senior-dev/persona.yaml": &fstest.MapFile{
			Data: []byte(`name: senior-dev
description: A senior developer persona
author: willvelida
packages:
  - name: git-workflow
  - name: raise-pull-requests
  - name: threat-modelling
instructions: |
  Prioritise code quality and security.
`),
		},

		// "empty-persona" has no packages — this should cause an error
		"personas/empty-persona/persona.yaml": &fstest.MapFile{
			Data: []byte(`name: empty-persona
description: A persona with no packages
packages: []
`),
		},

		// "bad-name" has a name in YAML that doesn't match directory
		"personas/bad-name/persona.yaml": &fstest.MapFile{
			Data: []byte(`name: wrong-name
description: Name mismatch test
packages:
  - name: git-workflow
`),
		},

		// "missing-pkg" references a package that doesn't exist
		"personas/missing-pkg/persona.yaml": &fstest.MapFile{
			Data: []byte(`name: missing-pkg
description: References a nonexistent package
packages:
  - name: git-workflow
  - name: does-not-exist
`),
		},
	}
}

// ---------- EmbeddedSource persona tests ----------

func TestEmbeddedSourceListPersonas(t *testing.T) {
	// This test checks that ListPersonas() correctly returns an error
	// when a persona directory contains an invalid persona.yaml
	// (in this case, "bad-name" has a name mismatch).
	src := NewEmbeddedSource(testEmbeddedFSWithPersonas())

	_, err := src.ListPersonas()
	// We expect an error because the test filesystem includes
	// "bad-name" which has a YAML name that doesn't match its
	// directory name.
	if err == nil {
		t.Fatal("expected error due to bad-name persona, but got nil")
	}

	t.Logf("got expected error: %v", err)
}

func TestEmbeddedSourceListPersonasHappyPath(t *testing.T) {
	// Create a clean filesystem with only valid personas.
	validFS := fstest.MapFS{
		"packages/git-workflow/package.yaml": &fstest.MapFile{
			Data: []byte("name: git-workflow\nversion: 0.1.0\ndescription: Git\n"),
		},
		"personas/senior-dev/persona.yaml": &fstest.MapFile{
			Data: []byte(`name: senior-dev
description: A senior developer persona
packages:
  - name: git-workflow
`),
		},
	}

	src := NewEmbeddedSource(validFS)

	personas, err := src.ListPersonas()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(personas) != 1 {
		t.Fatalf("expected 1 persona, got %d", len(personas))
	}

	if personas[0].Name != "senior-dev" {
		t.Errorf("Name: got %q, want %q", personas[0].Name, "senior-dev")
	}
	if personas[0].Description != "A senior developer persona" {
		t.Errorf("Description: got %q", personas[0].Description)
	}
}

func TestEmbeddedSourceListPersonasNoDirectory(t *testing.T) {
	// When there's no personas/ directory, ListPersonas should
	// return an empty list, NOT an error.
	noPersonasFS := fstest.MapFS{
		"packages/git-workflow/package.yaml": &fstest.MapFile{
			Data: []byte("name: git-workflow\nversion: 0.1.0\ndescription: Git\n"),
		},
	}

	src := NewEmbeddedSource(noPersonasFS)

	personas, err := src.ListPersonas()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(personas) != 0 {
		t.Errorf("expected 0 personas, got %d", len(personas))
	}
}

func TestEmbeddedSourceGetPersona(t *testing.T) {
	src := NewEmbeddedSource(testEmbeddedFSWithPersonas())

	persona, err := src.GetPersona("senior-dev")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that the persona was parsed correctly
	if persona.Name != "senior-dev" {
		t.Errorf("Name: got %q", persona.Name)
	}
	if persona.Description != "A senior developer persona" {
		t.Errorf("Description: got %q", persona.Description)
	}
	if persona.Author != "willvelida" {
		t.Errorf("Author: got %q", persona.Author)
	}
	if len(persona.Packages) != 3 {
		t.Fatalf("expected 3 packages, got %d", len(persona.Packages))
	}

	// Verify the package names are correct
	expectedPkgs := []string{"git-workflow", "raise-pull-requests", "threat-modelling"}
	for i, ref := range persona.Packages {
		if ref.Name != expectedPkgs[i] {
			t.Errorf("Package[%d]: got %q, want %q", i, ref.Name, expectedPkgs[i])
		}
	}

	// Check instructions were parsed
	if persona.Instructions == "" {
		t.Error("expected non-empty instructions")
	}
}

func TestEmbeddedSourceGetPersonaNotFound(t *testing.T) {
	src := NewEmbeddedSource(testEmbeddedFSWithPersonas())

	_, err := src.GetPersona("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent persona")
	}

	// The error should wrap ErrNotFound so callers can check with
	// errors.Is(). This is important for the Registry — it uses
	// ErrNotFound to know "try the next source."
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestEmbeddedSourceGetPersonaNameMismatch(t *testing.T) {
	// The "bad-name" persona has name: "wrong-name" in its YAML
	// but lives in the "bad-name" directory. This should be an error.
	src := NewEmbeddedSource(testEmbeddedFSWithPersonas())

	_, err := src.GetPersona("bad-name")
	if err == nil {
		t.Fatal("expected error for name mismatch")
	}

	t.Logf("got expected error: %v", err)
}

// ---------- PersonaResolver tests ----------

func TestPersonaResolverResolve(t *testing.T) {
	// Create a registry with our test filesystem, then use the
	// resolver to look up the "senior-dev" persona and resolve
	// all its packages.
	src := NewEmbeddedSource(testEmbeddedFSWithPersonas())
	reg := NewRegistry(src)
	resolver := NewPersonaResolver(reg)

	resolved, err := resolver.Resolve("senior-dev")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that the persona metadata is preserved
	if resolved.Persona.Name != "senior-dev" {
		t.Errorf("Persona.Name: got %q", resolved.Persona.Name)
	}

	// Check that all 3 packages were resolved
	if len(resolved.Packages) != 3 {
		t.Fatalf("expected 3 resolved packages, got %d", len(resolved.Packages))
	}

	// Verify each package has the expected name and version
	expected := map[string]string{
		"git-workflow":        "0.1.0",
		"raise-pull-requests": "0.1.0",
		"threat-modelling":    "0.2.0",
	}

	for _, rp := range resolved.Packages {
		wantVersion, ok := expected[rp.Package.Name]
		if !ok {
			t.Errorf("unexpected package: %q", rp.Package.Name)
			continue
		}
		if rp.Package.Version != wantVersion {
			t.Errorf("package %q version: got %q, want %q",
				rp.Package.Name, rp.Package.Version, wantVersion)
		}
		// Each resolved package should have a source
		if rp.Source == nil {
			t.Errorf("package %q has nil source", rp.Package.Name)
		}
	}
}

func TestPersonaResolverResolveNotFound(t *testing.T) {
	// Trying to resolve a persona that doesn't exist should fail.
	src := NewEmbeddedSource(testEmbeddedFSWithPersonas())
	reg := NewRegistry(src)
	resolver := NewPersonaResolver(reg)

	_, err := resolver.Resolve("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent persona")
	}
}

func TestPersonaResolverResolveEmptyPersona(t *testing.T) {
	// The "empty-persona" has packages: [] — the resolver should
	// reject this with a clear error message.
	src := NewEmbeddedSource(testEmbeddedFSWithPersonas())
	reg := NewRegistry(src)
	resolver := NewPersonaResolver(reg)

	_, err := resolver.Resolve("empty-persona")
	if err == nil {
		t.Fatal("expected error for empty persona")
	}

	t.Logf("got expected error: %v", err)
}

func TestPersonaResolverResolveMissingPackage(t *testing.T) {
	// The "missing-pkg" persona references "does-not-exist" — the
	// resolver should fail because it can't find that package.
	src := NewEmbeddedSource(testEmbeddedFSWithPersonas())
	reg := NewRegistry(src)
	resolver := NewPersonaResolver(reg)

	_, err := resolver.Resolve("missing-pkg")
	if err == nil {
		t.Fatal("expected error for persona with missing package")
	}

	t.Logf("got expected error: %v", err)
}

func TestPersonaResolverResolvePersonaDirect(t *testing.T) {
	// Test ResolvePersona() — the method that takes an already-loaded
	// persona struct instead of looking it up by name.
	src := NewEmbeddedSource(testEmbeddedFSWithPersonas())
	reg := NewRegistry(src)
	resolver := NewPersonaResolver(reg)

	// Build a persona struct manually (simulating a local file)
	persona := &model.Persona{
		Name:        "custom",
		Description: "A custom persona",
		Packages: []model.PackageRef{
			{Name: "git-workflow"},
		},
	}

	resolved, err := resolver.ResolvePersona(persona)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resolved.Packages) != 1 {
		t.Fatalf("expected 1 package, got %d", len(resolved.Packages))
	}

	if resolved.Packages[0].Package.Name != "git-workflow" {
		t.Errorf("Package.Name: got %q", resolved.Packages[0].Package.Name)
	}
}
