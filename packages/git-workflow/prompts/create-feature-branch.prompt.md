---
description: "Create a feature branch with conventional naming"
mode: agent
input:
  - name: feature_description
    description: "Brief description of the feature to implement"
---

# Create Feature Branch

Create a new feature branch for the following work:

${input:feature_description}

## Instructions

1. **Ensure the working tree is clean** — check `git status`

2. **Pull the latest from the default branch** (main or master)

3. **Derive a branch name** using conventional format:
   - Format: `feat/<short-kebab-description>`
   - Keep it concise — 3-5 words max
   - Use lowercase with hyphens

4. **Create and switch to the branch**:
   ```
   git checkout -b feat/<branch-name>
   ```

5. **Confirm** the branch was created and you're on it
