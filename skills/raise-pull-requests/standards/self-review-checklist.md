# Self-Review Checklist

Review your own pull request before requesting review from others. Self-review catches errors early, reduces review cycles, and shows respect for your reviewers' time.

## When to Self-Review

Self-review **every PR** before submitting — no exceptions. Open the diff on GitHub (or use `git diff`) and read through it as if you were the reviewer.

## Checklist

### Code Quality

- [ ] No leftover debug statements (`console.log`, `print`, `debugger`, `var_dump`)
- [ ] No commented-out code that should be removed
- [ ] No TODO or FIXME comments that should be resolved in this PR
- [ ] No hardcoded values that should be environment variables or configuration
- [ ] No dead code or unused imports
- [ ] Variable and function names are clear and descriptive
- [ ] Complex logic has comments explaining **why**, not what

### Correctness

- [ ] The code does what the PR description says it does
- [ ] Edge cases are handled (empty inputs, null values, boundary conditions)
- [ ] Error cases are handled gracefully (try/catch, error returns, fallbacks)
- [ ] No off-by-one errors in loops or array access
- [ ] No race conditions in concurrent code

### Security

- [ ] No secrets, tokens, API keys, or passwords in the diff
- [ ] No new vulnerable dependencies introduced
- [ ] Input validation is present where user input is accepted
- [ ] No SQL injection, XSS, or command injection risks
- [ ] Authentication and authorization checks are correct

### Tests

- [ ] New code has corresponding tests
- [ ] Edge cases are covered in tests
- [ ] Tests are meaningful — not just asserting `true`
- [ ] All existing tests still pass
- [ ] Test coverage is maintained or improved

### Standards

- [ ] Commits follow the [Commit Messages](../../git-workflow/standards/commit-messages.md) standard
- [ ] Branch follows the [Branch Naming](../../git-workflow/standards/branch-naming.md) standard
- [ ] PR size is within guidelines (see [PR Size](pr-size.md))
- [ ] No unrelated changes mixed into this PR

### Documentation

- [ ] Public APIs or interfaces have documentation
- [ ] README is updated if user-facing behavior changed
- [ ] Changelog is updated if the project maintains one
- [ ] Breaking changes are documented

### Final Checks

- [ ] Diff contains only intentional changes — no accidental whitespace or formatting
- [ ] File names and paths follow project conventions
- [ ] No merge conflict markers left in files (`<<<<<<<`, `=======`, `>>>>>>>`)
- [ ] The PR builds and tests pass locally
