package model

import (
	"fmt"
	"regexp"

	"github.com/willvelida/code-minions/internal/mcp"
)

// teamNamePattern defines the allowed characters in a team name.
// Only alphanumeric, hyphens, and underscores are permitted.
// This prevents injection of special characters into HTML comment
// markers used for instruction injection.
var teamNamePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

// Team is a named grouping of personas, allowing team leads to
// standardise which coding assistant configurations their team uses.
//
// Teams carry shared configuration beyond just "which personas to install":
//   - MCP: shared MCP server definitions installed for all team personas
//   - Instructions: org-wide instructions injected into every persona's agents
//
// MCP servers are team-level infrastructure (the server connection config),
// while individual personas choose which tools from those servers to use.
type Team struct {
	Name         string       `yaml:"name" json:"name"`
	Description  string       `yaml:"description" json:"description"`
	Author       string       `yaml:"author,omitempty" json:"author,omitempty"`
	Personas     []PersonaRef `yaml:"personas" json:"personas"`
	Config       TeamConfig   `yaml:"config,omitempty" json:"config,omitempty"`
	MCP          *mcp.Config  `yaml:"mcp,omitempty" json:"mcp,omitempty"`
	Instructions string       `yaml:"instructions,omitempty" json:"instructions,omitempty"`
}

// PersonaRef is a reference to a persona within a team.
// When Packages is non-empty, the persona is defined inline with its
// own curated package set. When Packages is empty, the persona is a
// reference to an existing persona definition resolved at install time.
type PersonaRef struct {
	Name     string       `yaml:"name" json:"name"`
	Version  string       `yaml:"version,omitempty" json:"version,omitempty"`
	Source   string       `yaml:"source,omitempty" json:"source,omitempty"`
	Packages []PackageRef `yaml:"packages,omitempty" json:"packages,omitempty"`
}

// TeamConfig holds team-level settings.
type TeamConfig struct {
	DefaultAssistant string `yaml:"default_assistant,omitempty" json:"default_assistant,omitempty"`
	EnforcePackages  bool   `yaml:"enforce_packages,omitempty" json:"enforce_packages,omitempty"`
}

// MaxInstructionLength is the maximum allowed length for team instructions.
// This guards against excessively large instruction blocks that could
// degrade LLM performance.
const MaxInstructionLength = 4096

// ValidateTeam checks the team's fields for correctness.
func ValidateTeam(t *Team) error {
	if t.Name == "" {
		return fmt.Errorf("team: name is required")
	}
	if !teamNamePattern.MatchString(t.Name) {
		return fmt.Errorf("team %q: name contains invalid characters (allowed: alphanumeric, hyphens, underscores)", t.Name)
	}
	if t.Description == "" {
		return fmt.Errorf("team %q: description is required", t.Name)
	}
	if len(t.Personas) == 0 {
		return fmt.Errorf("team %q: at least one persona is required", t.Name)
	}
	if len(t.Instructions) > MaxInstructionLength {
		return fmt.Errorf("team %q: instructions exceed maximum length (%d > %d)", t.Name, len(t.Instructions), MaxInstructionLength)
	}
	if t.MCP != nil {
		if err := t.MCP.Validate(); err != nil {
			return fmt.Errorf("team %q: %w", t.Name, err)
		}
	}
	return nil
}
