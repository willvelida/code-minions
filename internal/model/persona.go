package model

// Persona is a named grouping of packages that forms a coherent
// coding assistant identity (e.g. "senior-dev" = git-workflow +
// raise-pull-requests + threat-modelling).
type Persona struct {
	Name         string       `yaml:"name" json:"name"`
	Description  string       `yaml:"description" json:"description"`
	Author       string       `yaml:"author,omitempty" json:"author,omitempty"`
	Packages     []PackageRef `yaml:"packages" json:"packages"`
	Instructions string       `yaml:"instructions,omitempty" json:"instructions,omitempty"`
}

// PackageRef is a reference to a package within a persona or team.
type PackageRef struct {
	Name    string `yaml:"name" json:"name"`
	Version string `yaml:"version,omitempty" json:"version,omitempty"`
	Source  string `yaml:"source,omitempty" json:"source,omitempty"`
}
