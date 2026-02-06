# Python Security Standard

## Overview

Security requirements for Python code.

---

## Prohibited Practices

| Practice | Reason | Alternative |
|----------|--------|-------------|
| `eval()` / `exec()` | Code injection risk | Use `ast.literal_eval()` or proper parsing |
| `pickle` with untrusted data | Arbitrary code execution | Use JSON or validated serialisation |
| Hardcoded secrets | Credential exposure | Use environment variables |
| `shell=True` in subprocess | Command injection | Use argument lists |
| `assert` for validation | Disabled with `-O` flag | Use proper validation |

---

## Secure Practices

### Environment Variables for Secrets

```python
import os

api_key = os.environ.get("API_KEY")
if not api_key:
    raise ValueError("API_KEY environment variable required")
```

### Safe Subprocess Calls

```python
import subprocess
from pathlib import Path

# Use argument lists, not shell=True
result = subprocess.run(
    ["ls", "-la", str(Path.home())],
    capture_output=True,
    text=True,
    check=True,
)
```

### Input Validation with Pydantic

```python
from pydantic import BaseModel, Field

class UserInput(BaseModel):
    username: str = Field(..., min_length=1, max_length=50)
    email: str = Field(..., pattern=r"^[\w\.-]+@[\w\.-]+\.\w+$")
```

### Safe JSON Parsing

```python
import ast

# For literal Python structures
data = ast.literal_eval(user_input)

# For JSON data
import json
data = json.loads(user_input)
```

---

## Compliance Checklist

- [ ] No `eval()` or `exec()` with untrusted input
- [ ] No `pickle` with untrusted data
- [ ] No hardcoded secrets or credentials
- [ ] Subprocess calls use argument lists (no `shell=True`)
- [ ] Input validation uses proper libraries (Pydantic, etc.)
- [ ] Secrets loaded from environment variables
