# RULES — Frontend Data

**Scope:** TanStack Query, query keys, mutations, invalidation, health ping, locale switching.

---

## 1. Data Fetching Rules (TanStack Query)

### 1.1 Query Hooks
- `useSingleTypeDocument(slug, locale)` — 404 returns `undefined` (not error)
- `useCollectionDocuments(slug, start, size, locale)` — paginated
- `useCollectionDocument(slug, documentId, locale)` — single document
- `useContentTypes()` — list all content types
- `useLocales()` — returns `Locale[]` objects

### 1.2 Mutation Hooks
- Every `useMutation` **MUST** invalidate affected query keys on success
- Use `onMutationError` for consistent error toast handling
- Mutation hooks: `useUpdateDocument`, `usePublishDocument`, `useUnpublishDocument`, `useDuplicateCollectionDocument`, `useCreateLocale`, `useUpdateLocale`, `useDeleteLocale`, `useUpdateListFields`

### 1.3 Query Key Patterns
```typescript
['documents', 'single-type', slug]
['documents', 'collection-type', slug]
['documents', 'detail', slug, documentId, 'data']
['content-types']
['locales']
['media', 'list']
['users']
['roles']
['invites']
['access-tokens']
```

### 1.4 Health Ping (Exception to TanStack Query Rule)
- `HealthProvider` in `src/context/HealthContext.tsx` uses standalone `fetch` with `setTimeout` loop — NOT TanStack Query
- This is intentional: the ping is a side-effect loop with custom 10s/14m timing, not request-response data fetching
- Uses standalone `fetch` — **NEVER** the `api` axios instance (avoids auth interceptor)
- `useHealthStatus()` hook exposes `{ isApiHealthy: boolean }`

### 1.5 Invariants
- **NEVER** use raw `useEffect` + `useState` for API calls — use TanStack Query (exception: health ping loop)
- **ALWAYS** invalidate queries after mutations
- **ALWAYS** use query key constants (not string literals)

---

## 2. Locale Switching Rules

### 2.1 `useLocales()` Hook
- Fetches `Locale[]` objects (code + name + isDefault) on mount
- Returns full Locale objects, not just strings

### 2.2 LocaleSelector Behavior
- Displays language **name** — uses **code** as value
- Default locale pre-selected when no explicit selection
- On locale change: re-fetch document → form reset → `isDirty = false`
- All mutations forward `locale: activeLocale`

### 2.3 CollectionListPage Locale Integration
- Locale selector always visible (even with one locale)
- Changing locale: update state → reset pagination to `start = 0` → refetch

---

## 3. Boundaries Summary

| Rule | Detail |
|---|---|
| **Always** | Use TanStack Query for all server state |
| **Always** | Invalidate affected query keys on mutation success |
| **Always** | Display locale name in dropdowns (not code) |
| **Always** | Pre-select default locale when no selection |
| **Never** | Use raw `useEffect` + `useState` for API calls |
| **Never** | Hardcode locale lists — fetch from API |
