package mcp

import "fmt"

// OpenCodeTranslator converts canonical MCP config to OpenCode's
// opencode.json format.
//
// OpenCode format (under "mcp" key):
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
//
// Key differences from canonical:
//   - command is an array (command + args concatenated)
//   - env → environment
//   - type: "local" for stdio, "remote" for HTTP transports
//   - enabled: true always set
type OpenCodeTranslator struct{}

func (o *OpenCodeTranslator) ConfigPath() string { return "opencode.json" }
func (o *OpenCodeTranslator) ConfigKey() string  { return "mcp" }

func (o *OpenCodeTranslator) Translate(cfg *Config) (map[string]any, []string, error) {
	servers := make(map[string]any)
	var warnings []string

	for name, s := range cfg.Servers {
		switch s.Transport {
		case TransportStdio:
			// OpenCode combines command + args into a single array
			cmdArray := make([]string, 0, 1+len(s.Args))
			cmdArray = append(cmdArray, s.Command)
			cmdArray = append(cmdArray, s.Args...)

			entry := map[string]any{
				"type":    "local",
				"command": cmdArray,
				"enabled": true,
			}
			if len(s.Env) > 0 {
				entry["environment"] = s.Env
			}
			servers[name] = entry

		case TransportSSE, TransportStreamableHTTP:
			entry := map[string]any{
				"type":    "remote",
				"url":     s.URL,
				"enabled": true,
			}
			if len(s.Headers) > 0 {
				entry["headers"] = s.Headers
			}
			if len(s.Env) > 0 {
				entry["environment"] = s.Env
			}
			servers[name] = entry

		default:
			warnings = append(warnings, fmt.Sprintf("server %q: unsupported transport %q for OpenCode", name, s.Transport))
		}
	}

	return servers, warnings, nil
}
