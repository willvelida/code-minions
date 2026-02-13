package assistant

import (
	"strings"
	"testing"
)

func TestGetReturnsCorrectConfig(t *testing.T) {
	tests := []struct {
		name      string // Description shown in test output
		assistant string // Input to Get()
		agentDir  string // Expected AgentDir
		skillDir  string // Expected SkillDir
	}{
		{
			name:      "copilot uses .github/agents",
			assistant: "copilot",
			agentDir:  ".github/agents",
			skillDir:  "skills",
		},
		{
			name:      "claude uses .claude directories",
			assistant: "claude",
			agentDir:  ".claude/agents",
			skillDir:  ".claude/skills",
		},
		{
			name:      "opencode uses .opencode directories",
			assistant: "opencode",
			agentDir:  ".opencode/agents",
			skillDir:  ".opencode/skills",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := Get(tt.assistant)
			if err != nil {
				t.Fatalf("Get(%q) returned unexpected error: %v", tt.assistant, err)
			}

			// Check each field individually so failures are clear
			if cfg.Name != tt.assistant {
				t.Errorf("Name: got %q, want %q", cfg.Name, tt.assistant)
			}
			if cfg.AgentDir != tt.agentDir {
				t.Errorf("AgentDir: got %q, want %q", cfg.AgentDir, tt.agentDir)
			}
			if cfg.SkillDir != tt.skillDir {
				t.Errorf("SkillDir: got %q, want %q", cfg.SkillDir, tt.skillDir)
			}
		})
	}
}

func TestGetUnknownAssistantReturnsError(t *testing.T) {
	_, err := Get("unknown")
	if err == nil {
		t.Fatal("expected error for unknown assistant, got nil")
	}

	// The error message should include the invalid name
	if !strings.Contains(err.Error(), "unknown") {
		t.Errorf("error should mention the invalid name, got: %v", err)
	}

	// The error message should list valid options so the user knows what to type
	for _, name := range List() {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("error should mention valid option %q, got: %v", name, err)
		}
	}
}

func TestListReturnsAllAssistants(t *testing.T) {
	names := List()

	// We expect exactly 3 assistants
	if len(names) != 3 {
		t.Fatalf("List() returned %d names, want 3", len(names))
	}

	// Check each expected name is present
	expected := []string{"claude", "copilot", "opencode"}
	for i, want := range expected {
		if names[i] != want {
			t.Errorf("List()[%d]: got %q, want %q", i, names[i], want)
		}
	}
}

func TestGetReturnsCopy(t *testing.T) {
	cfg, err := Get("copilot")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Mutate the returned config
	cfg.AgentDir = "modified"

	// Fetch again — should still have the original value
	cfg2, err := Get("copilot")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg2.AgentDir == "modified" {
		t.Error("modifying returned Config should not affect the registry")
	}
}

// TestNewPathMapperRemapsPaths is a table-driven test that verifies
// NewPathMapper() correctly translates paths for each assistant.
//
// Each row specifies an assistant, an input path, and the expected output.
// This covers the three content types (agents, skills, standards) plus
// paths that don't match any prefix (pass-through).
func TestNewPathMapperRemapsPaths(t *testing.T) {
	tests := []struct {
		name      string // Description for test output
		assistant string // Which assistant's mapper to use
		input     string // Path coming out of the embedded FS (after prefix strip)
		expected  string // Where it should end up
	}{
		// --- Copilot: agents remap to .github/agents, skills stay ---
		{
			name:      "copilot remaps agents",
			assistant: "copilot",
			input:     "agents/my-agent.agent.md",
			expected:  ".github/agents/my-agent.agent.md",
		},
		{
			name:      "copilot keeps skills unchanged",
			assistant: "copilot",
			input:     "skills/my-skill/SKILL.md",
			expected:  "skills/my-skill/SKILL.md",
		},

		// --- Claude: agents and skills remap to .claude/ ---
		{
			name:      "claude remaps agents",
			assistant: "claude",
			input:     "agents/my-agent.md",
			expected:  ".claude/agents/my-agent.md",
		},
		{
			name:      "claude remaps skills",
			assistant: "claude",
			input:     "skills/my-skill/SKILL.md",
			expected:  ".claude/skills/my-skill/SKILL.md",
		},

		// --- OpenCode: agents and skills remap to .opencode/ ---
		{
			name:      "opencode remaps agents",
			assistant: "opencode",
			input:     "agents/my-agent.md",
			expected:  ".opencode/agents/my-agent.md",
		},
		{
			name:      "opencode remaps skills",
			assistant: "opencode",
			input:     "skills/my-skill/actions/create.md",
			expected:  ".opencode/skills/my-skill/actions/create.md",
		},

		// --- Edge cases ---
		{
			name:      "unknown prefix passes through",
			assistant: "copilot",
			input:     "docs/README.md",
			expected:  "docs/README.md",
		},
		{
			name:      "bare agents directory name with no trailing content",
			assistant: "claude",
			input:     "agents",
			expected:  ".claude/agents",
		},
		{
			name:      "nested skills path preserves full structure",
			assistant: "opencode",
			input:     "skills/dev-mentor/standards/checklist.md",
			expected:  ".opencode/skills/dev-mentor/standards/checklist.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := Get(tt.assistant)
			if err != nil {
				t.Fatalf("Get(%q) returned unexpected error: %v", tt.assistant, err)
			}

			mapper := cfg.NewPathMapper()
			got := mapper(tt.input)

			if got != tt.expected {
				t.Errorf("mapper(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
