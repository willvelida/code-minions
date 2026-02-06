# Bash Tooling Standard

## Overview

Standard tooling for Bash development: linting and formatting.

---

## Linter: ShellCheck

**ShellCheck** is the mandatory linter for all shell scripts.

**Usage:**

```bash
shellcheck script.sh
```

### ShellCheck Configuration

Create `.shellcheckrc` in the repository root:

```bash
# .shellcheckrc - ShellCheck configuration
shell=bash
severity=style
enable=all
```

---

## Formatter: shfmt

**shfmt** is the mandatory formatter for shell scripts.

**Usage:**

```bash
# Format a file
shfmt -w script.sh

# Format with specific options
shfmt -i 2 -ci -w script.sh
```

### shfmt Options

| Option | Purpose |
|--------|---------|
| `-i 2` | Use 2-space indentation |
| `-ci` | Indent switch cases |
| `-w` | Write result to file |
| `-d` | Display diff instead of writing |

---

## VS Code Integration

- **ShellCheck**: The `timonwong.shellcheck` extension provides real-time linting
- **shfmt**: The `foxundermoon.shell-format` extension provides format-on-save
