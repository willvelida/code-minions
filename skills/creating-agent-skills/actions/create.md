# Create Agent Skill

## Purpose

Generate a new, specification-compliant Agent Skill with proper structure, metadata, and instructions.

---

## Flow

### Step 1: Determine Approach 🛑

Present two options to the user:

| Option | Description | Best For |
|--------|-------------|----------|
| **A: Guided** | Walk through skill creation step-by-step | First-time skill authors, complex skills |
| **B: Quick** | Provide name and description, generate structure | Experienced authors, simple skills |

**🛑 STOP**: Wait for the user to select Option A or Option B before proceeding.

**Success Criteria:**
- [ ] User has explicitly selected their preferred approach

---

### Step 2: Gather Skill Requirements

#### Path A: Guided Creation 🛑

Collect from user through conversation:

1. **Purpose**: What does this skill help agents do?
2. **Triggers**: When should an agent use this skill?
3. **Scope**: What actions/workflows does it cover?
4. **Resources**: Will it need scripts, references, or assets?

**🛑 STOP**: Wait for user to provide requirements.

#### Path B: Quick Creation 🛑

Collect:
1. Skill name (must be kebab-case, lowercase, 1-64 chars)
2. Brief description of purpose and triggers

**🛑 STOP**: Wait for user to provide name and description.

**Success Criteria:**
- [ ] Skill purpose clearly defined
- [ ] Trigger conditions identified
- [ ] Scope established

---

### Step 3: Load Standards

Load from this skill's `standards/`:
- `specification.md` — Core format requirements
- `naming.md` — Naming conventions
- `descriptions.md` — Writing effective descriptions
- `structure.md` — Directory structure guidance
- `instructions.md` — Writing effective instructions

**Success Criteria:**
- [ ] All relevant standards loaded

---

### Step 4: Generate Skill Name

Apply naming conventions per `naming.md`:

| Rule | Requirement |
|------|-------------|
| Format | Lowercase letters, numbers, hyphens only |
| Length | 1-64 characters |
| Style | Gerund form preferred (e.g., `processing-pdfs`) |
| Restrictions | No `anthropic` or `claude` reserved words |
| Match | Must match parent directory name |

**Examples:**
- ✅ `processing-pdfs`, `analyzing-spreadsheets`, `creating-devcontainers`
- ❌ `PDF-Processing` (uppercase), `-pdf` (starts with hyphen), `helper` (vague)

**Success Criteria:**
- [ ] Name follows all conventions
- [ ] Name clearly indicates skill purpose

---

### Step 5: Write Skill Description

Apply guidelines per `descriptions.md`:

| Rule | Requirement |
|------|-------------|
| Length | 1-1024 characters |
| Voice | Third person ("Processes files" not "I process files") |
| Content | What the skill does AND when to use it |
| Keywords | Include specific trigger terms |

**Template:**
```
<What the skill does — capabilities>. Use when <trigger conditions — when to activate>.
```

**Examples:**
- ✅ `Extract text and tables from PDF files, fill forms, merge documents. Use when working with PDF files or when the user mentions PDFs, forms, or document extraction.`
- ❌ `Helps with PDFs.` (too vague, no triggers)

**Success Criteria:**
- [ ] Description follows third-person voice
- [ ] Includes what AND when
- [ ] Contains specific keywords for discovery

---

### Step 6: Plan Skill Structure

Determine directory structure per `structure.md`:

**Minimal Structure:**
```
skill-name/
└── SKILL.md          # Required
```

**Standard Structure** (for skills with procedures):
```
skill-name/
├── SKILL.md          # Main instructions + metadata
├── actions/          # Step-by-step procedures
│   ├── action-one.md
│   └── action-two.md
└── standards/        # Guidelines and conventions
    ├── guideline.md
    └── checklist.md
```

**Full Structure** (for complex skills):
```
skill-name/
├── SKILL.md          # Overview and navigation
├── actions/          # Procedures
├── standards/        # Guidelines
├── scripts/          # Executable code
├── references/       # Detailed documentation
└── assets/           # Templates, data files
```

**Success Criteria:**
- [ ] Structure matches skill complexity
- [ ] All needed directories identified

---

### Step 7: Write SKILL.md Content

Create the main skill file per `specification.md` and `instructions.md`:

#### Frontmatter (Required)
```yaml
---
name: skill-name
description: 'Description following standards/descriptions.md guidelines.'
license: MIT
allowed-tools: Bash
---
```

#### Body Content
Structure the body with:

1. **Overview** — Brief explanation of what the skill does
2. **Capabilities** — Table of actions/features (if applicable)
3. **Standards** — Table of bundled standards (if applicable)
4. **Principles** — Key concepts or guidelines (3-5 max)
5. **Usage** — How to use the skill
6. **References** — External documentation links

**Token Budget:**
- Keep SKILL.md body under 500 lines
- Move detailed content to separate files
- Use progressive disclosure

**Success Criteria:**
- [ ] Frontmatter complete with required fields
- [ ] Body under 500 lines
- [ ] Clear navigation to additional files

---

### Step 8: Create Action Files (If Applicable)

For each action, create a file in `actions/`:

**Action File Structure:**
```markdown
# Action Name

## Purpose
<One-line description of what this action accomplishes>

---

## Flow

### Step 1: <Step Name>
<Instructions>

**Success Criteria:**
- [ ] <Measurable outcome>

---

### Step 2: <Step Name>
...
```

**Guidelines:**
- Use numbered steps with clear headers
- Include 🛑 STOP markers where user input is required
- Add success criteria checkboxes for each step
- Keep instructions concise — agents are smart

**Success Criteria:**
- [ ] Each action has clear purpose
- [ ] Steps are numbered and sequential
- [ ] Stop points marked where input needed

---

### Step 9: Create Standards Files (If Applicable)

For each standard, create a file in `standards/`:

**Standards File Structure:**
```markdown
# Standard Name

## Purpose
<Why this standard exists>

## Rules

| Rule | Requirement |
|------|-------------|
| <Rule Name> | <Specific requirement> |

## Examples

**Good:**
<example>

**Bad:**
<counter-example>
```

**Guidelines:**
- Be specific and measurable
- Provide good and bad examples
- Keep each file focused on one topic

**Success Criteria:**
- [ ] Each standard has clear purpose
- [ ] Rules are specific and measurable
- [ ] Examples provided

---

### Step 10: Create Checklist

Create `standards/checklist.md` consolidating all validation criteria:

```markdown
# Skill Compliance Checklist

## Specification Compliance
- [ ] Name follows naming conventions
- [ ] Description is 1-1024 characters
- [ ] SKILL.md has required frontmatter
- [ ] Directory name matches skill name

## Quality Checks
- [ ] SKILL.md body under 500 lines
- [ ] Instructions are concise
- [ ] Examples provided where helpful
- [ ] File references are one level deep

## Testing
- [ ] Skill activates on expected triggers
- [ ] Instructions are clear and followable
- [ ] No time-sensitive information
```

**Success Criteria:**
- [ ] Checklist covers specification requirements
- [ ] Checklist covers quality guidelines

---

### Step 11: Validate Skill

Run validation per `checklist.md`:

1. **Specification compliance** — All required fields present and valid
2. **Quality checks** — Token budget, conciseness, examples
3. **Structure check** — Files referenced exist, no deep nesting

If using the skills-ref tool:
```bash
skills-ref validate ./skill-name
```

**Success Criteria:**
- [ ] All checklist items pass
- [ ] No validation errors

---

### Step 12: Present Results 🛑

Provide:

1. **Created Files** — List all files created with brief descriptions
2. **Skill Summary** — Name, description, capabilities
3. **Next Steps** — Testing recommendations, iteration guidance

**🛑 STOP**: Ask user if they want any modifications before finalising.
