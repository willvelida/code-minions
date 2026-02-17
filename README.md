# code-minions

```
  ██████╗ ██████╗ ██████╗ ███████╗
 ██╔════╝██╔═══██╗██╔══██╗██╔════╝
 ██║     ██║   ██║██║  ██║█████╗
 ██║     ██║   ██║██║  ██║██╔══╝
 ╚██████╗╚██████╔╝██████╔╝███████╗
  ╚═════╝ ╚═════╝ ╚═════╝ ╚══════╝
 ███╗   ███╗██╗███╗   ██╗██╗ ██████╗ ███╗   ██╗███████╗
 ████╗ ████║██║████╗  ██║██║██╔═══██╗████╗  ██║██╔════╝
 ██╔████╔██║██║██╔██╗ ██║██║██║   ██║██╔██╗ ██║███████╗
 ██║╚██╔╝██║██║██║╚██╗██║██║██║   ██║██║╚██╗██║╚════██║
 ██║ ╚═╝ ██║██║██║ ╚████║██║╚██████╔╝██║ ╚████║███████║
 ╚═╝     ╚═╝╚═╝╚═╝  ╚═══╝╚═╝ ╚═════╝ ╚═╝  ╚═══╝╚══════╝
```

[![Go Report Card](https://goreportcard.com/badge/github.com/willvelida/code-minions)](https://goreportcard.com/report/github.com/willvelida/code-minions)
[![codecov](https://codecov.io/gh/willvelida/code-minions/branch/main/graph/badge.svg)](https://codecov.io/gh/willvelida/code-minions)

A collection of reusable assets for AI coding agents to enhance AI-assisted development.

## Overview

code-minions provides structured knowledge that AI agents can load to perform development tasks consistently and effectively. It includes:

- **Packages** — Bundled agents and skills for common workflows (git operations, documentation, DevContainers)

## CLI

code-minions includes a CLI tool that installs packages (made up of Agents and Agent Skills) into your own repositories.

### Installation

#### From source (requires Go 1.25+)

```bash
go install github.com/willvelida/code-minions/cmd/code-minions@latest
```

#### From GitHub Releases

Download a pre-built binary from the [Releases page](https://github.com/willvelida/code-minions/releases).

#### Linux and macOS (via install script)

```bash
curl -fsSL https://raw.githubusercontent.com/willvelida/code-minions/main/install.sh | bash
```

#### Windows (via PowerShell)

```powershell
powershell -Command "iwr -useb https://raw.githubusercontent.com/willvelida/code-minions/main/install.ps1 | iex"
```

> **Security note:** The command above downloads and executes a remote script. To inspect it first:
>
> ```powershell
> Invoke-WebRequest -Uri "https://raw.githubusercontent.com/willvelida/code-minions/main/install.ps1" -OutFile "install.ps1"
> # Review install.ps1, then run:
> .\install.ps1
> ```

#### Windows (via Winget)

> **Note:** Winget availability depends on the package being approved in the [winget-pkgs](https://github.com/microsoft/winget-pkgs) repository. If the command below doesn't find the package yet, use the PowerShell install script above.

```powershell
winget install willvelida.code-minions
```

### Usage

```bash
# Install everything (all packages)
code-minions install

# Install a specific package (agent + skill)
code-minions install --package threat-modelling

# Install multiple packages
code-minions install --package threat-modelling,git-workflow

# Preview what would be installed without writing files
code-minions install --dry-run

# Overwrite existing files
code-minions install --force

# Install for a specific coding assistant
code-minions install --package developer-mentor --for copilot
code-minions install --package developer-mentor --for claude
code-minions install --package developer-mentor --for opencode

# List available packages, standards, and assistants
code-minions list

# Print version
code-minions version

# Output as JSON (works with list, install, uninstall, update, version)
code-minions list --json
code-minions install --package git-workflow --json

# Verbose output for debugging
code-minions install --package git-workflow --verbose

# Quiet mode for CI pipelines (exit code only)
code-minions install --package git-workflow --quiet
```

### Install flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--package` | string | all | Comma-separated list of packages to install |
| `--for` | string | | Target coding assistant (`copilot`, `claude`, `opencode`) |
| `--target` | string | `.` | Target directory for installation |
| `--dry-run` | bool | false | Show what would be installed without writing files |
| `--force` | bool | false | Overwrite existing files |
| `--json` | bool | false | Output results as JSON |
| `--verbose` / `-v` | bool | false | Show detailed output |
| `--quiet` / `-q` | bool | false | Suppress all output except errors |

When installing packages, an `AGENTS.md` file is automatically created if one does not already exist. Existing `AGENTS.md` files are never overwritten.

### Uninstalling

```bash
# Uninstall a specific package (prompts for confirmation)
code-minions uninstall --package threat-modelling

# Uninstall everything (--for is required to identify file locations)
code-minions uninstall --for copilot

# Skip the confirmation prompt (for CI/scripting)
code-minions uninstall --for copilot --yes

# Preview what would be removed (no prompt)
code-minions uninstall --package git-workflow --dry-run
```

The uninstall command prompts for confirmation before removing files. You will also be prompted separately before `AGENTS.md` is removed. In non-interactive environments (no TTY), the command aborts unless `--yes` is passed.

### Uninstall flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--package` | string | all | Comma-separated list of packages to uninstall |
| `--for` | string | | Target coding assistant (`copilot`, `claude`, `opencode`) — **required** when uninstalling everything |
| `--target` | string | `.` | Target directory to uninstall from |
| `--dry-run` | bool | false | Show what would be removed without deleting files |
| `--yes` / `-y` | bool | false | Skip confirmation prompt and proceed with removal |
| `--json` | bool | false | Output results as JSON |
| `--verbose` / `-v` | bool | false | Show detailed output |
| `--quiet` / `-q` | bool | false | Suppress all output except errors |

```bash
# Update all installed packages
code-minions update

# Update a specific package
code-minions update --package threat-modelling

# Update multiple packages
code-minions update --package threat-modelling,git-workflow

# Update for a specific coding assistant
code-minions update --package developer-mentor --for copilot

# Preview what would be updated
code-minions update --package threat-modelling --dry-run
```

Update overwrites installed files with the latest embedded content. When run with no flags, only packages already present in the target directory are updated — no new packages are installed. `AGENTS.md` is not modified during updates.

### Update flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--package` | string | | Comma-separated list of packages to update (omit to auto-detect installed) |
| `--for` | string | | Target coding assistant (`copilot`, `claude`, `opencode`) |
| `--target` | string | `.` | Target directory |
| `--dry-run` | bool | false | Show what would be updated without writing files |
| `--json` | bool | false | Output results as JSON |
| `--verbose` / `-v` | bool | false | Show detailed output |
| `--quiet` / `-q` | bool | false | Suppress all output except errors |

### Transferring

Migrate agent and skill files from one coding assistant's directory layout to another:

```bash
# Transfer from Copilot to Claude
code-minions transfer --from copilot --to claude

# Preview what would be transferred
code-minions transfer --from copilot --to claude --dry-run

# Overwrite existing files at the destination
code-minions transfer --from copilot --to claude --force

# Transfer and remove the old Copilot layout
code-minions transfer --from copilot --to claude --cleanup

# Transfer in a specific directory
code-minions transfer --from claude --to opencode --target ./my-project
```

Files are copied by default — the source layout is left in place. Use `--cleanup` to delete the source files after a successful copy.

MCP server configurations are automatically translated between the source and target formats when present. An `AGENTS.md` routing file is always regenerated (not copied) in the target layout because it contains assistant-specific paths.

### Transfer flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--from` | string | | Source coding assistant (`copilot`, `claude`, `opencode`) — **required** |
| `--to` | string | | Target coding assistant (`copilot`, `claude`, `opencode`) — **required** |
| `--target` | string | `.` | Working directory |
| `--dry-run` | bool | false | Preview changes without writing files |
| `--force` | bool | false | Overwrite existing files at the destination |
| `--cleanup` | bool | false | Delete source agent/skill files after successful copy |
| `--json` | bool | false | Output results as JSON |
| `--verbose` / `-v` | bool | false | Show detailed output |
| `--quiet` / `-q` | bool | false | Suppress all output except errors |

### JSON Output

The `version`, `list`, `install`, `uninstall`, `update`, and `transfer` commands support a `--json` flag for machine-readable output:

```bash
# List packages as JSON
code-minions list --json
# {"packages": [{"name": "git-workflow", ...}], "assistants": [...]}

# Install and capture results
code-minions install --package git-workflow --json
# {"copied": [...], "skipped": [...], "errors": [], "summary": {"copied": 2, ...}}

# Transfer and capture results
code-minions transfer --from copilot --to claude --json
# {"from": "copilot", "to": "claude", "files": {"copied": [...], ...}}

# Pipe to jq
code-minions list --json | jq '.packages[].name'
```

### Verbosity

The `--verbose` (`-v`), `--quiet` (`-q`), and `--json` flags are mutually exclusive — only one can be used at a time.

| Flag | Behaviour |
|------|----------|
| (none) | Normal human-readable output |
| `--verbose` / `-v` | Normal output plus extra details (package lists, skip reasons, hints) |
| `--quiet` / `-q` | Suppress all stdout; errors still go to stderr. Exit code signals success/failure |
| `--json` | Machine-readable JSON to stdout |

`--quiet` is designed for CI pipelines where only the exit code matters:

```bash
code-minions install --package git-workflow --for copilot --quiet
```

> **Note:** `--quiet` has no effect on `list` and `version` (these commands exist solely to display information). A warning is printed to stderr.

### Shell Completion

Generate shell completion scripts for tab-completion of commands and flags:

```bash
# Bash
source <(code-minions completion bash)

# Zsh
source <(code-minions completion zsh)

# Fish
code-minions completion fish | source

# PowerShell
code-minions completion powershell | Out-String | Invoke-Expression
```

To load completions automatically in every session, add the appropriate command to your shell profile. Run `code-minions completion --help` for per-shell instructions.

## Packages

Each package bundles an agent and its corresponding skill into a single installable unit.

```text
packages/
├── creating-agent-skill/       # Create and review Agent Skills following the open specification
├── creating-agents/            # Create and review agent definition files
├── creating-devcontainers/     # Create and review DevContainer configurations
├── creating-documentation/     # Create and review README files and repository documentation
├── developer-mentor/           # Guide users through development concepts without writing code
├── git-workflow/               # Git branching, commits, merges, and repository management
├── raise-pull-requests/        # PR preparation, submission, and code review
└── threat-modelling/           # STRIDE-based threat modelling for repositories and cloud infrastructure
```

## Coding Assistants

Use the `--for` flag to install packages into the directories expected by your coding assistant:

| Assistant | `--for` value | Agents placed in | Skills placed in |
|-----------|--------------|------------------|------------------|
| GitHub Copilot | `copilot` | `.github/agents/` | `skills/` |
| Claude Code | `claude` | `.claude/agents/` | `.claude/skills/` |
| OpenCode | `opencode` | `.opencode/agents/` | `.opencode/skills/` |

Without `--for`, files are installed to generic locations (`agents/`, `skills/`).

## Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, testing expectations, and PR guidelines.

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for a full release history.

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.
