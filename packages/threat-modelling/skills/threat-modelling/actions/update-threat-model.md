# Update Threat Model

## Purpose

Revise an existing threat model to reflect changes in the system architecture, address review findings, or incorporate new information.

---

## Flow

### Step 1: Locate Existing Threat Model 🛑

Find and load the current threat model document.

If no threat model is found:

**🛑 STOP**: Recommend the Analyse Repository or Assess Infrastructure actions to create one first.

**Success Criteria:**
- [ ] Existing threat model located and loaded

---

### Step 2: Determine Update Reason 🛑

Ask the user why the threat model needs updating:

| Reason | Implication |
|--------|------------|
| **System changes** | New components, removed components, changed data flows |
| **Review findings** | Issues identified during a threat model review |
| **New information** | New threat intelligence, changed risk context |
| **Periodic refresh** | Scheduled review cycle |

**🛑 STOP**: Wait for the user to explain what has changed or what needs updating.

**Success Criteria:**
- [ ] Update reason documented
- [ ] Scope of changes understood

---

### Step 3: Identify Changes

Based on the update reason:

#### For System Changes

Compare the current system (repository or deployed infrastructure) against the existing threat model:

| Check | Action |
|-------|--------|
| New components | Scan for components not in the current diagram |
| Removed components | Identify diagram elements that no longer exist |
| Changed data flows | Detect new or modified connections |
| Changed trust boundaries | Check for network or identity changes |
| New IaC resources | Look for infrastructure changes since last model |

#### For Review Findings

Load the review findings and map each to a section of the threat model:

| Finding Type | Update Action |
|-------------|--------------|
| Missing STRIDE categories | Add threat analysis for missing categories |
| Incomplete component coverage | Add missing components to diagram and analysis |
| Inaccurate diagram | Correct the Mermaid diagram |
| Missing priorities | Add likelihood, impact, and priority ratings |
| Weak recommendations | Strengthen mitigation recommendations |

#### For New Information

Document what has changed in the threat landscape:

| Input | Update Action |
|-------|--------------|
| New threat intelligence | Reassess likelihood ratings |
| Changed compliance requirements | Add relevant threats and controls |
| Incident learnings | Add threats based on real incidents |

**Success Criteria:**
- [ ] All changes identified
- [ ] Each change mapped to a threat model section

---

### Step 4: Update Architecture Diagram 🛑

Modify the Mermaid diagram to reflect changes:

1. Add new components, data flows, or trust boundaries
2. Remove elements that no longer exist
3. Update labels where names or descriptions have changed
4. Ensure the diagram remains readable

**🛑 STOP**: Present the updated diagram to the user. Ask:
1. Does this accurately reflect the current system?
2. Are there any other changes needed?

Do NOT proceed until the user confirms the updated diagram.

**Success Criteria:**
- [ ] Diagram updated with all changes
- [ ] User confirmed updated diagram

---

### Step 5: Update STRIDE Analysis

For any new or changed components and data flows:

1. Apply all six STRIDE categories per `standards/stride-framework.md`
2. Document new threats with full details (ID, category, component, description, impact, likelihood, mitigation)
3. Review existing threats for components that changed — update or remove as appropriate
4. Mark removed threats as "Resolved — component removed" rather than deleting them

For removed components:
- Mark associated threats as no longer applicable
- Keep in an appendix or mark with strikethrough for audit trail

**Success Criteria:**
- [ ] New components analysed with STRIDE
- [ ] Changed components re-analysed
- [ ] Removed components' threats marked appropriately

---

### Step 6: Re-Prioritise Threats with User 🛑

Present the updated threat list, highlighting what changed:

| Change Type | Indicator |
|-------------|-----------|
| New threat | 🆕 |
| Updated threat | 🔄 |
| Removed threat | ❌ |
| Unchanged | — |

Use the priority matrix to rate new threats:

| | High Impact | Medium Impact | Low Impact |
|---|------------|---------------|------------|
| **High Likelihood** | 🔴 Critical | 🟠 High | 🟡 Medium |
| **Medium Likelihood** | 🟠 High | 🟡 Medium | 🟢 Low |
| **Low Likelihood** | 🟡 Medium | 🟢 Low | 🟢 Low |

**🛑 STOP**: Work with the user to:
1. Confirm ratings for new and updated threats
2. Agree on which removed threats to archive vs delete
3. Finalise the updated priority list

**Success Criteria:**
- [ ] New threats prioritised
- [ ] Updated threats re-assessed
- [ ] User confirmed final threat list

---

### Step 7: Generate Updated Document 🛑

Update the threat model document following `standards/document-template.md`:

1. Update the Executive Summary with what changed and why
2. Replace the architecture diagram with the updated version
3. Update the STRIDE threat table
4. Update the prioritisation matrix
5. Update recommendations for new or changed threats
6. Add a change log entry at the bottom of the document

**Change Log Entry Format:**
```markdown
## Change Log

| Date | Reason | Summary of Changes |
|------|--------|--------------------|
| YYYY-MM-DD | <reason> | <brief description of what changed> |
```

**🛑 STOP**: Present the updated document to the user.

**⚠️ REMINDER: Do NOT implement any fixes. The output is an updated document only.**

**Success Criteria:**
- [ ] Document updated with all changes
- [ ] Change log entry added
- [ ] User confirmed updated document is complete
