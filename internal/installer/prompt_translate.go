package installer

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

// PromptTranslator converts generic .prompt.md files into the format
// expected by a specific coding assistant.
//
// Copilot uses .prompt.md as-is (pass-through). Other assistants need
// varying degrees of transformation: variable syntax, frontmatter
// schema, file extension, or even file format (Gemini uses TOML).
type PromptTranslator interface {
	// TranslateContent converts the file content and returns the translated
	// content along with the target filename. For Copilot, these are
	// unchanged. For other assistants, the file may be renamed and the
	// content transformed.
	TranslateContent(content []byte, filename string) ([]byte, string, error)
}

// NewPromptTranslator returns a translator for the named assistant.
func NewPromptTranslator(assistant string) PromptTranslator {
	switch assistant {
	case "copilot":
		return &PassthroughPromptTranslator{}
	case "claude":
		return &ClaudePromptTranslator{}
	case "opencode":
		return &OpenCodePromptTranslator{}
	case "gemini":
		return &GeminiPromptTranslator{}
	case "cursor":
		return &CursorPromptTranslator{}
	case "codex":
		return &CodexPromptTranslator{}
	default:
		return &PassthroughPromptTranslator{}
	}
}

// inputVarRegex matches ${input:var_name} variables in prompt content.
var inputVarRegex = regexp.MustCompile(`\$\{input:([^}]+)\}`)

// ---------------------------------------------------------------------------
// PassthroughPromptTranslator — Copilot (canonical format)
// ---------------------------------------------------------------------------

// PassthroughPromptTranslator returns content and filename unchanged.
// Used for Copilot, which defines the canonical .prompt.md format.
type PassthroughPromptTranslator struct{}

func (t *PassthroughPromptTranslator) TranslateContent(content []byte, filename string) ([]byte, string, error) {
	return content, filename, nil
}

// ---------------------------------------------------------------------------
// ClaudePromptTranslator — .prompt.md → .md with variable rewrite
// ---------------------------------------------------------------------------

// ClaudePromptTranslator converts the canonical .prompt.md format to
// Claude Code's command format:
//   - Rename .prompt.md → .md
//   - Convert input array → named frontmatter keys
//   - Translate ${input:x} → $x
//   - Drop mode field
type ClaudePromptTranslator struct{}

func (t *ClaudePromptTranslator) TranslateContent(content []byte, filename string) ([]byte, string, error) {
	if !strings.HasSuffix(filename, ".prompt.md") {
		return content, filename, nil
	}

	newName := toPromptFilename(filename, ".md")

	if content == nil {
		return nil, newName, nil
	}

	translated, err := translatePromptForClaude(content)
	if err != nil {
		return nil, "", fmt.Errorf("translating %s for Claude: %w", filename, err)
	}

	return translated, newName, nil
}

func translatePromptForClaude(content []byte) ([]byte, error) {
	frontmatter, body, hasFM := splitFrontmatter(content)
	if !hasFM {
		// No frontmatter — just rewrite variables
		return []byte(rewriteVarsClaude(string(body))), nil
	}

	// Parse frontmatter and rebuild for Claude
	var newLines []string
	inInputBlock := false
	currentInputName := ""

	for _, line := range strings.Split(frontmatter, "\n") {
		trimmed := strings.TrimSpace(line)

		// Skip mode field
		if strings.HasPrefix(trimmed, "mode:") {
			continue
		}

		// Detect input array start
		if strings.HasPrefix(trimmed, "input:") {
			inInputBlock = true
			continue
		}

		if inInputBlock {
			// Input array items: "- name: foo" or "  description: bar"
			if strings.HasPrefix(trimmed, "- name:") {
				currentInputName = strings.TrimSpace(strings.TrimPrefix(trimmed, "- name:"))
				continue
			}
			if strings.HasPrefix(trimmed, "name:") {
				currentInputName = strings.TrimSpace(strings.TrimPrefix(trimmed, "name:"))
				continue
			}
			if strings.HasPrefix(trimmed, "description:") && currentInputName != "" {
				desc := strings.TrimSpace(strings.TrimPrefix(trimmed, "description:"))
				desc = strings.Trim(desc, "\"'")
				newLines = append(newLines, currentInputName+": "+desc)
				currentInputName = ""
				continue
			}
			// End of input block — non-indented, non-array line
			if trimmed != "" && !strings.HasPrefix(trimmed, "-") && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
				inInputBlock = false
				// Fall through to process this line normally
			} else {
				continue
			}
		}

		newLines = append(newLines, line)
	}

	// Rebuild
	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.WriteString(strings.Join(newLines, "\n"))
	buf.WriteString("\n---\n")
	buf.WriteString(rewriteVarsClaude(string(body)))
	return buf.Bytes(), nil
}

// rewriteVarsClaude converts ${input:var_name} → $var_name
func rewriteVarsClaude(s string) string {
	return inputVarRegex.ReplaceAllString(s, "$$$1")
}

// ---------------------------------------------------------------------------
// OpenCodePromptTranslator — .prompt.md → .md with positional variables
// ---------------------------------------------------------------------------

// OpenCodePromptTranslator converts the canonical .prompt.md format to
// OpenCode's command format:
//   - Rename .prompt.md → .md
//   - Convert input array → positional $1/$2 or $ARGUMENTS
//   - Keep description, drop mode
type OpenCodePromptTranslator struct{}

func (t *OpenCodePromptTranslator) TranslateContent(content []byte, filename string) ([]byte, string, error) {
	if !strings.HasSuffix(filename, ".prompt.md") {
		return content, filename, nil
	}

	newName := toPromptFilename(filename, ".md")

	if content == nil {
		return nil, newName, nil
	}

	translated, err := translatePromptForOpenCode(content)
	if err != nil {
		return nil, "", fmt.Errorf("translating %s for OpenCode: %w", filename, err)
	}

	return translated, newName, nil
}

func translatePromptForOpenCode(content []byte) ([]byte, error) {
	frontmatter, body, hasFM := splitFrontmatter(content)
	if !hasFM {
		return content, nil
	}

	// Collect input names in order
	inputNames := extractInputNames(frontmatter)

	// Build new frontmatter — keep description, drop mode/input
	var newLines []string
	inInputBlock := false

	for _, line := range strings.Split(frontmatter, "\n") {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "mode:") {
			continue
		}
		if strings.HasPrefix(trimmed, "input:") {
			inInputBlock = true
			continue
		}
		if inInputBlock {
			if trimmed != "" && !strings.HasPrefix(trimmed, "-") && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
				inInputBlock = false
			} else {
				continue
			}
		}
		newLines = append(newLines, line)
	}

	// Rewrite variables in body
	bodyStr := string(body)
	if len(inputNames) <= 3 {
		// Positional: ${input:first} → $1, ${input:second} → $2
		for i, name := range inputNames {
			bodyStr = strings.ReplaceAll(bodyStr, fmt.Sprintf("${input:%s}", name), fmt.Sprintf("$%d", i+1))
		}
	} else {
		// Too many inputs — collapse to $ARGUMENTS
		bodyStr = inputVarRegex.ReplaceAllString(bodyStr, "$ARGUMENTS")
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.WriteString(strings.Join(newLines, "\n"))
	buf.WriteString("\n---\n")
	buf.WriteString(bodyStr)
	return buf.Bytes(), nil
}

// ---------------------------------------------------------------------------
// GeminiPromptTranslator — .prompt.md → .toml format conversion
// ---------------------------------------------------------------------------

// GeminiPromptTranslator converts the canonical .prompt.md format to
// Gemini CLI's TOML command format:
//   - Convert .prompt.md → .toml
//   - Extract body → TOML prompt key (triple-quoted string)
//   - Keep description as TOML string
//   - Translate ${input:x} → {{args}}
//   - Drop mode/input
type GeminiPromptTranslator struct{}

func (t *GeminiPromptTranslator) TranslateContent(content []byte, filename string) ([]byte, string, error) {
	if !strings.HasSuffix(filename, ".prompt.md") {
		return content, filename, nil
	}

	newName := toPromptFilename(filename, ".toml")

	if content == nil {
		return nil, newName, nil
	}

	translated, err := translatePromptForGemini(content)
	if err != nil {
		return nil, "", fmt.Errorf("translating %s for Gemini: %w", filename, err)
	}

	return translated, newName, nil
}

func translatePromptForGemini(content []byte) ([]byte, error) {
	frontmatter, body, hasFM := splitFrontmatter(content)

	var description string
	if hasFM {
		for _, line := range strings.Split(frontmatter, "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "description:") {
				description = strings.TrimSpace(strings.TrimPrefix(trimmed, "description:"))
				description = strings.Trim(description, "\"'")
				break
			}
		}
	}

	// Rewrite variables: ${input:x} → {{args}}
	bodyStr := inputVarRegex.ReplaceAllString(string(body), "{{args}}")

	// Build TOML output
	var buf bytes.Buffer
	if description != "" {
		buf.WriteString(fmt.Sprintf("description = %q\n", description))
	}
	buf.WriteString("prompt = \"\"\"\n")
	buf.WriteString(bodyStr)
	if !strings.HasSuffix(bodyStr, "\n") {
		buf.WriteString("\n")
	}
	buf.WriteString("\"\"\"\n")

	return buf.Bytes(), nil
}

// ---------------------------------------------------------------------------
// CursorPromptTranslator — .prompt.md → .mdc rule file (degraded)
// ---------------------------------------------------------------------------

// CursorPromptTranslator converts the canonical .prompt.md format to
// a Cursor rule file:
//   - Rename .prompt.md → .mdc
//   - Keep description, add alwaysApply: false
//   - Drop mode/input
//   - Strip variables → [placeholder]
type CursorPromptTranslator struct{}

func (t *CursorPromptTranslator) TranslateContent(content []byte, filename string) ([]byte, string, error) {
	if !strings.HasSuffix(filename, ".prompt.md") {
		return content, filename, nil
	}

	newName := toPromptFilename(filename, ".mdc")

	if content == nil {
		return nil, newName, nil
	}

	translated, err := translatePromptForCursor(content)
	if err != nil {
		return nil, "", fmt.Errorf("translating %s for Cursor: %w", filename, err)
	}

	return translated, newName, nil
}

func translatePromptForCursor(content []byte) ([]byte, error) {
	frontmatter, body, hasFM := splitFrontmatter(content)

	var description string
	if hasFM {
		for _, line := range strings.Split(frontmatter, "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "description:") {
				description = strings.TrimSpace(strings.TrimPrefix(trimmed, "description:"))
				description = strings.Trim(description, "\"'")
				break
			}
		}
	}

	// Strip variables → [placeholder_name]
	bodyStr := inputVarRegex.ReplaceAllString(string(body), "[$1]")

	// Build .mdc output
	var buf bytes.Buffer
	buf.WriteString("---\n")
	if description != "" {
		buf.WriteString(fmt.Sprintf("description: %s\n", description))
	}
	buf.WriteString("alwaysApply: false\n")
	buf.WriteString("---\n")
	buf.WriteString(bodyStr)

	return buf.Bytes(), nil
}

// ---------------------------------------------------------------------------
// CodexPromptTranslator — .prompt.md → .md reference doc (degraded)
// ---------------------------------------------------------------------------

// CodexPromptTranslator converts the canonical .prompt.md format to a
// reference document for Codex CLI:
//   - Rename .prompt.md → .md
//   - Keep description in frontmatter
//   - Strip mode/input
//   - Strip variables → [placeholder]
type CodexPromptTranslator struct{}

func (t *CodexPromptTranslator) TranslateContent(content []byte, filename string) ([]byte, string, error) {
	if !strings.HasSuffix(filename, ".prompt.md") {
		return content, filename, nil
	}

	newName := toPromptFilename(filename, ".md")

	if content == nil {
		return nil, newName, nil
	}

	translated, err := translatePromptForCodex(content)
	if err != nil {
		return nil, "", fmt.Errorf("translating %s for Codex: %w", filename, err)
	}

	return translated, newName, nil
}

func translatePromptForCodex(content []byte) ([]byte, error) {
	frontmatter, body, hasFM := splitFrontmatter(content)

	var description string
	if hasFM {
		for _, line := range strings.Split(frontmatter, "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "description:") {
				description = strings.TrimSpace(strings.TrimPrefix(trimmed, "description:"))
				description = strings.Trim(description, "\"'")
				break
			}
		}
	}

	// Strip variables → [placeholder_name]
	bodyStr := inputVarRegex.ReplaceAllString(string(body), "[$1]")

	var buf bytes.Buffer
	if description != "" {
		buf.WriteString("---\n")
		buf.WriteString(fmt.Sprintf("description: %s\n", description))
		buf.WriteString("---\n")
	}
	buf.WriteString(bodyStr)

	return buf.Bytes(), nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// toPromptFilename converts a .prompt.md filename to the target extension.
// "threat-assessment.prompt.md" → "threat-assessment.md" (ext=".md")
// "threat-assessment.prompt.md" → "threat-assessment.toml" (ext=".toml")
// "threat-assessment.prompt.md" → "threat-assessment.mdc" (ext=".mdc")
func toPromptFilename(name, ext string) string {
	return strings.TrimSuffix(name, ".prompt.md") + ext
}

// extractInputNames parses the input array from frontmatter and returns
// the variable names in order.
func extractInputNames(frontmatter string) []string {
	var names []string
	inInputBlock := false

	for _, line := range strings.Split(frontmatter, "\n") {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "input:") {
			inInputBlock = true
			continue
		}
		if inInputBlock {
			if strings.HasPrefix(trimmed, "- name:") {
				name := strings.TrimSpace(strings.TrimPrefix(trimmed, "- name:"))
				name = strings.Trim(name, "\"'")
				names = append(names, name)
				continue
			}
			if trimmed != "" && !strings.HasPrefix(trimmed, "-") && !strings.HasPrefix(trimmed, "description:") && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
				break
			}
		}
	}

	return names
}
