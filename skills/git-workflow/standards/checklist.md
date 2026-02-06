# Git Workflow Compliance Checklist

## Overview

Consolidated checklist for validating git workflow compliance across all actions.

---

## Repository Setup

- [ ] Repository cloned or forked appropriately for the workflow model
- [ ] Remotes configured correctly (`origin`, and `upstream` for forking workflow)
- [ ] Default branch identified (`main` or `develop` for Gitflow)
- [ ] `develop` branch exists (Gitflow only)

---

## Branch Management

- [ ] Branch created from correct base branch
- [ ] Branch name follows [Branch Naming](branch-naming.md) standard
- [ ] Branch name uses correct prefix (`feature/`, `bugfix/`, `hotfix/`, `release/`, etc.)
- [ ] Branch name uses kebab-case (lowercase, hyphens)
- [ ] Ticket number included when available

---

## Commits

- [ ] Each commit represents one logical change
- [ ] Commit message follows [Commit Messages](commit-messages.md) standard
- [ ] Commit message uses correct type (`feat`, `fix`, `docs`, etc.)
- [ ] Commit message description is present tense, imperative mood
- [ ] Commit message description under 72 characters
- [ ] Commits cryptographically signed (`commit.gpgsign=true` or `-S` flag)
- [ ] Commits include sign-off (`-s` flag) for DCO compliance
- [ ] No secrets, credentials, or tokens committed

---

## Syncing

- [ ] Branch synced with target branch before opening PR
- [ ] Rebase used for syncing (not merge, unless conflicts are complex)
- [ ] `--force-with-lease` used instead of `--force` when force pushing
- [ ] Never force pushed to `main`, `master`, or `develop`

---

## Pull Requests

- [ ] PR title follows conventional commit format
- [ ] PR description explains what and why
- [ ] Related issues referenced (`Closes #123`, `Refs #456`)
- [ ] At least one reviewer assigned
- [ ] All CI checks pass
- [ ] No unresolved conflicts
- [ ] Branch up to date with target branch

---

## Code Review

- [ ] Reviewed within one business day
- [ ] Comments use prefixes (`blocker:`, `suggestion:`, `nit:`, `question:`, `praise:`)
- [ ] All blockers resolved before approval
- [ ] Author responded to all comments

---

## Merging

- [ ] PR approved by required reviewers
- [ ] Correct merge method selected per [Merge Strategy](merge-strategy.md)
- [ ] Feature branch deleted after merge
- [ ] Local repository updated after merge

---

## Releases (Gitflow)

- [ ] All intended features merged into `develop`
- [ ] Release branch created from `develop`
- [ ] Version numbers bumped
- [ ] Changelog updated
- [ ] Release branch merged into `main`
- [ ] Release tagged with `v<version>`
- [ ] Release branch back-merged into `develop`
- [ ] Release branch deleted

---

## Hotfixes (Gitflow)

- [ ] Hotfix branch created from `main`
- [ ] Minimum fix applied
- [ ] Patch version bumped
- [ ] Hotfix merged into `main`
- [ ] Hotfix tagged with `v<version>`
- [ ] Hotfix back-merged into `develop` (or active release branch)
- [ ] Hotfix branch deleted

---

## Safety

- [ ] No `--force` pushes to protected branches
- [ ] No `git reset --hard` without explicit user request
- [ ] No `--no-verify` to skip hooks without explicit user request
- [ ] No git config changes without explicit user request
