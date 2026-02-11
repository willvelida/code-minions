# Guide Design

## Purpose

Walk the user through architecture and design decisions, helping them evaluate options and choose an approach — without designing or implementing it for them.

---

## Flow

### Step 1: Understand the Problem 🛑

Gather context about what the user is trying to build:

| Information | Purpose |
|-------------|---------|
| **Goal** | What are they trying to achieve? |
| **Constraints** | Timeline, team size, existing tech stack, budget |
| **Scale** | Expected load, data volume, number of users |
| **Current state** | Is this greenfield or modifying existing code? |

**🛑 STOP**: Wait for the user to describe their situation.

**Success Criteria:**
- [ ] Problem clearly understood
- [ ] Constraints identified
- [ ] Scale requirements known

---

### Step 2: Load Standards

Load from this skill's `standards/`:
- `mentoring-principles.md`
- `questioning-techniques.md`
- `boundaries.md`

**Success Criteria:**
- [ ] All relevant standards loaded

---

### Step 3: Explore What They've Considered 🛑

Before offering guidance, understand the user's thinking:

- "What approaches have you considered so far?"
- "What concerns do you have about your current direction?"
- "Are there any patterns or architectures you've been looking at?"

**🛑 STOP**: Wait for the user to share their current thinking.

**Success Criteria:**
- [ ] User's current approach understood
- [ ] Gaps in their thinking identified

---

### Step 4: Present Trade-offs

Help the user evaluate their options by presenting trade-offs:

| For Each Option | Discuss |
|-----------------|---------|
| **Benefits** | What problems does this approach solve well? |
| **Drawbacks** | What are the downsides or risks? |
| **Complexity** | How difficult is this to implement and maintain? |
| **Fit** | How well does this match their constraints? |
| **Evolution** | How does this approach handle future changes? |

**Level-Adaptive**: For beginners, illustrate simple design concepts with code examples (e.g., show a basic API route to explain REST). For intermediate/advanced, describe approaches in plain language. Do NOT generate full architecture implementations at any level.

**Success Criteria:**
- [ ] Multiple options presented with trade-offs
- [ ] No code or implementation details provided
- [ ] Options framed as choices, not directives

---

### Step 5: Guide the Decision 🛑

Ask questions that help the user decide:

- "Given your timeline, which approach feels most manageable?"
- "Which trade-offs are you most comfortable accepting?"
- "What would happen if your scale doubled — which approach handles that better?"
- "Which approach does your team have the most experience with?"

**🛑 STOP**: Wait for the user to make their decision.

**Success Criteria:**
- [ ] User made an informed decision
- [ ] Decision rationale is clear

---

### Step 6: Validate and Plan Next Steps

Once the user has chosen an approach:

1. Summarise their decision and rationale
2. Identify potential risks to watch for
3. Suggest what they should research or prototype first
4. Recommend which components to build in what order
5. Ask if they want to dive deeper into any specific aspect

**Level-Adaptive**: Recommendations should be directional ("start with the data layer") not prescriptive ("create a class called DataService with these methods") — unless the user is a beginner who needs a concrete starting point.

**Success Criteria:**
- [ ] Decision validated
- [ ] Next steps identified
- [ ] User feels confident in their direction
