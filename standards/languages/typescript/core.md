# TypeScript Core Standard

## Overview

Standards for TypeScript development including project structure, naming conventions, and coding style.

## Principles

1. **Type Safety** — Leverage TypeScript's type system fully
2. **Consistency** — Follow established conventions and patterns
3. **Modern Standards** — Use ES2022+ features with appropriate targets
4. **Developer Experience** — Optimise for IDE support and tooling

---

## Project Structure

### Recommended Layout

```
project/
├── src/
│   ├── index.ts           # Entry point
│   ├── types/             # Type definitions
│   │   └── index.ts
│   ├── utils/             # Utility functions
│   │   └── index.ts
│   ├── services/          # Business logic
│   │   └── index.ts
│   └── models/            # Data models
│       └── index.ts
├── tests/
│   ├── unit/
│   └── integration/
├── dist/                  # Compiled output
├── tsconfig.json
├── eslint.config.mjs
├── .prettierrc
├── package.json
└── README.md
```

### Barrel Exports

Use barrel exports (`index.ts`) for clean imports:

```typescript
// src/utils/index.ts
export * from './stringUtils';
export * from './dateUtils';
export * from './validators';
```

---

## TypeScript Configuration

### Base Configuration (`tsconfig.json`)

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "NodeNext",
    "moduleResolution": "NodeNext",
    "lib": ["ES2022"],
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true,
    "outDir": "./dist",
    "rootDir": "./src",
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": false,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noImplicitReturns": true,
    "noFallthroughCasesInSwitch": true,
    "noUncheckedIndexedAccess": true
  },
  "include": ["src/**/*"],
  "exclude": ["node_modules", "dist"]
}
```

### Strict Mode Requirements

All TypeScript projects **MUST** enable strict mode:

| Option | Purpose |
|--------|---------|
| `strict` | Enables all strict type-checking options |
| `noImplicitAny` | Error on expressions with implied `any` |
| `strictNullChecks` | Strict null and undefined checking |
| `strictFunctionTypes` | Strict function type checking |
| `strictBindCallApply` | Strict `bind`, `call`, and `apply` |
| `strictPropertyInitialization` | Strict property initialization |
| `noImplicitThis` | Error on `this` with implied `any` |
| `alwaysStrict` | Parse in strict mode |

---

## Naming Conventions

| Type | Convention | Example |
|------|------------|---------|
| Variables | camelCase | `userName`, `isActive` |
| Functions | camelCase | `getUserById`, `validateInput` |
| Classes | PascalCase | `UserService`, `HttpClient` |
| Interfaces | PascalCase | `UserProfile`, `ApiResponse` |
| Types | PascalCase | `UserId`, `ConfigOptions` |
| Enums | PascalCase | `UserRole`, `HttpStatus` |
| Constants | UPPER_SNAKE_CASE | `MAX_RETRIES`, `API_URL` |
| Files | kebab-case | `user-service.ts`, `api-client.ts` |

---

## Type Annotations

```typescript
// Explicit return types for public functions
function calculateTotal(items: Item[]): number {
  return items.reduce((sum, item) => sum + item.price, 0);
}

// Use interfaces for object shapes
interface UserProfile {
  id: string;
  name: string;
  email: string;
  createdAt: Date;
}

// Use type aliases for unions and complex types
type UserId = string | number;
type ApiResponse<T> = { data: T; status: number } | { error: string };

// Prefer readonly for immutable data
interface Config {
  readonly apiUrl: string;
  readonly timeout: number;
}
```

---

## Async/Await Standards

```typescript
// Always handle errors in async functions
async function fetchUser(id: string): Promise<User> {
  try {
    const response = await api.get<User>(`/users/${id}`);
    return response.data;
  } catch (error) {
    if (error instanceof ApiError) {
      throw new UserNotFoundError(id);
    }
    throw error;
  }
}

// Use Promise.all for parallel operations
async function fetchUserData(userId: string): Promise<UserData> {
  const [profile, settings, preferences] = await Promise.all([
    fetchProfile(userId),
    fetchSettings(userId),
    fetchPreferences(userId),
  ]);
  return { profile, settings, preferences };
}
```
