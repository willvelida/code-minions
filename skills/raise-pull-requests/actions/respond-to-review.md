# Respond to Review

Address reviewer feedback, push fixes, and re-request review. How you respond to feedback affects both the quality of the code and the health of the team.

## Prerequisites

- Pull request has been reviewed (at least one review with comments or change requests)
- You understand the difference between blockers, suggestions, and nits (see [Code Review](../../git-workflow/standards/code-review.md))

## Steps

### 1. Read All Feedback First

Before making any changes, read through **every comment** to understand the full picture. Reviewers may have related concerns across multiple files that are best addressed together.

Understand comment prefixes:

| Prefix | Meaning | Action Required |
|---|---|---|
| `blocker:` | Must be fixed before merge | Yes — fix before re-requesting review |
| `suggestion:` | Recommended but not required | Address or explain why not |
| `nit:` | Minor style or preference | Take it or leave it — acknowledge either way |
| `question:` | Clarification needed | Reply with an explanation |
| `praise:` | Positive feedback | No action needed — appreciate it |

### 2. Respond to Every Comment

Every review comment deserves a response. For each comment:

- **If you agree**: Make the change and reply with what you did
- **If you disagree**: Explain your reasoning respectfully — the reviewer may have context you don't, and vice versa
- **If it's unclear**: Ask for clarification before making changes

Examples:

```
✅ "Good catch — fixed in abc1234"
✅ "I considered that, but went with X because [reason]. Happy to change if you still feel strongly."
✅ "Can you clarify what you mean? I want to make sure I address the right concern."
```

### 3. Push Fixes as New Commits

During review, push fixes as **new commits** rather than amending or force-pushing. This lets reviewers see what changed since their last review:

```bash
# Make the requested changes
git add <files>
git commit -s -m "fix(scope): address review feedback

- Fix null check per reviewer comment
- Add missing error handling for edge case"
git push
```

**Do not force-push during review** unless explicitly asked by the reviewer. Force-pushing rewrites history and makes it harder to track incremental changes.

### 4. Resolve Conversations

After addressing a comment and pushing the fix:

1. Reply to the comment explaining what you changed
2. Mark the conversation as **resolved** on GitHub

Only the PR author or the commenter should resolve conversations. Do not resolve comments you haven't addressed.

### 5. Re-request Review

Once all feedback is addressed, re-request review from the original reviewers:

```bash
gh pr edit <pr-number> --add-reviewer <reviewer1>,<reviewer2>
```

Or use the GitHub UI:

1. Navigate to the PR
2. Click the refresh icon next to the reviewer's name in the sidebar
3. This notifies them that the PR is ready for another look

### 6. Handle Conflicting Feedback

If two reviewers give contradictory feedback:

1. Do not silently pick one — acknowledge both perspectives
2. Tag both reviewers in a comment explaining the conflict
3. Propose a resolution and ask for consensus
4. If no consensus, escalate to the team lead or tech lead

### 7. Know When to Push Back

Not all feedback needs to be accepted. It is appropriate to push back when:

- The suggestion is out of scope for this PR
- The change would introduce regression risk without clear benefit
- The feedback contradicts an established team convention

Always push back respectfully and with reasoning.

## Response Etiquette

- Respond within **one business day** of receiving review
- Be concise — reviewers are busy
- Say thank you — reviewing takes effort
- Never take feedback personally — it's about the code, not you
- If a discussion gets long, take it to a call or chat and summarize the outcome in the PR

## Checklist

- [ ] Read all feedback before making changes
- [ ] Every comment has a response
- [ ] Fixes pushed as new commits (not force-pushed)
- [ ] Resolved conversations are marked as resolved
- [ ] Review re-requested from original reviewers
