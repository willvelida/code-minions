package mcp

import (
	"encoding/json"
	"fmt"
)

// GeminiTranslator converts canonical MCP config to Gemini CLI's
// .gemini/settings.json format.
//
// Gemini format (under "mcpServers" key):
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
// Gemini CLI uses the same JSON schema as Claude Code and Cursor for
// MCP servers (mcpServers key, command/args/env for stdio, url for
// HTTP/SSE) but stores the config in .gemini/settings.json.
type GeminiTranslator struct{}

func (g *GeminiTranslator) ConfigPath() string { return ".gemini/settings.json" }
func (g *GeminiTranslator) ConfigKey() string  { return "mcpServers" }

func (g *GeminiTranslator) Translate(cfg *Config) (map[string]any, []string, error) {
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
			// Gemini CLI supports HTTP-based MCP servers via url
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
			warnings = append(warnings, fmt.Sprintf("server %q: unsupported transport %q for Gemini", name, s.Transport))
		}
	}

	return servers, warnings, nil
}

// GeminiReader parses Gemini CLI's .gemini/settings.json format.
//
// Gemini format:
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
type GeminiReader struct{}

func (r *GeminiReader) ConfigPath() string { return ".gemini/settings.json" }
func (r *GeminiReader) ConfigKey() string  { return "mcpServers" }

func (r *GeminiReader) Read(data []byte) (*Config, []string, error) {
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, nil, fmt.Errorf("failed to parse Gemini MCP config: %w", err)
	}

	raw, ok := doc["mcpServers"]
	if !ok {
		return &Config{Servers: map[string]Server{}}, nil, nil
	}

	var servers map[string]map[string]any
	if err := json.Unmarshal(raw, &servers); err != nil {
		return nil, nil, fmt.Errorf("failed to parse Gemini mcpServers: %w", err)
	}

	return readStandardFormat(servers, "Gemini")
}
