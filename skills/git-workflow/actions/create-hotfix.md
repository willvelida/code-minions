# Create Hotfix

Create an urgent fix that is applied directly to production. Hotfixes branch from `main` and merge back into both `main` and `develop`. This action applies to the **Gitflow Workflow**.

## Prerequisites

- A critical bug exists in the current production release (`main`)
- The fix cannot wait for the next scheduled release
- Repository uses the Gitflow branching model

## Steps

### 1. Sync the `main` Branch

```bash
git checkout main
git pull origin main
```

### 2. Create the Hotfix Branch

```bash
git checkout -b hotfix/<version> main
```

Increment the patch version: if current release is `v1.2.0`, the hotfix is `hotfix/1.2.1`.

### 3. Apply the Fix

Make the minimum changes needed to resolve the issue:

```bash
# Make changes
git add <files>
git commit -s -m "fix(<scope>): <description of the fix>

Closes #<issue-number>"
```

### 4. Bump the Version

```bash
# Update version in package.json, setup.py, etc.
git add package.json
git commit -s -m "chore(release): bump version to <version>"
```

### 5. Push the Hotfix Branch

```bash
git push -u origin hotfix/<version>
```

### 6. Open a Pull Request (optional but recommended)

Even for urgent fixes, a quick code review catches mistakes:

```bash
gh pr create \
  --base main \
  --title "fix(<scope>): <description>" \
  --body "## Hotfix

<Description of the critical issue and fix>

Closes #<issue-number>"
```

### 7. Merge into `main`

```bash
git checkout main
git pull origin main
git merge --no-ff hotfix/<version>
git push origin main
```

### 8. Tag the Hotfix Release

```bash
git tag -a v<version> -m "Hotfix v<version>"
git push origin v<version>
```

### 9. Back-Merge into `develop`

```bash
git checkout develop
git pull origin develop
git merge --no-ff hotfix/<version>
git push origin develop
```

If a release branch currently exists, merge the hotfix into the release branch instead of `develop`:

```bash
git checkout release/<release-version>
git merge --no-ff hotfix/<version>
git push origin release/<release-version>
```

### 10. Delete the Hotfix Branch

```bash
git branch -d hotfix/<version>
git push origin --delete hotfix/<version>
```

## Hotfix Checklist

- [ ] Hotfix branch created from `main`
- [ ] Minimum fix applied and tested
- [ ] Version number bumped
- [ ] Merged into `main` and tagged
- [ ] Back-merged into `develop` (or active release branch)
- [ ] Hotfix branch deleted
- [ ] Stakeholders notified of the fix
