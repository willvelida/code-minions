package installer

import (
	"strings"
	"testing"

	"github.com/willvelida/code-minions/internal/assistant"
	"github.com/willvelida/code-minions/internal/model"
	"github.com/willvelida/code-minions/internal/registry"
)

func TestBuildClaudeMDForPersona(t *testing.T) {
	cfg, err := assistant.Get("claude")
	if err != nil {
		t.Fatalf("failed to get claude config: %v", err)
	}

	resolved := &registry.ResolvedPersona{
		Persona: model.Persona{
			Name:        "senior-dev",
			Description: "A senior developer persona",
			Packages: []model.PackageRef{
				{Name: "git-workflow"},
				{Name: "threat-modelling"},
			},
		},
		Packages: []registry.ResolvedPackage{
			{Package: model.Package{Name: "git-workflow", Description: "Git workflow skill"}},
			{Package: model.Package{Name: "threat-modelling", Description: "Threat modelling skill"}},
		},
	}

	content := BuildClaudeMDForPersona(resolved, cfg)

	// Verify header
	if !strings.Contains(content, "# Project Instructions") {
		t.Error("missing header")
	}
	if !strings.Contains(content, "code-minions") {
		t.Error("missing code-minions reference")
	}

	// Verify persona section
	if !strings.Contains(content, "## Installed Persona: Senior Dev") {
		t.Error("missing persona heading")
	}
	if !strings.Contains(content, "A senior developer persona") {
		t.Error("missing persona description")
	}

	// Verify @import references for skills
	if !strings.Contains(content, "@.claude/skills/git-workflow/SKILL.md") {
		t.Error("missing @import for git-workflow skill")
	}
	if !strings.Contains(content, "@.claude/skills/threat-modelling/SKILL.md") {
		t.Error("missing @import for threat-modelling skill")
	}

	// Verify agent reference
	if !strings.Contains(content, "@.claude/agents/senior-dev.agent.md") {
		t.Error("missing @import for persona agent")
	}
}

func TestBuildClaudeMDForPersonaNoDescription(t *testing.T) {
	cfg, err := assistant.Get("claude")
	if err != nil {
		t.Fatalf("failed to get claude config: %v", err)
	}

	resolved := &registry.ResolvedPersona{
		Persona: model.Persona{
			Name: "basic",
			Packages: []model.PackageRef{
				{Name: "git-workflow"},
			},
		},
		Packages: []registry.ResolvedPackage{
			{Package: model.Package{Name: "git-workflow", Description: "Git workflow skill"}},
		},
	}

	content := BuildClaudeMDForPersona(resolved, cfg)

	// Should not have empty lines from missing description
	if strings.Contains(content, "## Installed Persona: Basic\n\n\n") {
		t.Error("should not have triple newlines when description is empty")
	}
}

func TestBuildClaudeMDForPackages(t *testing.T) {
	cfg, err := assistant.Get("claude")
	if err != nil {
		t.Fatalf("failed to get claude config: %v", err)
	}

	packages := []string{"git-workflow", "threat-modelling", "developer-mentor"}

	content := BuildClaudeMDForPackages(packages, cfg)

	// Verify header
	if !strings.Contains(content, "# Project Instructions") {
		t.Error("missing header")
	}
	if !strings.Contains(content, "code-minions") {
		t.Error("missing code-minions reference")
	}

	// Verify packages section
	if !strings.Contains(content, "## Installed Packages") {
		t.Error("missing packages heading")
	}

	// Verify @import references
	if !strings.Contains(content, "@.claude/skills/git-workflow/SKILL.md") {
		t.Error("missing @import for git-workflow")
	}
	if !strings.Contains(content, "@.claude/skills/threat-modelling/SKILL.md") {
		t.Error("missing @import for threat-modelling")
	}
	if !strings.Contains(content, "@.claude/skills/developer-mentor/SKILL.md") {
		t.Error("missing @import for developer-mentor")
	}

	// Verify display names are title-cased
	if !strings.Contains(content, "**Git Workflow**") {
		t.Error("missing title-cased display name for git-workflow")
	}
	if !strings.Contains(content, "**Developer Mentor**") {
		t.Error("missing title-cased display name for developer-mentor")
	}
}

func TestBuildClaudeMDForPackagesSinglePackage(t *testing.T) {
	cfg, err := assistant.Get("claude")
	if err != nil {
		t.Fatalf("failed to get claude config: %v", err)
	}

	packages := []string{"developer-mentor"}

	content := BuildClaudeMDForPackages(packages, cfg)

	if !strings.Contains(content, "**Developer Mentor**") {
		t.Error("missing package entry")
	}
	if !strings.Contains(content, "@.claude/skills/developer-mentor/SKILL.md") {
		t.Error("missing @import reference")
	}
}
