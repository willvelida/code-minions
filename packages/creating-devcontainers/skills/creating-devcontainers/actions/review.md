# Review DevContainer

## Purpose

Analyse an existing DevContainer configuration for standards compliance, security, and performance, then provide actionable recommendations.

---

## Flow

### Step 1: Locate Configuration

Find DevContainer files:

| File | Location | Required |
|------|----------|----------|
| devcontainer.json | `.devcontainer/devcontainer.json` | Yes |
| post-create.sh | `.devcontainer/post-create.sh` | Check existence |
| post-start.sh | `.devcontainer/post-start.sh` | Check existence |

**Failure:** No configuration found → Recommend Create action

---

### Step 2: Analyse Project Context

Detect project technologies to determine applicable language standards.

---

### Step 3: Load Standards

1. **Always load** from this skill's `standards/`:
   - `core.md`
   - `security.md`
   - `extensions.md`
   - `features.md`
   - `checklist.md`

2. **Per language**: Use the detected project language(s) to evaluate language-specific configuration.

---

### Step 4: Configuration Compliance Check

Evaluate against `checklist.md` configuration section:

- Base image
- Features and versioning
- Extensions
- Lifecycle hooks
- Container settings

---

### Step 5: Security Analysis

Evaluate against `security.md`:

| Priority | Checks |
|----------|--------|
| 🔴 Critical | Non-root user, privileged mode, hardcoded secrets |
| 🟠 High | Git safe.directory, host mounts, base image |
| 🟡 Medium | Feature versions, port forwarding |

---

### Step 6: Performance Analysis

| Factor | Question |
|--------|----------|
| Redundant features | `common-utils` or `git` features added unnecessarily? |
| Unnecessary features | Any features that could be removed? |
| Unused extensions | Any extensions not needed? |
| Hook optimisation | Could post-create commands be faster? |

---

### Step 7: Generate Findings Report

Categorise findings:

| Priority | Category |
|----------|----------|
| 🔴 Critical | Security issues - must fix immediately |
| 🟠 High | Standards compliance violations |
| 🟡 Medium | Performance/versioning improvements |
| 🟢 Low | Nice-to-have enhancements |

---

### Step 8: Present Results

Provide:

1. **Scores** — Compliance, Security, Performance (X/10)
2. **Critical Issues** — Security problems requiring immediate action
3. **Compliance Gaps** — Standards violations with specific fixes
4. **Quick Wins** — Easy changes with high impact

---

## Scoring Methodology

| Category | Weight | Calculation |
|----------|--------|-------------|
| Base Image | 20% | Correct = 100, Wrong = 0 |
| Essential Extensions | 20% | (present / required) × 100 |
| Language Requirements | 20% | (present / required) × 100 |
| Security Checks | 20% | (passed / total) × 100 |
| Versioning | 20% | (explicit / total features) × 100 |
