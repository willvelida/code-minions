package mcp

import "fmt"

// Translator converts a canonical MCP Config into the JSON structure
// expected by a specific coding assistant. Each assistant has its own
// file path, JSON key, and format quirks.
type Translator interface {
	// Translate converts the canonical config to the assistant's native
	// server map. The returned map is the value that goes under the
	// assistant's MCP config key (e.g. "servers" for Copilot).
	// Servers that are incompatible with this assistant (e.g. HTTP
	// transport on an assistant that only supports stdio) are skipped
	// and a warning is added to the returned slice.
	Translate(cfg *Config) (servers map[string]any, warnings []string, err error)

	// ConfigPath returns the relative file path for this assistant's
	// MCP configuration (e.g. ".vscode/mcp.json").
	ConfigPath() string

	// ConfigKey returns the top-level JSON key that holds MCP servers
	// (e.g. "servers", "mcpServers", "mcp").
	ConfigKey() string
}

// NewTranslator returns a Translator for the named assistant.
// Returns an error if the assistant is not recognised.
func NewTranslator(assistant string) (Translator, error) {
	switch assistant {
	case "copilot":
		return &CopilotTranslator{}, nil
	case "claude":
		return &ClaudeTranslator{}, nil
	case "opencode":
		return &OpenCodeTranslator{}, nil
	default:
		return nil, fmt.Errorf("no MCP translator for assistant %q", assistant)
	}
}
