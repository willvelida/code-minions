---
description: "Generate a comprehensive README for the current project"
mode: agent
input:
  - name: project_name
    description: "Name of the project"
---

# Create README

Generate a comprehensive README.md for **${input:project_name}**.

## Requirements

1. Analyse the repository structure, dependencies, and existing documentation

2. Generate a README with these sections:
   - **Title and badges** (build status, version, license)
   - **Description** — what the project does and why it exists
   - **Quick start** — minimum steps to get running
   - **Installation** — detailed setup instructions
   - **Usage** — common use cases with examples
   - **Configuration** — environment variables, config files
   - **Contributing** — how to contribute (or link to CONTRIBUTING.md)
   - **License** — license type and link

3. Follow writing style best practices:
   - Use active voice
   - Keep sentences concise
   - Include code examples with syntax highlighting
   - Use relative links for internal references

4. Ensure accessibility — add alt text to any images
