# Implementation Plan: Bundle Agents and Skills as Packages

> **GitHub Issue:** [#14 — Bundle agents and skills together](https://github.com/willvelida/code-minions/issues/14)
>
> **Status:** Planning
>
> **Breaking Change:** Yes — removes `--skills` and `--agents` flags from `install` command

---

## Design Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Package structure | `packages/<name>/` containing `agents/` and `skills/` subdirs | Self-contained, no registry file needed |
| Install command | `code-minions install --package <name>` (flag on existing command) | Simpler than a subcommand, consistent with `--standards` |
| Standards | Separate — installed via `--standards <lang>` | Skills have their own standards; avoids duplication when packages share a language standard |
| Output layout | Preserves `agents/` and `skills/` structure in user's repo | Consistent organisation across packages |
| Old flags | Remove `--skills` and `--agents` | Packages replace them (breaking change, acceptable for early project) |

---

## Package Directory Structure

Each package is a folder under `packages/` that mirrors the target repo layout:

```
packages/
├── threat-modelling/
│   ├── agents/
│   │   └── threat-modelling.agent.md
│   └── skills/
│       └── threat-modelling/
│           ├── SKILL.md
│           ├── actions/
│           │   └── ...
│           └── standards/
│               └── ...
├── developer-mentor/
│   ├── agents/
│   │   └── developer-mentor.agent.md
│   └── skills/
│       └── developer-mentor/
│           ├── SKILL.md
│           ├── actions/
│           │   └── ...
│           └── standards/
│               └── ...
├── git-workflow/
│   ├── agents/
│   │   └── git-workflow.agent.md
│   └── skills/
│       └── git-workflow/
│           ├── SKILL.md
│           ├── actions/
│           │   └── ...
│           └── standards/
│               └── ...
└── ...
```

When installed, the contents land in the user's repo preserving `agents/` and `skills/`:

```
user-repo/
├── agents/
│   └── threat-modelling.agent.md
└── skills/
    └── threat-modelling/
        ├── SKILL.md
        ├── actions/
        └── standards/
```

---

## Implementation Steps

### Step 1: Create the `packages/` directory structure

**What:** Physically reorganise the repo by creating `packages/<name>/` folders and moving agent + skill files into them.

**Files to create/move:**

- For each existing agent + skill pair (e.g., `threat-modelling`, `developer-mentor`, `git-workflow`):
  - Create `packages/<name>/agents/` and copy the agent `.agent.md` file into it
  - Create `packages/<name>/skills/<name>/` and copy the skill folder contents into it

**Packages to create (based on current agents with matching skills):**

| Package name | Agent file | Skill folder |
|---|---|---|
| `threat-modelling` | `threat-modelling.agent.md` | `skills/threat-modelling/` |
| `developer-mentor` | `developer-mentor.agent.md` | `skills/developer-mentor/` |
| `git-workflow` | `git-workflow.agent.md` | `skills/git-workflow/` |

**Naming mismatches to resolve (agent name ≠ skill name):**

| Agent file | Skill folder | Decision needed |
|---|---|---|
| `devcontainer.agent.md` | `skills/creating-devcontainers/` | Rename one to match? |
| `agent-skill.agent.md` | `skills/creating-agent-skill/` | Rename one to match? |

**After migration:** Remove the old top-level `agents/` and `skills/` directories (the `standards/` directory stays).

---

### Step 2: Update `embed.go`

**What:** Change the embed directive to include `packages` instead of `agents` and `skills`.

**Current:**

```go
//go:embed agents skills standards
var Content embed.FS
```

**New:**

```go
//go:embed packages standards
var Content embed.FS
```

**Go concept:** The `//go:embed` directive tells the Go compiler to bake files into the binary at build time. By changing it, we include the new `packages/` tree and stop including the old separate dirs.

---

### Step 3: Update `internal/cmd/install.go`

**What:** Replace `--skills` and `--agents` flags with `--package` flag. Update `buildDirList` to resolve packages.

**Changes:**

1. **Remove flags:** `--skills`, `--agents`
2. **Add flag:** `--package` (comma-separated string, e.g., `--package threat-modelling,git-workflow`)
3. **Update `buildDirList` signature** to accept `packageFlag string` instead of `includeAgents bool` and `skillsFlag string`
4. **Update `buildDirList` logic:**
   - When `--package` is provided: validate each package exists under `packages/<name>/` in the embedded FS, then walk the package dir to find its subdirectories
   - When no flags: install all packages (walk `packages/` to discover them all)

**Installer change needed:** The `Installer.Install` method currently copies files preserving the full embedded path. For packages, we need to **strip the `packages/<name>/` prefix** so that `packages/threat-modelling/agents/threat-modelling.agent.md` lands as `agents/threat-modelling.agent.md` in the target. This requires either:

- A new field on `Installer` (e.g., `StripPrefix string`), or
- A new method/parameter on `Install` that accepts a prefix to strip

---

### Step 4: Update `internal/installer/installer.go`

**What:** Add support for stripping a path prefix during installation.

**Approach:** Add a `StripPrefix` field to the `Installer` struct. In the `Install` method, when computing `targetPath`, strip this prefix from the embedded path before joining with the target directory.

**Example:**

- Embedded path: `packages/threat-modelling/agents/threat-modelling.agent.md`
- StripPrefix: `packages/threat-modelling`
- Target path becomes: `<target>/agents/threat-modelling.agent.md`

**Alternative:** Instead of modifying `Installer`, have `buildDirList` return a list of `{source, target}` pairs. This is more flexible but a bigger change.

---

### Step 5: Update `internal/cmd/list.go`

**What:** Add a "Packages" section to the `list` command output.

**Changes:**

- Read `packages/` directory from embedded FS
- For each package, list its name and optionally read the skill description from `packages/<name>/skills/<name>/SKILL.md`
- Consider whether to keep the separate "Agents" and "Skills" sections or replace with just "Packages" and "Standards"

---

### Step 6: Update tests

**Files to update:**

- `internal/cmd/install_test.go` — update `testContentFS()` to use `packages/` structure, update test cases for new `--package` flag
- `internal/installer/installer_test.go` — add tests for `StripPrefix` behaviour
- `internal/cmd/list_test.go` — update for new list output
- `embed_test.go` — validate `packages/` directory is embedded correctly

---

### Step 7: Update documentation

**Files to update:**

- `README.md` — update install command examples
- `agents/AGENTS.md` — review if this file is still needed or should move

---

## Open Questions

1. **Naming mismatches:** What should the package name be for `devcontainer.agent.md` + `skills/creating-devcontainers/`? Options: rename the agent to `creating-devcontainers.agent.md`, rename the skill to `devcontainer/`, or pick a new name.

2. **`agents/AGENTS.md`:** This is a shared routing file, not specific to one agent. Should it be installed with every package? Or become part of a "base" install? Or be removed?

3. **Install-all behaviour:** When no flags are given (`code-minions install`), should it install all packages + all standards? Or require explicit flags?

4. **Path prefix stripping:** Should we modify `Installer` with a `StripPrefix` field, or refactor `Install` to accept source→target mappings? (Simpler vs. more flexible)

---

## Command Examples (After Implementation)

```bash
# Install a single package
code-minions install --package threat-modelling

# Install multiple packages
code-minions install --package threat-modelling,git-workflow

# Install language standards
code-minions install --standards python,bash

# Install a package + standards together
code-minions install --package git-workflow --standards bash

# Install everything
code-minions install

# Dry run
code-minions install --package threat-modelling --dry-run

# List available packages and standards
code-minions list
```
