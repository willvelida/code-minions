package registry

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// ResolvedTree is the output of dependency resolution. It contains
// the full set of packages in topological install order (dependencies
// before dependents), along with metadata about which packages were
// explicitly requested vs pulled in transitively.
type ResolvedTree struct {
	// Order lists packages in topological install order: dependencies
	// appear before the packages that depend on them. Installing in
	// this order guarantees that every package's dependencies are
	// already present when it is installed.
	Order []ResolvedPackage

	// Direct contains the names of packages that were explicitly
	// requested by the user (not pulled in as transitive dependencies).
	Direct map[string]bool

	// Graph maps each package name to the names of its direct
	// dependencies. This is useful for tree visualisation and
	// the `why` command. Packages with no dependencies have an
	// empty (but present) slice.
	Graph map[string][]string
}

// DependencyResolver resolves the full transitive dependency tree
// for a set of packages. It uses the Registry to find packages across
// all registered sources, detects cycles, evaluates semver constraints,
// and produces a topologically-sorted install order.
//
// Usage mirrors PersonaResolver:
//
//	resolver := NewDependencyResolver(registry)
//	tree, err := resolver.Resolve(rootPackages)
//	// tree.Order is the install order (deps first)
type DependencyResolver struct {
	registry *Registry
}

// NewDependencyResolver creates a resolver backed by the given registry.
func NewDependencyResolver(registry *Registry) *DependencyResolver {
	return &DependencyResolver{registry: registry}
}

// nodeState tracks whether a node has been visited during DFS for
// cycle detection (white/grey/black colouring).
type nodeState int

const (
	unvisited nodeState = iota // white — not yet seen
	visiting                   // grey — currently on the DFS stack (cycle if revisited)
	visited                    // black — fully processed
)

// Resolve takes the root packages (explicitly requested by the user)
// and returns a ResolvedTree containing the full transitive closure
// in topological order.
//
// It returns an error if:
//   - A dependency is not found in any registry source
//   - A circular dependency is detected (with full cycle path)
//   - A semver constraint cannot be satisfied by the available version
//   - Multiple constraints on the same package are incompatible
func (r *DependencyResolver) Resolve(roots []ResolvedPackage) (*ResolvedTree, error) {
	tree := &ResolvedTree{
		Direct: make(map[string]bool, len(roots)),
		Graph:  make(map[string][]string),
	}

	// Track all resolved packages by name for deduplication.
	resolved := make(map[string]ResolvedPackage)

	// Track DFS state for cycle detection.
	state := make(map[string]nodeState)

	// Track the DFS path for cycle error messages.
	var path []string

	// Collect all semver constraints per package name. When the same
	// package is referenced by multiple dependents with different version
	// constraints, we need to verify the available version satisfies ALL
	// of them.
	constraints := make(map[string][]constraintEntry)

	// Seed the resolved map and mark roots as direct.
	for _, root := range roots {
		resolved[root.Package.Name] = root
		tree.Direct[root.Package.Name] = true
	}

	// topOrder collects packages in reverse-post-order (dependencies first).
	var topOrder []ResolvedPackage

	// DFS visit function with cycle detection and topological ordering.
	var visit func(name string) error
	visit = func(name string) error {
		switch state[name] {
		case visited:
			return nil // already fully processed
		case visiting:
			// Cycle detected — build the cycle path from the current
			// DFS stack. Find where the cycle starts and format a
			// clear error message.
			cycleStart := -1
			for i, p := range path {
				if p == name {
					cycleStart = i
					break
				}
			}
			cyclePath := append(path[cycleStart:], name)
			return fmt.Errorf("circular dependency detected: %s",
				strings.Join(cyclePath, " → "))
		}

		// Mark as visiting (grey) and push onto path.
		state[name] = visiting
		path = append(path, name)

		pkg, ok := resolved[name]
		if !ok {
			return fmt.Errorf("package %q not found (this should not happen — resolved map is inconsistent)", name)
		}

		// Build the dependency list for the graph.
		depNames := make([]string, 0, len(pkg.Package.Dependencies))

		for _, dep := range pkg.Package.Dependencies {
			depNames = append(depNames, dep.Name)

			// Record the constraint for later validation.
			if dep.Version != "" {
				constraints[dep.Name] = append(constraints[dep.Name], constraintEntry{
					constraint: dep.Version,
					requiredBy: name,
				})
			}

			// Resolve the dependency if we haven't seen it yet.
			if _, seen := resolved[dep.Name]; !seen {
				depPkg, depSrc, err := r.registry.ResolvePackage(dep.Name)
				if err != nil {
					return fmt.Errorf("package %q requires %q: %w", name, dep.Name, err)
				}
				resolved[dep.Name] = ResolvedPackage{
					Package: *depPkg,
					Source:  depSrc,
				}
			}

			// Recurse into the dependency.
			if err := visit(dep.Name); err != nil {
				return err
			}
		}

		tree.Graph[name] = depNames

		// Mark as visited (black) and pop from path.
		state[name] = visited
		path = path[:len(path)-1]

		// Post-order: add to topological order after all deps are processed.
		topOrder = append(topOrder, resolved[name])

		return nil
	}

	// Visit each root package.
	for _, root := range roots {
		if err := visit(root.Package.Name); err != nil {
			return nil, err
		}
	}

	// Validate all semver constraints.
	if err := validateConstraints(constraints, resolved); err != nil {
		return nil, err
	}

	tree.Order = topOrder
	return tree, nil
}

// constraintEntry pairs a semver constraint string with the package
// that declared it. This enables clear error messages like:
// "package X requires shared >=2.0.0 but available version is 1.5.0"
type constraintEntry struct {
	constraint string // e.g. ">=1.0.0", "^2.0.0"
	requiredBy string // name of the package that declared this constraint
}

// validateConstraints checks that the available version of each package
// satisfies all semver constraints declared by its dependents.
func validateConstraints(
	constraints map[string][]constraintEntry,
	resolved map[string]ResolvedPackage,
) error {
	for pkgName, entries := range constraints {
		pkg, ok := resolved[pkgName]
		if !ok {
			continue // shouldn't happen — resolver already checked
		}

		availableVersion := pkg.Package.Version
		if availableVersion == "" {
			// No version declared on the package — constraints cannot be
			// evaluated. This is a warning case, not an error, because
			// many packages start without version metadata.
			continue
		}

		ver, err := semver.NewVersion(availableVersion)
		if err != nil {
			return fmt.Errorf("package %q has invalid version %q: %w",
				pkgName, availableVersion, err)
		}

		for _, entry := range entries {
			c, err := semver.NewConstraint(entry.constraint)
			if err != nil {
				return fmt.Errorf("package %q declares invalid version constraint %q for %q: %w",
					entry.requiredBy, entry.constraint, pkgName, err)
			}

			if !c.Check(ver) {
				return fmt.Errorf(
					"version conflict: package %q requires %s %s, but available version is %s",
					entry.requiredBy, pkgName, entry.constraint, availableVersion,
				)
			}
		}
	}

	return nil
}
