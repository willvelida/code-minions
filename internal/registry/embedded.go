package registry

import (
	"errors"
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
	name    string
}

// NewEmbeddedSource creates a Source backed by an embedded filesystem.
// The filesystem must contain packages under a "packages/" directory.
func NewEmbeddedSource(content fs.FS) *EmbeddedSource {
	return &EmbeddedSource{content: content, name: "embedded"}
}

// NewEmbeddedSourceWithName creates a Source backed by an embedded
// filesystem with a custom name. This is used for local directory
// sources that behave like embedded sources but need a distinct name.
func NewEmbeddedSourceWithName(content fs.FS, name string) *EmbeddedSource {
	return &EmbeddedSource{content: content, name: name}
}

func (s *EmbeddedSource) Name() string { return s.name }
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
			return nil, fmt.Errorf("failed to read package %q: %w", entry.Name(), err)
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
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("package %q: %w", name, ErrNotFound)
		}
		return nil, fmt.Errorf("package %q: %w", name, err)
	}

	sub, err := fs.Sub(s.content, pkgDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create sub-filesystem for %q: %w", name, err)
	}

	return sub, nil
}

// ListPersonas returns all personas found under personas/ in the
// embedded filesystem. Each persona directory must contain a
// persona.yaml file.
//
// The directory layout is:
//
//	personas/
//	├── senior-dev/
//	│   └── persona.yaml
//	└── security-reviewer/
//	    └── persona.yaml
//
// If the personas/ directory doesn't exist, an empty list is
// returned (not an error) — it's valid for a source to have no
// personas.
func (s *EmbeddedSource) ListPersonas() ([]model.Persona, error) {
	// Try to read the personas/ directory.
	// If it doesn't exist, that's fine — just return an empty list.
	entries, err := fs.ReadDir(s.content, "personas")
	if err != nil {
		// No personas/ directory is not an error — this source just
		// doesn't have any personas. But real I/O errors should bubble up.
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading personas directory: %w", err)
	}

	var personas []model.Persona
	for _, entry := range entries {
		// Skip files at the top level — we only care about directories
		// (each directory is one persona).
		if !entry.IsDir() {
			continue
		}

		persona, err := s.readPersona(entry.Name())
		if err != nil {
			return nil, fmt.Errorf("failed to read persona %q: %w", entry.Name(), err)
		}
		personas = append(personas, *persona)
	}

	return personas, nil
}

// GetPersona returns the metadata for a named persona by reading
// its persona.yaml file from personas/<name>/persona.yaml.
func (s *EmbeddedSource) GetPersona(name string) (*model.Persona, error) {
	return s.readPersona(name)
}

// readPersona reads and parses the persona.yaml for a named persona.
// Returns ErrNotFound if the persona directory or manifest doesn't exist.
func (s *EmbeddedSource) readPersona(name string) (*model.Persona, error) {
	// Build the path to the persona's manifest file.
	// Example: "personas/senior-dev/persona.yaml"
	manifestPath := "personas/" + name + "/persona.yaml"

	// Try to read the file from the embedded filesystem.
	data, err := fs.ReadFile(s.content, manifestPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("persona %q: %w", name, ErrNotFound)
		}
		return nil, fmt.Errorf("reading persona %q manifest: %w", name, err)
	}

	// Parse the YAML into our Persona struct.
	// The struct fields have `yaml:` tags that tell the parser
	// which YAML key maps to which Go field.
	var persona model.Persona
	if err := yaml.Unmarshal(data, &persona); err != nil {
		return nil, fmt.Errorf("failed to parse persona.yaml for %q: %w", name, err)
	}

	// Safety check: the name in the YAML should match the directory name.
	// If the YAML doesn't specify a name, fill it in from the directory.
	// If the YAML specifies a *different* name, that's an error.
	if persona.Name == "" {
		persona.Name = name
	} else if persona.Name != name {
		return nil, fmt.Errorf("persona.yaml name %q does not match directory name %q", persona.Name, name)
	}

	return &persona, nil
}

// ListTeams returns an empty list — embedded source does not ship teams.
func (s *EmbeddedSource) ListTeams() ([]model.Team, error) {
	return nil, nil
}

// GetTeam returns an error — embedded source does not ship teams.
func (s *EmbeddedSource) GetTeam(name string) (*model.Team, error) {
	return nil, fmt.Errorf("team %q: %w", name, ErrNotFound)
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
			return nil, fmt.Errorf("package %q: %w", name, ErrNotFound)
		}

		// Directory exists but no package.yaml — build minimal package
		return s.buildFallbackPackage(name)
	}

	var pkg model.Package
	if err := yaml.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("failed to parse package.yaml for %q: %w", name, err)
	}

	// Ensure manifest name matches the directory-based package name.
	if pkg.Name == "" {
		pkg.Name = name
	} else if pkg.Name != name {
		return nil, fmt.Errorf("package.yaml name %q does not match directory name %q", pkg.Name, name)
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

	if walkErr := fs.WalkDir(s.content, pkgDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
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
	}); walkErr != nil {
		return nil, fmt.Errorf("failed to walk package %q: %w", name, walkErr)
	}

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
