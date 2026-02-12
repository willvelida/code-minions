---
name: threat-modelling
description: 'Analyse repositories and deployed cloud infrastructure to produce STRIDE-based threat model documents with Mermaid architecture diagrams, prioritised threats, and recommendations. Use when a user asks to create a threat model, identify threats, assess attack surfaces, review an existing threat model, or update a threat model after system changes.'
license: MIT
allowed-tools: Bash
---

# Threat Modelling

## Overview

This skill produces threat model documents by analysing codebases, infrastructure-as-code, and live cloud resources. It follows the [STRIDE framework](https://learn.microsoft.com/en-us/training/modules/tm-use-a-framework-to-identify-threats-and-find-ways-to-reduce-or-eliminate-risk/) and the [Microsoft Threat Modeling Fundamentals](https://learn.microsoft.com/en-us/training/paths/tm-threat-modeling-fundamentals/) four-phase approach: Design, Break, Fix (recommendations only), and Verify.

**This skill produces threat models. It does NOT implement fixes or perform penetration testing.**

## Capabilities

| Capability | Action | Description |
|------------|--------|-------------|
| Analyse Repository | `actions/analyse-repository.md` | Scan codebase and IaC to map architecture and confirm a data flow diagram |
| Assess Infrastructure | `actions/assess-infrastructure.md` | Query live cloud resources via MCP servers to map architecture and confirm a data flow diagram |
| Generate Threat Model | `actions/generate-threat-model.md` | Apply STRIDE to a confirmed diagram, produce threat model diagram and document |
| Review Threat Model | `actions/review-threat-model.md` | Evaluate an existing threat model for completeness and accuracy |
| Update Threat Model | `actions/update-threat-model.md` | Revise an existing threat model based on system changes |

## Standards

| Standard | File | Description |
|----------|------|-------------|
| STRIDE Framework | `standards/stride-framework.md` | STRIDE threat categories and identification guidance |
| Document Template | `standards/document-template.md` | Threat model document structure and Mermaid diagram format |
| Checklist | `standards/checklist.md` | Consolidated compliance and quality checklist |

## Principles

### 1. Analyse, Never Fix

This skill identifies and documents threats. It recommends mitigations but **never implements fixes** itself. The output is a document, not a code change.

### 2. Confirm Before Proceeding

Always confirm the architecture diagram with the user before identifying threats. After threat identification, work with the user to prioritise, resolve, or remove items from the model.

### 3. Follow STRIDE Systematically

Apply all six STRIDE categories to every component and data flow. Do not skip categories — document "No threats identified" where appropriate rather than omitting the analysis.

### 4. Cloud-Agnostic

When assessing deployed infrastructure, detect available MCP servers or ask the user which cloud provider to query. Do not assume a specific cloud platform.

### 5. Prioritise by Risk

All identified threats must be prioritised. Use a consistent risk-rating approach combining likelihood and impact so the user can focus on what matters most.

## Scope Boundaries

| In Scope | Out of Scope |
|----------|-------------|
| Repository and IaC analysis | Implementing security fixes |
| Live cloud infrastructure assessment via MCP | Penetration testing |
| STRIDE-based threat identification | Security patching |
| Architecture diagrams (Mermaid) | Code vulnerability scanning |
| Threat prioritisation with user | Automated remediation |
| Mitigation recommendations | Compliance auditing |
| Reviewing existing threat models | |
| Updating existing threat models | |

## Usage

1. Load this skill manifest
2. Identify the required capability (analyse, assess, review, or update)
3. Load the bundled standards from `standards/`
4. Execute the action following `actions/<capability>.md`

## References

- [Microsoft Threat Modeling Fundamentals](https://learn.microsoft.com/en-us/training/paths/tm-threat-modeling-fundamentals/)
- [STRIDE Framework](https://learn.microsoft.com/en-us/training/modules/tm-use-a-framework-to-identify-threats-and-find-ways-to-reduce-or-eliminate-risk/)
- [Threat Modeling Phases](https://learn.microsoft.com/en-us/training/modules/tm-introduction-to-threat-modeling/1b-threat-modeling-phases/)
- [Data-Flow Diagram Elements](https://learn.microsoft.com/en-us/training/modules/tm-create-a-threat-model-using-foundational-data-flow-diagram-elements/)
- [Prioritise Issues and Apply Security Controls](https://learn.microsoft.com/en-us/training/modules/tm-prioritize-your-issues-and-apply-security-controls/)
