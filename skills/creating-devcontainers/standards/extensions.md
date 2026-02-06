# DevContainer Extensions Standard

## Overview

Requirements for VS Code extensions in DevContainer configurations.

---

## Essential Extensions

All DevContainers **MUST** include these minimum required extensions, organised into logical bundles with documentation comments:

```jsonc
{
  "customizations": {
    "vscode": {
      "extensions": [
        /* ---------------------------------------------------------------------------
           GitHub Copilot
        --------------------------------------------------------------------------- */
        "GitHub.copilot",
        // Provides AI-assisted code completions and inline suggestions tailored to
        // the context of the file you're editing.

        "GitHub.copilot-chat",
        // Adds a chat-based Copilot interface for conversational guidance and
        // code explanations.

        /* ---------------------------------------------------------------------------
           Markdown and Documentation
        --------------------------------------------------------------------------- */
        "yzhang.markdown-all-in-one",
        // Common markdown conveniences: shortcuts, TOC generation, export commands.

        "DavidAnson.vscode-markdownlint",
        // Markdownlint rules for consistent style guidelines.

        "bierner.markdown-yaml-preamble",
        // Render YAML front matter in markdown preview.

        /* ---------------------------------------------------------------------------
           Code Quality and Spelling
        --------------------------------------------------------------------------- */
        "streetsidesoftware.code-spell-checker",
        // Inline spell checking for technical writing.

        "streetsidesoftware.code-spell-checker-australian-english",
        // Australian English dictionary for Code Spell Checker.

        /* ---------------------------------------------------------------------------
           Formatting
        --------------------------------------------------------------------------- */
        "esbenp.prettier-vscode"
        // Prettier formatter for markdown and other supported formats.
      ]
    }
  }
}
```

---

## Language Extensions

Language-specific extensions are defined in `standards/languages/<language>/development-environment.md`. When creating or reviewing a DevContainer, add the extensions for each included language.

---

## Extension Guidelines

### Review Guidelines

When reviewing existing DevContainers:

| Action | When |
|--------|------|
| **Add** | Essential extensions are missing |
| **Keep** | Project-relevant extensions that benefit the team |
| **Remove** | Only if extension is harmful, deprecated, or creates security vulnerabilities |

**Important**: Do NOT aggressively remove existing extensions. The essential extensions list is a **minimum requirement**, not a maximum.

### Creation Guidelines

When creating a new DevContainer, use ONLY:
- Essential extensions (listed above)
- Language-specific extensions from language standards

Personal tools belong in developer's VS Code user `settings.json`, not shared DevContainer configurations. Work with the developer to configure their user `settings.json` with the tools that are not part of the standards relevant to this project or repository.

### Bundling and Documentation

Extensions **MUST** be:

1. **Grouped into logical bundles** under descriptive headers
2. **Documented with comments** explaining purpose

#### Bundle Header Format

```jsonc
/* ---------------------------------------------------------------------------
   Bundle Name (e.g., "Python Development", "Database Tools")
--------------------------------------------------------------------------- */
```

#### Extension Comment Format

```jsonc
"extension.identifier",
// Brief description of what the extension provides.
```
