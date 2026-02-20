package cmd

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/willvelida/code-minions/internal/assistant"
	"github.com/willvelida/code-minions/internal/manifest"
	"github.com/willvelida/code-minions/internal/model"
	"github.com/willvelida/code-minions/internal/registry"
)

// initResult is the JSON output for the init command.
type initResult struct {
	ManifestPath string   `json:"manifest_path"`
	Assistant    string   `json:"assistant"`
	Template     string   `json:"template,omitempty"`
	Packages     []string `json:"packages"`
}

func newInitCommand(content fs.FS) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Scaffold a code-minions.yml project manifest",
		Long: `Create a code-minions.yml project manifest interactively.

The init command detects which coding assistants are configured in the
workspace, lets you choose a starting template, and writes a manifest
file that declares the desired configuration.

Templates provide curated package sets for common project archetypes:
minimal, standard, security, fullstack, and docs. Use --list-templates
to see all available templates.

In non-interactive mode (--yes flag or non-TTY stdin), the "standard"
template is used by default. Use --template to select a specific one.

The generated manifest can then be used by "code-minions install" to
reproducibly install the declared packages.`,
		Example: `  # Interactive setup
  code-minions init

  # Non-interactive with defaults (uses "standard" template)
  code-minions init --yes

  # Use a specific template
  code-minions init --template security

  # List available templates
  code-minions init --list-templates

  # Non-interactive with a specific template
  code-minions init --template fullstack --assistant copilot --yes

  # Specify assistant and packages explicitly (no template)
  code-minions init --assistant copilot --packages threat-modelling,git-workflow

  # Overwrite existing manifest
  code-minions init --force

  # JSON output for scripting
  code-minions init --assistant copilot --packages threat-modelling --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			assistantFlag, _ := cmd.Flags().GetString("assistant")
			packagesFlag, _ := cmd.Flags().GetString("packages")
			templateFlag, _ := cmd.Flags().GetString("template")
			listTemplatesFlag, _ := cmd.Flags().GetBool("list-templates")
			yesFlag, _ := cmd.Flags().GetBool("yes")
			forceFlag, _ := cmd.Flags().GetBool("force")

			mode := getOutputMode(cmd)

			return runInit(cmd, content, initOptions{
				target:        target,
				assistantFlag: assistantFlag,
				packagesFlag:  packagesFlag,
				templateFlag:  templateFlag,
				listTemplates: listTemplatesFlag,
				yes:           yesFlag,
				force:         forceFlag,
				mode:          mode,
			})
		},
	}

	cmd.Flags().String("target", ".", "Target directory for the manifest")
	cmd.Flags().String("assistant", "", "Assistant to configure (skip prompt)")
	cmd.Flags().String("packages", "", "Comma-separated packages to include (skip prompt)")
	cmd.Flags().String("template", "", "Use a predefined template (minimal, standard, security, fullstack, docs)")
	cmd.Flags().Bool("list-templates", false, "List available templates and exit")
	cmd.Flags().Bool("yes", false, "Accept defaults without prompting")
	cmd.Flags().Bool("force", false, "Overwrite existing code-minions.yml")

	cmd.MarkFlagsMutuallyExclusive("template", "packages")

	return cmd
}

// initOptions bundles all the resolved flags for the init command.
type initOptions struct {
	target        string
	assistantFlag string
	packagesFlag  string
	templateFlag  string
	listTemplates bool
	yes           bool
	force         bool
	mode          OutputMode
}

func runInit(cmd *cobra.Command, content fs.FS, opts initOptions) error {
	// --- Step 0: Handle --list-templates early exit ---
	if opts.listTemplates {
		return runListTemplates(cmd, content, opts)
	}

	// Resolve absolute target path
	absTarget, err := filepath.Abs(opts.target)
	if err != nil {
		return fmt.Errorf("failed to resolve target directory: %w", err)
	}

	manifestPath := manifest.DefaultPath(absTarget)

	// --- Step 1: Check for existing manifest ---
	exists, err := manifest.Exists(manifestPath)
	if err != nil {
		return err
	}
	if exists && !opts.force {
		return fmt.Errorf("code-minions.yml already exists in %s\n\nUse --force to overwrite", absTarget)
	}

	// Determine whether we're interactive.
	// Quiet mode is non-interactive (consistent with uninstall).
	interactive := isInteractiveFunc() && !opts.yes && opts.mode != OutputJSON && opts.mode != OutputQuiet

	// --- Step 2: Detect assistants ---
	detected := assistant.Detect(absTarget)
	verbosePrintf(cmd, opts.mode, "detected assistants: %v\n", detected)

	// --- Step 2.5: Choose template ---
	src := registry.NewEmbeddedSource(content)
	pkgModels, err := src.ListPackages()
	if err != nil {
		return fmt.Errorf("failed to list packages: %w", err)
	}

	selectedTemplate, err := resolveTemplate(cmd, opts, interactive)
	if err != nil {
		return err
	}

	var templateName string
	if selectedTemplate != nil {
		templateName = selectedTemplate.Name
		verbosePrintf(cmd, opts.mode, "selected template: %s\n", templateName)
	}

	// --- Step 3: Choose assistant ---
	chosenAssistant, err := resolveAssistant(cmd, opts, detected, interactive)
	if err != nil {
		return err
	}
	verbosePrintf(cmd, opts.mode, "selected assistant: %s\n", chosenAssistant)

	// --- Step 4: Choose packages (template or manual) ---
	var chosenPackages []string
	if selectedTemplate != nil {
		chosenPackages = resolveTemplatePackages(selectedTemplate, pkgModels)
	} else {
		chosenPackages, err = resolvePackages(cmd, opts, pkgModels, interactive)
		if err != nil {
			return err
		}
	}
	verbosePrintf(cmd, opts.mode, "selected packages: %v\n", chosenPackages)

	// --- Step 5: Build the project name ---
	projectName := filepath.Base(absTarget)

	// --- Step 6: Build and write manifest ---
	m := &manifest.ProjectManifest{
		Name:      projectName,
		Assistant: chosenAssistant,
		Template:  templateName,
		Packages:  chosenPackages,
	}

	if err := manifest.Save(manifestPath, m); err != nil {
		return err
	}

	// --- Step 7: Output ---
	result := initResult{
		ManifestPath: manifest.FileName,
		Assistant:    chosenAssistant,
		Template:     templateName,
		Packages:     nonNil(chosenPackages),
	}

	if opts.mode == OutputJSON {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(result)
	}

	if opts.mode != OutputQuiet {
		out := cmd.OutOrStdout()
		bold := color.New(color.Bold)
		green := color.New(color.FgGreen)

		if templateName != "" {
			_, _ = green.Fprintf(out, "✓ Created code-minions.yml with template %q\n", templateName)
		} else {
			_, _ = green.Fprintln(out, "✓ Created code-minions.yml")
		}
		_, _ = fmt.Fprintln(out)
		_, _ = bold.Fprintf(out, "  Project:   %s\n", projectName)
		_, _ = bold.Fprintf(out, "  Assistant: %s\n", chosenAssistant)
		if templateName != "" {
			_, _ = bold.Fprintf(out, "  Template:  %s\n", templateName)
		}
		_, _ = bold.Fprintf(out, "  Packages:  %s\n", strings.Join(chosenPackages, ", "))
		_, _ = fmt.Fprintln(out)
		_, _ = fmt.Fprintln(out, "Run 'code-minions install' to install the declared packages.")
	}

	// --- Step 9: Offer to install ---
	if interactive {
		ok, err := confirmPrompt(
			cmd.InOrStdin(),
			cmd.OutOrStdout(),
			"Run 'code-minions install' now? [y/N] ",
		)
		if err != nil {
			return err
		}
		if ok {
			_, _ = fmt.Fprintln(cmd.OutOrStdout())
			// Re-invoke install by finding the root command and executing it
			root := cmd.Root()
			installArgs := []string{"install"}
			if chosenAssistant != "" {
				installArgs = append(installArgs, "--for", chosenAssistant)
			}
			// Pass the selected packages so install doesn't default to all
			if len(chosenPackages) > 0 {
				installArgs = append(installArgs, "--package", strings.Join(chosenPackages, ","))
			}
			// Ensure install runs against the same target directory
			if opts.target != "" && opts.target != "." {
				installArgs = append(installArgs, "--target", opts.target)
			}
			root.SetArgs(installArgs)
			// Reset persistent output-mode flags so install does not inherit init's mode.
			for _, name := range []string{"json", "verbose", "quiet"} {
				if f := root.PersistentFlags().Lookup(name); f != nil {
					_ = f.Value.Set("false")
				}
			}
			return root.Execute()
		}
	}

	return nil
}

// resolveAssistant determines the assistant to use, either from flags,
// defaults (for non-interactive), or by prompting the user.
func resolveAssistant(cmd *cobra.Command, opts initOptions, detected []string, interactive bool) (string, error) {
	// Explicit flag takes priority
	if opts.assistantFlag != "" {
		if _, err := assistant.Get(opts.assistantFlag); err != nil {
			return "", err
		}
		return opts.assistantFlag, nil
	}

	// Non-interactive: use first detected or default to copilot
	if !interactive {
		if len(detected) > 0 {
			return detected[0], nil
		}
		return "copilot", nil
	}

	// Interactive: prompt the user
	chosen, err := selectAssistant(
		cmd.InOrStdin(),
		cmd.OutOrStdout(),
		detected,
		assistant.List(),
	)
	if err != nil {
		return "", err
	}

	// Validate the choice
	if _, err := assistant.Get(chosen); err != nil {
		return "", err
	}

	return chosen, nil
}

// resolvePackages determines which packages to include, either from flags,
// defaults (for non-interactive), or by prompting the user.
func resolvePackages(cmd *cobra.Command, opts initOptions, available []model.Package, interactive bool) ([]string, error) {
	// Explicit flag takes priority
	if opts.packagesFlag != "" {
		return validatePackageNames(opts.packagesFlag, available)
	}

	// Non-interactive: all packages
	if !interactive {
		return allPackageNames(available), nil
	}

	// Interactive: prompt the user
	return selectPackages(cmd.InOrStdin(), cmd.OutOrStdout(), available)
}

// validatePackageNames parses a comma-separated list of package names
// and validates each one against the available packages.
func validatePackageNames(input string, available []model.Package) ([]string, error) {
	validNames := make(map[string]bool)
	for _, pkg := range available {
		validNames[pkg.Name] = true
	}

	parts := strings.Split(input, ",")
	seen := make(map[string]bool)
	var selected []string
	for _, part := range parts {
		name := strings.TrimSpace(part)
		if name == "" {
			continue
		}
		if !validNames[name] {
			validList := make([]string, len(available))
			for i, pkg := range available {
				validList[i] = pkg.Name
			}
			return nil, fmt.Errorf("unknown package %q, available: %s", name, strings.Join(validList, ", "))
		}
		if !seen[name] {
			seen[name] = true
			selected = append(selected, name)
		}
	}

	if len(selected) == 0 {
		return nil, fmt.Errorf("no packages specified")
	}

	return selected, nil
}

// resolveTemplate determines which template to use, either from flags,
// defaults (for non-interactive), or by prompting the user.
// Returns nil when the user chooses "blank" (manual package selection).
func resolveTemplate(cmd *cobra.Command, opts initOptions, interactive bool) (*Template, error) {
	// Explicit --packages flag means no template
	if opts.packagesFlag != "" {
		return nil, nil
	}

	// Explicit --template flag takes priority
	if opts.templateFlag != "" {
		return getTemplate(opts.templateFlag)
	}

	// Non-interactive: use default template
	if !interactive {
		return getTemplate(defaultTemplateName)
	}

	// Interactive: prompt the user
	return selectTemplate(cmd.InOrStdin(), cmd.OutOrStdout(), listTemplates())
}

// runListTemplates handles the --list-templates flag by printing all available
// templates and returning without creating a manifest.
func runListTemplates(cmd *cobra.Command, content fs.FS, opts initOptions) error {
	templates := listTemplates()

	// Resolve fullstack packages for display
	src := registry.NewEmbeddedSource(content)
	pkgModels, err := src.ListPackages()
	if err != nil {
		return fmt.Errorf("failed to list packages: %w", err)
	}

	if opts.mode == OutputJSON {
		// For JSON output, resolve package lists
		type templateJSON struct {
			Name        string   `json:"name"`
			Description string   `json:"description"`
			Packages    []string `json:"packages"`
		}
		var out []templateJSON
		for _, t := range templates {
			pkgs := resolveTemplatePackages(&t, pkgModels)
			out = append(out, templateJSON{
				Name:        t.Name,
				Description: t.Description,
				Packages:    pkgs,
			})
		}
		return json.NewEncoder(cmd.OutOrStdout()).Encode(out)
	}

	if opts.mode != OutputQuiet {
		w := cmd.OutOrStdout()
		_, _ = fmt.Fprintln(w, "Available templates:")
		_, _ = fmt.Fprintln(w)
		for _, t := range templates {
			pkgs := resolveTemplatePackages(&t, pkgModels)
			_, _ = fmt.Fprintf(w, "  %-12s — %s\n", t.Name, t.Description)
			_, _ = fmt.Fprintf(w, "               Packages: %s\n", strings.Join(pkgs, ", "))
			_, _ = fmt.Fprintln(w)
		}
	}

	return nil
}
