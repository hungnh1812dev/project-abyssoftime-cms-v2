# RULES — Frontend (apps/web)

**Scope:** All React/TypeScript code in `apps/web/`. Cross-cutting rules for UI components, hooks, routing, forms, and state management.
**References:** SPEC.md §FE Spec, docs/guide.md

---

## 1. Project Structure Rules

### 1.1 Directory Layout
```
apps/web/src/
├── components/          # Reusable components
│   ├── ui/              # Shadcn UI primitives (owned, not imported)
│   ├── form/            # FormProvider, FormField, inputs/
│   ├── collection/      # Collection-specific components
│   ├── content-type/    # Content-type components
│   ├── locale/          # Locale components
│   ├── media/           # Media components
│   └── sidebar/         # Sidebar navigation
├── content-type-registry/  # Metadata-only content-type registry
├── context/             # React context providers
├── hooks/               # Custom hooks (data fetching, mutations)
├── lib/                 # Utilities, API client
├── pages/               # Page components
│   ├── admin/           # Admin pages
│   │   ├── layout/      # Layout components
│   │   ├── panels/      # Content-type panels
│   │   └── settings/    # Settings pages
│   └── auth/            # Auth pages (login, register, invite)
└── types/               # TypeScript type definitions
```

### 1.2 File Naming
- Components: `PascalCase.tsx` (e.g., `CollectionListPage.tsx`)
- Hooks: `camelCase.ts` (e.g., `useCollectionDocuments.ts`)
- Types: `camelCase.ts` (e.g., `cms.ts`)
- Tests: co-located in `__tests__/` directory (e.g., `__tests__/CollectionListPage.test.tsx`)

### 1.3 Exports
- **Named exports only** — no `export default`
- Lazy-loaded components use named export wrapper:
  ```tsx
  const SomePanel = lazy(() =>
    import('@/pages/admin/panels/SomePanel').then((m) => ({ default: m.SomePanel }))
  )
  ```

---

## 2. TypeScript Rules

### 2.1 Strict Mode
- `strict: true` in tsconfig — no exceptions
- **NEVER** use `any` type — use `unknown` or proper types
- All function parameters and return types must be typed
- Interface over type alias for object shapes

### 2.2 Type Definitions
- All shared types in `src/types/cms.ts`
- Key types: `ContentType`, `Document`, `MediaAsset`, `Locale`, `User`, `Role`, `FieldDefinition`
- API response types match backend JSON shapes exactly

### 2.3 Path Aliases
- `@/components` → `src/components/`
- `@/hooks` → `src/hooks/`
- `@/lib` → `src/lib/`
- `@/types` → `src/types/`
- `@/pages` → `src/pages/`
- `@/context` → `src/context/`
- `@/content-type-registry` → `src/content-type-registry/`

---

## 3. Component Library Rules

### 3.1 Shadcn UI
- Components are **copied** into `src/components/ui/` — fully owned
- **NEVER** import from a Shadcn npm package
- Built on Radix UI primitives (proper ARIA, keyboard nav, focus)
- Use TailwindCSS utilities directly — no CSS-in-JS

### 3.2 Available Input Components
| Component | Path | Use For |
|---|---|---|
| `TextInput` | `@/components/form/inputs/TextInput` | Short text, URLs |
| `NumberInput` | `@/components/form/inputs/NumberInput` | Integers, decimals |
| `BooleanInput` | `@/components/form/inputs/BooleanInput` | Toggles, flags |
| `RichTextInput` | `@/components/form/inputs/RichTextInput` | HTML content (CKEditor) |
| `JsonInput` | `@/components/form/inputs/JsonInput` | Arbitrary JSON (CodeMirror) |
| `MediaInput` | `@/components/form/inputs/MediaInput` | Image/file upload |
| `RepeatableComponentField` | `@/components/form/inputs/RepeatableComponentField` | Ordered component arrays |

### 3.3 Form Field Grid Layout
- Fields render in a responsive 6-column grid: `grid grid-cols-1 md:grid-cols-6 gap-4`
- Each field's `width` property controls its column span:
  - `"100%"` or omitted → `md:col-span-6` (full width)
  - `"50%"` → `md:col-span-3` (half width)
  - `"1/3"` → `md:col-span-2` (one-third width)
- Component fields always span full width (`md:col-span-6`)
- Mobile: all fields are full width (`grid-cols-1`)
- The grid applies at every nesting level (top-level, inside components, inside repeatable entries)

### 3.4 UI Design System
- Indigo color tokens
- Sidebar navigation
- Sticky action bar
- Dark mode support

---

## 4. FormProvider Rules

### 4.1 Lifecycle
| Moment | Behavior |
|---|---|
| Initial load | Fields rendered, pre-filled from server data |
| Clean state | Save button disabled (`isDirty === false`) |
| After edit | Save button enabled (`isDirty === true`) |
| Failed save | `toast.error(msg)`, form stays edited |
| Successful save | `toast.success('Saved')` → invalidate → reset → Save disabled |

### 4.2 FormProvider Props
- `query`: TanStack Query config for initial data fetch
- `mutationFn`: function to save data

### 4.3 Field Name Rules
- Flat name (`siteName`) → `{ siteName: "..." }`
- Dot-notation (`seo.title`) → `{ seo: { title: "..." } }`
- Array indexing (`skills.0.name`) → `{ skills: [{ name: "..." }] }`
- `FormProvider` handles conversion automatically

### 4.4 Invariants
- **NEVER** use `React.Children.map` or recursive child scanning
- **NEVER** use drag-and-drop or dynamic form engine
- FormProvider manages loading, submitting, isDirty — **NEVER** duplicate this state

---

## 5. Data Fetching Rules (TanStack Query)

### 5.1 Query Hooks
- `useSingleTypeDocument(slug, locale)` — 404 returns `undefined` (not error)
- `useCollectionDocuments(slug, start, size, locale)` — paginated
- `useCollectionDocument(slug, documentId, locale)` — single document
- `useContentTypes()` — list all content types
- `useLocales()` — returns `Locale[]` objects

### 5.2 Mutation Hooks
- Every `useMutation` **MUST** invalidate affected query keys on success
- Use `onMutationError` for consistent error toast handling
- Mutation hooks: `useUpdateDocument`, `usePublishDocument`, `useUnpublishDocument`, `useDuplicateCollectionDocument`, `useCreateLocale`, `useUpdateLocale`, `useDeleteLocale`, `useUpdateListFields`

### 5.3 Query Key Patterns
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

### 5.4 Health Ping (Exception to TanStack Query Rule)
- `HealthProvider` in `src/context/HealthContext.tsx` uses standalone `fetch` with `setTimeout` loop — NOT TanStack Query
- This is intentional: the ping is a side-effect loop with custom 10s/14m timing, not request-response data fetching
- Uses standalone `fetch` — **NEVER** the `api` axios instance (avoids auth interceptor)
- `useHealthStatus()` hook exposes `{ isApiHealthy: boolean }`

### 5.5 Invariants
- **NEVER** use raw `useEffect` + `useState` for API calls — use TanStack Query (exception: health ping loop)
- **ALWAYS** invalidate queries after mutations
- **ALWAYS** use query key constants (not string literals)

---

## 6. Routing Rules

### 6.1 Route Structure
| Route | Component | Description |
|---|---|---|
| `/admin/content-type/single-type/:slug` | SingleTypePage | Single-type editor |
| `/admin/content-type/collection-type/:slug` | CollectionListPage | Collection list |
| `/admin/content-type/collection-type/:slug/new` | CollectionDetailPage | Create entry |
| `/admin/content-type/collection-type/:slug/:id` | CollectionDetailPage | Edit entry |
| `/admin/settings/internationalize` | InternationalizePage | Locale management |
| `/admin/settings/media` | MediaLibraryPage | Media management |
| `/admin/settings/users` | UsersPage | User management |
| `/admin/settings/access-tokens` | AccessTokensPage | Token management |
| `/admin/settings/roles` | RolesPage | Permission matrix |

### 6.2 Custom Panels
- Registered as specific routes **before** generic catch-all
- Receive `ContentType` prop via a wrapper component
- Use lazy loading with named exports

### 6.3 Protected Routes
- `ProtectedRoute` component with optional `minRole` prop
- Settings links filtered by role in sidebar

---

## 7. Locale Switching Rules

### 7.1 `useLocales()` Hook
- Fetches `Locale[]` objects (code + name + isDefault) on mount
- Returns full Locale objects, not just strings

### 7.2 LocaleSelector Behavior
- Displays language **name** — uses **code** as value
- Default locale pre-selected when no explicit selection
- On locale change: re-fetch document → form reset → `isDirty = false`
- All mutations forward `locale: activeLocale`

### 7.3 CollectionListPage Locale Integration
- Locale selector always visible (even with one locale)
- Changing locale: update state → reset pagination to `start = 0` → refetch

---

## 8. Content-Type Registry Rules

### 8.1 Purpose
- Metadata-only module at `src/content-type-registry/index.ts`
- Maps content-type slugs to custom panel configuration
- Registry overrides take precedence over dynamic behavior

### 8.2 Column Override
- If a registry entry defines `columns`, the column chooser button is hidden
- Non-registry content types use the dynamic column chooser

---

## 9. Sidebar Rules

### 9.1 Content Types
- Eager metadata load from API
- Lazy component loading via `React.lazy`
- Grouped by kind (single-type, collection-type)

### 9.2 Settings Links (Role-Filtered)
| Link | Visible To |
|---|---|
| Media Library | All authenticated |
| Users | admin+ |
| Access Tokens | super_admin only |
| Roles | super_admin only |
| Internationalize | super_admin only |

---

## 10. Testing Rules (Frontend-Specific)

### 10.1 Test Stack
- Vitest + React Testing Library + MSW
- Tests in co-located `__tests__/` directories

### 10.2 Test Patterns
- Test user behavior, not implementation details
- Use `screen.getByRole`, `screen.getByText`, `screen.getByLabelText`
- Mock API calls with MSW, not by mocking hooks
- Verify query invalidation happens after mutations

### 10.3 Required Tests
- Every new component needs test file
- Every new hook needs test file
- Page tests: renders correctly, handles loading/error states
- Form tests: submit, validation, dirty state
- Dialog tests: open/close, confirm/cancel actions

---

## 11. Repeatable Component UI Rules

### 11.1 Non-Repeatable
- Rendered as collapsible `<fieldset>` with chevron toggle in legend
- Default state: expanded at depth=0 (top-level), collapsed at depth>=1 (nested)
- Collapsed header shows component name + first text field value as hint (truncated 60 chars)
- `aria-expanded` on toggle button; child grid unmounted when collapsed (form values preserved by react-hook-form)

### 11.2 Repeatable
- Rendered as list of bordered cards with controls
- Each entry: collapsible with chevron toggle, numbered header + Move Up/Move Down/Remove buttons
- Each entry starts collapsed by default; move/delete buttons always visible
- Move up disabled on first item; Move down disabled on last
- "Add entry" button at bottom — appends empty object (collapsed)
- Remove: no confirmation (immediate splice and re-index)

### 11.3 Form State
- Array under `fieldName` in form values
- Uses dot-notation with array indexing: `skills.0.category`, `skills.1.category`

---

## 12. Boundaries Summary

| Rule | Detail |
|---|---|
| **Always** | Use TanStack Query for all server state |
| **Always** | Invalidate affected query keys on mutation success |
| **Always** | Use named exports (no `export default`) |
| **Always** | Use TypeScript strict mode, no `any` |
| **Always** | Use Shadcn UI + TailwindCSS |
| **Always** | Display locale name in dropdowns (not code) |
| **Always** | Pre-select default locale when no selection |
| **Always** | FormProvider manages loading/submitting/isDirty |
| **Always** | Lazy-load panels with `React.lazy` |
| **Always** | Co-locate tests in `__tests__/` directories |
| **Never** | Use `any` type |
| **Never** | Use `export default` |
| **Never** | Use raw `useEffect` + `useState` for API calls |
| **Never** | Use `React.Children.map` in FormProvider |
| **Never** | Use drag-and-drop or dynamic form engine |
| **Never** | Hardcode locale lists — fetch from API |
| **Never** | Import Shadcn from npm — use copied components |
