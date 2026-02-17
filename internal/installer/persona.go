package installer

import (
	"fmt"
	"io/fs"
	"path"
	"sort"

	"github.com/willvelida/code-minions/internal/assistant"
	"github.com/willvelida/code-minions/internal/mcp"
	"github.com/willvelida/code-minions/internal/model"
	"github.com/willvelida/code-minions/internal/registry"
)

// PersonaInstaller installs all packages in a resolved persona for a
// specific coding assistant. It:
//
//  1. Installs each package's files using the existing Installer
//  2. Generates an assistant-specific "grouping artefact" that tells
//     the assistant which packages belong to this persona
//
// Think of it as a foreman: it doesn't lay bricks (that's the
// Installer's job), but it coordinates the work and adds the
// finishing touches (the grouping artefact).
type PersonaInstaller struct {
	// Resolved is the persona with all its packages fully looked up.
	// This comes from the PersonaResolver we built in Phase 1.
	Resolved *registry.ResolvedPersona

	// Content is the embedded filesystem that contains the package
	// files, including each package's mcp.yaml. We need this to
	// load MCP configurations — DownloadPackage strips the prefix,
	// but mcp.LoadConfig needs the full "packages/<name>/mcp.yaml" path.
	Content fs.FS

	// AssistantName is the target assistant (e.g. "copilot", "claude").
	AssistantName string

	// Target is the directory where files will be installed
	// (usually "." — the current project directory).
	Target string

	// Force overwrites existing files when true.
	Force bool

	// DryRun previews changes without writing files when true.
	DryRun bool
}

// PersonaResult tracks what happened during a persona installation.
// It includes per-package results plus any generated grouping files.
type PersonaResult struct {
	// PackageResults maps package name → its install result.
	// Each package gets its own Copied/Skipped/Errors lists.
	PackageResults map[string]*Result

	// GeneratedFiles lists the grouping artefact files that were
	// created (e.g. AGENTS.md section, persona agent file).
	GeneratedFiles []string

	// MCPResult captures what happened during MCP server merging.
	// Nil when no packages in the persona have mcp.yaml files.
	MCPResult *mcp.InstallResult

	// MCPServersByPackage tracks which MCP servers came from which
	// package. This is recorded in the manifest so that uninstall
	// can use reference counting to decide which servers to remove.
	MCPServersByPackage map[string][]string

	// Errors collects persona-level errors (not package-level).
	Errors []string
}

// TotalCopied returns the total number of files copied across all packages.
func (r *PersonaResult) TotalCopied() int {
	total := 0
	for _, pr := range r.PackageResults {
		total += len(pr.Copied)
	}
	return total + len(r.GeneratedFiles)
}

// TotalSkipped returns the total number of files skipped across all packages.
func (r *PersonaResult) TotalSkipped() int {
	total := 0
	for _, pr := range r.PackageResults {
		total += len(pr.Skipped)
	}
	return total
}

// TotalErrors returns the total number of errors across all packages
// plus persona-level errors.
func (r *PersonaResult) TotalErrors() int {
	total := len(r.Errors)
	for _, pr := range r.PackageResults {
		total += len(pr.Errors)
	}
	return total
}

// Install executes the full persona installation:
//
//  1. Look up the assistant's path mapper (agents/ → .github/agents/ etc.)
//  2. For each package in the persona, run the existing Installer
//  3. Generate the assistant-specific grouping artefact
//
// Each step is explained inline with comments.
func (pi *PersonaInstaller) Install() (*PersonaResult, error) {
	result := &PersonaResult{
		PackageResults: make(map[string]*Result),
	}

	// --- Step 1: Get the assistant configuration ---
	// This tells us how to map paths for this assistant.
	// For example, Copilot maps agents/ → .github/agents/
	assistantCfg, err := assistant.Get(pi.AssistantName)
	if err != nil {
		return nil, fmt.Errorf("unknown assistant %q: %w", pi.AssistantName, err)
	}
	pathMapper := assistantCfg.NewPathMapper()

	// --- Step 2: Install each package ---
	// We loop through every package in the persona and use the
	// existing Installer to copy its files. This is the same
	// Installer that `code-minions install --package X` uses.
	for _, rp := range pi.Resolved.Packages {
		pkgName := rp.Package.Name

		// Get the package's filesystem (its files).
		// DownloadPackage returns an fs.FS rooted at the package dir
		// (e.g. just agents/, skills/ — no "packages/git-workflow" prefix).
		pkgFS, err := rp.Source.DownloadPackage(pkgName, rp.Package.Version)
		if err != nil {
			result.Errors = append(result.Errors,
				fmt.Sprintf("failed to download package %q: %v", pkgName, err))
			continue
		}

		// Figure out which top-level directories this package has.
		// Typically "agents" and "skills", but we discover them
		// dynamically in case packages have other directories.
		dirs, err := listTopDirs(pkgFS)
		if err != nil {
			result.Errors = append(result.Errors,
				fmt.Sprintf("failed to list directories in package %q: %v", pkgName, err))
			continue
		}

		// Create an Installer for this package.
		// Notice we DON'T set StripPrefix because DownloadPackage
		// already returns a filesystem rooted at the package dir.
		inst := &Installer{
			Content:    pkgFS,
			Target:     pi.Target,
			Force:      pi.Force,
			DryRun:     pi.DryRun,
			PathMapper: pathMapper,
		}

		pkgResult, err := inst.Install(dirs)
		if err != nil {
			result.Errors = append(result.Errors,
				fmt.Sprintf("failed to install package %q: %v", pkgName, err))
			continue
		}

		result.PackageResults[pkgName] = pkgResult
	}

	// --- Step 2b: Process MCP servers for all packages ---
	// Each package may have an mcp.yaml that declares MCP servers.
	// We collect them all, detect cross-package conflicts (same
	// server name, different config), and merge the non-conflicting
	// ones into the assistant's native config file.
	//
	// This uses the same mcp.Install → mcp.Merge pipeline that
	// per-package installs use, but we run it for each package
	// and the Merge function handles deduplication automatically:
	//   - Same name + same config → skip (deduplicate)
	//   - Same name + different config → conflict (unless force)
	//   - New server → add
	if pi.Content != nil {
		mcpResult, serversByPkg := pi.installMCP(result)
		if mcpResult != nil {
			result.MCPResult = mcpResult
		}
		if len(serversByPkg) > 0 {
			result.MCPServersByPackage = serversByPkg
		}
	}

	// --- Step 3: Generate the grouping artefact ---
	// This creates the assistant-specific file that tells the
	// assistant about the persona. Each assistant needs a different
	// file format.
	generator, err := NewGroupingGenerator(assistantCfg, pi.Resolved, pi.Target, pi.DryRun, pi.Force)
	if err != nil {
		result.Errors = append(result.Errors,
			fmt.Sprintf("failed to create grouping generator: %v", err))
		return result, nil
	}

	// Pass MCP server names to the generator so it can include
	// them in the persona's agent file (e.g. as tools: or mcpServers:
	// in the frontmatter). Each generator formats this in its own
	// assistant-specific syntax.
	if result.MCPServersByPackage != nil {
		allServers := collectUniqueServers(result.MCPServersByPackage)
		generator.SetMCPServers(allServers)
	}

	generatedFiles, err := generator.Generate()
	if err != nil {
		result.Errors = append(result.Errors,
			fmt.Sprintf("failed to generate grouping artefact: %v", err))
		return result, nil
	}

	result.GeneratedFiles = generatedFiles

	// --- Step 4: Record installation in the manifest ---
	// This is the "receipt" step. We write down what we installed so
	// that future uninstall/update commands know exactly what to do.
	// We skip this step for dry-runs (no real files were written).
	if !pi.DryRun {
		if err := pi.recordManifest(result); err != nil {
			result.Errors = append(result.Errors,
				fmt.Sprintf("failed to record manifest: %v", err))
		}
	}

	return result, nil
}

// listTopDirs returns the names of top-level directories in an fs.FS.
// For a typical package this returns ["agents", "skills"].
// It skips files and metadata directories.
func listTopDirs(content fs.FS) ([]string, error) {
	entries, err := fs.ReadDir(content, ".")
	if err != nil {
		return nil, err
	}

	var dirs []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dirs = append(dirs, entry.Name())
	}

	return dirs, nil
}

// recordManifest writes the persona installation "receipt" to the
// manifest file (.code-minions/installed.json). This involves:
//
//  1. Loading the existing manifest (or creating a new one)
//  2. Recording each package with its files
//  3. Computing checksums for generated files
//  4. Recording the persona entry linking everything together
//  5. Saving the manifest back to disk
//
// If any of these steps fail, the error is returned but the
// installation itself is still considered successful (the files
// are already on disk). The manifest is "best effort" tracking.
func (pi *PersonaInstaller) recordManifest(result *PersonaResult) error {
	// Step 4a: Load (or create) the manifest.
	manifest, err := LoadManifest(pi.Target)
	if err != nil {
		return fmt.Errorf("load manifest: %w", err)
	}

	// Step 4b: Record each package that was successfully installed.
	var pkgNames []string
	for pkgName, pkgResult := range result.PackageResults {
		// Only record packages that actually installed files.
		if len(pkgResult.Copied) == 0 {
			continue
		}

		// Look up the package version from the resolved persona.
		var version string
		for _, rp := range pi.Resolved.Packages {
			if rp.Package.Name == pkgName {
				version = rp.Package.Version
				break
			}
		}

		if version == "" {
			return fmt.Errorf("record manifest: resolved package not found for %q", pkgName)
		}

		RecordInstall(manifest, pkgName, version, "embedded", pi.AssistantName, pkgResult.Copied)

		// Step 4b-ii: Stamp MCP server names onto the package entry.
		// This is how we track which servers each package "owns" —
		// the reference counting that makes safe uninstall possible.
		if servers, ok := result.MCPServersByPackage[pkgName]; ok && len(servers) > 0 {
			if entry := FindInstalled(manifest, pkgName); entry != nil {
				entry.MCPServers = servers
			}
		}

		pkgNames = append(pkgNames, pkgName)
	}

	// Step 4c: Compute checksums for generated files.
	// These are the grouping artefacts (AGENTS.md, persona agent, etc.).
	var trackedGenFiles []model.InstalledFile
	for _, genPath := range result.GeneratedFiles {
		tracked, err := NewInstalledFile(pi.Target, genPath, "generated")
		if err != nil {
			// Non-fatal: we still record the persona, just without
			// this file's checksum. Surface the warning so it's visible.
			result.Errors = append(result.Errors,
				fmt.Sprintf("checksum failed for %s: %v", genPath, err))
			continue
		}
		trackedGenFiles = append(trackedGenFiles, tracked)
	}

	// Step 4d: Record the persona entry.
	// Note: personas don't have their own version field — the version
	// is tracked per-package instead. We use an empty string here.
	RecordPersonaInstall(
		manifest,
		pi.Resolved.Persona.Name,
		"", // personas don't have a standalone version
		"embedded",
		pi.AssistantName,
		pkgNames,
		trackedGenFiles,
	)

	// Step 4e: Save the manifest.
	return SaveManifest(pi.Target, manifest)
}

// installMCP processes MCP servers for all packages in the persona.
// It collects each package's mcp.yaml, translates them to the target
// assistant's format, and merges them into the assistant's config file.
//
// Returns:
//   - mcpResult: the combined install result (nil if no packages have MCP)
//   - serversByPkg: map of package name → server names that were installed
//
// Why this is a separate method:
// MCP processing touches a shared file (e.g. .vscode/mcp.json), so we
// process all packages together rather than one at a time. This lets
// the Merge function handle deduplication across packages — if two
// packages declare the same "github" server with the same config,
// it's added once and both packages record ownership.
func (pi *PersonaInstaller) installMCP(result *PersonaResult) (*mcp.InstallResult, map[string][]string) {
	translator, err := mcp.NewTranslator(pi.AssistantName)
	if err != nil {
		// Not all assistants support MCP (e.g. Roo Code). That's fine.
		return nil, nil
	}

	// Collect MCP configs from all packages, tracking which server
	// names came from which package.
	//
	// serverOwnership maps server name → list of package names that
	// declared it. This is how we build the MCPServersByPackage map.
	serverOwnership := make(map[string][]string) // server name → []package name
	serversByPkg := make(map[string][]string)    // package name → []server name

	// allServers accumulates the translated server configs from all
	// packages. When two packages define the same server, the Merge
	// function will deduplicate (identical config) or conflict
	// (different config).
	allServers := make(map[string]any)

	for _, rp := range pi.Resolved.Packages {
		pkgName := rp.Package.Name
		pkgDir := path.Join("packages", pkgName)

		// Load this package's mcp.yaml (returns nil if none exists).
		cfg, err := mcp.LoadConfig(pi.Content, pkgDir)
		if err != nil {
			result.Errors = append(result.Errors,
				fmt.Sprintf("MCP (%s): failed to load config: %v", pkgName, err))
			continue
		}
		if cfg == nil {
			// This package has no MCP servers — that's normal.
			continue
		}

		// Translate to the assistant's native format.
		servers, warnings, err := translator.Translate(cfg)
		if err != nil {
			result.Errors = append(result.Errors,
				fmt.Sprintf("MCP (%s): translation failed: %v", pkgName, err))
			continue
		}
		if len(warnings) > 0 {
			for _, w := range warnings {
				result.Errors = append(result.Errors,
					fmt.Sprintf("MCP (%s) warning: %s", pkgName, w))
			}
		}

		// Track ownership: which package declared which server.
		for name, serverCfg := range servers {
			// Detect cross-package server conflicts: if another package
			// already declared a server with the same name, warn about
			// the overwrite. The downstream Merge function handles
			// deduplication for identical configs, but at this point
			// we're building a flat map, so last-write-wins.
			if existing, exists := allServers[name]; exists {
				// Only warn if configs actually differ.
				if !mcp.ServersEqual(existing, serverCfg) {
					owners := serverOwnership[name]
					result.Errors = append(result.Errors,
						fmt.Sprintf("MCP conflict: server %q declared by %v and %s with different configs — using %s's version",
							name, owners, pkgName, pkgName))
				}
			}
			serverOwnership[name] = append(serverOwnership[name], pkgName)
			serversByPkg[pkgName] = append(serversByPkg[pkgName], name)
			allServers[name] = serverCfg
		}
	}

	if len(allServers) == 0 {
		return nil, nil
	}

	// Sort server lists per package for deterministic manifest output.
	for pkg := range serversByPkg {
		sort.Strings(serversByPkg[pkg])
	}

	// Merge all collected servers into the assistant's config file.
	// This is the same mcp.Install flow, but we call Merge directly
	// with the pre-collected servers rather than loading from a single
	// package's mcp.yaml.
	mcpResult, err := mcp.InstallServers(
		pi.Target, translator, allServers, pi.Force, pi.DryRun,
	)
	if err != nil {
		result.Errors = append(result.Errors,
			fmt.Sprintf("MCP merge failed: %v", err))
		return nil, serversByPkg
	}

	// Update serversByPkg to only include servers that were actually
	// added (not skipped or conflicted). This ensures the manifest
	// accurately reflects what's in the config file.
	addedSet := make(map[string]bool, len(mcpResult.Merge.Added))
	for _, name := range mcpResult.Merge.Added {
		addedSet[name] = true
	}
	// Also count skipped servers as "owned" — they were already in
	// the config from a previous install, but this package still
	// claims them for reference counting purposes.
	for _, name := range mcpResult.Merge.Skipped {
		addedSet[name] = true
	}
	// Note: conflicted servers are NOT added to addedSet because
	// Merge did not install them (conflicts with --force go into
	// Added instead). Packages that declared conflicted servers
	// will not track them — they weren't written to the config.
	for pkg, servers := range serversByPkg {
		var kept []string
		for _, s := range servers {
			if addedSet[s] {
				kept = append(kept, s)
			}
		}
		if len(kept) > 0 {
			serversByPkg[pkg] = kept
		} else {
			delete(serversByPkg, pkg)
		}
	}

	return mcpResult, serversByPkg
}

// collectUniqueServers extracts a sorted, deduplicated list of server
// names from the MCPServersByPackage map. Multiple packages may
// declare the same server — this gives us one list of all unique
// server names for the persona.
func collectUniqueServers(serversByPkg map[string][]string) []string {
	seen := make(map[string]bool)
	for _, servers := range serversByPkg {
		for _, s := range servers {
			seen[s] = true
		}
	}
	result := make([]string, 0, len(seen))
	for s := range seen {
		result = append(result, s)
	}
	sort.Strings(result)
	return result
}
