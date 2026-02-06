# Create Release

Create a release branch, finalize it, and tag the release. This action applies to the **Gitflow Workflow**.

## Prerequisites

- `develop` branch contains all features intended for this release
- All features are complete and tested
- Repository uses the Gitflow branching model

## Steps

### 1. Sync the `develop` Branch

```bash
git checkout develop
git pull origin develop
```

### 2. Create the Release Branch

```bash
git checkout -b release/<version> develop
```

Follow semantic versioning: `release/1.2.0`, `release/2.0.0`.

### 3. Push the Release Branch

```bash
git push -u origin release/<version>
```

### 4. Perform Release Preparation

On the release branch, only the following changes are allowed:

- Version number bumps (package.json, setup.py, etc.)
- Changelog updates
- Minor bug fixes found during release testing
- Documentation updates for the release

**No new features.** New features go to `develop` for the next release.

```bash
# Example: bump version
git add package.json CHANGELOG.md
git commit -s -m "chore(release): bump version to <version>"
```

### 5. Merge into `main`

```bash
git checkout main
git pull origin main
git merge --no-ff release/<version>
git push origin main
```

### 6. Tag the Release

```bash
git tag -a v<version> -m "Release v<version>"
git push origin v<version>
```

### 7. Back-Merge into `develop`

Ensure any release-branch fixes are included in ongoing development:

```bash
git checkout develop
git pull origin develop
git merge --no-ff release/<version>
git push origin develop
```

Resolve any conflicts if they arise (see [Resolve Merge Conflicts](resolve-merge-conflicts.md)).

### 8. Delete the Release Branch

```bash
git branch -d release/<version>
git push origin --delete release/<version>
```

## Release Checklist

- [ ] All intended features are merged into `develop`
- [ ] Version numbers are bumped
- [ ] Changelog is updated
- [ ] Release branch is merged into `main`
- [ ] Release is tagged with `v<version>`
- [ ] Release branch is back-merged into `develop`
- [ ] Release branch is deleted
- [ ] Release notes are published on the hosting platform
