package lockfile

import "sort"

// StalenessResult describes how a lockfile relates to the project manifest.
type StalenessResult struct {
	// Fresh is true when the lockfile is fully up to date with the manifest.
	Fresh bool

	// Added lists packages that are in the manifest but not in the lockfile.
	// These need to be resolved and added.
	Added []string

	// Removed lists packages that are in the lockfile but not in the manifest.
	// These should be pruned from the lockfile.
	Removed []string
}

// CheckStaleness compares the manifest's package list against the lockfile
// and returns a result describing what's changed.
//
// Only direct packages are considered — transitive dependencies in the
// lockfile are not compared against the manifest (they are resolved
// during dependency resolution, not declared in the manifest).
//
// When lf is nil (no lockfile exists), all manifest packages are reported
// as Added and Fresh is false.
func CheckStaleness(lf *LockFile, manifestPackages []string) *StalenessResult {
	result := &StalenessResult{}

	if lf == nil {
		result.Added = make([]string, len(manifestPackages))
		copy(result.Added, manifestPackages)
		sort.Strings(result.Added)
		return result
	}

	// Build a set of direct lockfile packages.
	locked := make(map[string]bool)
	for name, pkg := range lf.Packages {
		if pkg.Direct {
			locked[name] = true
		}
	}

	// Build a set of manifest packages.
	manifest := make(map[string]bool, len(manifestPackages))
	for _, name := range manifestPackages {
		manifest[name] = true
	}

	// Find packages in manifest but not in lockfile (Added).
	for _, name := range manifestPackages {
		if !locked[name] {
			result.Added = append(result.Added, name)
		}
	}

	// Find direct packages in lockfile but not in manifest (Removed).
	for name := range locked {
		if !manifest[name] {
			result.Removed = append(result.Removed, name)
		}
	}

	sort.Strings(result.Added)
	sort.Strings(result.Removed)

	result.Fresh = len(result.Added) == 0 && len(result.Removed) == 0
	return result
}
