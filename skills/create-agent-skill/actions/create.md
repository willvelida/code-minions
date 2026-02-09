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

Load `standards/clarifying-questions.md` and use it to guide this step.

#### Path A: Guided Creation

Ask questions one category at a time. Wait for the user to respond before moving to the next category.

**Round 1: Purpose and Scope 🛑**

1. What task or workflow should this skill help agents perform?
2. What is out of scope for this skill?

**🛑 STOP**: Wait for the user to respond. If answers are vague, probe deeper per `standards/clarifying-questions.md` before continuing.

**Round 2: Trigger Conditions 🛑**

1. What words or phrases would a user say that should trigger this skill?
2. Are there situations where this skill should NOT be used?

**🛑 STOP**: Wait for the user to respond. Aim for at least 3 specific trigger keywords.

**Round 3: Actions and Workflows 🛑**

1. What distinct actions or workflows should the skill support?
2. For each action, what are the high-level steps?
3. Which steps require user input or decisions?

**🛑 STOP**: Wait for the user to respond. If more than 5 actions are listed, suggest splitting into multiple skills.

**Round 4: Standards, Structure, and Dependencies 🛑**

1. Are there naming conventions, style rules, or formats to follow?
2. Will the skill need scripts, templates, or reference material?
3. Does this skill depend on specific tools being available?

**🛑 STOP**: Wait for the user to respond.

#### Path B: Quick Creation 🛑

Collect:
1. Skill name (must be kebab-case, lowercase, 1-64 chars)
2. Brief description of purpose and triggers

**🛑 STOP**: Wait for user to provide name and description.

After receiving, ask one clarification round:
1. What should this skill NOT do? (scope boundaries)
2. Any specific tools or dependencies required?

**🛑 STOP**: Wait for user to respond. Proceed once scope is clear.

**Success Criteria:**
- [ ] Skill purpose clearly defined
- [ ] At least 3 trigger keywords identified
- [ ] Scope boundaries established (in and out)
- [ ] Actions/workflows listed
- [ ] Structure complexity determined

---

### Step 3: Confirm Requirements 🛑

Summarise your understanding back to the user using this template:

```
Based on our conversation, here's what I understand:

- **Purpose**: <one-sentence summary>
- **Triggers**: <keywords and conditions>
- **Scope**: <what's included> / <what's excluded>
- **Actions**: <list of actions>
- **Structure**: <minimal | standard | full>
- **Dependencies**: <tools, references, or none>

Does this capture your intent? Anything to add or change?
```

**🛑 STOP**: Wait for user confirmation. If the user requests changes, revisit the relevant questions from Step 2. Do not proceed until the user confirms.

**Success Criteria:**
- [ ] Summary presented to user
- [ ] User confirmed requirements are correct

---

### Step 4: Load Standards

Load from this skill's `standards/`:
- `specification.md` — Core format requirements
- `naming.md` — Naming conventions
- `descriptions.md` — Writing effective descriptions
- `structure.md` — Directory structure guidance
- `instructions.md` — Writing effective instructions
- `clarifying-questions.md` — Questioning framework (used in Steps 2-3)

**Success Criteria:**
- [ ] All relevant standards loaded

---

### Step 5: Generate Skill Name

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

### Step 6: Write Skill Description

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

### Step 7: Plan Skill Structure

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

### Step 8: Write SKILL.md Content

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

### Step 9: Create Action Files (If Applicable)

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

### Step 10: Create Standards Files (If Applicable)

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

### Step 11: Create Checklist

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

### Step 12: Validate Skill

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

### Step 13: Present Results 🛑

Provide:

1. **Created Files** — List all files created with brief descriptions
2. **Skill Summary** — Name, description, capabilities
3. **Next Steps** — Testing recommendations, iteration guidance

**🛑 STOP**: Ask user if they want any modifications before finalising.
