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

## Skills

```text
skills/
├── create-agent-skill/          # Create and review Agent Skills following the open specification
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

## Usage

Load skills and standards into your AI agent's context as needed:

1. **Load a skill** — Read the `SKILL.md` file to understand capabilities
2. **Execute an action** — Follow the procedure in `actions/<action>.md`
3. **Apply standards** — Load relevant standards from `standards/` for consistency

## Contributing

Contributions are welcome. Please open an issue to discuss changes before submitting a pull request.

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.
