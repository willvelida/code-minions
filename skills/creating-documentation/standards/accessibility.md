# Documentation Accessibility Standard

## Overview

Accessibility requirements for documentation to ensure all users can read and understand content, including those using assistive technologies.

---

## Why Accessibility Matters

Accessible documentation:
- Invites trust and exploration into your repository
- Increases clarity and usability for everyone
- Demonstrates commitment to inclusivity
- Builds good habits for all Markdown files

---

## Heading Structure

### Rules

| Rule | Requirement |
|------|-------------|
| Single H1 | Only one H1 heading (the title) per document |
| No skipping | Don't skip levels (H1 → H3 is wrong) |
| Hierarchy | H1 → H2 → H3 → H4 in order |
| Descriptive | Headings describe section content |

### Correct Example

```markdown
# Project Name          <!-- H1: Document title -->

## Installation         <!-- H2: Major section -->

### Prerequisites       <!-- H3: Subsection -->

### Install Steps       <!-- H3: Subsection -->

## Usage               <!-- H2: Major section -->

### Basic Usage        <!-- H3: Subsection -->

### Advanced Usage     <!-- H3: Subsection -->
```

### Incorrect Example

```markdown
# Project Name

#### Installation      <!-- ❌ Skipped H2 and H3 -->

## Usage

##### Basic Usage      <!-- ❌ Skipped H3 and H4 -->
```

**Why it matters:** Screen readers announce heading levels. Skipped levels confuse users about document structure.

---

## Link Text

### Rules

| Rule | Requirement |
|------|-------------|
| Descriptive | Link text describes the destination |
| No "click here" | Never use "click here", "here", "this link" |
| Unique | Different destinations have different link text |
| Contextual | Include context about where link goes |

### Correct Examples

```markdown
Read the [installation guide](docs/install.md) for detailed setup instructions.

See [GitHub's documentation on READMEs](https://docs.github.com/...) for more information.

The [API reference](api.md) contains all available methods.
```

### Incorrect Examples

```markdown
Click [here](docs/install.md) for installation.     <!-- ❌ "here" -->

For more info, see [this link](https://...).        <!-- ❌ "this link" -->

Read more [here](link1) and [here](link2).          <!-- ❌ Duplicate "here" -->
```

**Why it matters:** Screen readers present links in isolation. Users need to understand destinations without surrounding context.

---

## Image Alt Text

### Rules

| Rule | Requirement |
|------|-------------|
| All images | Every image must have alt text |
| Descriptive | Alt text describes image content/purpose |
| Concise | Keep alt text brief but informative |
| No redundancy | Don't start with "image of" or "photo of" |
| Screenshots | DO include "screenshot of..." for context |

### Correct Examples

```markdown
![Screenshot of the dashboard showing three panels: metrics, logs, and alerts](dashboard.png)

![PDFTools logo: a red document icon with white text](logo.png)

![Architecture diagram showing client, API server, and database connections](architecture.png)
```

### Incorrect Examples

```markdown
![](image.png)                          <!-- ❌ No alt text -->

![logo](logo.png)                       <!-- ❌ Not descriptive -->

![Image of the dashboard](dashboard.png) <!-- ❌ Starts with "Image of" -->
```

### Long Descriptions

For complex images (charts, detailed diagrams), use a details block:

```markdown
![Quarterly sales chart showing growth](chart.png)

<details>
<summary>Chart data details</summary>

| Quarter | Sales |
|---------|-------|
| Q1 | $100k |
| Q2 | $150k |
| Q3 | $200k |
| Q4 | $250k |

</details>
```

**Why it matters:** Screen readers read alt text to describe images. Without it, users miss visual information.

---

## Lists

### Rules

| Rule | Requirement |
|------|-------------|
| Proper syntax | Use Markdown list syntax (`-`, `*`, or `1.`) |
| No decorative bullets | Don't use emoji or symbols as bullets |
| Consistent markers | Use same marker style throughout a list |

### Correct Example

```markdown
## Features

- Fast PDF processing
- OCR support for scanned documents
- Batch processing capabilities
```

### Incorrect Example

```markdown
## Features

✨ Fast PDF processing
🔍 OCR support for scanned documents
📦 Batch processing capabilities
```

**Why it matters:** Screen readers recognize Markdown lists and announce them properly (e.g., "list of 3 items, item 1 of 3"). Custom bullets break this.

---

## Emoji Usage

### Rules

| Rule | Requirement |
|------|-------------|
| Sparingly | Use emoji sparingly, not excessively |
| Not critical | Don't convey critical info with emoji alone |
| No sequences | Avoid long emoji sequences |
| Consider variations | Some emoji render differently across platforms |

### Correct Example

```markdown
## Contributing 🤝

We welcome contributions! See our guide for details.
```

### Incorrect Example

```markdown
## 🚀🔥💯 FEATURES 🔥💯🚀

⚠️ IMPORTANT: You MUST install dependencies first!
```

**Why it matters:** Screen readers read emoji names aloud. "Face with stuck-out tongue and squinting eyes" repeated multiple times is jarring.

---

## Plain Language

### Rules

| Rule | Requirement |
|------|-------------|
| Short sentences | Aim for under 25 words per sentence |
| Simple words | Use common words over jargon |
| Active voice | "The function returns..." not "...is returned" |
| Define terms | Explain or link unfamiliar concepts |

### Correct Example

```markdown
Install the package using npm:

```bash
npm install project-name
```

This adds project-name to your dependencies.
```

### Incorrect Example

```markdown
The package installation procedure necessitates the utilization of the npm 
package management system, whereupon execution of the installation command 
will effectuate the addition of the requisite dependencies to your project 
configuration manifest.
```

---

## Testing Tools

### Editor Extensions

- [markdownlint for VS Code](https://marketplace.visualstudio.com/items?itemName=DavidAnson.vscode-markdownlint)
- [github-markdown-a11y-extension](https://github.com/iansan5653/github-markdown-a11y-extension)

### Writing Helpers

- [Hemingway App](https://hemingwayapp.com/) — Readability checker
- [Grammarly](https://www.grammarly.com/) — Grammar and clarity
- [Alex](https://alexjs.com/) — Inclusive language checker

### GitHub Actions

```yaml
on: [push, pull_request]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: DavidAnson/markdownlint-cli2-action@v11
      with:
        globs: |
          *.md
          !test/*.md
```
