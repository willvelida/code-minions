# Create DevContainer

## Purpose

Generate a new, standards-compliant DevContainer configuration for a project.

---

## Flow

### Step 1: Determine Approach 🛑

Present two options to the user:

| Option | Description | Best For |
|--------|-------------|----------|
| **A: Automatic** | Analyse repository to detect technologies | Existing repos, quick setup |
| **B: Manual** | User specifies requirements directly | Blank repos, specific needs |

**🛑 STOP**: Wait for the user to select Option A or Option B before proceeding.

**Success Criteria:**
- [ ] User has explicitly selected their preferred approach

---

### Step 2: Gather Requirements

#### Path A: Automatic Analysis

Analyse the repository to detect technologies, then present findings to the user for confirmation.

#### Path B: Manual Specification 🛑

Collect from user:
1. Primary development language(s) and versions
2. Frameworks in use
3. Additional tools needed

**🛑 STOP**: Wait for the user to provide their requirements.

**Success Criteria:**
- [ ] All project languages identified with specific versions
- [ ] Frameworks and tools catalogued
- [ ] User has confirmed requirements

---

### Step 3: Load Standards

1. **Always load** from this skill's `standards/`:
   - `core.md`
   - `security.md`
   - `extensions.md`
   - `features.md`

2. **Per language**: Load from `../../standards/languages/<language>/`:
   - `development-environment.md`
   - `tooling.md` (for tool configuration guidance)

**Success Criteria:**
- [ ] Core skill standards loaded
- [ ] All relevant language standards loaded

---

### Step 4: Build Configuration

Apply configuration per loaded standards:

1. **Base image & features** — Per `core.md`
2. **Feature versioning** — Per `features.md`
3. **Essential extensions** — Per `extensions.md`
4. **Language extensions** — Per language `development-environment.md` files
5. **Container settings** — Per `core.md`
6. **Security settings** — Per `security.md`

**Success Criteria:**
- [ ] All requirements from loaded standards applied

---

### Step 5: Create Lifecycle Hooks

Create hooks per `core.md`:

1. **Post-Create Hook** — `.devcontainer/post-create.sh`
2. **Post-Start Hook** — `.devcontainer/post-start.sh`

Include language-specific setup from language `tooling.md` files.

**Success Criteria:**
- [ ] Both hooks created per standards
- [ ] Language-specific setup included

---

### Step 6: Validate Configuration

Validate against `checklist.md`:

1. Configuration compliance
2. Security compliance
3. All checklist items pass

**Success Criteria:**
- [ ] All checklist items pass

---

### Step 7: Generate Output Files

Create files:

```
.devcontainer/
├── devcontainer.json
├── post-create.sh
└── post-start.sh
```

**Success Criteria:**
- [ ] Files created in correct location

---

### Step 8: Document

Summarise:

1. What was configured and why
2. Included features (with versions) and extensions
3. How to rebuild/customise

**Success Criteria:**
- [ ] User understands the configuration
