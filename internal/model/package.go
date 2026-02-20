package model

// Package represents a code-minions package with its metadata and contents.
// Every package (embedded or external) is described by a package.yaml manifest.
type Package struct {
	Name          string          `yaml:"name" json:"name"`
	Version       string          `yaml:"version" json:"version"`
	Description   string          `yaml:"description" json:"description"`
	Author        string          `yaml:"author,omitempty" json:"author,omitempty"`
	License       string          `yaml:"license,omitempty" json:"license,omitempty"`
	Contents      PackageContents `yaml:"contents,omitempty" json:"contents"`
	Compatibility []string        `yaml:"compatibility,omitempty" json:"compatibility,omitempty"`
	Dependencies  []DependencyRef `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`

	// Source is the name of the registry source this package came from.
	// Populated at runtime by Registry.ListPackages — not persisted in
	// package.yaml.
	Source string `yaml:"-" json:"source,omitempty"`
}

// PackageContents declares what files a package provides.
type PackageContents struct {
	Agents       []string `yaml:"agents,omitempty" json:"agents,omitempty"`
	Skills       []string `yaml:"skills,omitempty" json:"skills,omitempty"`
	MCP          []string `yaml:"mcp,omitempty" json:"mcp,omitempty"`
	Standards    []string `yaml:"standards,omitempty" json:"standards,omitempty"`
	Actions      []string `yaml:"actions,omitempty" json:"actions,omitempty"`
	Instructions []string `yaml:"instructions,omitempty" json:"instructions,omitempty"`
}

// DependencyRef is a reference to another package.
type DependencyRef struct {
	Name    string `yaml:"name" json:"name"`
	Version string `yaml:"version,omitempty" json:"version,omitempty"` // Semver constraint, e.g. ">=0.1.0"
	Source  string `yaml:"source,omitempty" json:"source,omitempty"`   // Source name, e.g. "clawhub", "github"
}
