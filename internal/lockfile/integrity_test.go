package lockfile

import (
	"io/fs"
	"testing"
	"testing/fstest"
)

func TestComputePackageIntegrityDeterministic(t *testing.T) {
	fsys := fstest.MapFS{
		"file-a.txt":     &fstest.MapFile{Data: []byte("content a")},
		"file-b.txt":     &fstest.MapFile{Data: []byte("content b")},
		"sub/file-c.txt": &fstest.MapFile{Data: []byte("content c")},
	}

	hash1, err := ComputePackageIntegrity(fsys)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}
	hash2, err := ComputePackageIntegrity(fsys)
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("same FS produced different hashes:\n  %s\n  %s", hash1, hash2)
	}

	if hash1[:7] != "sha256:" {
		t.Errorf("expected sha256 prefix, got: %s", hash1[:10])
	}
}

func TestComputePackageIntegrityOrderIndependent(t *testing.T) {
	// Two MapFS instances with the same content but declared in
	// different order should produce identical hashes because we
	// sort the paths before hashing.
	fs1 := fstest.MapFS{
		"z.txt": &fstest.MapFile{Data: []byte("z content")},
		"a.txt": &fstest.MapFile{Data: []byte("a content")},
	}
	fs2 := fstest.MapFS{
		"a.txt": &fstest.MapFile{Data: []byte("a content")},
		"z.txt": &fstest.MapFile{Data: []byte("z content")},
	}

	hash1, err := ComputePackageIntegrity(fs1)
	if err != nil {
		t.Fatalf("fs1 failed: %v", err)
	}
	hash2, err := ComputePackageIntegrity(fs2)
	if err != nil {
		t.Fatalf("fs2 failed: %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("different declaration order produced different hashes:\n  %s\n  %s", hash1, hash2)
	}
}

func TestComputePackageIntegrityDifferentContent(t *testing.T) {
	fs1 := fstest.MapFS{
		"file.txt": &fstest.MapFile{Data: []byte("content version 1")},
	}
	fs2 := fstest.MapFS{
		"file.txt": &fstest.MapFile{Data: []byte("content version 2")},
	}

	hash1, _ := ComputePackageIntegrity(fs1)
	hash2, _ := ComputePackageIntegrity(fs2)

	if hash1 == hash2 {
		t.Error("different content should produce different hashes")
	}
}

func TestComputePackageIntegrityDifferentPaths(t *testing.T) {
	fs1 := fstest.MapFS{
		"path-a.txt": &fstest.MapFile{Data: []byte("same content")},
	}
	fs2 := fstest.MapFS{
		"path-b.txt": &fstest.MapFile{Data: []byte("same content")},
	}

	hash1, _ := ComputePackageIntegrity(fs1)
	hash2, _ := ComputePackageIntegrity(fs2)

	if hash1 == hash2 {
		t.Error("same content at different paths should produce different hashes")
	}
}

func TestComputePackageIntegrityEmptyFS(t *testing.T) {
	fsys := fstest.MapFS{}

	hash, err := ComputePackageIntegrity(fsys)
	if err != nil {
		t.Fatalf("empty FS failed: %v", err)
	}

	if hash == "" {
		t.Error("expected non-empty hash for empty FS")
	}

	// Empty FS should still produce a valid sha256 hash
	if hash[:7] != "sha256:" {
		t.Errorf("expected sha256 prefix, got: %s", hash)
	}
}

func TestComputePackageIntegritySubdirectories(t *testing.T) {
	fsys := fstest.MapFS{
		"top.txt":           &fstest.MapFile{Data: []byte("top")},
		"sub/nested.txt":    &fstest.MapFile{Data: []byte("nested")},
		"sub/deep/leaf.txt": &fstest.MapFile{Data: []byte("leaf")},
	}

	hash, err := ComputePackageIntegrity(fsys)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}

	if hash == "" {
		t.Error("expected non-empty hash")
	}
}

func TestVerifyIntegrityMatch(t *testing.T) {
	fsys := fstest.MapFS{
		"file.txt": &fstest.MapFile{Data: []byte("hello")},
	}

	hash, err := ComputePackageIntegrity(fsys)
	if err != nil {
		t.Fatalf("compute failed: %v", err)
	}

	ok, err := VerifyIntegrity(fsys, hash)
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	if !ok {
		t.Error("expected integrity to match")
	}
}

func TestVerifyIntegrityMismatch(t *testing.T) {
	fsys := fstest.MapFS{
		"file.txt": &fstest.MapFile{Data: []byte("hello")},
	}

	ok, err := VerifyIntegrity(fsys, "sha256:0000000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	if ok {
		t.Error("expected integrity mismatch")
	}
}

func TestVerifyIntegrityEmptyRecorded(t *testing.T) {
	fsys := fstest.MapFS{
		"file.txt": &fstest.MapFile{Data: []byte("hello")},
	}

	ok, err := VerifyIntegrity(fsys, "")
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	if ok {
		t.Error("expected false for empty recorded hash")
	}
}

// errorFS is a test helper that returns an error when walking.
type errorFS struct{}

func (errorFS) Open(name string) (fs.File, error) {
	return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
}

func TestComputePackageIntegrityPropagatesWalkError(t *testing.T) {
	_, err := ComputePackageIntegrity(errorFS{})
	if err == nil {
		t.Error("expected error from broken FS")
	}
}
