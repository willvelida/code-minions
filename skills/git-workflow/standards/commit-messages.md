# Commit Messages Standard

All commits follow the [Conventional Commits](https://www.conventionalcommits.org/) specification for consistent, machine-readable commit history.

## Format

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

## Types

| Type | Purpose |
|---|---|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation only |
| `style` | Formatting/style changes (no logic change) |
| `refactor` | Code refactor (no feature or fix) |
| `perf` | Performance improvement |
| `test` | Add or update tests |
| `build` | Build system or dependency changes |
| `ci` | CI/CD configuration changes |
| `chore` | Maintenance or miscellaneous tasks |
| `revert` | Revert a previous commit |

## Scope

The scope is optional and indicates the module or area affected:

```
feat(auth): add JWT validation
fix(api): handle null response
docs(readme): update setup instructions
refactor(db): extract connection pooling
```

## Description Rules

- Use present tense: "add" not "added"
- Use imperative mood: "fix bug" not "fixes bug"
- Do not capitalize the first letter
- Do not end with a period
- Keep under 72 characters

## Body

The body is optional. Use it to explain **what** and **why** (not how):

```
fix(auth): correct token expiration check

The previous implementation compared timestamps using local time
instead of UTC, causing tokens to expire early for users in
negative UTC offsets.
```

## Footer

Use footers for metadata:

```
# Reference an issue
Refs #456

# Close an issue
Closes #123

# Breaking change
BREAKING CHANGE: `authenticate()` now requires an options object
```

## Breaking Changes

Indicate breaking changes with either:

### Exclamation mark in the type

```
feat!: remove deprecated /v1/users endpoint
```

### BREAKING CHANGE footer

```
feat(api): change authentication flow

BREAKING CHANGE: `login()` now returns a Promise instead of a callback
```

## Signed Commits

All commits **must** be signed in two ways:

### 1. Cryptographic Signature (`-S`)

Sign commits cryptographically using GPG or SSH keys. This verifies the commit author's identity.

```bash
# Configure SSH signing (recommended)
git config --global gpg.format ssh
git config --global user.signingkey ~/.ssh/id_rsa.pub
git config --global commit.gpgsign true

# Or configure GPG signing
git config --global commit.gpgsign true
git config --global user.signingkey YOUR_GPG_KEY_ID
```

With `commit.gpgsign = true`, all commits are automatically signed.

### 2. Sign-off (`-s`)

Add a `Signed-off-by:` line for Developer Certificate of Origin (DCO) compliance:

```bash
git commit -s -m "feat(auth): add token refresh endpoint"
```

### Combined Usage

```bash
# Both flags together (if gpgsign not set globally)
git commit -s -S -m "feat(auth): add token refresh endpoint"

# With global gpgsign=true, just use -s
git commit -s -m "feat(auth): add token refresh endpoint"
```

## Examples

```
feat(auth): add JWT token validation middleware
fix(api): correct null check in user lookup
docs(readme): update installation instructions
style(lint): apply prettier formatting
refactor(db): extract connection pool into module
perf(query): add index for user email lookup
test(auth): add unit tests for login endpoint
build(deps): upgrade express to v5
ci(actions): add staging deployment workflow
chore(scripts): clean up unused migration files
revert: revert "feat(api): add caching layer"
```
