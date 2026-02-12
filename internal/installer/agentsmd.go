package installer

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// AgentsMDHandler manages the AGENTS.md file during install and uninstall.
// It is handled separately because it is a shared resource across packages.
type AgentsMDHandler struct {
	Target     string
	PathMapper func(string) string
	DryRun     bool
	Stdin      io.Reader // For interactive prompts (os.Stdin in production, bytes.Buffer in tests)
}

const agentsMDFile = "AGENTS.md"

// ShouldSkip returns true if the path is an AGENTS.md file that should
// be handled separately from the normal install/uninstall flow.
func (h *AgentsMDHandler) ShouldSkip(outputPath string) bool {
	return filepath.Base(outputPath) == agentsMDFile
}

// OnInstall creates AGENTS.md if it does not already exist.
// Returns "created" or "skipped".
func (h *AgentsMDHandler) OnInstall(outputPath string, data []byte) (string, error) {
	targetPath := filepath.Join(h.Target, outputPath)

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

// OnUninstall prompts the user before removing AGENTS.md.
// Returns "removed" or "kept".
func (h *AgentsMDHandler) OnUninstall(outputPath string) (string, error) {
	targetPath := filepath.Join(h.Target, outputPath)

	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return "kept", nil
	} else if err != nil {
		return "", fmt.Errorf("failed to check %s: %w", targetPath, err)
	}

	if h.DryRun {
		return "kept", nil
	}

	fmt.Printf("  AGENTS.md exists at %s\n", targetPath)
	fmt.Print("  Do you also want to remove it? (y/N): ")

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

	return "kept", nil
}
