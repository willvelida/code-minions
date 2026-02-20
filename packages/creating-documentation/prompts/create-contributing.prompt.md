---
description: "Generate a CONTRIBUTING guide for the project"
mode: agent
input:
  - name: project_name
    description: "Name of the project"
---

# Create Contributing Guide

Generate a CONTRIBUTING.md for **${input:project_name}**.

## Requirements

1. Analyse the repository for:
   - Build system and toolchain
   - Testing framework and conventions
   - Existing contribution patterns (PRs, issues)

2. Generate a CONTRIBUTING.md with:
   - **Getting started** — prerequisites and local setup
   - **Development workflow** — branching, committing, testing
   - **Code standards** — style, linting, formatting
   - **Testing** — how to run tests, coverage expectations
   - **Pull request process** — what reviewers look for
   - **Issue reporting** — how to file bugs and feature requests
   - **Code of conduct** — link to or embed a code of conduct

3. Include commands for common development tasks

4. Keep instructions OS-agnostic where possible
