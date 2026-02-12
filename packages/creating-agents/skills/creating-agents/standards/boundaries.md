# Boundaries Standard

## Purpose

Defines how to write explicit boundaries for agent definitions using a three-tier system. Boundaries are the most critical safety mechanism — they prevent the agent from making destructive mistakes and clearly define the agent's operating scope.

---

## Three-Tier System

Every agent definition must include boundaries organised into three tiers:

| Tier | Symbol | Meaning | Example |
|------|--------|---------|---------|
| Always do | ✅ | Actions the agent must always perform | Run tests before finishing |
| Ask first | ⚠️ | Actions requiring explicit user approval | Adding new dependencies |
| Never do | 🚫 | Actions the agent must never perform | Committing secrets or API keys |

---

## Rules

| Rule | Requirement |
|------|-------------|
| All three tiers present | Every agent must define always, ask-first, and never rules |
| Specific over general | "Never edit `src/config/database.ts`" beats "Don't touch config" |
| File paths are explicit | Reference actual directories and files |
| Include secrets rule | Every agent must include "Never commit secrets or API keys" |
| Scope is clear | Define where the agent can read and where it can write |
| Production safety | Never-do tier must prevent production-affecting actions |

---

## Structure

```markdown
## Boundaries
- ✅ **Always:** [actions the agent must always do]
- ⚠️ **Ask first:** [actions requiring user approval]
- 🚫 **Never:** [actions the agent must never do]
```

---

## Common Boundary Categories

### Read/Write Scope

Define where the agent reads from and writes to:

```markdown
- ✅ **Always:** Read from `src/`, write to `tests/`
- 🚫 **Never:** Modify files in `src/`
```

### Safety and Secrets

Prevent destructive actions:

```markdown
- 🚫 **Never:** Commit secrets, API keys, or credentials
- 🚫 **Never:** Edit `.env` files or `node_modules/`
- 🚫 **Never:** Run destructive commands against production
```

### Approval Gates

Require user confirmation for impactful changes:

```markdown
- ⚠️ **Ask first:** Before modifying existing documentation
- ⚠️ **Ask first:** Before adding new dependencies to `package.json`
- ⚠️ **Ask first:** Database schema changes
```

---

## Examples

**Good — Specific, three-tier, file-path-aware:**

```markdown
## Boundaries
- ✅ **Always:** Write new files to `docs/`, follow the style examples, run markdownlint
- ✅ **Always:** Run `npm test` before considering work complete
- ⚠️ **Ask first:** Before modifying existing documents in a major way
- ⚠️ **Ask first:** Before adding new npm dependencies
- 🚫 **Never:** Modify code in `src/`, edit config files, commit secrets
- 🚫 **Never:** Delete or overwrite existing test files
```

**Bad — Vague, missing tiers:**

```markdown
## Rules
- Be careful with the code
- Don't break things
- Ask if unsure
```

---

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| Missing a tier | Include all three: always, ask-first, never |
| Vague rules ("be careful") | Use specific file paths and actions |
| No secrets rule | Always include "Never commit secrets or API keys" |
| No read/write scope | Define where the agent can and cannot write |
| Overly restrictive | Balance safety with the agent's ability to do its job |
