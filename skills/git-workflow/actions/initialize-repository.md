# Initialize Repository

Set up a local development environment by cloning or forking a repository and configuring remotes.

## Choose Setup Method

Determine which setup method to use based on the workflow model:

| Workflow | Method |
|---|---|
| Centralized | Clone the central repository |
| Feature Branch | Clone the central repository |
| Gitflow | Clone the central repository, create `develop` branch |
| Forking | Fork on server, then clone your fork |

## Centralized / Feature Branch Setup

### 1. Clone the Repository

```bash
git clone <repository-url>
cd <repository-name>
```

### 2. Verify Remote

```bash
git remote -v
# origin  <repository-url> (fetch)
# origin  <repository-url> (push)
```

## Gitflow Setup

### 1. Clone the Repository

```bash
git clone <repository-url>
cd <repository-name>
```

### 2. Create the `develop` Branch (if it does not exist)

```bash
git checkout -b develop
git push -u origin develop
```

### 3. Switch to `develop` as Working Base

```bash
git checkout develop
```

## Forking Workflow Setup

### 1. Fork the Repository

Fork the upstream repository on the hosting platform (GitHub, Bitbucket, GitLab).

### 2. Clone Your Fork

```bash
git clone <your-fork-url>
cd <repository-name>
```

### 3. Add Upstream Remote

```bash
git remote add upstream <original-repository-url>
```

### 4. Verify Remotes

```bash
git remote -v
# origin    <your-fork-url> (fetch)
# origin    <your-fork-url> (push)
# upstream  <original-repository-url> (fetch)
# upstream  <original-repository-url> (push)
```

## Post-Setup Verification

After any setup method, verify the environment:

```bash
# Confirm current branch
git branch

# Confirm remote configuration
git remote -v

# Pull latest changes
git pull
```
