package registry

import (
	"io/fs"

	"github.com/willvelida/code-minions/internal/model"
)

// Source is the interface for package, persona, and team discovery
// and retrieval. Implementations include embedded (embed.FS), git,
// clawhub, and generic URL sources.
type Source interface {
	// Name returns a human-readable identifier for this source.
	Name() string

	// Type returns the source type (embedded, git, clawhub, url).
	Type() string

	// ListPackages returns all available packages from this source.
	ListPackages() ([]model.Package, error)

	// GetPackage returns the metadata for a specific package.
	GetPackage(name string) (*model.Package, error)

	// DownloadPackage downloads a package's content and returns a
	// filesystem rooted at the package directory.
	DownloadPackage(name string, version string) (fs.FS, error)

	// ListPersonas returns all available personas from this source.
	ListPersonas() ([]model.Persona, error)

	// GetPersona returns a specific persona.
	GetPersona(name string) (*model.Persona, error)

	// ListTeams returns all available teams from this source.
	ListTeams() ([]model.Team, error)

	// GetTeam returns a specific team.
	GetTeam(name string) (*model.Team, error)

	// Search performs a text search across packages, personas, and teams.
	Search(query string) ([]model.SearchResult, error)
}
