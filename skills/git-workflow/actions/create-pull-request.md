# Create Pull Request

Open a pull request (PR) to propose merging your feature branch into the target branch. PRs enable code review and discussion before integration.

## Prerequisites

- Feature branch with committed changes pushed to remote
- Branch is synced with the target branch (see [Sync with Upstream](sync-with-upstream.md))
- All commits follow the [Commit Messages](../standards/commit-messages.md) standard

## Determine Target Branch

| Workflow | Target branch |
|---|---|
| Feature Branch | `main` |
| Gitflow (feature) | `develop` |
| Gitflow (release) | `main` |
| Gitflow (hotfix) | `main` and `develop` |
| Forking | `upstream/main` |

## Steps

### 1. Sync Before Opening

```bash
git fetch origin
git rebase origin/<target-branch>
git push --force-with-lease
```

### 2. Open the Pull Request

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

### 3. PR Title Format

Follow the same format as commit messages:

```
<type>[scope]: <description>
```

Examples:

```
feat(auth): add JWT token validation
fix(api): correct null pointer in user service
docs(readme): update deployment instructions
```

### 4. Assign Reviewers

- Assign at least one reviewer who is familiar with the affected area
- For critical changes, assign two or more reviewers
- Tag the team or code owners if configured

### 5. Address Review Feedback

When reviewers request changes:

```bash
# Make the requested changes
git add <files>
git commit -s -m "fix(scope): address review feedback"
git push
```

Avoid force-pushing during review unless explicitly asked — it makes it harder for reviewers to track incremental changes.

## PR Checklist

Before submitting a PR, verify:

- [ ] Branch is up to date with the target branch
- [ ] All commits follow the conventional commit format
- [ ] Tests pass locally
- [ ] No secrets or sensitive data in the diff
- [ ] PR description explains what and why
- [ ] Related issues are referenced

See [Code Review](../standards/code-review.md) for the full review standard.
