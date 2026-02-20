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

- **Packages** — Bundled agents, skills, instructions, and prompts for common workflows (git operations, documentation, DevContainers)

## CLI

code-minions includes a CLI tool that installs packages (made up of Agents, Agent Skills, Instructions, and Prompts) into your own repositories.

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
# Scaffold a project manifest interactively
code-minions init

# Non-interactive init with defaults (uses "standard" template)
code-minions init --yes

# Use a specific template
code-minions init --template security

# List available templates
code-minions init --list-templates

# Non-interactive with a specific template
code-minions init --template fullstack --assistant copilot --yes

# Specify assistant and packages explicitly (no template)
code-minions init --assistant copilot --packages threat-modelling,git-workflow

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
code-minions install --package developer-mentor --for codex
code-minions install --package developer-mentor --for cursor
code-minions install --package developer-mentor --for gemini
code-minions install --package developer-mentor --for opencode

# Install from an external package source (Git URL)
code-minions install --package my-skill --from github.com/org/packages

# Install from a configured named source
code-minions install --package my-skill --from my-team

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

### Init flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--assistant` | string | | Assistant to configure (skips prompt) |
| `--packages` | string | `""` (all in non-interactive mode) | Comma-separated packages to include (skips prompt) |
| `--target` | string | `.` | Target directory for the manifest |
| `--yes` | bool | false | Accept defaults without prompting |
| `--force` | bool | false | Overwrite existing `code-minions.yml` |
| `--json` | bool | false | Output results as JSON |

### Install flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--package` | string | all | Comma-separated list of packages to install |
| `--for` | string | | Target coding assistant (`claude`, `codex`, `copilot`, `cursor`, `gemini`, `opencode`) |
| `--from` | string | | Install from a specific source (configured name or Git URL) |
| `--target` | string | `.` | Target directory for installation |
| `--dry-run` | bool | false | Show what would be installed without writing files |
| `--force` | bool | false | Overwrite existing files |
| `--json` | bool | false | Output results as JSON |
| `--verbose` / `-v` | bool | false | Show detailed output |
| `--quiet` / `-q` | bool | false | Suppress all output except errors |

When installing packages, an `AGENTS.md` file is automatically created if one does not already exist. Existing `AGENTS.md` files are never overwritten.

When installing with `--for claude`, a `CLAUDE.md` file is also generated at the project root. This file uses Claude Code's native `@path/to/file` import syntax to reference installed skills and agents, giving Claude Code direct visibility into your installed capabilities. Like `AGENTS.md`, existing `CLAUDE.md` files are never overwritten.

#### Installing from external sources

The `--from` flag lets you install packages from Git repositories outside the built-in registry. The source must follow the standard `code-minions` layout with a `packages/` directory at the root.

```bash
# Install from a Git URL
code-minions install --package developer-mentor --from https://github.com/willvelida/MinionsHub.git

# Install from a bare GitHub path (https:// and .git are optional)
code-minions install --package developer-mentor --from github.com/willvelida/MinionsHub

# Install for a specific assistant
code-minions install --package developer-mentor --from github.com/willvelida/MinionsHub --for copilot

# Install multiple packages
code-minions install --package threat-modelling,git-workflow --from github.com/willvelida/MinionsHub

# Preview what would be installed
code-minions install --package developer-mentor --from github.com/willvelida/MinionsHub --dry-run
```

You can also configure named sources so you don't have to type the full URL every time:

```bash
# Add a named source
code-minions source add my-team --url https://github.com/my-org/agent-packages.git

# Install using the source name
code-minions install --package my-skill --from my-team

# List packages from all configured sources
code-minions list
```

The [willvelida/MinionsHub](https://github.com/willvelida/MinionsHub) repository is a public example of a compatible package source.

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

The uninstall command prompts for confirmation before removing files. You will also be prompted separately before `AGENTS.md` is removed. When uninstalling with `--for claude`, you will also be prompted before `CLAUDE.md` is removed. In non-interactive environments (no TTY), the command aborts unless `--yes` is passed.

### Uninstall flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--package` | string | all | Comma-separated list of packages to uninstall |
| `--for` | string | | Target coding assistant (`claude`, `codex`, `copilot`, `cursor`, `gemini`, `opencode`) — **required** when uninstalling everything |
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

Update overwrites installed files with the latest embedded content. When run with no flags, only packages already present in the target directory are updated — no new packages are installed. `AGENTS.md` and `CLAUDE.md` are not modified during updates.

### Update flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--package` | string | | Comma-separated list of packages to update (omit to auto-detect installed) |
| `--for` | string | | Target coding assistant (`claude`, `codex`, `copilot`, `cursor`, `gemini`, `opencode`) |
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

MCP server configurations are automatically translated between the source and target formats when present. An `AGENTS.md` routing file is always regenerated (not copied) in the target layout because it contains assistant-specific paths. When transferring to Claude, a `CLAUDE.md` file is also regenerated with the correct `@import` references for the target layout.

### Transfer flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--from` | string | | Source coding assistant (`claude`, `codex`, `copilot`, `cursor`, `gemini`, `opencode`) — **required** |
| `--to` | string | | Target coding assistant (`claude`, `codex`, `copilot`, `cursor`, `gemini`, `opencode`) — **required** |
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

| Assistant | `--for` value | Agents placed in | Skills placed in | Instructions placed in | Prompts placed in |
|-----------|--------------|------------------|------------------|------------------------|-------------------|
| GitHub Copilot | `copilot` | `.github/agents/` | `skills/` | `.github/instructions/` | `.github/prompts/` (`.prompt.md`) |
| Claude Code | `claude` | `.claude/agents/` | `.claude/skills/` | `.claude/instructions/` | `.claude/commands/` (`.md`) |
| Codex CLI | `codex` | `.agents/agents/` | `.agents/skills/` | `.agents/instructions/` | `.agents/prompts/` (`.md`) |
| Cursor | `cursor` | `.cursor/agents/` | `.cursor/skills/` | `.cursor/rules/` (as `.mdc`) | `.cursor/rules/` (`.mdc`) |
| Gemini CLI | `gemini` | `.gemini/agents/` | `.gemini/skills/` | `.gemini/instructions/` | `.gemini/commands/` (`.toml`) |
| OpenCode | `opencode` | `.opencode/agents/` | `.opencode/skills/` | `.opencode/instructions/` | `.opencode/commands/` (`.md`) |

Without `--for`, files are installed to generic locations (`agents/`, `skills/`, `instructions/`, `prompts/`).

Instruction files (`.instructions.md`) use `applyTo` frontmatter to target specific file patterns. For Cursor, they are automatically converted to `.mdc` format with `globs` frontmatter.

Prompt files (`.prompt.md`) are reusable slash commands that users can invoke from their assistant's chat interface. See [Prompts](#prompts) below for the canonical format and translation details.

## Prompts

Prompts are reusable slash commands (`.prompt.md` files) that users can invoke from their coding assistant's chat interface. Each package can include prompt files in a `prompts/` directory.

### Canonical format

The canonical format is GitHub Copilot's `.prompt.md` with YAML frontmatter:

```markdown
---
description: "Short description of what this prompt does"
mode: ask
input:
  - name: concept
    description: "The concept to explain"
  - name: scope
    description: "Optional scope constraint"
---

# Prompt Title

Prompt body using **${input:concept}** and **${input:scope}** variable syntax.
```

**Frontmatter fields:**

| Field | Required | Description |
|-------|----------|-------------|
| `description` | Yes | Short summary shown in the slash command picker |
| `mode` | No | Copilot-specific mode (`ask`, `edit`, `agent`) — dropped for other assistants |
| `input` | No | Array of named input variables the user is prompted to fill |

### Translation per assistant

When installing with `--for`, prompt files are automatically translated to each assistant's native format:

| Assistant | Format | Extension | Variable syntax | Notes |
|-----------|--------|-----------|-----------------|-------|
| Copilot | Markdown + YAML frontmatter | `.prompt.md` | `${input:name}` | Canonical — no translation |
| Claude | Markdown + YAML frontmatter | `.md` | `$name` | `input` array → named frontmatter keys; `mode` dropped |
| OpenCode | Markdown | `.md` | `$1`, `$2` / `$ARGUMENTS` | Positional args (≤3 inputs) or `$ARGUMENTS` (4+); `mode`/`input` dropped |
| Gemini | TOML | `.toml` | `{{args}}` | Body in triple-quoted `prompt` key; all variables → `{{args}}` |
| Cursor | MDC (Markdown Components) | `.mdc` | `[placeholder]` | Added `alwaysApply: false`; variables replaced with `[placeholder]` |
| Codex | Markdown reference | `.md` | `[placeholder]` | Reference document; variables replaced with `[placeholder]` |

### Example

A prompt file at `packages/developer-mentor/prompts/explain-concept.prompt.md`:

```markdown
---
description: "Get a level-adaptive explanation of a software concept"
mode: ask
input:
  - name: concept
    description: "The software concept to explain"
---

# Explain Concept

Explain **${input:concept}** using a level-adaptive mentoring approach.
```

Installed with `--for claude`, this becomes `.claude/commands/explain-concept.md`:

```markdown
---
description: "Get a level-adaptive explanation of a software concept"
concept: "The software concept to explain"
---

# Explain Concept

Explain **$concept** using a level-adaptive mentoring approach.
```

Installed with `--for gemini`, this becomes `.gemini/commands/explain-concept.toml`:

```toml
description = "Get a level-adaptive explanation of a software concept"
prompt = """

# Explain Concept

Explain **{{args}}** using a level-adaptive mentoring approach.
"""
```

## Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, testing expectations, and PR guidelines.

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for a full release history.

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.
