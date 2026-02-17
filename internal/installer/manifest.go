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

// ---------------------------------------------------------------------------
// Persona manifest operations
// ---------------------------------------------------------------------------
// These functions mirror the package-level CRUD above but operate on
// persona entries. A persona entry records:
//
//   - Which persona was installed
//   - For which assistant (copilot, claude, etc.)
//   - Which packages it brought in (by name)
//   - Which grouping files it generated (with checksums)
//
// Think of it as adding a "project receipt" alongside the individual
// "item receipts" (packages) we already track.

// RecordPersonaInstall adds or updates a persona entry in the manifest.
// It also stamps each package entry with the persona name so we can
// trace which persona brought in which package.
//
// Parameters:
//   - manifest: the manifest to modify (loaded with LoadManifest)
//   - name: persona name (e.g. "senior-dev")
//   - version: persona version (e.g. "0.1.0")
//   - source: where it came from (e.g. "embedded")
//   - assistant: target assistant (e.g. "copilot")
//   - packages: list of package names this persona installed
//   - generatedFiles: the grouping artefacts created (with checksums)
func RecordPersonaInstall(
	manifest *model.InstallManifest,
	name, version, source, assistant string,
	packages []string,
	generatedFiles []model.InstalledFile,
) {
	now := time.Now().UTC().Format(time.RFC3339)

	// Remove any existing entry for this persona (will be re-added fresh).
	// This handles re-installs cleanly.
	for i, p := range manifest.Personas {
		if p.Name == name {
			manifest.Personas = append(manifest.Personas[:i], manifest.Personas[i+1:]...)
			break
		}
	}

	// Add the persona entry.
	manifest.Personas = append(manifest.Personas, model.InstalledPersona{
		Name:           name,
		Version:        version,
		Source:         source,
		Assistant:      assistant,
		InstalledAt:    now,
		Packages:       packages,
		GeneratedFiles: generatedFiles,
	})

	// Stamp each package entry with the persona name.
	// This lets us know "git-workflow was installed by senior-dev".
	for _, pkgName := range packages {
		if entry := FindInstalled(manifest, pkgName); entry != nil {
			entry.Persona = name
		}
	}
}

// RecordPersonaUninstall removes a persona entry from the manifest.
// It also clears the persona stamp from any packages that were
// exclusively owned by this persona.
//
// Returns true if the persona was found and removed.
func RecordPersonaUninstall(manifest *model.InstallManifest, name string) bool {
	idx := -1
	for i, p := range manifest.Personas {
		if p.Name == name {
			idx = i
			break
		}
	}

	if idx == -1 {
		return false
	}

	// Get the package list before removing.
	packages := manifest.Personas[idx].Packages

	// Remove the persona entry.
	manifest.Personas = append(manifest.Personas[:idx], manifest.Personas[idx+1:]...)

	// For each package this persona owned, check if any OTHER persona
	// still references it. If not, clear the persona stamp.
	for _, pkgName := range packages {
		if !isPackageReferencedByAnyPersona(manifest, pkgName) {
			if entry := FindInstalled(manifest, pkgName); entry != nil {
				entry.Persona = ""
			}
		}
	}

	return true
}

// FindInstalledPersona returns the installed persona entry, or nil if not found.
func FindInstalledPersona(manifest *model.InstallManifest, name string) *model.InstalledPersona {
	for i := range manifest.Personas {
		if manifest.Personas[i].Name == name {
			return &manifest.Personas[i]
		}
	}
	return nil
}

// PersonaPackages returns the names of packages owned exclusively by
// this persona (not shared with any other persona). These are safe
// to remove during uninstall without affecting other personas.
func PersonaPackages(manifest *model.InstallManifest, personaName string) []string {
	persona := FindInstalledPersona(manifest, personaName)
	if persona == nil {
		return nil
	}

	var exclusive []string
	for _, pkgName := range persona.Packages {
		shared := false
		for _, other := range manifest.Personas {
			if other.Name == personaName {
				continue
			}
			for _, otherPkg := range other.Packages {
				if otherPkg == pkgName {
					shared = true
					break
				}
			}
			if shared {
				break
			}
		}
		if !shared {
			exclusive = append(exclusive, pkgName)
		}
	}

	return exclusive
}

// isPackageReferencedByAnyPersona checks whether any persona in the
// manifest still references the given package name.
func isPackageReferencedByAnyPersona(manifest *model.InstallManifest, pkgName string) bool {
	for _, persona := range manifest.Personas {
		for _, pkg := range persona.Packages {
			if pkg == pkgName {
				return true
			}
		}
	}
	return false
}

// IsMCPServerShared checks whether any installed package OTHER than
// excludePackage also claims the given MCP server name.
//
// This is the MCP equivalent of isPackageReferencedByAnyPersona —
// it enables safe uninstall by answering: "if I remove these packages,
// will another package still need this server?"
//
// excludePackages is a set of package names to ignore (typically all
// packages being uninstalled in the current operation). This prevents
// packages that are all being removed together from blocking each
// other's server cleanup.
//
// Example: packages "git-workflow" and "raise-pull-requests" both
// declare a "github" MCP server. When uninstalling both at once,
// IsMCPServerShared(manifest, "github", {"git-workflow","raise-pull-requests"})
// returns false because no OTHER package claims it.
func IsMCPServerShared(manifest *model.InstallManifest, serverName string, excludePackages map[string]bool) bool {
	for _, pkg := range manifest.Packages {
		if excludePackages[pkg.Name] {
			continue
		}
		for _, s := range pkg.MCPServers {
			if s == serverName {
				return true
			}
		}
	}
	return false
}

// ExclusiveMCPServers returns the MCP server names from the given
// packages that are NOT claimed by any other installed package outside
// the set. These are the only servers safe to remove from the
// assistant's config file.
//
// pkgNames should include ALL packages being removed in a single
// operation (e.g. all exclusive packages in a persona uninstall).
func ExclusiveMCPServers(manifest *model.InstallManifest, pkgNames []string) []string {
	excludeSet := make(map[string]bool, len(pkgNames))
	for _, n := range pkgNames {
		excludeSet[n] = true
	}

	seen := make(map[string]bool)
	var exclusive []string
	for _, pkgName := range pkgNames {
		pkg := FindInstalled(manifest, pkgName)
		if pkg == nil {
			continue
		}
		for _, serverName := range pkg.MCPServers {
			if seen[serverName] {
				continue // already collected
			}
			seen[serverName] = true
			if !IsMCPServerShared(manifest, serverName, excludeSet) {
				exclusive = append(exclusive, serverName)
			}
		}
	}
	return exclusive
}

// ---------------------------------------------------------------------------
// Team manifest operations
// ---------------------------------------------------------------------------
// These functions extend the manifest to track team-level configuration:
// team MCP servers, team instructions, and per-persona reference counting.

// RecordTeamInstall adds or updates a team entry in the manifest.
// It stamps each persona entry with the team name (for traceability)
// and records the team's MCP servers and instructions state.
//
// Parameters:
//   - manifest: the manifest to modify
//   - name: team name (e.g. "platform-engineering")
//   - version: team version
//   - source: where it came from (e.g. "embedded")
//   - assistant: target assistant (e.g. "copilot")
//   - personaNames: personas this team installed
//   - mcpServers: MCP server names the team contributed
//   - instructionsInjected: whether team instructions were written
//   - instructionsFile: relative path of the instructions file
func RecordTeamInstall(
	manifest *model.InstallManifest,
	name, version, source, assistant string,
	personaNames, mcpServers []string,
	instructionsInjected bool,
	instructionsFile string,
) {
	now := time.Now().UTC().Format(time.RFC3339)

	if manifest.InstalledAt == "" {
		manifest.InstalledAt = now
	}
	if assistant != "" {
		manifest.Assistant = assistant
	}

	// Clean up stale references from any existing entry before reinstall.
	// This ensures personas and packages from the old team definition
	// don't retain stale stamps when the team is reinstalled with
	// different personas or packages.
	cleanupTeamReferences(manifest, name)

	// Remove the old entry.
	for i, t := range manifest.Teams {
		if t.Name == name {
			manifest.Teams = append(manifest.Teams[:i], manifest.Teams[i+1:]...)
			break
		}
	}

	manifest.Teams = append(manifest.Teams, model.InstalledTeam{
		Name:                 name,
		Version:              version,
		Source:               source,
		InstalledAt:          now,
		Personas:             personaNames,
		MCPServers:           mcpServers,
		InstructionsInjected: instructionsInjected,
		InstructionsFile:     instructionsFile,
	})

	// Stamp each persona with the team name.
	for _, pName := range personaNames {
		if p := FindInstalledPersona(manifest, pName); p != nil {
			p.Team = name
		}
	}

	// Add team to ReferencedBy on packages brought in through personas.
	for _, pName := range personaNames {
		p := FindInstalledPersona(manifest, pName)
		if p == nil {
			continue
		}
		for _, pkgName := range p.Packages {
			pkg := FindInstalled(manifest, pkgName)
			if pkg == nil {
				continue
			}
			if !stringSliceContains(pkg.ReferencedBy, name) {
				pkg.ReferencedBy = append(pkg.ReferencedBy, name)
			}
		}
	}
}

// RecordTeamUninstall removes a team entry from the manifest.
// Returns true if the team was found and removed.
func RecordTeamUninstall(manifest *model.InstallManifest, name string) bool {
	idx := -1
	for i, t := range manifest.Teams {
		if t.Name == name {
			idx = i
			break
		}
	}
	if idx == -1 {
		return false
	}

	cleanupTeamReferences(manifest, name)
	manifest.Teams = append(manifest.Teams[:idx], manifest.Teams[idx+1:]...)
	return true
}

// cleanupTeamReferences clears persona team stamps and package
// ReferencedBy entries for the named team. Used both during uninstall
// and before reinstall to avoid stale references.
func cleanupTeamReferences(manifest *model.InstallManifest, name string) {
	team := FindInstalledTeam(manifest, name)
	if team == nil {
		return
	}

	// Clear team stamps from personas.
	for _, pName := range team.Personas {
		if p := FindInstalledPersona(manifest, pName); p != nil {
			if p.Team == name {
				p.Team = ""
			}
		}
	}

	// Remove team from ReferencedBy on packages.
	for _, pName := range team.Personas {
		p := FindInstalledPersona(manifest, pName)
		if p == nil {
			continue
		}
		for _, pkgName := range p.Packages {
			pkg := FindInstalled(manifest, pkgName)
			if pkg == nil {
				continue
			}
			pkg.ReferencedBy = removeFromSlice(pkg.ReferencedBy, name)
		}
	}
}

// FindInstalledTeam returns the installed team entry, or nil if not found.
func FindInstalledTeam(manifest *model.InstallManifest, name string) *model.InstalledTeam {
	for i := range manifest.Teams {
		if manifest.Teams[i].Name == name {
			return &manifest.Teams[i]
		}
	}
	return nil
}

// TeamExclusivePersonas returns persona names that belong ONLY to this
// team and not to any other team. These are safe to uninstall when
// removing the team without affecting other teams.
func TeamExclusivePersonas(manifest *model.InstallManifest, teamName string) []string {
	team := FindInstalledTeam(manifest, teamName)
	if team == nil {
		return nil
	}

	var exclusive []string
	for _, pName := range team.Personas {
		shared := false
		for _, other := range manifest.Teams {
			if other.Name == teamName {
				continue
			}
			for _, otherPersona := range other.Personas {
				if otherPersona == pName {
					shared = true
					break
				}
			}
			if shared {
				break
			}
		}
		if !shared {
			exclusive = append(exclusive, pName)
		}
	}
	return exclusive
}

// TeamExclusiveMCPServers returns MCP server names that belong ONLY
// to this team and are not claimed by any other team or any package
// outside the team's personas. These are safe to remove from the
// assistant's MCP config when uninstalling the team.
func TeamExclusiveMCPServers(manifest *model.InstallManifest, teamName string) []string {
	team := FindInstalledTeam(manifest, teamName)
	if team == nil {
		return nil
	}

	// Build the set of packages owned by this team's personas.
	teamPkgs := make(map[string]bool)
	for _, pName := range team.Personas {
		if p := FindInstalledPersona(manifest, pName); p != nil {
			for _, pkg := range p.Packages {
				teamPkgs[pkg] = true
			}
		}
	}

	var exclusive []string
	for _, serverName := range team.MCPServers {
		shared := false
		// Check other teams.
		for _, other := range manifest.Teams {
			if other.Name == teamName {
				continue
			}
			for _, s := range other.MCPServers {
				if s == serverName {
					shared = true
					break
				}
			}
			if shared {
				break
			}
		}
		if shared {
			continue
		}
		// Check packages outside this team.
		if IsMCPServerShared(manifest, serverName, teamPkgs) {
			continue
		}
		exclusive = append(exclusive, serverName)
	}
	return exclusive
}

// stringSliceContains checks if a string slice contains the given value.
func stringSliceContains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}

// removeFromSlice returns a new slice with the first occurrence of val removed.
// It creates a new backing array to avoid mutating the original slice.
func removeFromSlice(slice []string, val string) []string {
	for i, s := range slice {
		if s == val {
			result := make([]string, 0, len(slice)-1)
			result = append(result, slice[:i]...)
			result = append(result, slice[i+1:]...)
			return result
		}
	}
	return slice
}
