package mcp

import (
	"fmt"
	"sort"

	"gopkg.in/yaml.v3"
)

// ServerTransport defines how the client connects to the MCP server.
// MCP supports three transports:
//   - stdio: the server is a local subprocess communicating over stdin/stdout
//   - sse: the server is a remote HTTP endpoint using Server-Sent Events
//   - streamable-http: the server is a remote HTTP endpoint using streamable HTTP
type ServerTransport string

const (
	TransportStdio          ServerTransport = "stdio"
	TransportSSE            ServerTransport = "sse"
	TransportStreamableHTTP ServerTransport = "streamable-http"
)

// Server represents a single MCP server definition in the canonical format.
// This is the "source of truth" schema — translators convert it to each
// assistant's native JSON format.
type Server struct {
	Description string            `yaml:"description,omitempty" json:"description,omitempty"`
	Transport   ServerTransport   `yaml:"transport" json:"transport"`
	Command     string            `yaml:"command,omitempty" json:"command,omitempty"`   // stdio only
	Args        []string          `yaml:"args,omitempty" json:"args,omitempty"`         // stdio only
	URL         string            `yaml:"url,omitempty" json:"url,omitempty"`           // sse / streamable-http only
	Headers     map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`   // HTTP transports only
	Env         map[string]string `yaml:"env,omitempty" json:"env,omitempty"`           // Environment variables
	Required    bool              `yaml:"required,omitempty" json:"required,omitempty"` // Whether the package needs this MCP to function
}

// Config is the top-level mcp.yaml file structure.
type Config struct {
	Servers map[string]Server `yaml:"servers" json:"servers"`
}

// ParseConfig unmarshals raw YAML bytes into a Config and validates it.
func ParseConfig(data []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse mcp.yaml: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Validate checks that every server in the config has valid field
// combinations for its declared transport.
func (c *Config) Validate() error {
	if len(c.Servers) == 0 {
		return fmt.Errorf("mcp.yaml: no servers defined")
	}

	// Sort server names for deterministic error messages
	names := make([]string, 0, len(c.Servers))
	for name := range c.Servers {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		s := c.Servers[name]
		if err := validateServer(name, s); err != nil {
			return err
		}
	}
	return nil
}

// validateServer checks a single server definition for valid field combinations.
func validateServer(name string, s Server) error {
	switch s.Transport {
	case TransportStdio:
		if s.Command == "" {
			return fmt.Errorf("mcp.yaml: server %q (transport: stdio) requires a command", name)
		}
		if s.URL != "" {
			return fmt.Errorf("mcp.yaml: server %q (transport: stdio) must not set url", name)
		}
	case TransportSSE, TransportStreamableHTTP:
		if s.URL == "" {
			return fmt.Errorf("mcp.yaml: server %q (transport: %s) requires a url", name, s.Transport)
		}
		if s.Command != "" {
			return fmt.Errorf("mcp.yaml: server %q (transport: %s) must not set command", name, s.Transport)
		}
	case "":
		return fmt.Errorf("mcp.yaml: server %q is missing required field: transport", name)
	default:
		return fmt.Errorf("mcp.yaml: server %q has unknown transport %q (valid: stdio, sse, streamable-http)", name, s.Transport)
	}
	return nil
}

// EmptyEnvVars returns a map of server name → list of env var names that have
// empty string values. Callers use this to warn or prompt users.
func (c *Config) EmptyEnvVars() map[string][]string {
	result := make(map[string][]string)
	for name, s := range c.Servers {
		var empty []string
		// Sort keys for deterministic output
		keys := make([]string, 0, len(s.Env))
		for k := range s.Env {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if s.Env[k] == "" {
				empty = append(empty, k)
			}
		}
		if len(empty) > 0 {
			result[name] = empty
		}
	}
	return result
}
