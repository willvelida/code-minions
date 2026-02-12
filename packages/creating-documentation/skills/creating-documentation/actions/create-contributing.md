# Create CONTRIBUTING Guide

## Purpose

Generate a CONTRIBUTING.md file that helps potential contributors understand how to participate in the project effectively.

---

## Flow

### Step 1: Gather Project Context 🛑

Collect information about the project's contribution workflow:

| Information | Question |
|-------------|----------|
| Contribution types | What kinds of contributions are welcome? (bugs, features, docs) |
| Development setup | How do contributors set up their environment? |
| Code standards | Are there coding standards or style guides? |
| Testing | How should contributors test their changes? |
| Review process | How are contributions reviewed and merged? |
| Communication | Where should contributors ask questions? |

**🛑 STOP**: Wait for user to provide contribution workflow details.

**Success Criteria:**
- [ ] Contribution types identified
- [ ] Development workflow understood
- [ ] Review process documented

---

### Step 2: Load Standards

Load from this skill's `standards/`:
- `writing-style.md` — Plain language guidelines
- `markdown.md` — Formatting best practices
- `accessibility.md` — Accessibility requirements

**Success Criteria:**
- [ ] All relevant standards loaded

---

### Step 3: Generate CONTRIBUTING Structure

Build CONTRIBUTING.md with appropriate sections:

#### Standard Structure

```markdown
# Contributing to [Project Name]

Thank you for your interest in contributing!

## How Can I Contribute?

### Reporting Bugs
### Suggesting Features  
### Code Contributions

## Development Setup

### Prerequisites
### Installation
### Running Tests

## Style Guidelines

### Code Style
### Commit Messages

## Pull Request Process

## Code of Conduct

## Questions?
```

**Success Criteria:**
- [ ] All relevant sections included
- [ ] Structure matches project needs

---

### Step 4: Write Section Content

#### How Can I Contribute?

Describe accepted contribution types:
- Bug reports (with issue template guidance)
- Feature requests (with discussion process)
- Documentation improvements
- Code contributions

#### Development Setup

Provide step-by-step setup instructions:
1. Fork and clone
2. Install dependencies
3. Configure environment
4. Run tests to verify setup

#### Style Guidelines

Document:
- Code formatting requirements
- Commit message conventions
- Documentation standards

#### Pull Request Process

Explain:
1. Branch naming conventions
2. Required checks (tests, linting)
3. Review process
4. Merge criteria

**Success Criteria:**
- [ ] Each section provides actionable guidance
- [ ] Instructions are complete and testable
- [ ] Tone is welcoming and inclusive

---

### Step 5: Review and Present 🛑

Present the generated CONTRIBUTING.md to the user.

**🛑 STOP**: Wait for user feedback and approval.

**Success Criteria:**
- [ ] CONTRIBUTING.md presented to user
- [ ] User has approved or requested changes

---

### Step 6: Create File

Once approved, create the CONTRIBUTING.md file.

**File Locations:**
- Primary: `CONTRIBUTING.md` (repository root)
- Alternative: `.github/CONTRIBUTING.md`

**Success Criteria:**
- [ ] CONTRIBUTING.md created
- [ ] File contains approved content
