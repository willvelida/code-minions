# Agent Structure Standard

## Purpose

Defines the required sections and organisation for agent definition files (`agents.md`). A well-structured agent definition ensures the AI agent has all the context it needs, presented in an order that maximises effectiveness.

---

## File Format

Agent definitions use Markdown with optional YAML frontmatter:

```markdown
---
name: agent-name
description: One-sentence description of what this agent does
---

<agent instructions body>
```

### Frontmatter Fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Kebab-case identifier (e.g., `docs-agent`, `test-agent`) |
| `description` | Yes | One-sentence summary of the agent's role |

---

## Section Order

Present sections in this order. Agents reference commands frequently, so they appear early.

| Order | Section | Required | Purpose |
|-------|---------|----------|---------|
| 1 | Persona | Yes | Who the agent is and what it specialises in |
| 2 | Project Knowledge | Yes | Tech stack, file structure, key dependencies |
| 3 | Commands and Tools | Yes | Executable commands the agent can run |
| 4 | Standards / Code Style | Recommended | Naming conventions, patterns, and examples |
| 5 | Git Workflow | Recommended | Branch naming, commit messages, PR conventions |
| 6 | Boundaries | Yes | Always-do, ask-first, and never-do rules |

---

## Rules

| Rule | Requirement |
|------|-------------|
| One agent, one job | Each agent definition should focus on a single domain or task |
| Commands near the top | Executable commands appear before style guides and boundaries |
| Concrete over abstract | Use real file paths, real commands, real code snippets |
| Versions matter | Specify tech stack versions (e.g., "React 18" not "React") |
| File paths are explicit | Reference actual paths (e.g., `src/components/`) not generic descriptions |
| Keep it scannable | Use tables, bullet lists, and headers — not long paragraphs |

---

## Examples

**Good — Clear structure with all sections:**

```markdown
---
name: test-agent
description: Writes unit tests for TypeScript functions
---

You are a quality software engineer who writes comprehensive tests.

## Persona
- You specialise in writing unit and integration tests
- You understand Jest, React Testing Library, and Playwright
- Your output: test files that catch bugs early

## Project Knowledge
- **Tech Stack:** React 18, TypeScript 5.3, Vite, Tailwind CSS
- **File Structure:**
  - `src/` — Application source code
  - `tests/` — All test files (you WRITE here)
  - `src/components/` — React components

## Commands
- **Run tests:** `npm test` (runs Jest)
- **Run with coverage:** `npm test -- --coverage`
- **Lint:** `npm run lint`

## Code Style
- Test files: `<ComponentName>.test.tsx`
- Use `describe` blocks grouped by function
- Prefer `userEvent` over `fireEvent`

## Boundaries
- ✅ **Always:** Write to `tests/`, run tests before finishing
- ⚠️ **Ask first:** Adding new test dependencies
- 🚫 **Never:** Modify source code in `src/`, remove failing tests
```

**Bad — Vague, unstructured, missing key sections:**

```markdown
You are a helpful coding assistant. Help the user with their code.
Write good tests. Follow best practices.
```
