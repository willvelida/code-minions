# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [0.4.1](https://github.com/willvelida/code-minions/compare/v0.4.0...v0.4.1) — 2026-02-21

### Features

- add team install command ([#147](https://github.com/willvelida/code-minions/issues/147))


## [0.4.0](https://github.com/willvelida/code-minions/compare/v0.3.1...v0.4.0) — 2026-02-20

### Features

- remove language standards in favour of packages
- add shell completion command
- add --json output flag for scriptability
- add --verbose and --quiet flags
- add confirmation prompt before destructive uninstall
- add Long descriptions and Examples to install, uninstall, and list commands
- add Dev Container configuration
- add data model types for Package, Persona, Team, Source
- add registry interface and embedded source implementation
- add package.yaml manifests for all embedded packages
- add show and search commands
- wire install tracking manifest
- support MCP server definitions in packages
- pairwise MCP config translation between coding assistants
- add persona resolver
- add persona and team types to install manifest
- export MCP server installation for persona merging
- add persona installer with assistant grouping generators
- wire persona install, uninstall, update, and list into CLI
- add team configuration layer support
- add NewReversePathMapper to assistant.Config ([#63](https://github.com/willvelida/code-minions/issues/63))
- add Transfer engine for migrating files between assistants ([#63](https://github.com/willvelida/code-minions/issues/63))
- add transfer CLI command (Phase 3)
- add Cursor as a supported coding assistant ([#88](https://github.com/willvelida/code-minions/issues/88))
- add Gemini as a supported coding assistant ([#90](https://github.com/willvelida/code-minions/issues/90))
- add Codex CLI as supported coding assistant ([#89](https://github.com/willvelida/code-minions/issues/89))
- generate CLAUDE.md for Claude Code integration
- add code-minions init command ([#96](https://github.com/willvelida/code-minions/issues/96))
- wire project manifest into install, uninstall, and update commands
- external package sources (scenarios 1-3)
- add instructions primitive (.instructions.md) with file-pattern targeting ([#92](https://github.com/willvelida/code-minions/issues/92))
- add prompt primitive (.prompt.md) for reusable slash commands
- fetch remote packages from GitHub repositories ([#97](https://github.com/willvelida/code-minions/issues/97))
- add init templates for common project archetypes ([#110](https://github.com/willvelida/code-minions/issues/110))
- add team init wizard ([#141](https://github.com/willvelida/code-minions/issues/141))

### Bug Fixes

- correct comment in check-coverage.sh script
- use cmd /C for Windows clean target in Makefile
- use Git Bash as Make shell on Windows
- align help text with target comments
- handle errcheck lint for os.Chmod in cleanup
- address Copilot review feedback
- resolve errcheck lint errors and dry-run JSON corruption
- rewrite pre-commit hook to pass ShellCheck
- address review feedback on JSON consistency, README wording, and Bash compat
- handle errcheck lint and add codecov threshold
- add blank identifiers for errcheck on quiet error paths
- handle w.Close return values to satisfy errcheck lint
- address PR review comments
- address second round of PR review comments
- use precise .claude/agents/ path in install Long description
- address PR review comments on Long descriptions
- scope format-on-save to Go files only
- address lint errors and review comments
- guard nil manifest and deduplicate search results
- address review round 3 — yaml tags, error handling, ErrNotFound sentinel
- address review round 4 — ErrNotFound consistency, opencode paths, error surfacing
- address PR review comments and improve test coverage
- address second round of PR review comments
- address third round of PR review comments
- filter empty targets in --to flag and remove unused test helper
- address PR review — dedup targets, clarify warnings, rename sortedMapKeys
- address second round of PR review comments
- address PR review comments and lint failures
- address second round of PR review comments
- use path.Join for relative paths in persona generators
- address second-round review comments
- address third-round review comments
- address PR review comments
- address PR review comments and lint errors
- address PR review — regenerate AGENTS.md, DRY skip-files, agent-only cleanup
- align Gemini MCP transport and persona format with CLI conventions
- address PR review comments for Codex CLI ([#129](https://github.com/willvelida/code-minions/issues/129))
- scope CLAUDE.md generation to Claude only and add extractPackageNames tests
- address PR review feedback ([#134](https://github.com/willvelida/code-minions/issues/134))
- address second round of PR review feedback ([#134](https://github.com/willvelida/code-minions/issues/134))
- pass selected packages to install and clarify README default ([#134](https://github.com/willvelida/code-minions/issues/134))
- address PR review comments and lint-go failure
- address second round of PR review comments
- address third round of PR review comments
- clarify update comment to distinguish from install fallback
- address round-5 review comments
- address PR review - fix misleading comment and nil content optimization
- normalize spacing in applyTo-to-globs frontmatter conversion
- guard Cursor translator, scope manifest tracking, handle CRLF and uninstall errors
- add missing reverse mapping tests and remove dead nil check
- use forward-slash path joining and rename shadowed dir variables
- address PR review comments for prompt primitive
- use 0 as fallback defaultIdx in selectTemplate
- address PR review feedback on team init
- update test expectation for singular package grammar

### Refactor

- simplify Makefile by removing Windows shell detection
- assert all valid assistants via assistant.List()
- address PR review comments
- use registry in list command with --detail flag
- generate --for flag help text from assistant registry

### Documentation

- add SECURITY.md with vulnerability reporting process
- add automated CHANGELOG.md via git-cliff ([#43](https://github.com/willvelida/code-minions/issues/43))
- add GoDoc comments to exported types and functions ([#45](https://github.com/willvelida/code-minions/issues/45))
- add transfer command to README and CHANGELOG ([#63](https://github.com/willvelida/code-minions/issues/63))
- fix Claude example to show quoted description value
- update changelog for v0.4.0

### Styling

- format Go source files

### Testing

- add command-level tests for list command ([#39](https://github.com/willvelida/code-minions/issues/39))
- cover installer error branches (path traversal, write failures)
- harden write-failure test per review feedback
- add edge case tests for flag parsing
- add coverage tests for plain-text and --for paths
- add coverage tests for persona, grouping, and uninstall paths
- improve transfer command coverage

### CI

- add test coverage threshold enforcement
- add Makefile for local dev workflow
- add multi-OS test matrix to release workflow
- scope permissions to least privilege in release workflow
- add GitHub issue and PR templates
- align PR template checklist with CONTRIBUTING.md
- add changelog PR step to release workflow
- make codecov/patch informational, keep project as gate

### Miscellaneous

- add Dependabot for Go module and GitHub Actions updates
- add patterns wildcard to Dependabot groups
- remove EditorConfig extension (no .editorconfig file)


## [0.3.1](https://github.com/willvelida/code-minions/compare/v0.3.0...v0.3.1) — 2026-02-13

### Bug Fixes

- use minor version for golangci-lint (v2.9)
- use ldflags-injected Version for release binaries

### CI

- add golangci-lint to CI pipeline


## [0.3.0](https://github.com/willvelida/code-minions/compare/v0.2.0...v0.3.0) — 2026-02-13

### Features

- add --for flag to install for specific coding assistants ([#17](https://github.com/willvelida/code-minions/issues/17))
- add uninstall CLI command ([#15](https://github.com/willvelida/code-minions/issues/15))
- wire OnInstall to create AGENTS.md during package install
- add update command to CLI ([#21](https://github.com/willvelida/code-minions/issues/21))
- expand installation options for Windows ([#29](https://github.com/willvelida/code-minions/issues/29))

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


## [0.1.0](https://github.com/willvelida/code-minions/releases/tag/v0.1.0) — 2026-02-12

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
