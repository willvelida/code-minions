# Bash Development Environment Standard

## Overview

DevContainer configuration and VS Code settings for Bash development.

---

## DevContainer Features

```jsonc
{
  "features": {
    "ghcr.io/devcontainers-extra/features/shellcheck:1": { "version": "0.11.0" },
    "ghcr.io/devcontainers-extra/features/shfmt:1": { "version": "3.12.0" }
  }
}
```

**Note**: Bash is included in the mandatory base image (`mcr.microsoft.com/devcontainers/base:ubuntu-24.04`). The features above install ShellCheck and shfmt for linting and formatting.

---

## VS Code Extensions

```jsonc
{
  "extensions": [
    /* ---------------------------------------------------------------------------
       Bash Development
    --------------------------------------------------------------------------- */
    "mads-hartmann.bash-ide-vscode",
    // Bash Language Server - provides code completion, hover documentation,
    // go-to-definition, and symbol highlighting.

    "rogalmic.bash-debug",
    // Bash debugging support including breakpoints, stepping, and variable
    // inspection.

    "timonwong.shellcheck",
    // ShellCheck integration for real-time linting of shell scripts.

    "foxundermoon.shell-format"
    // shfmt integration for consistent shell script formatting.
  ]
}
```

| Extension | Purpose |
|-----------|---------|
| `mads-hartmann.bash-ide-vscode` | Language server for IDE features |
| `rogalmic.bash-debug` | Debugging support |
| `timonwong.shellcheck` | Real-time linting |
| `foxundermoon.shell-format` | Formatting with shfmt |

---

## VS Code Settings

```json
{
  "shellcheck.enable": true,
  "shellcheck.run": "onSave",
  "shellformat.flag": "-i 2 -ci",
  "[shellscript]": {
    "editor.defaultFormatter": "foxundermoon.shell-format",
    "editor.formatOnSave": true,
    "editor.tabSize": 2
  }
}
```

---

## Settings Explained

| Setting | Value | Purpose |
|---------|-------|---------|
| `shellcheck.enable` | `true` | Enable ShellCheck linting |
| `shellcheck.run` | `onSave` | Run ShellCheck when file is saved |
| `shellformat.flag` | `-i 2 -ci` | 2-space indent, indent case statements |
| `editor.defaultFormatter` | `foxundermoon.shell-format` | Use shfmt for formatting |
| `editor.formatOnSave` | `true` | Auto-format on save |
| `editor.tabSize` | `2` | Use 2-space indentation |
