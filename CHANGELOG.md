# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased](https://github.com/willvelida/code-minions/compare/v0.3.1...HEAD)

### Features

- remove language standards in favour of packages

### Bug Fixes

- correct comment in check-coverage.sh script
- use cmd /C for Windows clean target in Makefile
- use Git Bash as Make shell on Windows
- align help text with target comments
- handle errcheck lint for os.Chmod in cleanup

### Refactor

- simplify Makefile by removing Windows shell detection
- assert all valid assistants via assistant.List()

### Documentation

- add SECURITY.md with vulnerability reporting process

### Styling

- format Go source files

### Testing

- add command-level tests for list command (#39)
- cover installer error branches (path traversal, write failures)
- harden write-failure test per review feedback
- add edge case tests for flag parsing

### CI

- add test coverage threshold enforcement
- add Makefile for local dev workflow
- add multi-OS test matrix to release workflow
- scope permissions to least privilege in release workflow
- add GitHub issue and PR templates
- align PR template checklist with CONTRIBUTING.md

### Miscellaneous

- add Dependabot for Go module and GitHub Actions updates
- add patterns wildcard to Dependabot groups
## [0.3.1](https://github.com/willvelida/code-minions/compare/v0.3.0...v0.3.1) — 2026-02-13

### Bug Fixes

- use minor version for golangci-lint (v2.9)
- use ldflags-injected Version for release binaries

### CI

- add golangci-lint to CI pipeline
## [0.3.0](https://github.com/willvelida/code-minions/compare/v0.2.0...v0.3.0) — 2026-02-13

### Features

- add --for flag to install for specific coding assistants (#17)
- add uninstall CLI command (#15)
- wire OnInstall to create AGENTS.md during package install
- add update command to CLI (#21)
- expand installation options for Windows (#29)

### Bug Fixes

- stop installing standards.index.md to user repositories
- update only detects installed packages when no flags given
- address second round of PR review feedback
- intersect installed standards with embedded set
- rename deprecated archives.format to formats for GoReleaser v2
- address PR review comments on install scripts
- address second round of PR review comments
- guard ProgressPreference restore and require both auth env vars

### Refactor

- address PR [#26](https://github.com/willvelida/code-minions/issues/26) review feedback
- scan-then-intersect detection, extract helper, improve tests

### Documentation

- add uninstall usage and flags to README
## [0.2.0](https://github.com/willvelida/code-minions/compare/v0.1.1...v0.2.0) — 2026-02-12

### Features

- add package-based installation with --package flag
- update list command to show packages

### Bug Fixes

- skip root directory when stripping prefix

### Refactor

- reorganise agents and skills into packages

### Documentation

- update README for package-based installation

### Miscellaneous

- remove docs planning folder
## [0.1.1](https://github.com/willvelida/code-minions/compare/v0.1.0...v0.1.1) — 2026-02-12

### Bug Fixes

- use runtime/debug to read embedded module version

### Testing

- add unit tests for getVersion with injectable build info
## [0.1.0](https://github.com/willvelida/code-minions/compare/...v0.1.0) — 2026-02-12

### Features

- add git workflow skill with actions and standards
- add raise-pull-requests skill
- add creating-devcontainers skill and language standards
- add creating-agent-skills skill
- add creating-documentation skill and update README
- add developer-mentor skill
- rename skill and add clarifying questions
- add STRIDE-based threat modelling skill
- add creating-agents skill
- add agent-skill-expert agent definition
- add developer-mentor agent and rename to .expert.md convention
- add devcontainer agent and reorganise agent file structure
- add git-workflow agent and move .github agents back to agents/
- add threat-modelling agent
- add code-minions CLI installer tool

### Bug Fixes

- address PR review comments on creating-devcontainers skill
- address round 2 PR review comments
- address PR review feedback on developer-mentor
- escape backticks in tables and add missing checklist item
- address PR review feedback
- address second round of PR review feedback
- address third round of PR review feedback
- harden installer with path validation and error handling
- address fifth round of PR review feedback

### Refactor

- rename to gerund form for naming consistency
- standardise action file structure
- replace anti-code protocol with level-adaptive code policy
- rename agent files from expert to agent

### Documentation

- add pull request description template
- improve consistency and add missing elements
- remove markdown fences from checklist files
- require cryptographic commit signatures
- update ASCII art banner to CODE MINIONS text
- add developer-mentor to skills listing
- add AGENTS.md with system routing and structure overview
- add CLI installation and usage instructions

### Styling

- add language tags to code blocks

### CI

- add test workflow for pull requests
<!-- generated by git-cliff -->
