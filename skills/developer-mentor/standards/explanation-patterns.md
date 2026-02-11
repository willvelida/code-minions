# Explanation Patterns

## Purpose

Patterns and approaches for explaining software development concepts at different experience levels, without writing code.

---

## Choosing an Explanation Approach

| User Level | Primary Approach | Supporting Techniques |
|------------|------------------|----------------------|
| Beginner | Analogy-first | Simple language, real-world parallels, step-by-step |
| Intermediate | Concept-and-context | Trade-offs, comparisons, pattern names |
| Advanced | Principle-based | Design rationale, edge cases, theory |

---

## Explanation Patterns

### 1. Analogy Pattern

Best for: Beginners encountering a concept for the first time.

Structure:
1. Start with a familiar real-world scenario
2. Map the real-world elements to the technical concept
3. Explain where the analogy holds
4. Note where the analogy breaks down
5. Transition to technical terminology
6. For beginners: show a minimal working example that demonstrates the concept

**Level-Adaptive**: For beginners, follow the analogy with a code example. For intermediate/advanced, the analogy alone is usually sufficient.

### 2. Problem-Solution Pattern

Best for: Intermediate users who understand the basics but need to see the "why."

Structure:
1. Describe the problem that exists without this concept
2. Show how people struggled with the old approach
3. Explain how this concept solves the problem
4. Discuss what trade-offs the solution introduces
5. Connect to the user's specific situation

### 3. Compare-and-Contrast Pattern

Best for: Users who know a related concept but need to understand differences.

Structure:
1. Start with what they already know
2. Identify similarities between the known and new concept
3. Highlight key differences
4. Explain when to use each
5. Ask the user which fits their situation

### 4. Building Blocks Pattern

Best for: Complex concepts that build on simpler ones.

Structure:
1. Identify the prerequisite concepts
2. Verify the user understands each prerequisite
3. Combine the building blocks one at a time
4. Show how the full concept emerges from the parts
5. Verify understanding of the complete picture

**Level-Adaptive**: For beginners, demonstrate each building block with a small code example before combining. For intermediate/advanced, verbal description of each block is sufficient.

### 5. First Principles Pattern

Best for: Advanced users who want deep understanding.

Structure:
1. Start from fundamental truths or constraints
2. Build up the reasoning step by step
3. Show why this approach follows logically
4. Explore alternatives and why they fall short
5. Discuss implications and extensions

---

## Explaining With and Without Code

The approach depends on the user's assessed level (see `skill-assessment.md`):

### Beginners — Code as Teaching Tool

Code examples help beginners connect concepts to reality:

| Technique | Example |
|-----------|---------|
| Minimal working example | Show the simplest code that demonstrates the concept |
| Line-by-line annotation | Explain what each line does and why |
| Predict-and-verify | Ask "what do you think this will output?" before running |
| Modify-and-observe | Ask them to change one thing and predict the result |

### Intermediate — Conceptual First, Code When Stuck

| Instead of... | Try first... | Then if stuck... |
|---------------|-------------|------------------|
| Full code example | Describe the approach in plain language | Show a partial example |
| Working implementation | Discuss the pattern by name | Show pseudocode |
| Complete solution | Ask what they'd try | Show the tricky part only |

### Advanced — No Code, Concepts Only

| Instead of... | Use... |
|---------------|--------|
| Code snippets | Plain language description of what the logic does |
| Working examples | Conceptual walkthroughs with named components |
| Implementation details | High-level descriptions of the approach |
| Syntax examples | Reference to official documentation |

---

## Common Pitfalls

| Pitfall | Solution |
|---------|----------|
| Over-explaining basics | Check what the user already knows first |
| Using jargon without context | Define terms when first used |
| Assuming one learning style | Offer multiple angles on the same concept |
| Explaining everything at once | Break into digestible pieces with check-ins |
| Being too abstract | Connect to the user's specific situation |
