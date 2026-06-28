# RULES — Frontend Core

**Scope:** Project structure, TypeScript conventions, component library (Shadcn/Tailwind), repeatable component UI.

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

## 4. Repeatable Component UI Rules

### 4.1 Non-Repeatable
- Rendered as collapsible `<fieldset>` with chevron toggle in legend
- Default state: expanded at depth=0 (top-level), collapsed at depth>=1 (nested)
- Collapsed header shows component name + first text field value as hint (truncated 60 chars)
- `aria-expanded` on toggle button; child grid unmounted when collapsed (form values preserved by react-hook-form)

### 4.2 Repeatable
- Rendered as list of bordered cards with controls
- Each entry: collapsible with chevron toggle, numbered header + Move Up/Move Down/Remove buttons
- Each entry starts collapsed by default; move/delete buttons always visible
- Move up disabled on first item; Move down disabled on last
- "Add entry" button at bottom — appends empty object (collapsed)
- Remove: no confirmation (immediate splice and re-index)

### 4.3 Form State
- Array under `fieldName` in form values
- Uses dot-notation with array indexing: `skills.0.category`, `skills.1.category`

---

## 5. Boundaries Summary

| Rule | Detail |
|---|---|
| **Always** | Use named exports (no `export default`) |
| **Always** | Use TypeScript strict mode, no `any` |
| **Always** | Use Shadcn UI + TailwindCSS |
| **Never** | Use `any` type |
| **Never** | Use `export default` |
| **Never** | Import Shadcn from npm — use copied components |
