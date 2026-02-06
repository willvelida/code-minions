# Python Tooling Standard

## Overview

Standard tooling for Python development: package management, linting, formatting, and type checking.

---

## Package Manager: uv

**uv** is the mandatory package manager for Python projects.

**Rationale**:
- 10-100x faster than pip
- Built-in virtual environment management
- Lock file support for reproducibility
- Drop-in replacement for pip commands

**Usage:**

```bash
# Create virtual environment
uv venv

# Sync dependencies from pyproject.toml/uv.lock
uv sync

# Add a dependency
uv add requests

# Add a dev dependency
uv add --dev pytest

# Remove a dependency
uv remove requests

# Run a command in the virtual environment
uv run python script.py

# Run tests
uv run pytest
```

---

## Linter & Formatter: Ruff

**Ruff** is the mandatory linter and formatter for all Python projects.

**Rationale**:
- 10-100x faster than flake8, black, isort combined
- Single tool replaces multiple linters
- Highly configurable
- Auto-fix capabilities

**Usage:**

```bash
# Lint
ruff check .

# Lint with auto-fix
ruff check --fix .

# Format
ruff format .
```

### Ruff Configuration

Add to `pyproject.toml`:

```toml
[tool.ruff]
target-version = "py311"
line-length = 100
src = ["src"]

[tool.ruff.lint]
select = [
    "E",      # pycodestyle errors
    "W",      # pycodestyle warnings
    "F",      # Pyflakes
    "I",      # isort
    "B",      # flake8-bugbear
    "C4",     # flake8-comprehensions
    "UP",     # pyupgrade
    "ARG",    # flake8-unused-arguments
    "SIM",    # flake8-simplify
]
ignore = [
    "E501",   # line too long (handled by formatter)
]

[tool.ruff.lint.isort]
known-first-party = ["my_project"]

[tool.ruff.format]
quote-style = "double"
indent-style = "space"
docstring-code-format = true
```

### Enabled Lint Rules

| Rule Set | Code | Purpose |
|----------|------|---------|
| pycodestyle errors | `E` | PEP 8 style errors |
| pycodestyle warnings | `W` | PEP 8 style warnings |
| Pyflakes | `F` | Logical errors and undefined names |
| isort | `I` | Import sorting and organisation |
| flake8-bugbear | `B` | Common bugs and design problems |
| flake8-comprehensions | `C4` | Simplify list/dict comprehensions |
| pyupgrade | `UP` | Upgrade syntax for newer Python |
| flake8-unused-arguments | `ARG` | Detect unused function arguments |
| flake8-simplify | `SIM` | Simplify code patterns |

---

## Type Checker: Pylance / mypy

**Pylance** provides type checking through VS Code.

For CI/CD, use **mypy**:

```bash
uv add --dev mypy
uv run mypy src/
```

### Mypy Configuration

Add to `pyproject.toml`:

```toml
[tool.mypy]
python_version = "3.11"
strict = true
warn_return_any = true
warn_unused_ignores = true
disallow_untyped_defs = true
disallow_incomplete_defs = true
check_untyped_defs = true
```

### Pylance Type Checking Modes

| Mode | Description |
|------|-------------|
| `off` | No type checking |
| `basic` | Basic checks, minimal false positives |
| `standard` | **Recommended**: Balance of coverage and practicality |
| `strict` | Maximum strictness |

---

## Test Framework: pytest

**pytest** is the standard test framework.

### Configuration

Add to `pyproject.toml`:

```toml
[tool.pytest.ini_options]
testpaths = ["tests"]
python_files = ["test_*.py"]
python_functions = ["test_*"]
addopts = "-v --tb=short"
```

---

## Post-Start Hook

Add to `.devcontainer/post-start.d/10-sync-pyproject-dependencies.sh`:

```bash
#!/bin/bash
#
# Script Name: 10-sync-pyproject-dependencies.sh
# Description: Sync dependencies if pyproject.toml exists
#
# Usage: Executed automatically by post-start.sh
#
set -euo pipefail

if [ -f "pyproject.toml" ]; then
    uv sync
fi
```
