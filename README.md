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
```

### Install flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--package` | string | all | Comma-separated list of packages to install |
| `--for` | string | | Target coding assistant (`copilot`, `claude`, `opencode`) |
| `--target` | string | `.` | Target directory for installation |
| `--dry-run` | bool | false | Show what would be installed without writing files |
| `--force` | bool | false | Overwrite existing files |

When installing packages, an `AGENTS.md` file is automatically created if one does not already exist. Existing `AGENTS.md` files are never overwritten.

### Uninstalling

```bash
# Uninstall a specific package
code-minions uninstall --package threat-modelling

# Uninstall everything (--for is required to identify file locations)
code-minions uninstall --for copilot

# Preview what would be removed
code-minions uninstall --package git-workflow --dry-run
```

When uninstalling packages, you will be prompted before `AGENTS.md` is removed. If you decline (or stdin is empty, e.g. in CI), the file is kept.

### Uninstall flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--package` | string | all | Comma-separated list of packages to uninstall |
| `--for` | string | | Target coding assistant (`copilot`, `claude`, `opencode`) — **required** when uninstalling everything |
| `--target` | string | `.` | Target directory to uninstall from |
| `--dry-run` | bool | false | Show what would be removed without deleting files |

### Updating

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
