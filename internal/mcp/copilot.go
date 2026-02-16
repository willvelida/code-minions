package mcp

import "fmt"

// CopilotTranslator converts canonical MCP config to GitHub Copilot's
// .vscode/mcp.json format.
//
// Copilot format (under "servers" key):
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
type CopilotTranslator struct{}

func (c *CopilotTranslator) ConfigPath() string { return ".vscode/mcp.json" }
func (c *CopilotTranslator) ConfigKey() string  { return "servers" }

func (c *CopilotTranslator) Translate(cfg *Config) (map[string]any, []string, error) {
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
			// Copilot supports SSE via url field
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
			// Copilot supports streamable-http via url + type
			entry := map[string]any{
				"type": "http",
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
			warnings = append(warnings, fmt.Sprintf("server %q: unsupported transport %q for Copilot", name, s.Transport))
		}
	}

	return servers, warnings, nil
}
