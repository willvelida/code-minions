# DevContainer Core Standard

## Overview

Core requirements for DevContainer configurations including base image, lifecycle hooks, container configuration, and file structure.

## Principles

1. **Consistency** — Standardised configurations across all projects
2. **Security** — Non-root users, secure defaults, least privilege
3. **Performance** — Lightweight base images, optimised startup
4. **Developer Experience** — Essential tools included, clear documentation

---

## Base Image

### Mandatory Base Image

**ALL DevContainers MUST use:**

```docker
mcr.microsoft.com/devcontainers/base:ubuntu-24.04
```

This base image already includes:

- `common-utils` feature (with `vscode` user)
- `git` feature
- Zsh with Oh My Zsh

**Do NOT add these features.**

### Language Implementation

Use the base image with language-specific **features**:

```json
{
  "image": "mcr.microsoft.com/devcontainers/base:ubuntu-24.04",
  "features": {
    "ghcr.io/devcontainers/features/python:1": {
      "version": "3.12"
    }
  }
}
```

---

## Container Configuration

### Required Settings

```jsonc
{
  "name": "[Project Name]",
  "image": "mcr.microsoft.com/devcontainers/base:ubuntu-24.04",
  "remoteUser": "vscode",
  "privileged": false,
  "shutdownAction": "stopContainer",
  "init": true,
  "postCreateCommand": "bash .devcontainer/post-create.sh",
  "postStartCommand": "bash .devcontainer/post-start.sh"
}
```

| Setting | Value | Purpose |
|---------|-------|---------|
| `remoteUser` | `"vscode"` | Non-root user for security |
| `privileged` | `false` | Disable privileged mode |
| `shutdownAction` | `"stopContainer"` | Clean container shutdown |
| `init` | `true` | Enable init process (tini) for proper signal handling |

---

## Lifecycle Hooks

### Post-Create Hook (Mandatory)

**ALWAYS** create `.devcontainer/post-create.sh`:

```json
{
  "postCreateCommand": "bash .devcontainer/post-create.sh"
}
```

### Post-Start Hook (Mandatory)

**ALWAYS** create `.devcontainer/post-start.sh`:

```json
{
  "postStartCommand": "bash .devcontainer/post-start.sh"
}
```

### Script Execution

| Script | Purpose | Runs As |
|--------|---------|---------|
| `post-create.sh` | One-time setup after container creation | root |
| `post-start.sh` | Runs each time container starts | vscode user |

---

## Modular Script System

Lifecycle hooks use a modular `.d` directory pattern (similar to Linux init systems) for composable script execution.

### Directory Structure

```
.devcontainer/
├── devcontainer.json       # Main configuration (REQUIRED)
├── post-create.sh          # Orchestrator (REQUIRED)
├── post-start.sh           # Orchestrator (REQUIRED)
├── lib/
│   └── run-scripts.sh      # Shared runner library
├── post-create.d/          # Post-create scripts
│   └── 00-git-safe-directory.sh
└── post-start.d/           # Post-start scripts
    └── .gitkeep
```

### How It Works

1. `post-create.sh` and `post-start.sh` source `lib/run-scripts.sh`
2. The runner scans the corresponding `.d` directory for `*.sh` files
3. Scripts execute in **sorted order** (numerical, then alphabetical)
4. Execution is **fail-fast** — first failure stops all subsequent scripts

### Adding Scripts

Create scripts in the appropriate `.d` directory with a numeric prefix for ordering:

```bash
# Example: .devcontainer/post-create.d/10-install-tools.sh
#!/bin/bash
set -euo pipefail

# Install project-specific tools
apt-get update && apt-get install -y jq
```

**Naming convention:** `NN-descriptive-name.sh` where `NN` is a two-digit number.

Scripts with the same prefix execute in alphabetical order (e.g., `10-aaa.sh` before `10-zzz.sh`).

### Disabling Scripts

To temporarily disable a script without deleting it, rename with `.skip.sh` suffix:

```bash
# Disable a script
mv 10-install-tools.sh 10-install-tools.skip.sh

# Re-enable
mv 10-install-tools.skip.sh 10-install-tools.sh
```

The runner ignores all `*.skip.sh` files.

### Script Requirements

Each script in a `.d` directory **MUST**:

1. Include shebang: `#!/bin/bash`
2. Enable strict mode: `set -euo pipefail`
3. Be executable or invoked via `bash`

### Mandatory Scripts

| Directory | Script | Purpose |
|-----------|--------|---------|
| `post-create.d/` | `00-git-safe-directory.sh` | Configure git safe directory for workspace |

---

## File Structure

```
.devcontainer/
├── devcontainer.json       # Main configuration (REQUIRED)
├── post-create.sh          # Post-create orchestrator (REQUIRED)
├── post-start.sh           # Post-start orchestrator (REQUIRED)
├── lib/
│   └── run-scripts.sh      # Shared runner library (REQUIRED)
├── post-create.d/          # Post-create scripts (REQUIRED)
│   └── 00-git-safe-directory.sh
└── post-start.d/           # Post-start scripts (REQUIRED)
```

---

## References

- [DevContainer Specification](https://containers.dev/implementors/spec/)
- [VS Code DevContainer Documentation](https://code.visualstudio.com/docs/devcontainers/containers)
- [Base Image Source](https://github.com/devcontainers/images/blob/main/src/base-ubuntu/.devcontainer/devcontainer.json)
