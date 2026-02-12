# Review Agent Definition

## Purpose

Analyse an existing agent definition file for completeness, quality, and adherence to best practices, then provide actionable recommendations for improvement.

---

## Flow

### Step 1: Locate Agent Definition

Find the agent definition to review:

| Location | Common Paths |
|----------|-------------|
| Repository root | `AGENTS.md`, `agents.md` |
| GitHub agents | `.github/agents/<name>.md` |
| Project-specific | `agents/<name>.md` |

**Failure:** No agent definition found → Recommend Create action

**Success Criteria:**
- [ ] Agent definition file located and readable

---

### Step 2: Load Standards

Load from this skill's `standards/`:
- `agent-structure.md` — Section order and format
- `persona.md` — Persona quality guidelines
- `boundaries.md` — Three-tier boundary requirements
- `commands-and-tools.md` — Command documentation standards
- `code-style.md` — Code example requirements
- `checklist.md` — Full compliance checklist

**Success Criteria:**
- [ ] All review standards loaded

---

### Step 3: Frontmatter Check

Validate the YAML frontmatter:

| Check | Severity |
|-------|----------|
| `name` field present | 🔴 Critical |
| `name` is kebab-case | 🔴 Critical |
| `description` field present | 🔴 Critical |
| `description` states what the agent does | 🟠 High |
| `description` is one clear sentence | 🟡 Medium |

**Success Criteria:**
- [ ] Frontmatter compliance assessed
- [ ] Findings documented with severity

---

### Step 4: Persona Quality Analysis

Evaluate against `standards/persona.md`:

| Check | Severity |
|-------|----------|
| Clear role stated | 🔴 Critical |
| Specific specialisations listed (2-4) | 🟠 High |
| Output type defined | 🟠 High |
| Single responsibility | 🟠 High |
| Uses project-specific language | 🟡 Medium |
| Not "helpful assistant" or similarly vague | 🔴 Critical |

**Common Issues:**
- Too vague: "You are a helpful coding assistant"
- Multiple roles: "You write tests and also handle deployments"
- No output defined: Missing what the agent produces

**Success Criteria:**
- [ ] Persona quality assessed
- [ ] Improvement suggestions documented

---

### Step 5: Six Core Areas Audit

Check coverage of each core area per the article's best practices:

| Core Area | Present | Quality | Severity if Missing |
|-----------|---------|---------|---------------------|
| Commands | ☐ | — | 🔴 Critical |
| Testing | ☐ | — | 🔴 Critical |
| Project Structure | ☐ | — | 🟠 High |
| Code Style | ☐ | — | 🟠 High |
| Git Workflow | ☐ | — | 🟡 Medium |
| Boundaries | ☐ | — | 🔴 Critical |

For each area that is present, assess quality:
- **Commands**: Include flags? Purpose labels? Copy-pasteable?
- **Testing**: Framework specified? Coverage expectations?
- **Project Structure**: Real paths? Tech stack with versions?
- **Code Style**: Code snippets? Good/bad examples? Naming conventions?
- **Git Workflow**: Branch naming? Commit format?
- **Boundaries**: Three tiers? Specific file paths? Secrets rule?

**Success Criteria:**
- [ ] All six areas audited
- [ ] Coverage gaps identified

---

### Step 6: Commands Quality Check

Evaluate against `standards/commands-and-tools.md`:

| Check | Severity |
|-------|----------|
| Commands section appears early | 🟠 High |
| At least one test command | 🔴 Critical |
| Commands include flags and options | 🟠 High |
| Each command has a purpose label | 🟠 High |
| Commands are copy-pasteable | 🟠 High |
| No tool-name-only entries | 🟡 Medium |

**Success Criteria:**
- [ ] Command quality assessed

---

### Step 7: Boundaries Compliance Check

Evaluate against `standards/boundaries.md`:

| Check | Severity |
|-------|----------|
| ✅ Always-do tier present | 🔴 Critical |
| ⚠️ Ask-first tier present | 🔴 Critical |
| 🚫 Never-do tier present | 🔴 Critical |
| Secrets rule included | 🔴 Critical |
| File paths are specific | 🟠 High |
| Read/write scope defined | 🟠 High |

**Success Criteria:**
- [ ] Boundary compliance assessed

---

### Step 8: Quality and Format Check

Assess overall document quality:

| Check | Severity |
|-------|----------|
| Scannable format (tables, bullets, headers) | 🟠 High |
| Commands before style and boundaries | 🟡 Medium |
| Code examples over prose | 🟠 High |
| No vague language ("be careful", "best practices") | 🟡 Medium |
| Tech stack includes versions | 🟠 High |
| No time-sensitive information | 🟡 Medium |

**Success Criteria:**
- [ ] Quality assessed
- [ ] Format improvements identified

---

### Step 9: Generate Review Report 🛑

Present findings organised by severity:

```markdown
## Agent Definition Review: <agent-name>

### Summary
- **Overall Score:** [Excellent / Good / Needs Work / Insufficient]
- **Core Areas Covered:** X/6
- **Critical Issues:** X
- **Improvements Suggested:** X

### 🔴 Critical Issues
<list>

### 🟠 High Priority
<list>

### 🟡 Medium Priority
<list>

### 🟢 Suggestions
<list>

### Recommended Changes
<numbered list of specific, actionable changes>
```

**🛑 STOP**: Present the review and ask the user if they want help implementing any of the recommended changes.

**Success Criteria:**
- [ ] Review report presented with severity ratings
- [ ] Actionable recommendations provided
- [ ] User asked about implementing changes
