# Assess Level

## Purpose

Evaluate a developer's proficiency on a specific language, framework, or concept to determine the appropriate mentoring approach and code policy.

---

## Flow

### Step 1: Identify the Topic 🛑

Ask the user what they need help with:

| Information | Purpose |
|-------------|---------|
| **Language or framework** | What technology are they asking about? |
| **Specific area** | Is it setup, concepts, architecture, debugging? |
| **Context** | Are they starting a project, learning, or solving a problem? |

**🛑 STOP**: Wait for the user to describe their topic and context.

**Success Criteria:**
- [ ] Topic clearly identified
- [ ] Context understood

---

### Step 2: Load Standards

Load from this skill's `standards/`:
- `skill-assessment.md`
- `mentoring-principles.md`

**Success Criteria:**
- [ ] Assessment criteria loaded

---

### Step 3: Gather Signals

Ask 2-3 targeted questions to determine the user's level. Choose from:

**For unknown familiarity:**
- "Have you used [technology] before?"
- "What's your experience with [technology] so far?"

**For apparent beginners:**
- "Do you have [technology] installed and set up?"
- "Have you written any [language] code before?"

**For apparent intermediate/advanced:**
- "What have you tried so far?"
- "What approaches are you considering?"
- "What trade-offs are you weighing?"

Do not rapid-fire all questions. Ask one, listen, then follow up based on the response.

**Success Criteria:**
- [ ] At least 2 signal questions asked
- [ ] Responses analysed against assessment criteria

---

### Step 4: Determine Level and State It 🛑

Based on signals, classify the user's level on **this specific topic**:

| Level | Code Policy | Mentoring Style |
|-------|-------------|-----------------|
| **Beginner** | Show code, CLI commands, config | Explain step-by-step, provide examples |
| **Intermediate** | Conceptual first, code when stuck | Guide thinking, discuss trade-offs |
| **Advanced** | Socratic, no code | Ask questions, explore edge cases |

**🛑 STOP**: State your assessment to the user:

> "Based on what you've shared, it sounds like you're [level] with [topic]. I'll tailor my approach accordingly — [brief description of what that means]. Does that feel right, or would you like me to adjust?"

Wait for confirmation. If the user disagrees, adjust.

**Success Criteria:**
- [ ] Level determined per the skill-assessment standard
- [ ] Assessment communicated to the user
- [ ] User confirmed or corrected the assessment

---

### Step 5: Transition to Action

Based on the user's original request, route to the appropriate action:

| User Need | Action |
|-----------|--------|
| "What is X?" / "Explain Y" | `explain-concept.md` |
| "How should I design/build this?" | `guide-design.md` |
| "My code isn't working" | `debug-thinking.md` |
| "Is my approach good?" | `review-approach.md` |
| "What should I learn?" | `recommend-learning.md` |

Carry the assessed level into the next action — it determines explanation depth and code policy.

**Success Criteria:**
- [ ] Appropriate action identified
- [ ] Assessed level carried forward
