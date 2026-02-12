---
name: raise-pull-requests
description: 'Raise high-quality pull requests that are easy to review, provide clear context, and follow team conventions. Use when a user asks to open, write, prepare, or submit a pull request. Covers PR sizing, self-review, writing descriptions, submitting PRs, responding to review feedback, and reviewing others'' PRs. Follows GitHub best practices for collaboration and code review.'
license: MIT
compatibility: 'GitHub, GitHub CLI (gh) for PR commands'
allowed-tools: Bash
---

# Raise Pull Requests

## Overview

A skill for raising pull requests that are easy to review, clearly described, and follow team conventions. Good pull requests reduce review time, catch bugs earlier, and keep the team informed about what is changing and why.

This skill covers the full lifecycle of a pull request — from preparing your changes through responding to review feedback.

## Principles

### 1. Write Small, Focused Pull Requests

Each PR should fulfill a single purpose. Small PRs are easier and faster to review, leave less room for bugs, and provide a clearer history of changes. If your change touches multiple concerns, split it into separate PRs.

### 2. Provide Context and Guidance

Reviewers should not have to guess what your PR does or why. Write clear titles and descriptions, link related issues, and guide reviewers through the changes — especially when the PR touches multiple files.

### 3. Review Your Own Work First

Before requesting review, read through your own diff. Build and test locally. Catch typos, leftover debug code, and missing edge cases before someone else does.

### 4. Keep Your Team Informed

Use labels, linked issues, and project boards to make progress visible. A well-structured PR doubles as a status update — reducing the need for separate check-ins.

### 5. Respond Constructively to Feedback

Address every review comment. Push fixes as new commits during review so reviewers can track incremental changes. Re-request review only after all feedback is addressed.

## Actions

Step-by-step procedures for each phase of raising a pull request:

1. [Prepare Pull Request](actions/prepare-pull-request.md) — Scope, self-review, and readiness checks before opening
2. [Write PR Description](actions/write-pr-description.md) — Craft a clear title, description, and reviewer guidance
3. [Submit Pull Request](actions/submit-pull-request.md) — Open the PR, assign reviewers, and add labels
4. [Respond to Review](actions/respond-to-review.md) — Address feedback, push fixes, and re-request review
5. [Review Pull Request](actions/review-pull-request.md) — Review someone else's PR with constructive feedback

## Standards

Conventions and guidelines to follow when raising pull requests:

- [Pull Request Template](standards/pull-request-template.md) — Standard PR description template
- [PR Size](standards/pr-size.md) — Guidelines for keeping PRs small and focused
- [Self-Review Checklist](standards/self-review-checklist.md) — Pre-submission checklist for authors
- [PR Descriptions](standards/pr-descriptions.md) — Standards for titles, descriptions, and linking issues
- [Checklist](standards/checklist.md) — Consolidated compliance checklist

## Related Skills

This skill focuses on the **pull request lifecycle** and complements the broader [git-workflow](../git-workflow/SKILL.md) skill:

| Skill | Focus |
|-------|-------|
| `raise-pull-requests` | PR preparation, descriptions, submission, responding to reviews |
| `git-workflow` | Full git workflow including branching, commits, merging, releases |

For git operations like creating branches, committing changes, or merging PRs, see [git-workflow](../git-workflow/SKILL.md).

## References

- [GitHub — Helping others review your changes](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/getting-started/helping-others-review-your-changes)
- [Atlassian — The written and unwritten guide to pull requests](https://www.atlassian.com/blog/git/written-unwritten-guide-pull-requests)
