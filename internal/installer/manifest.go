package installer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/willvelida/code-minions/internal/model"
)

const (
	// ManifestDir is the directory name for code-minions metadata.
	ManifestDir = ".code-minions"
	// ManifestFile is the filename for the install tracking manifest.
	ManifestFile = "installed.json"
	// ManifestVersion is the current schema version for the manifest.
	ManifestVersion = 1
)

// ManifestPath returns the full path to the installed.json file
// within the given target directory.
func ManifestPath(target string) string {
	return filepath.Join(target, ManifestDir, ManifestFile)
}

// LoadManifest reads the install manifest from the target directory.
// Returns a zero-value manifest if the file does not exist.
func LoadManifest(target string) (*model.InstallManifest, error) {
	path := ManifestPath(target)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &model.InstallManifest{
				Version:  ManifestVersion,
				Packages: []model.InstalledPackage{},
				Personas: []model.InstalledPersona{},
				Teams:    []model.InstalledTeam{},
			}, nil
		}
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest model.InstallManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	return &manifest, nil
}

// SaveManifest writes the install manifest to the target directory.
// Creates the .code-minions/ directory and .gitignore if needed.
func SaveManifest(target string, manifest *model.InstallManifest) error {
	dir := filepath.Join(target, ManifestDir)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create manifest directory: %w", err)
	}

	// Ensure .code-minions is git-ignored
	gitignorePath := filepath.Join(dir, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		if err := os.WriteFile(gitignorePath, []byte("# Auto-generated — do not commit install tracking data\n*\n"), 0644); err != nil {
			return fmt.Errorf("failed to create .gitignore: %w", err)
		}
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	path := ManifestPath(target)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	return nil
}

// RecordInstall adds or updates a package entry in the manifest.
func RecordInstall(manifest *model.InstallManifest, name, version, source, assistant string, files []string) {
	now := time.Now().UTC().Format(time.RFC3339)

	// Update manifest-level fields
	if manifest.InstalledAt == "" {
		manifest.InstalledAt = now
	}
	if assistant != "" {
		manifest.Assistant = assistant
	}

	// Remove existing entry for this package (will be re-added)
	for i, p := range manifest.Packages {
		if p.Name == name {
			manifest.Packages = append(manifest.Packages[:i], manifest.Packages[i+1:]...)
			break
		}
	}

	manifest.Packages = append(manifest.Packages, model.InstalledPackage{
		Name:        name,
		Version:     version,
		Source:      source,
		InstalledAt: now,
		Files:       files,
	})
}

// RecordUninstall removes a package entry from the manifest.
// Returns true if the package was found and removed.
func RecordUninstall(manifest *model.InstallManifest, name string) bool {
	for i, p := range manifest.Packages {
		if p.Name == name {
			manifest.Packages = append(manifest.Packages[:i], manifest.Packages[i+1:]...)
			return true
		}
	}
	return false
}

// FindInstalled returns the installed package entry, or nil if not found.
func FindInstalled(manifest *model.InstallManifest, name string) *model.InstalledPackage {
	for i := range manifest.Packages {
		if manifest.Packages[i].Name == name {
			return &manifest.Packages[i]
		}
	}
	return nil
}
