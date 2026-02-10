# Writing Effective Descriptions

## Purpose

The `description` field enables skill discovery. Agents use it to choose the right skill from potentially 100+ available skills. A good description provides enough detail to know when to select the skill.

## Rules

| Rule | Requirement |
|------|-------------|
| Length | 1-1024 characters |
| Voice | Third person ("Processes files" not "I process files") |
| Content | Must describe what the skill does AND when to use it |
| Keywords | Include specific trigger terms for discovery |

## Template

```
<What the skill does — capabilities>. Use when <trigger conditions — when to activate>.
```

## Voice

Always write in third person. The description is injected into the system prompt, and inconsistent point-of-view causes discovery problems.

| Voice | Example | Status |
|-------|---------|--------|
| Third person | "Processes Excel files and generates reports" | ✅ Good |
| First person | "I can help you process Excel files" | ❌ Avoid |
| Second person | "You can use this to process Excel files" | ❌ Avoid |

## Be Specific

Include both what the skill does and specific triggers/contexts for when to use it.

### Good Examples

**PDF Processing:**
```
Extract text and tables from PDF files, fill forms, merge documents. Use when working with PDF files or when the user mentions PDFs, forms, or document extraction.
```

**Excel Analysis:**
```
Analyze Excel spreadsheets, create pivot tables, generate charts. Use when analyzing Excel files, spreadsheets, tabular data, or .xlsx files.
```

**Git Commit Helper:**
```
Generate descriptive commit messages by analyzing git diffs. Use when the user asks for help writing commit messages or reviewing staged changes.
```

**DevContainer Creation:**
```
Create and review DevContainer configurations that follow organisational standards. Use when a user asks to set up a dev container, configure a development environment, create a devcontainer.json, add lifecycle hooks, review an existing DevContainer for compliance, or improve container security.
```

### Bad Examples

| Description | Problem |
|-------------|---------|
| `Helps with documents` | Too vague, no specific capability |
| `Processes data` | No indication of what kind of data |
| `Does stuff with files` | Meaningless |
| `Processes PDFs` | Missing when to use |
| `A PDF skill` | No capability or trigger |

## Keywords

Include terms that agents will search for when matching tasks to skills:

| Skill Type | Keywords to Include |
|------------|---------------------|
| PDF processing | PDF, forms, extract, merge, document |
| Code review | review, pull request, PR, changes, diff |
| Data analysis | analyze, spreadsheet, Excel, CSV, data |
| DevContainers | devcontainer, development environment, container |

## Length Guidelines

- **Minimum:** Enough to convey capability + triggers (~50 characters)
- **Maximum:** 1024 characters
- **Sweet spot:** 100-300 characters

Too short lacks discovery keywords. Too long wastes tokens loaded at startup.

## Testing Your Description

Ask yourself:
1. Would an agent know when to use this skill?
2. Does it include the key terms users would mention?
3. Is it specific enough to distinguish from similar skills?
4. Is it written in third person?
