# Resolve Merge Conflicts

## Purpose

Handle conflicts that arise during rebase or merge operations. Conflicts occur when changes on different branches modify the same lines of code.

---

## Flow

### Step 1: Identify Conflicted Files

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

**Success Criteria:**
- [ ] All conflicted files identified

---

### Step 2: Open and Resolve Each Conflict

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

Choose one of:

- **Accept current** — keep your version
- **Accept incoming** — keep the other branch's version
- **Accept both** — combine both changes
- **Rewrite** — write a new version that incorporates both intents

Remove all conflict markers (`<<<<<<<`, `=======`, `>>>>>>>`).

**Success Criteria:**
- [ ] All conflict markers removed
- [ ] Resolved code is correct and compiles

---

### Step 3: Stage the Resolved Files

```bash
git add <resolved-file>
```

**Success Criteria:**
- [ ] All resolved files staged

---

### Step 4: Continue the Operation 🛑

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

**🛑 STOP**: If conflicts are too complex or you need to start over, confirm with the user before aborting:

```bash
# Abort a rebase
git rebase --abort

# Abort a merge
git merge --abort
```

This restores the branch to its state before the operation began.

**Success Criteria:**
- [ ] Rebase or merge operation completed
- [ ] No remaining conflicts

---

## Tips

- Sync frequently to reduce the size and frequency of conflicts (see [Sync with Upstream](sync-with-upstream.md))
- Keep branches short-lived — the longer a branch lives, the more likely conflicts become
- Resolve conflicts commit-by-commit during rebase for smaller, more manageable changes
- Use `git log --oneline --graph` to visualize diverged history
- When in doubt, communicate with the author of the conflicting changes
