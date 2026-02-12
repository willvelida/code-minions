# Sync with Upstream

## Purpose

Keep your local branch up to date with the remote to minimize merge conflicts and ensure you are building on the latest code.

---

## Flow

### Step 1: Choose Sync Method

| Workflow | Remote | Command |
|---|---|---|
| Centralized | `origin` | `git pull --rebase origin main` |
| Feature Branch | `origin` | `git pull --rebase origin main` |
| Gitflow | `origin` | `git pull --rebase origin develop` |
| Forking | `upstream` | `git fetch upstream && git rebase upstream/main` |

**Success Criteria:**
- [ ] Correct sync method identified for the workflow

---

### Step 2: Fetch Latest Changes

```bash
# Standard workflows
git fetch origin

# Forking workflow — also fetch upstream
git fetch upstream
```

**Success Criteria:**
- [ ] Latest changes fetched from remote

---

### Step 3: Rebase onto the Base Branch

Rebase keeps the commit history linear and avoids unnecessary merge commits.

```bash
# From your feature branch — Feature Branch workflow
git rebase origin/main

# From your feature branch — Gitflow workflow
git rebase origin/develop

# From your feature branch — Forking workflow
git rebase upstream/main
```

**Success Criteria:**
- [ ] Rebase completed successfully
- [ ] If conflicts occur, see [Resolve Merge Conflicts](resolve-merge-conflicts.md)

---

### Step 4: Force Push After Rebase (Feature Branches Only) 🛑

Rebasing rewrites history, so a force push is needed for branches that have already been pushed:

```bash
git push --force-with-lease
```

Use `--force-with-lease` instead of `--force` to prevent overwriting changes pushed by others.

**🛑 STOP**: Confirm the branch is a feature branch. **NEVER force push to `main`, `master`, or `develop`.**

**Success Criteria:**
- [ ] Force push only to a feature branch
- [ ] Used `--force-with-lease` (not `--force`)

---

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
