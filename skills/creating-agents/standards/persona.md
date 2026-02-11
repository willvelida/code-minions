# Persona Standard

## Purpose

Defines how to write effective agent personas that give the AI a clear identity, specific expertise, and focused output expectations. The persona is the most important section — it determines how the agent interprets every subsequent instruction.

---

## Rules

| Rule | Requirement |
|------|-------------|
| State a clear role | "You are a [specific role] for this project" |
| List specialisations | What skills, frameworks, and tools the agent knows |
| Define output type | What the agent produces (docs, tests, API routes, etc.) |
| Be specific | "Expert in React Testing Library" not "knows testing" |
| Third person in description | Frontmatter `description` uses third person |
| Second person in body | Body instructions address the agent as "you" |
| One persona per agent | Do not combine multiple roles into one agent |

---

## Structure

A persona section should cover three things:

1. **Role** — What the agent is (e.g., "You are a senior test engineer")
2. **Skills** — What the agent knows (e.g., "You are fluent in TypeScript and Jest")
3. **Output** — What the agent produces (e.g., "You write unit test files")

```markdown
## Persona
- You are a [role] for this project
- You specialise in [specific skills, frameworks, tools]
- You understand [domain knowledge relevant to the task]
- Your output: [what you produce] that [quality standard]
```

---

## Examples

**Good — Specific role, skills, and output:**

```markdown
## Persona
- You are an expert technical writer for this project
- You are fluent in Markdown and can read TypeScript code
- You write for a developer audience, focusing on clarity and practical examples
- Your output: documentation in `docs/` that helps new developers get started quickly
```

**Good — Narrow specialist:**

```markdown
## Persona
- You are a security analyst who reviews code for vulnerabilities
- You specialise in OWASP Top 10, dependency analysis, and secrets detection
- You understand Node.js, Express, and common attack vectors
- Your output: security reports with severity ratings and remediation steps
```

**Bad — Too vague:**

```markdown
## Persona
- You are a helpful assistant
- You can do many things
- You help with code
```

**Bad — Multiple roles combined:**

```markdown
## Persona
- You write tests, documentation, and also handle deployments
- You review code and fix bugs
```

---

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| "Helpful assistant" | State a specific role: "Senior test engineer" |
| No skills listed | List 2-4 specific frameworks or tools |
| No output defined | State what the agent produces |
| Multiple responsibilities | Split into separate agents |
| Generic language | Use project-specific terms and file paths |
