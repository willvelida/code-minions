package model

// InstallManifest tracks what's installed in a project.
// Stored at .code-minions/installed.json (git-ignored by default).
type InstallManifest struct {
	Version     int                `json:"version"`
	InstalledAt string             `json:"installed_at"`
	Assistant   string             `json:"assistant"`
	Packages    []InstalledPackage `json:"packages"`
	Personas    []InstalledPersona `json:"personas"`
	Teams       []InstalledTeam    `json:"teams"`
}

// InstalledPackage records a single installed package.
type InstalledPackage struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Source      string   `json:"source"`
	InstalledAt string   `json:"installed_at"`
	Persona     string   `json:"persona,omitempty"`
	Files       []string `json:"files"`
}

// InstalledPersona records a single installed persona.
type InstalledPersona struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Source      string `json:"source"`
	InstalledAt string `json:"installed_at"`
	Team        string `json:"team,omitempty"`
}

// InstalledTeam records a single installed team.
type InstalledTeam struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Source      string `json:"source"`
	InstalledAt string `json:"installed_at"`
}
