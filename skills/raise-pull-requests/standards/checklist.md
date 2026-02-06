# Pull Request Compliance Checklist

## Overview

Consolidated checklist for validating pull request compliance across all actions. Use this checklist to verify a PR is ready for submission and review.

---

## Preparation

- [ ] PR represents one logical change
- [ ] No unrelated changes mixed in (formatting, refactoring, dependency updates)
- [ ] Branch is synced with the target branch
- [ ] All merge conflicts resolved
- [ ] PR size within guidelines (< 200 lines changed, < 10 files)

---

## Self-Review

### Code Quality

- [ ] No leftover debug statements (`console.log`, `print`, `debugger`)
- [ ] No commented-out code that should be removed
- [ ] No TODO/FIXME comments that should be resolved in this PR
- [ ] No hardcoded values that should be configuration
- [ ] No dead code or unused imports
- [ ] Variable and function names are clear and descriptive

### Correctness

- [ ] Code does what the PR description says
- [ ] Edge cases handled (empty inputs, null values, boundaries)
- [ ] Error cases handled gracefully
- [ ] No off-by-one errors

### Security

- [ ] No secrets, tokens, API keys, or passwords in the diff
- [ ] No new vulnerable dependencies
- [ ] Input validation present where needed
- [ ] No injection vulnerabilities (SQL, XSS, command)

### Tests

- [ ] New code has corresponding tests
- [ ] Edge cases covered in tests
- [ ] All existing tests still pass

---

## PR Description

- [ ] Title follows `<type>[scope]: <description>` format
- [ ] Title under 72 characters
- [ ] Description explains what and why
- [ ] Changes listed with types and affected files
- [ ] Related issues linked (`Closes #123`, `Refs #456`)
- [ ] Review guidance provided for multi-file PRs
- [ ] Description under 4000 characters

---

## Submission

- [ ] PR targets the correct branch
- [ ] At least one reviewer assigned
- [ ] Labels reflect current status
- [ ] Draft status correct (draft if WIP, ready if complete)
- [ ] CI checks passing

---

## Standards Compliance

- [ ] Commits follow [Commit Messages](../../git-workflow/standards/commit-messages.md) standard
- [ ] Branch follows [Branch Naming](../../git-workflow/standards/branch-naming.md) standard
- [ ] PR follows [PR Descriptions](pr-descriptions.md) standard
- [ ] PR size follows [PR Size](pr-size.md) standard

---

## Review Response

- [ ] All review comments have a response
- [ ] Fixes pushed as new commits (not force-pushed during review)
- [ ] Resolved conversations marked as resolved
- [ ] Review re-requested from original reviewers
