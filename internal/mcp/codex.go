package mcp

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

// CodexTranslator converts canonical MCP config to Codex CLI's
// .codex/config.toml format.
//
// Codex format (TOML):
//
//	[mcp_servers.github]
//	type = "stdio"
//	command = "npx"
//	args = ["-y", "@modelcontextprotocol/server-github"]
//
//	[mcp_servers.github.env]
//	GITHUB_PERSONAL_ACCESS_TOKEN = ""
//
// Codex CLI stores MCP config in .codex/config.toml under [mcp_servers].
// Each server is a sub-table keyed by name. Transport types:
//   - "stdio" → command + args + env
//   - "sse"   → url + headers
//   - "streamable-http" → url + headers
type CodexTranslator struct{}

func (c *CodexTranslator) ConfigPath() string { return ".codex/config.toml" }
func (c *CodexTranslator) ConfigKey() string  { return "mcp_servers" }

func (c *CodexTranslator) Translate(cfg *Config) (map[string]any, []string, error) {
	servers := make(map[string]any)
	var warnings []string

	for name, s := range cfg.Servers {
		switch s.Transport {
		case TransportStdio:
			entry := map[string]any{
				"type":    "stdio",
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
			entry := map[string]any{
				"type": "sse",
				"url":  s.URL,
			}
			if len(s.Headers) > 0 {
				entry["headers"] = s.Headers
			}
			if len(s.Env) > 0 {
				entry["env"] = s.Env
			}
			servers[name] = entry

		case TransportStreamableHTTP:
			entry := map[string]any{
				"type": "streamable-http",
				"url":  s.URL,
			}
			if len(s.Headers) > 0 {
				entry["headers"] = s.Headers
			}
			if len(s.Env) > 0 {
				entry["env"] = s.Env
			}
			servers[name] = entry

		default:
			warnings = append(warnings, fmt.Sprintf("server %q: unsupported transport %q for Codex", name, s.Transport))
		}
	}

	return servers, warnings, nil
}

// CodexReader parses Codex CLI's .codex/config.toml format.
//
// Codex uses TOML with [mcp_servers.<name>] sections. Transport is
// determined by the "type" field:
//   - "stdio"            → command + args + env
//   - "sse"              → url + headers
//   - "streamable-http"  → url + headers
//
// When "type" is absent, presence of "command" implies stdio;
// presence of "url" implies SSE.
type CodexReader struct{}

func (r *CodexReader) ConfigPath() string { return ".codex/config.toml" }
func (r *CodexReader) ConfigKey() string  { return "mcp_servers" }

func (r *CodexReader) Read(data []byte) (*Config, []string, error) {
	var doc map[string]any
	if err := toml.Unmarshal(data, &doc); err != nil {
		return nil, nil, fmt.Errorf("failed to parse Codex MCP config: %w", err)
	}

	raw, ok := doc["mcp_servers"]
	if !ok {
		return &Config{Servers: map[string]Server{}}, nil, nil
	}

	serversRaw, ok := raw.(map[string]any)
	if !ok {
		return nil, nil, fmt.Errorf("failed to parse Codex mcp_servers: expected table, got %T", raw)
	}

	// Convert each server entry to map[string]any for readCodexFormat
	servers := make(map[string]map[string]any)
	for name, v := range serversRaw {
		m, ok := v.(map[string]any)
		if !ok {
			continue
		}
		servers[name] = m
	}

	return readCodexFormat(servers)
}

// readCodexFormat parses Codex's MCP server entries. Transport is
// determined by the "type" field, falling back to field detection:
//   - "stdio" or has "command"      → stdio
//   - "sse" or has "url" (no type)  → SSE
//   - "streamable-http"             → streamable HTTP
func readCodexFormat(servers map[string]map[string]any) (*Config, []string, error) {
	cfg := &Config{Servers: make(map[string]Server)}
	var warnings []string

	names := sortedKeys(servers)

	for _, name := range names {
		entry := servers[name]
		s := Server{}

		transportType, _ := entry["type"].(string)

		switch transportType {
		case "streamable-http":
			s.Transport = TransportStreamableHTTP
			url, _ := entry["url"].(string)
			if url == "" {
				warnings = append(warnings, fmt.Sprintf("server %q: streamable-http server missing url, skipping", name))
				continue
			}
			s.URL = url

		case "sse":
			s.Transport = TransportSSE
			url, _ := entry["url"].(string)
			if url == "" {
				warnings = append(warnings, fmt.Sprintf("server %q: SSE server missing url, skipping", name))
				continue
			}
			s.URL = url

		case "stdio":
			command, _ := entry["command"].(string)
			if command == "" {
				warnings = append(warnings, fmt.Sprintf("server %q: missing command in Codex config, skipping", name))
				continue
			}
			s.Transport = TransportStdio
			s.Command = command
			if argsRaw, ok := entry["args"]; ok {
				if args, err := toStringSlice(argsRaw); err == nil {
					s.Args = args
				}
			}

		case "":
			// No explicit type — detect from fields
			command, _ := entry["command"].(string)
			if command != "" {
				s.Transport = TransportStdio
				s.Command = command
				if argsRaw, ok := entry["args"]; ok {
					if args, err := toStringSlice(argsRaw); err == nil {
						s.Args = args
					}
				}
			} else if url, ok := entry["url"].(string); ok && url != "" {
				s.Transport = TransportSSE
				s.URL = url
			} else {
				warnings = append(warnings, fmt.Sprintf("server %q: missing command/url in Codex config, skipping", name))
				continue
			}

		default:
			warnings = append(warnings, fmt.Sprintf("server %q: unknown transport type %q in Codex config, skipping", name, transportType))
			continue
		}

		// Headers (applies to URL-based transports)
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
