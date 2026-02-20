---
description: Standards for devcontainer configuration files
applyTo: "**/.devcontainer/*.json,**/devcontainer.json"
---

# DevContainer Configuration Standards

When creating or editing devcontainer configuration files:

- Always include a `name` field that identifies the development environment
- Pin base images to specific tags or digests — never use `latest`
- Prefer features over manual Dockerfile installations for common tools
- Include only essential VS Code extensions — avoid personal preference extensions
- Set `remoteUser` to a non-root user for security
- Use `postCreateCommand` for project setup (dependency installation, build)
- Define port forwarding for any services the project exposes
- Include environment variables needed for development in `containerEnv`
- Add `mounts` for persistent caches (e.g. Go module cache, npm cache)
