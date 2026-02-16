package model

// SearchResult represents a search hit from a source.
type SearchResult struct {
	Kind        string `yaml:"kind" json:"kind"` // "package", "persona", "team"
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	Version     string `yaml:"version" json:"version"`
	Source      string `yaml:"source" json:"source"` // Source name
}
