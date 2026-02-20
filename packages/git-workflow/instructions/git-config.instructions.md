---
description: Git workflow and CI/CD configuration standards
applyTo: "**/.github/**,**/.gitlab-ci.yml,**/.gitlab/**"
---

# Git Workflow Standards

When creating or editing Git workflow and CI/CD configuration files:

- Use conventional branch naming: `feat/`, `fix/`, `chore/`, `docs/`, `refactor/`
- Write conventional commit messages: `type(scope): description`
- Configure branch protection rules for the default branch
- Require pull request reviews before merging
- Run CI checks (lint, test, build) on pull request events
- Use environment variables or secrets for sensitive values — never hardcode
- Pin action versions to full commit SHAs, not tags
- Include a `CODEOWNERS` file for automatic review assignment
