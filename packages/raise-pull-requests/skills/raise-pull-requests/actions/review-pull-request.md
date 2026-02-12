# Review Pull Request

Review someone else's pull request with constructive, actionable feedback. Good reviews catch bugs, improve quality, and share knowledge across the team.

## Prerequisites

- You have been assigned as a reviewer or are volunteering to review
- You are familiar with the area of code being changed (or willing to learn)

## Steps

### 1. Understand the Context

Before looking at code, read the PR description to understand:

- **What** is being changed
- **Why** the change is needed
- **How** to test or verify the change
- Any reviewer guidance the author provided (review order, feedback type requested)

Check linked issues for additional context.

### 2. Review at the Right Level

Match your review depth to the type of feedback requested:

| Request | Review depth |
|---|---|
| "Quick look" | High-level scan for obvious issues |
| "Architecture review" | Focus on design patterns and structure |
| "Full review" | Line-by-line analysis |
| No guidance specified | Default to a full review |

### 3. Check the Diff

Review the changes file by file. For each file:

- Does the change match the PR description?
- Are edge cases handled?
- Are error cases handled gracefully?
- Is the logic correct and complete?
- Is the code readable and well-organized?

```bash
# View the PR diff locally
gh pr diff <pr-number>

# Check out the PR branch locally for deeper testing
gh pr checkout <pr-number>
```

### 4. Use Comment Prefixes

Signal the weight of your feedback with prefixes:

| Prefix | Meaning |
|---|---|
| `blocker:` | Must be fixed before merge |
| `suggestion:` | Recommended but not required |
| `nit:` | Minor style or preference |
| `question:` | Clarification needed |
| `praise:` | Something done well |

Examples:

```
blocker: This will throw a NullReferenceException if `user` is null.

suggestion: Consider extracting this into a helper function for reuse.

nit: Prefer `const` over `let` here since the value is never reassigned.

question: Is this intentionally using a 30-second timeout? Seems high for this endpoint.

praise: Clean separation of concerns here — nice refactor.
```

### 5. Review the Checklist

Use the following checklist for each review:

#### Correctness

- [ ] Does the code do what the PR description says?
- [ ] Are edge cases handled?
- [ ] Are error cases handled gracefully?

#### Security

- [ ] No secrets, tokens, or credentials in the code
- [ ] No injection vulnerabilities (SQL, XSS, command injection)
- [ ] Input validation is present where needed

#### Quality

- [ ] Code is readable and well-organized
- [ ] No unnecessary duplication
- [ ] Complex logic has explanatory comments
- [ ] No dead code or leftover debug statements

#### Tests

- [ ] New code has corresponding tests
- [ ] Edge cases are covered in tests
- [ ] Existing tests still pass

#### Standards

- [ ] Commits follow the [Commit Messages](../../git-workflow/standards/commit-messages.md) standard
- [ ] PR description is complete and links related issues

### 6. Submit Your Review

Choose the appropriate review action:

| Condition | Action |
|---|---|
| All checks pass, no blockers | **Approve** |
| Minor suggestions only | **Approve** with comments |
| Blockers found | **Request changes** |
| Major architectural concerns | **Request changes** and discuss with author |

```bash
# Approve
gh pr review <pr-number> --approve --body "Looks good — clean implementation."

# Request changes
gh pr review <pr-number> --request-changes --body "A few blockers to address before merging."

# Comment only (no approval or rejection)
gh pr review <pr-number> --comment --body "Left some suggestions — nothing blocking."
```

### 7. Follow Up

After submitting your review:

- Monitor for the author's responses
- Re-review promptly when the author re-requests review
- Approve once all blockers are resolved

## Review Etiquette

- Review within **one business day** of being assigned
- Be constructive — suggest alternatives, not just problems
- Distinguish between blockers and preferences
- Acknowledge good work with `praise:` comments
- If a discussion gets long, suggest a call and summarize the outcome in the PR
- Remember: the goal is better code, not winning arguments
