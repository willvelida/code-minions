# TypeScript Tooling Standard

## Overview

Standard tooling for TypeScript development: package management, linting, and formatting.

---

## Package Manager: pnpm

**pnpm** is the recommended package manager for TypeScript projects.

**Rationale**:
- Faster installation than npm
- Disk-efficient through content-addressable storage
- Strict dependency resolution prevents phantom dependencies
- Drop-in replacement for npm commands

**Installation** (add to `.devcontainer/post-create.sh`):

```bash
npm install -g pnpm
```

**Usage:**

```bash
# Install dependencies
pnpm install

# Add a dependency
pnpm add express

# Add a dev dependency
pnpm add -D typescript

# Run scripts
pnpm run build
pnpm test
```

---

## Linter: ESLint

**ESLint** with TypeScript parser is the mandatory linter.

**Installation:**

```bash
pnpm add -D eslint @typescript-eslint/parser @typescript-eslint/eslint-plugin
```

**Configuration** (`eslint.config.mjs` - flat config):

```javascript
import eslint from '@eslint/js';
import tseslint from 'typescript-eslint';

export default tseslint.config(
  eslint.configs.recommended,
  ...tseslint.configs.recommended,
  {
    languageOptions: {
      parserOptions: {
        project: true,
        tsconfigRootDir: import.meta.dirname,
      },
    },
    rules: {
      '@typescript-eslint/explicit-function-return-type': 'warn',
      '@typescript-eslint/no-unused-vars': ['error', { argsIgnorePattern: '^_' }],
      '@typescript-eslint/no-explicit-any': 'warn',
    },
  }
);
```

---

## Formatter: Prettier

**Prettier** is the mandatory formatter for TypeScript projects.

**Note**: The `esbenp.prettier-vscode` extension is included in the core DevContainer extensions.

**Configuration** (`.prettierrc`):

```json
{
  "semi": true,
  "trailingComma": "es5",
  "singleQuote": true,
  "printWidth": 100,
  "tabWidth": 2,
  "useTabs": false
}
```

---

## TypeScript Compiler

The TypeScript compiler (`tsc`) is included with the `typescript` package.

**Usage:**

```bash
# Type check without emitting
pnpm exec tsc --noEmit

# Build
pnpm exec tsc

# Watch mode
pnpm exec tsc --watch
```
