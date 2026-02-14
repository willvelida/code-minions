# Security Policy

## Supported Versions

Only the latest release on `main` receives security fixes. Older versions are not backported.

| Version | Supported |
|---------|-----------|
| Latest release (`main`) | ✅ |
| Older releases | ❌ |

## Reporting a Vulnerability

**Please do not open a public issue for security vulnerabilities.**

To report a vulnerability privately:

1. Go to the [Security](https://github.com/willvelida/code-minions/security) tab of this repository
2. Click **"Report a vulnerability"**
3. Fill in the advisory form with the details described below

If private vulnerability reporting is unavailable, email **willvelida@hotmail.co.uk** with the subject line `[SECURITY] code-minions vulnerability report`.

## What to Include

To help us understand and fix the issue quickly, please include:

- **Description** — what the vulnerability is and which component is affected
- **Steps to reproduce** — a minimal set of steps to trigger the issue
- **Impact** — what an attacker could achieve by exploiting this (e.g., arbitrary file write, data exposure)
- **Suggested fix** (optional) — if you have an idea for how to resolve it

## Response Timeline

These are best-effort targets for a maintainer-run open-source project:

| Stage | Timeframe |
|-------|-----------|
| Acknowledgment | Within **7 days** |
| Initial assessment | Within **14 days** |
| Fix or mitigation | Within **30 days** for confirmed issues |

If you haven't heard back within the acknowledgment window, feel free to follow up on your report.

## Scope

### In scope

- Path traversal or arbitrary file write/delete in the installer
- Malicious content injection via packages or standards
- Command injection through CLI arguments
- Sensitive data exposure (credentials, tokens, secrets)

### Out of scope

- General bugs (use [Issues](https://github.com/willvelida/code-minions/issues))
- Feature requests
- Vulnerabilities in third-party dependencies (report those to the upstream project)

## Disclosure Policy

- We will work with you to understand and resolve the issue before any public disclosure
- A [GitHub Security Advisory](https://docs.github.com/en/code-security/security-advisories) will be published once a fix is released
- You will be credited in the advisory unless you prefer to remain anonymous
