# Agent Skills Specification

## Purpose

Core format specification for Agent Skills, defining required structure, frontmatter fields, and validation rules.

## Directory Structure

A skill is a directory containing at minimum a `SKILL.md` file:

```
skill-name/
└── SKILL.md          # Required
```

### Optional Directories

| Directory | Purpose | Contents |
|-----------|---------|----------|
| `scripts/` | Executable code | Self-contained scripts with documented dependencies |
| `references/` | Additional documentation | Detailed guides, API references, domain-specific files |
| `assets/` | Static resources | Templates, images, data files, schemas |
| `actions/` | Step-by-step procedures | Action files for workflows (convention from this repo) |
| `standards/` | Guidelines and conventions | Standards files (convention from this repo) |

## SKILL.md Format

The `SKILL.md` file must contain YAML frontmatter followed by Markdown content.

### Frontmatter (Required)

```yaml
---
name: skill-name
description: A description of what this skill does and when to use it.
---
```

### Frontmatter Fields

| Field | Required | Constraints |
|-------|----------|-------------|
| `name` | Yes | 1-64 chars, lowercase alphanumeric + hyphens, must match directory name |
| `description` | Yes | 1-1024 chars, non-empty, describes what AND when |
| `license` | No | License name or reference to bundled license file |
| `compatibility` | No | 1-500 chars, environment requirements if any |
| `metadata` | No | Key-value map for additional properties |
| `allowed-tools` | No | Space-delimited list of pre-approved tools (experimental) |

### Name Field Rules

| Rule | Requirement |
|------|-------------|
| Characters | Lowercase letters (`a-z`), numbers (`0-9`), hyphens (`-`) only |
| Length | 1-64 characters |
| Start/End | Must not start or end with hyphen |
| Consecutive | Must not contain consecutive hyphens (`--`) |
| Reserved | Must not contain `anthropic` or `claude` |
| Match | Must match parent directory name exactly |

### Description Field Rules

| Rule | Requirement |
|------|-------------|
| Length | 1-1024 characters |
| Voice | Third person ("Processes files" not "I process files") |
| Content | Must describe what the skill does AND when to use it |
| Keywords | Should include specific trigger terms for discovery |

## Body Content

The Markdown body after frontmatter contains skill instructions. No format restrictions — write whatever helps agents perform the task effectively.

### Recommended Sections

- Step-by-step instructions
- Examples of inputs and outputs
- Common edge cases
- References to additional files

### Token Budget

- Keep SKILL.md body under 500 lines
- Move detailed reference material to separate files
- Use progressive disclosure for efficient context use

## File References

When referencing other files, use relative paths from the skill root:

```markdown
See [the reference guide](references/REFERENCE.md) for details.

Run the extraction script:
scripts/extract.py
```

**Rule:** Keep file references one level deep from `SKILL.md`. Avoid deeply nested reference chains.

## Validation

Use the [skills-ref](https://github.com/agentskills/agentskills/tree/main/skills-ref) library to validate:

```bash
skills-ref validate ./skill-name
```

This checks frontmatter validity and naming conventions.
