---
name: git-workflow
description: 'Execute a complete git workflow for collaborative development. Use when user asks to initialize a repo, create branches, commit changes, sync with upstream, open pull requests, resolve conflicts, merge PRs, create releases, or apply hotfixes. Supports Centralized, Feature Branch, Gitflow, and Forking workflows with conventional commits, branch naming standards, and code review guidelines.'
license: MIT
compatibility: 'Git 2.x, GitHub CLI (gh) for PR commands'
allowed-tools: Bash
---

# Git Workflow

## Overview

A complete git workflow skill covering the full development lifecycle — from repository setup through release. Supports four workflow models based on team size and release cadence.

## Workflow Models

### 1. Centralized Workflow

All developers commit directly to `main`. Best for small teams or SVN migrations.

- Single branch (`main`)
- Pull with rebase to stay linear
- Conflict resolution before push

### 2. Feature Branch Workflow

All development happens on dedicated branches. Core idea: `main` never contains broken code.

- Create a branch per feature or fix
- Open a pull request for code review
- Merge into `main` after approval

### 3. Gitflow Workflow

Strict branching model designed around project releases. Extends Feature Branch with release management.

- `main` — production-ready code
- `develop` — integration branch for features
- `feature/*` — new features branch from `develop`
- `release/*` — release preparation branches
- `hotfix/*` — urgent fixes branch from `main`

### 4. Forking Workflow

Each developer has their own server-side fork. Common in open source.

- Fork the upstream repository
- Clone your fork locally
- Push to your fork, open PRs against upstream

## Workflow Selection Guide

| Criteria | Centralized | Feature Branch | Gitflow | Forking |
|---|---|---|---|---|
| Team size | 1–3 | 2–10 | 5+ | Any (open source) |
| Release cadence | Continuous | Continuous | Scheduled | Varies |
| Branch complexity | Minimal | Low | High | Medium |
| Code review | Optional | Required | Required | Required |
| Best for | Small teams, simple projects | Most teams | Versioned releases | Open source, external contributors |

## Capabilities

| Capability | Action | Description |
|------------|--------|-------------|
| Initialize Repository | `actions/initialize-repository.md` | Clone, set up remotes, or fork |
| Create Feature Branch | `actions/create-feature-branch.md` | Create and push a feature branch |
| Commit Changes | `actions/commit-changes.md` | Stage and commit with conventional commits |
| Sync with Upstream | `actions/sync-with-upstream.md` | Pull and rebase to stay current |
| Create Pull Request | `actions/create-pull-request.md` | Open a PR for code review |
| Resolve Merge Conflicts | `actions/resolve-merge-conflicts.md` | Handle conflicts during rebase or merge |
| Merge Pull Request | `actions/merge-pull-request.md` | Merge an approved PR into the target branch |
| Create Release | `actions/create-release.md` | Create release branches and tags (Gitflow) |
| Create Hotfix | `actions/create-hotfix.md` | Create and merge hotfix branches (Gitflow) |

## Standards

| Standard | File | Description |
|----------|------|-------------|
| Branch Naming | `standards/branch-naming.md` | Branch prefix and naming conventions |
| Commit Messages | `standards/commit-messages.md` | Conventional commit format and rules |
| Code Review | `standards/code-review.md` | Review checklist and guidelines |
| Merge Strategy | `standards/merge-strategy.md` | When to rebase vs merge |
| Checklist | `standards/checklist.md` | Consolidated compliance checklist |

## Git Safety Protocol

- NEVER update git config without explicit request
- NEVER run destructive commands (`--force`, `hard reset`) without explicit request
- NEVER skip hooks (`--no-verify`) unless user asks
- NEVER force push to `main`/`master`/`develop`
- If a commit fails due to hooks, fix the issue and create a NEW commit (do not amend)