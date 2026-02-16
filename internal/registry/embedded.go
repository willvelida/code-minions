package registry

import (
	"fmt"
	"io/fs"
	"strings"

	"github.com/willvelida/code-minions/internal/model"
	"gopkg.in/yaml.v3"
)

// EmbeddedSource wraps an embed.FS (or any fs.FS) containing packages
// under a "packages/" directory. This is the default source for
// built-in packages shipped with the CLI binary.
type EmbeddedSource struct {
	content fs.FS
}

// NewEmbeddedSource creates a Source backed by an embedded filesystem.
// The filesystem must contain packages under a "packages/" directory.
func NewEmbeddedSource(content fs.FS) *EmbeddedSource {
	return &EmbeddedSource{content: content}
}

func (s *EmbeddedSource) Name() string { return "embedded" }
func (s *EmbeddedSource) Type() string { return "embedded" }

// ListPackages returns all packages found under packages/ in the
// embedded filesystem. If a package has a package.yaml, its metadata
// is parsed; otherwise a minimal Package is constructed from the
// directory name.
func (s *EmbeddedSource) ListPackages() ([]model.Package, error) {
	entries, err := fs.ReadDir(s.content, "packages")
	if err != nil {
		return nil, fmt.Errorf("failed to read packages directory: %w", err)
	}

	var pkgs []model.Package
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pkg, err := s.readPackage(entry.Name())
		if err != nil {
			// If we can't read the package.yaml, build a minimal package
			// from the directory name for backward compatibility.
			pkg = &model.Package{Name: entry.Name()}
		}
		pkgs = append(pkgs, *pkg)
	}

	return pkgs, nil
}

// GetPackage returns the metadata for a named package.
func (s *EmbeddedSource) GetPackage(name string) (*model.Package, error) {
	return s.readPackage(name)
}

// DownloadPackage returns a sub-filesystem rooted at the package directory.
func (s *EmbeddedSource) DownloadPackage(name string, _ string) (fs.FS, error) {
	pkgDir := "packages/" + name

	// Verify the package exists
	if _, err := fs.Stat(s.content, pkgDir); err != nil {
		return nil, fmt.Errorf("package %q not found: %w", name, err)
	}

	sub, err := fs.Sub(s.content, pkgDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create sub-filesystem for %q: %w", name, err)
	}

	return sub, nil
}

// ListPersonas returns an empty list — embedded source does not
// currently ship personas.
func (s *EmbeddedSource) ListPersonas() ([]model.Persona, error) {
	return nil, nil
}

// GetPersona returns an error — embedded source does not ship personas.
func (s *EmbeddedSource) GetPersona(name string) (*model.Persona, error) {
	return nil, fmt.Errorf("persona %q not found in embedded source", name)
}

// ListTeams returns an empty list — embedded source does not ship teams.
func (s *EmbeddedSource) ListTeams() ([]model.Team, error) {
	return nil, nil
}

// GetTeam returns an error — embedded source does not ship teams.
func (s *EmbeddedSource) GetTeam(name string) (*model.Team, error) {
	return nil, fmt.Errorf("team %q not found in embedded source", name)
}

// Search performs a case-insensitive text search across package names
// and descriptions.
func (s *EmbeddedSource) Search(query string) ([]model.SearchResult, error) {
	pkgs, err := s.ListPackages()
	if err != nil {
		return nil, err
	}

	var results []model.SearchResult
	for _, pkg := range pkgs {
		if matchesQuery(query, pkg.Name, pkg.Description) {
			results = append(results, model.SearchResult{
				Kind:        "package",
				Name:        pkg.Name,
				Description: pkg.Description,
				Version:     pkg.Version,
				Source:      s.Name(),
			})
		}
	}

	return results, nil
}

// readPackage reads and parses the package.yaml for a named package.
// Falls back to directory-name identity if no package.yaml exists.
func (s *EmbeddedSource) readPackage(name string) (*model.Package, error) {
	manifestPath := "packages/" + name + "/package.yaml"

	data, err := fs.ReadFile(s.content, manifestPath)
	if err != nil {
		// Fall back: check if the directory exists at all
		pkgDir := "packages/" + name
		if _, statErr := fs.Stat(s.content, pkgDir); statErr != nil {
			return nil, fmt.Errorf("package %q not found", name)
		}

		// Directory exists but no package.yaml — build minimal package
		return s.buildFallbackPackage(name)
	}

	var pkg model.Package
	if err := yaml.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("failed to parse package.yaml for %q: %w", name, err)
	}

	return &pkg, nil
}

// buildFallbackPackage constructs a minimal Package from the directory
// structure when no package.yaml is present (backward compatibility).
func (s *EmbeddedSource) buildFallbackPackage(name string) (*model.Package, error) {
	pkg := &model.Package{
		Name: name,
	}

	// Try to read description from SKILL.md frontmatter
	skillPath := "packages/" + name + "/skills/" + name + "/SKILL.md"
	data, err := fs.ReadFile(s.content, skillPath)
	if err == nil {
		pkg.Description = extractFrontmatterField(string(data), "description")
		pkg.License = extractFrontmatterField(string(data), "license")
	}

	// Discover contents by walking the package directory
	pkgDir := "packages/" + name
	var agents, skills []string

	_ = fs.WalkDir(s.content, pkgDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}

		// Strip the package prefix to get the relative path
		rel := strings.TrimPrefix(path, pkgDir+"/")

		if strings.HasPrefix(rel, "agents/") {
			agents = append(agents, rel)
		} else if strings.HasPrefix(rel, "skills/") {
			skills = append(skills, rel)
		}
		return nil
	})

	if len(agents) > 0 {
		pkg.Contents.Agents = agents
	}
	if len(skills) > 0 {
		pkg.Contents.Skills = skills
	}

	return pkg, nil
}

// extractFrontmatterField extracts a single field value from YAML frontmatter
// in a Markdown file. Returns empty string if not found.
func extractFrontmatterField(content string, field string) string {
	lines := strings.Split(content, "\n")
	inFrontmatter := false

	for _, line := range lines {
		if strings.TrimSpace(line) == "---" {
			if inFrontmatter {
				return "" // End of frontmatter, field not found
			}
			inFrontmatter = true
			continue
		}

		if inFrontmatter && strings.HasPrefix(line, field+":") {
			val := strings.TrimPrefix(line, field+":")
			val = strings.TrimSpace(val)
			val = strings.Trim(val, "'\"")
			return val
		}
	}

	return ""
}
