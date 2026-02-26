package lockfile

import (
	"fmt"
	"sort"
	"time"

	"github.com/willvelida/code-minions/internal/registry"
)

// pinnable is satisfied by source implementations that expose
// version-pinning metadata for the lockfile. GitSource implements
// this via its HeadSHA() and URL() methods. Using a local interface
// (Go convention: define interfaces where consumed) avoids coupling
// the lockfile package to concrete source types.
type pinnable interface {
	HeadSHA() string
	URL() string
}

// FromResolvedTree converts a ResolvedTree (the output of dependency
// resolution) into a LockFile. It captures:
//   - Package type (builtin vs remote)
//   - Version strings
//   - Git commit SHAs for remote packages (via HeadSHA())
//   - Content integrity hashes for remote packages
//   - Direct/transitive classification
//   - Reverse dependency graph (DependencyOf)
//
// The cliVersion parameter is used as the version for builtin packages
// (since their content is baked into the binary and changes only when
// the CLI is rebuilt).
func FromResolvedTree(tree *registry.ResolvedTree, cliVersion string) (*LockFile, error) {
	lf := &LockFile{
		LockfileVersion: CurrentVersion,
		GeneratedAt:     time.Now().UTC().Format(time.RFC3339),
		CLIVersion:      cliVersion,
		Packages:        make(map[string]LockedPackage),
	}

	// Build reverse dependency map: child → parents.
	reverseDeps := buildReverseDeps(tree.Graph)

	for _, rp := range tree.Order {
		name := rp.Package.Name
		entry := LockedPackage{
			Direct: tree.Direct[name],
		}

		switch rp.Source.Type() {
		case "embedded":
			entry.Type = TypeBuiltin
			entry.Version = cliVersion

		case "git":
			entry.Type = TypeRemote
			entry.Version = rp.Package.Version

			// Extract pinning metadata via interface check.
			// GitSource satisfies pinnable; future source types
			// (OCI, HTTP registry) can do the same.
			if p, ok := rp.Source.(pinnable); ok {
				entry.Source = p.URL()
				entry.Resolved = p.HeadSHA()
			} else {
				entry.Source = rp.Source.Name()
			}

			// Compute content integrity hash.
			content, err := rp.Source.DownloadPackage(name, rp.Package.Version)
			if err != nil {
				return nil, fmt.Errorf("failed to download %q for integrity hash: %w", name, err)
			}
			integrity, err := ComputePackageIntegrity(content)
			if err != nil {
				return nil, fmt.Errorf("failed to compute integrity for %q: %w", name, err)
			}
			entry.Integrity = integrity

		default:
			// Unknown source type — record what we can.
			entry.Type = rp.Source.Type()
			entry.Version = rp.Package.Version
		}

		// Populate DependencyOf for transitive deps.
		if !entry.Direct {
			if parents, ok := reverseDeps[name]; ok && len(parents) > 0 {
				sorted := make([]string, len(parents))
				copy(sorted, parents)
				sort.Strings(sorted)
				entry.DependencyOf = sorted
			}
		}

		lf.Packages[name] = entry
	}

	return lf, nil
}

// buildReverseDeps converts a forward dependency graph (parent → children)
// into a reverse graph (child → parents).
func buildReverseDeps(graph map[string][]string) map[string][]string {
	reverse := make(map[string][]string)
	for parent, children := range graph {
		for _, child := range children {
			reverse[child] = append(reverse[child], parent)
		}
	}
	return reverse
}
