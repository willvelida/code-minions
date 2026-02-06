# Create Feature Branch

Create a dedicated branch for a new feature, bugfix, or task. All development happens in isolation from the main integration branch.

## Prerequisites

- Repository initialized (see [Initialize Repository](initialize-repository.md))
- Working tree is clean (`git status` shows no uncommitted changes)

## Determine Base Branch

| Workflow | Base branch |
|---|---|
| Feature Branch | `main` |
| Gitflow | `develop` |
| Forking | `main` (of your fork, synced with upstream) |

## Steps

### 1. Switch to the Base Branch

```bash
# Feature Branch / Forking workflow
git checkout main

# Gitflow workflow
git checkout develop
```

### 2. Pull Latest Changes

```bash
git pull origin <base-branch>

# Forking workflow — also sync with upstream
git fetch upstream
git rebase upstream/main
```

### 3. Create the Feature Branch

Follow the [Branch Naming](../standards/branch-naming.md) standard.

```bash
git checkout -b <branch-name>
```

Examples:

```bash
git checkout -b feature/user-authentication
git checkout -b bugfix/fix-login-redirect
git checkout -b chore/update-dependencies
```

### 4. Push the Branch to Remote

```bash
git push -u origin <branch-name>
```

The `-u` flag sets up tracking so future `git push` and `git pull` work without specifying the remote and branch.

## Verification

```bash
# Confirm you are on the new branch
git branch --show-current

# Confirm remote tracking is set
git branch -vv
```
