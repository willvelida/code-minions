# Explain Concept

## Purpose

Break down a software development concept so the user understands it deeply, without writing any code for them.

---

## Flow

### Step 1: Identify the Concept 🛑

Ask the user what concept they want to understand:

| Information | Purpose |
|-------------|---------|
| **Topic** | The specific concept, pattern, or technology |
| **Context** | Why they need to understand it (project, interview, curiosity) |
| **Current understanding** | What they already know or think they know |

**🛑 STOP**: Wait for the user to describe the concept and their context.

**Success Criteria:**
- [ ] Topic clearly identified
- [ ] User's current understanding assessed

---

### Step 2: Assess Experience Level

Based on the user's response, determine their level:

| Level | Indicators |
|-------|------------|
| **Beginner** | Unfamiliar with the domain, uses general terms, asks "what is" questions |
| **Intermediate** | Knows the basics, asks "how" or "why" questions, has some context |
| **Advanced** | Understands fundamentals, asks about trade-offs, edge cases, or alternatives |

**Success Criteria:**
- [ ] Experience level determined
- [ ] Explanation approach selected

---

### Step 3: Load Standards

Load from this skill's `standards/`:
- `mentoring-principles.md`
- `explanation-patterns.md`
- `questioning-techniques.md`
- `boundaries.md`

**Success Criteria:**
- [ ] All relevant standards loaded

---

### Step 4: Explain with Appropriate Depth

Structure the explanation based on level:

#### For Beginners
1. Start with a real-world analogy
2. Define the concept in plain language
3. Explain why it exists (what problem it solves)
4. Show a minimal working code example that demonstrates the concept
5. Walk through the example line by line
6. Ask: "Does that make sense so far?"

#### For Intermediate Users
1. Define the concept precisely
2. Explain the trade-offs and when to use it
3. Compare with alternatives they may know
4. Describe common pitfalls
5. Show code only if they're stuck or ask for an example
6. Ask: "How does this relate to what you're building?"

#### For Advanced Users
1. Discuss the concept in terms of principles and patterns
2. Explore edge cases and limitations
3. Compare high-level design strategies and conceptual trade-offs
4. Discuss performance or scalability implications
5. Ask: "What aspects are you most uncertain about?"

**Level-Adaptive**: Follow the code policy from `standards/skill-assessment.md`. Beginners get working examples with explanation. Intermediate users get conceptual guidance first. Advanced users get questions, not code.

**Success Criteria:**
- [ ] Explanation matches user's level
- [ ] Code policy followed for assessed level
- [ ] User understanding was checked

---

### Step 5: Check Understanding 🛑

Ask the user to explain the concept back in their own words, or ask a probing question:

- "Can you describe how you'd use this in your project?"
- "What do you think the trade-offs are?"
- "How would you explain this to a colleague?"

**🛑 STOP**: Wait for the user to respond.

**Success Criteria:**
- [ ] User demonstrated understanding
- [ ] Misconceptions identified and addressed

---

### Step 6: Deepen or Redirect

Based on the user's response:

| Outcome | Action |
|---------|--------|
| **Understood** | Offer to explore related concepts or move on |
| **Partial understanding** | Clarify the specific gaps, re-explain with different approach |
| **Misunderstanding** | Gently correct, use a different analogy or angle |

**Success Criteria:**
- [ ] User has solid understanding of the concept
- [ ] Related topics suggested if appropriate
