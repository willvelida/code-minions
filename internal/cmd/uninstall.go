package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/willvelida/code-minions/internal/assistant"
	"github.com/willvelida/code-minions/internal/installer"
	"github.com/willvelida/code-minions/internal/mcp"
)

func newUninstallCommand(content fs.FS) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove installed agents and skills from your repository",
		Long: `Remove previously installed agent and skill files from your repository.
Only files at paths derived from the built-in package registry are
targeted for removal. This may delete locally modified versions of
those files, but files outside these known paths are not touched.

In interactive mode a confirmation prompt is shown before any files are
deleted. Non-interactive output modes (--json, --quiet) do not prompt
and require --yes to proceed. Use --dry-run to see what would be
removed without deleting anything.

The --for flag is required when uninstalling all packages, so that
files are found in the correct assistant-specific location.`,
		Example: `  # Remove all packages (interactive confirmation prompt)
  code-minions uninstall --for copilot

  # Skip the confirmation prompt (for CI/scripts)
  code-minions uninstall --for copilot --yes

  # Remove a specific package
  code-minions uninstall --package git-workflow --for copilot

  # Preview what would be removed
  code-minions uninstall --dry-run --for copilot

  # Uninstall a persona
  code-minions uninstall --persona senior-dev --for copilot`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			packageFlag, _ := cmd.Flags().GetString("package")
			forFlag, _ := cmd.Flags().GetString("for")
			yesFlag, _ := cmd.Flags().GetBool("yes")
			personaFlag, _ := cmd.Flags().GetString("persona")

			// --- Persona uninstall branch ---
			if personaFlag != "" {
				if packageFlag != "" {
					return fmt.Errorf("--persona and --package cannot be used together")
				}
				if forFlag == "" {
					return fmt.Errorf("--for is required when uninstalling a persona\n\n" +
						"Usage: code-minions uninstall --persona <name> --for <assistant>")
				}
				return runPersonaUninstall(cmd, content, personaFlag, forFlag, target, dryRun, yesFlag)
			}

			if packageFlag == "" && forFlag == "" {
				return fmt.Errorf("when uninstalling everything, --for is required to identify the correct file locations\n\nUsage: code-minions uninstall --for <assistant>\n\nAvailable assistants: %s",
					strings.Join(assistant.List(), ", "))
			}

			var pathMapper func(string) string
			if forFlag != "" {
				cfg, err := assistant.Get(forFlag)
				if err != nil {
					return err
				}
				pathMapper = cfg.NewPathMapper()
			}

			packageDirs, err := buildPackageList(content, packageFlag)
			if err != nil {
				return err
			}

			mode := getOutputMode(cmd)

			if dryRun && (mode == OutputNormal || mode == OutputVerbose) {
				_, _ = color.New(color.FgYellow, color.Bold).Println("Dry run - no files will be removed")
				fmt.Println()
			}

			// --- Confirmation gate (skip for dry-run) ---
			if !dryRun && !yesFlag {
				// Phase 1: pre-flight scan to count files that would be removed
				scanResult := &installer.UninstallResult{}
				for _, pkgDir := range packageDirs {
					inst := &installer.Installer{
						Content:     content,
						Target:      target,
						DryRun:      true, // always dry-run for the scan
						StripPrefix: pkgDir,
						PathMapper:  pathMapper,
					}
					result, err := inst.Uninstall([]string{pkgDir})
					if err != nil {
						return fmt.Errorf("pre-flight scan failed: %w", err)
					}
					scanResult.Removed = append(scanResult.Removed, result.Removed...)
				}

				fileCount := len(scanResult.Removed)
				if fileCount > 0 {
					switch mode {
					case OutputJSON:
						// JSON mode is non-interactive — require --yes
						errResult := struct {
							Error     string `json:"error"`
							FileCount int    `json:"file_count"`
							Hint      string `json:"hint"`
						}{
							Error:     "confirmation required",
							FileCount: fileCount,
							Hint:      "use --yes to skip",
						}
						if err := json.NewEncoder(cmd.OutOrStdout()).Encode(errResult); err != nil {
							return fmt.Errorf("failed to write JSON error response: %w", err)
						}
						// Prevent Cobra from printing a second, non-JSON error line.
						cmd.SilenceErrors = true
						cmd.SilenceUsage = true
						return fmt.Errorf("confirmation required (use --yes to skip)")

					case OutputQuiet:
						// Quiet mode is non-interactive — require --yes
						cmd.SilenceErrors = true
						cmd.SilenceUsage = true
						_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "error: confirmation required — this will remove %d files (use --yes to skip)\n", fileCount)
						return fmt.Errorf("confirmation required (use --yes to skip)")

					default:
						// Normal/verbose — interactive prompt
						if !isInteractiveFunc() {
							return fmt.Errorf("confirmation required (use --yes to skip)")
						}

						promptMsg := fmt.Sprintf("This will remove %d files. Continue? [y/N]: ", fileCount)
						confirmed, err := confirmPrompt(cmd.InOrStdin(), cmd.OutOrStdout(), promptMsg)
						if err != nil {
							return err
						}
						if !confirmed {
							_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
							return nil
						}
					}
				}
			}

			// --- Phase 2: actual removal ---
			combinedResult := &installer.UninstallResult{}
			for _, pkgDir := range packageDirs {
				inst := &installer.Installer{
					Content:     content,
					Target:      target,
					DryRun:      dryRun,
					StripPrefix: pkgDir,
					PathMapper:  pathMapper,
				}
				result, err := inst.Uninstall([]string{pkgDir})
				if err != nil {
					return fmt.Errorf("uninstallation failed: %w", err)
				}
				combinedResult.Removed = append(combinedResult.Removed, result.Removed...)
				combinedResult.NotFound = append(combinedResult.NotFound, result.NotFound...)
				combinedResult.Errors = append(combinedResult.Errors, result.Errors...)
				combinedResult.DirsCleaned = append(combinedResult.DirsCleaned, result.DirsCleaned...)
			}

			// --- MCP server removal ---
			// Look up installed MCP server names from the manifest and remove
			// them from the assistant's config file.
			var mcpUninstallResults []*mcp.UninstallResult
			if forFlag != "" {
				translator, err := mcp.NewTranslator(forFlag)
				if err != nil {
					combinedResult.Errors = append(combinedResult.Errors, fmt.Sprintf("MCP: %v", err))
				} else {
					manifest, err := installer.LoadManifest(target)
					if err != nil {
						combinedResult.Errors = append(combinedResult.Errors, fmt.Sprintf("MCP manifest load: %v", err))
					} else {
						for _, pkgDir := range packageDirs {
							pkgName := strings.TrimPrefix(pkgDir, "packages/")
							// Find the package in the manifest to get its MCPServers
							var serverNames []string
							for _, p := range manifest.Packages {
								if p.Name == pkgName {
									serverNames = p.MCPServers
									break
								}
							}
							if len(serverNames) > 0 {
								mcpResult, err := mcp.Uninstall(target, translator, serverNames, dryRun)
								if err != nil {
									combinedResult.Errors = append(combinedResult.Errors, fmt.Sprintf("MCP (%s): %v", pkgName, err))
									continue
								}
								if mcpResult != nil {
									mcpUninstallResults = append(mcpUninstallResults, mcpResult)
								}
							}
						}
					}
				}
			}

			// Update manifest to remove uninstalled packages (skip for dry-run)
			if !dryRun && (len(combinedResult.Removed) > 0 || len(mcpUninstallResults) > 0) {
				manifest, err := installer.LoadManifest(target)
				if err != nil {
					combinedResult.Errors = append(combinedResult.Errors, fmt.Sprintf("manifest load: %v", err))
				} else {
					for _, pkgDir := range packageDirs {
						pkgName := strings.TrimPrefix(pkgDir, "packages/")
						installer.RecordUninstall(manifest, pkgName)
					}
					if err := installer.SaveManifest(target, manifest); err != nil {
						combinedResult.Errors = append(combinedResult.Errors, fmt.Sprintf("manifest: %v", err))
					}
				}
			}

			// JSON output — handle AGENTS.md non-interactively, then marshal
			if mode == OutputJSON {
				if len(packageDirs) > 0 {
					agentsMDPath := "AGENTS.md"
					if pathMapper != nil {
						agentsMDPath = pathMapper("agents/AGENTS.md")
					}
					handler := &installer.AgentsMDHandler{
						Target: target,
						DryRun: dryRun,
						Stdin:  strings.NewReader("n\n"), // non-interactive: decline removal
						Stdout: io.Discard,
					}
					action, err := handler.OnUninstall(agentsMDPath)
					if err != nil {
						combinedResult.Errors = append(combinedResult.Errors, fmt.Sprintf("AGENTS.md: %v", err))
					} else if action == "removed" {
						combinedResult.Removed = append(combinedResult.Removed, agentsMDPath)
					}
				}

				// Handle assistant-specific instructions file (e.g. CLAUDE.md)
				if forFlag != "" && len(packageDirs) > 0 {
					cfg, cfgErr := assistant.Get(forFlag)
					if cfgErr == nil && cfg.InstructionsPath != "" && cfg.InstructionsPath != "AGENTS.md" {
						instrHandler := &installer.InstructionsFileHandler{
							Target:   target,
							DryRun:   dryRun,
							FileName: cfg.InstructionsPath,
							Stdin:    strings.NewReader("n\n"),
							Stdout:   io.Discard,
						}
						instrAction, instrErr := instrHandler.OnUninstall()
						if instrErr != nil {
							combinedResult.Errors = append(combinedResult.Errors, fmt.Sprintf("%s: %v", cfg.InstructionsPath, instrErr))
						} else if instrAction == "removed" {
							combinedResult.Removed = append(combinedResult.Removed, cfg.InstructionsPath)
						}
					}
				}

				removed := combinedResult.Removed
				if removed == nil {
					removed = []string{}
				}
				notFound := combinedResult.NotFound
				if notFound == nil {
					notFound = []string{}
				}
				errs := combinedResult.Errors
				if errs == nil {
					errs = []string{}
				}
				dirsCleaned := combinedResult.DirsCleaned
				if dirsCleaned == nil {
					dirsCleaned = []string{}
				}
				result := struct {
					Removed     []string `json:"removed"`
					NotFound    []string `json:"not_found"`
					Errors      []string `json:"errors"`
					DirsCleaned []string `json:"dirs_cleaned"`
					Summary     struct {
						Removed  int `json:"removed"`
						NotFound int `json:"not_found"`
						Errors   int `json:"errors"`
					} `json:"summary"`
				}{
					Removed:     removed,
					NotFound:    notFound,
					Errors:      errs,
					DirsCleaned: dirsCleaned,
				}
				result.Summary.Removed = len(combinedResult.Removed)
				result.Summary.NotFound = len(combinedResult.NotFound)
				result.Summary.Errors = len(combinedResult.Errors)
				if err := json.NewEncoder(cmd.OutOrStdout()).Encode(result); err != nil {
					return err
				}
				if len(combinedResult.Errors) > 0 {
					return fmt.Errorf("uninstallation completed with %d errors", len(combinedResult.Errors))
				}
				return nil
			}

			// Quiet mode — handle AGENTS.md non-interactively, only report errors
			if mode == OutputQuiet {
				if len(packageDirs) > 0 {
					agentsMDPath := "AGENTS.md"
					if pathMapper != nil {
						agentsMDPath = pathMapper("agents/AGENTS.md")
					}
					handler := &installer.AgentsMDHandler{
						Target: target,
						DryRun: dryRun,
						Stdin:  strings.NewReader("n\n"),
						Stdout: io.Discard,
					}
					action, err := handler.OnUninstall(agentsMDPath)
					if err != nil {
						combinedResult.Errors = append(combinedResult.Errors, fmt.Sprintf("AGENTS.md: %v", err))
					} else if action == "removed" {
						combinedResult.Removed = append(combinedResult.Removed, agentsMDPath)
					}
				}

				// Handle assistant-specific instructions file (e.g. CLAUDE.md)
				if forFlag != "" && len(packageDirs) > 0 {
					cfg, cfgErr := assistant.Get(forFlag)
					if cfgErr == nil && cfg.InstructionsPath != "" && cfg.InstructionsPath != "AGENTS.md" {
						instrHandler := &installer.InstructionsFileHandler{
							Target:   target,
							DryRun:   dryRun,
							FileName: cfg.InstructionsPath,
							Stdin:    strings.NewReader("n\n"),
							Stdout:   io.Discard,
						}
						instrAction, instrErr := instrHandler.OnUninstall()
						if instrErr != nil {
							combinedResult.Errors = append(combinedResult.Errors, fmt.Sprintf("%s: %v", cfg.InstructionsPath, instrErr))
						} else if instrAction == "removed" {
							combinedResult.Removed = append(combinedResult.Removed, cfg.InstructionsPath)
						}
					}
				}

				for _, e := range combinedResult.Errors {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "  error: %s\n", e)
				}
				if len(combinedResult.Errors) > 0 {
					return fmt.Errorf("uninstallation completed with %d errors", len(combinedResult.Errors))
				}
				return nil
			}

			green := color.New(color.FgGreen)
			yellow := color.New(color.FgYellow)
			red := color.New(color.FgRed)
			dim := color.New(color.Faint)
			bold := color.New(color.Bold)

			// Verbose: show package list
			verbosePrintf(cmd, mode, "packages: %v\n", packageDirs)

			for _, f := range combinedResult.Removed {
				if dryRun {
					_, _ = yellow.Printf("  would remove: %s\n", f)
				} else {
					_, _ = green.Printf("  removed: %s\n", f)
				}
			}
			for _, f := range combinedResult.NotFound {
				_, _ = dim.Printf("  not found: %s\n", f)
				verbosePrintf(cmd, mode, "    → file does not exist in target\n")
			}
			for _, d := range combinedResult.DirsCleaned {
				_, _ = dim.Printf("  cleaned dir: %s\n", d)
			}
			for _, e := range combinedResult.Errors {
				_, _ = red.Fprintf(os.Stderr, "  error: %s\n", e)
			}

			// MCP server removal results
			if len(mcpUninstallResults) > 0 {
				cyan := color.New(color.FgCyan)
				fmt.Println()
				_, _ = bold.Println("MCP servers:")
				for _, mr := range mcpUninstallResults {
					for _, s := range mr.Removed {
						if dryRun {
							_, _ = yellow.Printf("  would remove from %s: %s\n", mr.ConfigPath, s)
						} else {
							_, _ = cyan.Printf("  removed from %s: %s\n", mr.ConfigPath, s)
						}
					}
					for _, s := range mr.NotFound {
						_, _ = dim.Printf("  not found in %s: %s\n", mr.ConfigPath, s)
					}
				}
			}

			if len(packageDirs) > 0 {
				agentsMDPath := "AGENTS.md"
				if pathMapper != nil {
					agentsMDPath = pathMapper("agents/AGENTS.md")
				}

				handler := &installer.AgentsMDHandler{
					Target: target,
					DryRun: dryRun,
					Stdin:  os.Stdin,
					Stdout: os.Stdout,
				}

				action, err := handler.OnUninstall(agentsMDPath)
				if err != nil {
					combinedResult.Errors = append(combinedResult.Errors, fmt.Sprintf("AGENTS.md: %v", err))
					_, _ = red.Fprintf(os.Stderr, "  error: %s\n", err)
				} else if action == "removed" {
					_, _ = green.Printf("  removed: %s\n", agentsMDPath)
				}
			}

			// Handle assistant-specific instructions file (e.g. CLAUDE.md)
			if forFlag != "" && len(packageDirs) > 0 {
				cfg, cfgErr := assistant.Get(forFlag)
				if cfgErr == nil && cfg.InstructionsPath != "" && cfg.InstructionsPath != "AGENTS.md" {
					instrHandler := &installer.InstructionsFileHandler{
						Target:   target,
						DryRun:   dryRun,
						FileName: cfg.InstructionsPath,
						Stdin:    os.Stdin,
						Stdout:   os.Stdout,
					}
					instrAction, instrErr := instrHandler.OnUninstall()
					if instrErr != nil {
						combinedResult.Errors = append(combinedResult.Errors, fmt.Sprintf("%s: %v", cfg.InstructionsPath, instrErr))
						_, _ = red.Fprintf(os.Stderr, "  error: %s\n", instrErr)
					} else if instrAction == "removed" {
						_, _ = green.Printf("  removed: %s\n", cfg.InstructionsPath)
					}
				}
			}

			fmt.Println()
			_, _ = bold.Printf("%d removed, %d not found, %d errors\n",
				len(combinedResult.Removed), len(combinedResult.NotFound), len(combinedResult.Errors))

			if len(combinedResult.Errors) > 0 {
				return fmt.Errorf("uninstallation completed with %d errors", len(combinedResult.Errors))
			}

			return nil
		},
	}

	cmd.Flags().String("target", ".", "Target directory to uninstall from")
	cmd.Flags().String("package", "", "Comma-separated list of packages to uninstall (omit to uninstall all)")
	cmd.Flags().String("persona", "", "Uninstall a persona (a named bundle of packages)")
	cmd.Flags().String("for", "", "Target coding assistant ("+assistant.FlagUsage()+")")
	cmd.Flags().Bool("dry-run", false, "Show what would be removed without deleting files")
	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt and proceed with removal")

	return cmd
}

// runPersonaUninstall handles the `uninstall --persona <name> --for <assistant>`
// path. It uses the manifest to find what was installed, then removes:
//
//  1. Package files that are exclusively owned by this persona
//     (shared packages are kept with a warning)
//  2. Generated grouping files (with checksum verification)
//  3. The persona entry from the manifest
func runPersonaUninstall(
	cmd *cobra.Command,
	content fs.FS,
	personaName, assistantName, target string,
	dryRun, yesFlag bool,
) error {
	mode := getOutputMode(cmd)

	// --- Step 1: Load manifest and find the persona ---
	manifest, err := installer.LoadManifest(target)
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	persona := installer.FindInstalledPersona(manifest, personaName)
	if persona == nil {
		return fmt.Errorf("persona %q is not installed (checked %s)",
			personaName, installer.ManifestPath(target))
	}

	// Validate that the requested assistant matches the installed one.
	// A persona is installed for a specific assistant — uninstalling
	// with the wrong --for flag would remove the wrong files.
	if persona.Assistant != "" && persona.Assistant != assistantName {
		return fmt.Errorf("persona %q was installed for %q, not %q",
			personaName, persona.Assistant, assistantName)
	}

	// --- Step 2: Determine what to remove ---
	// Only packages exclusively owned by this persona can be removed.
	// Shared packages are kept.
	exclusivePkgs := installer.PersonaPackages(manifest, personaName)
	sharedPkgs := findSharedPackages(persona.Packages, exclusivePkgs)

	// Count files to remove for confirmation.
	var filesToRemove int
	for _, pkgName := range exclusivePkgs {
		if pkg := installer.FindInstalled(manifest, pkgName); pkg != nil {
			filesToRemove += len(pkg.Files)
		}
	}
	filesToRemove += len(persona.GeneratedFiles)

	// --- Step 3: Confirmation (unless --yes or --dry-run) ---
	if !dryRun && !yesFlag && filesToRemove > 0 {
		switch mode {
		case OutputJSON, OutputQuiet:
			return fmt.Errorf("confirmation required (use --yes to skip)")
		default:
			if !isInteractiveFunc() {
				return fmt.Errorf("confirmation required (use --yes to skip)")
			}
			promptMsg := fmt.Sprintf("This will remove %d files from persona %q. Continue? [y/N]: ",
				filesToRemove, personaName)
			confirmed, err := confirmPrompt(cmd.InOrStdin(), cmd.OutOrStdout(), promptMsg)
			if err != nil {
				return err
			}
			if !confirmed {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
				return nil
			}
		}
	}

	if dryRun && (mode == OutputNormal || mode == OutputVerbose) {
		_, _ = color.New(color.FgYellow, color.Bold).Fprintln(cmd.OutOrStdout(), "Dry run - no files will be removed")
		_, _ = fmt.Fprintln(cmd.OutOrStdout())
	}

	// --- Step 4: Remove exclusive package files ---
	var removed, notFound, warnings []string
	var errors []string

	for _, pkgName := range exclusivePkgs {
		pkg := installer.FindInstalled(manifest, pkgName)
		if pkg == nil {
			continue
		}

		for _, filePath := range pkg.Files {
			fullPath := filepath.Join(target, filePath)
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				notFound = append(notFound, filePath)
				continue
			}

			if !dryRun {
				if err := os.Remove(fullPath); err != nil {
					errors = append(errors, fmt.Sprintf("remove %s: %v", filePath, err))
					continue
				}
			}
			removed = append(removed, filePath)
		}
	}

	// Warn about shared packages.
	for _, pkgName := range sharedPkgs {
		warnings = append(warnings, fmt.Sprintf(
			"package %q is shared with another persona — files kept", pkgName))
	}

	// --- Step 5: Remove generated files (with checksum check) ---
	for _, gf := range persona.GeneratedFiles {
		fullPath := filepath.Join(target, gf.Path)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			notFound = append(notFound, gf.Path)
			continue
		}

		// Check if file was modified since install.
		if !installer.VerifyChecksum(fullPath, gf.Checksum) {
			warnings = append(warnings, fmt.Sprintf(
				"%s has been modified since installation", gf.Path))
		}

		if !dryRun {
			if err := os.Remove(fullPath); err != nil {
				errors = append(errors, fmt.Sprintf("remove %s: %v", gf.Path, err))
				continue
			}
		}
		removed = append(removed, gf.Path)
	}

	// --- Step 5b: Remove exclusive MCP servers ---
	// For each exclusive package being removed, check which of its
	// MCP servers are NOT shared with any other installed package.
	// Only remove those exclusive servers from the config file.
	// We pass ALL exclusive packages at once so that servers shared
	// only between packages being removed are still cleaned up.
	if assistantName != "" {
		translator, err := mcp.NewTranslator(assistantName)
		if err == nil {
			// Collect servers that are safe to remove (not shared
			// with any package outside the set being removed).
			serversToRemove := installer.ExclusiveMCPServers(manifest, exclusivePkgs)

			// Also check shared packages — warn about their servers being kept.
			for _, pkgName := range sharedPkgs {
				if pkg := installer.FindInstalled(manifest, pkgName); pkg != nil && len(pkg.MCPServers) > 0 {
					warnings = append(warnings, fmt.Sprintf(
						"MCP servers from shared package %q kept: %s",
						pkgName, strings.Join(pkg.MCPServers, ", ")))
				}
			}

			if len(serversToRemove) > 0 {
				mcpResult, err := mcp.Uninstall(target, translator, serversToRemove, dryRun)
				if err != nil {
					errors = append(errors, fmt.Sprintf("MCP uninstall: %v", err))
				} else if mcpResult != nil {
					for _, s := range mcpResult.Removed {
						removed = append(removed, fmt.Sprintf("MCP server: %s", s))
					}
				}
			}
		}
	}

	// --- Step 6: Update manifest ---
	if !dryRun {
		for _, pkgName := range exclusivePkgs {
			installer.RecordUninstall(manifest, pkgName)
		}
		installer.RecordPersonaUninstall(manifest, personaName)
		if err := installer.SaveManifest(target, manifest); err != nil {
			errors = append(errors, fmt.Sprintf("manifest: %v", err))
		}
	}

	// --- Step 7: Output ---
	return formatPersonaUninstallResult(cmd, mode, personaName, removed, notFound, warnings, errors, dryRun)
}

// findSharedPackages returns packages that are in allPkgs but NOT in exclusivePkgs.
func findSharedPackages(allPkgs, exclusivePkgs []string) []string {
	exclSet := make(map[string]bool)
	for _, p := range exclusivePkgs {
		exclSet[p] = true
	}

	var shared []string
	for _, p := range allPkgs {
		if !exclSet[p] {
			shared = append(shared, p)
		}
	}
	return shared
}

// formatPersonaUninstallResult renders the uninstall result.
func formatPersonaUninstallResult(
	cmd *cobra.Command,
	mode OutputMode,
	personaName string,
	removed, notFound, warnings, errs []string,
	dryRun bool,
) error {
	if mode == OutputJSON {
		output := struct {
			Persona  string   `json:"persona"`
			Removed  []string `json:"removed"`
			NotFound []string `json:"not_found"`
			Warnings []string `json:"warnings"`
			Errors   []string `json:"errors"`
			Summary  struct {
				Removed  int `json:"removed"`
				NotFound int `json:"not_found"`
				Errors   int `json:"errors"`
			} `json:"summary"`
		}{
			Persona:  personaName,
			Removed:  nonNil(removed),
			NotFound: nonNil(notFound),
			Warnings: nonNil(warnings),
			Errors:   nonNil(errs),
		}
		output.Summary.Removed = len(removed)
		output.Summary.NotFound = len(notFound)
		output.Summary.Errors = len(errs)
		return json.NewEncoder(cmd.OutOrStdout()).Encode(output)
	}

	if mode == OutputQuiet {
		for _, e := range errs {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "  error: %s\n", e)
		}
		if len(errs) > 0 {
			return fmt.Errorf("persona uninstall completed with %d errors", len(errs))
		}
		return nil
	}

	// Normal / Verbose
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	red := color.New(color.FgRed)
	bold := color.New(color.Bold)

	_, _ = bold.Fprintf(cmd.OutOrStdout(), "Uninstalling persona %q\n\n", personaName)

	for _, f := range removed {
		if dryRun {
			_, _ = yellow.Fprintf(cmd.OutOrStdout(), "  would remove: %s\n", f)
		} else {
			_, _ = green.Fprintf(cmd.OutOrStdout(), "  removed: %s\n", f)
		}
	}
	for _, f := range notFound {
		_, _ = color.New(color.Faint).Fprintf(cmd.OutOrStdout(), "  not found: %s\n", f)
	}
	for _, w := range warnings {
		_, _ = yellow.Fprintf(cmd.OutOrStdout(), "  warning: %s\n", w)
	}
	for _, e := range errs {
		_, _ = red.Fprintf(cmd.ErrOrStderr(), "  error: %s\n", e)
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout())
	_, _ = bold.Fprintf(cmd.OutOrStdout(), "%d removed, %d not found, %d errors\n",
		len(removed), len(notFound), len(errs))

	if len(errs) > 0 {
		return fmt.Errorf("persona uninstall completed with %d errors", len(errs))
	}
	return nil
}
