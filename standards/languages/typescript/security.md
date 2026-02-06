# TypeScript Security Standard

## Overview

Security requirements for TypeScript code.

---

## Prohibited Practices

| Practice | Reason | Alternative |
|----------|--------|-------------|
| `any` type | Bypasses type safety | Use `unknown` or proper types |
| `@ts-ignore` | Hides type errors | Fix the underlying issue |
| `eval()` | Code injection risk | Use safer alternatives |
| Non-null assertion (`!`) | Can cause runtime errors | Use proper null checking |

---

## Secure Practices

### Use `unknown` Instead of `any`

```typescript
// Use unknown for external data
function processExternalData(data: unknown): UserProfile {
  if (isUserProfile(data)) {
    return data;
  }
  throw new ValidationError('Invalid user profile data');
}
```

### Type Guards for Runtime Validation

```typescript
function isUserProfile(data: unknown): data is UserProfile {
  return (
    typeof data === 'object' &&
    data !== null &&
    'id' in data &&
    'name' in data &&
    'email' in data
  );
}
```

### Runtime Validation with Zod

```typescript
import { z } from 'zod';

const UserSchema = z.object({
  id: z.string().uuid(),
  name: z.string().min(1),
  email: z.string().email(),
});

type User = z.infer<typeof UserSchema>;

function parseUser(data: unknown): User {
  return UserSchema.parse(data);
}
```

### Proper Null Handling

```typescript
// ❌ Avoid non-null assertion
const user = getUser()!;

// ✅ Use proper null checking
const user = getUser();
if (!user) {
  throw new Error('User not found');
}
```

---

## Compliance Checklist

- [ ] No `any` types (use `unknown` or proper types)
- [ ] No `@ts-ignore` comments
- [ ] No `eval()` usage
- [ ] Minimal use of non-null assertions
- [ ] Type guards for external data validation
- [ ] Proper error handling in async code
