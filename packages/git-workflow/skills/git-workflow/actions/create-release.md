# Create Release

## Purpose

Create a release branch, finalize it, and tag the release. This action applies to the **Gitflow Workflow**.

---

## Flow

### Step 1: Verify Release Readiness 🛑

Confirm:

- `develop` branch contains all features intended for this release
- All features are complete and tested
- Repository uses the Gitflow branching model

**🛑 STOP**: Confirm with the user that all intended features are merged into `develop` and the version number for this release.

**Success Criteria:**
- [ ] All features merged into `develop`
- [ ] Version number confirmed

---

### Step 2: Sync the `develop` Branch

```bash
git checkout develop
git pull origin develop
```

**Success Criteria:**
- [ ] On `develop` with latest changes

---

### Step 3: Create the Release Branch

```bash
git checkout -b release/<version> develop
```

Follow semantic versioning: `release/1.2.0`, `release/2.0.0`.

```bash
git push -u origin release/<version>
```

**Success Criteria:**
- [ ] Release branch created from `develop`
- [ ] Branch pushed to remote

---

### Step 4: Perform Release Preparation

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

**Success Criteria:**
- [ ] Version numbers bumped
- [ ] Changelog updated
- [ ] No new features on the release branch

---

### Step 5: Merge into `main` 🛑

**🛑 STOP**: Confirm with the user that the release is ready to merge into production.

```bash
git checkout main
git pull origin main
git merge --no-ff release/<version>
git push origin main
```

**Success Criteria:**
- [ ] Release merged into `main`

---

### Step 6: Tag the Release

```bash
git tag -a v<version> -m "Release v<version>"
git push origin v<version>
```

**Success Criteria:**
- [ ] Release tagged with `v<version>`
- [ ] Tag pushed to remote

---

### Step 7: Back-Merge into `develop`

Ensure any release-branch fixes are included in ongoing development:

```bash
git checkout develop
git pull origin develop
git merge --no-ff release/<version>
git push origin develop
```

Resolve any conflicts if they arise (see [Resolve Merge Conflicts](resolve-merge-conflicts.md)).

**Success Criteria:**
- [ ] Release back-merged into `develop`

---

### Step 8: Delete the Release Branch

```bash
git branch -d release/<version>
git push origin --delete release/<version>
```

**Success Criteria:**
- [ ] Release branch deleted (local and remote)

---

## Release Checklist

- [ ] All intended features are merged into `develop`
- [ ] Version numbers are bumped
- [ ] Changelog is updated
- [ ] Release branch is merged into `main`
- [ ] Release is tagged with `v<version>`
- [ ] Release branch is back-merged into `develop`
- [ ] Release branch is deleted
- [ ] Release notes are published on the hosting platform
