# Review Agent Skill

## Purpose

Analyse an existing Agent Skill for specification compliance, quality, and best practices adherence, then provide actionable recommendations.

---

## Flow

### Step 1: Locate Skill

Find the skill to review:

| File | Location | Required |
|------|----------|----------|
| SKILL.md | `<skill-dir>/SKILL.md` | Yes |
| actions/ | `<skill-dir>/actions/` | Check existence |
| standards/ | `<skill-dir>/standards/` | Check existence |
| scripts/ | `<skill-dir>/scripts/` | Check existence |
| references/ | `<skill-dir>/references/` | Check existence |

**Failure:** No SKILL.md found → Recommend Create action

**Success Criteria:**
- [ ] SKILL.md located and readable
- [ ] Directory structure catalogued

---

### Step 2: Load Standards

Load from this skill's `standards/`:
- `specification.md`
- `naming.md`
- `descriptions.md`
- `structure.md`
- `instructions.md`
- `checklist.md`

**Success Criteria:**
- [ ] All review standards loaded

---

### Step 3: Frontmatter Compliance Check

Validate YAML frontmatter against `specification.md`:

| Field | Check | Severity |
|-------|-------|----------|
| `name` | Present, 1-64 chars, lowercase alphanumeric + hyphens | 🔴 Critical |
| `name` | Matches parent directory name | 🔴 Critical |
| `name` | No reserved words (`anthropic`, `claude`) | 🔴 Critical |
| `name` | No consecutive hyphens, doesn't start/end with hyphen | 🔴 Critical |
| `description` | Present, 1-1024 chars | 🔴 Critical |
| `description` | Third-person voice | 🟠 High |
| `description` | Includes what AND when to use | 🟠 High |
| `description` | Contains specific trigger keywords | 🟡 Medium |
| `license` | Present (recommended) | 🟢 Low |
| `allowed-tools` | Valid if present | 🟡 Medium |

**Success Criteria:**
- [ ] All critical checks pass
- [ ] Findings documented with severity

---

### Step 4: Description Quality Analysis

Evaluate against `descriptions.md`:

| Quality Factor | Check |
|----------------|-------|
| Specificity | Does it clearly state what the skill does? |
| Triggers | Does it explain when to use the skill? |
| Keywords | Does it include terms agents will search for? |
| Voice | Is it written in third person? |
| Length | Is it concise but informative? |

**Common Issues:**
- Too vague: "Helps with files" → No specific capability
- Missing triggers: "Processes PDFs" → When should it activate?
- Wrong voice: "I can help you..." → Should be third person
- Too long: Over-explaining basic concepts

**Success Criteria:**
- [ ] Description quality assessed
- [ ] Improvement suggestions documented

---

### Step 5: Structure Analysis

Evaluate against `structure.md`:

| Check | Severity |
|-------|----------|
| SKILL.md body under 500 lines | 🟠 High |
| File references one level deep | 🟠 High |
| Referenced files exist | 🔴 Critical |
| No deeply nested reference chains | 🟠 High |
| Directory structure appropriate for complexity | 🟡 Medium |
| Files named descriptively | 🟡 Medium |

**Progressive Disclosure Check:**
- Is detailed content moved to separate files?
- Are reference files loaded only when needed?
- Is the token budget respected?

**Success Criteria:**
- [ ] Structure analysis complete
- [ ] Token budget compliance assessed

---

### Step 6: Instructions Quality Analysis

Evaluate SKILL.md body and action files against `instructions.md`:

| Quality Factor | Check |
|----------------|-------|
| Conciseness | Does it avoid explaining what agents already know? |
| Freedom Level | Is specificity appropriate for task fragility? |
| Examples | Are concrete examples provided where helpful? |
| Terminology | Is terminology consistent throughout? |
| Time-sensitivity | Does it avoid dates that will become outdated? |

**Action Files (if present):**

| Check | Severity |
|-------|----------|
| Clear purpose statement | 🟠 High |
| Numbered sequential steps | 🟠 High |
| Stop points for user input | 🟡 Medium |
| Success criteria per step | 🟡 Medium |
| Appropriate detail level | 🟡 Medium |

**Success Criteria:**
- [ ] Instructions quality assessed
- [ ] Action files reviewed (if present)

---

### Step 7: Standards Files Analysis (If Present)

Evaluate `standards/` files:

| Check | Severity |
|-------|----------|
| Each file has clear purpose | 🟠 High |
| Rules are specific and measurable | 🟠 High |
| Good and bad examples provided | 🟡 Medium |
| Consistent formatting | 🟡 Medium |
| Checklist consolidates all criteria | 🟡 Medium |

**Success Criteria:**
- [ ] Standards files reviewed
- [ ] Quality and completeness assessed

---

### Step 8: Scripts Analysis (If Present)

Evaluate `scripts/` files against best practices:

| Check | Severity |
|-------|----------|
| Scripts handle errors explicitly | 🟠 High |
| No "magic numbers" without explanation | 🟡 Medium |
| Dependencies documented | 🟠 High |
| Clear documentation | 🟡 Medium |
| Unix-style paths (forward slashes) | 🟠 High |

**Success Criteria:**
- [ ] Scripts reviewed for quality
- [ ] Dependencies identified

---

### Step 9: Anti-Pattern Detection

Check for common anti-patterns:

| Anti-Pattern | Detection | Severity |
|--------------|-----------|----------|
| Windows paths | Backslashes in file paths | 🟠 High |
| Too many options | Multiple approaches without clear default | 🟠 High |
| Over-explaining | Explaining concepts agents know | 🟡 Medium |
| Vague names | Generic names like `helper`, `utils` | 🟡 Medium |
| Time-sensitive info | Dates or version-specific instructions | 🟡 Medium |
| Inconsistent terminology | Mixed terms for same concept | 🟡 Medium |
| Deep nesting | Reference chains more than 1 level | 🟠 High |

**Success Criteria:**
- [ ] Anti-patterns identified
- [ ] Fixes documented

---

### Step 10: Generate Findings Report

Categorise all findings:

| Priority | Category | Action |
|----------|----------|--------|
| 🔴 Critical | Specification violations | Must fix — skill may not work |
| 🟠 High | Quality issues | Should fix — impacts effectiveness |
| 🟡 Medium | Improvements | Consider fixing — enhances quality |
| 🟢 Low | Nice-to-have | Optional — minor enhancements |

**Success Criteria:**
- [ ] All findings categorised
- [ ] Specific fixes documented

---

### Step 11: Present Results

Provide:

### Summary Scores

| Category | Score | Notes |
|----------|-------|-------|
| Specification Compliance | X/10 | Frontmatter, naming, structure |
| Description Quality | X/10 | Clarity, triggers, keywords |
| Instructions Quality | X/10 | Conciseness, examples, consistency |
| Overall | X/10 | Weighted average |

### Critical Issues (🔴)
<List any specification violations requiring immediate fix>

### High Priority (🟠)
<List quality issues that should be addressed>

### Improvements (🟡)
<List medium priority enhancements>

### Quick Wins
<List easy changes with high impact>

### Recommended Actions
1. <Specific action with file/line reference>
2. <Specific action with file/line reference>
...

**Success Criteria:**
- [ ] Scores calculated
- [ ] Actionable recommendations provided
- [ ] Findings prioritised by severity
