---
description: "Stage and commit changes with conventional commit messages"
mode: agent
input:
  - name: scope
    description: "The scope or area of changes (e.g. auth, api, docs)"
---

# Commit Changes

Stage and commit the current changes with a conventional commit message.

## Scope

${input:scope}

## Instructions

1. **Review the current changes** using `git diff` and `git status`

2. **Determine the commit type** based on the changes:
   - `feat:` — new feature
   - `fix:` — bug fix
   - `docs:` — documentation only
   - `refactor:` — code restructuring without behaviour change
   - `test:` — adding or updating tests
   - `chore:` — maintenance, dependencies, tooling

3. **Write a conventional commit message**:
   - Format: `type(scope): concise description`
   - Subject line ≤ 72 characters
   - Use imperative mood ("add" not "added")
   - Include a body if the change needs explanation

4. **Stage the relevant files** — group related changes, avoid mixing concerns

5. **Commit** with the constructed message
