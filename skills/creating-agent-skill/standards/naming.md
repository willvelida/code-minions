# Naming Conventions

## Purpose

Consistent naming makes skills easier to discover, reference, and maintain. These conventions apply to skill names, directory names, and file names.

## Skill Name Rules

| Rule | Requirement |
|------|-------------|
| Characters | Lowercase letters (`a-z`), numbers (`0-9`), hyphens (`-`) only |
| Length | 1-64 characters |
| Start/End | Must not start or end with hyphen |
| Consecutive | Must not contain consecutive hyphens (`--`) |
| Reserved | Must not contain `anthropic` or `claude` |
| Match | Must match parent directory name exactly |

## Recommended Style

Use **gerund form** (verb + -ing) for skill names — this clearly describes the activity or capability.

### Good Examples (Gerund Form)

- `processing-pdfs`
- `analyzing-spreadsheets`
- `managing-databases`
- `testing-code`
- `writing-documentation`
- `creating-devcontainers`
- `creating-agent-skill`

### Acceptable Alternatives

- Noun phrases: `pdf-processing`, `spreadsheet-analysis`
- Action-oriented: `process-pdfs`, `analyze-spreadsheets`

### Avoid

| Pattern | Problem | Example |
|---------|---------|---------|
| Vague names | Unclear purpose | `helper`, `utils`, `tools` |
| Overly generic | Too broad | `documents`, `data`, `files` |
| Reserved words | Specification violation | `anthropic-helper`, `claude-tools` |
| Uppercase | Invalid characters | `PDF-Processing` |
| Underscores | Invalid characters | `pdf_processing` |
| Spaces | Invalid characters | `pdf processing` |

## Directory Names

| Directory | Convention | Examples |
|-----------|------------|----------|
| Skill root | Match `name` field exactly | `creating-agent-skill/` |
| Actions | `actions/` | `actions/create.md` |
| Standards | `standards/` | `standards/naming.md` |
| Scripts | `scripts/` | `scripts/validate.py` |
| References | `references/` | `references/api.md` |
| Assets | `assets/` | `assets/template.json` |

## File Names

| File Type | Convention | Examples |
|-----------|------------|----------|
| Main skill file | `SKILL.md` (uppercase) | `SKILL.md` |
| Action files | Kebab-case `.md` | `create.md`, `review-compliance.md` |
| Standards files | Kebab-case `.md` | `naming.md`, `self-review-checklist.md` |
| Scripts | Kebab-case or snake_case | `validate.py`, `extract_data.py` |
| Reference docs | Kebab-case `.md` or uppercase | `api-reference.md`, `REFERENCE.md` |

## Consistency Benefits

Consistent naming:
- Makes skills easier to reference in documentation
- Helps agents understand skill purpose at a glance
- Simplifies searching and organising skills
- Maintains a professional, cohesive skill library

## Examples

### Valid Skill Names

```
processing-pdfs
data-analysis
code-review
git-workflow
creating-devcontainers
raise-pull-requests
```

### Invalid Skill Names

```
PDF-Processing      # Uppercase not allowed
-pdf-processing     # Cannot start with hyphen
pdf--processing     # Consecutive hyphens not allowed
pdf_processing      # Underscores not allowed
claude-helper       # Reserved word
helper              # Too vague
```
