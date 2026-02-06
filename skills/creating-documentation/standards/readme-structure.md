# README Structure Standard

## Overview

Guidelines for organizing README content to maximize clarity and usability. A well-structured README helps users quickly find the information they need.

---

## Section Hierarchy

### Required Sections

Every README **must** include these sections in order:

| Section | Purpose | Content |
|---------|---------|---------|
| **Title** | Project name | H1 heading with project name |
| **Description** | What it does | 1-3 sentences explaining the project |
| **Installation** | How to set up | Step-by-step setup instructions |
| **Usage** | How to use | Basic usage with examples |
| **License** | Legal terms | License type and link |

### Recommended Sections

Include these sections when applicable:

| Section | When to Include | Content |
|---------|-----------------|---------|
| **Badges** | CI/CD, npm, etc. | Status badges after title |
| **Features** | Multiple notable features | Bulleted feature list |
| **Requirements** | Prerequisites needed | System requirements, dependencies |
| **Configuration** | Configurable options | Configuration instructions |
| **API Reference** | Exposing an API | API documentation or link |
| **Examples** | Complex usage | Additional usage examples |
| **Contributing** | Open to contributions | How to contribute |
| **Support** | Support channels exist | Where to get help |
| **Roadmap** | Future plans exist | Planned features |
| **Acknowledgments** | Credits due | Credits and thanks |
| **Changelog** | Version history | Link to CHANGELOG.md |

---

## Section Guidelines

### Title

```markdown
# Project Name
```

- Use exact project name
- Single H1 heading only
- No emoji in title (accessibility)

### Description

```markdown
Project Name is a [type] that [does what] for [audience].
```

**Good Example:**
```markdown
# PDFTools

PDFTools is a Python library for extracting text and tables from PDF documents. 
It supports scanned documents via OCR and can handle encrypted files.
```

**Bad Example:**
```markdown
# PDFTools

Welcome to PDFTools! 👋 This is the best PDF tool ever made!!!
```

### Badges (Optional)

Place immediately after title/description:

```markdown
# Project Name

Brief description.

[![Build Status](https://img.shields.io/...)](link)
[![License](https://img.shields.io/...)](link)
[![Version](https://img.shields.io/...)](link)
```

**Guidelines:**
- Maximum 5-6 badges
- Include alt text in badge images
- Order: Build → Coverage → Version → License → Downloads

### Installation

```markdown
## Installation

### Prerequisites

- Node.js 18 or higher
- npm or yarn

### Install

```bash
npm install project-name
```
```

**Guidelines:**
- List prerequisites first
- Use code blocks with language tags
- Provide multiple methods if applicable (npm, yarn, manual)
- Include verification command if possible

### Usage

```markdown
## Usage

```javascript
const project = require('project-name');

const result = project.doSomething('input');
console.log(result); // Expected output
```
```

**Guidelines:**
- Start with simplest possible example
- Show expected output
- Use complete, runnable code
- Link to more examples if needed

### License

```markdown
## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
```

**Guidelines:**
- State license type clearly
- Link to LICENSE file
- Keep brief

---

## Section Ordering

Recommended order for comprehensive READMEs:

1. Title
2. Badges (optional)
3. Description
4. Table of Contents (if README is long)
5. Features (optional)
6. Requirements/Prerequisites
7. Installation
8. Usage
9. Configuration (optional)
10. API Reference (optional)
11. Examples (optional)
12. Contributing
13. Support (optional)
14. Roadmap (optional)
15. Acknowledgments (optional)
16. License

---

## Table of Contents

Include a table of contents when README exceeds ~300 lines:

```markdown
## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
- [Configuration](#configuration)
- [Contributing](#contributing)
- [License](#license)
```

**Note:** GitHub auto-generates a table of contents accessible via the outline menu.

---

## File Placement

README.md location priority (GitHub checks in this order):

1. Repository root (recommended)
2. `.github/` directory
3. `docs/` directory

**Always place README.md in the repository root** for maximum visibility.
