package installer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/willvelida/code-minions/internal/model"
)

func TestManifestPathReturnsExpectedPath(t *testing.T) {
	path := ManifestPath("/project")
	want := filepath.Join("/project", ".code-minions", "installed.json")
	if path != want {
		t.Errorf("got %q, want %q", path, want)
	}
}

func TestLoadManifestReturnsEmptyWhenNoFile(t *testing.T) {
	target := t.TempDir()

	manifest, err := LoadManifest(target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if manifest.Version != ManifestVersion {
		t.Errorf("version = %d, want %d", manifest.Version, ManifestVersion)
	}
	if len(manifest.Packages) != 0 {
		t.Errorf("packages should be empty, got %d", len(manifest.Packages))
	}
}

func TestSaveAndLoadManifestRoundTrip(t *testing.T) {
	target := t.TempDir()

	original := &model.InstallManifest{
		Version:     1,
		InstalledAt: "2026-01-01T00:00:00Z",
		Assistant:   "copilot",
		Packages: []model.InstalledPackage{
			{
				Name:        "git-workflow",
				Version:     "0.1.0",
				Source:      "embedded",
				InstalledAt: "2026-01-01T00:00:00Z",
				Files:       []string{"agents/git-workflow.agent.md"},
			},
		},
		Personas: []model.InstalledPersona{},
		Teams:    []model.InstalledTeam{},
	}

	if err := SaveManifest(target, original); err != nil {
		t.Fatalf("SaveManifest failed: %v", err)
	}

	// Verify .gitignore was created
	gitignorePath := filepath.Join(target, ManifestDir, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		t.Error(".gitignore should be created in .code-minions/")
	}

	loaded, err := LoadManifest(target)
	if err != nil {
		t.Fatalf("LoadManifest failed: %v", err)
	}

	if loaded.Assistant != "copilot" {
		t.Errorf("assistant = %q, want copilot", loaded.Assistant)
	}
	if len(loaded.Packages) != 1 {
		t.Fatalf("packages count = %d, want 1", len(loaded.Packages))
	}
	if loaded.Packages[0].Name != "git-workflow" {
		t.Errorf("package name = %q, want git-workflow", loaded.Packages[0].Name)
	}
}

func TestRecordInstallAddsPackage(t *testing.T) {
	manifest := &model.InstallManifest{
		Version:  1,
		Packages: []model.InstalledPackage{},
		Personas: []model.InstalledPersona{},
		Teams:    []model.InstalledTeam{},
	}

	RecordInstall(manifest, "git-workflow", "0.1.0", "embedded", "copilot",
		[]string{"agents/git-workflow.agent.md"})

	if len(manifest.Packages) != 1 {
		t.Fatalf("expected 1 package, got %d", len(manifest.Packages))
	}
	if manifest.Packages[0].Name != "git-workflow" {
		t.Errorf("name = %q, want git-workflow", manifest.Packages[0].Name)
	}
	if manifest.InstalledAt == "" {
		t.Error("installed_at should be set")
	}
}

func TestRecordInstallReplacesExisting(t *testing.T) {
	manifest := &model.InstallManifest{
		Version: 1,
		Packages: []model.InstalledPackage{
			{Name: "git-workflow", Version: "0.0.1", Source: "embedded"},
		},
		Personas: []model.InstalledPersona{},
		Teams:    []model.InstalledTeam{},
	}

	RecordInstall(manifest, "git-workflow", "0.2.0", "embedded", "copilot",
		[]string{"agents/git-workflow.agent.md"})

	if len(manifest.Packages) != 1 {
		t.Fatalf("expected 1 package (replaced), got %d", len(manifest.Packages))
	}
	if manifest.Packages[0].Version != "0.2.0" {
		t.Errorf("version = %q, want 0.2.0", manifest.Packages[0].Version)
	}
}

func TestRecordUninstallRemovesPackage(t *testing.T) {
	manifest := &model.InstallManifest{
		Version: 1,
		Packages: []model.InstalledPackage{
			{Name: "git-workflow", Version: "0.1.0"},
			{Name: "developer-mentor", Version: "0.1.0"},
		},
		Personas: []model.InstalledPersona{},
		Teams:    []model.InstalledTeam{},
	}

	removed := RecordUninstall(manifest, "git-workflow")
	if !removed {
		t.Error("should return true when package was found")
	}
	if len(manifest.Packages) != 1 {
		t.Fatalf("expected 1 package remaining, got %d", len(manifest.Packages))
	}
	if manifest.Packages[0].Name != "developer-mentor" {
		t.Errorf("remaining package = %q, want developer-mentor", manifest.Packages[0].Name)
	}
}

func TestRecordUninstallReturnsFalseWhenNotFound(t *testing.T) {
	manifest := &model.InstallManifest{
		Version:  1,
		Packages: []model.InstalledPackage{},
		Personas: []model.InstalledPersona{},
		Teams:    []model.InstalledTeam{},
	}

	removed := RecordUninstall(manifest, "nonexistent")
	if removed {
		t.Error("should return false when package not found")
	}
}

func TestFindInstalledReturnsNilWhenNotFound(t *testing.T) {
	manifest := &model.InstallManifest{
		Version:  1,
		Packages: []model.InstalledPackage{},
	}

	if pkg := FindInstalled(manifest, "nonexistent"); pkg != nil {
		t.Error("should return nil for nonexistent package")
	}
}

func TestFindInstalledReturnsPackage(t *testing.T) {
	manifest := &model.InstallManifest{
		Version: 1,
		Packages: []model.InstalledPackage{
			{Name: "git-workflow", Version: "0.1.0"},
		},
	}

	pkg := FindInstalled(manifest, "git-workflow")
	if pkg == nil {
		t.Fatal("should find installed package")
	}
	if pkg.Version != "0.1.0" {
		t.Errorf("version = %q, want 0.1.0", pkg.Version)
	}
}
