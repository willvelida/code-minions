# Agent Definition Checklist

## Purpose

Consolidated checklist for validating agent definitions against quality guidelines and best practices from analysis of 2,500+ repositories.

---

## Requirements Gathering

- [ ] Clarifying questions asked before generation began
- [ ] Agent purpose stated in one clear sentence
- [ ] Target domain or task identified
- [ ] Tech stack documented with versions
- [ ] File structure of the target project understood
- [ ] Requirements confirmed by user

---

## Frontmatter

- [ ] `name` field present
- [ ] `name` is kebab-case (lowercase, hyphens only)
- [ ] `description` field present
- [ ] `description` states what the agent does in one sentence

---

## Persona

- [ ] Clear role stated ("You are a [specific role]")
- [ ] Specialisations listed (2-4 specific skills/frameworks)
- [ ] Output type defined (what the agent produces)
- [ ] Single responsibility (one persona, one job)
- [ ] Uses project-specific language, not generic terms

---

## Project Knowledge

- [ ] Tech stack listed with versions
- [ ] File structure documented with actual paths
- [ ] Key dependencies mentioned
- [ ] Read/write scope defined (where agent reads, where it writes)

---

## Commands and Tools

- [ ] Commands section appears early in the document
- [ ] At least one build command included
- [ ] At least one test command included
- [ ] At least one lint or format command included
- [ ] Commands include flags and options (not just tool names)
- [ ] Each command has a purpose label
- [ ] Commands are copy-pasteable and executable

---

## Code Style

- [ ] Naming conventions documented (functions, classes, files)
- [ ] At least one good code example provided
- [ ] At least one bad code example provided (contrast)
- [ ] Examples use the project's actual language and frameworks
- [ ] Error handling patterns shown

---

## Git Workflow (If Applicable)

- [ ] Branch naming conventions stated
- [ ] Commit message format defined
- [ ] PR process described

---

## Boundaries

- [ ] ✅ Always-do tier present with specific actions
- [ ] ⚠️ Ask-first tier present with approval gates
- [ ] 🚫 Never-do tier present with hard restrictions
- [ ] "Never commit secrets or API keys" included
- [ ] Read/write scope explicitly defined with file paths
- [ ] No production-affecting actions without approval
- [ ] Boundaries are specific (file paths, not vague guidance)

---

## Six Core Areas Coverage

- [ ] Commands — Executable commands with flags
- [ ] Testing — Test framework, commands, coverage expectations
- [ ] Project structure — File layout and what lives where
- [ ] Code style — Naming, patterns, concrete examples
- [ ] Git workflow — Branch naming, commits, PR process
- [ ] Boundaries — What the agent must never touch

---

## Quality Checks

- [ ] Document is scannable (tables, bullets, headers — not long paragraphs)
- [ ] Commands appear before style guides and boundaries
- [ ] Code examples over prose explanations
- [ ] No vague language ("be careful", "follow best practices")
- [ ] Tech stack includes versions
- [ ] File paths are real and specific
- [ ] No time-sensitive information (or properly isolated)

---

## Anti-Patterns Avoided

- [ ] No "helpful assistant" persona (too vague)
- [ ] No multiple roles in one agent (split into separate agents)
- [ ] No tool names without full commands
- [ ] No prose-only style guidance (needs code examples)
- [ ] No missing boundary tiers
- [ ] No generic file paths ("the config file" — specify which one)
- [ ] No overly long paragraphs (use structured formatting)
