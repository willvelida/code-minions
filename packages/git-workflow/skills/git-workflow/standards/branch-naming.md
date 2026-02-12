# Branch Naming Standard

Consistent branch naming makes it easy to identify the purpose of a branch at a glance and enables automation around CI/CD, code review, and cleanup.

## Format

```
<prefix>/<short-description>
```

- **Prefix**: Indicates the type of work
- **Short description**: Kebab-case summary of the change (lowercase, hyphens between words)

## Prefixes

| Prefix | Purpose | Example |
|---|---|---|
| `feature/` | New feature or enhancement | `feature/user-authentication` |
| `bugfix/` | Bug fix (non-urgent) | `bugfix/fix-login-redirect` |
| `hotfix/` | Urgent production fix | `hotfix/1.2.1` or `hotfix/patch-null-check` |
| `release/` | Release preparation | `release/1.2.0` |
| `chore/` | Maintenance, refactoring, tooling | `chore/update-dependencies` |
| `docs/` | Documentation changes | `docs/update-api-reference` |
| `test/` | Adding or updating tests | `test/add-auth-unit-tests` |
| `ci/` | CI/CD pipeline changes | `ci/add-staging-deploy` |

## Rules

1. **Always use a prefix** — bare branch names like `login-fix` are not allowed
2. **Use kebab-case** — lowercase letters, numbers, and hyphens only
3. **Keep it short** — aim for 3–5 words in the description
4. **No special characters** — no spaces, underscores, or dots
5. **Include ticket number when available** — e.g., `feature/PROJ-123-user-auth`

## Protected Branches

The following branches must never be worked on directly:

| Branch | Purpose |
|---|---|
| `main` | Production-ready code |
| `develop` | Integration branch (Gitflow) |

All changes to protected branches flow through pull requests.

## Examples

```
feature/add-payment-gateway
feature/PROJ-456-search-filters
bugfix/fix-date-parsing
bugfix/PROJ-789-null-pointer
hotfix/1.3.1
release/2.0.0
chore/upgrade-node-20
docs/add-contributing-guide
test/integration-tests-auth
ci/add-docker-build-step
```

## Cleanup

Delete branches after they are merged:

```bash
# Delete remote branch
git push origin --delete <branch-name>

# Delete local branch
git branch -d <branch-name>

# Prune stale remote-tracking references
git fetch --prune
```
