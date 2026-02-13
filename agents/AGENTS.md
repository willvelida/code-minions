# Code Minions

## Purpose

This repository is a toolkit of various different AI-Assisted development capabilities that I use to enhance my development.

It includes standards, Agent Skills, and Agent configurations that have helped me during my software development.

These agents provides a structured approach to AI-assisted development through specialized agents and skills. The system emphasizes isolated context for specific tasks, ensuring optimal token usage and focused expertise.

## Philosophy

- **Specialisation** — Each agent focuses on a specific domain, providing deep expertise
- **Isolation** — Subagents operate with isolated context, reducing noise and improving accuracy
- **Skills** — Self-contained capabilities that agents can invoke, bundling actions with relevant standards

## Structure

```bash
code-minions/
├── AGENTS.md                 # This file - system overview and routing
├── agents/                   # Agent definitions
└── skills/                   # Self-contained skill packages
    ├── creating-agent-skill/ # Skill for creating new agent skills
    ├── creating-agents/      # Agent definition creation skill
    ├── creating-devcontainers/ # DevContainer creation skill
    ├── creating-documentation/ # Documentation creation skill
    ├── developer-mentor/     # Developer mentoring skill
    ├── git-workflow/         # Git workflow skill
    ├── raise-pull-requests/  # Pull request skill
    └── threat-modelling/     # Threat modelling skill
        ├── SKILL.md          # Skill manifest (each skill follows this structure)
        ├── actions/          # Executable actions
        └── standards/        # Skill-specific standards
```

### Skill-Bundled Standards

Skills will typically bundle their own generalised standards in `skills/<skill>/standards/`. Each skill will document how it manages precedence across the skill's generalised standards and the project specific standards.

## Routing

When a query matches a specific domain, delegate to the appropriate subagent for isolated, focused processing.

| Domain              | Keywords                                                                                  | Subagent                    | Skill                            |
|---------------------|-------------------------------------------------------------------------------------------|-----------------------------|----------------------------------|
| Agent Skills        | agent skill, create skill, SKILL.md, skill specification, review skill                    | `Agent-Skill-Agent`         | `skills/creating-agent-skill/`   |
| Agents              | agent, agents.md, agent definition, persona, custom agent, copilot agent                  | `Agent-Agent`               | `skills/creating-agents/`        |
| DevContainer        | devcontainer, dev container, .devcontainer, development container, devcontainer.json       | `DevContainer-Agent`        | `skills/creating-devcontainers/` |
| Documentation       | README, documentation, CONTRIBUTING, docs, writing, markdown                              | `Documentation-Agent`       | `skills/creating-documentation/` |
| Developer Mentor    | explain, mentor, teach, concept, learning, design guidance, debug thinking, approach       | `Developer-Mentor-Agent`    | `skills/developer-mentor/`       |
| Git Workflow        | git, branch, commit, merge, rebase, release, hotfix, conventional commit                  | `Git-Workflow-Agent`        | `skills/git-workflow/`           |
| Pull Requests       | pull request, PR, code review, PR description, submit PR, review feedback                 | `Pull-Request-Agent`        | `skills/raise-pull-requests/`    |
| Threat Modelling    | threat model, STRIDE, security, attack surface, threat assessment, data flow diagram      | `Threat-Modelling-Agent`    | `skills/threat-modelling/`       |

### How to Route

1. Identify the primary domain of the user's query
2. Match against the keywords in the routing table
3. Delegate to the specified subagent
4. The subagent will load its skill manifest (`SKILL.md`) and relevant standards

### When NOT to Route

- General questions that don't match a specific domain
- Questions spanning multiple domains (handle at orchestrator level)
- Simple queries that don't require specialist knowledge