# Submit Pull Request

Open the pull request, assign reviewers, add labels, and link to project boards. This is the step where your prepared changes become visible to the team.

## Prerequisites

- Changes are prepared (see [Prepare Pull Request](prepare-pull-request.md))
- PR title and description are written (see [Write PR Description](write-pr-description.md))
- Branch is pushed to the remote

## Steps

### 1. Push Your Branch

If you haven't already, push your branch to the remote:

```bash
git push -u origin <branch-name>
```

### 2. Open the Pull Request 🛑

**🛑 STOP**: Confirm with the user whether to open as a **draft** (for early feedback/WIP) or **ready for review** (complete and tested).

#### Using GitHub CLI

```bash
gh pr create \
  --base <target-branch> \
  --title "<type>[scope]: <description>" \
  --body "<PR description>" \
  --reviewer <reviewer1>,<reviewer2> \
  --label <label1>,<label2>
```

#### Using a Draft PR

If your work is not yet ready for full review but you want early feedback or visibility:

```bash
gh pr create \
  --base <target-branch> \
  --title "<type>[scope]: <description>" \
  --body "<PR description>" \
  --draft
```

Convert to ready when preparation is complete:

```bash
gh pr ready <pr-number>
```

#### Using the Web UI

1. Navigate to the repository on GitHub
2. Click **New Pull Request** or use the prompt on the branch
3. Select your feature branch as the source
4. Select the target branch as the base
5. Fill in the title and description using the [Pull Request Template](../standards/pull-request-template.md)
6. Click **Create Pull Request** (or **Create Draft Pull Request**)

### 3. Determine the Target Branch

| Workflow | Target branch |
|---|---|
| Feature Branch | `main` |
| Gitflow (feature) | `develop` |
| Gitflow (release) | `main` |
| Gitflow (hotfix) | `main` and `develop` |
| Forking | `upstream/main` |

### 4. Assign Reviewers

- Assign at least **one reviewer** who is familiar with the affected area
- For critical changes, assign **two or more reviewers**
- Tag the team or code owners if configured (via `CODEOWNERS` file)
- Avoid assigning too many reviewers — it diffuses responsibility

```bash
# Add reviewers after creation
gh pr edit <pr-number> --add-reviewer <username1>,<username2>
```

### 5. Add Labels

Use labels to communicate the status and type of your PR:

| Label | Purpose |
|---|---|
| `ready-for-review` | PR is complete and ready for review |
| `work-in-progress` | PR is still being worked on |
| `blocked` | PR is blocked by an external dependency |
| `needs-discussion` | PR requires team discussion before proceeding |
| `breaking-change` | PR introduces a breaking change |

```bash
gh pr edit <pr-number> --add-label "ready-for-review"
```

### 6. Link to Issues and Projects

Link related issues so they auto-close when the PR merges:

- Add `Closes #<issue-number>` in the PR description
- Or link manually via the GitHub sidebar

Connect to project boards for tracking:

```bash
# Link to a project (if using GitHub Projects)
gh pr edit <pr-number> --add-project "<project-name>"
```

### 7. Verify CI Status

After submitting, monitor the CI checks:

```bash
# Check PR status
gh pr checks <pr-number>
```

If checks fail, fix the issues and push new commits. Do not wait for reviewers to point out CI failures.

## Post-Submission Checklist

- [ ] PR is open against the correct target branch
- [ ] Reviewers are assigned
- [ ] Labels reflect the current status
- [ ] Related issues are linked
- [ ] CI checks are passing
- [ ] PR is not a draft (unless intentionally so)
