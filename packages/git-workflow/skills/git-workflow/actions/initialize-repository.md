# Initialize Repository

## Purpose

Set up a local development environment by cloning or forking a repository and configuring remotes.

---

## Flow

### Step 1: Choose Setup Method 🛑

Determine which setup method to use based on the workflow model:

| Workflow | Method |
|---|---|
| Centralized | Clone the central repository |
| Feature Branch | Clone the central repository |
| Gitflow | Clone the central repository, create `develop` branch |
| Forking | Fork on server, then clone your fork |

**🛑 STOP**: Confirm the workflow model with the user before proceeding.

**Success Criteria:**
- [ ] Workflow model confirmed by user

---

### Step 2: Clone or Fork

#### Centralized / Feature Branch Setup

```bash
git clone <repository-url>
cd <repository-name>
```

#### Gitflow Setup

```bash
git clone <repository-url>
cd <repository-name>

# Create develop branch if it does not exist
git checkout -b develop
git push -u origin develop

# Switch to develop as working base
git checkout develop
```

#### Forking Workflow Setup

1. Fork the upstream repository on the hosting platform (GitHub, Bitbucket, GitLab)
2. Clone your fork:

```bash
git clone <your-fork-url>
cd <repository-name>

# Add upstream remote
git remote add upstream <original-repository-url>
```

**Success Criteria:**
- [ ] Repository cloned locally
- [ ] Correct base branch checked out
- [ ] `develop` branch exists (Gitflow only)
- [ ] `upstream` remote configured (Forking only)

---

### Step 3: Verify Remotes

```bash
# Confirm remote configuration
git remote -v
```

Expected output:

```
# Standard workflows
origin  <repository-url> (fetch)
origin  <repository-url> (push)

# Forking workflow
origin    <your-fork-url> (fetch)
origin    <your-fork-url> (push)
upstream  <original-repository-url> (fetch)
upstream  <original-repository-url> (push)
```

**Success Criteria:**
- [ ] `origin` remote points to the correct repository
- [ ] `upstream` remote configured (Forking only)

---

### Step 4: Post-Setup Verification

```bash
# Confirm current branch
git branch

# Pull latest changes
git pull
```

**Success Criteria:**
- [ ] On the correct branch (`main` or `develop`)
- [ ] Latest changes pulled
- [ ] Repository ready for development
