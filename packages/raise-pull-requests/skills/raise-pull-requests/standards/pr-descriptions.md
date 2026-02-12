# PR Descriptions Standard

Every pull request must have a clear, structured description that helps reviewers understand what changed and why. Good descriptions reduce review time, prevent misunderstandings, and serve as documentation for future reference.

## Title Format

Use the same format as conventional commit messages:

```
<type>[scope]: <description>
```

### Rules

- Use present tense, imperative mood: "add" not "added"
- Keep under **72 characters**
- Be specific: "fix null pointer in user lookup" not "fix bug"
- Include the scope when the change targets a specific module or area

### Examples

```
feat(auth): add JWT token validation middleware
fix(api): correct null check in user service
docs(readme): update deployment instructions
refactor(db): extract connection pool into module
chore(deps): update lodash to 4.17.21
test(auth): add unit tests for login endpoint
```

## Description Structure

Use the [Pull Request Template](pull-request-template.md) as the base. Every description must include:

### 1. Summary (Required)

A brief explanation of **what** this PR does and **why**. One to three sentences.

```markdown
## 📝 Description

Add JWT token validation middleware to protect API endpoints.
Previously, all endpoints were publicly accessible. This change
enforces authentication on all routes except health checks.
```

### 2. Changes (Required)

List each logical change with its type, what was changed, why, and which files are affected:

```markdown
## 🚀 Changes

**✨ feat(middleware): Add JWT validation middleware**
What: New middleware that validates JWT tokens on incoming requests
Why: API endpoints need authentication before accessing protected resources
📁 Files: `src/middleware/auth.ts` (`validateToken`, `extractUser`)
```

### 3. Related Issues (Required when applicable)

Link to tracking issues so they auto-close on merge:

```markdown
## 🔗 Related Issues

Closes #123
Refs #456
```

Use these keywords to auto-close issues: `Closes`, `Fixes`, `Resolves`.

Use `Refs` to reference related issues without closing them.

### 4. Additional Context (Optional)

Include anything else reviewers need:

- Screenshots or recordings for UI changes
- Architecture decisions and trade-offs
- Migration notes or deployment steps
- Known limitations or follow-up work planned

```markdown
## 🙏 Additional Context

This middleware is intentionally permissive for `/health` and `/docs`.
A follow-up PR will add role-based access control.
```

## Reviewer Guidance

For PRs that touch multiple files, include guidance on review order and the type of feedback you need:

```markdown
**Review order:**
1. `src/config/auth.ts` — new configuration
2. `src/middleware/auth.ts` — core logic
3. `src/routes/api.ts` — middleware applied

**Feedback requested:** Architecture review — is middleware the right pattern here?
```

## Character Limit

The full PR description must not exceed **4000 characters** (including all text, emojis, spaces, and formatting). This ensures compatibility with the GitHub API and keeps descriptions scannable.

## Anti-Patterns

| ❌ Avoid | ✅ Instead |
|---|---|
| Empty description | Always fill in the template |
| "Fixed stuff" | Explain what was fixed and why |
| Pasting the full diff | Summarize changes at a high level |
| Wall of text with no structure | Use headings, lists, and the template |
| Only linking an issue with no summary | Summarize even if an issue exists — reviewers shouldn't have to leave the PR |
