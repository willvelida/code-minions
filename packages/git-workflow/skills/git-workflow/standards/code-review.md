# Code Review Standard

Guidelines for reviewing pull requests. Good code review catches bugs, improves code quality, and shares knowledge across the team.

## Reviewer Responsibilities

- Review within **one business day** of being assigned
- Be constructive and specific — suggest alternatives, not just problems
- Approve when satisfied, request changes when blockers exist
- Distinguish between blockers, suggestions, and nits

## Comment Prefixes

Use prefixes to signal the weight of feedback:

| Prefix | Meaning |
|---|---|
| `blocker:` | Must be fixed before merge |
| `suggestion:` | Recommended but not required |
| `nit:` | Minor style or preference — take it or leave it |
| `question:` | Clarification needed, not necessarily a change |
| `praise:` | Something done well — positive feedback matters |

## Review Checklist

### Correctness

- [ ] Does the code do what the PR description says?
- [ ] Are edge cases handled?
- [ ] Are error cases handled gracefully?
- [ ] Is the logic correct and complete?

### Security

- [ ] No secrets, tokens, or credentials in the code
- [ ] No SQL injection, XSS, or other vulnerability risks
- [ ] Input validation is present where needed
- [ ] Authentication and authorization checks are correct

### Quality

- [ ] Code is readable and well-organized
- [ ] Functions and variables have clear names
- [ ] No unnecessary duplication
- [ ] Complex logic has comments explaining why
- [ ] No dead code or commented-out blocks

### Tests

- [ ] New code has corresponding tests
- [ ] Edge cases are covered
- [ ] Tests are meaningful (not just asserting `true`)
- [ ] Existing tests still pass

### Standards Compliance

- [ ] Commits follow the [Commit Messages](commit-messages.md) standard
- [ ] Branch follows the [Branch Naming](branch-naming.md) standard
- [ ] PR description is complete and links related issues

## Approval Guidelines

| Condition | Action |
|---|---|
| All checks pass, no blockers | Approve |
| Minor suggestions only | Approve with comments |
| Blockers found | Request changes |
| Major architectural concerns | Request changes, discuss with author |

## Author Responsibilities

- Keep PRs small and focused — one logical change per PR
- Write a clear PR description (what, why, how to test)
- Respond to all review comments
- Do not merge your own PR without at least one approval (unless the team has agreed otherwise)
- Resolve or address every comment before re-requesting review
