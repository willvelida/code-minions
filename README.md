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

A collection of reusable skills and standards for AI coding agents to enhance developer workflows.

## Overview

code-minions provides structured knowledge that AI agents can load to perform development tasks consistently and effectively. It includes:

- **Skills** — Step-by-step procedures for common workflows (git operations, documentation, DevContainers)
- **Standards** — Language-specific guidelines for consistent code quality

## CLI

code-minions includes a CLI tool that installs agents, skills, and standards into your own repositories.

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
# Install everything into the current directory
code-minions install

# Install only specific skills
code-minions install --skills git-workflow,threat-modelling

# Install only agents
code-minions install --agents

# Install only language standards
code-minions install --standards python,typescript

# Preview what would be installed without writing files
code-minions install --dry-run

# Overwrite existing files
code-minions install --force

# List available agents, skills, and standards
code-minions list

# Print version
code-minions version
```

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--skills` | string | all | Comma-separated list of skills to install |
| `--standards` | string | all | Comma-separated list of language standards |
| `--agents` | bool | false | Include agents in the installation |
| `--target` | string | `.` | Target directory for installation |
| `--dry-run` | bool | false | Show what would be installed without writing files |
| `--force` | bool | false | Overwrite existing files |

## Skills

```text
skills/
├── creating-agent-skill/        # Create and review Agent Skills following the open specification
├── creating-devcontainers/     # Create and review DevContainer configurations
├── creating-documentation/     # Create and review README files and repository documentation
├── developer-mentor/           # Guide users through development concepts without writing code
├── git-workflow/               # Git branching, commits, merges, and repository management
├── raise-pull-requests/        # PR preparation, submission, and code review
└── threat-modelling/           # STRIDE-based threat modelling for repositories and cloud infrastructure
```

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

## Loading Skills Directly

Load skills and standards into your AI agent's context without the CLI:

1. **Load a skill** — Read the `SKILL.md` file to understand capabilities
2. **Execute an action** — Follow the procedure in `actions/<action>.md`
3. **Apply standards** — Load relevant standards from `standards/` for consistency

## Contributing

Contributions are welcome. Please open an issue to discuss changes before submitting a pull request.

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.
