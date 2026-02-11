# Commit Changes

## Purpose

Stage and commit changes using the Conventional Commits specification. Each commit should represent one logical change.

---

## Flow

### Step 1: Review Changes

```bash
# Check which files have changed
git status --porcelain

# Review the full diff
git diff
```

**Success Criteria:**
- [ ] Changes reviewed and understood

---

### Step 2: Stage Files

Stage only the files related to one logical change:

```bash
# Stage specific files
git add path/to/file1 path/to/file2

# Stage by pattern
git add src/components/*
git add *.test.*

# Stage all changes (use with caution)
git add .

# Interactive staging for partial file changes
git add -p
```

**Never stage secrets**: `.env`, credentials, private keys, or tokens.

**Success Criteria:**
- [ ] Only related files staged
- [ ] No secrets or sensitive files staged

---

### Step 3: Verify Staged Changes 🛑

```bash
git diff --staged
```

**🛑 STOP**: Review the staged diff. Confirm all staged changes are intentional and belong to one logical change.

**Success Criteria:**
- [ ] Staged diff reviewed
- [ ] All changes belong to one logical unit

---

### Step 4: Generate the Commit Message

Analyze the staged diff to determine:

- **Type**: What kind of change (see [Commit Messages standard](../standards/commit-messages.md))
- **Scope**: What module or area is affected
- **Description**: One-line summary in present tense, imperative mood, under 72 characters

**Success Criteria:**
- [ ] Type correctly identified
- [ ] Description is present tense, imperative mood, under 72 characters

---

### Step 5: Execute the Commit

**Important:** All commits must be cryptographically signed. Ensure signing is configured:

```bash
# Verify signing is enabled (should return "true")
git config --global commit.gpgsign
```

If not configured, see [Commit Messages standard](../standards/commit-messages.md#signed-commits).

```bash
# Single-line commit (with sign-off)
git commit -s -m "<type>[scope]: <description>"

# Multi-line commit with body and/or footer
git commit -s -m "<type>[scope]: <description>

<optional body explaining the what and why>

<optional footer: Closes #123, BREAKING CHANGE, etc.>"
```

**Note:** With `commit.gpgsign=true` set globally, commits are automatically cryptographically signed. The `-s` flag adds the DCO sign-off line.

**Success Criteria:**
- [ ] Commit message follows conventional commit format
- [ ] Commit is signed and has sign-off
- [ ] No hook errors

---

### Step 6: Push the Commit

```bash
git push
```

**Success Criteria:**
- [ ] Commit pushed to remote

---

## Commit Message Quick Reference

```
feat(auth): add JWT token validation middleware
fix(api): correct null check in user lookup
docs(readme): update installation instructions
refactor(db): extract connection pool into module
test(auth): add unit tests for login endpoint
```

## Rules

- One logical change per commit
- Present tense: "add" not "added"
- Imperative mood: "fix bug" not "fixes bug"
- Description under 72 characters
- Reference issues when applicable: `Closes #123`, `Refs #456`
- **Cryptographic signature required** (`commit.gpgsign=true` or `-S` flag)
- Sign-off with `-s` flag for DCO compliance

See [Commit Messages](../standards/commit-messages.md) for the full standard.
