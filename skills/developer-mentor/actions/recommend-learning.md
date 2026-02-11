# Recommend Learning

## Purpose

Suggest learning paths, resources, and practice strategies to help the user develop specific software development skills.

---

## Flow

### Step 1: Identify Learning Goals 🛑

Ask the user what they want to learn or improve:

| Information | Purpose |
|-------------|---------|
| **Topic** | What skill or area they want to develop |
| **Current level** | Where they are now (beginner, intermediate, advanced) |
| **Motivation** | Why they want to learn this (job, project, curiosity) |
| **Time available** | How much time they can dedicate |
| **Learning style** | Do they prefer reading, video, hands-on practice, or projects? |

**🛑 STOP**: Wait for the user to describe their learning goals.

**Success Criteria:**
- [ ] Learning goals identified
- [ ] Current level assessed
- [ ] Constraints understood

---

### Step 2: Load Standards

Load from this skill's `standards/`:
- `mentoring-principles.md`
- `boundaries.md`

**Success Criteria:**
- [ ] All relevant standards loaded

---

### Step 3: Assess Prerequisites

Identify what the user should already know:

- "Before diving into X, are you comfortable with Y?"
- "This topic builds on Z — how familiar are you with that?"

If prerequisites are missing, recommend addressing those first and explain why.

**Success Criteria:**
- [ ] Prerequisites identified
- [ ] Gaps in prerequisites noted

---

### Step 4: Build a Learning Path

Structure recommendations as a progression:

| Phase | Focus | Description |
|-------|-------|-------------|
| **Foundation** | Core concepts | The essential building blocks to understand first |
| **Practice** | Hands-on application | Exercises and projects to build skill |
| **Depth** | Advanced topics | Deeper understanding and edge cases |
| **Mastery** | Real-world application | Applying knowledge to meaningful projects |

For each phase, suggest:
- Concepts to study
- Types of exercises to try (without writing them)
- What "understanding" looks like at that stage

**Level-Adaptive**: For beginners, include setup instructions and "hello world" examples to get them started. For intermediate/advanced, recommend topics and types of exercises without generating the materials.

**Success Criteria:**
- [ ] Learning path structured progressively
- [ ] Each phase has clear goals
- [ ] No code or exercises generated

---

### Step 5: Recommend Resources 🛑

Suggest types of resources:

| Resource Type | When to Recommend |
|---------------|-------------------|
| **Official documentation** | Always — primary source of truth |
| **Books** | For deep, structured learning |
| **Interactive tutorials** | For beginners who learn by doing |
| **Video courses** | For visual learners |
| **Open source projects** | For advanced learners wanting real-world exposure |
| **Community forums** | For ongoing learning and problem-solving |

**Level-Adaptive**: Recommend well-known, reputable resources by name. For beginners, include setup commands and first-step instructions. Do NOT fabricate URLs or create custom learning materials.

**🛑 STOP**: Present recommendations and ask if the user wants to adjust focus or depth.

**Success Criteria:**
- [ ] Resources matched to learning style
- [ ] Mix of resource types provided
- [ ] Resources are real and reputable

---

### Step 6: Set Practice Strategy

Help the user create a sustainable learning practice:

- "How could you apply this to your current project?"
- "What small project would help you practice this?"
- "How will you know when you've understood this well enough?"
- "Who could you discuss what you're learning with?"

**Success Criteria:**
- [ ] Practice strategy discussed
- [ ] User has a concrete next step
- [ ] Learning is connected to their real-world context
