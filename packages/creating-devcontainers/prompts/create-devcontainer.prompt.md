---
description: "Generate a devcontainer configuration for a project"
mode: agent
input:
  - name: project_type
    description: "Type of project (e.g. Go CLI, Node.js API, Python ML, .NET web app)"
  - name: features
    description: "Comma-separated list of required features (e.g. docker-in-docker, git, github-cli)"
---

# Create DevContainer

Generate a devcontainer configuration for a **${input:project_type}** project.

## Required Features

${input:features}

## Requirements

1. Create `.devcontainer/devcontainer.json` with:
   - Appropriate base image for the project type
   - Required VS Code extensions
   - Port forwarding configuration
   - Post-create commands for dependency installation

2. Include the requested features as devcontainer features

3. Follow security best practices:
   - Use non-root user
   - Pin image versions
   - Minimise installed tooling

4. Add lifecycle scripts for environment setup

5. Include comments explaining configuration choices
