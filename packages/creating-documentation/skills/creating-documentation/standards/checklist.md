# Documentation Compliance Checklist

## Purpose

Consolidated checklist for validating repository documentation against quality, accessibility, and best practice requirements.

---

## README Structure

### Required Sections

- [ ] Title (H1) present and matches project name
- [ ] Description explains what the project does
- [ ] Installation section with step-by-step instructions
- [ ] Usage section with code examples
- [ ] License section or reference to LICENSE file

### Recommended Sections

- [ ] Prerequisites/Requirements listed if needed
- [ ] Contributing section or link to CONTRIBUTING.md
- [ ] Support/Contact information provided
- [ ] Badges (if applicable) are functional

### Structure Quality

- [ ] Sections in logical order
- [ ] Table of contents for long documents (300+ lines)
- [ ] No orphaned or empty sections

---

## Accessibility

### Heading Structure

- [ ] Single H1 (document title only)
- [ ] No skipped heading levels (H1 → H2 → H3)
- [ ] Headings are descriptive
- [ ] Heading hierarchy is logical

### Images

- [ ] All images have alt text
- [ ] Alt text is descriptive (not just "image" or "logo")
- [ ] Complex images have extended descriptions
- [ ] No critical information conveyed only through images

### Links

- [ ] No "click here" or "here" link text
- [ ] Link text describes destination
- [ ] No duplicate link text for different URLs
- [ ] All links are functional (not broken)

### Lists

- [ ] Proper Markdown list syntax used
- [ ] No emoji or symbols as bullet points
- [ ] Consistent list markers throughout

### Emoji

- [ ] Emoji used sparingly
- [ ] Critical information not conveyed by emoji alone
- [ ] No long emoji sequences

---

## Writing Quality

### Clarity

- [ ] Purpose is clear within first paragraph
- [ ] Technical terms are explained or linked
- [ ] No assumed knowledge without explanation
- [ ] Active voice used (not passive)

### Conciseness

- [ ] Sentences under 25 words (generally)
- [ ] Paragraphs under 5 sentences
- [ ] No filler words or phrases
- [ ] Information is not duplicated

### Consistency

- [ ] Consistent terminology throughout
- [ ] Consistent capitalization of terms
- [ ] Consistent code style in examples
- [ ] Consistent formatting patterns

### Accuracy

- [ ] Installation instructions work as written
- [ ] Code examples are runnable
- [ ] Links point to correct destinations
- [ ] Version numbers are current

---

## Markdown Quality

### Code Blocks

- [ ] All code blocks have language tags
- [ ] Code examples are properly formatted
- [ ] Commands are distinguishable from output
- [ ] Placeholders are clearly marked

### Formatting

- [ ] Consistent list marker style
- [ ] Proper heading spacing (blank lines)
- [ ] Tables are properly formatted
- [ ] No raw URLs (all links are formatted)

### Links and References

- [ ] Relative links for repo files
- [ ] Absolute links for external sites
- [ ] No broken internal links
- [ ] Section links use correct anchors

---

## Content Completeness

### For Users

- [ ] Can a new user understand what this project does?
- [ ] Can a new user install the project from README alone?
- [ ] Can a new user run basic examples from README?
- [ ] Is there guidance on where to get help?

### For Contributors

- [ ] Is there a CONTRIBUTING.md or contributing section?
- [ ] Are code standards documented?
- [ ] Is the development setup explained?
- [ ] Is the review process documented?

### For Maintainers

- [ ] Is the license clearly stated?
- [ ] Is project status indicated?
- [ ] Are maintainers identified?
- [ ] Is the release process documented (if applicable)?

---

## Anti-Patterns to Avoid

- [ ] No "click here" or "here" links
- [ ] No skipped heading levels
- [ ] No images without alt text
- [ ] No emoji as list bullets
- [ ] No time-sensitive language ("recently", "coming soon")
- [ ] No undefined abbreviations
- [ ] No passive voice overuse
- [ ] No overly long sentences (>25 words)
- [ ] No broken or outdated links
- [ ] No placeholder content ("Lorem ipsum", "TODO")

---

## Severity Guide

| Severity | Impact | Action |
|----------|--------|--------|
| 🔴 Critical | Blocks understanding/usage | Must fix |
| 🟠 High | Significant usability impact | Should fix |
| 🟡 Medium | Minor usability impact | Consider fixing |
| 🟢 Low | Style/preference | Nice to have |

### Critical Issues (🔴)

- Missing installation instructions
- Missing usage examples
- Broken code examples
- Missing alt text on key images
- Skipped heading levels
- No license information

### High Priority (🟠)

- "Click here" link text
- Missing prerequisites
- Outdated information
- No contribution guidelines
- Passive voice overuse

### Medium Priority (🟡)

- Missing badges
- No table of contents for long docs
- Inconsistent formatting
- Minor accessibility issues

### Low Priority (🟢)

- Style inconsistencies
- Missing optional sections
- Verbose language
