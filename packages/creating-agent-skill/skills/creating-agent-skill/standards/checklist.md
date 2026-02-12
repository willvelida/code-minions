# Skill Compliance Checklist

## Purpose

Consolidated checklist for validating Agent Skills against specification requirements and quality guidelines.

---

## Requirements Gathering

- [ ] Clarifying questions asked before generation began
- [ ] Skill purpose stated in one clear sentence
- [ ] At least 3 trigger keywords identified
- [ ] Scope boundaries defined (what's in and what's out)
- [ ] Actions/workflows listed with high-level steps
- [ ] Structure complexity determined (minimal, standard, or full)
- [ ] Requirements summary confirmed by user

---

## Specification Compliance

### Frontmatter

- [ ] `name` field present
- [ ] `name` is 1-64 characters
- [ ] `name` contains only lowercase letters, numbers, hyphens
- [ ] `name` does not start or end with hyphen
- [ ] `name` does not contain consecutive hyphens (`--`)
- [ ] `name` does not contain reserved words (`anthropic`, `claude`)
- [ ] `name` matches parent directory name exactly
- [ ] `description` field present
- [ ] `description` is 1-1024 characters
- [ ] `description` is non-empty

### Structure

- [ ] `SKILL.md` file exists in skill root
- [ ] All referenced files exist
- [ ] File references use forward slashes (not backslashes)
- [ ] File references are one level deep (no nested chains)

---

## Description Quality

- [ ] Written in third person voice
- [ ] Describes what the skill does
- [ ] Describes when to use the skill
- [ ] Contains specific trigger keywords
- [ ] Concise but informative (100-300 chars ideal)

---

## Instructions Quality

- [ ] SKILL.md body under 500 lines
- [ ] Instructions are concise (no over-explaining)
- [ ] Terminology is consistent throughout
- [ ] No time-sensitive information (or properly isolated)
- [ ] Appropriate degree of freedom for task type
- [ ] Examples provided where helpful

---

## Action Files (If Present)

- [ ] Each action has clear purpose statement
- [ ] Steps are numbered and sequential
- [ ] Stop points (🛑) marked where user input required
- [ ] Success criteria provided per step
- [ ] Appropriate detail level (not over-explaining)

---

## Standards Files (If Present)

- [ ] Each standard has clear purpose
- [ ] Rules are specific and measurable
- [ ] Good and bad examples provided
- [ ] Consistent formatting across files

---

## Scripts (If Present)

- [ ] Scripts handle errors explicitly
- [ ] No "magic numbers" without explanation
- [ ] Dependencies clearly documented
- [ ] Unix-style paths (forward slashes)
- [ ] Clear documentation/comments

---

## Progressive Disclosure

- [ ] Metadata (name + description) is ~100 tokens
- [ ] SKILL.md body is <5000 tokens (recommended)
- [ ] Detailed content moved to separate files
- [ ] Reference files focused and specific

---

## Anti-Patterns Avoided

- [ ] No Windows-style paths (backslashes)
- [ ] No excessive options without clear default
- [ ] No vague names (`helper`, `utils`, `tools`)
- [ ] No deeply nested reference chains
- [ ] No time-sensitive dates or versions
- [ ] No inconsistent terminology

---

## Testing

- [ ] Skill activates on expected trigger keywords
- [ ] Instructions are clear and followable
- [ ] Referenced files load correctly
- [ ] Scripts execute without errors (if present)

---

## Final Validation

Run the skills-ref validator:

```bash
skills-ref validate ./skill-name
```

- [ ] No validation errors
- [ ] No validation warnings (or warnings acknowledged)
