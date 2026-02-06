# DevContainer Security Standard

## Overview

Security requirements for DevContainer configurations. Focus on risk rationale and guidance.

---

## 🔴 Critical Requirements

### Non-Root User

**ALWAYS** use a non-root user:

```json
{
  "remoteUser": "vscode"
}
```

**Risk**: Running as root bypasses security boundaries. A compromised process gains full container access and potentially host access through mounted volumes.

### No Privileged Mode

**NEVER** use privileged mode without documented justification:

```jsonc
// ❌ PROHIBITED without justification
{
  "privileged": true
}
```

**Risk**: Privileged mode grants full host capabilities, effectively disabling container isolation. The container can access host devices, bypass security modules, and modify host kernel parameters.

**Exception**: Docker-in-Docker may require `privileged: true`. Document the justification in the configuration.

### No Hardcoded Secrets

**NEVER** hardcode credentials in configuration:

```jsonc
// ❌ PROHIBITED
{
  "containerEnv": {
    "API_KEY": "sk-1234567890abcdef"
  }
}
```

**Risk**: Secrets in configuration files are committed to version control, visible in container inspection, and may leak through logs or error messages.

**Alternative**: Use environment variables, mounted secrets, or credential managers.

---

## 🟠 High Priority Requirements

### Git Safe Directory

**ALWAYS** configure git safe.directory in post-create hook:

```bash
git config --global --add safe.directory "${containerWorkspaceFolder:-/workspaces}"
```

**Risk**: Without this configuration, git may refuse to operate due to ownership mismatches between host and container users.

### Restricted Host Mounts

Limit host filesystem mounts:

```jsonc
// ❌ AVOID broad mounts
{
  "mounts": [
    "source=${localEnv:HOME},target=/host-home,type=bind"
  ]
}

// ✅ Use specific paths with read-only where possible
{
  "mounts": [
    "source=${localEnv:HOME}/.ssh,target=/home/vscode/.ssh,type=bind,readonly"
  ]
}
```

**Risk**: Broad host mounts expose sensitive files. A compromised container could read or modify host data.

### Secure Base Image

**ALWAYS** use the mandatory base image:

```docker
mcr.microsoft.com/devcontainers/base:ubuntu-24.04
```

**Risk**: Unofficial or outdated images may contain known vulnerabilities, malware, or insecure defaults.

---

## 🟡 Medium Priority Requirements

### Versioned Features

**ALWAYS** specify explicit versions (see `features.md`).

**Risk**: Unversioned features reduce reproducibility and may introduce unexpected changes or vulnerabilities.

### Port Forwarding

Review and restrict port forwarding:

```jsonc
// ✅ Only forward necessary ports
{
  "forwardPorts": [3000, 5432],
  "portsAttributes": {
    "3000": {
      "label": "Web App",
      "onAutoForward": "notify"
    }
  }
}
```

**Risk**: Unnecessary port forwarding increases attack surface.

---

## Best Practices

1. **Least Privilege** — Only grant permissions necessary for development
2. **Secrets Management** — Use environment variables or credential managers
3. **Mount Restrictions** — Use specific paths rather than broad host mounts
4. **Read-Only Mounts** — Use `readonly` flag where possible
5. **Port Security** — Only forward ports required for development
6. **Regular Updates** — Use versioned features for reproducible, updatable configurations
