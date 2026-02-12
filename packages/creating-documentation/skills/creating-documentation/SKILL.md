---
name: creating-documentation
description: 'Create and review README files and repository documentation that follows best practices for clarity, accessibility, and discoverability. Use when a user asks to create a README, write project documentation, improve existing documentation, add a CONTRIBUTING guide, review documentation for accessibility, or document a new repository.'
license: MIT
allowed-tools: Bash
---

# Creating Documentation

## Overview

This skill provides capabilities for creating and reviewing README files and other repository documentation that follows established best practices for structure, accessibility, and user experience.

Good documentation is the front door to your project. It helps users understand what your project does, how to use it, and how to contribute. This skill ensures documentation is clear, accessible, and follows industry standards.

## Capabilities

| Capability | Action | Description |
|------------|--------|-------------|
| Create README | `actions/create-readme.md` | Generate a comprehensive README for a repository |
| Create CONTRIBUTING | `actions/create-contributing.md` | Generate contribution guidelines |
| Review | `actions/review-documentation.md` | Analyse existing documentation for quality and accessibility |

## Standards

This skill bundles the following standards in `standards/`:

| Standard | File | Description |
|----------|------|-------------|
| README Structure | `readme-structure.md` | Essential sections and organization for README files |
| Accessibility | `accessibility.md` | Accessibility requirements for inclusive documentation |
| Writing Style | `writing-style.md` | Plain language and readability guidelines |
| Markdown | `markdown.md` | Markdown formatting best practices |
| Checklist | `checklist.md` | Consolidated compliance and quality checklist |

## Principles

### 1. Audience First

Write for your audience. A README's primary purpose is to help visitors quickly understand your project and get started. Don't assume prior knowledge—provide context and links for unfamiliar concepts.

### 2. Progressive Disclosure

Present information in order of importance:
- **Title & Description**: What is this project? (immediate)
- **Installation & Usage**: How do I use it? (primary)
- **Contributing & License**: How can I help? (secondary)

### 3. Accessibility Matters

Documentation should be accessible to all users:
- Use proper heading hierarchy (don't skip levels)
- Provide alt text for all images
- Write descriptive link text (not "click here")
- Use plain language and short sentences

### 4. Keep It Current

Outdated documentation is worse than no documentation. Include only information that can be maintained. Use automation where possible.

## Usage

1. Load this skill manifest
2. Identify the required capability (create or review)
3. Load the bundled standards from `standards/`
4. Execute the action following `actions/<capability>.md`

## References

- [Make a README](https://www.makeareadme.com/)
- [GitHub - About READMEs](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-readmes)
- [5 Tips for Making Your GitHub Profile Page Accessible](https://github.blog/developer-skills/github/5-tips-for-making-your-github-profile-page-accessible/)
- [Choose a License](https://choosealicense.com/)
- [Keep a Changelog](https://keepachangelog.com/)
