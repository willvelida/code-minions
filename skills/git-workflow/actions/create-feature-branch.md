# Create Feature Branch

## Purpose

Create a dedicated branch for a new feature, bugfix, or task. All development happens in isolation from the main integration branch.

---

## Flow

### Step 1: Determine Base Branch

| Workflow | Base branch |
|---|---|
| Feature Branch | `main` |
| Gitflow | `develop` |
| Forking | `main` (of your fork, synced with upstream) |

**Success Criteria:**
- [ ] Correct base branch identified for the workflow model

---

### Step 2: Switch to Base Branch and Pull Latest

```bash
# Feature Branch / Forking workflow
git checkout main

# Gitflow workflow
git checkout develop

# Pull latest changes
git pull origin <base-branch>

# Forking workflow — also sync with upstream
git fetch upstream
git rebase upstream/main
```

**Success Criteria:**
- [ ] On the correct base branch
- [ ] Latest changes pulled

---

### Step 3: Create the Branch

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

**Success Criteria:**
- [ ] Branch name follows naming standard (prefix, kebab-case)
- [ ] Branch created from correct base

---

### Step 4: Push the Branch to Remote

```bash
git push -u origin <branch-name>
```

The `-u` flag sets up tracking so future `git push` and `git pull` work without specifying the remote and branch.

**Success Criteria:**
- [ ] Branch pushed to remote
- [ ] Tracking configured

---

### Step 5: Verify

```bash
# Confirm you are on the new branch
git branch --show-current

# Confirm remote tracking is set
git branch -vv
```

**Success Criteria:**
- [ ] On the new branch
- [ ] Remote tracking confirmed
