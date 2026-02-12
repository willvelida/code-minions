# Review Threat Model

## Purpose

Evaluate an existing threat model document for completeness, accuracy, and adherence to STRIDE methodology, then provide actionable findings.

---

## Flow

### Step 1: Locate Threat Model 🛑

Find the existing threat model document.

| Check | Action |
|-------|--------|
| File location | Look for `docs/threat-model.md` or similar in the repository |
| File format | Confirm it is a markdown document |
| Mermaid diagram | Check if an architecture diagram is present |

If no threat model is found:

**🛑 STOP**: Ask the user where the threat model is located. If none exists, recommend the Analyse Repository or Assess Infrastructure actions instead.

**Success Criteria:**
- [ ] Threat model document located and readable
- [ ] Document structure understood

---

### Step 2: Load Standards

Load from this skill's `standards/`:
- `stride-framework.md`
- `document-template.md`
- `checklist.md`

**Success Criteria:**
- [ ] All review standards loaded

---

### Step 3: Structure Compliance Check

Validate the document against `standards/document-template.md`:

| Section | Check | Severity |
|---------|-------|----------|
| Executive Summary | Present and meaningful | 🟠 High |
| Architecture Diagram | Present, uses Mermaid, renders correctly | 🔴 Critical |
| Components listed | All components labelled and described | 🟠 High |
| Data flows shown | Flows labelled with data description | 🟠 High |
| Trust boundaries marked | Boundaries clearly indicated | 🟠 High |
| STRIDE Threat Table | Present with all required columns | 🔴 Critical |
| All STRIDE categories covered | No categories skipped | 🔴 Critical |
| Prioritisation Matrix | Threats prioritised with likelihood and impact | 🟠 High |
| Recommendations | Mitigations listed (not implemented) | 🟠 High |

**Success Criteria:**
- [ ] All template sections checked
- [ ] Missing sections identified

---

### Step 4: Architecture Diagram Review

Evaluate the Mermaid diagram:

| Check | Severity |
|-------|----------|
| Diagram renders without errors | 🔴 Critical |
| All components from the system are represented | 🟠 High |
| Data flows have directional arrows | 🟠 High |
| Data flows are labelled | 🟡 Medium |
| Trust boundaries are clearly marked | 🟠 High |
| External entities are distinguished from internal components | 🟡 Medium |
| Diagram is readable (not overly complex) | 🟡 Medium |

If the repository or infrastructure has changed since the threat model was created, flag components or flows that may be missing or outdated.

**Success Criteria:**
- [ ] Diagram accuracy assessed
- [ ] Gaps identified if system has changed

---

### Step 5: STRIDE Coverage Review

For each component and data flow in the architecture diagram, verify that all six STRIDE categories have been considered:

| Category | Check |
|----------|-------|
| Spoofing | Threats related to identity and authentication assessed? |
| Tampering | Threats related to data integrity assessed? |
| Repudiation | Threats related to accountability and logging assessed? |
| Information Disclosure | Threats related to data confidentiality assessed? |
| Denial of Service | Threats related to availability assessed? |
| Elevation of Privilege | Threats related to authorisation assessed? |

Flag:
- Components with missing STRIDE categories
- Generic threats that lack specificity
- Threats missing likelihood or impact ratings

**Success Criteria:**
- [ ] STRIDE coverage verified for each component
- [ ] Gaps in coverage identified

---

### Step 6: Prioritisation Review

Evaluate the threat prioritisation:

| Check | Severity |
|-------|----------|
| Every threat has a likelihood rating | 🟠 High |
| Every threat has an impact rating | 🟠 High |
| Priority is consistent with likelihood × impact | 🟠 High |
| Critical threats have clear recommendations | 🟠 High |
| No threats left unprioritised | 🟡 Medium |

**Success Criteria:**
- [ ] Prioritisation logic is consistent
- [ ] All threats are rated

---

### Step 7: Generate Findings Report 🛑

Categorise all findings:

| Priority | Category | Action |
|----------|----------|--------|
| 🔴 Critical | Missing sections or broken diagram | Must fix |
| 🟠 High | Incomplete STRIDE coverage or missing priorities | Should fix |
| 🟡 Medium | Improvements to clarity or detail | Consider fixing |
| 🟢 Low | Minor formatting or style | Optional |

Present findings as:

### Summary Scores

| Category | Score | Notes |
|----------|-------|-------|
| Structure Compliance | X/10 | Template adherence |
| Diagram Accuracy | X/10 | Completeness and correctness |
| STRIDE Coverage | X/10 | Systematic threat identification |
| Prioritisation Quality | X/10 | Consistent and complete ratings |
| Overall | X/10 | Weighted average |

### Findings by Priority
List each finding with:
- What the issue is
- Where it is in the document
- How to fix it

**🛑 STOP**: Present findings to the user. Ask if they would like to proceed with updating the threat model using the Update Threat Model action.

**⚠️ REMINDER: Do NOT implement any fixes to the system. Only recommend changes to the threat model document.**

**Success Criteria:**
- [ ] Findings categorised by severity
- [ ] Scores calculated
- [ ] Actionable recommendations provided
