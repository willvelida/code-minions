package model

import (
	"encoding/json"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestPackageYAMLRoundTrip(t *testing.T) {
	input := `name: git-workflow
version: 0.1.0
description: Git workflow skill
author: willvelida
license: MIT
contents:
    agents:
        - agents/git-workflow.agent.md
    skills:
        - skills/git-workflow/SKILL.md
compatibility:
    - copilot
    - claude
dependencies:
    - name: some-dep
      version: '>=0.1.0'
      source: github
`

	var pkg Package
	if err := yaml.Unmarshal([]byte(input), &pkg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if pkg.Name != "git-workflow" {
		t.Errorf("Name: got %q, want %q", pkg.Name, "git-workflow")
	}
	if pkg.Version != "0.1.0" {
		t.Errorf("Version: got %q, want %q", pkg.Version, "0.1.0")
	}
	if pkg.Description != "Git workflow skill" {
		t.Errorf("Description: got %q, want %q", pkg.Description, "Git workflow skill")
	}
	if pkg.Author != "willvelida" {
		t.Errorf("Author: got %q, want %q", pkg.Author, "willvelida")
	}
	if pkg.License != "MIT" {
		t.Errorf("License: got %q, want %q", pkg.License, "MIT")
	}
	if len(pkg.Contents.Agents) != 1 {
		t.Fatalf("Contents.Agents: got %d, want 1", len(pkg.Contents.Agents))
	}
	if pkg.Contents.Agents[0] != "agents/git-workflow.agent.md" {
		t.Errorf("Contents.Agents[0]: got %q", pkg.Contents.Agents[0])
	}
	if len(pkg.Contents.Skills) != 1 {
		t.Fatalf("Contents.Skills: got %d, want 1", len(pkg.Contents.Skills))
	}
	if len(pkg.Compatibility) != 2 {
		t.Fatalf("Compatibility: got %d, want 2", len(pkg.Compatibility))
	}
	if len(pkg.Dependencies) != 1 {
		t.Fatalf("Dependencies: got %d, want 1", len(pkg.Dependencies))
	}
	if pkg.Dependencies[0].Name != "some-dep" {
		t.Errorf("Dependencies[0].Name: got %q", pkg.Dependencies[0].Name)
	}
	if pkg.Dependencies[0].Version != ">=0.1.0" {
		t.Errorf("Dependencies[0].Version: got %q", pkg.Dependencies[0].Version)
	}
	if pkg.Dependencies[0].Source != "github" {
		t.Errorf("Dependencies[0].Source: got %q", pkg.Dependencies[0].Source)
	}

	// Round trip: marshal back
	out, err := yaml.Marshal(&pkg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	if len(out) == 0 {
		t.Fatal("marshal produced empty output")
	}
}

func TestPackageMinimalYAML(t *testing.T) {
	input := `name: minimal
version: 0.1.0
description: A minimal package
`

	var pkg Package
	if err := yaml.Unmarshal([]byte(input), &pkg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if pkg.Name != "minimal" {
		t.Errorf("Name: got %q, want %q", pkg.Name, "minimal")
	}
	if pkg.Author != "" {
		t.Errorf("Author should be empty, got %q", pkg.Author)
	}
	if pkg.License != "" {
		t.Errorf("License should be empty, got %q", pkg.License)
	}
	if len(pkg.Contents.Agents) != 0 {
		t.Errorf("Contents.Agents should be empty, got %d", len(pkg.Contents.Agents))
	}
	if len(pkg.Dependencies) != 0 {
		t.Errorf("Dependencies should be empty, got %d", len(pkg.Dependencies))
	}
}

func TestPersonaYAMLRoundTrip(t *testing.T) {
	input := `name: senior-dev
description: Senior developer persona
author: willvelida
packages:
    - name: git-workflow
      version: '>=0.1.0'
    - name: raise-pull-requests
    - name: threat-modelling
instructions: |
    Prioritise code quality and security.
`

	var persona Persona
	if err := yaml.Unmarshal([]byte(input), &persona); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if persona.Name != "senior-dev" {
		t.Errorf("Name: got %q, want %q", persona.Name, "senior-dev")
	}
	if len(persona.Packages) != 3 {
		t.Fatalf("Packages: got %d, want 3", len(persona.Packages))
	}
	if persona.Packages[0].Name != "git-workflow" {
		t.Errorf("Packages[0].Name: got %q", persona.Packages[0].Name)
	}
	if persona.Packages[0].Version != ">=0.1.0" {
		t.Errorf("Packages[0].Version: got %q", persona.Packages[0].Version)
	}
	if persona.Instructions == "" {
		t.Error("Instructions should not be empty")
	}
}

func TestTeamYAMLRoundTrip(t *testing.T) {
	input := `name: platform-engineering
description: Platform engineering team
author: willvelida
personas:
    - name: senior-dev
      version: '>=1.0.0'
    - name: security-reviewer
config:
    default_assistant: copilot
    enforce_packages: true
`

	var team Team
	if err := yaml.Unmarshal([]byte(input), &team); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if team.Name != "platform-engineering" {
		t.Errorf("Name: got %q, want %q", team.Name, "platform-engineering")
	}
	if len(team.Personas) != 2 {
		t.Fatalf("Personas: got %d, want 2", len(team.Personas))
	}
	if team.Config.DefaultAssistant != "copilot" {
		t.Errorf("Config.DefaultAssistant: got %q", team.Config.DefaultAssistant)
	}
	if !team.Config.EnforcePackages {
		t.Error("Config.EnforcePackages: got false, want true")
	}
}

func TestInstallManifestJSONRoundTrip(t *testing.T) {
	manifest := InstallManifest{
		Version:     1,
		InstalledAt: "2026-02-16T10:30:00Z",
		Assistant:   "copilot",
		Packages: []InstalledPackage{
			{
				Name:        "git-workflow",
				Version:     "0.1.0",
				Source:      "embedded",
				InstalledAt: "2026-02-16T10:30:00Z",
				Persona:     "senior-dev",
				Files:       []string{".github/agents/git-workflow.agent.md"},
			},
		},
		Personas: []InstalledPersona{
			{
				Name:        "senior-dev",
				Version:     "1.0.0",
				Source:      "embedded",
				InstalledAt: "2026-02-16T10:30:00Z",
				Team:        "platform-engineering",
			},
		},
		Teams: []InstalledTeam{
			{
				Name:        "platform-engineering",
				Version:     "1.0.0",
				Source:      "company-packages",
				InstalledAt: "2026-02-16T10:30:00Z",
			},
		},
	}

	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var got InstallManifest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if got.Version != 1 {
		t.Errorf("Version: got %d, want 1", got.Version)
	}
	if got.Assistant != "copilot" {
		t.Errorf("Assistant: got %q", got.Assistant)
	}
	if len(got.Packages) != 1 {
		t.Fatalf("Packages: got %d, want 1", len(got.Packages))
	}
	if got.Packages[0].Persona != "senior-dev" {
		t.Errorf("Packages[0].Persona: got %q", got.Packages[0].Persona)
	}
	if len(got.Personas) != 1 {
		t.Fatalf("Personas: got %d, want 1", len(got.Personas))
	}
	if got.Personas[0].Team != "platform-engineering" {
		t.Errorf("Personas[0].Team: got %q", got.Personas[0].Team)
	}
	if len(got.Teams) != 1 {
		t.Fatalf("Teams: got %d, want 1", len(got.Teams))
	}
}

func TestSearchResultJSON(t *testing.T) {
	result := SearchResult{
		Kind:        "package",
		Name:        "git-workflow",
		Description: "Git workflow",
		Version:     "0.1.0",
		Source:      "embedded",
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var got SearchResult
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if got.Kind != "package" {
		t.Errorf("Kind: got %q", got.Kind)
	}
	if got.Name != "git-workflow" {
		t.Errorf("Name: got %q", got.Name)
	}
	if got.Source != "embedded" {
		t.Errorf("Source: got %q", got.Source)
	}
}
