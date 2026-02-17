package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/willvelida/code-minions/internal/assistant"
)

// TransferOptions configures a transfer between two assistant layouts.
type TransferOptions struct {
	FromAssistant string // Source assistant name
	ToAssistant   string // Target assistant name
	TargetDir     string // Working directory (default ".")
	Force         bool   // Overwrite existing files
	DryRun        bool   // Preview without writing
	Cleanup       bool   // Delete source files after successful copy
}

// TransferResult tracks what happened during a transfer.
type TransferResult struct {
	Copied   []string // Files successfully copied (destination paths)
	Skipped  []string // Files skipped (already exist at destination)
	Errors   []string // Error messages
	Cleaned  []string // Source files deleted (when --cleanup is used)
	Warnings []string // Non-fatal warnings (e.g. missing source dirs)
}

// transferSkipFiles lists filenames that should not be transferred.
// AGENTS.md is excluded because it is regenerated for the target layout.
var transferSkipFiles = map[string]bool{
	"package.yaml": true,
	"mcp.yaml":     true,
	"AGENTS.md":    true,
}

// Transfer copies agent and skill files from one assistant's layout to another.
//
// Flow:
//  1. Look up source and target assistant configs
//  2. Scan source assistant's agent and skill directories
//  3. For each file: reverse-map to generic path, forward-map to target path
//  4. Copy file to target (respecting Force and DryRun)
//  5. Optionally clean up source files (when Cleanup is true)
//  6. Return summary
func Transfer(opts TransferOptions) (*TransferResult, error) {
	result := &TransferResult{}

	if opts.FromAssistant == opts.ToAssistant {
		return nil, fmt.Errorf("source and target assistant cannot be the same (%q)", opts.FromAssistant)
	}

	fromCfg, err := assistant.Get(opts.FromAssistant)
	if err != nil {
		return nil, err
	}
	toCfg, err := assistant.Get(opts.ToAssistant)
	if err != nil {
		return nil, err
	}

	if opts.TargetDir == "" {
		opts.TargetDir = "."
	}

	absTarget, err := filepath.Abs(opts.TargetDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve target directory: %w", err)
	}

	reverseMapper := fromCfg.NewReversePathMapper()
	forwardMapper := toCfg.NewPathMapper()

	// Track source files that were successfully copied (for cleanup)
	var copiedSourcePaths []string
	var sourceDirPaths []string

	// Scan source directories
	sourceDirs := []string{fromCfg.AgentDir, fromCfg.SkillDir}
	agentDirExists := false
	skillDirExists := false

	for i, dir := range sourceDirs {
		sourceDir := filepath.Join(opts.TargetDir, dir)
		info, err := os.Stat(sourceDir)
		if os.IsNotExist(err) {
			label := "agent"
			if i == 1 {
				label = "skill"
			}
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("%s directory %q does not exist for %s", label, dir, fromCfg.Description))
			continue
		}
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to stat %s: %v", dir, err))
			continue
		}
		if !info.IsDir() {
			result.Errors = append(result.Errors, fmt.Sprintf("%s is not a directory", dir))
			continue
		}

		if i == 0 {
			agentDirExists = true
		} else {
			skillDirExists = true
		}

		// Walk the source directory
		err = filepath.WalkDir(sourceDir, func(absPath string, d os.DirEntry, walkErr error) error {
			if walkErr != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("error reading %s: %v", absPath, walkErr))
				return nil
			}

			// Compute the relative path from the target directory
			relPath, err := filepath.Rel(opts.TargetDir, absPath)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to compute relative path for %s: %v", absPath, err))
				return nil
			}

			// Normalise to forward slashes for consistent path mapping
			relPath = filepath.ToSlash(relPath)

			if d.IsDir() {
				sourceDirPaths = append(sourceDirPaths, absPath)
				return nil
			}

			// Skip metadata and AGENTS.md files
			if transferSkipFiles[d.Name()] {
				return nil
			}

			// Reverse-map: assistant-specific → generic
			genericPath := reverseMapper(relPath)

			// Forward-map: generic → target assistant-specific
			destRelPath := forwardMapper(genericPath)

			destPath := filepath.Join(opts.TargetDir, filepath.FromSlash(destRelPath))

			// Path traversal protection
			absDest, err := filepath.Abs(destPath)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to resolve path %s: %v", destRelPath, err))
				return nil
			}
			rel, err := filepath.Rel(absTarget, absDest)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to compute relative path for %s: %v", destRelPath, err))
				return nil
			}
			if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
				result.Errors = append(result.Errors, fmt.Sprintf("path %s escapes target directory", destRelPath))
				return nil
			}

			// Check if destination already exists
			if !opts.Force {
				if _, err := os.Stat(destPath); err == nil {
					result.Skipped = append(result.Skipped, destRelPath)
					return nil
				} else if !os.IsNotExist(err) {
					result.Errors = append(result.Errors, fmt.Sprintf("failed to stat %s: %v", destPath, err))
					return nil
				}
			}

			if opts.DryRun {
				result.Copied = append(result.Copied, destRelPath)
				copiedSourcePaths = append(copiedSourcePaths, absPath)
				return nil
			}

			// Read source file
			data, err := os.ReadFile(absPath)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to read %s: %v", relPath, err))
				return nil
			}

			// Create destination directory
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to create directory for %s: %v", destPath, err))
				return nil
			}

			// Write to destination
			if err := os.WriteFile(destPath, data, 0644); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to write %s: %v", destRelPath, err))
			} else {
				result.Copied = append(result.Copied, destRelPath)
				copiedSourcePaths = append(copiedSourcePaths, absPath)
			}

			return nil
		})

		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to walk directory %s: %v", dir, err))
		}
	}

	// If neither source directory exists, return an error
	if !agentDirExists && !skillDirExists {
		return nil, fmt.Errorf("no agent or skill files found for %s — neither %q nor %q exists",
			fromCfg.Description, fromCfg.AgentDir, fromCfg.SkillDir)
	}

	// Warn if source dirs existed but contained no files
	if agentDirExists || skillDirExists {
		if len(result.Copied) == 0 && len(result.Skipped) == 0 && len(result.Errors) == 0 {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("source directories for %s exist but contain no transferable files", fromCfg.Description))
		}
	}

	// Cleanup source files if requested
	if opts.Cleanup && len(copiedSourcePaths) > 0 {
		if opts.DryRun {
			// In dry-run mode, report what would be cleaned up
			for _, p := range copiedSourcePaths {
				relPath, _ := filepath.Rel(opts.TargetDir, p)
				result.Cleaned = append(result.Cleaned, filepath.ToSlash(relPath))
			}
			// Also report AGENTS.md cleanup in dry-run
			for _, dir := range sourceDirs {
				agentsMD := filepath.Join(opts.TargetDir, dir, "AGENTS.md")
				if _, err := os.Stat(agentsMD); err == nil {
					relPath, _ := filepath.Rel(opts.TargetDir, agentsMD)
					result.Cleaned = append(result.Cleaned, filepath.ToSlash(relPath))
				}
			}
		} else {
			// Delete successfully copied source files
			for _, p := range copiedSourcePaths {
				if err := os.Remove(p); err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("cleanup: failed to remove %s: %v", p, err))
				} else {
					relPath, _ := filepath.Rel(opts.TargetDir, p)
					result.Cleaned = append(result.Cleaned, filepath.ToSlash(relPath))
				}
			}

			// Also remove AGENTS.md from source dirs during cleanup
			for _, dir := range sourceDirs {
				agentsMD := filepath.Join(opts.TargetDir, dir, "AGENTS.md")
				if _, err := os.Stat(agentsMD); err == nil {
					if err := os.Remove(agentsMD); err != nil {
						result.Errors = append(result.Errors, fmt.Sprintf("cleanup: failed to remove %s: %v", agentsMD, err))
					} else {
						relPath, _ := filepath.Rel(opts.TargetDir, agentsMD)
						result.Cleaned = append(result.Cleaned, filepath.ToSlash(relPath))
					}
				}
			}

			// Clean up empty source directories (deepest first)
			cleanTransferDirs(sourceDirPaths)
		}
	}

	return result, nil
}

// cleanTransferDirs removes empty directories from deepest to shallowest.
// Note: the original cleanEmptyDirs in uninstall.go returns []string;
// this version is used internally by Transfer which tracks cleaned files separately.
func cleanTransferDirs(dirs []string) {
	sort.Slice(dirs, func(i, j int) bool {
		return strings.Count(dirs[i], string(os.PathSeparator)) >
			strings.Count(dirs[j], string(os.PathSeparator))
	})

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		if len(entries) == 0 {
			_ = os.Remove(dir)
		}
	}
}
