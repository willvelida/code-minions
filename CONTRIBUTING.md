# Contributing to code-minions

Thank you for your interest in contributing! This guide will help you get set up and understand the workflow.

## Prerequisites

- **Go 1.25+** — see [go.dev/dl](https://go.dev/dl/) for installation
- **Git** — any recent version
- **GNU Make** (optional, for local workflow shortcuts) — on Linux typically pre-installed; on macOS install via Homebrew (`brew install make`) and use `gmake`; on Windows install via `choco install make` and run from a **Git Bash** terminal
- **golangci-lint** (optional, for local linting) — see [golangci-lint.run](https://golangci-lint.run/welcome/install/)

## Getting Started

1. **Fork and clone** the repository:

   ```bash
   git clone https://github.com/<your-username>/code-minions.git
   cd code-minions
   ```

2. **Build** to verify your setup:

   ```bash
   go build ./cmd/code-minions
   ```

3. **Run tests**:

   ```bash
   go test ./...
   ```

4. **Set up the pre-push hook** (recommended):

   ```bash
   git config core.hooksPath .githooks
   ```

   This runs a coverage threshold check before each push. You can bypass it in a pinch with `git push --no-verify`, but CI will still enforce the threshold.

## Using Make

The repository includes a `Makefile` with shortcuts for common development tasks. Run `make help` to see all available targets:

| Target | Description |
|--------|-------------|
| `make build` | Compile the binary (default target) |
| `make test` | Run all tests with coverage summary |
| `make lint` | Run golangci-lint |
| `make fmt` | Format all Go source files |
| `make snapshot` | Run GoReleaser in snapshot mode |
| `make install` | Install the binary to `$GOPATH/bin` |
| `make coverage` | Generate coverage profile and open HTML report |
| `make check-coverage` | Run the coverage threshold check |
| `make clean` | Remove build artifacts and coverage files |

Make is **optional** — all commands work standalone too. The sections below show both forms.

## Making Changes

### Branch Naming

Create a branch from `main` using the format `<type>/<short-description>`:

```
ci/coverage-threshold
docs/contributing-guide
feat/new-install-flag
fix/uninstall-empty-dir
```

Common types: `feat`, `fix`, `docs`, `ci`, `refactor`, `test`, `chore`.

### Commit Messages

Use [Conventional Commits](https://www.conventionalcommits.org/) format:

```
<type>: <short summary>
```

Examples:

```
feat: add --for flag to install command
fix: handle missing AGENTS.md gracefully
docs: add CONTRIBUTING.md
ci: add coverage threshold enforcement
test: add uninstall edge case tests
```

Keep the summary concise (≤ 72 characters). Use the commit body for additional context when needed.

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage summary
go test ./... -cover

# Generate a coverage profile and inspect it
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out     # per-function breakdown
go tool cover -html=coverage.out     # visual HTML report
```

### Coverage Expectations

- **Minimum threshold: 70%** — CI will fail if total coverage drops below this
- The threshold is configured in [.testcoverage.yml](.testcoverage.yml)
- `cmd/code-minions/main.go` is excluded from coverage (bootstrap code with no testable logic)
- If you add new packages, make sure to include tests — don't let coverage regress

The pre-push hook (see [Getting Started](#getting-started)) runs the same coverage check locally so you catch issues before pushing.

### Coverage Check Script

You can run the threshold check manually at any time:

```bash
./scripts/check-coverage.sh       # uses default 70% threshold
./scripts/check-coverage.sh 75    # override threshold
```

## Code Style

### Formatting

Go code must be formatted with `gofmt`. Most editors do this automatically on save.

```bash
gofmt -w .
```

### Linting

CI runs [golangci-lint](https://golangci-lint.run/) on all PRs. To run it locally:

```bash
golangci-lint run
```

Only new issues are flagged on PRs — you won't be asked to fix pre-existing lint warnings in unrelated code.

## Submitting a Pull Request

1. **Open an issue first** — discuss the change before investing time in implementation
2. **Create a branch** from `main` following the [branch naming convention](#branch-naming)
3. **Make your changes** with clear, well-scoped commits
4. **Ensure all checks pass locally** (or use `make build`, `make test`, `make check-coverage`):
   - `go build ./cmd/code-minions` — builds successfully
   - `go test ./... -cover` — tests pass
   - `./scripts/check-coverage.sh` — coverage meets the threshold
5. **Push and open a PR** against `main`

### PR Checklist

Before requesting review, verify:

- [ ] Tests pass (`go test ./...`)
- [ ] Coverage meets the 70% threshold
- [ ] Code is formatted (`gofmt`)
- [ ] Commit messages follow Conventional Commits format
- [ ] New functionality includes tests
- [ ] PR description explains *what* and *why*

### What Reviewers Look For

- **Correctness** — does the change do what it claims?
- **Tests** — are new code paths covered?
- **Clarity** — is the code easy to follow?
- **Scope** — is the PR focused on one concern?

## CI Pipeline

Every push and PR triggers the CI pipeline, which runs:

| Job | What it does |
|-----|-------------|
| `test` | Builds and runs tests on Ubuntu, macOS, and Windows |
| `coverage` | Enforces 70% coverage threshold and uploads reports to [Codecov](https://codecov.io/gh/willvelida/code-minions) |
| `lint-go` | Runs golangci-lint |
| `lint-scripts` | Lints shell and PowerShell scripts (ShellCheck, PSScriptAnalyzer) |
| `goreleaser-check` | Validates release configuration |

All jobs must pass before a PR can be merged.

## Questions?

Open an issue — happy to help!
