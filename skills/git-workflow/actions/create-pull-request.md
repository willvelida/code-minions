# Create Pull Request

## Purpose

Open a pull request (PR) to propose merging your feature branch into the target branch. PRs enable code review and discussion before integration.

---

## Flow

### Step 1: Determine Target Branch

| Workflow | Target branch |
|---|---|
| Feature Branch | `main` |
| Gitflow (feature) | `develop` |
| Gitflow (release) | `main` |
| Gitflow (hotfix) | `main` and `develop` |
| Forking | `upstream/main` |

**Success Criteria:**
- [ ] Correct target branch identified

---

### Step 2: Sync Before Opening

```bash
git fetch origin
git rebase origin/<target-branch>
git push --force-with-lease
```

**Success Criteria:**
- [ ] Branch is up to date with the target branch
- [ ] All commits follow the [Commit Messages](../standards/commit-messages.md) standard

---

### Step 3: Open the Pull Request 🛑

#### Using GitHub CLI

```bash
gh pr create \
  --base <target-branch> \
  --title "<type>[scope]: <description>" \
  --body "## Summary
<What this PR does and why>

## Changes
- <Change 1>
- <Change 2>

## Testing
- <How this was tested>

## Related Issues
Closes #<issue-number>"
```

#### Using the Web UI

1. Navigate to the repository on the hosting platform
2. Click "New Pull Request" or "Create Pull Request"
3. Select your feature branch as the source
4. Select the target branch as the base
5. Fill in the title and description
6. Assign reviewers
7. Submit

**🛑 STOP**: Confirm the PR title, description, and target branch with the user before submitting.

**Success Criteria:**
- [ ] PR title follows conventional commit format
- [ ] PR description explains what and why
- [ ] Related issues referenced

---

### Step 4: Assign Reviewers

- Assign at least one reviewer who is familiar with the affected area
- For critical changes, assign two or more reviewers
- Tag the team or code owners if configured

**Success Criteria:**
- [ ] At least one reviewer assigned

---

### Step 5: Address Review Feedback

When reviewers request changes:

```bash
# Make the requested changes
git add <files>
git commit -s -m "fix(scope): address review feedback"
git push
```

Avoid force-pushing during review unless explicitly asked — it makes it harder for reviewers to track incremental changes.

**Success Criteria:**
- [ ] All review comments addressed
- [ ] Changes pushed as new commits

---

## PR Checklist

Before submitting a PR, verify:

- [ ] Branch is up to date with the target branch
- [ ] All commits follow the conventional commit format
- [ ] Tests pass locally
- [ ] No secrets or sensitive data in the diff
- [ ] PR description explains what and why
- [ ] Related issues are referenced

See [Code Review](../standards/code-review.md) for the full review standard.
