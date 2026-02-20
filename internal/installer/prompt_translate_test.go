package installer

import (
	"strings"
	"testing"
)

// samplePromptContent is a canonical .prompt.md file used by all translator tests.
var samplePromptContent = []byte(`---
description: "Perform a STRIDE threat assessment on the codebase"
mode: agent
input:
  - name: scope
    description: "Focus area for the assessment"
  - name: threat_level
    description: "Minimum severity to report"
---

# Threat Assessment

Focus area: ${input:scope}

Minimum severity: ${input:threat_level}
`)

// sampleSingleInputPrompt has just one input for testing $ARGUMENTS vs positional.
var sampleSingleInputPrompt = []byte(`---
description: "Explain a software concept"
mode: ask
input:
  - name: concept
    description: "The concept to explain"
---

Explain ${input:concept} in detail.
`)

// --- Factory ---

func TestNewPromptTranslatorReturnsCorrectType(t *testing.T) {
	tests := []struct {
		assistant string
		wantType  string
	}{
		{"copilot", "*installer.PassthroughPromptTranslator"},
		{"claude", "*installer.ClaudePromptTranslator"},
		{"opencode", "*installer.OpenCodePromptTranslator"},
		{"gemini", "*installer.GeminiPromptTranslator"},
		{"cursor", "*installer.CursorPromptTranslator"},
		{"codex", "*installer.CodexPromptTranslator"},
		{"unknown", "*installer.PassthroughPromptTranslator"},
	}

	for _, tt := range tests {
		t.Run(tt.assistant, func(t *testing.T) {
			translator := NewPromptTranslator(tt.assistant)
			got := typeName(translator)
			if got != tt.wantType {
				t.Errorf("NewPromptTranslator(%q) type = %s, want %s", tt.assistant, got, tt.wantType)
			}
		})
	}
}

func typeName(v interface{}) string {
	switch v.(type) {
	case *PassthroughPromptTranslator:
		return "*installer.PassthroughPromptTranslator"
	case *ClaudePromptTranslator:
		return "*installer.ClaudePromptTranslator"
	case *OpenCodePromptTranslator:
		return "*installer.OpenCodePromptTranslator"
	case *GeminiPromptTranslator:
		return "*installer.GeminiPromptTranslator"
	case *CursorPromptTranslator:
		return "*installer.CursorPromptTranslator"
	case *CodexPromptTranslator:
		return "*installer.CodexPromptTranslator"
	default:
		return "unknown"
	}
}

// --- Passthrough (Copilot) ---

func TestPassthroughPromptTranslator(t *testing.T) {
	translator := &PassthroughPromptTranslator{}

	content, name, err := translator.TranslateContent(samplePromptContent, "threat-assessment.prompt.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if name != "threat-assessment.prompt.md" {
		t.Errorf("filename = %q, want %q", name, "threat-assessment.prompt.md")
	}

	if string(content) != string(samplePromptContent) {
		t.Error("content should be unchanged")
	}
}

func TestPassthroughSkipsNonPromptFiles(t *testing.T) {
	translator := &PassthroughPromptTranslator{}
	input := []byte("# Agent\nSome content")

	content, name, err := translator.TranslateContent(input, "my-agent.agent.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if name != "my-agent.agent.md" {
		t.Errorf("filename = %q, want %q", name, "my-agent.agent.md")
	}

	if string(content) != string(input) {
		t.Error("non-prompt files should be unchanged")
	}
}

// --- Claude ---

func TestClaudePromptTranslator(t *testing.T) {
	translator := &ClaudePromptTranslator{}

	content, name, err := translator.TranslateContent(samplePromptContent, "threat-assessment.prompt.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// File renamed to .md
	if name != "threat-assessment.md" {
		t.Errorf("filename = %q, want %q", name, "threat-assessment.md")
	}

	result := string(content)

	// description preserved
	if !strings.Contains(result, "description:") {
		t.Error("should preserve description")
	}

	// mode should be dropped
	if strings.Contains(result, "mode:") {
		t.Error("should drop mode field")
	}

	// input array should be converted to named keys
	if strings.Contains(result, "input:") {
		t.Error("should convert input array to named keys")
	}

	// Named keys should be present with quoted descriptions
	if !strings.Contains(result, `scope: "Focus area for the assessment"`) {
		t.Error("should have named key 'scope'")
	}
	if !strings.Contains(result, `threat_level: "Minimum severity to report"`) {
		t.Error("should have named key 'threat_level'")
	}

	// Variables rewritten: ${input:scope} → $scope
	if strings.Contains(result, "${input:scope}") {
		t.Error("should rewrite ${input:scope} to $scope")
	}
	if !strings.Contains(result, "$scope") {
		t.Error("should contain $scope variable")
	}
	if !strings.Contains(result, "$threat_level") {
		t.Error("should contain $threat_level variable")
	}
}

func TestClaudePromptTranslatorNilContent(t *testing.T) {
	translator := &ClaudePromptTranslator{}

	content, name, err := translator.TranslateContent(nil, "test.prompt.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if name != "test.md" {
		t.Errorf("filename = %q, want %q", name, "test.md")
	}
	if content != nil {
		t.Error("content should be nil when input is nil")
	}
}

func TestClaudeSkipsNonPromptFiles(t *testing.T) {
	translator := &ClaudePromptTranslator{}
	input := []byte("# Some file")

	_, name, err := translator.TranslateContent(input, "my-agent.agent.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "my-agent.agent.md" {
		t.Errorf("non-prompt files should keep their name, got %q", name)
	}
}

// --- OpenCode ---

func TestOpenCodePromptTranslatorPositional(t *testing.T) {
	translator := &OpenCodePromptTranslator{}

	content, name, err := translator.TranslateContent(samplePromptContent, "threat-assessment.prompt.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if name != "threat-assessment.md" {
		t.Errorf("filename = %q, want %q", name, "threat-assessment.md")
	}

	result := string(content)

	// mode and input should be dropped from frontmatter
	if strings.Contains(result, "mode:") {
		t.Error("should drop mode field")
	}
	if strings.Contains(result, "input:") {
		t.Error("should drop input array")
	}

	// Variables should be positional (2 inputs → $1, $2)
	if !strings.Contains(result, "$1") {
		t.Error("should contain $1 for first input")
	}
	if !strings.Contains(result, "$2") {
		t.Error("should contain $2 for second input")
	}
	if strings.Contains(result, "${input:scope}") {
		t.Error("should not contain original ${input:scope} syntax")
	}
}

func TestOpenCodePromptTranslatorSingleInput(t *testing.T) {
	translator := &OpenCodePromptTranslator{}

	content, name, err := translator.TranslateContent(sampleSingleInputPrompt, "explain-concept.prompt.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if name != "explain-concept.md" {
		t.Errorf("filename = %q, want %q", name, "explain-concept.md")
	}

	result := string(content)

	// Single input should use $1 (positional since ≤3 inputs)
	if !strings.Contains(result, "$1") {
		t.Error("single input should use $1")
	}
}

func TestOpenCodePromptTranslatorNilContent(t *testing.T) {
	translator := &OpenCodePromptTranslator{}

	content, name, err := translator.TranslateContent(nil, "test.prompt.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if name != "test.md" {
		t.Errorf("filename = %q, want %q", name, "test.md")
	}
	if content != nil {
		t.Error("content should be nil when input is nil")
	}
}

// --- Gemini ---

func TestGeminiPromptTranslator(t *testing.T) {
	translator := &GeminiPromptTranslator{}

	content, name, err := translator.TranslateContent(samplePromptContent, "threat-assessment.prompt.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Extension changed to .toml
	if name != "threat-assessment.toml" {
		t.Errorf("filename = %q, want %q", name, "threat-assessment.toml")
	}

	result := string(content)

	// Should be TOML format
	if !strings.Contains(result, "prompt = \"\"\"") {
		t.Error("should contain TOML prompt key with triple-quoted string")
	}
	if !strings.Contains(result, "\"\"\"") {
		t.Error("should have triple-quote closing")
	}

	// Description should be TOML string
	if !strings.Contains(result, `description = "`) {
		t.Error("should have TOML description key")
	}

	// Variables should be {{args}}
	if !strings.Contains(result, "{{args}}") {
		t.Error("should rewrite variables to {{args}}")
	}
	if strings.Contains(result, "${input:") {
		t.Error("should not contain original ${input:} syntax")
	}

	// Should NOT contain YAML frontmatter delimiters
	if strings.Contains(result, "---") {
		t.Error("should not contain YAML frontmatter delimiters in TOML output")
	}
}

func TestGeminiPromptTranslatorNilContent(t *testing.T) {
	translator := &GeminiPromptTranslator{}

	content, name, err := translator.TranslateContent(nil, "test.prompt.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if name != "test.toml" {
		t.Errorf("filename = %q, want %q", name, "test.toml")
	}
	if content != nil {
		t.Error("content should be nil when input is nil")
	}
}

// --- Cursor ---

func TestCursorPromptTranslator(t *testing.T) {
	translator := &CursorPromptTranslator{}

	content, name, err := translator.TranslateContent(samplePromptContent, "threat-assessment.prompt.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Extension changed to .mdc
	if name != "threat-assessment.mdc" {
		t.Errorf("filename = %q, want %q", name, "threat-assessment.mdc")
	}

	result := string(content)

	// Should have alwaysApply: false
	if !strings.Contains(result, "alwaysApply: false") {
		t.Error("should have alwaysApply: false")
	}

	// Should keep description
	if !strings.Contains(result, "description:") {
		t.Error("should preserve description")
	}

	// Variables should be stripped to [placeholder]
	if strings.Contains(result, "${input:") {
		t.Error("should strip ${input:} variables")
	}
	if !strings.Contains(result, "[scope]") {
		t.Error("should replace variables with [placeholder], expected [scope]")
	}
	if !strings.Contains(result, "[threat_level]") {
		t.Error("should replace variables with [placeholder], expected [threat_level]")
	}

	// Should NOT have mode or input
	if strings.Contains(result, "mode:") {
		t.Error("should not contain mode")
	}
	if strings.Contains(result, "input:") {
		t.Error("should not contain input")
	}
}

func TestCursorPromptTranslatorNilContent(t *testing.T) {
	translator := &CursorPromptTranslator{}

	content, name, err := translator.TranslateContent(nil, "test.prompt.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if name != "test.mdc" {
		t.Errorf("filename = %q, want %q", name, "test.mdc")
	}
	if content != nil {
		t.Error("content should be nil when input is nil")
	}
}

// --- Codex ---

func TestCodexPromptTranslator(t *testing.T) {
	translator := &CodexPromptTranslator{}

	content, name, err := translator.TranslateContent(samplePromptContent, "threat-assessment.prompt.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Extension changed to .md
	if name != "threat-assessment.md" {
		t.Errorf("filename = %q, want %q", name, "threat-assessment.md")
	}

	result := string(content)

	// Should keep description
	if !strings.Contains(result, "description:") {
		t.Error("should preserve description")
	}

	// Variables should be stripped to [placeholder]
	if strings.Contains(result, "${input:") {
		t.Error("should strip ${input:} variables")
	}
	if !strings.Contains(result, "[scope]") {
		t.Error("should replace variables with [placeholder]")
	}

	// Should NOT have mode or input
	if strings.Contains(result, "mode:") {
		t.Error("should not contain mode")
	}
	if strings.Contains(result, "input:") {
		t.Error("should not contain input")
	}
}

func TestCodexPromptTranslatorNilContent(t *testing.T) {
	translator := &CodexPromptTranslator{}

	content, name, err := translator.TranslateContent(nil, "test.prompt.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if name != "test.md" {
		t.Errorf("filename = %q, want %q", name, "test.md")
	}
	if content != nil {
		t.Error("content should be nil when input is nil")
	}
}

// --- Helper: toPromptFilename ---

func TestToPromptFilename(t *testing.T) {
	tests := []struct {
		input    string
		ext      string
		expected string
	}{
		{"threat-assessment.prompt.md", ".md", "threat-assessment.md"},
		{"threat-assessment.prompt.md", ".toml", "threat-assessment.toml"},
		{"threat-assessment.prompt.md", ".mdc", "threat-assessment.mdc"},
		{"create-skill.prompt.md", ".md", "create-skill.md"},
	}

	for _, tt := range tests {
		got := toPromptFilename(tt.input, tt.ext)
		if got != tt.expected {
			t.Errorf("toPromptFilename(%q, %q) = %q, want %q", tt.input, tt.ext, got, tt.expected)
		}
	}
}

// --- Helper: extractInputNames ---

func TestExtractInputNames(t *testing.T) {
	fm := `description: "Test"
mode: agent
input:
  - name: scope
    description: "Focus area"
  - name: threat_level
    description: "Severity"`

	names := extractInputNames(fm)
	if len(names) != 2 {
		t.Fatalf("got %d names, want 2", len(names))
	}
	if names[0] != "scope" {
		t.Errorf("names[0] = %q, want %q", names[0], "scope")
	}
	if names[1] != "threat_level" {
		t.Errorf("names[1] = %q, want %q", names[1], "threat_level")
	}
}

func TestExtractInputNamesEmpty(t *testing.T) {
	fm := `description: "Test"
mode: agent`

	names := extractInputNames(fm)
	if len(names) != 0 {
		t.Errorf("got %d names, want 0", len(names))
	}
}

// --- No frontmatter ---

func TestClaudePromptNoFrontmatter(t *testing.T) {
	translator := &ClaudePromptTranslator{}
	input := []byte("# Just a body\n\n${input:foo} content here\n")

	content, name, err := translator.TranslateContent(input, "test.prompt.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if name != "test.md" {
		t.Errorf("filename = %q, want %q", name, "test.md")
	}

	result := string(content)
	// Variables should still be rewritten
	if !strings.Contains(result, "$foo") {
		t.Error("should rewrite variables even without frontmatter")
	}
}

func TestGeminiPromptNoFrontmatter(t *testing.T) {
	translator := &GeminiPromptTranslator{}
	input := []byte("# Just a body\n\n${input:foo} content here\n")

	content, name, err := translator.TranslateContent(input, "test.prompt.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if name != "test.toml" {
		t.Errorf("filename = %q, want %q", name, "test.toml")
	}

	result := string(content)
	if !strings.Contains(result, "prompt = \"\"\"") {
		t.Error("should wrap body in TOML prompt key")
	}
	if !strings.Contains(result, "{{args}}") {
		t.Error("should rewrite variables to {{args}}")
	}
	// No description when no frontmatter
	if strings.Contains(result, "description") {
		t.Error("should not have description without frontmatter")
	}
}
