---
name: creating-devcontainers
description: 'Create and review DevContainer configurations that follow organisational standards. Use when a user asks to set up a dev container, configure a development environment, create a devcontainer.json, add lifecycle hooks, review an existing DevContainer for compliance, or improve container security. Covers base images, features, extensions, lifecycle hooks, and security hardening.'
license: MIT
compatibility: 'VS Code with Dev Containers extension (ms-vscode-remote.remote-containers)'
allowed-tools: Bash
---

# DevContainer Skill

## Overview

This skill provides capabilities for creating and reviewing DevContainer configurations that comply with organisational standards.

## Capabilities

| Capability | Action | Description |
|------------|--------|-------------|
| Create | `actions/create.md` | Generate a standards-compliant DevContainer configuration |
| Review | `actions/review.md` | Analyse an existing DevContainer for compliance and improvements |

## Standards

This skill bundles the following standards in `standards/`:

| Standard | File | Description |
|----------|------|-------------|
| Core | `core.md` | Base image, lifecycle hooks, container configuration, file structure |
| Extensions | `extensions.md` | VS Code extension requirements and guidelines |
| Features | `features.md` | DevContainer features configuration and versioning |
| Security | `security.md` | Security requirements and risk guidance |
| Checklist | `checklist.md` | Consolidated compliance checklist |
| Template | `template.json` | Base configuration template |

## Usage

1. Load this skill manifest
2. Identify the required capability (create or review)
3. Load the bundled standards from `standards/`
4. Ask the user what language(s) and tooling the DevContainer needs to accommodate
5. Execute the action following `actions/<capability>.md`
