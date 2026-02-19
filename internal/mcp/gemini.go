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
// Gemini CLI stores MCP config in .gemini/settings.json under the
// "mcpServers" key. Unlike Copilot/Claude (which use "type": "http"
// to distinguish streamable-http from SSE), Gemini uses separate
// field names:
//   - "command" + "args" + "env" → stdio
//   - "url"     → SSE (SSEClientTransport)
//   - "httpUrl"  → Streamable HTTP (StreamableHTTPClientTransport)
//
// Headers work with both "url" and "httpUrl".
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

		case TransportSSE:
			// Gemini uses "url" for SSE transport
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

		case TransportStreamableHTTP:
			// Gemini uses "httpUrl" for streamable-http transport
			entry := map[string]any{
				"httpUrl": s.URL,
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
// Gemini uses different field names to distinguish transports:
//   - "command" + "args" → stdio
//   - "url"              → SSE
//   - "httpUrl"           → streamable HTTP
//
// This differs from Copilot/Claude which use "type": "http" as a
// discriminator. We therefore cannot use readStandardFormat() and
// need Gemini-specific parsing.
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

	return readGeminiFormat(servers)
}

// readGeminiFormat parses Gemini's MCP server entries. Transport is
// determined by which URL field is present:
//   - "httpUrl" → streamable-http
//   - "url"     → SSE
//   - "command" → stdio
func readGeminiFormat(servers map[string]map[string]any) (*Config, []string, error) {
	cfg := &Config{Servers: make(map[string]Server)}
	var warnings []string

	names := sortedKeys(servers)

	for _, name := range names {
		entry := servers[name]
		s := Server{}

		if httpURL, ok := entry["httpUrl"].(string); ok && httpURL != "" {
			// httpUrl → streamable-http transport
			s.Transport = TransportStreamableHTTP
			s.URL = httpURL
		} else if url, ok := entry["url"].(string); ok && url != "" {
			// url → SSE transport
			s.Transport = TransportSSE
			s.URL = url
		} else {
			// Fallback: stdio transport
			s.Transport = TransportStdio
			command, _ := entry["command"].(string)
			if command == "" {
				warnings = append(warnings, fmt.Sprintf("server %q: missing command/url/httpUrl in Gemini config, skipping", name))
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

		// Headers (applies to both url and httpUrl)
		if headersRaw, ok := entry["headers"]; ok {
			if headersMap, err := toStringMap(headersRaw); err == nil {
				s.Headers = headersMap
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
