---
description: "Scaffold a new agent skill with SKILL.md, actions, and standards"
mode: agent
input:
  - name: skill_name
    description: "Name for the new skill (e.g. code-review, deploy-pipeline)"
  - name: purpose
    description: "What the skill should help the agent do"
---

# Create Agent Skill

Create a new agent skill called **${input:skill_name}** with the following purpose:

${input:purpose}

## Requirements

1. Create the skill directory structure:
   - `skills/${input:skill_name}/SKILL.md` — skill manifest with frontmatter
   - `skills/${input:skill_name}/actions/` — step-by-step action files
   - `skills/${input:skill_name}/standards/` — quality standards and checklists

2. The SKILL.md must include:
   - YAML frontmatter with `name`, `description`, `version`
   - A clear scope section defining what the skill does and doesn't do
   - References to actions and standards

3. Create at least one action file with step-by-step instructions

4. Create a `checklist.md` standard for quality validation

5. Follow the agent skills specification for file naming and structure
