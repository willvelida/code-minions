# Bash Core Standard

## Overview

Standards for Bash/Shell script development including script structure, naming conventions, and coding style.

## Principles

1. **Portability** — Scripts should work across local DevContainers, GitHub Codespaces, and CI/CD pipelines
2. **Safety** — Use strict mode and error handling
3. **Readability** — Clear structure and documentation
4. **Security** — Avoid common security pitfalls

---

## Script Standards

### Shebang Requirements

All scripts **MUST** start with:

```bash
#!/usr/bin/env bash
```

This format ensures portability across different systems where Bash may be installed in different locations.

### Strict Mode (Mandatory)

All scripts **MUST** include strict mode settings:

```bash
#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'
```

| Flag | Purpose |
|------|---------|
| `-e` | Exit immediately on error |
| `-u` | Treat unset variables as errors |
| `-o pipefail` | Return exit code of first failed command in pipeline |
| `IFS=$'\n\t'` | Safer word splitting |

### Error Handling

```bash
# Trap for cleanup on exit
cleanup() {
    echo "Cleaning up..."
}
trap cleanup EXIT

# Error handling function
error_exit() {
    echo "ERROR: $1" >&2
    exit "${2:-1}"
}
```

---

## File Headers

```bash
#!/usr/bin/env bash
#
# Script Name: script_name.sh
# Description: Brief description of what the script does
#
# Usage: ./script_name.sh [options] <arguments>
#
# Dependencies:
#   - dependency1
#   - dependency2
#

set -euo pipefail
```

**Note**: Author, date, and version information should be managed via Git history and release tags.

---

## Variable Standards

```bash
# Use lowercase for local variables
local my_variable="value"

# Use UPPERCASE for exported/environment variables
export MY_ENV_VAR="value"

# Always quote variables
echo "${my_variable}"

# Use default values
: "${MY_VAR:=default_value}"
```

---

## Function Standards

```bash
# Use snake_case for function names
my_function() {
    local param1="$1"
    local param2="${2:-default}"
    
    # Function body
}

# Document complex functions
# Description: Brief description of what the function does
# Arguments:
#   $1 - Description of first argument
#   $2 - Description of second argument (optional)
# Returns:
#   0 on success, non-zero on failure
```

---

## Naming Conventions

| Type | Convention | Example |
|------|------------|---------|
| Local variables | snake_case | `my_variable`, `file_path` |
| Environment variables | UPPER_SNAKE_CASE | `MY_ENV_VAR`, `API_KEY` |
| Functions | snake_case | `process_file`, `validate_input` |
| Scripts | kebab-case or snake_case | `deploy-app.sh`, `post_create.sh` |
| Constants | UPPER_SNAKE_CASE | `MAX_RETRIES`, `DEFAULT_TIMEOUT` |
