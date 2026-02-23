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
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Source       string   `json:"source"`
	InstalledAt  string   `json:"installed_at"`
	Persona      string   `json:"persona,omitempty"`
	ReferencedBy []string `json:"referenced_by,omitempty"`
	Files        []string `json:"files"`
	MCPServers   []string `json:"mcp_servers,omitempty"`

	// Direct is true when the package was explicitly requested by the
	// user (via --package flag or code-minions.yml). False when the
	// package was pulled in as a transitive dependency.
	Direct bool `json:"direct"`

	// DependencyOf lists the names of packages that depend on this one.
	// This enables safe uninstall checks ("cannot remove X: still needed
	// by Y") and the `why` command. A package can be both Direct and
	// a DependencyOf another package (user explicitly installs something
	// that is also transitively required).
	DependencyOf []string `json:"dependency_of,omitempty"`
}

// InstalledPersona records a single installed persona.
// It links to the packages it brought in and the grouping files
// it generated (like AGENTS.md sections or persona agent files).
//
// This is the "receipt" for a persona installation — it tells us
// everything we need to know to cleanly uninstall or update later.
type InstalledPersona struct {
	// Name is the persona identifier (e.g. "senior-dev").
	Name string `json:"name"`

	// Version is the persona version at install time.
	Version string `json:"version"`

	// Source identifies where the persona came from (e.g. "embedded").
	Source string `json:"source"`

	// Assistant is the target assistant (e.g. "copilot", "claude").
	Assistant string `json:"assistant"`

	// InstalledAt is the UTC timestamp of installation (RFC 3339).
	InstalledAt string `json:"installed_at"`

	// Team is the team that owns this persona (optional).
	Team string `json:"team,omitempty"`

	// Packages lists the names of packages this persona installed.
	// These names match entries in the manifest's Packages slice,
	// which in turn track the individual files.
	Packages []string `json:"packages"`

	// GeneratedFiles lists the grouping artefacts this persona created.
	// These are assistant-specific files like persona agent definitions
	// or routing table sections. They have checksums so we can detect
	// user modifications before removing them.
	GeneratedFiles []InstalledFile `json:"generated_files,omitempty"`
}

// InstalledFile records a single installed file with its checksum.
// The checksum lets us detect if the user has modified the file since
// installation — important for safe uninstall and update operations.
//
// Think of it like a serial number on a receipt: if the item you're
// returning matches the serial number, the return is automatic.
// If it doesn't match, the store asks if you're sure.
type InstalledFile struct {
	// Path is the relative path where the file was written
	// (e.g. ".github/agents/git-workflow.agent.md").
	Path string `json:"path"`

	// Type describes what kind of file this is (optional).
	// Examples: "agent", "skill", "routing-table", "persona-agent".
	// Empty for regular package files.
	Type string `json:"type,omitempty"`

	// Checksum is the SHA-256 hash of the file contents at install time.
	// Format: "sha256:<hex>" (e.g. "sha256:abc123...").
	Checksum string `json:"checksum"`
}

// InstalledTeam records a single installed team.
type InstalledTeam struct {
	Name                 string   `json:"name"`
	Version              string   `json:"version"`
	Source               string   `json:"source"`
	InstalledAt          string   `json:"installed_at"`
	Personas             []string `json:"personas,omitempty"`
	MCPServers           []string `json:"mcp_servers,omitempty"`
	InstructionsInjected bool     `json:"instructions_injected,omitempty"`
	InstructionsFile     string   `json:"instructions_file,omitempty"`
}
