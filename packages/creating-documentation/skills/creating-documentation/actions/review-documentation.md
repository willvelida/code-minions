# Review Documentation

## Purpose

Analyse existing repository documentation for quality, completeness, accessibility, and best practices adherence, then provide actionable recommendations.

---

## Flow

### Step 1: Locate Documentation

Find documentation files in the repository:

| File | Location | Priority |
|------|----------|----------|
| README.md | Root, `.github/`, or `docs/` | Required |
| CONTRIBUTING.md | Root or `.github/` | Recommended |
| CODE_OF_CONDUCT.md | Root or `.github/` | Recommended |
| LICENSE | Root | Recommended |
| CHANGELOG.md | Root | Optional |
| docs/ | `docs/` directory | Optional |

**Failure:** No README.md found → Recommend Create README action

**Success Criteria:**
- [ ] Documentation files located and catalogued
- [ ] Missing recommended files noted

---

### Step 2: Load Standards

Load from this skill's `standards/`:
- `readme-structure.md`
- `accessibility.md`
- `writing-style.md`
- `markdown.md`
- `checklist.md`

**Success Criteria:**
- [ ] All review standards loaded

---

### Step 3: README Structure Analysis

Evaluate README against `readme-structure.md`:

| Section | Status | Severity |
|---------|--------|----------|
| Title/Name | Present? | 🔴 Critical |
| Description | Present, clear? | 🔴 Critical |
| Installation | Present, complete? | 🔴 Critical |
| Usage | Present, with examples? | 🔴 Critical |
| License | Present or referenced? | 🟠 High |
| Contributing | Present or linked? | 🟡 Medium |
| Support/Contact | Present? | 🟢 Low |

**Common Issues:**
- Missing installation instructions
- Usage examples that don't work
- Outdated information
- Missing license information

**Success Criteria:**
- [ ] All sections evaluated
- [ ] Missing sections documented with severity

---

### Step 4: Accessibility Audit

Evaluate against `accessibility.md`:

#### Heading Structure

| Check | Severity |
|-------|----------|
| Single H1 (title only) | 🔴 Critical |
| No skipped heading levels | 🔴 Critical |
| Logical heading hierarchy | 🟠 High |

#### Images

| Check | Severity |
|-------|----------|
| All images have alt text | 🔴 Critical |
| Alt text is descriptive | 🟠 High |
| Decorative images handled appropriately | 🟡 Medium |

#### Links

| Check | Severity |
|-------|----------|
| No "click here" or "here" links | 🔴 Critical |
| Link text describes destination | 🟠 High |
| No duplicate link text for different URLs | 🟡 Medium |
| External links work (not broken) | 🟠 High |

#### Lists

| Check | Severity |
|-------|----------|
| Proper markdown list syntax used | 🟠 High |
| No emoji/symbols as bullet points | 🟠 High |

#### Emoji Usage

| Check | Severity |
|-------|----------|
| Emoji not used for critical information | 🟠 High |
| Emoji used sparingly | 🟡 Medium |
| No long emoji sequences | 🟡 Medium |

**Success Criteria:**
- [ ] All accessibility checks completed
- [ ] Issues documented with severity

---

### Step 5: Writing Quality Analysis

Evaluate against `writing-style.md`:

| Quality Factor | Check |
|----------------|-------|
| Clarity | Is the purpose immediately clear? |
| Conciseness | Is information presented efficiently? |
| Readability | Short sentences, plain language? |
| Consistency | Consistent terminology throughout? |
| Completeness | All necessary information included? |
| Accuracy | Information is current and correct? |

**Readability Checks:**
- Sentences under 25 words (ideal)
- Paragraphs under 5 sentences
- Technical jargon explained or linked
- Active voice preferred

**Success Criteria:**
- [ ] Writing quality assessed
- [ ] Improvement suggestions documented

---

### Step 6: Markdown Quality Check

Evaluate against `markdown.md`:

| Check | Severity |
|-------|----------|
| Code blocks have language tags | 🟠 High |
| Consistent formatting style | 🟡 Medium |
| Tables properly formatted | 🟡 Medium |
| No HTML when Markdown suffices | 🟢 Low |
| Relative links used for repo files | 🟡 Medium |

**Success Criteria:**
- [ ] Markdown quality assessed
- [ ] Formatting issues documented

---

### Step 7: Generate Review Report 🛑

Compile findings into a structured report:

```markdown
# Documentation Review Report

## Summary
- Overall Score: [Good/Needs Work/Poor]
- Critical Issues: [count]
- Recommendations: [count]

## Critical Issues (Must Fix)
[List issues with 🔴 severity]

## High Priority (Should Fix)
[List issues with 🟠 severity]

## Medium Priority (Consider Fixing)
[List issues with 🟡 severity]

## Low Priority (Nice to Have)
[List issues with 🟢 severity]

## Accessibility Score
[Summary of accessibility compliance]

## Recommendations
[Prioritized list of improvements]
```

**🛑 STOP**: Present report to user and await feedback.

**Success Criteria:**
- [ ] Report generated with all findings
- [ ] Issues prioritized by severity
- [ ] Actionable recommendations provided

---

### Step 8: Offer Fixes

After presenting the report, offer to:

1. **Fix critical issues** — Apply fixes for must-fix items
2. **Improve accessibility** — Address accessibility violations
3. **Enhance content** — Improve writing quality and completeness
4. **Regenerate** — Create new documentation using Create actions

**Success Criteria:**
- [ ] User has chosen which improvements to apply
- [ ] Selected improvements implemented
