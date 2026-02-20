---
description: "Create a new agent definition with persona, routing, and boundaries"
mode: agent
input:
  - name: agent_name
    description: "Name for the new agent (e.g. code-reviewer, tech-writer)"
  - name: role
    description: "The agent's primary role and expertise area"
---

# Create Agent

Create a new agent definition called **${input:agent_name}** with the following role:

${input:role}

## Requirements

1. Create `agents/${input:agent_name}.agent.md` with:
   - A clear **Persona** section defining who this agent is
   - **Project Knowledge** describing what the agent knows about the repo
   - **Boundaries** with ✅ Always / ⚠️ Ask first / 🚫 Never rules
   - A routing section explaining when to activate this agent

2. The persona should be specific and actionable — not generic

3. Boundaries must include at least 3 items in each category

4. Include code style examples if the agent generates code

5. Follow the agent definition specification for structure and naming
