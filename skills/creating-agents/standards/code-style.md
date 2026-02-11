# Code Style Standard

## Purpose

Defines how to document code style expectations in agent definitions. Code examples are far more effective than written descriptions — one real snippet showing your preferred style beats three paragraphs describing it.

---

## Rules

| Rule | Requirement |
|------|-------------|
| Show, don't tell | Provide actual code snippets, not prose descriptions |
| Good and bad examples | Show what to do AND what not to do |
| Cover naming conventions | Functions, classes, constants, files |
| Include error handling | Show preferred error handling patterns |
| Keep it focused | 2-4 examples per category, not exhaustive guides |
| Match the tech stack | Examples must use the project's actual language and frameworks |

---

## Categories to Cover

### Naming Conventions

Define naming rules for each code element:

```markdown
**Naming conventions:**
- Functions: camelCase (`getUserData`, `calculateTotal`)
- Classes: PascalCase (`UserService`, `DataController`)
- Constants: UPPER_SNAKE_CASE (`API_KEY`, `MAX_RETRIES`)
- Files: kebab-case (`user-service.ts`, `data-controller.ts`)
- Test files: `<name>.test.ts` or `<name>.spec.ts`
```

### Code Patterns

Show preferred patterns with real code:

```markdown
**Code style example:**

✅ Good — descriptive names, proper error handling:
```typescript
async function fetchUserById(id: string): Promise<User> {
  if (!id) throw new Error('User ID required');

  const response = await api.get(`/users/${id}`);
  return response.data;
}
```

❌ Bad — vague names, no error handling:
```typescript
async function get(x) {
  return await api.get('/users/' + x).data;
}
```
```

### File Organisation

Show how files should be structured:

```markdown
**File organisation:**
- One component per file
- Imports ordered: external → internal → types
- Exports at bottom of file
```

---

## Examples

**Good — Real code, good/bad comparison:**

```markdown
## Code Style

**Naming:**
- Components: PascalCase (`UserProfile.tsx`)
- Hooks: `use` prefix (`useAuth`, `useForm`)
- Utils: camelCase (`formatDate.ts`)

**Component pattern:**
✅ Good:
```tsx
export function UserProfile({ userId }: UserProfileProps) {
  const { data, isLoading } = useUser(userId);

  if (isLoading) return <Spinner />;
  return <ProfileCard user={data} />;
}
```

❌ Bad:
```tsx
export default function(props) {
  const d = getData(props.id);
  return <div>{d.name}</div>;
}
```
```

**Bad — No code examples:**

```markdown
## Code Style
Write clean code. Use good names. Handle errors properly.
Follow best practices. Keep functions small.
```

---

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| Prose without code | Add at least one code snippet per convention |
| Only good examples | Show bad examples too — agents learn from contrast |
| Generic snippets | Use code from your actual project or tech stack |
| Exhaustive style guide | Keep to 2-4 key patterns — link to full guides if needed |
| No naming conventions | Always cover function, class, and file naming |
