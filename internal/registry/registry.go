package registry

import (
	"errors"
	"fmt"
	"strings"

	"github.com/willvelida/code-minions/internal/model"
)

// ErrNotFound is a sentinel error returned by Source implementations
// when the requested entity does not exist. Registry methods use this
// to distinguish "not found" (try next source) from real failures.
var ErrNotFound = errors.New("not found")

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
		if err != nil && !errors.Is(err, ErrNotFound) {
			return nil, nil, fmt.Errorf("source %q: %w", src.Name(), err)
		}
	}
	return nil, nil, fmt.Errorf("package %q: %w", name, ErrNotFound)
}

// ResolvePersona finds a persona by name across all sources.
func (r *Registry) ResolvePersona(name string) (*model.Persona, Source, error) {
	for _, src := range r.sources {
		persona, err := src.GetPersona(name)
		if err == nil && persona != nil {
			return persona, src, nil
		}
		if err != nil && !errors.Is(err, ErrNotFound) {
			return nil, nil, fmt.Errorf("source %q: %w", src.Name(), err)
		}
	}
	return nil, nil, fmt.Errorf("persona %q: %w", name, ErrNotFound)
}

// ResolveTeam finds a team by name across all sources.
func (r *Registry) ResolveTeam(name string) (*model.Team, Source, error) {
	for _, src := range r.sources {
		team, err := src.GetTeam(name)
		if err == nil && team != nil {
			return team, src, nil
		}
		if err != nil && !errors.Is(err, ErrNotFound) {
			return nil, nil, fmt.Errorf("source %q: %w", src.Name(), err)
		}
	}
	return nil, nil, fmt.Errorf("team %q: %w", name, ErrNotFound)
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
				pkg.Source = src.Name()
				all = append(all, pkg)
			}
		}
	}

	return all, nil
}

// Search queries all sources and merges results, deduplicated by kind+name.
// When the same entity appears in multiple sources, the first source wins.
func (r *Registry) Search(query string) ([]model.SearchResult, error) {
	seen := make(map[string]bool)
	var all []model.SearchResult

	for _, src := range r.sources {
		results, err := src.Search(query)
		if err != nil {
			return nil, fmt.Errorf("source %q: %w", src.Name(), err)
		}
		for _, res := range results {
			key := res.Kind + ":" + res.Name
			if !seen[key] {
				seen[key] = true
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
