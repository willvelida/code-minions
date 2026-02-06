# DevContainer Compliance Checklist

## Overview

Consolidated checklist for validating DevContainer configurations against all standards.

---

## Configuration

- [ ] Uses mandatory base image (`mcr.microsoft.com/devcontainers/base:ubuntu-24.04`)
- [ ] Language-specific features with explicit versions
- [ ] No redundant features (`common-utils`, `git`)
- [ ] Essential extensions included
- [ ] All extensions bundled and documented with comments
- [ ] Language-specific extensions included per language standards
- [ ] Required container settings configured:
  - [ ] `remoteUser: "vscode"`
  - [ ] `privileged: false`
  - [ ] `shutdownAction: "stopContainer"`
  - [ ] `init: true`

---

## Lifecycle Hooks

- [ ] `post-create.sh` exists
- [ ] `post-create.sh` includes `git config --global --add safe.directory`
- [ ] `post-start.sh` exists
- [ ] Uses `bash .devcontainer/*.sh` execution pattern for both hooks
- [ ] Language-specific setup included in hooks

---

## Security — Critical

- [ ] Non-root user configured (`remoteUser: "vscode"`)
- [ ] No `privileged: true` (unless documented justification)
- [ ] No hardcoded secrets or credentials

---

## Security — High Priority

- [ ] Git safe.directory configured in post-create hook
- [ ] Host mounts are minimal and restricted
- [ ] Using mandatory base image

---

## Security — Medium Priority

- [ ] All features have explicit versions (no `latest`, `lts`)
- [ ] Port forwarding limited to development needs

---

## File Structure

- [ ] `.devcontainer/devcontainer.json` exists
- [ ] `.devcontainer/post-create.sh` exists
- [ ] `.devcontainer/post-start.sh` exists
