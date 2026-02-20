package cmd

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/willvelida/code-minions/internal/model"
)

// selectPackages displays a numbered list of packages and reads the user's
// selection from stdin. The user can enter:
//   - Comma-separated numbers: "1,3" → packages at indices 0 and 2
//   - "all" → all packages
//   - Empty line (just Enter) → all packages (default)
//
// Returns the selected package names, or an error if the input is invalid.
func selectPackages(stdin io.Reader, stdout io.Writer, packages []model.Package) ([]string, error) {
	if len(packages) == 0 {
		return []string{}, nil
	}

	// Display the numbered list
	if _, err := fmt.Fprintln(stdout, "\nAvailable packages:"); err != nil {
		return nil, err
	}
	for i, pkg := range packages {
		desc := pkg.Description
		if desc == "" {
			desc = "(no description)"
		}
		if _, err := fmt.Fprintf(stdout, "  [%d] %s — %s\n", i+1, pkg.Name, desc); err != nil {
			return nil, err
		}
	}

	if _, err := fmt.Fprint(stdout, "\nEnter package numbers (comma-separated), or 'all' [default: all]: "); err != nil {
		return nil, err
	}

	// Read the user's input
	reader := bufio.NewReader(stdin)
	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		// EOF with no input — use default (all)
		return allPackageNames(packages), nil
	}

	input := strings.TrimSpace(line)

	// Empty input or "all" → select everything
	if input == "" || strings.EqualFold(input, "all") {
		return allPackageNames(packages), nil
	}

	// Parse comma-separated numbers
	return parsePackageSelection(input, packages)
}

// parsePackageSelection parses a comma-separated string of 1-based indices
// and returns the corresponding package names.
func parsePackageSelection(input string, packages []model.Package) ([]string, error) {
	parts := strings.Split(input, ",")
	seen := make(map[int]bool)
	var selected []string

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}

		num, err := strconv.Atoi(trimmed)
		if err != nil {
			return nil, fmt.Errorf("invalid selection %q: expected a number", trimmed)
		}

		if num < 1 || num > len(packages) {
			return nil, fmt.Errorf("invalid selection %d: must be between 1 and %d", num, len(packages))
		}

		// Deduplicate: ignore repeated numbers
		idx := num - 1
		if !seen[idx] {
			seen[idx] = true
			selected = append(selected, packages[idx].Name)
		}
	}

	if len(selected) == 0 {
		return nil, fmt.Errorf("no packages selected")
	}

	return selected, nil
}

// allPackageNames returns the names of all packages in order.
func allPackageNames(packages []model.Package) []string {
	names := make([]string, len(packages))
	for i, pkg := range packages {
		names[i] = pkg.Name
	}
	return names
}

// selectAssistant displays detected assistants and reads the user's choice.
// If detected is non-empty, the first one is shown as the default.
// Falls back to "copilot" if nothing is detected and the user presses Enter.
func selectAssistant(stdin io.Reader, stdout io.Writer, detected []string, allAssistants []string) (string, error) {
	defaultChoice := "copilot"
	if len(detected) > 0 {
		defaultChoice = detected[0]
	}

	// Show detected assistants
	if len(detected) > 0 {
		if _, err := fmt.Fprintf(stdout, "\nDetected assistants: %s\n", strings.Join(detected, ", ")); err != nil {
			return "", err
		}
	}

	// Show all available assistants
	if _, err := fmt.Fprintf(stdout, "Available assistants: %s\n", strings.Join(allAssistants, ", ")); err != nil {
		return "", err
	}

	if _, err := fmt.Fprintf(stdout, "Choose assistant [default: %s]: ", defaultChoice); err != nil {
		return "", err
	}

	reader := bufio.NewReader(stdin)
	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		return defaultChoice, nil
	}

	input := strings.TrimSpace(line)
	if input == "" {
		return defaultChoice, nil
	}

	return input, nil
}

// selectTemplate displays a numbered list of templates plus a "blank" option
// and reads the user's choice. Returns the selected Template or nil for blank
// (choose packages manually).
func selectTemplate(stdin io.Reader, stdout io.Writer, templates []Template) (*Template, error) {
	if len(templates) == 0 {
		return nil, nil
	}

	// Find index of the default template ("standard")
	defaultIdx := 0 // fallback to first item
	for i, t := range templates {
		if t.Name == defaultTemplateName {
			defaultIdx = i
			break
		}
	}

	// Display the numbered list
	if _, err := fmt.Fprintln(stdout, "\nChoose a starting template:"); err != nil {
		return nil, err
	}
	for i, t := range templates {
		if _, err := fmt.Fprintf(stdout, "  [%d] %-12s — %s\n", i+1, t.Name, t.Description); err != nil {
			return nil, err
		}
	}
	// Add "blank" as the last option
	blankIdx := len(templates) + 1
	if _, err := fmt.Fprintf(stdout, "  [%d] %-12s — %s\n", blankIdx, "blank", "Choose packages manually"); err != nil {
		return nil, err
	}

	if _, err := fmt.Fprintf(stdout, "\nEnter template number [default: %d]: ", defaultIdx+1); err != nil {
		return nil, err
	}

	// Read the user's input
	reader := bufio.NewReader(stdin)
	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		// EOF with no input — use default
		return &templates[defaultIdx], nil
	}

	input := strings.TrimSpace(line)

	// Empty input → use default
	if input == "" {
		return &templates[defaultIdx], nil
	}

	num, err := strconv.Atoi(input)
	if err != nil {
		return nil, fmt.Errorf("invalid selection %q: expected a number", input)
	}

	if num < 1 || num > blankIdx {
		return nil, fmt.Errorf("invalid selection %d: must be between 1 and %d", num, blankIdx)
	}

	// "blank" option
	if num == blankIdx {
		return nil, nil
	}

	return &templates[num-1], nil
}
