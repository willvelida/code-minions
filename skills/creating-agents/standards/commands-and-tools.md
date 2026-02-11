# Commands and Tools Standard

## Purpose

Defines how to document executable commands in agent definitions. Commands are the most frequently referenced section — agents use them to build, test, lint, and validate their work. Putting commands early and making them specific dramatically improves agent performance.

---

## Rules

| Rule | Requirement |
|------|-------------|
| Commands near the top | Place the commands section before code style and boundaries |
| Include flags and options | `pytest -v --tb=short` not just `pytest` |
| State the purpose | Each command has a label explaining what it does |
| Include output expectations | Describe what success looks like when relevant |
| Group by category | Organise commands into build, test, lint, and run groups |
| Use actual commands | Commands must be copy-pasteable and executable |

---

## Structure

```markdown
## Commands
- **Build:** `npm run build` (compiles TypeScript, outputs to `dist/`)
- **Test:** `npm test` (runs Jest, must pass before commits)
- **Test with coverage:** `npm test -- --coverage` (minimum 80% coverage)
- **Lint:** `npm run lint` (ESLint with project rules)
- **Lint fix:** `npm run lint --fix` (auto-fixes style errors)
- **Format:** `npx prettier --write .` (formats all files)
```

---

## Command Categories

### Build Commands

Commands that compile, transpile, or bundle the project:

```markdown
- **Build:** `npm run build` (compiles TypeScript to `dist/`)
- **Build docs:** `npm run docs:build` (generates documentation site)
```

### Test Commands

Commands that run test suites. Include framework, flags, and coverage:

```markdown
- **Unit tests:** `npm test` (runs Jest)
- **Integration tests:** `npm run test:integration` (requires running database)
- **Coverage:** `npm test -- --coverage` (target: 80%+)
- **Single file:** `npm test -- --testPathPattern=<filename>`
```

### Lint and Format Commands

Commands that check and fix code style:

```markdown
- **Lint:** `npm run lint` (ESLint with project config)
- **Lint fix:** `npm run lint --fix` (auto-fix safe issues)
- **Format:** `npx prettier --write .` (format all files)
- **Markdown lint:** `npx markdownlint docs/` (validate markdown)
```

### Run Commands

Commands that start the application or development server:

```markdown
- **Dev server:** `npm run dev` (starts on http://localhost:3000)
- **Production:** `npm start` (serves built files from `dist/`)
```

---

## Examples

**Good — Specific with flags, purpose, and grouping:**

```markdown
## Commands you can use
**Build:**
- `npm run build` — Compiles TypeScript, outputs to `dist/`
- `npm run build:watch` — Rebuilds on file changes

**Test:**
- `npm test` — Runs Jest (must pass before commits)
- `npm test -- --coverage` — Generates coverage report (target 80%+)
- `npm test -- --testPathPattern=auth` — Runs only auth tests

**Lint:**
- `npm run lint` — Checks code style with ESLint
- `npx markdownlint docs/` — Validates markdown formatting
```

**Bad — Just tool names, no context:**

```markdown
## Commands
- npm
- jest
- eslint
```

---

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| Tool names only ("jest") | Full commands with flags: `npx jest --verbose` |
| No purpose labels | Add what each command does: "Runs unit tests" |
| Missing test commands | Always include at least one test command |
| No lint/format commands | Include linting so agents can validate their work |
| Platform-specific commands | Note when commands differ across OS |
