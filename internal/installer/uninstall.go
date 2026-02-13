package installer

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// UninstallResult tracks what happened during uninstallation
type UninstallResult struct {
	Removed     []string
	NotFound    []string
	Errors      []string
	DirsCleaned []string
}

// Uninstall removes previously installed files from the target directory.
// It uses the embedded FS as the source of truth for which files to remove,
// then cleans up any empty directories.
func (i *Installer) Uninstall(dirs []string) (*UninstallResult, error) {
	result := &UninstallResult{}
	var dirPaths []string

	absTarget, err := filepath.Abs(i.Target)
	if err != nil {
		return result, fmt.Errorf("failed to resolve target directory: %w", err)
	}

	for _, dir := range dirs {
		err := fs.WalkDir(i.Content, dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("error reading %s: %v", path, err))
				return nil
			}

			outputPath := path
			if i.StripPrefix != "" {
				if path == i.StripPrefix {
					return nil
				}
				if strings.HasPrefix(path, i.StripPrefix) {
					outputPath = strings.TrimPrefix(path, i.StripPrefix)
					outputPath = strings.TrimPrefix(outputPath, "/")
				}
			}

			if i.PathMapper != nil {
				outputPath = i.PathMapper(outputPath)
			}

			targetPath := filepath.Join(i.Target, outputPath)

			absPath, err := filepath.Abs(targetPath)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to resolve path %s: %v", path, err))
				return nil
			}
			relPath, err := filepath.Rel(absTarget, absPath)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to compute relative path for %s: %v", path, err))
				return nil
			}
			if relPath == ".." || strings.HasPrefix(relPath, ".."+string(os.PathSeparator)) {
				result.Errors = append(result.Errors, fmt.Sprintf("path %s escapes target directory", path))
				return nil
			}

			if d.IsDir() {
				dirPaths = append(dirPaths, targetPath)
				return nil
			}

			if _, err := os.Stat(targetPath); os.IsNotExist(err) {
				result.NotFound = append(result.NotFound, outputPath)
				return nil
			} else if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to stat %s: %v", targetPath, err))
				return nil
			}

			if i.DryRun {
				result.Removed = append(result.Removed, outputPath)
				return nil
			}

			if err := os.Remove(targetPath); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to remove %s: %v", targetPath, err))
			} else {
				result.Removed = append(result.Removed, outputPath)
			}

			return nil
		})

		if err != nil {
			return result, fmt.Errorf("failed to walk directory %s: %w", dir, err)
		}
	}

	// Clean up empty directories (skip in dry-run mode)
	if !i.DryRun {
		result.DirsCleaned = cleanEmptyDirs(dirPaths)
	}

	return result, nil
}

// cleanEmptyDirs removes empty directories from deepest to shallowest.
// Uses os.Remove so non-empty directories are left in place.
func cleanEmptyDirs(dirs []string) []string {
	sort.Slice(dirs, func(i, j int) bool {
		return strings.Count(dirs[i], string(os.PathSeparator)) >
			strings.Count(dirs[j], string(os.PathSeparator))
	})

	var cleaned []string
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		if len(entries) == 0 {
			if err := os.Remove(dir); err == nil {
				cleaned = append(cleaned, dir)
			}
		}
	}

	return cleaned
}
