package installer

import (
	"bytes"
	"fmt"
	"strings"
)

// InstructionTranslator converts generic .instructions.md files into the
// format expected by a specific coding assistant.
//
// Most assistants accept the generic format as-is (pass-through). Cursor
// is the notable exception — it uses .mdc files with `globs` frontmatter
// instead of `applyTo`.
type InstructionTranslator interface {
	// TranslateContent converts the file content and returns the translated
	// content along with the target filename. For most assistants, these
	// are unchanged. For Cursor, the file is renamed to .mdc and the
	// frontmatter is converted.
	TranslateContent(content []byte, filename string) ([]byte, string, error)
}

// NewInstructionTranslator returns a translator for the named assistant.
// Assistants that accept the generic .instructions.md format natively use
// PassthroughInstructionTranslator to return content unchanged.
func NewInstructionTranslator(assistant string) InstructionTranslator {
	switch assistant {
	case "cursor":
		return &CursorInstructionTranslator{}
	default:
		// All other assistants accept .instructions.md as-is.
		return &PassthroughInstructionTranslator{}
	}
}

// PassthroughInstructionTranslator returns content and filename unchanged.
type PassthroughInstructionTranslator struct{}

func (t *PassthroughInstructionTranslator) TranslateContent(content []byte, filename string) ([]byte, string, error) {
	return content, filename, nil
}

// CursorInstructionTranslator converts .instructions.md files to .mdc
// format with Cursor-native frontmatter.
//
// Translation rules:
//   - applyTo → globs (Cursor's native glob key)
//   - Add alwaysApply: false (since we're targeting specific file patterns)
//   - Rename file from .instructions.md to .mdc
type CursorInstructionTranslator struct{}

func (t *CursorInstructionTranslator) TranslateContent(content []byte, filename string) ([]byte, string, error) {
	// Only transform .instructions.md files; pass everything else through
	// unchanged so that agent/skill files aren't accidentally renamed.
	if !strings.HasSuffix(filename, ".instructions.md") {
		return content, filename, nil
	}

	newName := toCursorFilename(filename)

	// When content is nil (filename-only transform during uninstall),
	// skip the frontmatter translation entirely.
	if content == nil {
		return nil, newName, nil
	}

	translated, err := translateFrontmatter(content)
	if err != nil {
		return nil, "", fmt.Errorf("translating %s for Cursor: %w", filename, err)
	}

	return translated, newName, nil
}

// toCursorFilename converts a generic instruction filename to a Cursor .mdc filename.
// "security.instructions.md" → "security.mdc"
//
// Callers must ensure the name ends in ".instructions.md" before calling;
// non-instruction files are guarded upstream in TranslateContent.
func toCursorFilename(name string) string {
	return strings.TrimSuffix(name, ".instructions.md") + ".mdc"
}

// translateFrontmatter converts generic instruction frontmatter to Cursor format.
//
// Input:
//
//	---
//	description: Security standards
//	applyTo: '**/*.{tf,bicep}'
//	---
//
// Output:
//
//	---
//	description: Security standards
//	globs: '**/*.{tf,bicep}'
//	alwaysApply: false
//	---
func translateFrontmatter(content []byte) ([]byte, error) {
	// Split into frontmatter and body
	frontmatter, body, hasFrontmatter := splitFrontmatter(content)
	if !hasFrontmatter {
		// No frontmatter — wrap in Cursor frontmatter with alwaysApply: true
		// (since there's no pattern, it should always apply)
		var buf bytes.Buffer
		buf.WriteString("---\nalwaysApply: true\n---\n")
		buf.Write(content)
		return buf.Bytes(), nil
	}

	// Convert frontmatter lines
	var newLines []string
	hasAlwaysApply := false
	hasGlobs := false

	for _, line := range strings.Split(frontmatter, "\n") {
		trimmed := strings.TrimSpace(line)

		// Convert applyTo → globs
		if strings.HasPrefix(trimmed, "applyTo:") {
			value := strings.TrimSpace(strings.TrimPrefix(trimmed, "applyTo:"))
			newLines = append(newLines, "globs: "+value)
			hasGlobs = true
			continue
		}

		// Pass through other keys
		if strings.HasPrefix(trimmed, "alwaysApply:") {
			hasAlwaysApply = true
		}
		newLines = append(newLines, line)
	}

	// Add alwaysApply: false if we have globs and it wasn't already set
	if hasGlobs && !hasAlwaysApply {
		newLines = append(newLines, "alwaysApply: false")
	}
	// If no globs (no applyTo), add alwaysApply: true
	if !hasGlobs && !hasAlwaysApply {
		newLines = append(newLines, "alwaysApply: true")
	}

	// Reconstruct
	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.WriteString(strings.Join(newLines, "\n"))
	buf.WriteString("\n---\n")
	buf.Write(body)
	return buf.Bytes(), nil
}

// splitFrontmatter separates YAML frontmatter from body content.
// Returns the frontmatter content (without delimiters), the body,
// and whether frontmatter was found.
func splitFrontmatter(content []byte) (string, []byte, bool) {
	s := string(content)

	// Normalise line endings so downstream splitting works uniformly.
	s = strings.ReplaceAll(s, "\r\n", "\n")

	// Must start with ---
	if !strings.HasPrefix(strings.TrimSpace(s), "---") {
		return "", content, false
	}

	// Find the closing ---
	trimmed := strings.TrimSpace(s)
	rest := trimmed[3:] // skip opening ---
	rest = strings.TrimPrefix(rest, "\n")

	closingIdx := strings.Index(rest, "\n---")
	if closingIdx < 0 {
		return "", content, false
	}

	frontmatter := rest[:closingIdx]
	body := rest[closingIdx+4:] // skip \n---
	// Trim leading newline from body
	body = strings.TrimPrefix(body, "\n")

	return frontmatter, []byte(body), true
}
