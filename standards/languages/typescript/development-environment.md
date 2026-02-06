# TypeScript Development Environment Standard

## Overview

DevContainer configuration and VS Code settings for TypeScript development.

---

## DevContainer Features

```jsonc
{
  "features": {
    "ghcr.io/devcontainers/features/node:1": {
      "version": "22",
      "nodeGypDependencies": true
    }
  }
}
```

**Note**: Node.js is required for TypeScript development. The feature includes npm as the default package manager.

---

## VS Code Extensions

```jsonc
{
  "extensions": [
    /* ---------------------------------------------------------------------------
       TypeScript Development
    --------------------------------------------------------------------------- */
    "dbaeumer.vscode-eslint",
    // ESLint integration for real-time linting with TypeScript support.

    "ms-vscode.vscode-typescript-next"
    // Provides the latest TypeScript language features and improvements.
  ]
}
```

| Extension | Purpose |
|-----------|---------|
| `dbaeumer.vscode-eslint` | ESLint integration for linting |
| `ms-vscode.vscode-typescript-next` | Latest TypeScript language features |

**Note**: Prettier (`esbenp.prettier-vscode`) is included in the core DevContainer extensions and applies to TypeScript files.

---

## VS Code Settings

```json
{
  "typescript.preferences.importModuleSpecifier": "relative",
  "typescript.updateImportsOnFileMove.enabled": "always",
  "typescript.suggest.autoImports": true,
  "typescript.inlayHints.parameterNames.enabled": "all",
  "typescript.inlayHints.functionLikeReturnTypes.enabled": true,
  "editor.defaultFormatter": "esbenp.prettier-vscode",
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.fixAll.eslint": "explicit",
    "source.organizeImports": "explicit"
  },
  "[typescript]": {
    "editor.defaultFormatter": "esbenp.prettier-vscode"
  },
  "[typescriptreact]": {
    "editor.defaultFormatter": "esbenp.prettier-vscode"
  }
}
```

---

## Settings Explained

| Setting | Value | Purpose |
|---------|-------|---------|
| `typescript.preferences.importModuleSpecifier` | `relative` | Use relative paths for imports |
| `typescript.updateImportsOnFileMove.enabled` | `always` | Auto-update imports when files move |
| `typescript.suggest.autoImports` | `true` | Enable auto-import suggestions |
| `typescript.inlayHints.parameterNames.enabled` | `all` | Show parameter name hints |
| `editor.defaultFormatter` | `esbenp.prettier-vscode` | Use Prettier for formatting |
| `editor.formatOnSave` | `true` | Auto-format on save |
| `source.fixAll.eslint` | `explicit` | Apply ESLint fixes on save |
| `source.organizeImports` | `explicit` | Organise imports on save |
