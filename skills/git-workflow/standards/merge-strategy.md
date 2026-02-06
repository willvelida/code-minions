# Merge Strategy Standard

Guidance on which merge method to use and when. The right strategy keeps the commit history clean and meaningful.

## Merge Methods

### Squash and Merge

Combines all commits from the feature branch into a single commit on the target branch.

```
main:  A --- B --- C --- S (squashed)
```

**When to use:**
- Feature branch has many small or "work-in-progress" commits
- You want one clean commit per PR in the target branch
- Default choice for most feature branches

**Pros:**
- Clean, linear history on `main`
- Easy to revert an entire feature (one commit)
- Noisy branch history does not pollute the main log

**Cons:**
- Individual commit granularity is lost
- Co-authors may not be reflected (add `Co-authored-by` trailers)

### Rebase and Merge

Replays each commit from the feature branch onto the tip of the target branch.

```
main:  A --- B --- C --- D1 --- D2 --- D3 (rebased)
```

**When to use:**
- Each commit on the branch is meaningful and atomic
- You want a linear history with full commit detail
- Branch has been kept clean through interactive rebasing

**Pros:**
- Linear history with individual commits preserved
- Full granularity of changes
- Easy to bisect

**Cons:**
- Messy if the branch has "fixup" or "WIP" commits
- Requires disciplined committing

### Merge Commit (No Fast-Forward)

Creates a merge commit that ties the two branches together.

```
main:  A --- B --- C --------- M (merge commit)
                \             /
feature:         D1 --- D2 ---
```

**When to use:**
- You want to preserve the branch topology
- Required for Gitflow release and hotfix branches
- Team convention prefers explicit merge points

**Pros:**
- Full history of branch and merge points
- Clear record of when features were integrated
- Easy to see the "shape" of development

**Cons:**
- History can become cluttered with merge commits
- Harder to read in a flat log

## Recommended Defaults

| Workflow | Default method |
|---|---|
| Feature Branch | Squash and merge |
| Gitflow (feature → develop) | Squash and merge |
| Gitflow (release → main) | Merge commit (`--no-ff`) |
| Gitflow (hotfix → main) | Merge commit (`--no-ff`) |
| Gitflow (back-merge → develop) | Merge commit (`--no-ff`) |
| Forking | Squash and merge |

## Rebase vs Merge for Syncing

When keeping a feature branch up to date with the target branch:

| Approach | Command | When to use |
|---|---|---|
| Rebase | `git rebase origin/main` | Default — keeps history linear |
| Merge | `git merge origin/main` | When rebase causes too many conflicts on a long-lived branch |

**Prefer rebase** for syncing. Use merge only as a fallback for complex conflict scenarios.

## Force Push Policy

After rebasing a branch that has already been pushed:

```bash
git push --force-with-lease
```

- **Always use `--force-with-lease`** instead of `--force` to prevent overwriting others' work
- **NEVER force push** to `main`, `master`, or `develop`
- Force push is acceptable on feature branches that only you are working on
- Communicate with collaborators before force pushing a shared branch
