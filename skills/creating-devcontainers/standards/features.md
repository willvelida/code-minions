# DevContainer Features Standard

## Overview

Requirements for DevContainer features configuration and versioning.

---

## Versioning

**MANDATORY**: All features MUST specify explicit versions.

```jsonc
// ❌ PROHIBITED
{
  "features": {
    "ghcr.io/devcontainers/features/python:1": {
      "version": "latest"
    }
  }
}

// ✅ REQUIRED
{
  "features": {
    "ghcr.io/devcontainers/features/python:1": {
      "version": "3.12"
    }
  }
}
```

| Prohibited | Reason |
|------------|--------|
| `"latest"` | Non-reproducible builds, unexpected breaking changes |
| `"lts"` | Version changes over time, non-deterministic |
| Omitted version | Defaults to latest, same issues |

---

## Feature Sources

### Official Features

Prefer official DevContainer features from:

```
ghcr.io/devcontainers/features/<feature>:<major-version>
```

### Community Features

Community features from `ghcr.io/devcontainers-extra/features/` are acceptable when:

1. No official equivalent exists
2. Feature is well-maintained
3. Version is explicitly specified

---

## Excluded Features

The mandatory base image (`mcr.microsoft.com/devcontainers/base:ubuntu-24.04`) already includes:

- `common-utils` (provides `vscode` user)
- `git`

**Do NOT add these features** — they are redundant and may cause conflicts.

---

## References

- [Official DevContainer Features](https://containers.dev/features)
- [Features Specification](https://containers.dev/implementors/features/)
