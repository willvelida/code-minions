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

A collection of reusable skills and standards for AI coding agents to enhance AI-assisted development.

## Overview

code-minions provides structured knowledge that AI agents can load to perform development tasks consistently and effectively. It includes:

- **Packages** — Bundled agents and skills for common workflows (git operations, documentation, DevContainers)
- **Standards** — Language-specific guidelines for consistent code quality

## CLI

code-minions includes a CLI tool that installs packages (made up of Agents and Agent Skills) and standards into your own repositories.

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

#### Windows (via Winget)

> **Note:** Winget availability depends on the package being approved in the [winget-pkgs](https://github.com/microsoft/winget-pkgs) repository. If the command below doesn't find the package yet, use the PowerShell install script above.

```powershell
winget install willvelida.code-minions
```

### Usage

```bash
# Install everything (all packages + all standards)
code-minions install

# Install a specific package (agent + skill)
code-minions install --package threat-modelling

# Install multiple packages
code-minions install --package threat-modelling,git-workflow

# Install only language standards
code-minions install --standards python,typescript

# Install a package and standards together
code-minions install --package git-workflow --standards bash

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
| `--standards` | string | all | Comma-separated list of language standards |
| `--for` | string | | Target coding assistant (`copilot`, `claude`, `opencode`) |
| `--target` | string | `.` | Target directory for installation |
| `--dry-run` | bool | false | Show what would be installed without writing files |
| `--force` | bool | false | Overwrite existing files |

When installing packages, an `AGENTS.md` file is automatically created if one does not already exist. Existing `AGENTS.md` files are never overwritten.

### Uninstalling

```bash
# Uninstall a specific package
code-minions uninstall --package threat-modelling

# Uninstall language standards
code-minions uninstall --standards python

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
| `--standards` | string | all | Comma-separated list of language standards |
| `--for` | string | | Target coding assistant (`copilot`, `claude`, `opencode`) — **required** when uninstalling everything |
| `--target` | string | `.` | Target directory to uninstall from |
| `--dry-run` | bool | false | Show what would be removed without deleting files |

### Updating

```bash
# Update all installed packages and standards
code-minions update

# Update a specific package
code-minions update --package threat-modelling

# Update multiple packages
code-minions update --package threat-modelling,git-workflow

# Update only language standards
code-minions update --standards python,typescript

# Update for a specific coding assistant
code-minions update --package developer-mentor --for copilot

# Preview what would be updated
code-minions update --package threat-modelling --dry-run
```

Update overwrites installed files with the latest embedded content. When run with no flags, only packages and standards already present in the target directory are updated — no new packages are installed. `AGENTS.md` is not modified during updates.

### Update flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--package` | string | | Comma-separated list of packages to update (omit to auto-detect installed) |
| `--standards` | string | | Comma-separated list of language standards (omit to auto-detect installed) |
| `--for` | string | | Target coding assistant (`copilot`, `claude`, `opencode`) |
| `--target` | string | `.` | Target directory |
| `--dry-run` | bool | false | Show what would be updated without writing files |

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

Without `--for`, files are installed to generic locations (`agents/`, `skills/`, `standards/`). Standards are always placed in `standards/` regardless of the assistant.

## Standards

Language standards provide consistent guidelines for development practices.

```text
standards/
└── languages/
    ├── bash/                   # Portable shell scripting with ShellCheck and shfmt
    ├── python/                 # Modern Python with uv, Ruff, and strict typing
    └── typescript/             # Type-safe JavaScript with ESLint, Prettier, and pnpm
```

Each language includes standards for:
- **Core** — Project structure, naming, coding style
- **Tooling** — Package manager, linter, formatter, testing
- **Security** — Secure practices and prohibited patterns
- **Development Environment** — DevContainer and VS Code setup

See the [standards index](standards/languages/standards.index.md) for details. Additional languages will be added over time.

## Contributing

Contributions are welcome. Please open an issue to discuss changes before submitting a pull request.

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.
