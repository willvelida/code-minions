# Prepare Pull Request

Scope your changes, self-review your diff, and verify readiness before opening a pull request. Preparation catches errors early and reduces review cycles.

## Prerequisites

- Feature branch with committed changes (see [git-workflow: Create Feature Branch](../../git-workflow/actions/create-feature-branch.md))
- All commits follow the [Commit Messages](../../git-workflow/standards/commit-messages.md) standard
- Working tree is clean (`git status` shows no uncommitted changes)

## Steps

### 1. Verify PR Scope 🛑

Each PR should represent **one logical change**. Before opening, ask:

- Does this PR do exactly one thing?
- Can any part of this change be split into a separate PR?
- Are there unrelated changes mixed in (formatting, refactoring, dependency updates)?

If the answer to any of these is yes, split the work into separate branches and PRs.

**🛑 STOP**: If splitting is needed, confirm the split strategy with the user before proceeding.

See [PR Size](../standards/pr-size.md) for detailed sizing guidelines.

### 2. Sync with the Target Branch

Ensure your branch is up to date to avoid conflicts during review:

```bash
git fetch origin
git rebase origin/<target-branch>
```

Resolve any conflicts before proceeding (see [git-workflow: Resolve Merge Conflicts](../../git-workflow/actions/resolve-merge-conflicts.md)).

### 3. Self-Review Your Diff

Review your own changes as if you were the reviewer:

```bash
# View the full diff against the target branch
git diff origin/<target-branch>...HEAD

# View a summary of changed files
git diff origin/<target-branch>...HEAD --stat
```

While reviewing, check for:

- Leftover debug statements (`console.log`, `print`, `debugger`)
- Commented-out code that should be removed
- TODO comments that should be resolved
- Hardcoded values that should be configuration
- Missing error handling or edge cases
- Typos in code, comments, and documentation

See [Self-Review Checklist](../standards/self-review-checklist.md) for the complete checklist.

### 4. Review for Security

Before submitting, check for security issues:

- No secrets, tokens, API keys, or credentials in the diff
- No new vulnerable dependencies introduced
- Input validation is present where needed
- Authentication and authorization checks are correct

```bash
# Search for common secret patterns in your changes
git diff origin/<target-branch>...HEAD | grep -iE "(password|secret|token|api_key|apikey|private_key)"
```

### 5. Run Tests Locally

Verify your changes pass all tests before opening the PR:

```bash
# Run the project's test suite (adjust to your project)
npm test        # Node.js
pytest          # Python
dotnet test     # .NET
go test ./...   # Go
```

### 6. Build Locally

Confirm the project builds without errors:

```bash
# Build the project (adjust to your project)
npm run build       # Node.js
python -m build     # Python
dotnet build        # .NET
go build ./...      # Go
```

### 7. Verify CI Will Pass

If your project has a CI configuration, review what checks will run and make sure your changes satisfy them:

- Linting rules pass
- Test coverage thresholds are met
- Build succeeds
- Security scans pass

## Readiness Checklist

Before moving to [Write PR Description](write-pr-description.md), confirm:

- [ ] PR represents one logical change
- [ ] Branch is synced with the target branch
- [ ] Self-review complete — no debug code, no secrets, no typos
- [ ] Tests pass locally
- [ ] Build succeeds locally
- [ ] No unresolved merge conflicts
