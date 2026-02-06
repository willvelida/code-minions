# PR Size Standard

Keep pull requests small and focused. Small PRs are easier to review, faster to merge, less likely to introduce bugs, and provide a clearer history of changes.

## Size Guidelines

| Metric | Target | Maximum |
|---|---|---|
| Lines changed | < 200 | 400 |
| Files changed | < 10 | 20 |
| Logical changes | 1 | 1 |

These are guidelines, not hard rules. A 500-line PR that adds a single well-defined feature may be fine. A 50-line PR that mixes a bug fix with a refactor should be split.

## One Logical Change Per PR

Each PR should represent **one logical change**. A logical change is one of:

- A single feature or user story
- A single bug fix
- A single refactoring
- A single dependency update
- A single documentation update

### Do Not Mix

| ❌ Don't mix | ✅ Split into separate PRs |
|---|---|
| Feature + formatting fixes | PR 1: formatting, PR 2: feature |
| Bug fix + refactor | PR 1: bug fix, PR 2: refactor |
| Dependency update + code changes | PR 1: dependency update, PR 2: code changes |
| Multiple unrelated bug fixes | One PR per bug fix |

## When Large PRs Are Unavoidable

Some changes are inherently large (e.g., initial project setup, large migrations, generated code). In these cases:

1. **Add a reviewer guide** — Explain the review order and what to focus on
2. **Break into reviewable chunks** — Use multiple commits with clear messages so reviewers can go commit-by-commit
3. **Mark generated code** — Clearly identify auto-generated files so reviewers can skip them
4. **Consider a stacked PR approach** — Break the work into a chain of dependent PRs, each building on the last

## Stacked PRs

For large features, use stacked (chained) PRs:

```
main
 └── feature/auth-part-1 (PR #1: database schema)
      └── feature/auth-part-2 (PR #2: API endpoints)
           └── feature/auth-part-3 (PR #3: UI integration)
```

Each PR is small and reviewable on its own. Merge them in order, rebasing the next PR onto the updated target branch.

## Checking PR Size

Before opening a PR, check the size of your diff:

```bash
# Count lines changed against the target branch
git diff origin/<target-branch>...HEAD --stat

# Count total added/removed lines
git diff origin/<target-branch>...HEAD --shortstat
```

If the diff is too large, consider splitting the branch before opening the PR.

## Why Small PRs Matter

- **Faster reviews** — Reviewers can focus and provide better feedback
- **Fewer bugs** — Smaller changes are easier to reason about
- **Easier rollbacks** — If something breaks, the blast radius is small
- **Better history** — Each merge commit tells a clear story
- **Less merge conflict risk** — Short-lived branches diverge less from the target
