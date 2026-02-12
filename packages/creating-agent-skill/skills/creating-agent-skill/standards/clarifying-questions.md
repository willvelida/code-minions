# Clarifying Questions

## Purpose

Framework for gathering complete, unambiguous requirements from users before generating an Agent Skill. Asking the right questions upfront avoids rework and produces higher-quality skills.

---

## Question Categories

Ask questions in this order. Each category builds on earlier answers.

### 1. Purpose and Scope

Establish what the skill does and its boundaries.

| Question | Why It Matters |
|----------|---------------|
| What task or workflow should this skill help agents perform? | Defines core capability |
| What is out of scope for this skill? | Prevents scope creep |
| Does a similar skill already exist that this extends or replaces? | Avoids duplication |

### 2. Trigger Conditions

Determine when an agent should activate this skill.

| Question | Why It Matters |
|----------|---------------|
| What words or phrases would a user say that should trigger this skill? | Drives description keywords |
| In what context or situation should an agent reach for this skill? | Clarifies activation conditions |
| Are there situations where this skill should NOT be used? | Defines negative triggers |

### 3. Actions and Workflows

Understand the procedures the skill needs to cover.

| Question | Why It Matters |
|----------|---------------|
| What are the distinct actions or workflows this skill should support? | Determines action files needed |
| For each action, what are the high-level steps? | Shapes action file structure |
| Which steps require user input or decisions? | Identifies stop points |

### 4. Standards and Constraints

Identify rules, conventions, or quality criteria.

| Question | Why It Matters |
|----------|---------------|
| Are there naming conventions, style rules, or formats to follow? | Drives standards files |
| Are there things the skill should always or never do? | Defines guardrails |
| How should quality be measured for this skill's output? | Shapes checklist criteria |

### 5. Complexity and Structure

Determine the appropriate skill structure.

| Question | Why It Matters |
|----------|---------------|
| Is this a simple single-purpose skill or a multi-action skill? | Determines structure pattern |
| Will the skill need scripts, templates, or reference material? | Identifies additional directories |
| How much domain knowledge does the agent need beyond its training? | Gauges instruction depth |

### 6. Integration and Dependencies

Identify external requirements.

| Question | Why It Matters |
|----------|---------------|
| Does this skill depend on specific tools being available? | Populates `allowed-tools` |
| Does it reference external APIs, services, or documentation? | Identifies references needed |
| Should this skill work alongside other skills? | Considers interactions |

---

## Handling Incomplete Answers

### Vague Answers

When a user gives a vague response, probe with a specific follow-up:

| Vague Answer | Follow-Up |
|-------------|-----------|
| "It should help with testing" | "What kind of testing — unit tests, integration tests, end-to-end? For which languages or frameworks?" |
| "General code quality" | "Which aspects — formatting, linting, complexity, security, naming conventions?" |
| "It should be flexible" | "Can you give an example of two different ways someone might use it?" |

### Overly Broad Scope

When a skill tries to do too much:

| Signal | Response |
|--------|----------|
| More than 4-5 distinct actions | "This covers a lot of ground. Could we split it into two focused skills?" |
| Unrelated capabilities grouped together | "These seem like separate concerns. Which is the primary focus?" |
| Description exceeding 500 characters | "The description is getting long. What's the core capability in one sentence?" |

### Missing Information

When key details are absent, ask directly rather than assuming:

| Missing Detail | Question |
|----------------|----------|
| No trigger conditions | "When should an agent choose this skill over doing the task without it?" |
| No scope boundaries | "What should this skill explicitly NOT do?" |
| No quality criteria | "How would you know if the skill produced a good result versus a bad one?" |

---

## Sufficiency Check

Before proceeding to generation, confirm all of the following:

- [ ] Skill purpose can be stated in one clear sentence
- [ ] At least 3 trigger keywords identified
- [ ] Scope boundaries defined (what's in and what's out)
- [ ] Actions/workflows listed with high-level steps
- [ ] Structure complexity determined (minimal, standard, or full)
- [ ] User has confirmed understanding is correct

**Template for confirmation:**

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

**🛑 STOP**: Wait for user confirmation before generating the skill.
