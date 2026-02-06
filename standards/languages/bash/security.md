# Bash Security Standard

## Overview

Security requirements for Bash scripts.

---

## Prohibited Practices

| Practice | Reason | Alternative |
|----------|--------|-------------|
| `eval "$user_input"` | Command injection risk | Use arrays and proper quoting |
| Unquoted variables | Word splitting attacks | Always quote: `"${var}"` |
| `source` untrusted files | Code injection | Validate before sourcing |
| Hardcoded secrets | Credential exposure | Use environment variables |

---

## Secure Practices

### Input Validation

```bash
# Validate input format
if [[ ! "$input" =~ ^[a-zA-Z0-9_-]+$ ]]; then
    error_exit "Invalid input format"
fi
```

### Safe Command Execution

```bash
# Use arrays for commands with arguments
cmd=("ls" "-la" "${directory}")
"${cmd[@]}"
```

### Secure Temporary Files

```bash
# Create secure temp file
tmp_file=$(mktemp)
trap 'rm -f "$tmp_file"' EXIT
```

### Environment Variables for Secrets

```bash
# Check for required secrets
: "${API_KEY:?API_KEY environment variable is required}"
```

### Always Quote Variables

```bash
# ❌ Unsafe
echo $user_input

# ✅ Safe
echo "${user_input}"
```

---

## Compliance Checklist

- [ ] Script has appropriate shebang (`#!/usr/bin/env bash`)
- [ ] Strict mode enabled (`set -euo pipefail`)
- [ ] Variables are properly quoted
- [ ] Error handling implemented
- [ ] ShellCheck passes without errors
- [ ] No hardcoded secrets or credentials
- [ ] Input validation for user-provided data
