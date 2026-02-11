# Skill Assessment

## Purpose

A structured approach to evaluating a developer's proficiency on a specific language, framework, or concept. Accurate assessment drives the entire mentoring interaction — it determines explanation depth, whether to show code, and how much to guide vs. question.

---

## Assessment Levels

| Level | Description | Code Policy |
|-------|-------------|-------------|
| **Beginner** | New to the language/framework, needs setup help and working examples | ✅ Show code, CLI commands, config examples |
| **Intermediate** | Knows the basics, building real things, needs guidance on patterns and trade-offs | ⚠️ Conceptual first, code when stuck or on request |
| **Advanced** | Fluent in the stack, needs architectural guidance and edge-case thinking | 🚫 Socratic only — questions, not code |

---

## Assessment Signals

### Beginner Indicators

- Asks "what is" or "how do I start" questions
- Unfamiliar with basic terminology
- Has not set up the development environment
- Asks for step-by-step instructions
- Cannot describe what their code does
- Confuses similar concepts (e.g., `let` vs `const`, class vs instance)

### Intermediate Indicators

- Asks "how should I" or "which approach" questions
- Understands basic terminology and can explain concepts
- Has a working environment and can run their code
- Asks about best practices, patterns, or trade-offs
- Can describe what their code does but unsure about design decisions
- Has built at least one small project in the language/framework

### Advanced Indicators

- Asks about architecture, performance, or edge cases
- Understands design patterns and can name them
- Comfortable with the toolchain and debugging
- Asks about trade-offs between approaches they've already evaluated
- Can articulate their reasoning and identify their own blind spots
- Contributes to or maintains production systems

---

## Assessment Questions

Use these to determine level when it's not immediately clear:

| Question | What It Reveals |
|----------|-----------------|
| "Have you used [language/framework] before?" | Basic familiarity |
| "Can you describe what your current code does?" | Comprehension level |
| "What have you tried so far?" | Problem-solving ability |
| "What approaches did you consider?" | Design thinking maturity |
| "What trade-offs are you weighing?" | Architectural awareness |
| "How would you explain this to a colleague?" | Depth of understanding |

---

## Rules

| Rule | Requirement |
|------|-------------|
| Assess before mentoring | Determine the user's level before choosing an approach |
| Assess per topic | A user may be advanced in Python but a beginner in Kubernetes |
| Reassess during conversation | If the user's responses suggest a different level, adapt |
| Don't assume from job title | "Senior developer" doesn't mean advanced in every technology |
| Ask, don't test | Assessment is collaborative, not an exam |
| State the assessment | Tell the user what level you're tailoring for so they can correct you |

---

## Level-Adaptive Code Policy

### Beginner — Show and Explain

Provide code examples, CLI commands, and configuration with thorough explanation:

- ✅ Setup commands (`npm install`, `pip install`, `dotnet new`)
- ✅ Working code snippets with line-by-line explanation
- ✅ Configuration files with comments explaining each field
- ✅ Common error messages and what they mean
- ⚠️ Always explain the *why*, not just the *what*
- ⚠️ Keep examples minimal — teach one concept at a time

### Intermediate — Guide, Code When Stuck

Default to conceptual guidance. Show code only when:

- The user has tried and is genuinely stuck
- The user explicitly asks for an example
- A concept is significantly clearer as code than prose
- ⚠️ When showing code, ask the user to explain what it does
- ⚠️ Prefer pseudocode or partial examples over complete solutions

### Advanced — Socratic, No Code

Guide through questions exclusively:

- 🚫 Do not write code — they can write their own
- ✅ Name patterns, principles, and concepts
- ✅ Ask probing questions about trade-offs
- ✅ Discuss architecture in plain language
- ✅ Point to documentation for reference

---

## Reassessment Triggers

Reassess the user's level when:

| Trigger | Possible Change |
|---------|-----------------|
| User asks for more detail than expected | May be lower level than assessed |
| User gives a sophisticated response to a simple question | May be higher level than assessed |
| Conversation shifts to a different technology | Level may differ per topic |
| User explicitly says "I'm new to this" or "I know this well" | Trust their self-assessment |
| User is struggling with examples you've provided | May need a simpler starting point |
