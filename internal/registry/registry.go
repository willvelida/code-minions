package registry

import (
	"fmt"
	"strings"

	"github.com/willvelida/code-minions/internal/model"
)

// Registry aggregates multiple Sources and resolves packages,
// personas, and teams across all of them. Sources are tried in
// order — first match wins.
type Registry struct {
	sources []Source
}

// NewRegistry creates a registry with the given sources.
// Source order determines priority: the first source containing
// a matching entity wins.
func NewRegistry(sources ...Source) *Registry {
	return &Registry{sources: sources}
}

// ResolvePackage finds a package by name across all sources,
// returning the first match and the source it came from.
func (r *Registry) ResolvePackage(name string) (*model.Package, Source, error) {
	for _, src := range r.sources {
		pkg, err := src.GetPackage(name)
		if err == nil && pkg != nil {
			return pkg, src, nil
		}
	}
	return nil, nil, fmt.Errorf("package %q not found in any source", name)
}

// ResolvePersona finds a persona by name across all sources.
func (r *Registry) ResolvePersona(name string) (*model.Persona, Source, error) {
	for _, src := range r.sources {
		persona, err := src.GetPersona(name)
		if err == nil && persona != nil {
			return persona, src, nil
		}
	}
	return nil, nil, fmt.Errorf("persona %q not found in any source", name)
}

// ResolveTeam finds a team by name across all sources.
func (r *Registry) ResolveTeam(name string) (*model.Team, Source, error) {
	for _, src := range r.sources {
		team, err := src.GetTeam(name)
		if err == nil && team != nil {
			return team, src, nil
		}
	}
	return nil, nil, fmt.Errorf("team %q not found in any source", name)
}

// ListPackages returns packages from all sources, deduplicated by name.
// When the same package exists in multiple sources, the first source wins.
func (r *Registry) ListPackages() ([]model.Package, error) {
	seen := make(map[string]bool)
	var all []model.Package

	for _, src := range r.sources {
		pkgs, err := src.ListPackages()
		if err != nil {
			return nil, fmt.Errorf("source %q: %w", src.Name(), err)
		}
		for _, pkg := range pkgs {
			if !seen[pkg.Name] {
				seen[pkg.Name] = true
				all = append(all, pkg)
			}
		}
	}

	return all, nil
}

// Search queries all sources and merges results, deduplicated by package name.
// When the same package appears in multiple sources, the first source wins.
func (r *Registry) Search(query string) ([]model.SearchResult, error) {
	seen := make(map[string]bool)
	var all []model.SearchResult

	for _, src := range r.sources {
		results, err := src.Search(query)
		if err != nil {
			return nil, fmt.Errorf("source %q: %w", src.Name(), err)
		}
		for _, res := range results {
			if !seen[res.Name] {
				seen[res.Name] = true
				all = append(all, res)
			}
		}
	}

	return all, nil
}

// Sources returns the list of registered sources.
func (r *Registry) Sources() []Source {
	dst := make([]Source, len(r.sources))
	copy(dst, r.sources)
	return dst
}

// matchesQuery returns true if the query appears (case-insensitive)
// in any of the given fields.
func matchesQuery(query string, fields ...string) bool {
	q := strings.ToLower(query)
	for _, f := range fields {
		if strings.Contains(strings.ToLower(f), q) {
			return true
		}
	}
	return false
}
