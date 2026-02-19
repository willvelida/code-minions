package mcp

import (
	"encoding/json"
	"fmt"
)

// CursorTranslator converts canonical MCP config to Cursor's
// .cursor/mcp.json format.
//
// Cursor format (under "mcpServers" key):
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
//
// Cursor uses the same JSON schema as Claude Code for MCP servers
// (mcpServers key, command/args/env for stdio, url for HTTP) but
// stores the config in .cursor/mcp.json instead of
// .claude/settings.local.json.
type CursorTranslator struct{}

func (c *CursorTranslator) ConfigPath() string { return ".cursor/mcp.json" }
func (c *CursorTranslator) ConfigKey() string  { return "mcpServers" }

func (c *CursorTranslator) Translate(cfg *Config) (map[string]any, []string, error) {
	servers := make(map[string]any)
	var warnings []string

	for name, s := range cfg.Servers {
		switch s.Transport {
		case TransportStdio:
			entry := map[string]any{
				"command": s.Command,
			}
			if len(s.Args) > 0 {
				entry["args"] = s.Args
			}
			if len(s.Env) > 0 {
				entry["env"] = s.Env
			}
			servers[name] = entry

		case TransportSSE, TransportStreamableHTTP:
			// Cursor supports HTTP-based MCP servers via url
			entry := map[string]any{
				"url": s.URL,
			}
			if len(s.Headers) > 0 {
				entry["headers"] = s.Headers
			}
			if len(s.Env) > 0 {
				entry["env"] = s.Env
			}
			servers[name] = entry

		default:
			warnings = append(warnings, fmt.Sprintf("server %q: unsupported transport %q for Cursor", name, s.Transport))
		}
	}

	return servers, warnings, nil
}

// CursorReader parses Cursor's .cursor/mcp.json format.
//
// Cursor format:
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
type CursorReader struct{}

func (r *CursorReader) ConfigPath() string { return ".cursor/mcp.json" }
func (r *CursorReader) ConfigKey() string  { return "mcpServers" }

func (r *CursorReader) Read(data []byte) (*Config, []string, error) {
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, nil, fmt.Errorf("failed to parse Cursor MCP config: %w", err)
	}

	raw, ok := doc["mcpServers"]
	if !ok {
		return &Config{Servers: map[string]Server{}}, nil, nil
	}

	var servers map[string]map[string]any
	if err := json.Unmarshal(raw, &servers); err != nil {
		return nil, nil, fmt.Errorf("failed to parse Cursor mcpServers: %w", err)
	}

	return readStandardFormat(servers, "Cursor")
}
