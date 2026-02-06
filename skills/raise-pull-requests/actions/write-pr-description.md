# Write PR Description

Craft a clear, informative pull request title and description that helps reviewers understand what changed and why. Good descriptions reduce review time and back-and-forth.

## Prerequisites

- Changes are prepared and ready (see [Prepare Pull Request](prepare-pull-request.md))
- You know the purpose, scope, and impact of your changes

## Steps

### 1. Write the PR Title

Use the same format as conventional commit messages:

```
<type>[scope]: <description>
```

Examples:

```
feat(auth): add JWT token validation middleware
fix(api): correct null pointer in user service
docs(readme): update deployment instructions
refactor(db): extract connection pool into module
```

Rules:

- Use present tense, imperative mood ("add" not "added")
- Keep under 72 characters
- Be specific — "fix bug" is too vague; "fix null pointer in user lookup" is clear

### 2. Write the PR Description

Use the [Pull Request Template](../standards/pull-request-template.md) as a starting point. Every PR description should include:

#### Purpose

Explain **what** this PR does and **why** the change is needed. Link to the issue, ticket, or discussion that motivated the work.

```markdown
## 📝 Description

Add JWT token validation middleware to protect API endpoints.
Previously, all endpoints were publicly accessible.
```

#### Changes

List each change with its type and the files affected. This gives reviewers a roadmap of what to look at.

```markdown
## 🚀 Changes

**✨ feat(middleware): Add JWT validation middleware**
What: New middleware that validates JWT tokens on incoming requests
Why: API endpoints need authentication before accessing protected resources
📁 Files: `src/middleware/auth.ts` (`validateToken`, `extractUser`)

**✨ feat(config): Add JWT configuration**
What: Environment variables for JWT secret and expiration
Why: Token validation needs configurable secret and TTL
📁 Files: `src/config/auth.ts` (new file)
```

#### Related Issues

Link to tracking issues so they auto-close when the PR merges:

```markdown
## 🔗 Related Issues

Closes #123
Refs #456
```

#### Additional Context

Include anything else reviewers need — screenshots, architecture decisions, migration notes, or known limitations:

```markdown
## 🙏 Additional Context

This middleware is intentionally permissive for the `/health` and `/docs` endpoints.
A follow-up PR will add role-based access control.
```

### 3. Guide Reviewers Through the Changes

For PRs that touch multiple files, tell reviewers where to start and what order to follow:

```markdown
**Review order:**
1. Start with `src/config/auth.ts` — new configuration
2. Then `src/middleware/auth.ts` — core middleware logic
3. Finally `src/routes/api.ts` — where the middleware is applied
```

### 4. Specify the Type of Feedback Needed

Help reviewers focus their effort:

```markdown
**Feedback requested:**
- Architecture: Is middleware the right pattern here?
- Edge cases: Are all token expiration scenarios handled?
- Quick look only — logic is straightforward
```

### 5. Character Limit

Keep the full PR description (including all text, emojis, and formatting) under **4000 characters**. This ensures compatibility with GitHub API limits and keeps descriptions scannable.

## Description Quality Checklist

- [ ] Title follows `<type>[scope]: <description>` format
- [ ] Description explains what and why
- [ ] Changes are listed with types and affected files
- [ ] Related issues are linked with `Closes` or `Refs`
- [ ] Review guidance is provided for multi-file PRs
- [ ] Description is under 4000 characters

See [PR Descriptions](../standards/pr-descriptions.md) for the full standard.
