package cmd

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// promptTeamName reads a team name from stdin and validates it.
func promptTeamName(stdin io.Reader, stdout io.Writer) (string, error) {
	if _, err := fmt.Fprint(stdout, "Team name: "); err != nil {
		return "", err
	}

	reader := bufio.NewReader(stdin)
	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		return "", fmt.Errorf("failed to read team name")
	}

	name := strings.TrimSpace(line)
	if name == "" {
		return "", fmt.Errorf("team name is required")
	}

	return name, nil
}

// promptTeamDescription reads a team description from stdin.
func promptTeamDescription(stdin io.Reader, stdout io.Writer) (string, error) {
	if _, err := fmt.Fprint(stdout, "Team description: "); err != nil {
		return "", err
	}

	reader := bufio.NewReader(stdin)
	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		return "", fmt.Errorf("failed to read team description")
	}

	desc := strings.TrimSpace(line)
	if desc == "" {
		return "", fmt.Errorf("team description is required")
	}

	return desc, nil
}

// promptPersonaName reads a persona name from stdin.
func promptPersonaName(stdin io.Reader, stdout io.Writer) (string, error) {
	if _, err := fmt.Fprint(stdout, "Persona name: "); err != nil {
		return "", err
	}

	reader := bufio.NewReader(stdin)
	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		return "", fmt.Errorf("failed to read persona name")
	}

	name := strings.TrimSpace(line)
	if name == "" {
		return "", fmt.Errorf("persona name is required")
	}

	return name, nil
}

// promptAssistant reads a default assistant from stdin with a default of "copilot".
func promptAssistant(stdin io.Reader, stdout io.Writer) (string, error) {
	if _, err := fmt.Fprint(stdout, "Default assistant [copilot]: "); err != nil {
		return "", err
	}

	reader := bufio.NewReader(stdin)
	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		return "copilot", nil
	}

	input := strings.TrimSpace(line)
	if input == "" {
		return "copilot", nil
	}

	return input, nil
}
