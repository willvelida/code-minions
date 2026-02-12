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

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--package` | string | all | Comma-separated list of packages to install |
| `--standards` | string | all | Comma-separated list of language standards |
| `--for` | string | | Target coding assistant (`copilot`, `claude`, `opencode`) |
| `--target` | string | `.` | Target directory for installation |
| `--dry-run` | bool | false | Show what would be installed without writing files |
| `--force` | bool | false | Overwrite existing files |

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
