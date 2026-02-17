package installer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/willvelida/code-minions/internal/model"
)

// checksumPrefix is the prefix for SHA-256 checksums in the manifest.
// We include the algorithm name so that if we ever switch to a different
// hash algorithm (e.g. SHA-3), old manifests are still readable.
const checksumPrefix = "sha256:"

// ComputeChecksum computes the SHA-256 checksum of a byte slice.
//
// Returns the checksum in the format "sha256:<hex>".
// This is the "serial number" we stamp on each installed file so
// we can later detect if the user has modified it.
func ComputeChecksum(data []byte) string {
	hash := sha256.Sum256(data)
	return checksumPrefix + hex.EncodeToString(hash[:])
}

// ComputeFileChecksum reads a file from disk and returns its checksum.
//
// This is used during uninstall/update to compare the current file
// against the checksum recorded at install time. If they don't match,
// the user has modified the file and we should warn before deleting.
func ComputeFileChecksum(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file for checksum: %w", err)
	}
	return ComputeChecksum(data), nil
}

// VerifyChecksum compares a file's current checksum against the
// recorded checksum from the manifest. Returns:
//
//   - true if the file has NOT been modified (checksums match)
//   - false if the file HAS been modified or doesn't exist
//
// This is the core of the "safe uninstall" feature: we only
// auto-delete files that match their original checksum.
func VerifyChecksum(path string, recorded string) bool {
	current, err := ComputeFileChecksum(path)
	if err != nil {
		return false // File missing or unreadable → treat as modified
	}
	return current == recorded
}

// NewInstalledFile creates an InstalledFile entry by reading the file
// from disk and computing its checksum. This is called right after
// writing a file during installation.
//
// Parameters:
//   - basePath: the target directory (e.g. the project root)
//   - relativePath: the file path relative to basePath
//   - fileType: what kind of file (e.g. "persona-agent", "routing-table");
//     pass "" for regular package files
func NewInstalledFile(basePath, relativePath, fileType string) (model.InstalledFile, error) {
	fullPath := filepath.Join(basePath, relativePath)

	checksum, err := ComputeFileChecksum(fullPath)
	if err != nil {
		return model.InstalledFile{}, fmt.Errorf(
			"failed to compute checksum for %q: %w", relativePath, err)
	}

	return model.InstalledFile{
		Path:     relativePath,
		Type:     fileType,
		Checksum: checksum,
	}, nil
}
