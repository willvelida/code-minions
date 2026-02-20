package installer

import (
	"strings"
	"testing"
)

func TestPassthroughTranslator(t *testing.T) {
	translator := NewInstructionTranslator("copilot")
	content := []byte("---\napplyTo: '**/*.py'\n---\n# Python Standards\n")

	out, name, err := translator.TranslateContent(content, "python.instructions.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Content should be unchanged
	if string(out) != string(content) {
		t.Errorf("content should be unchanged\ngot:  %q\nwant: %q", string(out), string(content))
	}

	// Filename should be unchanged
	if name != "python.instructions.md" {
		t.Errorf("filename: got %q, want %q", name, "python.instructions.md")
	}
}

func TestCursorTranslatorConvertsApplyToToGlobs(t *testing.T) {
	translator := NewInstructionTranslator("cursor")
	content := []byte("---\ndescription: Python standards\napplyTo: '**/*.py'\n---\n# Python Standards\nFollow PEP 8.\n")

	out, name, err := translator.TranslateContent(content, "python.instructions.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outStr := string(out)

	// Should have globs instead of applyTo
	if !strings.Contains(outStr, "globs:") {
		t.Error("output should contain 'globs:'")
	}
	if strings.Contains(outStr, "applyTo:") {
		t.Error("output should not contain 'applyTo:'")
	}

	// Should have alwaysApply: false
	if !strings.Contains(outStr, "alwaysApply: false") {
		t.Error("output should contain 'alwaysApply: false'")
	}

	// Should preserve description
	if !strings.Contains(outStr, "description: Python standards") {
		t.Error("output should preserve description")
	}

	// Should preserve body
	if !strings.Contains(outStr, "Follow PEP 8.") {
		t.Error("output should preserve body content")
	}

	// Filename should be .mdc
	if name != "python.mdc" {
		t.Errorf("filename: got %q, want %q", name, "python.mdc")
	}
}

func TestCursorTranslatorNoFrontmatter(t *testing.T) {
	translator := NewInstructionTranslator("cursor")
	content := []byte("# Some Instructions\nDo this.\n")

	out, name, err := translator.TranslateContent(content, "general.instructions.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outStr := string(out)

	// Should add alwaysApply: true when no frontmatter (no pattern to target)
	if !strings.Contains(outStr, "alwaysApply: true") {
		t.Error("output should contain 'alwaysApply: true' when no frontmatter")
	}

	// Should preserve body
	if !strings.Contains(outStr, "# Some Instructions") {
		t.Error("output should preserve body content")
	}

	// Filename should be .mdc
	if name != "general.mdc" {
		t.Errorf("filename: got %q, want %q", name, "general.mdc")
	}
}

func TestCursorFilenameConversion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"security.instructions.md", "security.mdc"},
		{"python-standards.instructions.md", "python-standards.mdc"},
		{"readme.md", "readme.mdc"},
		{"noext", "noext.mdc"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toCursorFilename(tt.input)
			if got != tt.expected {
				t.Errorf("toCursorFilename(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSplitFrontmatter(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		hasFrontmatter bool
		frontmatter    string
		body           string
	}{
		{
			name:           "with frontmatter",
			input:          "---\nkey: value\n---\nBody content",
			hasFrontmatter: true,
			frontmatter:    "key: value",
			body:           "Body content",
		},
		{
			name:           "no frontmatter",
			input:          "# Just a heading\nSome content",
			hasFrontmatter: false,
			body:           "# Just a heading\nSome content",
		},
		{
			name:           "multiple frontmatter fields",
			input:          "---\ndescription: Test\napplyTo: '**/*.go'\n---\n# Go Standards",
			hasFrontmatter: true,
			frontmatter:    "description: Test\napplyTo: '**/*.go'",
			body:           "# Go Standards",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm, body, hasFM := splitFrontmatter([]byte(tt.input))

			if hasFM != tt.hasFrontmatter {
				t.Errorf("hasFrontmatter: got %v, want %v", hasFM, tt.hasFrontmatter)
			}

			if hasFM {
				if fm != tt.frontmatter {
					t.Errorf("frontmatter:\ngot:  %q\nwant: %q", fm, tt.frontmatter)
				}
			}

			if string(body) != tt.body {
				t.Errorf("body:\ngot:  %q\nwant: %q", string(body), tt.body)
			}
		})
	}
}

func TestNewInstructionTranslator(t *testing.T) {
	tests := []struct {
		assistant string
		wantType  string
	}{
		{"copilot", "*installer.PassthroughInstructionTranslator"},
		{"claude", "*installer.PassthroughInstructionTranslator"},
		{"opencode", "*installer.PassthroughInstructionTranslator"},
		{"cursor", "*installer.CursorInstructionTranslator"},
		{"gemini", "*installer.PassthroughInstructionTranslator"},
		{"codex", "*installer.PassthroughInstructionTranslator"},
	}

	for _, tt := range tests {
		t.Run(tt.assistant, func(t *testing.T) {
			translator := NewInstructionTranslator(tt.assistant)
			if translator == nil {
				t.Fatal("translator should not be nil")
			}
		})
	}
}
