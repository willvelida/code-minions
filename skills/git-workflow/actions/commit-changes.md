# Commit Changes

Stage and commit changes using the Conventional Commits specification. Each commit should represent one logical change.

## Prerequisites

- On a feature branch (see [Create Feature Branch](create-feature-branch.md))
- Changes exist in the working tree

## Steps

### 1. Review Changes

```bash
# Check which files have changed
git status --porcelain

# Review the full diff
git diff
```

### 2. Stage Files

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

### 3. Verify Staged Changes

```bash
git diff --staged
```

### 4. Generate the Commit Message

Analyze the staged diff to determine:

- **Type**: What kind of change (see [Commit Messages standard](../standards/commit-messages.md))
- **Scope**: What module or area is affected
- **Description**: One-line summary in present tense, imperative mood, under 72 characters

### 5. Execute the Commit

```bash
# Single-line commit
git commit -s -m "<type>[scope]: <description>"

# Multi-line commit with body and/or footer
git commit -s -m "<type>[scope]: <description>

<optional body explaining the what and why>

<optional footer: Closes #123, BREAKING CHANGE, etc.>"
```

### 6. Push the Commit

```bash
git push
```

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
- Sign commits with `-s` flag for DCO compliance

See [Commit Messages](../standards/commit-messages.md) for the full standard.
