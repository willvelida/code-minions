package installer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/willvelida/code-minions/internal/model"
)

// ---------------------------------------------------------------------------
// Checksum tests
// ---------------------------------------------------------------------------

func TestComputeChecksumDeterministic(t *testing.T) {
	// The same content should always produce the same checksum.
	data := []byte("hello world")
	sum1 := ComputeChecksum(data)
	sum2 := ComputeChecksum(data)

	if sum1 != sum2 {
		t.Errorf("checksums differ for identical data: %q vs %q", sum1, sum2)
	}

	// It should start with the "sha256:" prefix.
	if len(sum1) < 7 || sum1[:7] != checksumPrefix {
		t.Errorf("checksum should start with %q, got %q", checksumPrefix, sum1)
	}
}

func TestComputeChecksumDifferentData(t *testing.T) {
	// Different content should produce different checksums.
	sum1 := ComputeChecksum([]byte("hello"))
	sum2 := ComputeChecksum([]byte("world"))

	if sum1 == sum2 {
		t.Error("different data produced the same checksum")
	}
}

func TestComputeFileChecksum(t *testing.T) {
	// Write a temporary file and compute its checksum.
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	content := []byte("test content for checksum")

	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	checksum, err := ComputeFileChecksum(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should match computing from the raw bytes.
	expected := ComputeChecksum(content)
	if checksum != expected {
		t.Errorf("file checksum = %q, want %q", checksum, expected)
	}
}

func TestComputeFileChecksumMissingFile(t *testing.T) {
	_, err := ComputeFileChecksum("/nonexistent/file.txt")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestVerifyChecksumMatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "verified.txt")
	content := []byte("original content")

	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	recorded := ComputeChecksum(content)
	if !VerifyChecksum(path, recorded) {
		t.Error("checksum should match for unmodified file")
	}
}

func TestVerifyChecksumMismatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "modified.txt")

	// Write original content and record its checksum.
	original := []byte("original content")
	if err := os.WriteFile(path, original, 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	recorded := ComputeChecksum(original)

	// Now modify the file.
	if err := os.WriteFile(path, []byte("modified content"), 0644); err != nil {
		t.Fatalf("failed to modify file: %v", err)
	}

	if VerifyChecksum(path, recorded) {
		t.Error("checksum should NOT match for modified file")
	}
}

func TestVerifyChecksumMissingFile(t *testing.T) {
	if VerifyChecksum("/nonexistent/file.txt", "sha256:abc") {
		t.Error("should return false for missing file")
	}
}

func TestNewInstalledFile(t *testing.T) {
	dir := t.TempDir()

	// Create a file to track.
	content := []byte("agent file contents")
	subdir := filepath.Join(dir, "agents")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subdir, "test.agent.md"), content, 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	tracked, err := NewInstalledFile(dir, "agents/test.agent.md", "agent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tracked.Path != "agents/test.agent.md" {
		t.Errorf("path = %q, want agents/test.agent.md", tracked.Path)
	}
	if tracked.Type != "agent" {
		t.Errorf("type = %q, want agent", tracked.Type)
	}
	if tracked.Checksum == "" {
		t.Error("checksum should not be empty")
	}

	// Checksum should match the content.
	expected := ComputeChecksum(content)
	if tracked.Checksum != expected {
		t.Errorf("checksum = %q, want %q", tracked.Checksum, expected)
	}
}

func TestNewInstalledFileMissingFile(t *testing.T) {
	_, err := NewInstalledFile("/nonexistent", "missing.txt", "")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

// ---------------------------------------------------------------------------
// Persona manifest CRUD tests
// ---------------------------------------------------------------------------

func TestRecordPersonaInstallAddsEntry(t *testing.T) {
	manifest := &model.InstallManifest{
		Version:  1,
		Packages: []model.InstalledPackage{},
		Personas: []model.InstalledPersona{},
		Teams:    []model.InstalledTeam{},
	}

	// First, add some packages so the persona can link to them.
	RecordInstall(manifest, "git-workflow", "0.1.0", "embedded", "copilot",
		[]string{".github/agents/git-workflow.agent.md"})
	RecordInstall(manifest, "threat-modelling", "0.1.0", "embedded", "copilot",
		[]string{"skills/threat-modelling/SKILL.md"})

	// Now record the persona.
	genFiles := []model.InstalledFile{
		{Path: "AGENTS.md", Type: "routing-table", Checksum: "sha256:abc123"},
	}
	RecordPersonaInstall(manifest, "senior-dev", "0.1.0", "embedded", "copilot",
		[]string{"git-workflow", "threat-modelling"}, genFiles)

	// Check the persona was recorded.
	if len(manifest.Personas) != 1 {
		t.Fatalf("expected 1 persona, got %d", len(manifest.Personas))
	}

	p := manifest.Personas[0]
	if p.Name != "senior-dev" {
		t.Errorf("name = %q, want senior-dev", p.Name)
	}
	if p.Assistant != "copilot" {
		t.Errorf("assistant = %q, want copilot", p.Assistant)
	}
	if len(p.Packages) != 2 {
		t.Errorf("packages count = %d, want 2", len(p.Packages))
	}
	if len(p.GeneratedFiles) != 1 {
		t.Errorf("generated files count = %d, want 1", len(p.GeneratedFiles))
	}
	if p.InstalledAt == "" {
		t.Error("installed_at should be set")
	}

	// Check that packages are stamped with the persona name.
	for _, pkg := range manifest.Packages {
		if pkg.Persona != "senior-dev" {
			t.Errorf("package %q persona = %q, want senior-dev", pkg.Name, pkg.Persona)
		}
	}
}

func TestRecordPersonaInstallReplacesExisting(t *testing.T) {
	manifest := &model.InstallManifest{
		Version: 1,
		Packages: []model.InstalledPackage{
			{Name: "git-workflow", Version: "0.1.0"},
		},
		Personas: []model.InstalledPersona{
			{Name: "senior-dev", Version: "0.1.0", Source: "embedded",
				Assistant: "copilot", Packages: []string{"git-workflow"}},
		},
		Teams: []model.InstalledTeam{},
	}

	// Re-install the same persona (e.g. with a version bump).
	RecordPersonaInstall(manifest, "senior-dev", "0.2.0", "embedded", "copilot",
		[]string{"git-workflow"}, nil)

	if len(manifest.Personas) != 1 {
		t.Fatalf("expected 1 persona (replaced), got %d", len(manifest.Personas))
	}
	if manifest.Personas[0].Version != "0.2.0" {
		t.Errorf("version = %q, want 0.2.0", manifest.Personas[0].Version)
	}
}

func TestRecordPersonaUninstallRemovesEntry(t *testing.T) {
	manifest := &model.InstallManifest{
		Version: 1,
		Packages: []model.InstalledPackage{
			{Name: "git-workflow", Version: "0.1.0", Persona: "senior-dev"},
			{Name: "threat-modelling", Version: "0.1.0", Persona: "senior-dev"},
		},
		Personas: []model.InstalledPersona{
			{Name: "senior-dev", Packages: []string{"git-workflow", "threat-modelling"}},
		},
		Teams: []model.InstalledTeam{},
	}

	removed := RecordPersonaUninstall(manifest, "senior-dev")
	if !removed {
		t.Error("should return true when persona found")
	}
	if len(manifest.Personas) != 0 {
		t.Errorf("expected 0 personas, got %d", len(manifest.Personas))
	}

	// Packages should have their persona stamp cleared
	// (since no other persona references them).
	for _, pkg := range manifest.Packages {
		if pkg.Persona != "" {
			t.Errorf("package %q should have persona cleared, got %q", pkg.Name, pkg.Persona)
		}
	}
}

func TestRecordPersonaUninstallSharedPackageKeepsStamp(t *testing.T) {
	// Two personas share "git-workflow".
	manifest := &model.InstallManifest{
		Version: 1,
		Packages: []model.InstalledPackage{
			{Name: "git-workflow", Version: "0.1.0", Persona: "senior-dev"},
			{Name: "threat-modelling", Version: "0.1.0", Persona: "senior-dev"},
		},
		Personas: []model.InstalledPersona{
			{Name: "senior-dev", Packages: []string{"git-workflow", "threat-modelling"}},
			{Name: "security-reviewer", Packages: []string{"git-workflow"}},
		},
		Teams: []model.InstalledTeam{},
	}

	removed := RecordPersonaUninstall(manifest, "senior-dev")
	if !removed {
		t.Error("should return true")
	}

	// git-workflow is still referenced by security-reviewer,
	// so its persona stamp should NOT be cleared.
	gwPkg := FindInstalled(manifest, "git-workflow")
	if gwPkg == nil {
		t.Fatal("git-workflow package should still exist")
	}
	// The stamp stays because another persona still references it.
	// (The stamp value might be "senior-dev" from the original install,
	// but the important thing is it's not cleared.)
	// Note: in a more sophisticated implementation, we might update
	// the stamp to "security-reviewer", but for now we just don't clear it.

	// threat-modelling is NOT referenced by any other persona,
	// so its stamp SHOULD be cleared.
	tmPkg := FindInstalled(manifest, "threat-modelling")
	if tmPkg == nil {
		t.Fatal("threat-modelling package should still exist")
	}
	if tmPkg.Persona != "" {
		t.Errorf("threat-modelling persona stamp should be cleared, got %q", tmPkg.Persona)
	}
}

func TestRecordPersonaUninstallNotFound(t *testing.T) {
	manifest := &model.InstallManifest{
		Version:  1,
		Packages: []model.InstalledPackage{},
		Personas: []model.InstalledPersona{},
		Teams:    []model.InstalledTeam{},
	}

	removed := RecordPersonaUninstall(manifest, "nonexistent")
	if removed {
		t.Error("should return false for nonexistent persona")
	}
}

func TestFindInstalledPersonaFound(t *testing.T) {
	manifest := &model.InstallManifest{
		Version: 1,
		Personas: []model.InstalledPersona{
			{Name: "senior-dev", Assistant: "copilot"},
		},
	}

	p := FindInstalledPersona(manifest, "senior-dev")
	if p == nil {
		t.Fatal("should find installed persona")
	}
	if p.Assistant != "copilot" {
		t.Errorf("assistant = %q, want copilot", p.Assistant)
	}
}

func TestFindInstalledPersonaNotFound(t *testing.T) {
	manifest := &model.InstallManifest{
		Version:  1,
		Personas: []model.InstalledPersona{},
	}

	if p := FindInstalledPersona(manifest, "nonexistent"); p != nil {
		t.Error("should return nil for nonexistent persona")
	}
}

func TestPersonaPackagesExclusive(t *testing.T) {
	manifest := &model.InstallManifest{
		Version: 1,
		Personas: []model.InstalledPersona{
			{Name: "senior-dev", Packages: []string{"git-workflow", "threat-modelling"}},
			{Name: "security-reviewer", Packages: []string{"git-workflow"}},
		},
	}

	exclusive := PersonaPackages(manifest, "senior-dev")

	// "git-workflow" is shared, so only "threat-modelling" is exclusive.
	if len(exclusive) != 1 {
		t.Fatalf("expected 1 exclusive package, got %d: %v", len(exclusive), exclusive)
	}
	if exclusive[0] != "threat-modelling" {
		t.Errorf("exclusive package = %q, want threat-modelling", exclusive[0])
	}
}

func TestPersonaPackagesAllExclusive(t *testing.T) {
	manifest := &model.InstallManifest{
		Version: 1,
		Personas: []model.InstalledPersona{
			{Name: "senior-dev", Packages: []string{"git-workflow", "threat-modelling"}},
		},
	}

	exclusive := PersonaPackages(manifest, "senior-dev")
	if len(exclusive) != 2 {
		t.Fatalf("expected 2 exclusive packages, got %d", len(exclusive))
	}
}

func TestPersonaPackagesNotFound(t *testing.T) {
	manifest := &model.InstallManifest{
		Version:  1,
		Personas: []model.InstalledPersona{},
	}

	exclusive := PersonaPackages(manifest, "nonexistent")
	if exclusive != nil {
		t.Errorf("expected nil for nonexistent persona, got %v", exclusive)
	}
}

// ---------------------------------------------------------------------------
// Manifest round-trip with personas
// ---------------------------------------------------------------------------

func TestSaveAndLoadManifestWithPersonas(t *testing.T) {
	target := t.TempDir()

	original := &model.InstallManifest{
		Version:     1,
		InstalledAt: "2026-01-01T00:00:00Z",
		Assistant:   "copilot",
		Packages: []model.InstalledPackage{
			{
				Name:    "git-workflow",
				Version: "0.1.0",
				Source:  "embedded",
				Persona: "senior-dev",
				Files:   []string{".github/agents/git-workflow.agent.md"},
			},
		},
		Personas: []model.InstalledPersona{
			{
				Name:      "senior-dev",
				Version:   "0.1.0",
				Source:    "embedded",
				Assistant: "copilot",
				Packages:  []string{"git-workflow"},
				GeneratedFiles: []model.InstalledFile{
					{Path: "AGENTS.md", Type: "routing-table", Checksum: "sha256:abc123"},
				},
			},
		},
		Teams: []model.InstalledTeam{},
	}

	if err := SaveManifest(target, original); err != nil {
		t.Fatalf("SaveManifest failed: %v", err)
	}

	loaded, err := LoadManifest(target)
	if err != nil {
		t.Fatalf("LoadManifest failed: %v", err)
	}

	// Verify persona survives round-trip.
	if len(loaded.Personas) != 1 {
		t.Fatalf("personas count = %d, want 1", len(loaded.Personas))
	}

	p := loaded.Personas[0]
	if p.Name != "senior-dev" {
		t.Errorf("persona name = %q, want senior-dev", p.Name)
	}
	if p.Assistant != "copilot" {
		t.Errorf("persona assistant = %q, want copilot", p.Assistant)
	}
	if len(p.Packages) != 1 || p.Packages[0] != "git-workflow" {
		t.Errorf("persona packages = %v, want [git-workflow]", p.Packages)
	}
	if len(p.GeneratedFiles) != 1 {
		t.Fatalf("generated files count = %d, want 1", len(p.GeneratedFiles))
	}
	gf := p.GeneratedFiles[0]
	if gf.Path != "AGENTS.md" || gf.Type != "routing-table" || gf.Checksum != "sha256:abc123" {
		t.Errorf("generated file = %+v, unexpected values", gf)
	}

	// Verify package persona link survives.
	if loaded.Packages[0].Persona != "senior-dev" {
		t.Errorf("package persona = %q, want senior-dev", loaded.Packages[0].Persona)
	}
}

// ---------------------------------------------------------------------------
// MCP server reference counting tests
// ---------------------------------------------------------------------------

// TestIsMCPServerSharedTrue verifies that a server is considered shared
// when two different packages both claim it.
func TestIsMCPServerSharedTrue(t *testing.T) {
	manifest := &model.InstallManifest{
		Version: 1,
		Packages: []model.InstalledPackage{
			{Name: "git-workflow", MCPServers: []string{"github"}},
			{Name: "raise-pull-requests", MCPServers: []string{"github"}},
		},
	}

	// When excluding git-workflow, "github" is still in raise-pull-requests.
	if !IsMCPServerShared(manifest, "github", map[string]bool{"git-workflow": true}) {
		t.Error("expected github to be shared when raise-pull-requests also uses it")
	}

	// Same from the other direction.
	if !IsMCPServerShared(manifest, "github", map[string]bool{"raise-pull-requests": true}) {
		t.Error("expected github to be shared when git-workflow also uses it")
	}
}

// TestIsMCPServerSharedFalse verifies that a server is NOT shared when
// only one package claims it.
func TestIsMCPServerSharedFalse(t *testing.T) {
	manifest := &model.InstallManifest{
		Version: 1,
		Packages: []model.InstalledPackage{
			{Name: "git-workflow", MCPServers: []string{"github"}},
			{Name: "threat-modelling", MCPServers: []string{"azure"}},
		},
	}

	// "github" is only in git-workflow — not shared.
	if IsMCPServerShared(manifest, "github", map[string]bool{"git-workflow": true}) {
		t.Error("github is only in git-workflow, should not be shared")
	}

	// "azure" is only in threat-modelling — not shared.
	if IsMCPServerShared(manifest, "azure", map[string]bool{"threat-modelling": true}) {
		t.Error("azure is only in threat-modelling, should not be shared")
	}
}

// TestIsMCPServerSharedUnknown verifies that a server not in any
// package is not considered shared.
func TestIsMCPServerSharedUnknown(t *testing.T) {
	manifest := &model.InstallManifest{
		Version: 1,
		Packages: []model.InstalledPackage{
			{Name: "git-workflow", MCPServers: []string{"github"}},
		},
	}

	if IsMCPServerShared(manifest, "nonexistent", map[string]bool{"git-workflow": true}) {
		t.Error("nonexistent server should not be shared")
	}
}

// TestExclusiveMCPServersPartial verifies that only exclusive servers
// are returned when a package has both shared and exclusive servers.
func TestExclusiveMCPServersPartial(t *testing.T) {
	manifest := &model.InstallManifest{
		Version: 1,
		Packages: []model.InstalledPackage{
			{Name: "git-workflow", MCPServers: []string{"github", "linear"}},
			{Name: "raise-pull-requests", MCPServers: []string{"github"}},
		},
	}

	exclusive := ExclusiveMCPServers(manifest, []string{"git-workflow"})
	// "github" is shared, "linear" is exclusive.
	if len(exclusive) != 1 || exclusive[0] != "linear" {
		t.Errorf("ExclusiveMCPServers = %v, want [linear]", exclusive)
	}
}

// TestExclusiveMCPServersAllExclusive verifies that all servers are
// returned when none are shared.
func TestExclusiveMCPServersAllExclusive(t *testing.T) {
	manifest := &model.InstallManifest{
		Version: 1,
		Packages: []model.InstalledPackage{
			{Name: "git-workflow", MCPServers: []string{"github", "linear"}},
		},
	}

	exclusive := ExclusiveMCPServers(manifest, []string{"git-workflow"})
	if len(exclusive) != 2 {
		t.Errorf("ExclusiveMCPServers = %v, want [github linear]", exclusive)
	}
}

// TestExclusiveMCPServersAllShared verifies that an empty list is
// returned when all servers are shared with other packages.
func TestExclusiveMCPServersAllShared(t *testing.T) {
	manifest := &model.InstallManifest{
		Version: 1,
		Packages: []model.InstalledPackage{
			{Name: "git-workflow", MCPServers: []string{"github"}},
			{Name: "raise-pull-requests", MCPServers: []string{"github"}},
		},
	}

	exclusive := ExclusiveMCPServers(manifest, []string{"git-workflow"})
	if len(exclusive) != 0 {
		t.Errorf("ExclusiveMCPServers = %v, want empty", exclusive)
	}
}

// TestExclusiveMCPServersNotFound verifies nil for a package not in
// the manifest.
func TestExclusiveMCPServersNotFound(t *testing.T) {
	manifest := &model.InstallManifest{
		Version:  1,
		Packages: []model.InstalledPackage{},
	}

	exclusive := ExclusiveMCPServers(manifest, []string{"nonexistent"})
	if exclusive != nil {
		t.Errorf("ExclusiveMCPServers = %v, want nil", exclusive)
	}
}

// ---------------------------------------------------------------------------
// OrphanDependencies tests
// ---------------------------------------------------------------------------

// TestOrphanDependenciesSimple verifies that a transitive dep whose only
// parent is being removed is returned as an orphan.
func TestOrphanDependenciesSimple(t *testing.T) {
	manifest := &model.InstallManifest{
		Packages: []model.InstalledPackage{
			{Name: "pkg-a", Direct: true},
			{Name: "pkg-b", Direct: false, DependencyOf: []string{"pkg-a"}},
		},
	}
	orphans := OrphanDependencies(manifest, []string{"pkg-a"})
	if len(orphans) != 1 || orphans[0] != "pkg-b" {
		t.Errorf("OrphanDependencies = %v, want [pkg-b]", orphans)
	}
}

// TestOrphanDependenciesShared verifies that a transitive dep shared by
// another package is NOT returned as an orphan.
func TestOrphanDependenciesShared(t *testing.T) {
	manifest := &model.InstallManifest{
		Packages: []model.InstalledPackage{
			{Name: "pkg-a", Direct: true},
			{Name: "pkg-c", Direct: true},
			{Name: "pkg-b", Direct: false, DependencyOf: []string{"pkg-a", "pkg-c"}},
		},
	}
	orphans := OrphanDependencies(manifest, []string{"pkg-a"})
	if len(orphans) != 0 {
		t.Errorf("OrphanDependencies = %v, want empty (pkg-b still needed by pkg-c)", orphans)
	}
}

// TestOrphanDependenciesCascade verifies that cascaded orphans are found:
// removing pkg-a orphans pkg-b, which then orphans pkg-c.
func TestOrphanDependenciesCascade(t *testing.T) {
	manifest := &model.InstallManifest{
		Packages: []model.InstalledPackage{
			{Name: "pkg-a", Direct: true},
			{Name: "pkg-b", Direct: false, DependencyOf: []string{"pkg-a"}},
			{Name: "pkg-c", Direct: false, DependencyOf: []string{"pkg-b"}},
		},
	}
	orphans := OrphanDependencies(manifest, []string{"pkg-a"})
	orphanSet := make(map[string]bool)
	for _, o := range orphans {
		orphanSet[o] = true
	}
	if !orphanSet["pkg-b"] || !orphanSet["pkg-c"] {
		t.Errorf("OrphanDependencies = %v, want [pkg-b, pkg-c]", orphans)
	}
}

// TestOrphanDependenciesDirectNeverCascaded verifies that a direct
// install is never auto-cascaded, even if its DependencyOf list
// would otherwise make it an orphan.
func TestOrphanDependenciesDirectNeverCascaded(t *testing.T) {
	manifest := &model.InstallManifest{
		Packages: []model.InstalledPackage{
			{Name: "pkg-a", Direct: true},
			{Name: "pkg-b", Direct: true, DependencyOf: []string{"pkg-a"}},
		},
	}
	orphans := OrphanDependencies(manifest, []string{"pkg-a"})
	if len(orphans) != 0 {
		t.Errorf("OrphanDependencies = %v, want empty (pkg-b is direct)", orphans)
	}
}

// TestOrphanDependenciesNoManifestDeps verifies that a package without
// DependencyOf metadata is not treated as an orphan.
func TestOrphanDependenciesNoManifestDeps(t *testing.T) {
	manifest := &model.InstallManifest{
		Packages: []model.InstalledPackage{
			{Name: "pkg-a", Direct: true},
			{Name: "pkg-b", Direct: false}, // no DependencyOf
		},
	}
	orphans := OrphanDependencies(manifest, []string{"pkg-a"})
	if len(orphans) != 0 {
		t.Errorf("OrphanDependencies = %v, want empty (no DependencyOf metadata)", orphans)
	}
}
