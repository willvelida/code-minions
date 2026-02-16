package registry

import (
	"testing"
	"testing/fstest"
)

func testEmbeddedFS() fstest.MapFS {
	return fstest.MapFS{
		"packages/git-workflow/package.yaml": &fstest.MapFile{
			Data: []byte(`name: git-workflow
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
`),
		},
		"packages/git-workflow/agents/git-workflow.agent.md": &fstest.MapFile{
			Data: []byte("# Git Agent"),
		},
		"packages/git-workflow/skills/git-workflow/SKILL.md": &fstest.MapFile{
			Data: []byte("---\nname: git-workflow\ndescription: 'Git workflow skill'\nlicense: MIT\n---\n# Git"),
		},
		"packages/developer-mentor/agents/developer-mentor.agent.md": &fstest.MapFile{
			Data: []byte("# Mentor Agent"),
		},
		"packages/developer-mentor/skills/developer-mentor/SKILL.md": &fstest.MapFile{
			Data: []byte("---\nname: developer-mentor\ndescription: 'Socratic development mentoring'\nlicense: MIT\n---\n# Mentor"),
		},
	}
}

// ---------- EmbeddedSource tests ----------

func TestEmbeddedSourceNameAndType(t *testing.T) {
	src := NewEmbeddedSource(testEmbeddedFS())

	if src.Name() != "embedded" {
		t.Errorf("Name: got %q, want %q", src.Name(), "embedded")
	}
	if src.Type() != "embedded" {
		t.Errorf("Type: got %q, want %q", src.Type(), "embedded")
	}
}

func TestEmbeddedSourceListPackages(t *testing.T) {
	src := NewEmbeddedSource(testEmbeddedFS())

	pkgs, err := src.ListPackages()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}

	// Find the package with a package.yaml
	var gitPkg, mentorPkg *struct {
		name    string
		version string
		desc    string
	}

	for _, p := range pkgs {
		switch p.Name {
		case "git-workflow":
			gitPkg = &struct {
				name    string
				version string
				desc    string
			}{p.Name, p.Version, p.Description}
		case "developer-mentor":
			mentorPkg = &struct {
				name    string
				version string
				desc    string
			}{p.Name, p.Version, p.Description}
		}
	}

	if gitPkg == nil {
		t.Fatal("git-workflow package not found")
	}
	if gitPkg.version != "0.1.0" {
		t.Errorf("git-workflow version: got %q, want %q", gitPkg.version, "0.1.0")
	}
	if gitPkg.desc != "Git workflow skill" {
		t.Errorf("git-workflow description: got %q", gitPkg.desc)
	}

	// developer-mentor has no package.yaml, so it falls back
	if mentorPkg == nil {
		t.Fatal("developer-mentor package not found")
	}
	if mentorPkg.desc != "Socratic development mentoring" {
		t.Errorf("developer-mentor description (fallback): got %q", mentorPkg.desc)
	}
}

func TestEmbeddedSourceGetPackage(t *testing.T) {
	src := NewEmbeddedSource(testEmbeddedFS())

	pkg, err := src.GetPackage("git-workflow")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pkg.Name != "git-workflow" {
		t.Errorf("Name: got %q", pkg.Name)
	}
	if pkg.Version != "0.1.0" {
		t.Errorf("Version: got %q", pkg.Version)
	}
	if pkg.License != "MIT" {
		t.Errorf("License: got %q", pkg.License)
	}
}

func TestEmbeddedSourceGetPackageFallback(t *testing.T) {
	src := NewEmbeddedSource(testEmbeddedFS())

	pkg, err := src.GetPackage("developer-mentor")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pkg.Name != "developer-mentor" {
		t.Errorf("Name: got %q", pkg.Name)
	}
	// Fallback reads description from SKILL.md frontmatter
	if pkg.Description != "Socratic development mentoring" {
		t.Errorf("Description: got %q", pkg.Description)
	}
	if pkg.License != "MIT" {
		t.Errorf("License: got %q", pkg.License)
	}
}

func TestEmbeddedSourceGetPackageNotFound(t *testing.T) {
	src := NewEmbeddedSource(testEmbeddedFS())

	_, err := src.GetPackage("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent package")
	}
}

func TestEmbeddedSourceDownloadPackage(t *testing.T) {
	src := NewEmbeddedSource(testEmbeddedFS())

	sub, err := src.DownloadPackage("git-workflow", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be able to read files relative to the package root
	data, err := fstest.MapFS(nil).Open(".")
	_ = data
	// Just verify the sub-filesystem is not nil
	if sub == nil {
		t.Fatal("expected non-nil sub-filesystem")
	}
}

func TestEmbeddedSourceDownloadPackageNotFound(t *testing.T) {
	src := NewEmbeddedSource(testEmbeddedFS())

	_, err := src.DownloadPackage("nonexistent", "")
	if err == nil {
		t.Fatal("expected error for nonexistent package")
	}
}

func TestEmbeddedSourceSearch(t *testing.T) {
	src := NewEmbeddedSource(testEmbeddedFS())

	results, err := src.Search("git")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].Name != "git-workflow" {
		t.Errorf("Name: got %q", results[0].Name)
	}
	if results[0].Kind != "package" {
		t.Errorf("Kind: got %q", results[0].Kind)
	}
	if results[0].Source != "embedded" {
		t.Errorf("Source: got %q", results[0].Source)
	}
}

func TestEmbeddedSourceSearchCaseInsensitive(t *testing.T) {
	src := NewEmbeddedSource(testEmbeddedFS())

	results, err := src.Search("MENTOR")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].Name != "developer-mentor" {
		t.Errorf("Name: got %q", results[0].Name)
	}
}

func TestEmbeddedSourceSearchNoMatch(t *testing.T) {
	src := NewEmbeddedSource(testEmbeddedFS())

	results, err := src.Search("nonexistent-query")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestEmbeddedSourcePersonasAndTeamsEmpty(t *testing.T) {
	src := NewEmbeddedSource(testEmbeddedFS())

	personas, err := src.ListPersonas()
	if err != nil {
		t.Fatalf("ListPersonas: unexpected error: %v", err)
	}
	if len(personas) != 0 {
		t.Errorf("ListPersonas: expected 0, got %d", len(personas))
	}

	_, err = src.GetPersona("any")
	if err == nil {
		t.Error("GetPersona: expected error")
	}

	teams, err := src.ListTeams()
	if err != nil {
		t.Fatalf("ListTeams: unexpected error: %v", err)
	}
	if len(teams) != 0 {
		t.Errorf("ListTeams: expected 0, got %d", len(teams))
	}

	_, err = src.GetTeam("any")
	if err == nil {
		t.Error("GetTeam: expected error")
	}
}

// ---------- extractFrontmatterField tests ----------

func TestExtractFrontmatterField(t *testing.T) {
	tests := []struct {
		name    string
		content string
		field   string
		want    string
	}{
		{
			name:    "extracts description",
			content: "---\nname: test\ndescription: 'A test skill'\nlicense: MIT\n---\n# Test",
			field:   "description",
			want:    "A test skill",
		},
		{
			name:    "extracts license",
			content: "---\nname: test\nlicense: MIT\n---\n# Test",
			field:   "license",
			want:    "MIT",
		},
		{
			name:    "returns empty when field missing",
			content: "---\nname: test\n---\n# Test",
			field:   "description",
			want:    "",
		},
		{
			name:    "returns empty when no frontmatter",
			content: "# Just a heading",
			field:   "description",
			want:    "",
		},
		{
			name:    "handles double-quoted values",
			content: "---\ndescription: \"A quoted skill\"\n---",
			field:   "description",
			want:    "A quoted skill",
		},
		{
			name:    "handles unquoted values",
			content: "---\nlicense: Apache-2.0\n---",
			field:   "license",
			want:    "Apache-2.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractFrontmatterField(tt.content, tt.field)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
