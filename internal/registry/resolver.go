package registry

import (
	"fmt"

	"github.com/willvelida/code-minions/internal/model"
)

// ResolvedPersona is the result of resolving a persona — it contains
// the persona metadata AND the full Package structs for every package
// referenced by that persona.
//
// Think of it this way:
//   - model.Persona says: "I need packages: git-workflow, threat-modelling"
//   - ResolvedPersona says: "Here are the actual packages with all their
//     files, versions, descriptions, and the source they came from"
type ResolvedPersona struct {
	Persona  model.Persona     // The original persona metadata
	Packages []ResolvedPackage // Each package fully resolved with its source
}

// ResolvedPackage pairs a Package with the Source it was found in.
// We track the source because during installation we'll need to
// call source.DownloadPackage() to get the actual files.
type ResolvedPackage struct {
	Package model.Package
	Source  Source
}

// PersonaResolver looks up a persona and resolves all of its package
// references into full Package structs. It uses the Registry to find
// packages across all registered sources.
//
// Usage:
//
//	resolver := NewPersonaResolver(registry)
//	resolved, err := resolver.Resolve("senior-dev")
//	// resolved.Packages now contains the full Package for each
//	// package referenced by the "senior-dev" persona
type PersonaResolver struct {
	registry *Registry
}

// NewPersonaResolver creates a resolver backed by the given registry.
// The registry is used to find both the persona itself and its packages.
func NewPersonaResolver(registry *Registry) *PersonaResolver {
	return &PersonaResolver{registry: registry}
}

// Resolve looks up the named persona and resolves each of its package
// references to a full Package struct.
//
// It returns an error if:
//   - The persona is not found in any source
//   - Any of the persona's packages cannot be found
//   - The persona has no packages (empty personas are not useful)
func (r *PersonaResolver) Resolve(personaName string) (*ResolvedPersona, error) {
	// Step 1: Find the persona in the registry.
	// The registry checks all sources in priority order.
	persona, _, err := r.registry.ResolvePersona(personaName)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve persona: %w", err)
	}

	// Step 2: Validate that the persona has at least one package.
	// A persona with no packages is like a recipe with no ingredients
	// — it's valid YAML but not useful.
	if len(persona.Packages) == 0 {
		return nil, fmt.Errorf("persona %q has no packages", personaName)
	}

	// Step 3: Resolve each package reference to a full Package struct.
	// We iterate over the persona's PackageRef list and look up each
	// one in the registry.
	resolved := &ResolvedPersona{
		Persona:  *persona,
		Packages: make([]ResolvedPackage, 0, len(persona.Packages)),
	}

	for _, ref := range persona.Packages {
		// Look up the package by name across all sources.
		pkg, src, err := r.registry.ResolvePackage(ref.Name)
		if err != nil {
			return nil, fmt.Errorf("persona %q: package %q: %w",
				personaName, ref.Name, err)
		}

		resolved.Packages = append(resolved.Packages, ResolvedPackage{
			Package: *pkg,
			Source:  src,
		})
	}

	return resolved, nil
}

// ResolvePersona is a convenience method that resolves a Persona
// struct directly (when you already have the struct and don't need
// to look it up by name).
//
// This is useful when the persona comes from a different source
// than the registry (e.g. a local persona.yaml file).
func (r *PersonaResolver) ResolvePersona(persona *model.Persona) (*ResolvedPersona, error) {
	if len(persona.Packages) == 0 {
		return nil, fmt.Errorf("persona %q has no packages", persona.Name)
	}

	resolved := &ResolvedPersona{
		Persona:  *persona,
		Packages: make([]ResolvedPackage, 0, len(persona.Packages)),
	}

	for _, ref := range persona.Packages {
		pkg, src, err := r.registry.ResolvePackage(ref.Name)
		if err != nil {
			return nil, fmt.Errorf("persona %q: package %q: %w",
				persona.Name, ref.Name, err)
		}

		resolved.Packages = append(resolved.Packages, ResolvedPackage{
			Package: *pkg,
			Source:  src,
		})
	}

	return resolved, nil
}
