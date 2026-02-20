package model

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/willvelida/code-minions/internal/mcp"
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
	// New fields should be zero-valued when absent (backward compat)
	if team.MCP != nil {
		t.Error("MCP should be nil when absent")
	}
	if team.Instructions != "" {
		t.Error("Instructions should be empty when absent")
	}
	// Personas without inline packages should have empty Packages slice
	if len(team.Personas[0].Packages) != 0 {
		t.Errorf("Personas[0].Packages should be empty, got %d", len(team.Personas[0].Packages))
	}
}

func TestTeamYAMLWithInlinePackages(t *testing.T) {
	input := `name: platform-engineering
description: Standard setup for the platform team
personas:
    - name: senior-dev
      packages:
        - name: git-workflow
        - name: developer-mentor
        - name: raise-pull-requests
    - name: security-reviewer
      packages:
        - name: git-workflow
        - name: threat-modelling
config:
    default_assistant: copilot
    enforce_packages: true
`

	var team Team
	if err := yaml.Unmarshal([]byte(input), &team); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if team.Name != "platform-engineering" {
		t.Errorf("Name: got %q", team.Name)
	}
	if len(team.Personas) != 2 {
		t.Fatalf("Personas: got %d, want 2", len(team.Personas))
	}

	// First persona has 3 inline packages
	p0 := team.Personas[0]
	if p0.Name != "senior-dev" {
		t.Errorf("Personas[0].Name: got %q", p0.Name)
	}
	if len(p0.Packages) != 3 {
		t.Fatalf("Personas[0].Packages: got %d, want 3", len(p0.Packages))
	}
	if p0.Packages[0].Name != "git-workflow" {
		t.Errorf("Personas[0].Packages[0].Name: got %q", p0.Packages[0].Name)
	}
	if p0.Packages[1].Name != "developer-mentor" {
		t.Errorf("Personas[0].Packages[1].Name: got %q", p0.Packages[1].Name)
	}
	if p0.Packages[2].Name != "raise-pull-requests" {
		t.Errorf("Personas[0].Packages[2].Name: got %q", p0.Packages[2].Name)
	}

	// Second persona has 2 inline packages
	p1 := team.Personas[1]
	if p1.Name != "security-reviewer" {
		t.Errorf("Personas[1].Name: got %q", p1.Name)
	}
	if len(p1.Packages) != 2 {
		t.Fatalf("Personas[1].Packages: got %d, want 2", len(p1.Packages))
	}

	// Round trip: marshal and unmarshal should preserve packages
	out, err := yaml.Marshal(&team)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	var roundTripped Team
	if err := yaml.Unmarshal(out, &roundTripped); err != nil {
		t.Fatalf("failed to unmarshal round-tripped YAML: %v", err)
	}
	if len(roundTripped.Personas[0].Packages) != 3 {
		t.Errorf("round-trip Personas[0].Packages: got %d, want 3", len(roundTripped.Personas[0].Packages))
	}
	if len(roundTripped.Personas[1].Packages) != 2 {
		t.Errorf("round-trip Personas[1].Packages: got %d, want 2", len(roundTripped.Personas[1].Packages))
	}
}

func TestTeamYAMLWithMCPAndInstructions(t *testing.T) {
	input := `name: platform-engineering
description: Platform engineering team
personas:
    - name: frontend-dev
    - name: backend-dev
config:
    default_assistant: copilot
mcp:
    servers:
        github:
            description: GitHub API via MCP
            transport: stdio
            command: npx
            args:
                - -y
                - '@modelcontextprotocol/server-github'
            env:
                GITHUB_TOKEN: ""
        postgres:
            transport: stdio
            command: npx
            args:
                - -y
                - '@modelcontextprotocol/server-postgres'
            env:
                DATABASE_URL: ""
instructions: |
    All agents must follow coding standards.
    Never commit secrets.
`

	var team Team
	if err := yaml.Unmarshal([]byte(input), &team); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if team.Name != "platform-engineering" {
		t.Errorf("Name: got %q", team.Name)
	}
	if team.MCP == nil {
		t.Fatal("MCP should not be nil")
	}
	if len(team.MCP.Servers) != 2 {
		t.Fatalf("MCP.Servers: got %d, want 2", len(team.MCP.Servers))
	}
	gh, ok := team.MCP.Servers["github"]
	if !ok {
		t.Fatal("expected github server")
	}
	if gh.Command != "npx" {
		t.Errorf("github.Command: got %q", gh.Command)
	}
	if gh.Description != "GitHub API via MCP" {
		t.Errorf("github.Description: got %q", gh.Description)
	}
	if gh.Env["GITHUB_TOKEN"] != "" {
		t.Errorf("github.Env[GITHUB_TOKEN]: got %q, want empty", gh.Env["GITHUB_TOKEN"])
	}
	pg, ok := team.MCP.Servers["postgres"]
	if !ok {
		t.Fatal("expected postgres server")
	}
	if pg.Env["DATABASE_URL"] != "" {
		t.Errorf("postgres.Env[DATABASE_URL]: got %q, want empty", pg.Env["DATABASE_URL"])
	}
	if team.Instructions == "" {
		t.Fatal("Instructions should not be empty")
	}
	if team.Instructions != "All agents must follow coding standards.\nNever commit secrets.\n" {
		t.Errorf("Instructions: got %q", team.Instructions)
	}

	// Round trip
	out, err := yaml.Marshal(&team)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	if len(out) == 0 {
		t.Fatal("marshal produced empty output")
	}
}

func TestValidateTeam(t *testing.T) {
	tests := []struct {
		name    string
		team    Team
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid minimal team",
			team: Team{
				Name:        "my-team",
				Description: "A team",
				Personas:    []PersonaRef{{Name: "dev"}},
			},
		},
		{
			name: "missing name",
			team: Team{
				Description: "A team",
				Personas:    []PersonaRef{{Name: "dev"}},
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "missing description",
			team: Team{
				Name:     "my-team",
				Personas: []PersonaRef{{Name: "dev"}},
			},
			wantErr: true,
			errMsg:  "description is required",
		},
		{
			name: "no personas",
			team: Team{
				Name:        "my-team",
				Description: "A team",
				Personas:    []PersonaRef{},
			},
			wantErr: true,
			errMsg:  "at least one persona",
		},
		{
			name: "instructions too long",
			team: Team{
				Name:         "my-team",
				Description:  "A team",
				Personas:     []PersonaRef{{Name: "dev"}},
				Instructions: string(make([]byte, MaxInstructionLength+1)),
			},
			wantErr: true,
			errMsg:  "exceed maximum length",
		},
		{
			name: "instructions at max length",
			team: Team{
				Name:         "my-team",
				Description:  "A team",
				Personas:     []PersonaRef{{Name: "dev"}},
				Instructions: string(make([]byte, MaxInstructionLength)),
			},
		},
		{
			name: "invalid MCP config",
			team: Team{
				Name:        "my-team",
				Description: "A team",
				Personas:    []PersonaRef{{Name: "dev"}},
				MCP: &mcp.Config{
					Servers: map[string]mcp.Server{
						"bad-server": {Transport: "stdio"},
					},
				},
			},
			wantErr: true,
			errMsg:  "requires a command",
		},
		{
			name: "invalid team name",
			team: Team{
				Name:        "my-->team",
				Description: "A team",
				Personas:    []PersonaRef{{Name: "dev"}},
			},
			wantErr: true,
			errMsg:  "invalid characters",
		},
		{
			name: "valid team with inline packages",
			team: Team{
				Name:        "my-team",
				Description: "A team",
				Personas: []PersonaRef{
					{
						Name: "dev",
						Packages: []PackageRef{
							{Name: "git-workflow"},
							{Name: "developer-mentor"},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTeam(&tt.team)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestInstallManifestJSONRoundTrip(t *testing.T) {
	manifest := InstallManifest{
		Version:     1,
		InstalledAt: "2026-02-16T10:30:00Z",
		Assistant:   "copilot",
		Packages: []InstalledPackage{
			{
				Name:         "git-workflow",
				Version:      "0.1.0",
				Source:       "embedded",
				InstalledAt:  "2026-02-16T10:30:00Z",
				Persona:      "senior-dev",
				Files:        []string{".github/agents/git-workflow.agent.md"},
				ReferencedBy: []string{"platform-engineering"},
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
				Name:                 "platform-engineering",
				Version:              "1.0.0",
				Source:               "company-packages",
				InstalledAt:          "2026-02-16T10:30:00Z",
				Personas:             []string{"senior-dev", "security-reviewer"},
				MCPServers:           []string{"github", "postgres"},
				InstructionsInjected: true,
				InstructionsFile:     ".github/copilot-instructions.md",
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
	if len(got.Packages[0].ReferencedBy) != 1 || got.Packages[0].ReferencedBy[0] != "platform-engineering" {
		t.Errorf("Packages[0].ReferencedBy: got %v", got.Packages[0].ReferencedBy)
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
	team := got.Teams[0]
	if len(team.Personas) != 2 {
		t.Errorf("Teams[0].Personas: got %v, want [senior-dev security-reviewer]", team.Personas)
	}
	if len(team.MCPServers) != 2 {
		t.Errorf("Teams[0].MCPServers: got %v, want [github postgres]", team.MCPServers)
	}
	if !team.InstructionsInjected {
		t.Error("Teams[0].InstructionsInjected: got false, want true")
	}
	if team.InstructionsFile != ".github/copilot-instructions.md" {
		t.Errorf("Teams[0].InstructionsFile: got %q", team.InstructionsFile)
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
