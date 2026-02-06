# Resolve Merge Conflicts

Handle conflicts that arise during rebase or merge operations. Conflicts occur when changes on different branches modify the same lines of code.

## Prerequisites

- A rebase or merge operation is in progress and has paused due to conflicts
- Or a PR has been flagged with conflicts by the hosting platform

## Identify Conflicted Files

```bash
git status
```

Conflicted files appear under "Unmerged paths":

```
Unmerged paths:
  (use "git add <file>..." to mark resolution)
    both modified:   src/auth/handler.ts
    both modified:   src/config.ts
```

## Resolve Conflicts

### 1. Open Conflicted Files

Conflict markers in the file look like:

```
<<<<<<< HEAD
// Your changes (current branch)
const timeout = 5000;
=======
// Incoming changes (target branch)
const timeout = 3000;
>>>>>>> origin/main
```

### 2. Edit to Resolve

Choose one of:

- **Accept current** — keep your version
- **Accept incoming** — keep the other branch's version
- **Accept both** — combine both changes
- **Rewrite** — write a new version that incorporates both intents

Remove all conflict markers (`<<<<<<<`, `=======`, `>>>>>>>`).

### 3. Stage the Resolved Files

```bash
git add <resolved-file>
```

### 4. Continue the Operation

#### If Rebasing

```bash
git rebase --continue
```

If more conflicts appear on subsequent commits, repeat steps 1–3.

#### If Merging

```bash
git commit
```

Git opens the editor with a pre-filled merge commit message.

## Abort if Needed

If the conflict is too complex or you need to start over:

```bash
# Abort a rebase
git rebase --abort

# Abort a merge
git merge --abort
```

This restores the branch to its state before the operation began.

## Tips

- Sync frequently to reduce the size and frequency of conflicts (see [Sync with Upstream](sync-with-upstream.md))
- Keep branches short-lived — the longer a branch lives, the more likely conflicts become
- Resolve conflicts commit-by-commit during rebase for smaller, more manageable changes
- Use `git log --oneline --graph` to visualize diverged history
- When in doubt, communicate with the author of the conflicting changes
