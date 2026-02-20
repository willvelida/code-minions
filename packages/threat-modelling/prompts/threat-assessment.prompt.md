---
description: "Perform a STRIDE threat assessment on the codebase"
mode: agent
input:
  - name: scope
    description: "Focus area for the assessment (e.g. API layer, auth flow, data pipeline)"
  - name: threat_level
    description: "Minimum severity to report (low, medium, high, critical)"
---

# Threat Assessment

Analyse the repository for security threats using the STRIDE framework.

## Scope

Focus area: ${input:scope}

## Threat Level

Minimum severity to report: ${input:threat_level}

## Instructions

1. **Identify trust boundaries** in the scoped area:
   - External inputs (user data, API calls, file uploads)
   - Service boundaries (microservices, third-party integrations)
   - Data stores (databases, caches, file systems)

2. **Map data flows** across trust boundaries

3. **Apply STRIDE categories** to each data flow:
   - **S**poofing — can an attacker impersonate a legitimate user/service?
   - **T**ampering — can data be modified in transit or at rest?
   - **R**epudiation — can actions be denied without evidence?
   - **I**nformation disclosure — can sensitive data leak?
   - **D**enial of service — can the system be overwhelmed?
   - **E**levation of privilege — can an attacker gain unauthorised access?

4. **Prioritise threats** by severity (${input:threat_level} and above)

5. **Recommend mitigations** for each identified threat

6. **Output a threat model document** with:
   - Architecture diagram (Mermaid format)
   - Threat table with ID, category, description, severity, mitigation
   - Summary of findings and recommended next steps
