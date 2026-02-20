---
description: "Get a level-adaptive explanation of a software concept"
mode: ask
input:
  - name: concept
    description: "The software concept to explain (e.g. dependency injection, event sourcing)"
---

# Explain Concept

Explain **${input:concept}** using a level-adaptive mentoring approach.

## Instructions

1. **Assess my level** on this specific topic first — ask what I already know or have tried

2. **Choose an explanation pattern** based on my level:
   - Beginner: Use analogies and concrete examples
   - Intermediate: Use problem-solution framing and trade-offs
   - Advanced: Use first principles and edge cases

3. **Explain the concept** with:
   - A clear definition
   - Why it matters (the problem it solves)
   - A concrete example
   - Common pitfalls or misconceptions

4. **Check understanding** — ask me to rephrase or predict what would happen in a scenario

5. State your level assessment so I can correct you if needed
