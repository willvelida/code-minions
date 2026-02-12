package installer

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Installer copies embedded content to a target directory
type Installer struct {
	Content     fs.FS
	Target      string
	Force       bool
	DryRun      bool
	StripPrefix string
	PathMapper  func(path string) string // Optional: transforms output paths (e.g. for assistant-specific layouts)
}

// Result tracks what happened during installation
type Result struct {
	Copied  []string
	Skipped []string
	Errors  []string
}

func (i *Installer) Install(dirs []string) (*Result, error) {
	result := &Result{}

	for _, dir := range dirs {
		err := fs.WalkDir(i.Content, dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("error reading %s: %v", path, err))
				return nil
			}

			outputPath := path
			if i.StripPrefix != "" {
				// Root entry is the package dir itself — skip it
				if path == i.StripPrefix {
					return nil
				}
				// For children, strip the prefix + leading slash
				if strings.HasPrefix(path, i.StripPrefix) {
					outputPath = strings.TrimPrefix(path, i.StripPrefix)
					outputPath = strings.TrimPrefix(outputPath, "/")
				}
			}

			// Apply optional path mapping (e.g. agents/ → .github/agents/ for Copilot)
			if i.PathMapper != nil {
				outputPath = i.PathMapper(outputPath)
			}

			targetPath := filepath.Join(i.Target, outputPath)

			// Ensure the resolved path stays within the target directory
			absTarget, err := filepath.Abs(i.Target)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to resolve target directory: %v", err))
				return nil
			}
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

			// Create directories
			if d.IsDir() {
				if !i.DryRun {
					if err := os.MkdirAll(targetPath, 0755); err != nil {
						result.Errors = append(result.Errors, fmt.Sprintf("failed to create directory %s: %v", path, err))
					}
				}
				return nil
			}

			// Handle files
			if !i.Force {
				if _, err := os.Stat(targetPath); err == nil {
					result.Skipped = append(result.Skipped, outputPath)
					return nil
				} else if !os.IsNotExist(err) {
					result.Errors = append(result.Errors, fmt.Sprintf("failed to stat %s: %v", targetPath, err))
					return nil
				}
			}

			if i.DryRun {
				result.Copied = append(result.Copied, outputPath)
				return nil
			}

			// Read from embedded FS and write to target
			data, err := fs.ReadFile(i.Content, path)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to read %s: %v", outputPath, err))
				return nil
			}

			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to create directory for %s: %v", targetPath, err))
				return nil
			}

			if err := os.WriteFile(targetPath, data, 0644); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to write %s: %v", targetPath, err))
			} else {
				result.Copied = append(result.Copied, outputPath)
			}

			return nil
		})

		if err != nil {
			return result, fmt.Errorf("failed to walk directory %s: %w", dir, err)
		}
	}

	return result, nil
}
