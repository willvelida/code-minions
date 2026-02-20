package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/willvelida/code-minions/internal/model"
	"github.com/willvelida/code-minions/internal/registry"
)

func TestGetTemplateKnownName(t *testing.T) {
	tmpl, err := getTemplate("standard")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tmpl.Name != "standard" {
		t.Errorf("Name: got %q, want %q", tmpl.Name, "standard")
	}
	if tmpl.Description == "" {
		t.Error("expected non-empty description")
	}
}

func TestGetTemplateUnknownName(t *testing.T) {
	_, err := getTemplate("nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown template, got nil")
	}
	if !strings.Contains(err.Error(), "unknown template") {
		t.Errorf("error should mention 'unknown template', got: %v", err)
	}
}

func TestGetTemplateAllBuiltins(t *testing.T) {
	for _, name := range []string{"minimal", "standard", "security", "fullstack", "docs"} {
		tmpl, err := getTemplate(name)
		if err != nil {
			t.Errorf("getTemplate(%q): unexpected error: %v", name, err)
			continue
		}
		if tmpl.Name != name {
			t.Errorf("getTemplate(%q): Name = %q", name, tmpl.Name)
		}
	}
}

func TestListTemplatesReturnsAll(t *testing.T) {
	templates := listTemplates()
	if len(templates) != 5 {
		t.Errorf("expected 5 templates, got %d", len(templates))
	}
}

func TestTemplateNamesAreUnique(t *testing.T) {
	names := templateNames()
	seen := make(map[string]bool)
	for _, name := range names {
		if seen[name] {
			t.Errorf("duplicate template name: %q", name)
		}
		seen[name] = true
	}
}

func TestTemplateNamesMatchBuiltins(t *testing.T) {
	names := templateNames()
	templates := listTemplates()
	if len(names) != len(templates) {
		t.Fatalf("templateNames() returned %d, listTemplates() returned %d", len(names), len(templates))
	}
	for i, name := range names {
		if name != templates[i].Name {
			t.Errorf("index %d: templateNames()=%q, listTemplates()=%q", i, name, templates[i].Name)
		}
	}
}

func TestResolveTemplatePackagesStaticTemplate(t *testing.T) {
	tmpl := &Template{
		Name:     "test",
		Packages: []string{"git-workflow", "threat-modelling"},
	}
	available := []model.Package{
		{Name: "git-workflow"},
		{Name: "threat-modelling"},
		{Name: "developer-mentor"},
	}

	result := resolveTemplatePackages(tmpl, available)
	if len(result) != 2 {
		t.Fatalf("expected 2 packages, got %d: %v", len(result), result)
	}
	if result[0] != "git-workflow" || result[1] != "threat-modelling" {
		t.Errorf("unexpected packages: %v", result)
	}
}

func TestResolveTemplatePackagesFullstack(t *testing.T) {
	tmpl := &Template{
		Name:     "fullstack",
		Packages: nil, // nil means all available
	}
	available := []model.Package{
		{Name: "git-workflow"},
		{Name: "threat-modelling"},
		{Name: "developer-mentor"},
	}

	result := resolveTemplatePackages(tmpl, available)
	if len(result) != 3 {
		t.Fatalf("expected 3 packages (all), got %d: %v", len(result), result)
	}
	for i, pkg := range available {
		if result[i] != pkg.Name {
			t.Errorf("result[%d]: got %q, want %q", i, result[i], pkg.Name)
		}
	}
}

func TestTemplatePackagesExistInEmbeddedRegistry(t *testing.T) {
	// Build a content FS that mirrors the real embedded packages
	content := testContentFSWithAllPackages()
	src := registry.NewEmbeddedSource(content)
	pkgModels, err := src.ListPackages()
	if err != nil {
		t.Fatalf("failed to list packages: %v", err)
	}

	validPkgs := make(map[string]bool)
	for _, pkg := range pkgModels {
		validPkgs[pkg.Name] = true
	}

	for _, tmpl := range listTemplates() {
		if tmpl.Packages == nil {
			continue // fullstack uses all packages, always valid
		}
		for _, pkgName := range tmpl.Packages {
			if !validPkgs[pkgName] {
				t.Errorf("template %q references unknown package %q", tmpl.Name, pkgName)
			}
		}
	}
}

func TestSelectTemplateDefaultChoice(t *testing.T) {
	templates := listTemplates()
	stdin := strings.NewReader("\n")
	var stdout bytes.Buffer

	result, err := selectTemplate(stdin, &stdout, templates)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected default template, got nil (blank)")
	}
	if result.Name != defaultTemplateName {
		t.Errorf("expected default %q, got %q", defaultTemplateName, result.Name)
	}
}

func TestSelectTemplateSpecificChoice(t *testing.T) {
	templates := listTemplates()
	// Choose first template
	stdin := strings.NewReader("1\n")
	var stdout bytes.Buffer

	result, err := selectTemplate(stdin, &stdout, templates)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected template, got nil")
	}
	if result.Name != templates[0].Name {
		t.Errorf("expected %q, got %q", templates[0].Name, result.Name)
	}
}

func TestSelectTemplateBlankChoice(t *testing.T) {
	templates := listTemplates()
	// Blank is always len(templates)+1
	blankIdx := len(templates) + 1
	stdin := strings.NewReader(fmt.Sprintf("%d\n", blankIdx))
	var stdout bytes.Buffer

	result, err := selectTemplate(stdin, &stdout, templates)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil (blank), got template %q", result.Name)
	}
}

func TestSelectTemplateInvalidInput(t *testing.T) {
	templates := listTemplates()
	stdin := strings.NewReader("abc\n")
	var stdout bytes.Buffer

	_, err := selectTemplate(stdin, &stdout, templates)
	if err == nil {
		t.Fatal("expected error for invalid input, got nil")
	}
	if !strings.Contains(err.Error(), "invalid selection") {
		t.Errorf("expected 'invalid selection' error, got: %v", err)
	}
}

func TestSelectTemplateOutOfRange(t *testing.T) {
	templates := listTemplates()
	stdin := strings.NewReader("99\n")
	var stdout bytes.Buffer

	_, err := selectTemplate(stdin, &stdout, templates)
	if err == nil {
		t.Fatal("expected error for out-of-range input, got nil")
	}
	if !strings.Contains(err.Error(), "must be between") {
		t.Errorf("expected range error, got: %v", err)
	}
}

func TestSelectTemplateZeroIndex(t *testing.T) {
	templates := listTemplates()
	stdin := strings.NewReader("0\n")
	var stdout bytes.Buffer

	_, err := selectTemplate(stdin, &stdout, templates)
	if err == nil {
		t.Fatal("expected error for index 0, got nil")
	}
}

func TestSelectTemplateEmptyList(t *testing.T) {
	stdin := strings.NewReader("\n")
	var stdout bytes.Buffer

	result, err := selectTemplate(stdin, &stdout, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil for empty template list, got %v", result)
	}
}

func TestSelectTemplateOutputFormat(t *testing.T) {
	templates := listTemplates()
	stdin := strings.NewReader("\n")
	var stdout bytes.Buffer

	_, err := selectTemplate(stdin, &stdout, templates)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Choose a starting template:") {
		t.Error("output should contain 'Choose a starting template:' header")
	}
	if !strings.Contains(output, "[1]") {
		t.Error("output should contain numbered list")
	}
	if !strings.Contains(output, "blank") {
		t.Error("output should contain 'blank' option")
	}
}

// testContentFSWithAllPackages builds a test FS with all packages referenced by templates.
func testContentFSWithAllPackages() fstest.MapFS {
	return fstest.MapFS{
		"packages/git-workflow/package.yaml": &fstest.MapFile{
			Data: []byte("name: git-workflow\nversion: 0.1.0\ndescription: Git workflow conventions\n"),
		},
		"packages/git-workflow/agents/git-workflow.agent.md": &fstest.MapFile{
			Data: []byte("# Git Agent"),
		},
		"packages/developer-mentor/package.yaml": &fstest.MapFile{
			Data: []byte("name: developer-mentor\nversion: 0.1.0\ndescription: Developer mentoring skill\n"),
		},
		"packages/developer-mentor/agents/developer-mentor.agent.md": &fstest.MapFile{
			Data: []byte("# Mentor Agent"),
		},
		"packages/raise-pull-requests/package.yaml": &fstest.MapFile{
			Data: []byte("name: raise-pull-requests\nversion: 0.1.0\ndescription: Raise high-quality PRs\n"),
		},
		"packages/raise-pull-requests/agents/raise-pull-requests.agent.md": &fstest.MapFile{
			Data: []byte("# PR Agent"),
		},
		"packages/threat-modelling/package.yaml": &fstest.MapFile{
			Data: []byte("name: threat-modelling\nversion: 0.1.0\ndescription: STRIDE-based threat modelling\n"),
		},
		"packages/threat-modelling/agents/threat-modelling.agent.md": &fstest.MapFile{
			Data: []byte("# Threat Agent"),
		},
		"packages/creating-documentation/package.yaml": &fstest.MapFile{
			Data: []byte("name: creating-documentation\nversion: 0.1.0\ndescription: Create documentation\n"),
		},
		"packages/creating-documentation/agents/creating-documentation.agent.md": &fstest.MapFile{
			Data: []byte("# Docs Agent"),
		},
	}
}
