# Python Core Standard

## Overview

Standards for Python development including project structure, naming conventions, and coding style.

## Principles

1. **PEP Compliance** — Follow PEP-8, PEP-621, and other Python Enhancement Proposals
2. **Modern Tooling** — Use fast, modern tools (uv, Ruff) over legacy alternatives
3. **Type Safety** — Leverage type hints for better code quality
4. **Reproducibility** — Ensure consistent environments across machines

---

## Project Structure

### Recommended Layout

```
project/
├── src/
│   └── my_project/
│       ├── __init__.py
│       ├── main.py
│       └── utils/
│           ├── __init__.py
│           └── helpers.py
├── tests/
│   ├── __init__.py
│   ├── conftest.py
│   └── test_main.py
├── pyproject.toml
├── README.md
└── .python-version
```

### Source Layout

Use the `src/` layout for packages:

- Prevents accidental imports of development code
- Clear separation between source and tests
- Better compatibility with build tools

### Python Version File

Create `.python-version` to specify Python version:

```
3.12
```

---

## Project Configuration

### pyproject.toml (PEP-621)

All Python projects **MUST** use `pyproject.toml` for project metadata:

```toml
[project]
name = "my-project"
version = "0.1.0"
description = "A brief description of the project"
readme = "README.md"
requires-python = ">=3.11"
dependencies = [
    "requests>=2.28.0",
]

[project.optional-dependencies]
dev = [
    "ruff>=0.1.0",
    "mypy>=1.0.0",
    "pytest>=7.0.0",
]

[build-system]
requires = ["hatchling"]
build-backend = "hatchling.build"
```

---

## Naming Conventions

| Type | Convention | Example |
|------|------------|---------|
| Variables | snake_case | `user_name`, `is_active` |
| Functions | snake_case | `get_user_by_id`, `validate_input` |
| Classes | PascalCase | `UserService`, `HttpClient` |
| Constants | UPPER_SNAKE_CASE | `MAX_RETRIES`, `API_URL` |
| Modules | snake_case | `user_service.py`, `api_client.py` |
| Packages | lowercase | `mypackage`, `utilities` |

---

## Type Hints

Type hints are **mandatory** for all Python code.

### Basic Types

```python
name: str = "Alice"
count: int = 42
is_active: bool = True

def greet(name: str) -> str:
    return f"Hello, {name}"

def log_message(message: str) -> None:
    print(message)
```

### Collection Types (Python 3.9+)

```python
def process_items(items: list[str]) -> dict[str, int]:
    return {item: len(item) for item in items}

names: set[str] = {"alice", "bob"}
mapping: dict[str, list[int]] = {"numbers": [1, 2, 3]}
```

### Optional and Union Types (Python 3.10+)

```python
def find_user(user_id: str) -> User | None:
    ...

def parse_value(value: str) -> int | float | None:
    ...
```

### When to Use `Any`

Avoid `Any` when possible. Valid uses:

```python
# Valid: Truly dynamic data from external sources
def parse_json_response(response: str) -> dict[str, Any]:
    ...
```

---

## Docstrings (Google Style)

```python
def fetch_user(user_id: str, include_profile: bool = False) -> User:
    """Fetch a user by their ID.

    Args:
        user_id: The unique identifier of the user.
        include_profile: Whether to include the full profile data.

    Returns:
        The User object with the requested data.

    Raises:
        UserNotFoundError: If no user exists with the given ID.
    """
    ...
```

---

## Import Organisation

Ruff handles import sorting automatically. Manual organisation:

```python
# Standard library
import os
from pathlib import Path

# Third-party
import requests
from pydantic import BaseModel

# Local
from my_project.utils import helpers
```
