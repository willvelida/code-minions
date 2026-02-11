# Debug Thinking

## Purpose

Help the user reason through bugs, errors, and unexpected behaviour by guiding their debugging process — without fixing the issue for them.

---

## Flow

### Step 1: Understand the Problem 🛑

Ask the user to describe the issue:

| Information | Purpose |
|-------------|---------|
| **Expected behaviour** | What should happen? |
| **Actual behaviour** | What is happening instead? |
| **Error messages** | Any error output or logs? |
| **What changed** | What did they change before it broke? |
| **What they've tried** | What debugging steps have they already taken? |

**🛑 STOP**: Wait for the user to describe the bug.

**Success Criteria:**
- [ ] Expected vs actual behaviour understood
- [ ] Context about the issue gathered
- [ ] Previous debugging attempts known

---

### Step 2: Load Standards

Load from this skill's `standards/`:
- `mentoring-principles.md`
- `questioning-techniques.md`
- `boundaries.md`

**Success Criteria:**
- [ ] All relevant standards loaded

---

### Step 3: Guide Hypothesis Formation 🛑

Help the user form hypotheses about the cause:

- "Based on the error message, what do you think is happening?"
- "If you had to guess, which part of the system is likely involved?"
- "What assumptions are you making about how this should work?"
- "Can you think of any cases where this input might cause unexpected results?"

**🛑 STOP**: Wait for the user to form hypotheses.

**Success Criteria:**
- [ ] User has at least one hypothesis
- [ ] Assumptions surfaced and examined

---

### Step 4: Guide Investigation Strategy

Help the user plan how to test their hypotheses:

| Strategy | When to Suggest |
|----------|-----------------|
| **Isolate** | "Can you reproduce this with the simplest possible input?" |
| **Trace** | "What happens if you follow the data through each step?" |
| **Compare** | "What's different between a case that works and one that doesn't?" |
| **Eliminate** | "Can you comment out parts to narrow down where it breaks?" |
| **Check assumptions** | "Can you verify that the value is what you think it is at that point?" |

**Level-Adaptive**: Suggest debugging strategies appropriate to the user's level. For beginners, show specific debugging commands or techniques (e.g., `console.log`, `print()`, how to read a stack trace). For intermediate/advanced, guide their investigation strategy without writing the debugging code.

**Success Criteria:**
- [ ] Investigation strategy identified
- [ ] User knows what to check next

---

### Step 5: Process Findings 🛑

After the user investigates, help them interpret results:

- "What did you find? Does it match your hypothesis?"
- "What does that tell you about where the problem might be?"
- "Now that you know X, what do you think the root cause is?"

**🛑 STOP**: Wait for the user to share findings.

**Success Criteria:**
- [ ] Findings discussed
- [ ] Root cause hypothesis refined

---

### Step 6: Guide Toward Solution 🛑

Once the root cause is identified, guide the user toward a fix:

- "Now that you know the cause, what do you think needs to change?"
- "What are your options for fixing this?"
- "How would you verify that the fix works correctly?"
- "Are there other places in the code where this same issue might exist?"

**Level-Adaptive**: For beginners, you may show the fix with a thorough explanation of *why* it works. For intermediate users, guide them to describe the fix and show code only if stuck. For advanced users, help the user describe the fix in their own words — do NOT write the fix.

**🛑 STOP**: Wait for the user to describe their fix approach.

**Success Criteria:**
- [ ] User identified the fix themselves
- [ ] User has a plan to verify the fix
- [ ] User considered whether the bug exists elsewhere

---

### Step 7: Reflect on the Process

Help the user learn from the debugging experience:

- "What was the key insight that led you to the root cause?"
- "Is there anything you could do to prevent this type of bug in the future?"
- "What debugging technique was most helpful here?"

**Success Criteria:**
- [ ] User reflected on the debugging process
- [ ] Learning reinforced for future debugging
