# Create README

## Purpose

Generate a comprehensive, accessible README file for a repository that helps users understand, install, and use the project.

---

## Flow

### Step 1: Determine Approach 🛑

Present two options to the user:

| Option | Description | Best For |
|--------|-------------|----------|
| **A: Automatic** | Analyse repository to detect project details | Existing repos with code |
| **B: Manual** | User provides project information | New repos, specific requirements |

**🛑 STOP**: Wait for the user to select Option A or Option B before proceeding.

**Success Criteria:**
- [ ] User has explicitly selected their preferred approach

---

### Step 2: Gather Project Information

#### Path A: Automatic Analysis

Analyse the repository to detect:

| Information | Detection Method |
|-------------|-----------------|
| Project name | Directory name, package.json, setup.py, etc. |
| Description | Existing README, package description |
| Language(s) | File extensions, config files |
| Dependencies | package.json, requirements.txt, Cargo.toml, etc. |
| Build system | Makefile, npm scripts, setup.py, etc. |
| License | LICENSE file |
| Tests | Test directories, test scripts |

Present findings to user for confirmation.

#### Path B: Manual Specification 🛑

Collect from user:

1. **Project name**: What is the project called?
2. **Description**: What does the project do? (1-2 sentences)
3. **Primary use case**: Who is this for and what problem does it solve?
4. **Installation method**: How do users install/set up the project?
5. **Basic usage**: How do users run/use the project?
6. **License**: What license applies? (MIT, Apache 2.0, etc.)

**🛑 STOP**: Wait for user to provide project information.

**Success Criteria:**
- [ ] Project name identified
- [ ] Description captured
- [ ] Installation method known
- [ ] Basic usage understood

---

### Step 3: Load Standards

Load from this skill's `standards/`:
- `readme-structure.md` — Section organization
- `accessibility.md` — Accessibility requirements
- `writing-style.md` — Plain language guidelines
- `markdown.md` — Formatting best practices

**Success Criteria:**
- [ ] All relevant standards loaded

---

### Step 4: Generate README Structure

Build README following `readme-structure.md`:

#### Required Sections

```markdown
# Project Name

Brief description of what the project does.

## Installation

How to install the project.

## Usage

How to use the project with examples.

## License

License information.
```

#### Recommended Sections (include if applicable)

```markdown
## Features          <!-- If multiple notable features -->
## Requirements      <!-- If specific dependencies needed -->
## Configuration     <!-- If configurable -->
## API Reference     <!-- If exposing an API -->
## Contributing      <!-- If accepting contributions -->
## Support           <!-- If support channels exist -->
## Acknowledgments   <!-- If crediting others -->
```

**Success Criteria:**
- [ ] All required sections included
- [ ] Recommended sections included where applicable
- [ ] Logical section ordering

---

### Step 5: Write Section Content

For each section, apply standards:

#### Title & Description
- Use the project name as H1 heading
- Write a clear, concise description (1-2 sentences)
- Optionally add badges for build status, version, license

#### Installation
- List prerequisites/requirements first
- Provide step-by-step installation commands
- Use code blocks with appropriate language tags
- Include verification step if applicable

#### Usage
- Start with the simplest possible example
- Show expected output where helpful
- Link to more complex examples if needed
- Use code blocks with syntax highlighting

#### License
- State the license type
- Link to the LICENSE file or full license text

**Success Criteria:**
- [ ] Each section follows writing style guidelines
- [ ] Code examples are complete and runnable
- [ ] Links are descriptive (not "click here")

---

### Step 6: Apply Accessibility Standards

Validate against `accessibility.md`:

| Check | Requirement |
|-------|-------------|
| Headings | Proper hierarchy (H1 → H2 → H3, no skipping) |
| Images | All images have descriptive alt text |
| Links | Descriptive text (not "here" or "click here") |
| Lists | Proper markdown list syntax (not decorative bullets) |
| Language | Plain language, short sentences |
| Emoji | Used sparingly, not for critical information |

**Success Criteria:**
- [ ] All accessibility checks pass

---

### Step 7: Review and Present 🛑

Present the generated README to the user:

1. Show the complete README content
2. Highlight key sections
3. Note any sections that may need user input

**🛑 STOP**: Wait for user feedback and approval.

**Success Criteria:**
- [ ] README presented to user
- [ ] User has approved or requested changes

---

### Step 8: Create File

Once approved, create the README.md file in the repository root.

**File Location:** `README.md` (repository root)

**Success Criteria:**
- [ ] README.md created in repository root
- [ ] File contains approved content
