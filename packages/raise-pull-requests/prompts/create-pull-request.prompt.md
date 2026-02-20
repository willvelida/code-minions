---
description: "Generate a pull request with structured description"
mode: agent
input:
  - name: target_branch
    description: "Target branch for the PR (e.g. main, develop)"
---

# Create Pull Request

Generate a pull request targeting **${input:target_branch}**.

## Instructions

1. **Analyse the changes** on the current branch vs ${input:target_branch}:
   - Review `git log` for commit history
   - Review `git diff` for full changeset

2. **Generate a PR title** using conventional format:
   - Format: `type(scope): concise description`
   - Match the primary commit type

3. **Generate a structured PR description** with:
   - **Summary** — what this PR does and why
   - **Changes** — bullet list of specific changes
   - **Testing** — how changes were verified
   - **Checklist** — standard review items (tests pass, docs updated, etc.)

4. **Check PR readiness**:
   - Are there any uncommitted changes?
   - Is the branch up to date with ${input:target_branch}?
   - Are all tests passing?

5. **Create the PR** using the GitHub CLI or provide the command to do so
