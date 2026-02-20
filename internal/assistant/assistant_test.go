package assistant

import (
	"strings"
	"testing"
)

func TestGetReturnsCorrectConfig(t *testing.T) {
	tests := []struct {
		name           string // Description shown in test output
		assistant      string // Input to Get()
		agentDir       string // Expected AgentDir
		skillDir       string // Expected SkillDir
		instructionDir string // Expected InstructionDir
		mcpConfigPath  string // Expected MCPConfigPath
		mcpConfigKey   string // Expected MCPConfigKey
	}{
		{
			name:           "copilot uses .github/agents",
			assistant:      "copilot",
			agentDir:       ".github/agents",
			skillDir:       "skills",
			instructionDir: ".github/instructions",
			mcpConfigPath:  ".vscode/mcp.json",
			mcpConfigKey:   "servers",
		},
		{
			name:           "claude uses .claude directories",
			assistant:      "claude",
			agentDir:       ".claude/agents",
			skillDir:       ".claude/skills",
			instructionDir: ".claude/instructions",
			mcpConfigPath:  ".claude/settings.local.json",
			mcpConfigKey:   "mcpServers",
		},
		{
			name:           "opencode uses .opencode directories",
			assistant:      "opencode",
			agentDir:       ".opencode/agents",
			skillDir:       ".opencode/skills",
			instructionDir: ".opencode/instructions",
			mcpConfigPath:  "opencode.json",
			mcpConfigKey:   "mcp",
		},
		{
			name:           "cursor uses .cursor directories",
			assistant:      "cursor",
			agentDir:       ".cursor/agents",
			skillDir:       ".cursor/skills",
			instructionDir: ".cursor/rules",
			mcpConfigPath:  ".cursor/mcp.json",
			mcpConfigKey:   "mcpServers",
		},
		{
			name:           "gemini uses .gemini directories",
			assistant:      "gemini",
			agentDir:       ".gemini/agents",
			skillDir:       ".gemini/skills",
			instructionDir: ".gemini/instructions",
			mcpConfigPath:  ".gemini/settings.json",
			mcpConfigKey:   "mcpServers",
		},
		{
			name:           "codex uses .agents directories",
			assistant:      "codex",
			agentDir:       ".agents/agents",
			skillDir:       ".agents/skills",
			instructionDir: ".agents/instructions",
			mcpConfigPath:  ".codex/config.toml",
			mcpConfigKey:   "mcp_servers",
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
			if cfg.InstructionDir != tt.instructionDir {
				t.Errorf("InstructionDir: got %q, want %q", cfg.InstructionDir, tt.instructionDir)
			}
			if cfg.MCPConfigPath != tt.mcpConfigPath {
				t.Errorf("MCPConfigPath: got %q, want %q", cfg.MCPConfigPath, tt.mcpConfigPath)
			}
			if cfg.MCPConfigKey != tt.mcpConfigKey {
				t.Errorf("MCPConfigKey: got %q, want %q", cfg.MCPConfigKey, tt.mcpConfigKey)
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

	// We expect exactly 6 assistants
	if len(names) != 6 {
		t.Fatalf("List() returned %d names, want 6", len(names))
	}

	// Check each expected name is present
	expected := []string{"claude", "codex", "copilot", "cursor", "gemini", "opencode"}
	for i, want := range expected {
		if names[i] != want {
			t.Errorf("List()[%d]: got %q, want %q", i, names[i], want)
		}
	}
}

func TestFlagUsageContainsAllAssistants(t *testing.T) {
	usage := FlagUsage()

	// Should contain every registered assistant name
	for _, name := range List() {
		if !strings.Contains(usage, name) {
			t.Errorf("FlagUsage() should contain %q, got: %s", name, usage)
		}
	}

	// Should be comma-separated
	if usage != strings.Join(List(), ", ") {
		t.Errorf("FlagUsage() = %q, want %q", usage, strings.Join(List(), ", "))
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

		// --- Cursor: agents and skills remap to .cursor/ ---
		{
			name:      "cursor remaps agents",
			assistant: "cursor",
			input:     "agents/my-agent.md",
			expected:  ".cursor/agents/my-agent.md",
		},
		{
			name:      "cursor remaps skills",
			assistant: "cursor",
			input:     "skills/my-skill/SKILL.md",
			expected:  ".cursor/skills/my-skill/SKILL.md",
		},

		// --- Gemini: agents and skills remap to .gemini/ ---
		{
			name:      "gemini remaps agents",
			assistant: "gemini",
			input:     "agents/my-agent.md",
			expected:  ".gemini/agents/my-agent.md",
		},
		{
			name:      "gemini remaps skills",
			assistant: "gemini",
			input:     "skills/my-skill/SKILL.md",
			expected:  ".gemini/skills/my-skill/SKILL.md",
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

		// --- Instructions ---
		{
			name:      "copilot remaps instructions to .github/instructions",
			assistant: "copilot",
			input:     "instructions/security.instructions.md",
			expected:  ".github/instructions/security.instructions.md",
		},
		{
			name:      "claude remaps instructions to .claude/instructions",
			assistant: "claude",
			input:     "instructions/security.instructions.md",
			expected:  ".claude/instructions/security.instructions.md",
		},
		{
			name:      "cursor remaps instructions to .cursor/rules",
			assistant: "cursor",
			input:     "instructions/security.instructions.md",
			expected:  ".cursor/rules/security.instructions.md",
		},
		{
			name:      "gemini remaps instructions to .gemini/instructions",
			assistant: "gemini",
			input:     "instructions/security.instructions.md",
			expected:  ".gemini/instructions/security.instructions.md",
		},
		{
			name:      "codex remaps instructions to .agents/instructions",
			assistant: "codex",
			input:     "instructions/security.instructions.md",
			expected:  ".agents/instructions/security.instructions.md",
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

// TestNewReversePathMapperRemapsPaths verifies that NewReversePathMapper()
// correctly translates assistant-specific paths back to the generic layout.
func TestNewReversePathMapperRemapsPaths(t *testing.T) {
	tests := []struct {
		name      string
		assistant string
		input     string
		expected  string
	}{
		// --- Copilot: .github/agents → agents, skills stays ---
		{
			name:      "copilot reverses agents",
			assistant: "copilot",
			input:     ".github/agents/foo.md",
			expected:  "agents/foo.md",
		},
		{
			name:      "copilot keeps skills unchanged (identity)",
			assistant: "copilot",
			input:     "skills/bar/SKILL.md",
			expected:  "skills/bar/SKILL.md",
		},

		// --- Claude: .claude/agents → agents, .claude/skills → skills ---
		{
			name:      "claude reverses agents",
			assistant: "claude",
			input:     ".claude/agents/foo.md",
			expected:  "agents/foo.md",
		},
		{
			name:      "claude reverses skills",
			assistant: "claude",
			input:     ".claude/skills/bar/SKILL.md",
			expected:  "skills/bar/SKILL.md",
		},

		// --- OpenCode: .opencode/agents → agents, .opencode/skills → skills ---
		{
			name:      "opencode reverses agents",
			assistant: "opencode",
			input:     ".opencode/agents/foo.md",
			expected:  "agents/foo.md",
		},
		{
			name:      "opencode reverses skills",
			assistant: "opencode",
			input:     ".opencode/skills/bar/SKILL.md",
			expected:  "skills/bar/SKILL.md",
		},

		// --- Cursor: .cursor/agents → agents, .cursor/skills → skills ---
		{
			name:      "cursor reverses agents",
			assistant: "cursor",
			input:     ".cursor/agents/foo.md",
			expected:  "agents/foo.md",
		},
		{
			name:      "cursor reverses skills",
			assistant: "cursor",
			input:     ".cursor/skills/bar/SKILL.md",
			expected:  "skills/bar/SKILL.md",
		},

		// --- Gemini: .gemini/agents → agents, .gemini/skills → skills ---
		{
			name:      "gemini reverses agents",
			assistant: "gemini",
			input:     ".gemini/agents/foo.md",
			expected:  "agents/foo.md",
		},
		{
			name:      "gemini reverses skills",
			assistant: "gemini",
			input:     ".gemini/skills/bar/SKILL.md",
			expected:  "skills/bar/SKILL.md",
		},

		// --- Codex: .agents/agents → agents, .agents/skills → skills ---
		{
			name:      "codex reverses agents",
			assistant: "codex",
			input:     ".agents/agents/foo.md",
			expected:  "agents/foo.md",
		},
		{
			name:      "codex reverses skills",
			assistant: "codex",
			input:     ".agents/skills/bar/SKILL.md",
			expected:  "skills/bar/SKILL.md",
		},

		// --- Edge cases ---
		{
			name:      "unmapped path passes through",
			assistant: "claude",
			input:     "docs/README.md",
			expected:  "docs/README.md",
		},
		{
			name:      "bare agent directory name",
			assistant: "claude",
			input:     ".claude/agents",
			expected:  "agents",
		},
		{
			name:      "nested skills path preserves structure",
			assistant: "opencode",
			input:     ".opencode/skills/dev-mentor/standards/checklist.md",
			expected:  "skills/dev-mentor/standards/checklist.md",
		},

		// --- Instructions (reverse) ---
		{
			name:      "copilot reverses instructions",
			assistant: "copilot",
			input:     ".github/instructions/security.instructions.md",
			expected:  "instructions/security.instructions.md",
		},
		{
			name:      "claude reverses instructions",
			assistant: "claude",
			input:     ".claude/instructions/security.instructions.md",
			expected:  "instructions/security.instructions.md",
		},
		{
			name:      "cursor reverses instructions from rules",
			assistant: "cursor",
			input:     ".cursor/rules/security.mdc",
			expected:  "instructions/security.mdc",
		},
		{
			name:      "opencode reverses instructions",
			assistant: "opencode",
			input:     ".opencode/instructions/security.instructions.md",
			expected:  "instructions/security.instructions.md",
		},
		{
			name:      "gemini reverses instructions",
			assistant: "gemini",
			input:     ".gemini/instructions/security.instructions.md",
			expected:  "instructions/security.instructions.md",
		},
		{
			name:      "codex reverses instructions",
			assistant: "codex",
			input:     ".agents/instructions/security.instructions.md",
			expected:  "instructions/security.instructions.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := Get(tt.assistant)
			if err != nil {
				t.Fatalf("Get(%q) returned unexpected error: %v", tt.assistant, err)
			}

			mapper := cfg.NewReversePathMapper()
			got := mapper(tt.input)

			if got != tt.expected {
				t.Errorf("mapper(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// TestPathMapperRoundTrip verifies that applying NewPathMapper() followed by
// NewReversePathMapper() returns the original path (and vice versa).
func TestPathMapperRoundTrip(t *testing.T) {
	genericPaths := []string{
		"agents/my-agent.agent.md",
		"skills/dev-mentor/SKILL.md",
		"skills/dev-mentor/actions/create.md",
		"agents/git-workflow.agent.md",
		"instructions/security.instructions.md",
		"instructions/python-standards.instructions.md",
	}

	for _, assistant := range List() {
		t.Run(assistant, func(t *testing.T) {
			cfg, err := Get(assistant)
			if err != nil {
				t.Fatalf("Get(%q) returned unexpected error: %v", assistant, err)
			}

			forward := cfg.NewPathMapper()
			reverse := cfg.NewReversePathMapper()

			for _, p := range genericPaths {
				// generic → assistant → generic
				mapped := forward(p)
				roundTripped := reverse(mapped)
				if roundTripped != p {
					t.Errorf("round-trip failed for %q: forward(%q) = %q, reverse(%q) = %q",
						p, p, mapped, mapped, roundTripped)
				}
			}
		})
	}
}
