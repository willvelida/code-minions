# Threat Model Checklist

## Purpose

Consolidated checklist for validating threat model documents against quality requirements and STRIDE methodology.

---

## Scope and Boundaries

- [ ] Skill produced a document only — no fixes implemented
- [ ] No penetration testing was performed
- [ ] Scope was confirmed with the user before analysis
- [ ] Out-of-scope items documented

---

## Architecture Diagram

- [ ] Mermaid diagram present and renders correctly
- [ ] All system components represented
- [ ] Data stores identified
- [ ] External entities shown
- [ ] Data flows labelled with directional arrows
- [ ] Trust boundaries clearly marked using subgraphs
- [ ] Diagram confirmed by user before threat identification

---

## STRIDE Coverage

- [ ] Spoofing threats assessed for all external entities and processes
- [ ] Tampering threats assessed for all data stores, data flows, and processes
- [ ] Repudiation threats assessed for all external entities and processes
- [ ] Information Disclosure threats assessed for all data stores, data flows, and processes
- [ ] Denial of Service threats assessed for all processes, data stores, and data flows
- [ ] Elevation of Privilege threats assessed for all processes
- [ ] No STRIDE categories skipped — "No threats identified" documented where appropriate

---

## Threat Documentation

- [ ] Every threat has a unique ID (T-001, T-002, etc.)
- [ ] Every threat has a STRIDE category
- [ ] Every threat identifies the affected component or data flow
- [ ] Every threat has a specific, clear description
- [ ] Every threat has an impact rating (High / Medium / Low)
- [ ] Every threat has a likelihood rating (High / Medium / Low)
- [ ] Every threat has a priority derived from the priority matrix
- [ ] Every threat has a recommended mitigation

---

## Prioritisation

- [ ] All threats prioritised using the likelihood × impact matrix
- [ ] Priorities reviewed and confirmed with user
- [ ] Threats user chose to remove are documented as excluded
- [ ] Critical and high-priority threats have specific recommendations

---

## Document Structure

- [ ] Executive Summary present with key findings counts
- [ ] Architecture Diagram section present
- [ ] Threat Model Diagram present and renders correctly
- [ ] STRIDE Threat Table present with all required columns
- [ ] Prioritisation Matrix present (by priority and by component)
- [ ] Recommendations section present
- [ ] Change Log present (for updated models)

---

## Recommendations Quality

- [ ] Recommendations are specific and actionable
- [ ] Recommendations reference their threat ID
- [ ] Recommendations explain rationale
- [ ] Recommendations do NOT include implementation code
- [ ] Recommendations are grouped by priority

---

## Infrastructure Assessment (If Applicable)

- [ ] Cloud provider identified
- [ ] MCP server used to query live resources
- [ ] All resource types queried (compute, networking, data, identity, APIs, secrets)
- [ ] Public-facing resources flagged
- [ ] Network and identity boundaries mapped as trust boundaries

---

## Review and Update (If Applicable)

- [ ] All template sections checked for completeness
- [ ] STRIDE coverage verified for every component
- [ ] Changes clearly marked (new, updated, removed)
- [ ] Change log entry added with date and reason
- [ ] User confirmed final document
