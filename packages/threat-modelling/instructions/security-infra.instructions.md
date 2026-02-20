---
description: Security standards for infrastructure-as-code files
applyTo: "**/*.tf,**/*.bicep,**/*.yaml,**/*.yml"
---

# Security Standards for Infrastructure Files

When creating or editing infrastructure-as-code files:

- Use managed identities instead of keys, passwords, or connection strings
- Enable encryption at rest for all storage and database resources
- Enable encryption in transit (TLS/HTTPS) for all network communication
- Apply the principle of least privilege for IAM roles and policies
- Never hardcode secrets — use key vault or secret manager references
- Enable logging and monitoring for all deployed resources
- Use private endpoints or service endpoints instead of public access where possible
- Tag resources with owner, environment, and cost centre for governance
- Pin provider and module versions to avoid unexpected changes
