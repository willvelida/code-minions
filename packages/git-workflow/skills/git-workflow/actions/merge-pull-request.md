# Merge Pull Request

## Purpose

Merge an approved pull request into the target branch. This is the final step to integrate a feature, fix, or change into the codebase.

---

## Flow

### Step 1: Verify Merge Readiness

Confirm:

- Pull request has been reviewed and approved (see [Code Review](../standards/code-review.md))
- All CI checks pass
- No unresolved conflicts
- Branch is up to date with the target branch

**Success Criteria:**
- [ ] PR approved by required reviewers
- [ ] All CI checks pass
- [ ] No unresolved conflicts

---

### Step 2: Choose Merge Method 🛑

See [Merge Strategy](../standards/merge-strategy.md) for detailed guidance.

| Method | When to use |
|---|---|
| Squash and merge | Default for most feature branches — creates one clean commit |
| Rebase and merge | When each commit is meaningful and atomic |
| Merge commit | When you want to preserve the full branch history |

**🛑 STOP**: Confirm the merge method with the user before proceeding. Default to "Squash and merge" if no preference given.

**Success Criteria:**
- [ ] Merge method confirmed

---

### Step 3: Execute the Merge

#### Option A: Squash and Merge (Recommended)

```bash
# Using GitHub CLI
gh pr merge <pr-number> --squash --delete-branch

# Using Git manually
git checkout <target-branch>
git merge --squash <feature-branch>
git commit -s -m "<type>[scope]: <description>

Squashed commits from <feature-branch>.

Closes #<issue-number>"
git push origin <target-branch>
```

#### Option B: Rebase and Merge

```bash
# Using GitHub CLI
gh pr merge <pr-number> --rebase --delete-branch

# Using Git manually
git checkout <target-branch>
git rebase <feature-branch>
git push origin <target-branch>
```

#### Option C: Merge Commit

```bash
# Using GitHub CLI
gh pr merge <pr-number> --merge --delete-branch

# Using Git manually
git checkout <target-branch>
git merge --no-ff <feature-branch>
git push origin <target-branch>
```

**Success Criteria:**
- [ ] Merge completed successfully
- [ ] Changes visible on the target branch

---

### Step 4: Post-Merge Cleanup

```bash
# Delete remote branch
git push origin --delete <feature-branch>

# Delete local branch
git branch -d <feature-branch>

# Update local repository
git checkout <target-branch>
git pull origin <target-branch>
```

**Success Criteria:**
- [ ] Feature branch deleted (remote and local)
- [ ] Local repository updated

---

### Step 5: Notify the Team

If the merge introduces changes others should know about (API changes, dependency updates, breaking changes), communicate via the team's agreed channel.

**Success Criteria:**
- [ ] Team notified of impactful changes (if applicable)

---

## Gitflow-Specific Merges

For Gitflow, some merges target multiple branches:

### Release Branch Merge

```bash
# Merge release into main
git checkout main
git merge --no-ff release/<version>
git tag -a v<version> -m "Release v<version>"
git push origin main --tags

# Back-merge into develop
git checkout develop
git merge --no-ff release/<version>
git push origin develop

# Delete the release branch
git branch -d release/<version>
git push origin --delete release/<version>
```

### Hotfix Branch Merge

See [Create Hotfix](create-hotfix.md) for the complete hotfix workflow.
