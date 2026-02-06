# Documentation Writing Style Standard

## Overview

Guidelines for writing clear, concise, and user-friendly documentation. Good writing helps users understand your project quickly and reduces support burden.

---

## Core Principles

### 1. Clarity Over Cleverness

Write to inform, not to impress. Choose the clearest way to express an idea.

| Instead of | Write |
|------------|-------|
| "Utilize" | "Use" |
| "Facilitate" | "Help" or "Enable" |
| "Leverage" | "Use" |
| "Implement" | "Add" or "Create" |
| "Instantiate" | "Create" |

### 2. Audience Awareness

Know your reader:
- **New users**: Need context, simple examples, prerequisites
- **Experienced users**: Need reference material, advanced options
- **Contributors**: Need architecture, coding standards, workflow

### 3. Progressive Disclosure

Present information in order of importance:
1. What does this do? (Immediate)
2. How do I use it? (Primary)
3. How does it work? (Secondary)
4. How can I extend it? (Advanced)

---

## Sentence Structure

### Keep Sentences Short

| Guideline | Target |
|-----------|--------|
| Ideal length | 15-20 words |
| Maximum length | 25 words |
| Paragraphs | 3-5 sentences |

### Use Active Voice

Active voice is clearer and more direct.

| Passive (avoid) | Active (preferred) |
|-----------------|-------------------|
| "The file is created by the function" | "The function creates the file" |
| "Errors are thrown when..." | "The method throws errors when..." |
| "The data is processed" | "The processor handles the data" |

### Front-Load Important Information

Put the key information first:

**Good:** "Run `npm install` to install dependencies."

**Poor:** "In order to proceed with the installation of the required dependencies for this project, you will need to run the npm install command."

---

## Technical Writing Guidelines

### Code Examples

1. **Keep examples minimal** — Show the smallest working example
2. **Use realistic values** — `user@example.com`, not `foo@bar.baz`
3. **Show output** — Include expected results when helpful
4. **Make it runnable** — Users should be able to copy and run

```javascript
// Good: Minimal, realistic, shows output
const user = getUser('alice@example.com');
console.log(user.name); // "Alice Smith"
```

```javascript
// Bad: Too much setup, unrealistic, no output
const config = { /* many options */ };
const client = new Client(config);
const connectionOptions = { /* more options */ };
await client.connect(connectionOptions);
const query = buildQuery({ /* options */ });
const result = await client.execute(query);
// ... what does result contain?
```

### Command Line Examples

- Use `$` prefix for shell commands (don't include in copy)
- Show realistic output where helpful
- Separate commands from output

```bash
$ npm install project-name
added 42 packages in 2.3s

$ npm test
✓ All tests passed (23 tests)
```

### Placeholders

Use clear placeholder syntax:

| Style | Example | Use for |
|-------|---------|---------|
| `<brackets>` | `git clone <repository-url>` | Required values |
| `[brackets]` | `command [--optional-flag]` | Optional values |
| `UPPER_SNAKE` | `export API_KEY=YOUR_API_KEY` | Environment variables |

---

## Tone and Voice

### Be Welcoming

Documentation is often a user's first interaction with your project.

**Good:**
```markdown
## Contributing

We welcome contributions! Whether you're fixing bugs, improving docs, 
or proposing new features, we'd love to hear from you.
```

**Poor:**
```markdown
## Contributing

Read the guidelines. Follow the process. PRs without tests will be rejected.
```

### Be Confident but Not Arrogant

**Good:** "This library provides fast PDF processing."

**Poor:** "This is the fastest PDF library ever created!"

### Avoid Assumptions

Don't assume knowledge. Link to explanations for unfamiliar concepts.

**Good:** "Clone the repository using [git](https://git-scm.com/)."

**Poor:** "Clone the repo." (assumes user knows git)

---

## Common Patterns

### Prerequisites Section

```markdown
## Prerequisites

Before you begin, ensure you have:

- Node.js 18 or higher ([download](https://nodejs.org/))
- npm or yarn package manager
- A GitHub account (for contributing)
```

### Installation Section

```markdown
## Installation

Install via npm:

```bash
npm install project-name
```

Or with yarn:

```bash
yarn add project-name
```
```

### Configuration Section

```markdown
## Configuration

Create a config file at `~/.project/config.json`:

```json
{
  "apiKey": "your-api-key",
  "timeout": 30000
}
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `apiKey` | string | required | Your API key |
| `timeout` | number | `30000` | Request timeout in ms |
```

---

## Things to Avoid

### Filler Words and Phrases

Remove these without losing meaning:
- "In order to" → "To"
- "It is important to note that" → (delete)
- "As a matter of fact" → (delete)
- "Basically" → (delete)
- "Simply" → (delete) — what's simple to you may not be simple to others

### Ambiguous Pronouns

Be specific about what "it", "this", and "that" refer to.

**Ambiguous:** "After installing the package and configuring it, run it."

**Clear:** "After installing the package and configuring the options, run the application."

### Absolute Time References

Avoid:
- "Recently added feature"
- "Coming soon"
- "Last month we..."
- "As of 2024..."

Use version numbers instead:
- "Added in v2.0"
- "Planned for v3.0"
- "Deprecated in v1.5"
