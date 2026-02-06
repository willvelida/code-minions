# Python Development Environment Standard

## Overview

DevContainer configuration and VS Code settings for Python development.

---

## DevContainer Features

```jsonc
{
  "features": {
    "ghcr.io/devcontainers/features/python:1": {
      "version": "3.12"
    },
    "ghcr.io/devcontainers-extra/features/uv:1": {},
    "ghcr.io/devcontainers-extra/features/ruff:1": {}
  }
}
```

**Note**: The Python feature installs Python and pip. Additional tools (uv, Ruff) are installed via devcontainer features for better container image layer caching.

---

## VS Code Extensions

```jsonc
{
  "extensions": [
    /* ---------------------------------------------------------------------------
       Python Development
    --------------------------------------------------------------------------- */
    "ms-python.python",
    // Core Python language support including debugging, IntelliSense, and Jupyter
    // notebook integration.

    "ms-python.vscode-pylance",
    // Fast language server providing IntelliSense, type checking, auto-imports,
    // and go-to-definition.

    "charliermarsh.ruff",
    // Extremely fast Python linter and formatter. Handles style enforcement,
    // import sorting, and code formatting.

    "tamasfe.even-better-toml"
    // TOML file support for pyproject.toml editing.
  ]
}
```

| Extension | Purpose |
|-----------|---------|
| `ms-python.python` | Core Python language support, debugging |
| `ms-python.vscode-pylance` | Language server: IntelliSense, type checking |
| `charliermarsh.ruff` | Linter and formatter |
| `tamasfe.even-better-toml` | TOML file support |

---

## Pylance and Ruff Coexistence

Pylance and Ruff serve complementary purposes:

- **Pylance** provides IDE intelligence and static type analysis
- **Ruff** provides fast linting rules and code formatting

To avoid duplicate diagnostics, configure Pylance to defer linting to Ruff with the following settings:

```json
{
  "python.analysis.typeCheckingMode": "standard",
  "python.analysis.diagnosticSeverityOverrides": {
    "reportMissingTypeStubs": "none"
  }
}
```

---

## VS Code Settings

```json
{
  "python.defaultInterpreterPath": "${workspaceFolder}/.venv/bin/python",
  "python.analysis.typeCheckingMode": "standard",
  "python.analysis.autoImportCompletions": true,
  "python.analysis.diagnosticSeverityOverrides": {
    "reportMissingTypeStubs": "none"
  },
  "[python]": {
    "editor.defaultFormatter": "charliermarsh.ruff",
    "editor.formatOnSave": true,
    "editor.codeActionsOnSave": {
      "source.fixAll.ruff": "explicit",
      "source.organizeImports.ruff": "explicit"
    }
  },
  "ruff.organizeImports": true,
  "ruff.fixAll": true
}
```

---

## Settings Explained

| Setting | Value | Purpose |
|---------|-------|---------|
| `python.defaultInterpreterPath` | `.venv/bin/python` | Use project virtual environment |
| `python.analysis.typeCheckingMode` | `standard` | Balanced type checking strictness |
| `python.analysis.autoImportCompletions` | `true` | Enable auto-import suggestions |
| `editor.defaultFormatter` | `charliermarsh.ruff` | Use Ruff for formatting |
| `editor.formatOnSave` | `true` | Auto-format on save |
| `source.fixAll.ruff` | `explicit` | Apply Ruff fixes on save |
| `source.organizeImports.ruff` | `explicit` | Sort imports on save |
