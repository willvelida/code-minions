package cmd

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/willvelida/code-minions/internal/assistant"
	"github.com/willvelida/code-minions/internal/installer"
	"github.com/willvelida/code-minions/internal/mcp"
	"github.com/willvelida/code-minions/internal/model"
	"github.com/willvelida/code-minions/internal/registry"
	"gopkg.in/yaml.v3"
)

// teamInstallOptions bundles all the resolved flags for team install.
type teamInstallOptions struct {
	forFlag     string
	fileFlag    string
	fromFlag    string
	targetFlag  string
	personaFlag string
	force       bool
	dryRun      bool
	mode        OutputMode
}

// teamInstallResult is the JSON output for the team install command.
type teamInstallResult struct {
	Team                 string                     `json:"team"`
	Assistant            string                     `json:"assistant"`
	File                 string                     `json:"file"`
	Personas             []teamInstallPersonaResult `json:"personas"`
	MCPServers           []string                   `json:"mcp_servers"`
	InstructionsInjected bool                       `json:"instructions_injected"`
	Summary              teamInstallSummary         `json:"summary"`
}

// teamInstallPersonaResult captures per-persona results in JSON output.
type teamInstallPersonaResult struct {
	Name           string                     `json:"name"`
	Packages       []teamInstallPackageResult `json:"packages"`
	GeneratedFiles []string                   `json:"generated_files"`
	Errors         []string                   `json:"errors"`
}

// teamInstallPackageResult captures per-package results in JSON output.
type teamInstallPackageResult struct {
	Name    string `json:"name"`
	Copied  int    `json:"copied"`
	Skipped int    `json:"skipped"`
	Errors  int    `json:"errors"`
}

// teamInstallSummary is the summary section in JSON output.
type teamInstallSummary struct {
	Personas  int `json:"personas"`
	Packages  int `json:"packages"`
	Copied    int `json:"copied"`
	Generated int `json:"generated"`
	Skipped   int `json:"skipped"`
	Errors    int `json:"errors"`
}

// personaInstallEntry pairs a persona name with its install result.
type personaInstallEntry struct {
	name   string
	result *installer.PersonaResult
}

func newTeamInstallCommand(content fs.FS) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install personas and packages from a team configuration",
		Long: `Install reads a team.yaml file and installs all defined
personas with their packages for the specified coding assistant.

Each persona's packages are installed, assistant-specific grouping
artefacts are generated, and MCP servers are configured. The
installation is recorded in .code-minions/installed.json for
tracking and uninstall support.

The --for flag specifies the target assistant. If omitted, the
default_assistant from team.yaml's config section is used.`,
		Example: `  # Install from team.yaml for GitHub Copilot
  code-minions team install --for copilot

  # Use default assistant from team.yaml
  code-minions team install

  # Install with packages from an external source
  code-minions team install --for copilot --from github.com/org/packages

  # Install only a specific persona
  code-minions team install --for copilot --persona back-end

  # Preview what would be installed
  code-minions team install --for copilot --dry-run

  # JSON output for scripting
  code-minions team install --for copilot --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			forFlag, _ := cmd.Flags().GetString("for")
			fileFlag, _ := cmd.Flags().GetString("file")
			fromFlag, _ := cmd.Flags().GetString("from")
			targetFlag, _ := cmd.Flags().GetString("target")
			personaFlag, _ := cmd.Flags().GetString("persona")
			force, _ := cmd.Flags().GetBool("force")
			dryRun, _ := cmd.Flags().GetBool("dry-run")

			mode := getOutputMode(cmd)

			return runTeamInstall(cmd, content, teamInstallOptions{
				forFlag:     forFlag,
				fileFlag:    fileFlag,
				fromFlag:    fromFlag,
				targetFlag:  targetFlag,
				personaFlag: personaFlag,
				force:       force,
				dryRun:      dryRun,
				mode:        mode,
			})
		},
	}

	cmd.Flags().String("for", "", "Target assistant (e.g. copilot, claude)")
	cmd.Flags().String("file", "team.yaml", "Path to team configuration file")
	cmd.Flags().String("from", "", "Package source name or Git URL")
	cmd.Flags().String("target", ".", "Target directory for installation")
	cmd.Flags().String("persona", "", "Install only a specific persona from the team")
	cmd.Flags().Bool("force", false, "Overwrite existing files")
	cmd.Flags().Bool("dry-run", false, "Preview changes without writing files")

	return cmd
}

func runTeamInstall(cmd *cobra.Command, content fs.FS, opts teamInstallOptions) error {
	// --- Step 1: Read and parse team.yaml ---
	teamFilePath := opts.fileFlag
	if !filepath.IsAbs(teamFilePath) {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
		teamFilePath = filepath.Join(cwd, teamFilePath)
	}

	data, err := os.ReadFile(teamFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s not found\n\nRun 'code-minions team init' to create a team configuration", opts.fileFlag)
		}
		return fmt.Errorf("failed to read %s: %w", opts.fileFlag, err)
	}

	var team model.Team
	if err := yaml.Unmarshal(data, &team); err != nil {
		return fmt.Errorf("failed to parse %s: %w", opts.fileFlag, err)
	}
	verbosePrintf(cmd, opts.mode, "loaded team %q from %s\n", team.Name, opts.fileFlag)

	// --- Step 2: Validate team structure ---
	if err := model.ValidateTeam(&team); err != nil {
		return fmt.Errorf("invalid team configuration: %w", err)
	}

	// --- Step 3: Resolve assistant (--for flag → team config default → error) ---
	assistantName := opts.forFlag
	if assistantName == "" {
		assistantName = team.Config.DefaultAssistant
	}
	if assistantName == "" {
		return fmt.Errorf("--for is required (no default_assistant in team config)\n\n"+
			"Usage: code-minions team install --for <assistant>\n\n"+
			"Available assistants: %s", strings.Join(assistant.List(), ", "))
	}
	if _, err := assistant.Get(assistantName); err != nil {
		return err
	}
	verbosePrintf(cmd, opts.mode, "target assistant: %s\n", assistantName)

	// --- Step 4: Filter personas if --persona is set ---
	personas := team.Personas
	if opts.personaFlag != "" {
		var filtered []model.PersonaRef
		for _, p := range personas {
			if p.Name == opts.personaFlag {
				filtered = append(filtered, p)
				break
			}
		}
		if len(filtered) == 0 {
			var names []string
			for _, p := range personas {
				names = append(names, p.Name)
			}
			return fmt.Errorf("persona %q not found in team %q\n\nAvailable personas: %s",
				opts.personaFlag, team.Name, strings.Join(names, ", "))
		}
		personas = filtered
		verbosePrintf(cmd, opts.mode, "installing persona: %s\n", opts.personaFlag)
	}

	// --- Step 5: Build registry ---
	var reg *registry.Registry
	if opts.fromFlag != "" {
		reg, err = registry.BuildRegistryWithFrom(content, opts.fromFlag)
		if err != nil {
			return fmt.Errorf("failed to resolve source %q: %w", opts.fromFlag, err)
		}
	} else {
		reg, err = registry.BuildRegistry(content)
		if err != nil {
			return fmt.Errorf("failed to build registry: %w", err)
		}
	}
	resolver := registry.NewPersonaResolver(reg)

	if opts.dryRun && (opts.mode == OutputNormal || opts.mode == OutputVerbose) {
		printDryRunBanner(cmd.OutOrStdout())
	}

	// --- Step 6: Install each persona ---
	var installedPersonas []personaInstallEntry
	var allErrors []string
	totalPackages := 0

	for _, pRef := range personas {
		// Convert PersonaRef → model.Persona for the resolver
		persona := &model.Persona{
			Name:     pRef.Name,
			Packages: pRef.Packages,
		}

		verbosePrintf(cmd, opts.mode, "resolving persona %q (%d packages)\n", pRef.Name, len(pRef.Packages))

		resolved, err := resolver.ResolvePersona(persona)
		if err != nil {
			allErrors = append(allErrors, fmt.Sprintf("persona %q: %v", pRef.Name, err))
			continue
		}

		pi := &installer.PersonaInstaller{
			Resolved:      resolved,
			Content:       content,
			AssistantName: assistantName,
			Target:        opts.targetFlag,
			Force:         opts.force,
			DryRun:        opts.dryRun,
		}

		result, err := pi.Install()
		if err != nil {
			allErrors = append(allErrors, fmt.Sprintf("persona %q: installation failed: %v", pRef.Name, err))
			continue
		}

		installedPersonas = append(installedPersonas, personaInstallEntry{
			name:   pRef.Name,
			result: result,
		})
		totalPackages += len(result.PackageResults)
	}

	// --- Step 7: Merge team-level MCP servers ---
	var teamMCPServers []string
	if team.MCP != nil && len(team.MCP.Servers) > 0 {
		translator, err := mcp.NewTranslator(assistantName)
		if err != nil {
			verbosePrintf(cmd, opts.mode, "skipping team MCP: %v\n", err)
		} else {
			servers, warnings, err := translator.Translate(team.MCP)
			if err != nil {
				allErrors = append(allErrors, fmt.Sprintf("team MCP translation failed: %v", err))
			} else {
				for _, w := range warnings {
					allErrors = append(allErrors, fmt.Sprintf("team MCP warning: %s", w))
				}
				mcpResult, err := mcp.InstallServers(opts.targetFlag, translator, servers, opts.force, opts.dryRun)
				if err != nil {
					allErrors = append(allErrors, fmt.Sprintf("team MCP merge failed: %v", err))
				} else if mcpResult != nil {
					for name := range servers {
						teamMCPServers = append(teamMCPServers, name)
					}
					sort.Strings(teamMCPServers)
				}
			}
		}
	}

	// --- Step 8: Inject team-level instructions ---
	var instructionsInjected bool
	var instructionsFile string
	if team.Instructions != "" && !opts.dryRun {
		assistantCfg, _ := assistant.Get(assistantName)
		instrPath := assistantCfg.InstructionsPath
		absPath, err := installer.InjectTeamInstructions(
			opts.targetFlag, instrPath, team.Name, team.Instructions,
		)
		if err != nil {
			allErrors = append(allErrors, fmt.Sprintf("failed to inject team instructions: %v", err))
		} else if absPath != "" {
			instructionsInjected = true
			instructionsFile = instrPath
			verbosePrintf(cmd, opts.mode, "injected team instructions into %s\n", instrPath)
		}
	}

	// --- Step 9: Record team install in manifest ---
	if !opts.dryRun && len(installedPersonas) > 0 {
		manifest, err := installer.LoadManifest(opts.targetFlag)
		if err != nil {
			allErrors = append(allErrors, fmt.Sprintf("failed to load manifest: %v", err))
		} else {
			personaNames := make([]string, len(installedPersonas))
			for i, p := range installedPersonas {
				personaNames[i] = p.name
			}

			installer.RecordTeamInstall(
				manifest,
				team.Name,
				"",
				"team.yaml",
				assistantName,
				personaNames,
				teamMCPServers,
				instructionsInjected,
				instructionsFile,
			)

			if err := installer.SaveManifest(opts.targetFlag, manifest); err != nil {
				allErrors = append(allErrors, fmt.Sprintf("failed to save manifest: %v", err))
			}
		}
	}

	// --- Step 10: Format output ---
	return formatTeamInstallResult(cmd, opts, &team, installedPersonas, teamMCPServers, instructionsInjected, allErrors)
}

// formatTeamInstallResult renders the team install result in the
// appropriate output mode (JSON, quiet, normal/verbose).
func formatTeamInstallResult(
	cmd *cobra.Command,
	opts teamInstallOptions,
	team *model.Team,
	installedPersonas []personaInstallEntry,
	teamMCPServers []string,
	instructionsInjected bool,
	allErrors []string,
) error {
	// Compute totals
	totalCopied := 0
	totalSkipped := 0
	totalGenerated := 0
	totalErrors := len(allErrors)
	totalPackages := 0

	for _, p := range installedPersonas {
		totalCopied += p.result.TotalCopied()
		totalSkipped += p.result.TotalSkipped()
		totalGenerated += len(p.result.GeneratedFiles)
		totalErrors += p.result.TotalErrors()
		totalPackages += len(p.result.PackageResults)
	}

	// --- JSON output ---
	if opts.mode == OutputJSON {
		result := teamInstallResult{
			Team:                 team.Name,
			Assistant:            opts.forFlag,
			File:                 opts.fileFlag,
			MCPServers:           nonNil(teamMCPServers),
			InstructionsInjected: instructionsInjected,
		}

		if result.Assistant == "" {
			result.Assistant = team.Config.DefaultAssistant
		}

		for _, p := range installedPersonas {
			pr := teamInstallPersonaResult{
				Name:           p.name,
				GeneratedFiles: nonNil(p.result.GeneratedFiles),
				Errors:         nonNil(p.result.Errors),
			}

			pkgNames := sortedKeys(p.result.PackageResults)
			for _, pkgName := range pkgNames {
				pkgResult := p.result.PackageResults[pkgName]
				pr.Packages = append(pr.Packages, teamInstallPackageResult{
					Name:    pkgName,
					Copied:  len(pkgResult.Copied),
					Skipped: len(pkgResult.Skipped),
					Errors:  len(pkgResult.Errors),
				})
			}
			if pr.Packages == nil {
				pr.Packages = []teamInstallPackageResult{}
			}

			result.Personas = append(result.Personas, pr)
		}
		if result.Personas == nil {
			result.Personas = []teamInstallPersonaResult{}
		}

		result.Summary = teamInstallSummary{
			Personas:  len(installedPersonas),
			Packages:  totalPackages,
			Copied:    totalCopied,
			Generated: totalGenerated,
			Skipped:   totalSkipped,
			Errors:    totalErrors,
		}

		if err := json.NewEncoder(cmd.OutOrStdout()).Encode(result); err != nil {
			return err
		}
		if totalErrors > 0 {
			return fmt.Errorf("team installation completed with %d errors", totalErrors)
		}
		return nil
	}

	// --- Quiet output ---
	if opts.mode == OutputQuiet {
		for _, e := range allErrors {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "  error: %s\n", e)
		}
		for _, p := range installedPersonas {
			for _, e := range p.result.Errors {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "  error: %s\n", e)
			}
			for _, pr := range p.result.PackageResults {
				for _, e := range pr.Errors {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "  error: %s\n", e)
				}
			}
		}
		if totalErrors > 0 {
			return fmt.Errorf("team installation completed with %d errors", totalErrors)
		}
		return nil
	}

	// --- Normal / Verbose output ---
	out := cmd.OutOrStdout()
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	red := color.New(color.FgRed)
	bold := color.New(color.Bold)
	cyan := color.New(color.FgCyan)

	assistantName := opts.forFlag
	if assistantName == "" {
		assistantName = team.Config.DefaultAssistant
	}

	_, _ = bold.Fprintf(out, "Installing team %q for %s\n\n", team.Name, assistantName)

	for _, p := range installedPersonas {
		_, _ = cyan.Fprintf(out, "  Persona: %s\n", p.name)

		pkgNames := sortedKeys(p.result.PackageResults)
		for _, pkgName := range pkgNames {
			pr := p.result.PackageResults[pkgName]
			_, _ = fmt.Fprintf(out, "    Package: %s\n", pkgName)
			for _, f := range pr.Copied {
				if opts.dryRun {
					_, _ = yellow.Fprintf(out, "      would copy: %s\n", f)
				} else {
					_, _ = green.Fprintf(out, "      copied: %s\n", f)
				}
			}
			for _, f := range pr.Skipped {
				_, _ = yellow.Fprintf(out, "      skipped (exists): %s\n", f)
			}
			for _, e := range pr.Errors {
				_, _ = red.Fprintf(cmd.ErrOrStderr(), "      error: %s\n", e)
			}
		}

		if len(p.result.GeneratedFiles) > 0 {
			for _, f := range p.result.GeneratedFiles {
				if opts.dryRun {
					_, _ = yellow.Fprintf(out, "    would generate: %s\n", f)
				} else {
					_, _ = green.Fprintf(out, "    generated: %s\n", f)
				}
			}
		}

		for _, e := range p.result.Errors {
			_, _ = red.Fprintf(cmd.ErrOrStderr(), "    error: %s\n", e)
		}

		_, _ = fmt.Fprintln(out)
	}

	// Team-level errors
	for _, e := range allErrors {
		_, _ = red.Fprintf(cmd.ErrOrStderr(), "  error: %s\n", e)
	}

	// Summary
	if len(installedPersonas) > 0 {
		_, _ = green.Fprintf(out, "✓ Installed team %q\n", team.Name)
		_, _ = bold.Fprintf(out, "  Personas: %s\n", formatInstalledPersonaSummary(installedPersonas))
		_, _ = bold.Fprintf(out, "  Total: %d copied, %d generated, %d skipped, %d errors\n",
			totalCopied, totalGenerated, totalSkipped, totalErrors)
		_, _ = fmt.Fprintln(out)
	}

	if totalErrors > 0 {
		return fmt.Errorf("team installation completed with %d errors", totalErrors)
	}

	return nil
}

// formatInstalledPersonaSummary returns a human-readable summary like
// "front-end (1 package), back-end (3 packages)".
func formatInstalledPersonaSummary(personas []personaInstallEntry) string {
	parts := make([]string, len(personas))
	for i, p := range personas {
		count := len(p.result.PackageResults)
		noun := "packages"
		if count == 1 {
			noun = "package"
		}
		parts[i] = fmt.Sprintf("%s (%d %s)", p.name, count, noun)
	}
	return strings.Join(parts, ", ")
}
