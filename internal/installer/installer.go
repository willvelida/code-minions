package installer

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// SkipFiles lists filenames that should never be copied to the target
// directory (e.g. package manifests are metadata, not installable content).
var SkipFiles = map[string]bool{
	"package.yaml": true,
	"mcp.yaml":     true,
}

// Installer copies embedded package content to a target directory.
type Installer struct {
	Content         fs.FS
	Target          string
	Force           bool
	DryRun          bool
	StripPrefix     string
	PathMapper      func(path string) string                                      // Optional: transforms output paths (e.g. for assistant-specific layouts)
	FileTransformer func(content []byte, filename string) ([]byte, string, error) // Optional: transforms file content and name (e.g. .instructions.md → .mdc for Cursor)
}

// Result tracks what happened during an installation.
type Result struct {
	Copied  []string
	Skipped []string
	Errors  []string
	Actions []FileAction // Populated only during dry-run with categorised operations
}

// Install walks the embedded filesystem directories and copies their contents
// to the target directory. It respects Force, DryRun, and PathMapper settings.
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

			// Skip metadata files that shouldn't be installed
			if !d.IsDir() && SkipFiles[d.Name()] {
				return nil
			}

			// Apply optional file transformer (e.g. .instructions.md → .mdc for Cursor).
			// This runs before PathMapper so directory remapping still applies
			// to the (possibly renamed) file.
			var fileTransformApplied bool
			if i.FileTransformer != nil && !d.IsDir() {
				// We'll apply the actual content transform during the write phase below.
				// For now, just transform the filename portion of the output path.
				parentDir := ""
				name := outputPath
				if idx := strings.LastIndex(outputPath, "/"); idx >= 0 {
					parentDir = outputPath[:idx+1]
					name = outputPath[idx+1:]
				}
				_, newName, err := i.FileTransformer(nil, name)
				if err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("failed to transform %s: %v", outputPath, err))
					return nil
				}
				if newName != name {
					outputPath = parentDir + newName
					fileTransformApplied = true
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
			if i.DryRun {
				action := FileAction{Path: outputPath}
				_, statErr := os.Stat(targetPath)
				if os.IsNotExist(statErr) {
					// File does not exist — would be created
					action.Kind = ActionCreate
					result.Copied = append(result.Copied, outputPath)
				} else if statErr == nil {
					// File exists — compare content to categorise
					sourceData, readErr := fs.ReadFile(i.Content, path)
					if readErr != nil {
						result.Errors = append(result.Errors, fmt.Sprintf("failed to read %s: %v", path, readErr))
						return nil
					}
					// Apply content transformation to source data for accurate comparison
					if i.FileTransformer != nil && fileTransformApplied {
						origName := d.Name()
						transformedData, _, transformErr := i.FileTransformer(sourceData, origName)
						if transformErr != nil {
							result.Errors = append(result.Errors, fmt.Sprintf("failed to transform %s: %v", path, transformErr))
							return nil
						}
						sourceData = transformedData
					}
					targetData, readErr := os.ReadFile(targetPath)
					if readErr != nil {
						result.Errors = append(result.Errors, fmt.Sprintf("failed to read %s: %v", targetPath, readErr))
						return nil
					}
					if bytes.Equal(sourceData, targetData) {
						action.Kind = ActionUnchanged
						result.Skipped = append(result.Skipped, outputPath)
					} else if i.Force {
						action.Kind = ActionModify
						result.Copied = append(result.Copied, outputPath)
					} else {
						// File differs but --force not set — would be skipped
						action.Kind = ActionSkipped
						action.Annotation = "exists (use --force to overwrite)"
						result.Skipped = append(result.Skipped, outputPath)
					}
				} else {
					result.Errors = append(result.Errors, fmt.Sprintf("failed to stat %s: %v", targetPath, statErr))
					return nil
				}
				result.Actions = append(result.Actions, action)
				return nil
			}

			// Non-dry-run: skip existing files unless --force
			if !i.Force {
				if _, err := os.Stat(targetPath); err == nil {
					result.Skipped = append(result.Skipped, outputPath)
					return nil
				} else if !os.IsNotExist(err) {
					result.Errors = append(result.Errors, fmt.Sprintf("failed to stat %s: %v", targetPath, err))
					return nil
				}
			}

			// Read from embedded FS and write to target
			data, err := fs.ReadFile(i.Content, path)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to read %s: %v", outputPath, err))
				return nil
			}

			// Apply content transformation if a file transformer is set
			// and the filename was transformed (indicating this file type
			// needs content transformation too).
			if i.FileTransformer != nil && fileTransformApplied {
				origName := d.Name()
				transformedData, _, transformErr := i.FileTransformer(data, origName)
				if transformErr != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("failed to transform content of %s: %v", outputPath, transformErr))
					return nil
				}
				data = transformedData
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
