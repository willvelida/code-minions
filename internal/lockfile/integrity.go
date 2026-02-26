package lockfile

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

const checksumPrefix = "sha256:"

// ComputePackageIntegrity computes a deterministic content hash for a
// package's file tree. The hash captures both file paths and contents
// so that renames and content changes are detected.
//
// Algorithm:
//  1. Walk the fs.FS and collect all file paths
//  2. Sort paths lexicographically (forward slashes, regardless of OS)
//  3. For each file: hash = sha256(path + "\x00" + content)
//  4. Concatenate all individual hashes in sorted order
//  5. Final hash = sha256(concatenated hashes)
//  6. Return as "sha256:<hex>"
//
// The null byte separator prevents a file named "a" with content "b\x00c"
// from colliding with a file named "a\x00b" with content "c".
func ComputePackageIntegrity(content fs.FS) (string, error) {
	var paths []string

	err := fs.WalkDir(content, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		// Normalise to forward slashes for cross-platform determinism.
		paths = append(paths, filepath.ToSlash(path))
		return nil
	})
	if err != nil {
		return "", err
	}

	sort.Strings(paths)

	// Hash each file individually, then combine.
	var fileHashes []byte
	for _, path := range paths {
		h, err := hashFile(content, path)
		if err != nil {
			return "", err
		}
		fileHashes = append(fileHashes, h...)
	}

	// Empty package (no files) gets a hash of the empty input.
	final := sha256.Sum256(fileHashes)
	return checksumPrefix + hex.EncodeToString(final[:]), nil
}

// hashFile computes sha256(path + "\x00" + content) for a single file.
func hashFile(fsys fs.FS, path string) ([]byte, error) {
	f, err := fsys.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()

	h := sha256.New()
	h.Write([]byte(path))
	h.Write([]byte{0x00})

	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

// VerifyIntegrity checks that a package's current content matches
// the recorded integrity hash. Returns true if the hashes match.
func VerifyIntegrity(content fs.FS, recorded string) (bool, error) {
	if recorded == "" {
		return false, nil
	}

	// Strip the "sha256:" prefix for comparison.
	recorded = strings.TrimPrefix(recorded, checksumPrefix)

	current, err := ComputePackageIntegrity(content)
	if err != nil {
		return false, err
	}

	current = strings.TrimPrefix(current, checksumPrefix)
	return current == recorded, nil
}
