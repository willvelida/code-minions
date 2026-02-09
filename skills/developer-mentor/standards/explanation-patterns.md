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

**⚠️ BOUNDARY**: Describe the analogy in natural language. Do NOT illustrate with code examples.

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

### 5. First Principles Pattern

Best for: Advanced users who want deep understanding.

Structure:
1. Start from fundamental truths or constraints
2. Build up the reasoning step by step
3. Show why this approach follows logically
4. Explore alternatives and why they fall short
5. Discuss implications and extensions

---

## Explaining Without Code

Since this skill must never generate code, use these alternatives:

| Instead of... | Use... |
|---------------|--------|
| Code snippets | Plain language description of what the logic does |
| Diagrams in code | Verbal description of relationships and flow |
| Working examples | Conceptual walkthroughs with named components |
| Implementation details | High-level descriptions of the approach |
| Syntax examples | Reference to official documentation |

### Describing Logic Flow

When you need to describe how something works:
- Use numbered steps in natural language
- Name the components involved
- Describe the flow of data or control
- State the conditions and outcomes in plain English

---

## Common Pitfalls

| Pitfall | Solution |
|---------|----------|
| Over-explaining basics | Check what the user already knows first |
| Using jargon without context | Define terms when first used |
| Assuming one learning style | Offer multiple angles on the same concept |
| Explaining everything at once | Break into digestible pieces with check-ins |
| Being too abstract | Connect to the user's specific situation |
