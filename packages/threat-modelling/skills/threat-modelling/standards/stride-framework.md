# STRIDE Framework

## Purpose

Reference guide for applying the STRIDE threat classification framework systematically to identify threats in a system architecture.

---

## STRIDE Categories

| Category | Threat | Property Violated | Question |
|----------|--------|-------------------|----------|
| **S**poofing | Pretending to be something or someone else | Authentication | Can an attacker impersonate a user, service, or component? |
| **T**ampering | Modifying data or code without authorisation | Integrity | Can data be altered in transit, at rest, or during processing? |
| **R**epudiation | Claiming to not have performed an action | Non-repudiation | Can a user or system deny performing an action without accountability? |
| **I**nformation Disclosure | Exposing data to unauthorised parties | Confidentiality | Can sensitive data be accessed by unauthorised users or systems? |
| **D**enial of Service | Making a system unavailable | Availability | Can a component be overwhelmed, crashed, or made inaccessible? |
| **E**levation of Privilege | Gaining access beyond what is authorised | Authorisation | Can an attacker gain higher privileges than intended? |

---

## Applying STRIDE to Elements

### External Entities (Users, Third-Party Systems)

| Category | Applicable | Common Threats |
|----------|-----------|----------------|
| Spoofing | ✅ | Impersonation, credential theft, session hijacking |
| Tampering | ❌ | (applies to data flows instead) |
| Repudiation | ✅ | Denying transactions, false claims |
| Information Disclosure | ❌ | (applies to data flows and stores) |
| Denial of Service | ❌ | (applies to processes and stores) |
| Elevation of Privilege | ❌ | (applies to processes) |

### Processes (Services, Applications, Functions)

| Category | Applicable | Common Threats |
|----------|-----------|----------------|
| Spoofing | ✅ | Service impersonation, rogue processes |
| Tampering | ✅ | Code injection, configuration tampering |
| Repudiation | ✅ | Missing audit logs, unsigned actions |
| Information Disclosure | ✅ | Error messages leaking data, debug endpoints |
| Denial of Service | ✅ | Resource exhaustion, crash exploits |
| Elevation of Privilege | ✅ | Privilege escalation, insecure defaults |

### Data Stores (Databases, Caches, Files)

| Category | Applicable | Common Threats |
|----------|-----------|----------------|
| Spoofing | ❌ | (applies to entities and processes) |
| Tampering | ✅ | Unauthorised writes, SQL injection |
| Repudiation | ❌ | (applies to entities and processes) |
| Information Disclosure | ✅ | Unencrypted storage, excessive access |
| Denial of Service | ✅ | Storage exhaustion, lock contention |
| Elevation of Privilege | ❌ | (applies to processes) |

### Data Flows (Network Connections, API Calls)

| Category | Applicable | Common Threats |
|----------|-----------|----------------|
| Spoofing | ❌ | (applies to entities and processes) |
| Tampering | ✅ | Man-in-the-middle, unencrypted channels |
| Repudiation | ❌ | (applies to entities and processes) |
| Information Disclosure | ✅ | Eavesdropping, unencrypted data in transit |
| Denial of Service | ✅ | Network flooding, connection exhaustion |
| Elevation of Privilege | ❌ | (applies to processes) |

---

## Threat Documentation Format

For each identified threat, document:

| Field | Description | Example |
|-------|-------------|---------|
| ID | Unique identifier | `T-001` |
| Category | STRIDE category | Spoofing |
| Component | Affected component or data flow | API Gateway → Auth Service |
| Description | What the threat is | An attacker could forge JWT tokens to impersonate authenticated users |
| Impact | What happens if exploited | Unauthorised access to user data and actions |
| Likelihood | How likely (High / Medium / Low) | Medium |
| Priority | Risk rating from priority matrix | 🟠 High |
| Recommended Mitigation | How to address (recommendation only) | Implement token signature validation with key rotation |

---

## Priority Matrix

Combine likelihood and impact to determine priority:

| | High Impact | Medium Impact | Low Impact |
|---|------------|---------------|------------|
| **High Likelihood** | 🔴 Critical | 🟠 High | 🟡 Medium |
| **Medium Likelihood** | 🟠 High | 🟡 Medium | 🟢 Low |
| **Low Likelihood** | 🟡 Medium | 🟢 Low | 🟢 Low |

### Likelihood Guidance

| Rating | Description |
|--------|-------------|
| High | Easily exploitable, well-known attack vector, minimal skill required |
| Medium | Exploitable with moderate effort or specific conditions |
| Low | Requires significant effort, insider access, or unlikely conditions |

### Impact Guidance

| Rating | Description |
|--------|-------------|
| High | Data breach, full system compromise, significant financial or reputational damage |
| Medium | Partial data exposure, service degradation, limited blast radius |
| Low | Minor information leak, temporary inconvenience, easily recoverable |

---

## Common Mitigations by Category

These are **recommendations only** — this skill does not implement them.

| Category | Common Mitigations |
|----------|-------------------|
| Spoofing | Strong authentication, MFA, certificate pinning, token validation |
| Tampering | Input validation, integrity checks, signed payloads, immutable infrastructure |
| Repudiation | Audit logging, digital signatures, timestamps, centralised log management |
| Information Disclosure | Encryption (at rest and in transit), access controls, data classification |
| Denial of Service | Rate limiting, auto-scaling, circuit breakers, redundancy |
| Elevation of Privilege | Least privilege, RBAC, input validation, security boundaries |

## References

- [STRIDE Framework — Microsoft Learn](https://learn.microsoft.com/en-us/training/modules/tm-use-a-framework-to-identify-threats-and-find-ways-to-reduce-or-eliminate-risk/)
- [Threat Modeling Phases](https://learn.microsoft.com/en-us/training/modules/tm-introduction-to-threat-modeling/1b-threat-modeling-phases/)
