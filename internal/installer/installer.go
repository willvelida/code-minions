package installer

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// Installer copies embedded content to a target directory
type Installer struct {
	Content fs.FS
	Target  string
	Force   bool
	DryRun  bool
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

			targetPath := filepath.Join(i.Target, path)

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
					result.Skipped = append(result.Skipped, path)
					return nil
				}
			}

			if i.DryRun {
				result.Copied = append(result.Copied, path)
				return nil
			}

			// Read from embedded FS and write to target
			data, err := fs.ReadFile(i.Content, path)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to read %s: %v", path, err))
				return nil
			}

			if err := os.WriteFile(targetPath, data, 0644); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to write %s: %v", targetPath, err))
			} else {
				result.Copied = append(result.Copied, path)
			}

			return nil
		})

		if err != nil {
			return result, fmt.Errorf("failed to walk directory %s: %w", dir, err)
		}
	}

	return result, nil
}
