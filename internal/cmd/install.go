package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/willvelida/code-minions/internal/assistant"
	"github.com/willvelida/code-minions/internal/installer"
	"github.com/willvelida/code-minions/internal/mcp"
	"github.com/willvelida/code-minions/internal/registry"
)

func newInstallCommand(content fs.FS) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install agents and skills into your repository",
		Long: `Install copies agent definitions and skill files from the built-in
package registry into your repository. Files are placed relative to the
target directory, organised into agents/ and skills/ directories.

When --for is specified, files are placed in the assistant-specific
location (e.g. .github/agents/ for GitHub Copilot, .claude/agents/ for Claude).

An AGENTS.md routing file is created at the root of the target to help
AI assistants discover installed agents. If AGENTS.md already exists,
it is left unchanged.

Existing files are skipped unless --force is used. Use --dry-run to
preview changes without writing any files.`,
		Example: `  # Install all packages into the current directory
  code-minions install

  # Install for GitHub Copilot (places files in .github/agents/)
  code-minions install --for copilot

  # Install a specific package
  code-minions install --package git-workflow

  # Preview what would be installed
  code-minions install --dry-run

  # Overwrite existing files
  code-minions install --force

  # Install a persona (a bundle of packages)
  code-minions install --persona senior-dev --for copilot

  # Preview a persona install
  code-minions install --persona senior-dev --for copilot --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			force, _ := cmd.Flags().GetBool("force")
			packageFlag, _ := cmd.Flags().GetString("package")
			forFlag, _ := cmd.Flags().GetString("for")
			personaFlag, _ := cmd.Flags().GetString("persona")

			// --- Persona install branch ---
			// When --persona is set, we take a completely different path:
			// resolve the persona, install all its packages via PersonaInstaller,
			// and generate the grouping artefact.
			if personaFlag != "" {
				// --persona and --package are mutually exclusive.
				// You're either installing a persona OR individual packages.
				if packageFlag != "" {
					return fmt.Errorf("--persona and --package cannot be used together\n\n" +
						"Use --persona to install a bundle of packages, or --package for individual packages")
				}
				// --for is required for persona installs because the grouping
				// artefact is assistant-specific (AGENTS.md vs .agent.md etc.)
				if forFlag == "" {
					return fmt.Errorf("--for is required when installing a persona\n\n"+
						"Usage: code-minions install --persona <name> --for <assistant>\n\n"+
						"Available assistants: %s", strings.Join(assistant.List(), ", "))
				}
				return runPersonaInstall(cmd, content, personaFlag, forFlag, target, force, dryRun)
			}

			// If --for is set, look up the assistant config and build a path mapper
			var pathMapper func(string) string
			if forFlag != "" {
				cfg, err := assistant.Get(forFlag)
				if err != nil {
					return err
				}
				pathMapper = cfg.NewPathMapper()
			}

			// Build the list of package directories to install
			packageDirs, err := buildPackageList(content, packageFlag)
			if err != nil {
				return err
			}

			mode := getOutputMode(cmd)

			if dryRun && (mode == OutputNormal || mode == OutputVerbose) {
				_, _ = color.New(color.FgYellow, color.Bold).Println("Dry run - no files will be written")
				fmt.Println()
			}

			// Install each package (with prefix stripping)
			combinedResult := &installer.Result{}
			for _, pkgDir := range packageDirs {
				inst := &installer.Installer{
					Content:     content,
					Target:      target,
					Force:       force,
					DryRun:      dryRun,
					StripPrefix: pkgDir,
					PathMapper:  pathMapper,
				}
				result, err := inst.Install([]string{pkgDir})
				if err != nil {
					return fmt.Errorf("installation failed: %w", err)
				}
				combinedResult.Copied = append(combinedResult.Copied, result.Copied...)
				combinedResult.Skipped = append(combinedResult.Skipped, result.Skipped...)
				combinedResult.Errors = append(combinedResult.Errors, result.Errors...)
			}

			// Create AGENTS.md if installing packages
			if len(packageDirs) > 0 {
				agentsMDPath := "AGENTS.md"
				if pathMapper != nil {
					agentsMDPath = pathMapper("agents/AGENTS.md")
				}

				handlerStdout := io.Writer(os.Stdout)
				if mode == OutputJSON || mode == OutputQuiet {
					handlerStdout = io.Discard
				}
				handler := &installer.AgentsMDHandler{
					Target: target,
					DryRun: dryRun,
					Stdin:  os.Stdin,
					Stdout: handlerStdout,
				}

				action, err := handler.OnInstall(agentsMDPath, []byte(installer.DefaultAgentsMDContent))
				if err != nil {
					combinedResult.Errors = append(combinedResult.Errors, fmt.Sprintf("AGENTS.md: %v", err))
				} else if action == "created" {
					combinedResult.Copied = append(combinedResult.Copied, agentsMDPath)
				} else {
					combinedResult.Skipped = append(combinedResult.Skipped, agentsMDPath)
				}
			}

			// Create assistant-specific instructions file (e.g. CLAUDE.md)
			// when --for is set and the assistant has an InstructionsPath
			// distinct from AGENTS.md.
			if forFlag != "" && len(packageDirs) > 0 {
				cfg, cfgErr := assistant.Get(forFlag)
				if cfgErr == nil && cfg.InstructionsPath != "" && cfg.InstructionsPath != "AGENTS.md" {
					instrStdout := io.Writer(os.Stdout)
					if mode == OutputJSON || mode == OutputQuiet {
						instrStdout = io.Discard
					}
					instrHandler := &installer.InstructionsFileHandler{
						Target:   target,
						DryRun:   dryRun,
						FileName: cfg.InstructionsPath,
						Stdin:    os.Stdin,
						Stdout:   instrStdout,
					}
					// Build package names for content generation
					var pkgNames []string
					for _, pkgDir := range packageDirs {
						pkgNames = append(pkgNames, strings.TrimPrefix(pkgDir, "packages/"))
					}
					instrContent := installer.BuildClaudeMDForPackages(pkgNames, cfg)
					instrAction, instrErr := instrHandler.OnInstall([]byte(instrContent))
					if instrErr != nil {
						combinedResult.Errors = append(combinedResult.Errors,
							fmt.Sprintf("%s: %v", cfg.InstructionsPath, instrErr))
					} else if instrAction == "created" {
						combinedResult.Copied = append(combinedResult.Copied, cfg.InstructionsPath)
					} else {
						combinedResult.Skipped = append(combinedResult.Skipped, cfg.InstructionsPath)
					}
				}
			}

			// --- MCP server processing ---
			// When --for is set, translate each package's mcp.yaml into
			// the assistant's native JSON format and merge it in.
			mcpResults := make(map[string]*mcp.InstallResult) // package name → result
			if forFlag != "" {
				translator, err := mcp.NewTranslator(forFlag)
				if err != nil {
					combinedResult.Errors = append(combinedResult.Errors, fmt.Sprintf("MCP: %v", err))
				} else {
					for _, pkgDir := range packageDirs {
						pkgName := strings.TrimPrefix(pkgDir, "packages/")
						mcpResult, err := mcp.Install(content, pkgDir, target, translator, force, dryRun)
						if err != nil {
							combinedResult.Errors = append(combinedResult.Errors, fmt.Sprintf("MCP (%s): %v", pkgName, err))
							continue
						}
						if mcpResult != nil {
							mcpResults[pkgName] = mcpResult
						}
					}
				}
			}

			// Record installation in manifest (skip for dry-run)
			if !dryRun && (len(combinedResult.Copied) > 0 || len(mcpResults) > 0) {
				src := registry.NewEmbeddedSource(content)
				manifest, err := installer.LoadManifest(target)
				if err != nil {
					combinedResult.Errors = append(combinedResult.Errors, fmt.Sprintf("manifest load: %v", err))
				} else {
					for _, pkgDir := range packageDirs {
						pkgName := strings.TrimPrefix(pkgDir, "packages/")
						var version string
						if pkg, err := src.GetPackage(pkgName); err == nil {
							version = pkg.Version
						}
						// Collect files that belong to this package
						var pkgFiles []string
						for _, f := range combinedResult.Copied {
							if strings.HasPrefix(f, "agents/"+pkgName+"/") ||
								strings.HasPrefix(f, "agents/"+pkgName+".") ||
								strings.HasPrefix(f, "skills/"+pkgName+"/") ||
								strings.HasPrefix(f, ".github/agents/"+pkgName+"/") ||
								strings.HasPrefix(f, ".github/agents/"+pkgName+".") ||
								strings.HasPrefix(f, ".claude/agents/"+pkgName+"/") ||
								strings.HasPrefix(f, ".claude/agents/"+pkgName+".") ||
								strings.HasPrefix(f, ".claude/skills/"+pkgName+"/") ||
								strings.HasPrefix(f, ".opencode/agents/"+pkgName+"/") ||
								strings.HasPrefix(f, ".opencode/agents/"+pkgName+".") ||
								strings.HasPrefix(f, ".opencode/skills/"+pkgName+"/") {
								pkgFiles = append(pkgFiles, f)
							}
						}
						installer.RecordInstall(manifest, pkgName, version, "embedded", forFlag, pkgFiles)

						// Attach only actually-added MCP server names to the manifest entry.
						// RecordInstall creates a fresh entry (without MCPServers), so we
						// locate it via FindInstalled and set the field directly.
						if mcpResult, ok := mcpResults[pkgName]; ok && len(mcpResult.Merge.Added) > 0 {
							addedServers := make([]string, len(mcpResult.Merge.Added))
							copy(addedServers, mcpResult.Merge.Added)
							sort.Strings(addedServers)
							if entry := installer.FindInstalled(manifest, pkgName); entry != nil {
								entry.MCPServers = addedServers
							}
						}
					}
					if err := installer.SaveManifest(target, manifest); err != nil {
						combinedResult.Errors = append(combinedResult.Errors, fmt.Sprintf("manifest: %v", err))
					}
				}
			}

			// JSON output
			if mode == OutputJSON {
				copied := combinedResult.Copied
				if copied == nil {
					copied = []string{}
				}
				skipped := combinedResult.Skipped
				if skipped == nil {
					skipped = []string{}
				}
				errs := combinedResult.Errors
				if errs == nil {
					errs = []string{}
				}

				// Build MCP section for JSON
				type mcpJSONEntry struct {
					Package    string   `json:"package"`
					ConfigPath string   `json:"config_path"`
					Added      []string `json:"added"`
					Skipped    []string `json:"skipped"`
					Conflicts  []string `json:"conflicts"`
					Warnings   []string `json:"warnings"`
				}
				var mcpEntries []mcpJSONEntry
				mcpPkgNames := sortedKeys(mcpResults)
				for _, pkgName := range mcpPkgNames {
					mr := mcpResults[pkgName]
					entry := mcpJSONEntry{
						Package:    pkgName,
						ConfigPath: mr.ConfigPath,
						Added:      nonNil(mr.Merge.Added),
						Skipped:    nonNil(mr.Merge.Skipped),
						Conflicts:  nonNil(mr.Merge.Conflict),
						Warnings:   nonNil(mr.Merge.Warnings),
					}
					mcpEntries = append(mcpEntries, entry)
				}
				if mcpEntries == nil {
					mcpEntries = []mcpJSONEntry{}
				}

				result := struct {
					Copied  []string       `json:"copied"`
					Skipped []string       `json:"skipped"`
					Errors  []string       `json:"errors"`
					MCP     []mcpJSONEntry `json:"mcp"`
					Summary struct {
						Copied  int `json:"copied"`
						Skipped int `json:"skipped"`
						Errors  int `json:"errors"`
					} `json:"summary"`
				}{
					Copied:  copied,
					Skipped: skipped,
					Errors:  errs,
					MCP:     mcpEntries,
				}
				result.Summary.Copied = len(combinedResult.Copied)
				result.Summary.Skipped = len(combinedResult.Skipped)
				result.Summary.Errors = len(combinedResult.Errors)
				if err := json.NewEncoder(cmd.OutOrStdout()).Encode(result); err != nil {
					return err
				}
				if len(combinedResult.Errors) > 0 {
					return fmt.Errorf("installation completed with %d errors", len(combinedResult.Errors))
				}
				return nil
			}

			// Quiet mode — only report errors to stderr
			if mode == OutputQuiet {
				for _, e := range combinedResult.Errors {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "  error: %s\n", e)
				}
				if len(combinedResult.Errors) > 0 {
					return fmt.Errorf("installation completed with %d errors", len(combinedResult.Errors))
				}
				return nil
			}

			green := color.New(color.FgGreen)
			yellow := color.New(color.FgYellow)
			red := color.New(color.FgRed)
			bold := color.New(color.Bold)

			// Verbose: show package list
			verbosePrintf(cmd, mode, "packages: %v\n", packageDirs)

			// Print results
			for _, f := range combinedResult.Copied {
				if dryRun {
					_, _ = yellow.Printf("  would copy: %s\n", f)
				} else {
					_, _ = green.Printf("  copied: %s\n", f)
				}
			}
			for _, f := range combinedResult.Skipped {
				_, _ = yellow.Printf("  skipped (exists): %s\n", f)
				verbosePrintf(cmd, mode, "    → use --force to overwrite\n")
			}
			for _, e := range combinedResult.Errors {
				_, _ = red.Fprintf(os.Stderr, "  error: %s\n", e)
			}

			// MCP server results
			if len(mcpResults) > 0 {
				cyan := color.New(color.FgCyan)
				fmt.Println()
				_, _ = bold.Println("MCP servers:")
				mcpPkgNames := sortedKeys(mcpResults)
				for _, pkgName := range mcpPkgNames {
					mr := mcpResults[pkgName]
					for _, s := range mr.Merge.Added {
						if dryRun {
							_, _ = yellow.Printf("  would add to %s: %s (package: %s)\n", mr.ConfigPath, s, pkgName)
						} else {
							_, _ = cyan.Printf("  added to %s: %s (package: %s)\n", mr.ConfigPath, s, pkgName)
						}
					}
					for _, s := range mr.Merge.Skipped {
						_, _ = yellow.Printf("  skipped (identical): %s in %s\n", s, mr.ConfigPath)
					}
					for _, s := range mr.Merge.Conflict {
						_, _ = yellow.Printf("  conflict: %s in %s (use --force to overwrite)\n", s, mr.ConfigPath)
					}
					for _, w := range mr.Merge.Warnings {
						_, _ = yellow.Printf("  warning: %s\n", w)
					}
				}
			}

			// Summary
			fmt.Println()
			_, _ = bold.Printf("%d copied, %d skipped, %d errors\n",
				len(combinedResult.Copied), len(combinedResult.Skipped), len(combinedResult.Errors))

			if len(combinedResult.Errors) > 0 {
				return fmt.Errorf("installation completed with %d errors", len(combinedResult.Errors))
			}

			return nil
		},
	}

	cmd.Flags().String("target", ".", "Target directory for installation")
	cmd.Flags().String("package", "", "Comma-separated list of packages to install (omit to install all)")
	cmd.Flags().String("persona", "", "Install a persona (a named bundle of packages)")
	cmd.Flags().String("for", "", "Target coding assistant ("+assistant.FlagUsage()+")")
	cmd.Flags().Bool("dry-run", false, "Show what would be installed without writing files")
	cmd.Flags().Bool("force", false, "Overwrite existing files")

	return cmd
}

func buildPackageList(content fs.FS, packageFlag string) ([]string, error) {
	// No flag = install all packages
	if packageFlag == "" {
		var dirs []string
		packages, err := listSubDirs(content, "packages")
		if err != nil {
			return nil, fmt.Errorf("failed to list packages: %w", err)
		}
		for _, pkg := range packages {
			dirs = append(dirs, "packages/"+pkg)
		}
		return dirs, nil
	}

	var dirs []string
	packages := strings.Split(packageFlag, ",")
	for _, pkg := range packages {
		pkg = strings.TrimSpace(pkg)
		if pkg == "" {
			continue
		}
		dirPath := "packages/" + pkg

		if _, err := fs.Stat(content, dirPath); err != nil {
			available, err := listSubDirs(content, "packages")
			if err != nil {
				return nil, fmt.Errorf("package %q not found", pkg)
			}
			return nil, fmt.Errorf("package %q not found\n\nAvailable packages:\n  %s",
				pkg, strings.Join(available, "\n  "))
		}

		dirs = append(dirs, dirPath)
	}

	// De-duplicate dirs
	seen := make(map[string]bool)
	unique := dirs[:0]
	for _, d := range dirs {
		if !seen[d] {
			seen[d] = true
			unique = append(unique, d)
		}
	}

	return unique, nil
}

func listSubDirs(content fs.FS, dir string) ([]string, error) {
	entries, err := fs.ReadDir(content, dir)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	return names, nil
}

// runPersonaInstall handles the `install --persona <name> --for <assistant>`
// path. This is separate from the main install logic because persona
// installation uses a different flow:
//
//  1. Resolve the persona (look up all its packages)
//  2. Install all packages via PersonaInstaller
//  3. Generate the assistant-specific grouping artefact
//  4. Record everything in the manifest
//
// The PersonaInstaller handles steps 2-4 internally (we built it in
// Phase 2 and wired manifest tracking in Phase 3). This function
// just handles the CLI concerns: parsing flags, resolving the persona,
// and formatting output.
func runPersonaInstall(
	cmd *cobra.Command,
	content fs.FS,
	personaName, assistantName, target string,
	force, dryRun bool,
) error {
	mode := getOutputMode(cmd)

	// --- Step 1: Resolve the persona ---
	// Create a registry with the embedded source, then use the
	// PersonaResolver to look up the persona and all its packages.
	src := registry.NewEmbeddedSource(content)
	reg := registry.NewRegistry(src)
	resolver := registry.NewPersonaResolver(reg)

	resolved, err := resolver.Resolve(personaName)
	if err != nil {
		return fmt.Errorf("failed to resolve persona %q: %w", personaName, err)
	}

	if dryRun && (mode == OutputNormal || mode == OutputVerbose) {
		_, _ = color.New(color.FgYellow, color.Bold).Fprintln(cmd.OutOrStdout(), "Dry run - no files will be written")
		_, _ = fmt.Fprintln(cmd.OutOrStdout())
	}

	// --- Step 2: Install via PersonaInstaller ---
	pi := &installer.PersonaInstaller{
		Resolved:      resolved,
		Content:       content,
		AssistantName: assistantName,
		Target:        target,
		Force:         force,
		DryRun:        dryRun,
	}

	result, err := pi.Install()
	if err != nil {
		return fmt.Errorf("persona installation failed: %w", err)
	}

	// --- Step 3: Format output ---
	return formatPersonaResult(cmd, mode, personaName, assistantName, result, dryRun)
}

// formatPersonaResult renders the persona install result in the
// appropriate output mode (JSON, quiet, normal/verbose).
func formatPersonaResult(
	cmd *cobra.Command,
	mode OutputMode,
	personaName, assistantName string,
	result *installer.PersonaResult,
	dryRun bool,
) error {
	// --- JSON output ---
	if mode == OutputJSON {
		type pkgJSON struct {
			Name    string   `json:"name"`
			Copied  []string `json:"copied"`
			Skipped []string `json:"skipped"`
			Errors  []string `json:"errors"`
		}
		var pkgs []pkgJSON
		for name, pr := range result.PackageResults {
			pkgs = append(pkgs, pkgJSON{
				Name:    name,
				Copied:  nonNil(pr.Copied),
				Skipped: nonNil(pr.Skipped),
				Errors:  nonNil(pr.Errors),
			})
		}
		if pkgs == nil {
			pkgs = []pkgJSON{}
		}

		output := struct {
			Persona        string    `json:"persona"`
			Assistant      string    `json:"assistant"`
			Packages       []pkgJSON `json:"packages"`
			GeneratedFiles []string  `json:"generated_files"`
			Errors         []string  `json:"errors"`
			Summary        struct {
				Copied    int `json:"copied"`
				Skipped   int `json:"skipped"`
				Errors    int `json:"errors"`
				Generated int `json:"generated"`
			} `json:"summary"`
		}{
			Persona:        personaName,
			Assistant:      assistantName,
			Packages:       pkgs,
			GeneratedFiles: nonNil(result.GeneratedFiles),
			Errors:         nonNil(result.Errors),
		}
		output.Summary.Copied = result.TotalCopied()
		output.Summary.Skipped = result.TotalSkipped()
		output.Summary.Errors = result.TotalErrors()
		output.Summary.Generated = len(result.GeneratedFiles)

		if err := json.NewEncoder(cmd.OutOrStdout()).Encode(output); err != nil {
			return err
		}
		if result.TotalErrors() > 0 {
			return fmt.Errorf("persona installation completed with %d errors", result.TotalErrors())
		}
		return nil
	}

	// --- Quiet output ---
	if mode == OutputQuiet {
		for _, e := range result.Errors {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "  error: %s\n", e)
		}
		for _, pr := range result.PackageResults {
			for _, e := range pr.Errors {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "  error: %s\n", e)
			}
		}
		if result.TotalErrors() > 0 {
			return fmt.Errorf("persona installation completed with %d errors", result.TotalErrors())
		}
		return nil
	}

	// --- Normal / Verbose output ---
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	red := color.New(color.FgRed)
	bold := color.New(color.Bold)
	cyan := color.New(color.FgCyan)

	_, _ = bold.Fprintf(cmd.OutOrStdout(), "Installing persona %q for %s\n\n", personaName, assistantName)

	// Per-package results (sorted for deterministic output)
	pkgNames := make([]string, 0, len(result.PackageResults))
	for k := range result.PackageResults {
		pkgNames = append(pkgNames, k)
	}
	sort.Strings(pkgNames)
	for _, pkgName := range pkgNames {
		pr := result.PackageResults[pkgName]
		_, _ = cyan.Fprintf(cmd.OutOrStdout(), "  Package: %s\n", pkgName)
		for _, f := range pr.Copied {
			if dryRun {
				_, _ = yellow.Fprintf(cmd.OutOrStdout(), "    would copy: %s\n", f)
			} else {
				_, _ = green.Fprintf(cmd.OutOrStdout(), "    copied: %s\n", f)
			}
		}
		for _, f := range pr.Skipped {
			_, _ = yellow.Fprintf(cmd.OutOrStdout(), "    skipped (exists): %s\n", f)
		}
		for _, e := range pr.Errors {
			_, _ = red.Fprintf(cmd.ErrOrStderr(), "    error: %s\n", e)
		}
	}

	// Generated files
	if len(result.GeneratedFiles) > 0 {
		_, _ = fmt.Fprintln(cmd.OutOrStdout())
		_, _ = bold.Fprintln(cmd.OutOrStdout(), "  Generated:")
		for _, f := range result.GeneratedFiles {
			if dryRun {
				_, _ = yellow.Fprintf(cmd.OutOrStdout(), "    would generate: %s\n", f)
			} else {
				_, _ = green.Fprintf(cmd.OutOrStdout(), "    generated: %s\n", f)
			}
		}
	}

	// Persona-level errors
	for _, e := range result.Errors {
		_, _ = red.Fprintf(cmd.ErrOrStderr(), "  error: %s\n", e)
	}

	// Summary
	_, _ = fmt.Fprintln(cmd.OutOrStdout())
	_, _ = bold.Fprintf(cmd.OutOrStdout(), "%d copied, %d skipped, %d generated, %d errors\n",
		result.TotalCopied(), result.TotalSkipped(), len(result.GeneratedFiles), result.TotalErrors())

	if result.TotalErrors() > 0 {
		return fmt.Errorf("persona installation completed with %d errors", result.TotalErrors())
	}

	return nil
}
