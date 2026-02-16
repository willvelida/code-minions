package model

// SearchResult represents a search hit from a source.
type SearchResult struct {
	Kind        string `json:"kind"` // "package", "persona", "team"
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Source      string `json:"source"` // Source name
}
