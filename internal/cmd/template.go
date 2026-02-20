package cmd

import (
	"fmt"
	"strings"

	"github.com/willvelida/code-minions/internal/model"
)

// Template represents a predefined package set for a project archetype.
type Template struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Packages    []string `json:"packages"`
}

// builtinTemplates is the ordered list of templates shown in the interactive
// selector. Order matters — it determines display order.
var builtinTemplates = []Template{
	{
		Name:        "minimal",
		Description: "Bare essentials for any project",
		Packages:    []string{"git-workflow"},
	},
	{
		Name:        "standard",
		Description: "Good defaults for most projects",
		Packages:    []string{"git-workflow", "developer-mentor", "raise-pull-requests"},
	},
	{
		Name:        "security",
		Description: "Security-conscious development",
		Packages:    []string{"threat-modelling", "git-workflow", "raise-pull-requests"},
	},
	{
		Name:        "fullstack",
		Description: "Everything included",
		Packages:    nil, // resolved at runtime to all available packages
	},
	{
		Name:        "docs",
		Description: "Documentation-focused projects",
		Packages:    []string{"creating-documentation", "developer-mentor"},
	},
}

// defaultTemplateName is the template used when --yes is set without --template.
const defaultTemplateName = "standard"

// getTemplate returns the named template, or an error if not found.
func getTemplate(name string) (*Template, error) {
	for i := range builtinTemplates {
		if builtinTemplates[i].Name == name {
			return &builtinTemplates[i], nil
		}
	}
	return nil, fmt.Errorf("unknown template %q, available: %s", name, strings.Join(templateNames(), ", "))
}

// listTemplates returns all built-in templates.
func listTemplates() []Template {
	return builtinTemplates
}

// templateNames returns just the names of all built-in templates.
func templateNames() []string {
	names := make([]string, len(builtinTemplates))
	for i, t := range builtinTemplates {
		names[i] = t.Name
	}
	return names
}

// resolveTemplatePackages returns the package names for a template.
// For the "fullstack" template (nil Packages), it returns all available
// package names from the provided list.
func resolveTemplatePackages(t *Template, available []model.Package) []string {
	if t.Packages != nil {
		return t.Packages
	}
	// nil Packages means "all available" (used by fullstack)
	names := make([]string, len(available))
	for i, pkg := range available {
		names[i] = pkg.Name
	}
	return names
}
