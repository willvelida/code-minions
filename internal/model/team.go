package model

// Team is a named grouping of personas, allowing team leads to
// standardise which coding assistant configurations their team uses.
type Team struct {
	Name        string       `yaml:"name" json:"name"`
	Description string       `yaml:"description" json:"description"`
	Author      string       `yaml:"author,omitempty" json:"author,omitempty"`
	Personas    []PersonaRef `yaml:"personas" json:"personas"`
	Config      TeamConfig   `yaml:"config,omitempty" json:"config,omitempty"`
}

// PersonaRef is a reference to a persona within a team.
type PersonaRef struct {
	Name    string `yaml:"name" json:"name"`
	Version string `yaml:"version,omitempty" json:"version,omitempty"`
	Source  string `yaml:"source,omitempty" json:"source,omitempty"`
}

// TeamConfig holds team-level settings.
type TeamConfig struct {
	DefaultAssistant string `yaml:"default_assistant,omitempty" json:"default_assistant,omitempty"`
	EnforcePackages  bool   `yaml:"enforce_packages,omitempty" json:"enforce_packages,omitempty"`
}
