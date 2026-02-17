package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// teamMarkerStart returns the opening marker for a team's instructions block.
func teamMarkerStart(teamName string) string {
	return fmt.Sprintf("<!-- code-minions:team:%s:start -->", teamName)
}

// teamMarkerEnd returns the closing marker for a team's instructions block.
func teamMarkerEnd(teamName string) string {
	return fmt.Sprintf("<!-- code-minions:team:%s:end -->", teamName)
}

// InjectTeamInstructions appends (or replaces) a team's instructions block
// in the assistant's instructions file using HTML comment markers.
//
// The block is delimited by:
//
//	<!-- code-minions:team:TEAM_NAME:start -->
//	... instructions ...
//	<!-- code-minions:team:TEAM_NAME:end -->
//
// If the file does not exist, it is created with just the team block.
// If the markers already exist, the block is replaced in place.
// If the markers do not exist, the block is appended at the end.
//
// Returns the absolute path of the file written, or an error.
func InjectTeamInstructions(targetDir, instructionsPath, teamName, instructions string) (string, error) {
	if instructions == "" {
		return "", nil
	}

	absPath := filepath.Join(targetDir, instructionsPath)

	// Read existing content (or start empty)
	var existing string
	if data, err := os.ReadFile(absPath); err == nil {
		existing = string(data)
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to read %s: %w", instructionsPath, err)
	}

	block := formatBlock(teamName, instructions)

	var result string
	startMarker := teamMarkerStart(teamName)
	endMarker := teamMarkerEnd(teamName)

	startIdx := strings.Index(existing, startMarker)
	endIdx := strings.Index(existing, endMarker)

	if startIdx >= 0 && endIdx >= 0 && endIdx > startIdx {
		// Replace existing block
		result = existing[:startIdx] + block + existing[endIdx+len(endMarker):]
	} else {
		// Append
		result = existing
		if result != "" && !strings.HasSuffix(result, "\n") {
			result += "\n"
		}
		if result != "" {
			result += "\n"
		}
		result += block
	}

	// Ensure trailing newline
	if !strings.HasSuffix(result, "\n") {
		result += "\n"
	}

	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create directory for %s: %w", instructionsPath, err)
	}
	if err := os.WriteFile(absPath, []byte(result), 0644); err != nil {
		return "", fmt.Errorf("failed to write %s: %w", instructionsPath, err)
	}

	return absPath, nil
}

// RemoveTeamInstructions removes a team's instructions block from the
// assistant's instructions file. If the resulting file is empty (or
// whitespace only), it is deleted.
//
// Returns true if the block was found and removed, false if not present.
func RemoveTeamInstructions(targetDir, instructionsPath, teamName string) (bool, error) {
	absPath := filepath.Join(targetDir, instructionsPath)

	data, err := os.ReadFile(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to read %s: %w", instructionsPath, err)
	}

	content := string(data)
	startMarker := teamMarkerStart(teamName)
	endMarker := teamMarkerEnd(teamName)

	startIdx := strings.Index(content, startMarker)
	endIdx := strings.Index(content, endMarker)

	if startIdx < 0 || endIdx < 0 || endIdx <= startIdx {
		return false, nil
	}

	// Remove the block including any surrounding blank line
	before := content[:startIdx]
	after := content[endIdx+len(endMarker):]

	// Clean up extra blank lines at the join point
	before = strings.TrimRight(before, "\n")
	after = strings.TrimLeft(after, "\n")

	var result string
	if before != "" && after != "" {
		result = before + "\n\n" + after
	} else {
		result = before + after
	}

	// If file is now empty (or whitespace), delete it
	if strings.TrimSpace(result) == "" {
		if err := os.Remove(absPath); err != nil && !os.IsNotExist(err) {
			return false, fmt.Errorf("failed to remove empty %s: %w", instructionsPath, err)
		}
		return true, nil
	}

	// Ensure trailing newline
	if !strings.HasSuffix(result, "\n") {
		result += "\n"
	}

	if err := os.WriteFile(absPath, []byte(result), 0644); err != nil {
		return false, fmt.Errorf("failed to write %s: %w", instructionsPath, err)
	}

	return true, nil
}

// formatBlock produces the marker-delimited instruction block.
func formatBlock(teamName, instructions string) string {
	// Ensure instructions don't have trailing newline before we add the end marker
	instructions = strings.TrimRight(instructions, "\n")
	return fmt.Sprintf("%s\n%s\n%s",
		teamMarkerStart(teamName),
		instructions,
		teamMarkerEnd(teamName),
	)
}
