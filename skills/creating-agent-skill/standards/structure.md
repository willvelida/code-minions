# Skill Structure

## Purpose

Define directory structure patterns and progressive disclosure strategies for efficient context use.

## Progressive Disclosure

Skills use progressive disclosure to manage context efficiently:

1. **Discovery** (~100 tokens): At startup, agents load only `name` and `description`
2. **Activation** (<5000 tokens): When triggered, agent reads full `SKILL.md`
3. **Execution** (as needed): Agent loads referenced files or executes scripts

This keeps agents fast while providing access to more context on demand.

## Structure Patterns

### Minimal Structure

For simple skills with just instructions:

```
skill-name/
└── SKILL.md          # Required
```

**Use when:** Single-purpose skills, simple procedures, no separate standards needed.

### Standard Structure

For skills with procedures and guidelines:

```
skill-name/
├── SKILL.md          # Main instructions + metadata
├── actions/          # Step-by-step procedures
│   ├── create.md
│   └── review.md
└── standards/        # Guidelines and conventions
    ├── naming.md
    └── checklist.md
```

**Use when:** Skills with multiple actions, team conventions, or compliance requirements.

### Full Structure

For complex skills with executable code:

```
skill-name/
├── SKILL.md          # Overview and navigation
├── actions/          # Procedures
│   └── ...
├── standards/        # Guidelines
│   └── ...
├── scripts/          # Executable code
│   ├── validate.py
│   └── process.sh
├── references/       # Detailed documentation
│   ├── api.md
│   └── examples.md
└── assets/           # Templates, data files
    ├── template.json
    └── schema.yaml
```

**Use when:** Skills with utility scripts, extensive reference material, or templates.

## Token Budget

| Content | Budget | Location |
|---------|--------|----------|
| Metadata | ~100 tokens | Frontmatter (loaded at startup) |
| Instructions | <5000 tokens | SKILL.md body (loaded when activated) |
| Resources | As needed | Separate files (loaded on demand) |

**Rule:** Keep SKILL.md body under 500 lines.

## File References

Reference other files using relative paths from skill root:

```markdown
See [the reference guide](references/api.md) for details.

Run the validation script:
scripts/validate.py
```

### One Level Deep

Keep references one level deep from SKILL.md. Agents may partially read deeply nested files.

**Bad — Too Deep:**
```
SKILL.md → advanced.md → details.md → actual-info.md
```

**Good — One Level:**
```
SKILL.md → advanced.md
SKILL.md → details.md
SKILL.md → api-reference.md
```

### Table of Contents

For reference files over 100 lines, include a table of contents:

```markdown
# API Reference

## Contents
- Authentication and setup
- Core methods (create, read, update, delete)
- Advanced features
- Error handling
- Code examples

## Authentication and setup
...
```

## Directory Guidelines

### scripts/

Contains executable code that agents can run:
- Self-contained or clearly document dependencies
- Include helpful error messages
- Handle edge cases gracefully

### references/

Contains additional documentation loaded on demand:
- `REFERENCE.md` — Detailed technical reference
- Domain-specific files (`finance.md`, `legal.md`)
- Keep individual files focused

### assets/

Contains static resources:
- Templates (document, configuration)
- Data files (lookup tables, schemas)
- Images (diagrams, examples)

### actions/

Contains step-by-step procedures (convention from this repo):
- One file per action/workflow
- Clear purpose and sequential steps
- Success criteria per step

### standards/

Contains guidelines and conventions (convention from this repo):
- One file per topic
- Specific, measurable rules
- Good and bad examples

## File Naming

| File Type | Convention |
|-----------|------------|
| Skill manifest | `SKILL.md` (uppercase) |
| Action files | Kebab-case `.md` |
| Standards files | Kebab-case `.md` |
| Scripts | Kebab-case or snake_case |
| References | Kebab-case `.md` |

**Always use forward slashes** in file paths, even on Windows:
- ✅ `scripts/helper.py`
- ❌ `scripts\helper.py`
