package cmd

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/willvelida/code-minions/internal/assistant"
	"github.com/willvelida/code-minions/internal/installer"
	"github.com/willvelida/code-minions/internal/registry"
)

func newUpdateCommand(content fs.FS) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update installed packages with the latest versions",
		Long: `Update overwrites previously installed files with the latest embedded
content. When run with no flags, only packages that are already
installed in the target directory are updated.

AGENTS.md is not modified during updates.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			packageFlag, _ := cmd.Flags().GetString("package")
			forFlag, _ := cmd.Flags().GetString("for")
			personaFlag, _ := cmd.Flags().GetString("persona")
			mode := getOutputMode(cmd)

			// --- Persona update branch ---
			// Update means: re-resolve and re-install with force=true.
			// Note: resolveForFlag is intentionally NOT called here. Persona
			// installs target a single, explicitly-named assistant, so --for
			// is required and must not be auto-resolved from the manifest.
			if personaFlag != "" {
				if packageFlag != "" {
					return fmt.Errorf("--persona and --package cannot be used together")
				}
				if forFlag == "" {
					return fmt.Errorf("--for is required when updating a persona\n\n" +
						"Usage: code-minions update --persona <name> --for <assistant>")
				}
				// Re-install with force=true is the update.
				return runPersonaInstall(cmd, content, personaFlag, forFlag, target, true, dryRun)
			}

			// If --for is set, look up the assistant config and build a path mapper
			forFlag, err := resolveForFlag(cmd, forFlag, target, mode)
			if err != nil {
				return err
			}

			var pathMapper func(string) string
			if forFlag != "" {
				cfg, err := assistant.Get(forFlag)
				if err != nil {
					return err
				}
				pathMapper = cfg.NewPathMapper()
			}

			// When --package is provided, use buildPackageList (same as install).
			// When no flags are provided, detect what's already installed.
			var packageDirs []string
			if packageFlag != "" {
				var err error
				packageDirs, err = buildPackageList(content, packageFlag, target)
				if err != nil {
					return err
				}
			} else {
				var err error
				packageDirs, err = detectInstalled(content, target, pathMapper)
				if err != nil {
					return err
				}
				if len(packageDirs) == 0 {
					if mode == OutputJSON {
						return json.NewEncoder(cmd.OutOrStdout()).Encode(struct {
							Updated []string `json:"updated"`
							Errors  []string `json:"errors"`
							Summary struct {
								Updated int `json:"updated"`
								Errors  int `json:"errors"`
							} `json:"summary"`
						}{Updated: []string{}, Errors: []string{}})
					}
					if mode != OutputQuiet {
						fmt.Println("No installed packages found. Use 'code-minions install' to install packages first.")
					}
					return nil
				}
			}

			if dryRun && (mode == OutputNormal || mode == OutputVerbose) {
				_, _ = color.New(color.FgYellow, color.Bold).Println("Dry run - no files will be written")
				fmt.Println()
			}

			// Update each package (with prefix stripping, force always true)
			combinedResult := &installer.Result{}
			for _, pkgDir := range packageDirs {
				if err := runUpdate(content, target, pkgDir, dryRun, pathMapper, combinedResult); err != nil {
					return err
				}
			}

			// No AGENTS.md handling — update leaves it untouched

			// Update manifest with new versions (skip for dry-run)
			if !dryRun && len(combinedResult.Copied) > 0 {
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
					}
					if err := installer.SaveManifest(target, manifest); err != nil {
						combinedResult.Errors = append(combinedResult.Errors, fmt.Sprintf("manifest: %v", err))
					}
				}
			}

			// JSON output
			if mode == OutputJSON {
				updated := combinedResult.Copied
				if updated == nil {
					updated = []string{}
				}
				errs := combinedResult.Errors
				if errs == nil {
					errs = []string{}
				}
				result := struct {
					Updated []string `json:"updated"`
					Errors  []string `json:"errors"`
					Summary struct {
						Updated int `json:"updated"`
						Errors  int `json:"errors"`
					} `json:"summary"`
				}{
					Updated: updated,
					Errors:  errs,
				}
				result.Summary.Updated = len(combinedResult.Copied)
				result.Summary.Errors = len(combinedResult.Errors)
				if err := json.NewEncoder(cmd.OutOrStdout()).Encode(result); err != nil {
					return err
				}
				if len(combinedResult.Errors) > 0 {
					return fmt.Errorf("update completed with %d errors", len(combinedResult.Errors))
				}
				return nil
			}

			// Quiet mode — only report errors to stderr
			if mode == OutputQuiet {
				for _, e := range combinedResult.Errors {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "  error: %s\n", e)
				}
				if len(combinedResult.Errors) > 0 {
					return fmt.Errorf("update completed with %d errors", len(combinedResult.Errors))
				}
				return nil
			}

			green := color.New(color.FgGreen)
			yellow := color.New(color.FgYellow)
			red := color.New(color.FgRed)
			bold := color.New(color.Bold)

			// Verbose: show package list and detection method
			if packageFlag != "" {
				verbosePrintf(cmd, mode, "packages (explicit): %v\n", packageDirs)
			} else {
				verbosePrintf(cmd, mode, "packages (auto-detected): %v\n", packageDirs)
			}

			// Print results — say "updated" instead of "copied"
			for _, f := range combinedResult.Copied {
				if dryRun {
					_, _ = yellow.Printf("  would update: %s\n", f)
				} else {
					_, _ = green.Printf("  updated: %s\n", f)
				}
			}
			for _, e := range combinedResult.Errors {
				_, _ = red.Fprintf(os.Stderr, "  error: %s\n", e)
			}

			// Summary
			fmt.Println()
			if dryRun {
				_, _ = bold.Printf("%d would be updated, %d errors\n",
					len(combinedResult.Copied), len(combinedResult.Errors))
			} else {
				_, _ = bold.Printf("%d updated, %d errors\n",
					len(combinedResult.Copied), len(combinedResult.Errors))
			}

			if len(combinedResult.Errors) > 0 {
				return fmt.Errorf("update completed with %d errors", len(combinedResult.Errors))
			}

			return nil
		},
	}

	cmd.Flags().String("target", ".", "Target directory for update")
	cmd.Flags().String("package", "", "Comma-separated list of packages to update (omit to auto-detect installed packages)")
	cmd.Flags().String("persona", "", "Update a persona (re-install all its packages with latest content)")
	cmd.Flags().String("for", "", "Target coding assistant ("+assistant.FlagUsage()+")")
	cmd.Flags().Bool("dry-run", false, "Show what would be updated without writing files")
	cmd.Flags().Bool("update-assistant", false, "Update the manifest's assistant field when --for differs")

	return cmd
}

// runUpdate runs the installer for a set of directories, appending
// results into combined. When stripPrefix is non-empty it is set on
// the installer (used for packages). When dirs are provided they
// override the default of []string{stripPrefix}.
func runUpdate(content fs.FS, target, stripPrefix string, dryRun bool, pathMapper func(string) string, combined *installer.Result, dirs ...string) error {
	if len(dirs) == 0 {
		dirs = []string{stripPrefix}
	}
	inst := &installer.Installer{
		Content:     content,
		Target:      target,
		Force:       true,
		DryRun:      dryRun,
		StripPrefix: stripPrefix,
		PathMapper:  pathMapper,
	}
	result, err := inst.Install(dirs)
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}
	combined.Copied = append(combined.Copied, result.Copied...)
	combined.Skipped = append(combined.Skipped, result.Skipped...)
	combined.Errors = append(combined.Errors, result.Errors...)
	return nil
}

// detectInstalled scans the target directory to find which packages
// are already installed. It scans the installed skills/ directory once
// and intersects with the embedded package set.
func detectInstalled(content fs.FS, target string, pathMapper func(string) string) ([]string, error) {
	var dirs []string

	// Build a set of available (embedded) package names for quick lookup.
	available, err := listSubDirs(content, "packages")
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}
	embeddedPkgs := make(map[string]struct{}, len(available))
	for _, pkg := range available {
		embeddedPkgs[pkg] = struct{}{}
	}

	// Scan the installed skills directories once and intersect with
	// the embedded package set.
	skillsBase := "skills"
	if pathMapper != nil {
		skillsBase = pathMapper(skillsBase)
	}
	skillsFullPath := filepath.Join(target, skillsBase)

	skillEntries, err := os.ReadDir(skillsFullPath)
	if err != nil {
		// If skills/ doesn't exist, that's fine — just no packages installed.
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to list installed skills in %s: %w", skillsFullPath, err)
		}
	} else {
		for _, entry := range skillEntries {
			if !entry.IsDir() {
				continue
			}
			name := entry.Name()
			if _, ok := embeddedPkgs[name]; ok {
				dirs = append(dirs, "packages/"+name)
			}
		}
	}

	return dirs, nil
}
