package lockfile

import "testing"

func TestCheckStalenessNilLockfile(t *testing.T) {
	result := CheckStaleness(nil, []string{"pkg-a", "pkg-b"})

	if result.Fresh {
		t.Error("expected staleness when lockfile is nil")
	}
	if len(result.Added) != 2 {
		t.Fatalf("Added: got %d, want 2", len(result.Added))
	}
	// Should be sorted
	if result.Added[0] != "pkg-a" || result.Added[1] != "pkg-b" {
		t.Errorf("Added: got %v, want [pkg-a pkg-b]", result.Added)
	}
	if len(result.Removed) != 0 {
		t.Errorf("Removed: got %v, want empty", result.Removed)
	}
}

func TestCheckStalenessFresh(t *testing.T) {
	lf := New("v1.0.0")
	lf.Packages["pkg-a"] = LockedPackage{Type: TypeBuiltin, Version: "v1.0.0", Direct: true}
	lf.Packages["pkg-b"] = LockedPackage{Type: TypeBuiltin, Version: "v1.0.0", Direct: true}

	result := CheckStaleness(lf, []string{"pkg-a", "pkg-b"})

	if !result.Fresh {
		t.Error("expected fresh when manifest matches lockfile")
	}
	if len(result.Added) != 0 {
		t.Errorf("Added: got %v, want empty", result.Added)
	}
	if len(result.Removed) != 0 {
		t.Errorf("Removed: got %v, want empty", result.Removed)
	}
}

func TestCheckStalenessAddedPackages(t *testing.T) {
	lf := New("v1.0.0")
	lf.Packages["pkg-a"] = LockedPackage{Type: TypeBuiltin, Version: "v1.0.0", Direct: true}

	result := CheckStaleness(lf, []string{"pkg-a", "pkg-b", "pkg-c"})

	if result.Fresh {
		t.Error("expected stale when manifest has new packages")
	}
	if len(result.Added) != 2 {
		t.Fatalf("Added: got %d, want 2", len(result.Added))
	}
	if result.Added[0] != "pkg-b" || result.Added[1] != "pkg-c" {
		t.Errorf("Added: got %v, want [pkg-b pkg-c]", result.Added)
	}
}

func TestCheckStalenessRemovedPackages(t *testing.T) {
	lf := New("v1.0.0")
	lf.Packages["pkg-a"] = LockedPackage{Type: TypeBuiltin, Version: "v1.0.0", Direct: true}
	lf.Packages["pkg-b"] = LockedPackage{Type: TypeBuiltin, Version: "v1.0.0", Direct: true}
	lf.Packages["pkg-c"] = LockedPackage{Type: TypeBuiltin, Version: "v1.0.0", Direct: true}

	result := CheckStaleness(lf, []string{"pkg-a"})

	if result.Fresh {
		t.Error("expected stale when manifest has removed packages")
	}
	if len(result.Removed) != 2 {
		t.Fatalf("Removed: got %d, want 2", len(result.Removed))
	}
	if result.Removed[0] != "pkg-b" || result.Removed[1] != "pkg-c" {
		t.Errorf("Removed: got %v, want [pkg-b pkg-c]", result.Removed)
	}
}

func TestCheckStalenessAddedAndRemoved(t *testing.T) {
	lf := New("v1.0.0")
	lf.Packages["old-pkg"] = LockedPackage{Type: TypeBuiltin, Version: "v1.0.0", Direct: true}

	result := CheckStaleness(lf, []string{"new-pkg"})

	if result.Fresh {
		t.Error("expected stale")
	}
	if len(result.Added) != 1 || result.Added[0] != "new-pkg" {
		t.Errorf("Added: got %v, want [new-pkg]", result.Added)
	}
	if len(result.Removed) != 1 || result.Removed[0] != "old-pkg" {
		t.Errorf("Removed: got %v, want [old-pkg]", result.Removed)
	}
}

func TestCheckStalenessIgnoresTransitiveDeps(t *testing.T) {
	lf := New("v1.0.0")
	lf.Packages["direct-pkg"] = LockedPackage{Type: TypeBuiltin, Version: "v1.0.0", Direct: true}
	// Transitive dep — not in manifest, but should not be reported as Removed
	lf.Packages["transitive-dep"] = LockedPackage{
		Type:         TypeBuiltin,
		Version:      "v1.0.0",
		Direct:       false,
		DependencyOf: []string{"direct-pkg"},
	}

	result := CheckStaleness(lf, []string{"direct-pkg"})

	if !result.Fresh {
		t.Error("expected fresh — transitive deps should not cause staleness")
	}
	if len(result.Removed) != 0 {
		t.Errorf("Removed: got %v, want empty (transitive deps should be ignored)", result.Removed)
	}
}

func TestCheckStalenessEmptyManifest(t *testing.T) {
	lf := New("v1.0.0")
	lf.Packages["pkg-a"] = LockedPackage{Type: TypeBuiltin, Version: "v1.0.0", Direct: true}

	result := CheckStaleness(lf, []string{})

	if result.Fresh {
		t.Error("expected stale when manifest is empty but lockfile has packages")
	}
	if len(result.Removed) != 1 {
		t.Errorf("Removed: got %v, want [pkg-a]", result.Removed)
	}
}

func TestCheckStalenessEmptyBoth(t *testing.T) {
	lf := New("v1.0.0")
	result := CheckStaleness(lf, []string{})

	if !result.Fresh {
		t.Error("expected fresh when both manifest and lockfile are empty")
	}
}
