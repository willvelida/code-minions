package installer

import (
	"testing"

	"github.com/willvelida/code-minions/internal/model"
)

func newTestManifest() *model.InstallManifest {
	return &model.InstallManifest{
		Version:  1,
		Packages: []model.InstalledPackage{},
		Personas: []model.InstalledPersona{},
		Teams:    []model.InstalledTeam{},
	}
}

func TestRecordTeamInstall_Basic(t *testing.T) {
	m := newTestManifest()

	// Pre-populate personas and packages
	RecordPersonaInstall(m, "senior-dev", "1.0.0", "embedded", "copilot",
		[]string{"git-workflow", "raise-prs"}, nil)
	RecordInstall(m, "git-workflow", "0.1.0", "embedded", "copilot", []string{"agents/gw.md"})
	RecordInstall(m, "raise-prs", "0.1.0", "embedded", "copilot", []string{"agents/rp.md"})

	RecordTeamInstall(m, "platform-eng", "1.0.0", "embedded", "copilot",
		[]string{"senior-dev"}, []string{"github", "postgres"}, true, ".github/copilot-instructions.md")

	if len(m.Teams) != 1 {
		t.Fatalf("expected 1 team, got %d", len(m.Teams))
	}

	team := m.Teams[0]
	if team.Name != "platform-eng" {
		t.Errorf("name = %q", team.Name)
	}
	if len(team.Personas) != 1 || team.Personas[0] != "senior-dev" {
		t.Errorf("personas = %v", team.Personas)
	}
	if len(team.MCPServers) != 2 {
		t.Errorf("mcp_servers = %v", team.MCPServers)
	}
	if !team.InstructionsInjected {
		t.Error("instructions_injected should be true")
	}
	if team.InstructionsFile != ".github/copilot-instructions.md" {
		t.Errorf("instructions_file = %q", team.InstructionsFile)
	}

	// Persona should be stamped with the team name
	p := FindInstalledPersona(m, "senior-dev")
	if p == nil {
		t.Fatal("persona not found")
	}
	if p.Team != "platform-eng" {
		t.Errorf("persona.Team = %q, want platform-eng", p.Team)
	}

	// Packages should have ReferencedBy
	pkg := FindInstalled(m, "git-workflow")
	if pkg == nil {
		t.Fatal("package not found")
	}
	if len(pkg.ReferencedBy) != 1 || pkg.ReferencedBy[0] != "platform-eng" {
		t.Errorf("git-workflow.ReferencedBy = %v", pkg.ReferencedBy)
	}
}

func TestRecordTeamInstall_ReinstallReplaces(t *testing.T) {
	m := newTestManifest()

	RecordTeamInstall(m, "my-team", "1.0.0", "embedded", "copilot",
		nil, []string{"github"}, false, "")
	RecordTeamInstall(m, "my-team", "2.0.0", "embedded", "copilot",
		nil, []string{"github", "postgres"}, true, "CLAUDE.md")

	if len(m.Teams) != 1 {
		t.Fatalf("expected 1 team after reinstall, got %d", len(m.Teams))
	}
	if m.Teams[0].Version != "2.0.0" {
		t.Errorf("version = %q, want 2.0.0", m.Teams[0].Version)
	}
	if len(m.Teams[0].MCPServers) != 2 {
		t.Errorf("mcp_servers = %v", m.Teams[0].MCPServers)
	}
}

func TestRecordTeamUninstall_Basic(t *testing.T) {
	m := newTestManifest()

	RecordPersonaInstall(m, "dev", "1.0.0", "embedded", "copilot", []string{"pkg-a"}, nil)
	RecordInstall(m, "pkg-a", "0.1.0", "embedded", "copilot", []string{"a.md"})
	RecordTeamInstall(m, "team-x", "1.0.0", "embedded", "copilot",
		[]string{"dev"}, []string{"github"}, true, "CLAUDE.md")

	// Verify stamps are set
	if p := FindInstalledPersona(m, "dev"); p.Team != "team-x" {
		t.Fatal("persona should be stamped before uninstall")
	}

	removed := RecordTeamUninstall(m, "team-x")
	if !removed {
		t.Fatal("expected removed=true")
	}
	if len(m.Teams) != 0 {
		t.Errorf("teams should be empty, got %d", len(m.Teams))
	}

	// Persona team stamp should be cleared
	if p := FindInstalledPersona(m, "dev"); p.Team != "" {
		t.Errorf("persona.Team should be cleared, got %q", p.Team)
	}

	// Package ReferencedBy should be cleared
	if pkg := FindInstalled(m, "pkg-a"); len(pkg.ReferencedBy) != 0 {
		t.Errorf("pkg.ReferencedBy should be empty, got %v", pkg.ReferencedBy)
	}
}

func TestRecordTeamUninstall_NotFound(t *testing.T) {
	m := newTestManifest()
	if RecordTeamUninstall(m, "nope") {
		t.Error("should return false for absent team")
	}
}

func TestFindInstalledTeam(t *testing.T) {
	m := newTestManifest()
	RecordTeamInstall(m, "team-a", "1.0.0", "embedded", "copilot", nil, nil, false, "")

	if team := FindInstalledTeam(m, "team-a"); team == nil {
		t.Error("expected to find team-a")
	}
	if team := FindInstalledTeam(m, "team-b"); team != nil {
		t.Error("expected nil for absent team")
	}
}

func TestTeamExclusivePersonas(t *testing.T) {
	m := newTestManifest()

	// Two teams share "shared-persona", "team-a" uniquely owns "a-only"
	RecordTeamInstall(m, "team-a", "1.0.0", "embedded", "copilot",
		[]string{"shared-persona", "a-only"}, nil, false, "")
	RecordTeamInstall(m, "team-b", "1.0.0", "embedded", "copilot",
		[]string{"shared-persona", "b-only"}, nil, false, "")

	exclusive := TeamExclusivePersonas(m, "team-a")
	if len(exclusive) != 1 || exclusive[0] != "a-only" {
		t.Errorf("expected [a-only], got %v", exclusive)
	}

	exclusiveB := TeamExclusivePersonas(m, "team-b")
	if len(exclusiveB) != 1 || exclusiveB[0] != "b-only" {
		t.Errorf("expected [b-only], got %v", exclusiveB)
	}
}

func TestTeamExclusivePersonas_NotFound(t *testing.T) {
	m := newTestManifest()
	if result := TeamExclusivePersonas(m, "nope"); result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestTeamExclusiveMCPServers(t *testing.T) {
	m := newTestManifest()

	// team-a has github + postgres; team-b also has github
	RecordTeamInstall(m, "team-a", "1.0.0", "embedded", "copilot",
		nil, []string{"github", "postgres"}, false, "")
	RecordTeamInstall(m, "team-b", "1.0.0", "embedded", "copilot",
		nil, []string{"github", "sentry"}, false, "")

	exclusive := TeamExclusiveMCPServers(m, "team-a")
	if len(exclusive) != 1 || exclusive[0] != "postgres" {
		t.Errorf("expected [postgres], got %v", exclusive)
	}
}

func TestTeamExclusiveMCPServers_AlsoChecksPackages(t *testing.T) {
	m := newTestManifest()

	// Package outside the team also claims "github"
	RecordInstall(m, "standalone-pkg", "0.1.0", "embedded", "copilot", []string{"s.md"})
	pkg := FindInstalled(m, "standalone-pkg")
	pkg.MCPServers = []string{"github"}

	RecordTeamInstall(m, "team-a", "1.0.0", "embedded", "copilot",
		nil, []string{"github", "postgres"}, false, "")

	exclusive := TeamExclusiveMCPServers(m, "team-a")
	if len(exclusive) != 1 || exclusive[0] != "postgres" {
		t.Errorf("expected [postgres] (github shared with standalone-pkg), got %v", exclusive)
	}
}

func TestTeamExclusiveMCPServers_NotFound(t *testing.T) {
	m := newTestManifest()
	if result := TeamExclusiveMCPServers(m, "nope"); result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestStringSliceContains(t *testing.T) {
	if !stringSliceContains([]string{"a", "b", "c"}, "b") {
		t.Error("should find 'b'")
	}
	if stringSliceContains([]string{"a", "b"}, "z") {
		t.Error("should not find 'z'")
	}
	if stringSliceContains(nil, "a") {
		t.Error("nil slice should not contain anything")
	}
}

func TestRemoveFromSlice(t *testing.T) {
	result := removeFromSlice([]string{"a", "b", "c"}, "b")
	if len(result) != 2 || result[0] != "a" || result[1] != "c" {
		t.Errorf("got %v", result)
	}
	// Value not found — returns unchanged
	result = removeFromSlice([]string{"a", "b"}, "z")
	if len(result) != 2 {
		t.Errorf("should be unchanged, got %v", result)
	}
}
