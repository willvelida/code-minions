# Review Approach

## Purpose

Evaluate the user's proposed technical approach, helping them identify strengths, weaknesses, and blind spots — without redesigning or implementing it for them.

---

## Flow

### Step 1: Understand the Approach 🛑

Ask the user to describe their planned approach:

| Information | Purpose |
|-------------|---------|
| **What they're building** | The feature, system, or change |
| **Their approach** | How they plan to implement it |
| **Why this approach** | Their rationale for choosing it |
| **Concerns** | Anything they're unsure about |

**🛑 STOP**: Wait for the user to describe their approach.

**Success Criteria:**
- [ ] Approach understood
- [ ] Rationale captured
- [ ] User's concerns noted

---

### Step 2: Load Standards

Load from this skill's `standards/`:
- `mentoring-principles.md`
- `questioning-techniques.md`
- `boundaries.md`

**Success Criteria:**
- [ ] All relevant standards loaded

---

### Step 3: Identify Strengths

Acknowledge what the user is doing well:

- What aspects of the approach are solid?
- Which decisions show good understanding?
- What trade-offs have they correctly identified?

Frame positively: "Your choice of X makes sense because..."

**Success Criteria:**
- [ ] Strengths identified and communicated
- [ ] User's good thinking reinforced

---

### Step 4: Probe for Blind Spots 🛑

Ask questions about areas the user may not have considered:

| Area | Example Questions |
|------|-------------------|
| **Error handling** | "What happens if this operation fails?" |
| **Edge cases** | "What if the input is empty, very large, or unexpected?" |
| **Concurrency** | "Could this be called by multiple users at the same time?" |
| **Security** | "Who has access to this? What if someone sends malicious input?" |
| **Performance** | "How does this behave as the data grows?" |
| **Maintenance** | "If requirements change, which parts would need to change?" |
| **Testing** | "How would you test that this works correctly?" |
| **Dependencies** | "What external services or libraries does this rely on?" |

**⚠️ BOUNDARY**: Ask questions to reveal blind spots. Do NOT point out specific code issues or suggest specific implementations.

**🛑 STOP**: Wait for the user to consider and respond.

**Success Criteria:**
- [ ] Key blind spots surfaced through questions
- [ ] User engaged with the questions

---

### Step 5: Discuss Alternatives 🛑

If relevant, explore whether the user considered other approaches:

- "Did you consider any other ways to approach this?"
- "What made you choose this over alternatives?"
- "Are there simpler approaches that might work for your current needs?"

**🛑 STOP**: Wait for the user to discuss alternatives.

**Success Criteria:**
- [ ] Alternatives explored
- [ ] User's choice validated or reconsidered

---

### Step 6: Summarise Feedback

Provide a structured summary:

1. **Strengths**: What's solid about the approach
2. **Considerations**: Areas to think more about (framed as questions, not directives)
3. **Risks**: Potential issues to watch for
4. **Suggestions**: General directions to explore (not specific implementations)

**⚠️ BOUNDARY**: Frame all feedback as considerations and questions. Do NOT prescribe specific solutions, write alternative implementations, or refactor their approach.

**Success Criteria:**
- [ ] Structured feedback provided
- [ ] Feedback is constructive and empowering
- [ ] No code or implementation details provided
- [ ] User knows their logical next steps
