---
description: Security coding standards for application source code
applyTo: "**/*.go,**/*.py,**/*.ts,**/*.js,**/*.java,**/*.cs"
---

# Security Coding Standards

When creating or editing application source code:

- Validate and sanitise all external inputs (user input, API responses, file contents)
- Use parameterised queries — never concatenate user input into SQL or commands
- Store secrets in environment variables or secret managers, not in source code
- Apply the principle of least privilege for file system and network access
- Use constant-time comparison for security-sensitive string comparisons
- Log security-relevant events (authentication, authorisation, data access) without including sensitive data
- Handle errors explicitly — never swallow errors in security-sensitive code paths
- Use established cryptographic libraries — never implement custom cryptography
- Set appropriate timeouts for HTTP clients and database connections
- Prefer allow-lists over deny-lists for input validation
