package mcp

import "fmt"

// ClaudeTranslator converts canonical MCP config to Claude Code's
// .claude/settings.local.json format.
//
// Claude format (under "mcpServers" key):
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
// We target .claude/settings.local.json (not settings.json) to avoid
// committing secrets to version control.
type ClaudeTranslator struct{}

func (c *ClaudeTranslator) ConfigPath() string { return ".claude/settings.local.json" }
func (c *ClaudeTranslator) ConfigKey() string  { return "mcpServers" }

func (c *ClaudeTranslator) Translate(cfg *Config) (map[string]any, []string, error) {
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
			// Claude Code supports HTTP-based MCP servers via url
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
			warnings = append(warnings, fmt.Sprintf("server %q: unsupported transport %q for Claude", name, s.Transport))
		}
	}

	return servers, warnings, nil
}
