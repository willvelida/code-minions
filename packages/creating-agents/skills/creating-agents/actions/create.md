# Create Agent Definition

## Purpose

Generate a new agent definition file (`agents.md`) with a clear persona, project knowledge, executable commands, code style examples, and explicit boundaries.

---

## Flow

### Step 1: Determine Approach 🛑

Present two options to the user:

| Option | Description | Best For |
|--------|-------------|----------|
| **A: Guided** | Walk through agent creation step-by-step | First-time authors, complex agents |
| **B: Quick** | Provide agent name and role, generate definition | Experienced authors, simple agents |

**🛑 STOP**: Wait for the user to select Option A or Option B before proceeding.

**Success Criteria:**
- [ ] User has explicitly selected their preferred approach

---

### Step 2: Gather Agent Requirements

#### Path A: Guided Creation

Ask questions one category at a time. Wait for the user to respond before moving to the next category.

**Round 1: Role and Purpose 🛑**

1. What specific task or domain should this agent handle?
2. What is out of scope for this agent?
3. What persona best fits this role? (e.g., "senior test engineer", "technical writer", "security analyst")

**🛑 STOP**: Wait for the user to respond. If answers are vague, probe for specifics.

**Round 2: Project Context 🛑**

1. What is the tech stack? (languages, frameworks, versions)
2. What is the file structure? (key directories and their purposes)
3. What key dependencies or tools does the project use?

**🛑 STOP**: Wait for the user to respond. If they are unsure, scan the workspace for `package.json`, `pyproject.toml`, `Cargo.toml`, or similar to detect the stack.

**Round 3: Commands and Testing 🛑**

1. What build command(s) does the project use?
2. What test command(s) and framework?
3. What lint or format command(s)?
4. Any other commands the agent should know about?

**🛑 STOP**: Wait for the user to respond. If they are unsure, check `package.json` scripts, `Makefile`, or similar.

**Round 4: Style and Boundaries 🛑**

1. Are there naming conventions to follow? (functions, classes, files)
2. What files or directories should the agent never touch?
3. What actions should require user approval before proceeding?
4. Any existing style guides, linter configs, or `.editorconfig`?

**🛑 STOP**: Wait for the user to respond.

#### Path B: Quick Creation 🛑

Collect:
1. Agent name (kebab-case, e.g., `test-agent`, `docs-agent`)
2. One-sentence role description
3. Target repository or project (if not obvious from workspace)

**🛑 STOP**: Wait for user to provide name and role.

After receiving, scan the workspace to automatically detect:
- Tech stack from config files (`package.json`, `pyproject.toml`, etc.)
- File structure from directory listing
- Commands from scripts/build files
- Lint/format configuration from tool configs

Present findings and ask for confirmation before proceeding.

**Success Criteria:**
- [ ] Agent purpose and role clearly defined
- [ ] Tech stack documented with versions
- [ ] File structure understood
- [ ] Commands identified
- [ ] Boundaries discussed

---

### Step 3: Confirm Requirements 🛑

Summarise your understanding back to the user:

```
Based on our conversation, here's the agent I'll create:

- **Name**: <agent-name>
- **Role**: <specific role description>
- **Tech Stack**: <technologies with versions>
- **Key Commands**: <build, test, lint>
- **Write Scope**: <where the agent writes>
- **Read Scope**: <where the agent reads>
- **Never Touch**: <restricted files/directories>

Does this capture your intent? Anything to add or change?
```

**🛑 STOP**: Wait for user confirmation. If the user requests changes, revisit the relevant questions from Step 2. Do not proceed until the user confirms.

**Success Criteria:**
- [ ] Summary presented to user
- [ ] User confirmed requirements are correct

---

### Step 4: Load Standards

Load from this skill's `standards/`:
- `agent-structure.md` — Section order and file format
- `persona.md` — Writing effective personas
- `boundaries.md` — Three-tier boundary system
- `commands-and-tools.md` — Documenting commands
- `code-style.md` — Providing code examples

**Success Criteria:**
- [ ] All relevant standards loaded

---

### Step 5: Write Frontmatter

Create the YAML frontmatter:

```yaml
---
name: <agent-name>
description: <One-sentence description of what this agent does>
---
```

| Rule | Requirement |
|------|-------------|
| `name` | Kebab-case, descriptive (e.g., `test-agent`, `docs-agent`, `lint-agent`) |
| `description` | Third person, states what the agent does |

**Success Criteria:**
- [ ] Name is kebab-case and descriptive
- [ ] Description clearly states the agent's role

---

### Step 6: Write Persona Section

Apply guidelines per `standards/persona.md`:

```markdown
You are a [specific role] for this project.

## Persona
- You specialise in [specific skills, frameworks, tools]
- You understand [domain knowledge relevant to the task]
- Your output: [what you produce] that [quality standard]
```

**Success Criteria:**
- [ ] Clear role stated
- [ ] 2-4 specific specialisations listed
- [ ] Output type defined

---

### Step 7: Write Project Knowledge Section

Document the project context:

```markdown
## Project Knowledge
- **Tech Stack:** [technologies with versions]
- **File Structure:**
  - `src/` — [what's here]
  - `tests/` — [what's here]
  - `docs/` — [what's here]
- **Key Dependencies:** [important packages/libraries]
```

**Success Criteria:**
- [ ] Tech stack includes versions
- [ ] File paths are real and specific
- [ ] Key dependencies documented

---

### Step 8: Write Commands Section

Apply guidelines per `standards/commands-and-tools.md`. Place this section early:

```markdown
## Commands
- **Build:** `<build command>` (<what it does>)
- **Test:** `<test command>` (<framework and expectations>)
- **Lint:** `<lint command>` (<what it checks>)
```

**Success Criteria:**
- [ ] At least build, test, and lint commands included
- [ ] Commands include flags and options
- [ ] Each command has a purpose label

---

### Step 9: Write Code Style Section

Apply guidelines per `standards/code-style.md`:

- Document naming conventions
- Provide at least one good code example
- Provide at least one bad code example for contrast

**Success Criteria:**
- [ ] Naming conventions documented
- [ ] Good and bad examples provided
- [ ] Examples use the project's actual language

---

### Step 10: Write Boundaries Section

Apply guidelines per `standards/boundaries.md`:

```markdown
## Boundaries
- ✅ **Always:** [actions the agent must always do]
- ⚠️ **Ask first:** [actions requiring user approval]
- 🚫 **Never:** [actions the agent must never do]
```

**Success Criteria:**
- [ ] All three tiers present
- [ ] "Never commit secrets or API keys" included
- [ ] File paths are specific
- [ ] Read/write scope clearly defined

---

### Step 11: Validate Agent Definition

Run validation per `standards/checklist.md`:

1. **Six core areas** — Commands, testing, project structure, code style, git workflow, boundaries
2. **Quality checks** — Scannable format, commands early, code examples over prose
3. **Anti-patterns** — No vague personas, no tool-only commands, no missing boundary tiers

**Success Criteria:**
- [ ] All checklist items pass
- [ ] No anti-patterns present

---

### Step 12: Present Results 🛑

Provide:

1. **File Location** — Where the agent definition was created
2. **Agent Summary** — Name, role, key capabilities
3. **Next Steps** — How to test the agent, iteration guidance

Recommend the user test the agent with a real task immediately and iterate based on observed behavior.

**🛑 STOP**: Ask user if they want any modifications before finalising.
