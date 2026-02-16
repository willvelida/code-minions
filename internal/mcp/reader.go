package mcp

import (
	"encoding/json"
	"fmt"
	"sort"
)

// Reader parses an assistant's native MCP config JSON into the canonical
// format. Each assistant has its own Reader implementation that understands
// the JSON schema used by that assistant.
type Reader interface {
	// Read parses the given JSON bytes and returns the canonical Config.
	// Returns an empty Config (not nil, no error) when the file exists
	// but contains no MCP servers.
	Read(data []byte) (*Config, []string, error)

	// ConfigPath returns the relative file path for this assistant's
	// MCP configuration (e.g. ".vscode/mcp.json").
	ConfigPath() string

	// ConfigKey returns the top-level JSON key that holds MCP servers
	// (e.g. "servers", "mcpServers", "mcp").
	ConfigKey() string
}

// NewReader returns a Reader for the named assistant.
// Returns an error if the assistant is not recognised.
func NewReader(assistant string) (Reader, error) {
	switch assistant {
	case "copilot":
		return &CopilotReader{}, nil
	case "claude":
		return &ClaudeReader{}, nil
	case "opencode":
		return &OpenCodeReader{}, nil
	default:
		return nil, fmt.Errorf("no MCP reader for assistant %q", assistant)
	}
}

// --- Copilot Reader ---

// CopilotReader parses GitHub Copilot's .vscode/mcp.json format.
//
// Copilot format:
//
//	{
//	  "servers": {
//	    "github": {
//	      "command": "npx",
//	      "args": ["-y", "@modelcontextprotocol/server-github"],
//	      "env": { "GITHUB_PERSONAL_ACCESS_TOKEN": "" }
//	    }
//	  }
//	}
type CopilotReader struct{}

func (r *CopilotReader) ConfigPath() string { return ".vscode/mcp.json" }
func (r *CopilotReader) ConfigKey() string  { return "servers" }

func (r *CopilotReader) Read(data []byte) (*Config, []string, error) {
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, nil, fmt.Errorf("failed to parse Copilot MCP config: %w", err)
	}

	raw, ok := doc["servers"]
	if !ok {
		return &Config{Servers: map[string]Server{}}, nil, nil
	}

	var servers map[string]map[string]any
	if err := json.Unmarshal(raw, &servers); err != nil {
		return nil, nil, fmt.Errorf("failed to parse Copilot servers: %w", err)
	}

	return readStandardFormat(servers, "Copilot")
}

// --- Claude Reader ---

// ClaudeReader parses Claude Code's .claude/settings.local.json format.
//
// Claude format:
//
//	{
//	  "mcpServers": {
//	    "github": {
//	      "command": "npx",
//	      "args": ["-y", "@modelcontextprotocol/server-github"],
//	      "env": { "GITHUB_PERSONAL_ACCESS_TOKEN": "" }
//	    }
//	  }
//	}
type ClaudeReader struct{}

func (r *ClaudeReader) ConfigPath() string { return ".claude/settings.local.json" }
func (r *ClaudeReader) ConfigKey() string  { return "mcpServers" }

func (r *ClaudeReader) Read(data []byte) (*Config, []string, error) {
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, nil, fmt.Errorf("failed to parse Claude MCP config: %w", err)
	}

	raw, ok := doc["mcpServers"]
	if !ok {
		return &Config{Servers: map[string]Server{}}, nil, nil
	}

	var servers map[string]map[string]any
	if err := json.Unmarshal(raw, &servers); err != nil {
		return nil, nil, fmt.Errorf("failed to parse Claude mcpServers: %w", err)
	}

	return readStandardFormat(servers, "Claude")
}

// --- OpenCode Reader ---

// OpenCodeReader parses OpenCode's opencode.json format.
//
// OpenCode format:
//
//	{
//	  "mcp": {
//	    "github": {
//	      "type": "local",
//	      "command": ["npx", "-y", "@modelcontextprotocol/server-github"],
//	      "environment": { "GITHUB_PERSONAL_ACCESS_TOKEN": "" },
//	      "enabled": true
//	    }
//	  }
//	}
type OpenCodeReader struct{}

func (r *OpenCodeReader) ConfigPath() string { return "opencode.json" }
func (r *OpenCodeReader) ConfigKey() string  { return "mcp" }

func (r *OpenCodeReader) Read(data []byte) (*Config, []string, error) {
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, nil, fmt.Errorf("failed to parse OpenCode config: %w", err)
	}

	raw, ok := doc["mcp"]
	if !ok {
		return &Config{Servers: map[string]Server{}}, nil, nil
	}

	var servers map[string]map[string]any
	if err := json.Unmarshal(raw, &servers); err != nil {
		return nil, nil, fmt.Errorf("failed to parse OpenCode mcp: %w", err)
	}

	cfg := &Config{Servers: make(map[string]Server)}
	var warnings []string

	// Sort for deterministic output
	names := sortedKeys(servers)

	for _, name := range names {
		entry := servers[name]

		// Check enabled flag — skip disabled servers
		if enabled, ok := entry["enabled"]; ok {
			if b, ok := enabled.(bool); ok && !b {
				warnings = append(warnings, fmt.Sprintf("server %q is disabled in OpenCode config, skipping", name))
				continue
			}
		}

		s := Server{}

		// Determine transport from "type" field
		ocType, _ := entry["type"].(string)
		switch ocType {
		case "local", "":
			s.Transport = TransportStdio

			// OpenCode stores command as an array: ["npx", "-y", "..."]
			cmdRaw, ok := entry["command"]
			if !ok {
				warnings = append(warnings, fmt.Sprintf("server %q: missing command field, skipping", name))
				continue
			}

			cmdArray, err := toStringSlice(cmdRaw)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("server %q: invalid command field: %v, skipping", name, err))
				continue
			}

			if len(cmdArray) == 0 {
				warnings = append(warnings, fmt.Sprintf("server %q: empty command array, skipping", name))
				continue
			}

			s.Command = cmdArray[0]
			if len(cmdArray) > 1 {
				s.Args = cmdArray[1:]
			}

		case "remote":
			s.Transport = TransportSSE
			url, _ := entry["url"].(string)
			if url == "" {
				warnings = append(warnings, fmt.Sprintf("server %q: remote type but missing url, skipping", name))
				continue
			}
			s.URL = url

		default:
			warnings = append(warnings, fmt.Sprintf("server %q: unknown type %q, skipping", name, ocType))
			continue
		}

		// Environment → Env (OpenCode uses "environment" instead of "env")
		if envRaw, ok := entry["environment"]; ok {
			if envMap, err := toStringMap(envRaw); err == nil {
				s.Env = envMap
			}
		}

		// Headers
		if headersRaw, ok := entry["headers"]; ok {
			if headersMap, err := toStringMap(headersRaw); err == nil {
				s.Headers = headersMap
			}
		}

		cfg.Servers[name] = s
	}

	return cfg, warnings, nil
}

// --- Shared helpers ---

// readStandardFormat parses the "standard" MCP JSON format used by Copilot and
// Claude (command/args/env with transport inferred from fields).
func readStandardFormat(servers map[string]map[string]any, assistantName string) (*Config, []string, error) {
	cfg := &Config{Servers: make(map[string]Server)}
	var warnings []string

	names := sortedKeys(servers)

	for _, name := range names {
		entry := servers[name]
		s := Server{}

		// Determine transport: if "url" is present, it's HTTP; otherwise stdio
		if url, ok := entry["url"].(string); ok && url != "" {
			// Check for explicit type to distinguish SSE vs streamable-http
			if t, ok := entry["type"].(string); ok && t == "http" {
				s.Transport = TransportStreamableHTTP
			} else {
				s.Transport = TransportSSE
			}
			s.URL = url

			// Headers
			if headersRaw, ok := entry["headers"]; ok {
				if headersMap, err := toStringMap(headersRaw); err == nil {
					s.Headers = headersMap
				}
			}
		} else {
			s.Transport = TransportStdio
			command, _ := entry["command"].(string)
			if command == "" {
				warnings = append(warnings, fmt.Sprintf("server %q: missing command field in %s config, skipping", name, assistantName))
				continue
			}
			s.Command = command

			// Args
			if argsRaw, ok := entry["args"]; ok {
				if args, err := toStringSlice(argsRaw); err == nil {
					s.Args = args
				}
			}
		}

		// Env
		if envRaw, ok := entry["env"]; ok {
			if envMap, err := toStringMap(envRaw); err == nil {
				s.Env = envMap
			}
		}

		cfg.Servers[name] = s
	}

	return cfg, warnings, nil
}

// toStringSlice converts a JSON value to []string. JSON arrays unmarshalled
// into any are []any, so we convert each element.
func toStringSlice(v any) ([]string, error) {
	switch val := v.(type) {
	case []any:
		result := make([]string, 0, len(val))
		for _, item := range val {
			s, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("expected string, got %T", item)
			}
			result = append(result, s)
		}
		return result, nil
	case []string:
		return val, nil
	default:
		return nil, fmt.Errorf("expected array, got %T", v)
	}
}

// toStringMap converts a JSON value to map[string]string. JSON objects
// unmarshalled into any are map[string]any, so we convert each value.
func toStringMap(v any) (map[string]string, error) {
	switch val := v.(type) {
	case map[string]any:
		result := make(map[string]string, len(val))
		for k, item := range val {
			s, ok := item.(string)
			if !ok {
				continue // skip non-string values gracefully
			}
			result[k] = s
		}
		return result, nil
	case map[string]string:
		return val, nil
	default:
		return nil, fmt.Errorf("expected object, got %T", v)
	}
}

// sortedKeys returns the keys of a map[string]T sorted alphabetically.
// NOTE: intentionally duplicated from internal/cmd/output.go to avoid
// a cross-package dependency between mcp and cmd.
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
