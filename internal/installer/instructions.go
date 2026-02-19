package installer

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// InstructionsFileHandler manages assistant-specific instruction files
// (e.g. CLAUDE.md, .cursor/rules/instructions.mdc) during install and
// uninstall. It follows the same contract as AgentsMDHandler but uses
// a configurable filename from assistant.Config.InstructionsPath.
//
// AGENTS.md is the universal routing file handled by AgentsMDHandler.
// InstructionsFileHandler handles the *assistant-native* instructions
// file — the one the assistant reads directly for project context.
type InstructionsFileHandler struct {
	Target   string    // Root directory of the target project
	DryRun   bool      // Preview without writing
	FileName string    // Relative path, e.g. "CLAUDE.md"
	Stdin    io.Reader // For user input (os.Stdin in production)
	Stdout   io.Writer // For prompts and messages
}

// OnInstall creates the instructions file if it does not already exist.
// Returns "created" or "skipped".
func (h *InstructionsFileHandler) OnInstall(data []byte) (string, error) {
	targetPath := filepath.Join(h.Target, h.FileName)

	if _, err := os.Stat(targetPath); err == nil {
		return "skipped", nil
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to check %s: %w", targetPath, err)
	}

	if h.DryRun {
		return "created", nil
	}

	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create directory for %s: %w", targetPath, err)
	}

	if err := os.WriteFile(targetPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write %s: %w", targetPath, err)
	}

	return "created", nil
}

// OnUninstall prompts the user before removing the instructions file.
// Returns "removed" or "kept".
func (h *InstructionsFileHandler) OnUninstall() (string, error) {
	targetPath := filepath.Join(h.Target, h.FileName)

	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return "kept", nil
	} else if err != nil {
		return "", fmt.Errorf("failed to check %s: %w", targetPath, err)
	}

	if h.DryRun {
		return "kept", nil
	}

	displayName := filepath.Base(h.FileName)
	_, _ = fmt.Fprintf(h.Stdout, "  %s exists at %s\n", displayName, targetPath)
	_, _ = fmt.Fprintf(h.Stdout, "  Do you also want to remove it? (y/N): ")

	scanner := bufio.NewScanner(h.Stdin)
	if scanner.Scan() {
		answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
		if answer == "y" || answer == "yes" {
			if err := os.Remove(targetPath); err != nil {
				return "", fmt.Errorf("failed to remove %s: %w", targetPath, err)
			}
			return "removed", nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return "kept", nil
}
