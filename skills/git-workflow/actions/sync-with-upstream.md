# Sync with Upstream

Keep your local branch up to date with the remote to minimize merge conflicts and ensure you are building on the latest code.

## Prerequisites

- Repository initialized with remotes configured (see [Initialize Repository](initialize-repository.md))
- On a feature branch or base branch

## Choose Sync Method

| Workflow | Remote | Command |
|---|---|---|
| Centralized | `origin` | `git pull --rebase origin main` |
| Feature Branch | `origin` | `git pull --rebase origin main` |
| Gitflow | `origin` | `git pull --rebase origin develop` |
| Forking | `upstream` | `git fetch upstream && git rebase upstream/main` |

## Steps

### 1. Fetch Latest Changes

```bash
# Standard workflows
git fetch origin

# Forking workflow — also fetch upstream
git fetch upstream
```

### 2. Rebase onto the Base Branch

Rebase keeps the commit history linear and avoids unnecessary merge commits.

```bash
# From your feature branch — Feature Branch workflow
git rebase origin/main

# From your feature branch — Gitflow workflow
git rebase origin/develop

# From your feature branch — Forking workflow
git rebase upstream/main
```

### 3. Handle Conflicts (if any)

If conflicts occur during rebase, see [Resolve Merge Conflicts](resolve-merge-conflicts.md).

### 4. Force Push After Rebase (feature branches only)

Rebasing rewrites history, so a force push is needed for branches that have already been pushed:

```bash
git push --force-with-lease
```

Use `--force-with-lease` instead of `--force` to prevent overwriting changes pushed by others.

**NEVER force push to `main`, `master`, or `develop`.**

## Sync Your Fork's Default Branch (Forking Workflow)

To keep your fork's `main` in sync with upstream:

```bash
git checkout main
git fetch upstream
git rebase upstream/main
git push origin main
```

## When to Sync

- Before creating a new feature branch
- Before opening a pull request
- When your PR has conflicts with the target branch
- Regularly during long-running feature branches (at least daily)
