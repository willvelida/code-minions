package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
)

// isInteractiveFunc checks whether stdin and stdout are interactive terminals.
// It is a package-level variable so tests can override it.
var isInteractiveFunc = func() bool {
	inFd := os.Stdin.Fd()
	outFd := os.Stdout.Fd()

	stdinIsTTY := isatty.IsTerminal(inFd) || isatty.IsCygwinTerminal(inFd)
	stdoutIsTTY := isatty.IsTerminal(outFd) || isatty.IsCygwinTerminal(outFd)

	return stdinIsTTY && stdoutIsTTY
}

// confirmPrompt writes message to stdout and reads a single line from stdin.
// Returns true only if the user types "y" or "yes" (case-insensitive).
// Any other input (including empty/Enter) returns false.
func confirmPrompt(stdin io.Reader, stdout io.Writer, message string) (bool, error) {
	if _, err := fmt.Fprint(stdout, message); err != nil {
		return false, fmt.Errorf("failed to write confirmation prompt: %w", err)
	}

	scanner := bufio.NewScanner(stdin)
	if scanner.Scan() {
		answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
		return answer == "y" || answer == "yes", nil
	}

	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("failed to read confirmation: %w", err)
	}

	// EOF without any input — treat as "no"
	return false, nil
}
