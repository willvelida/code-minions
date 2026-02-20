---
description: Standards for agent skill file structure and YAML frontmatter
applyTo: "**/SKILL.md,**/*.agent.md"
---

# Agent Skill File Standards

When creating or editing agent skill files:

- Every SKILL.md must include YAML frontmatter with `name`, `description`, and `version`
- Action files belong in an `actions/` subdirectory and follow the verb-noun naming convention
- Standards files belong in a `standards/` subdirectory
- Keep the skill description under 160 characters for display compatibility
- Reference actions and standards using relative paths from the SKILL.md location
- Include a checklist standard (`standards/checklist.md`) for quality validation
