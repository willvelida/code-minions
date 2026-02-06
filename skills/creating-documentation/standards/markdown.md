# Markdown Formatting Standard

## Overview

Best practices for Markdown formatting in documentation to ensure consistency, readability, and proper rendering across platforms.

---

## Code Blocks

### Always Use Language Tags

Specify the language for syntax highlighting:

````markdown
```javascript
const x = 1;
```
````

**Common language tags:**
- `javascript`, `typescript`, `python`, `bash`, `json`, `yaml`, `markdown`
- `jsx`, `tsx` for React code
- `shell` or `bash` for terminal commands
- `text` or `plaintext` for no highlighting

### Inline Code

Use backticks for:
- File names: `README.md`
- Commands: `npm install`
- Code references: `const`, `function`, `true`
- Paths: `/usr/local/bin`

Don't use backticks for:
- Emphasis (use **bold** or *italic*)
- Product names
- General technical terms

---

## Links

### Relative vs Absolute Links

| Link Type | When to Use | Example |
|-----------|-------------|---------|
| Relative | Files in same repo | `[Guide](docs/guide.md)` |
| Absolute | External sites | `[GitHub](https://github.com)` |

Relative links work across branches and forks.

### Link Syntax

```markdown
<!-- Inline link -->
[Link text](url)

<!-- Reference link -->
[Link text][ref]

[ref]: url "Optional title"
```

### Section Links

Link to headings within the same document:

```markdown
See the [Installation](#installation) section.
```

GitHub converts headings to lowercase anchors with hyphens:
- `## Getting Started` → `#getting-started`
- `## API Reference` → `#api-reference`

---

## Tables

### Basic Table Syntax

```markdown
| Column 1 | Column 2 | Column 3 |
|----------|----------|----------|
| Data 1   | Data 2   | Data 3   |
| Data 4   | Data 5   | Data 6   |
```

### Column Alignment

```markdown
| Left | Center | Right |
|:-----|:------:|------:|
| L    |   C    |     R |
```

### Table Guidelines

- Keep tables simple (max 4-5 columns)
- Use consistent column widths
- Align numeric data right
- Consider lists for complex data

---

## Lists

### Unordered Lists

Use `-` or `*` consistently:

```markdown
- Item one
- Item two
  - Nested item
  - Another nested
- Item three
```

### Ordered Lists

```markdown
1. First step
2. Second step
3. Third step
```

Or let Markdown auto-number:

```markdown
1. First step
1. Second step
1. Third step
```

### Task Lists

```markdown
- [x] Completed task
- [ ] Incomplete task
- [ ] Another task
```

---

## Headings

### ATX Style (Preferred)

```markdown
# Heading 1
## Heading 2
### Heading 3
```

### Guidelines

- Use ATX style (`#`) not underline style
- Add blank line before and after headings
- Don't end headings with punctuation
- Keep headings concise

---

## Emphasis

| Style | Syntax | Use for |
|-------|--------|---------|
| **Bold** | `**text**` | Important terms, warnings |
| *Italic* | `*text*` | Technical terms, emphasis |
| ~~Strikethrough~~ | `~~text~~` | Deprecated content |
| `Code` | `` `text` `` | Code, commands, file names |

**Don't combine excessively:** ~~***`text`***~~ is hard to read.

---

## Blockquotes

Use for:
- Quotes from documentation
- Important notes
- Callouts

```markdown
> **Note:** This is an important note.

> **Warning:** This action cannot be undone.
```

### GitHub Alerts (GitHub-specific)

```markdown
> [!NOTE]
> Useful information.

> [!TIP]
> Helpful advice.

> [!IMPORTANT]
> Key information.

> [!WARNING]
> Urgent information.

> [!CAUTION]
> Negative potential consequences.
```

---

## Images

### Basic Syntax

```markdown
![Alt text](image-url)
```

### With Title

```markdown
![Alt text](image-url "Image title")
```

### Sizing (GitHub/HTML)

GitHub Markdown doesn't support image sizing. Use HTML when needed:

```html
<img src="image.png" alt="Description" width="400">
```

### Image Guidelines

- Always include descriptive alt text
- Use relative paths for repo images
- Optimize image file sizes
- Consider dark/light mode compatibility

---

## Horizontal Rules

Use for major section breaks (sparingly):

```markdown
---
```

or

```markdown
***
```

---

## Escaping Characters

Escape special Markdown characters with backslash:

```markdown
\*not italic\*
\# not a heading
\[not a link\]
```

---

## File Naming

| Convention | Example |
|------------|---------|
| README | `README.md` (uppercase) |
| Other docs | `lowercase-with-hyphens.md` |
| No spaces | Use hyphens, not spaces |
| Descriptive | `installation-guide.md` not `doc1.md` |

---

## Common Mistakes

### Inconsistent Formatting

❌ Mixed list markers:
```markdown
* Item one
- Item two
+ Item three
```

✅ Consistent markers:
```markdown
- Item one
- Item two
- Item three
```

### Missing Blank Lines

❌ No spacing:
```markdown
## Heading
Content immediately after.
```

✅ Proper spacing:
```markdown
## Heading

Content after blank line.
```

### Raw URLs

❌ Raw URL:
```markdown
Check out https://example.com for more.
```

✅ Linked text:
```markdown
Check out [the documentation](https://example.com) for more.
```
