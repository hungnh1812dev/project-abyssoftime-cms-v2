# SPEC — personal-cms

A lightweight, code-first Personal Headless CMS. Developers define content panels manually in React; the Go backend enforces strict data contracts via Clean Architecture.

---

## 1. Objective

Build a self-hosted headless CMS for developers and administrators who prefer code over drag-and-drop builders. Content panels are hard-coded React pages. The backend exposes a REST API (GraphQL later) backed by MongoDB, deployed as a single Render.com service running `docker-compose`.

Every content entry follows a **draft → publish** workflow: editors save changes as a draft at any time; the public read API keeps serving the last published version until an explicit Publish action promotes the draft. Content types are grouped as **single-type** (one auto-created singleton entry, e.g. Homepage settings) or **collection-type** (many entries, e.g. Blog posts).

Content-**type structure** (the schema: which fields exist, their types) is defined as code — JSON definition files checked into the repo — never through the API or UI. On API startup, a sync step reads every definition file and creates, updates, or removes the corresponding `ContentType` records in MongoDB. The web UI is for content **data** only (create/edit/delete entries); it never defines or edits the structure itself.

**Not in scope (v1):** PostgreSQL support (document storage is designed for future migration), Localization/i18n beyond basic locale field.

---

## 2. Commands

### Native development (recommended)

MongoDB runs as a container; API and web run natively. Full step-by-step walkthrough lives in `docs/local-dev.md` — summarized here so the spec is self-contained.

**Prerequisites:** Go 1.21+, Node.js 20 LTS, Docker or Podman (for the MongoDB container only — no native Mongo install needed).

| Command | Description |
|---|---|
| `make mongo-start` | Start MongoDB container (`cms-mongo`, port 27017, persistent volume `cms-mongo-data`). Idempotent. |
| `make mongo-stop` | Stop the MongoDB container |
| `make dev` | Start API + web in parallel (Ctrl-C stops both) |
| `make dev-api` | Start Go API server only |
| `make dev-web` | Start Vite dev server only |
| `make test-api` | `go test ./...` inside `apps/api` |
| `make test-web` | `vitest run` inside `apps/web` |

Podman users: prefix any `make` target with `CONTAINER_CLI=podman`.

**Quick start:** `cd apps/web && npm install && cd ../..` → `make mongo-start` → `make dev`. Web app at `http://localhost:5173`, API at `http://localhost:8080`.

**Environment variables** (copy `.env.example` → `.env`, never commit `.env`):

| Variable | Description | Default |
|---|---|---|
| `PORT` | API listen port | `8080` |
| `MONGODB_URI` | MongoDB connection string | `mongodb://localhost:27017/cms` |
| `JWT_SECRET` | Secret for signing JWTs | *(required, no default)* |
| `CLOUDINARY_CLOUD_NAME` / `CLOUDINARY_API_KEY` / `CLOUDINARY_API_SECRET` | Cloudinary credentials | *(required for media upload)* |
| `CONTENT_TYPES_DIR` | Dir of JSON content-type definitions synced on boot | `content-types` |
| `STORAGE_PROVIDER` | Active media adapter | `s3` \| `cloudinary` |
| `VITE_API_URL` | API base URL for the Vite proxy | `http://localhost:8080` |

**Troubleshooting:**
- *"mongodb connect" error on boot* → MongoDB container isn't running; `make mongo-start` (verify with `docker ps | grep cms-mongo`).
- *"JWT_SECRET not set" panic* → export `JWT_SECRET` before `make dev-api`.
- *Vite proxy returns 502* → API isn't running; start it with `make dev-api` or use `make dev`.
- *Port already in use (8080/5173)* → `lsof -ti:8080 | xargs kill` and retry.

### Docker Compose (full stack)
| Command | Description |
|---|---|
| `docker-compose -f docker-compose.yml -f docker-compose.dev.yml up --build` | First run: build images + start all services with hot reload |
| `docker-compose -f docker-compose.yml -f docker-compose.dev.yml up` | Start services after images are already built |
| `docker-compose down` | Stop and remove all containers |
| `docker-compose logs -f api` | Tail API logs |
| `docker-compose logs -f web` | Tail web (Vite) logs |

### Pre-deploy verification (local)
| Command | Description |
|---|---|
| `docker-compose run --rm api go vet ./...` | Go static analysis in a clean container |
| `docker-compose run --rm api go test ./...` | Backend tests in a clean container |
| `docker-compose run --rm web npm run lint` | Frontend lint in a clean container |
| `docker-compose run --rm web npm run build` | Verify production frontend build compiles |
| `docker-compose build` | Build production images |
| `docker-compose up` | Start production stack locally for smoke testing |
| `docker-compose down` | Tear down after verification |

### Production (Render.com)
| Command | Description |
|---|---|
| `docker-compose up --build` | Build production images and start all services |
| `docker-compose build` | Rebuild production images without starting |

### Backend (`apps/api`)
| Command | Description |
|---|---|
| `go run ./cmd/server` | Start the API server |
| `go test ./...` | Run all tests |
| `go build -o bin/server ./cmd/server` | Compile production binary |
| `go vet ./...` | Static analysis |

### Frontend (`apps/web`)
| Command | Description |
|---|---|
| `npm run dev` | Start Vite dev server |
| `npm run build` | Production build |
| `npm run lint` | ESLint check |
| `npm run preview` | Preview production build locally |

---

## 3. Project Structure

```
personal-cms/
├── apps/
│   ├── api/                          # Go backend
│   │   ├── cmd/
│   │   │   └── server/               # main.go — entry point; runs content-type sync on boot
│   │   ├── content-types/            # JSON schema-as-code definition files (one per content type)
│   │   ├── internal/
│   │   │   ├── domain/               # Entities, value objects, repository interfaces
│   │   │   │   ├── entity/           # User, Document (draft/published record), ContentType, MediaAsset
│   │   │   │   └── repository/       # Pure interfaces (no DB code)
│   │   │   ├── usecase/              # Application business logic (DB-agnostic)
│   │   │   │   ├── auth/
│   │   │   │   ├── document/
│   │   │   │   ├── content_type/     # includes Sync: reconciles JSON definitions → DB on startup
│   │   │   │   └── media/
│   │   │   ├── infrastructure/       # Concrete adapters
│   │   │   │   ├── mongodb/          # MongoDB repository implementations
│   │   │   │   └── storage/          # S3 / Cloudinary adapters
│   │   │   └── delivery/
│   │   │       └── http/             # REST handlers, middleware, router
│   │   │           ├── handler/
│   │   │           ├── middleware/
│   │   │           └── router.go
│   │   ├── pkg/                      # Shared utilities (jwt, errors, pagination)
│   │   ├── Dockerfile
│   │   └── go.mod                    # module: project-abyssoftime-cms-v2/api
│   │
│   └── web/                          # React frontend
│       ├── src/
│       │   ├── pages/
│       │   │   ├── auth/             # Login, Register pages
│       │   │   └── admin/
│       │   │       ├── panels/       # Hard-coded content panels (one file per panel)
│       │   │       │   ├── single/       # Single-type panels (singleton entry: edit/save/publish only)
│       │   │       │   └── collection/   # Collection-type panels (list + entry edit/save/publish)
│       │   │       └── layout/       # Admin shell, sidebar (grouped: Single Types / Collection Types), nav
│       │   ├── components/
│       │   │   ├── form/
│       │   │   │   ├── FormProvider.tsx
│       │   │   │   ├── FormField.tsx
│       │   │   │   └── inputs/
│       │   │   │       ├── TextInput.tsx
│       │   │   │       ├── RichTextInput.tsx
│       │   │   │       ├── JsonInput.tsx
│       │   │   │       ├── BooleanInput.tsx
│       │   │   │       ├── MediaInput.tsx
│       │   │   │       └── NumberInput.tsx
│       │   │   └── ui/               # Shadcn-generated components
│       │   ├── hooks/                # TanStack Query hooks: useQuery + useMutation per resource
│       │   ├── lib/
│       │   │   ├── api.ts            # Axios/fetch client with JWT interceptors
│       │   │   ├── queryClient.ts    # TanStack QueryClient singleton + default options
│       │   │   └── utils.ts
│       │   ├── router.tsx            # React Router routes
│       │   └── main.tsx
│       ├── Dockerfile
│       ├── vite.config.ts
│       └── package.json
│
├── .github/
│   └── workflows/
│       └── ci.yml                    # Lint → test → build → push images → deploy to Render
├── docker-compose.yml                # api + web + mongodb services
└── SPEC.md
```

---

## 4. Code Style

### Go (Backend)
- **Architecture**: Strict Clean Architecture — `usecase` layer must import only `domain`; `delivery` and `infrastructure` import `usecase` and `domain`. Zero cross-layer leakage.
- **Error handling**: Wrap errors at the usecase boundary; handlers map domain errors to HTTP status codes. No naked `error` strings in HTTP responses.
- **Naming**: `PascalCase` for exported identifiers; `camelCase` for unexported. Repository interfaces use `<Entity>Repository` (e.g., `DocumentRepository`).
- **Database IDs**: Document entities use `documentId` as the primary domain identifier (replaces `entryId` and MongoDB `_id`). Higher layers (usecase, handler, frontend) only work with `documentId` and content-type `slug` — never with MongoDB `_id` or generic `id`. ContentType entities retain their MongoDB `_id` as `ID`.
- **Cascade deletion**: Implemented in the usecase layer, not as DB-level triggers, to remain DB-agnostic.
- **No `init()` functions** in business logic packages.

### React (Frontend)
- **TypeScript**: Strict mode enabled. No `any`.
- **Component files**: One component per file. Named exports only (no default exports for components).
- **Form system rules** (enforced by code review):
  - `FormProvider` manages `loading`, `submitting`, and initial data fetch state automatically.
  - `FormField` is the only mechanism for injecting `register`/`control` into inputs.
  - `FormProvider` must never use `React.Children.map` or any recursive child scanning.
  - Nested field names use dot notation (`block.house.title`); `FormProvider` converts to nested JSON on submit.
- **Panels**: Each content panel is a standalone page file in `src/pages/admin/panels/`. No dynamic panel engine. The generic `ContentTypePanelPage` handles any content type automatically; see `docs/guide.md` for the full walkthrough on writing a custom panel (query hook → mutations → `FormProvider`/`FormField` form → route registration).
- **Styling**: TailwindCSS utility classes. Shadcn UI components as the base. No inline `style` props.
- **Imports**: Path aliases (`@/components`, `@/lib`, etc.) configured in `vite.config.ts`.

### Data Fetching (TanStack Query)
All server state is managed exclusively through **TanStack Query** (`@tanstack/react-query`). Raw `useEffect` + `useState` for API calls is not permitted.

**Pattern: query hooks in `src/hooks/`**
- One `useQuery` hook per resource (e.g., `useDocument`, `useDocumentList`, `useContentType`).
- One `useMutation` hook per write operation (e.g., `useUpdateDocument`, `usePublishDocument`, `useDeleteDocument`).
- Query keys are namespaced by resource: `['documents', contentTypeSlug]`, `['documents', 'detail', contentTypeSlug, documentId, locale]`.
- All document hooks accept `contentTypeSlug` (not `contentTypeId`) as the primary identifier for routing to the correct per-content-type collection.

**Refetch after mutation — mandatory rule:**
Every `useMutation` that changes data **must** call `queryClient.invalidateQueries` in its `onSuccess` callback to invalidate all affected query keys. The frontend must never display stale data after a successful write.

```ts
// Example pattern
const { mutate } = useMutation({
  mutationFn: (data) => api.put(`/api/content-types/${slug}/documents/${documentId}`, { data }),
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['documents', slug] });
  },
});
```

**FormProvider integration:**
`FormProvider` uses a `useQuery` hook (passed as a prop or config) to fetch initial data, and calls the panel's `useMutation` on submit. The `isLoading` / `isFetching` / `isPending` states from TanStack Query drive `FormProvider`'s `loading` and `submitting` states — no manual state duplication.

### Auth
- Access token: short-lived JWT (15 min), stored in memory (React state/context).
- Refresh token: long-lived JWT (7 days), stored in `HttpOnly` cookie.
- All protected routes require `Authorization: Bearer <access_token>` header.
- Roles: `admin` (full access) and `guest` (read-only).

### Domain Rules: Draft & Publish
- Every content entry is identified by a `documentId` and stored as **two separate records in a per-content-type MongoDB collection** (`documents_<content_type_slug>`): a `draft` record and a `published` record (`version: "draft" | "published"`).
- Each content type's documents live in their own standalone collection, created automatically during content-type sync on API boot. This design prepares for future PostgreSQL migration where each content type maps to its own table.
- `draft` record: holds the latest edits, `createdAt`/`updatedAt`/`createdBy`/`updatedBy`, **no** `publishedAt`/`publishedBy`.
- `published` record: only exists once the entry has been published at least once; additionally carries `publishedAt`/`publishedBy`.
- Every record (draft and published) also carries `locale` (defaults to `"en"`) — set automatically by the system, not authored by editors in v1.
- **Documents are only created on explicit save.** No document exists until the user performs a Save action. Publish/unpublish buttons are only active when a document has been saved at least once.
- **Save** (`usecase/document.Save`): upserts the `draft` record's `data`, `updatedAt`, and `updatedBy` (current user). Never touches the `published` record. Never returned by the public read API.
- **Publish** (`usecase/document.Publish`): copies `draft.data` → `published` record (creating it if absent), sets `published.updatedAt = draft.updatedAt`, `published.publishedAt = now()`, and `published.publishedBy = current user`.
- Entry `status` is **computed, never stored**: `draft` (no published record exists), `modified` (`draft.updatedAt > published.updatedAt`), `published` (timestamps match).
- The public/content read API resolves an entry to its `published` record only. If no `published` record exists, the entry does not exist from the API's perspective (404) — the draft is invisible to readers, however recent.
- Admin edit screens read the `draft` record (so editors always see their latest unpublished work) plus the computed `status` for a draft/modified/published badge.
- Applies uniformly to **every** content type — there is no per-type opt-out in v1.

### Domain Rules: Content Type Kinds
- `ContentType.kind: "single" | "collection"`.
- **Single-type**: at most one entry per content type. **No auto-created singleton** — the entry is created on the user's first explicit Save. The UI shows an empty form until then; Publish/Unpublish buttons are disabled until a document exists. No create/delete UI — only edit, Save (draft), and Publish.
- **Collection-type**: zero or more entries, each with its own `documentId`. Standard list + create/edit/delete, each entry carrying its own independent draft/published pair and status.
- The admin sidebar groups navigation into two sections — **Single Types** and **Collection Types** — driven by `ContentType.kind`, not by folder location.

### Domain Rules: Content-Type Schema as Code
- Content-type **structure** (fields, types, `kind`) is defined in JSON files under `apps/api/content-types/*.json` — never created or edited via the API or UI. The UI only manages content **data** (entries).
- On every API startup, `usecase/content_type.Sync` reads all definition files and reconciles them against the `ContentType` records in MongoDB:
  - **New file** → create the `ContentType` and its per-content-type document collection (`documents_<slug>`) with indexes.
  - **Changed file** (fields added/changed) → update the `ContentType` record's schema in place.
  - **Field removed from a file** → drop the field from the `ContentType` schema, but leave already-stored entry data untouched (the orphaned key stays in MongoDB, simply no longer shown or editable).
  - **File deleted** → delete the `ContentType`, cascade-delete all its entries (draft + published), and drop the per-content-type collection.
- Sync is one-directional: JSON definitions are always the source of truth; nothing the UI or API does ever writes back to the definition files.
- JSON schema files declare only content fields. The system fields (`createdAt`, `updatedAt`, `publishedAt`, `createdBy`, `updatedBy`, `publishedBy`, `locale` — see Draft & Publish rules above) are never declared in the schema; they're injected automatically on every record.
- **This restriction applies only to content-type *structure*, never to content *data*.** Creating, updating, deleting, saving, and publishing entries (documents) stays fully available through the existing API and UI for every content type — schema-as-code only removes the ability to create/edit/delete the *type itself* (its name, slug, kind, field list).

---

## 5. Testing Strategy

### Backend
- **Unit tests**: Every usecase tested in isolation with mock repositories (interface-based). Located in `internal/usecase/<name>/<name>_test.go`.
- **Draft/Publish tests**: cover all three computed statuses (`draft`, `modified`, `published`), Save never mutating the published record, Publish syncing `updatedAt`/`publishedAt`, and the public read API returning 404 for entries with no published record yet.
- **Single-type tests**: single-type content types do not auto-create a singleton entry; documents are only created on explicit save. Create/delete operations are rejected for single-type content.
- **Schema sync tests**: new definition file creates a `ContentType` and its per-content-type document collection; changed file updates its schema in place; removed field drops from schema but leaves stored entry data untouched; deleted file cascades-deletes the `ContentType`, all its entries, and drops the per-content-type collection. Run against a mock repository, not real files, except for one integration test reading actual fixture JSON files.
- **Integration tests**: MongoDB repository implementations tested against a real MongoDB instance spun up via Docker in CI. Located in `internal/infrastructure/mongodb/`.
- **HTTP handler tests**: Use `httptest` to test handlers with mock usecases.
- **Coverage target**: ≥ 80% on `internal/usecase/` packages.
- **No mocking the DB in usecase tests** — mock the repository interface instead.

### Frontend
- **Component tests**: Vitest + React Testing Library for `FormProvider`, `FormField`, and all input components.
- **Form integration tests**: Render a minimal panel with `FormProvider` + `FormField` + input, submit, assert the resulting JSON shape matches the dot-notation contract.
- **Query/mutation tests**: Wrap components in a test `QueryClientProvider`; use `msw` (Mock Service Worker) to intercept API calls. Assert that after a successful mutation, the relevant query key is invalidated and refetched.
- **No E2E tests in v1** (deferred).

### CI (GitHub Actions)
```
on: push to main / pull_request
jobs:
  api: go vet → go test ./...
  web: npm run lint → npm run build
  docker: build both images (smoke test only)
  deploy: trigger Render deploy hook (main branch only)
```

---

## 6. Boundaries

### Always (automatic, no confirmation needed)
- `FormProvider` automatically manages `loading` state during fetch and submit.
- `FormProvider` automatically binds fetched data into form inputs as initial values.
- Dot-notation `name` attributes are automatically deserialized into nested JSON on form submit.
- Deleting any entity cascades deletion of all child sub-objects and related entities (usecase layer).
- JWT access token is automatically refreshed via the refresh token before expiry.
- Every `useMutation` that writes data must call `queryClient.invalidateQueries` on success — the UI always reflects the latest server state after any change.
- Save always writes the draft record only; it never publishes implicitly.
- A single-type content type starts with no document; the first Save creates it. Publish/Unpublish are only available after a document exists.
- Each content type's documents are stored in a standalone MongoDB collection (`documents_<slug>`), created during schema sync on API startup.
- Entry `status` (draft/modified/published) is always computed from timestamps, never persisted as a stored field.
- Content-type schema sync runs automatically on every API startup — no manual trigger needed. Sync also ensures per-content-type collections exist with proper indexes.
- `createdAt`, `updatedAt`, `publishedAt`, `createdBy`, `updatedBy`, `publishedBy`, and `locale` are injected automatically on every record — never authored by editors or declared in a content-type's JSON schema.

### Ask before (require explicit approval)
- Starting any new coding task — confirm scope and approach with the user first.
- Any feature or implementation with multiple valid approaches — present options, do not choose autonomously.
- `MediaInput` opens the OS file picker before uploading to storage.
- Choosing which storage adapter (S3 vs Cloudinary) is active per environment/deployment — confirm with the user before changing the default.

### Never
- Never read, edit, create, or expose `.env` or any environment variable files.
- Never use a drag-and-drop or dynamic form engine — panels are hard-coded React pages only.
- Never use `React.Children.map` or recursive child scanning in `FormProvider`.
- Never couple usecase logic to a specific database — all DB access is behind repository interfaces.
- Never auto-choose an implementation path when multiple options exist — always ask.
- Never add GraphQL or PostgreSQL support until REST + MongoDB are fully complete and the user authorizes the next phase.
- Never let the public/content read API return draft data — it only ever resolves the `published` record.
- Never expose delete actions for single-type content in the UI or API — only edit, Save, and Publish. The first Save implicitly creates the document.
- Never add an API or UI path to create, edit, or delete a `ContentType`'s **structure** — structure changes only ever come from editing JSON definition files and restarting/syncing. (Content **data**/entries are unaffected by this rule — see below.)
- Never let content-type sync write back to the JSON definition files — sync is one-directional (files → DB).
- Never remove or restrict create/update/delete/save/publish on content **data** (documents/entries) — every content type, single or collection, keeps full data CRUD via the API and UI. Only the type's *structure* is JSON-only.

---

## Resolved Decisions

1. **Media storage**: Support both AWS S3 and Cloudinary from day one, behind the storage interface (`internal/infrastructure/storage/`), selectable via config/env.
2. **Go module path**: `project-abyssoftime-cms-v2/api`.
3. **Render deployment**: Single Render service running `docker-compose up` (the full stack as one service, not split per Docker service).

---

## 7. Web — Content-Type Management System (Refactor)

**Scope**: Web layer only. API is untouched.

### 7.1 Objective

Refactor the web admin panel's content-type handling to:
- Enforce a strict form lifecycle: dirty-tracking, toast notifications on save success/failure, and post-save form reset to server state
- Introduce `ContentTypeLayout` — a render-prop wrapper that handles the standard layout shell, with `renderHeader` and `renderActions` escape hatches for custom UI
- Replace ad-hoc column display in collection-type tables with a React-side column registry
- Restructure routes under `/admin` to distinguish single-type from collection-type paths
- Sidebar loads only metadata eagerly; component source code is loaded on demand (React.lazy), never at sidebar mount time
- Locale switching is in scope for both `SingleTypePage` and `CollectionDetailPage`

---

### 7.2 Routing Structure

All routes remain under `/admin`. Sub-paths are restructured:

| Route | Component | Description |
|---|---|---|
| `/admin/content-type/single-type/:slug` | `SingleTypePage` | Single-type edit form |
| `/admin/content-type/collection-type/:slug` | `CollectionListPage` | Collection-type list + "Add new item" |
| `/admin/content-type/collection-type/:slug/:id` | `CollectionDetailPage` | Collection-type item edit form |

Existing `/admin/content-types/:slug` and `/admin/content-types/:slug/:id` routes are removed.

---

### 7.3 Components

#### `ContentTypeLayout`

Standard layout shell for any content-type form page. Exported from `src/components/content-type/ContentTypeLayout.tsx`.

```ts
interface ContentTypeLayoutProps {
  title: string
  status?: string
  // Replaces the entire header section if provided
  renderHeader?: (defaultHeader: ReactNode) => ReactNode
  // Appends action buttons (Publish, Unpublish, locale select, Save) to the right of the header
  renderActions?: () => ReactNode
  children: ReactNode
}
```

Default render: `title` + status badge on the left; `renderActions()` slot on the right. `renderHeader` overrides the entire header row when provided.

---

#### Content-Type Registry

A metadata-only module at `src/content-type-registry/index.ts`. **No component imports here.**

```ts
interface CollectionColumnDef {
  key: string
  label: string
  type: 'text' | 'boolean' | 'number' | 'image'
}

interface ContentTypeRegistration {
  slug: string
  kind: 'single' | 'collection'
  // For collection types: defines which columns appear in the list table
  columns?: CollectionColumnDef[]
  // Optional custom layout wrapper; default ContentTypeLayout used if omitted
  wrapper?: React.ComponentType<ContentTypeLayoutProps>
}

export const contentTypeRegistry: ContentTypeRegistration[]
```

The sidebar reads `slug`, `kind`, and `name` (from the API) and uses the registry to resolve column definitions and custom wrappers. No component code is bundled at the registry level.

---

#### `FormProvider` — Enhanced Lifecycle

`FormProvider` is updated to cover the full form lifecycle:

| Moment | Behaviour |
|---|---|
| Initial load | All fields rendered; pre-filled from server data if available, empty otherwise |
| Clean state | `Save` button disabled (`isDirty === false`) |
| After any edit | `Save` button enabled (`isDirty === true`) |
| Failed save | `toast.error(serverMessage)` — form stays in edited state |
| Successful save | `toast.success('Saved')` → `queryClient.invalidateQueries(queryKey)` → `reset(newServerData)` → Save disabled again |

`isDirty` is exposed on `FormStateContext` so action slots can read it:

```ts
interface FormState {
  loading: boolean
  submitting: boolean
  isDirty: boolean
}
```

---

### 7.4 Single-Type Page (`SingleTypePage`)

Replaces `SingleTypePanel`. Responsibilities:
- Fetches document via `useDocuments(contentType.Slug)`
- When no document exists (first visit), renders an empty form based on content-type fields; first Save creates the document
- Maintains local `locale` state (defaults to first locale from `useLocales()`)
- When `useLocales()` returns more than one locale, renders a locale `<select>` in `renderActions`
- Switching locale resets the form to the new locale's document data; `isDirty` becomes `false`
- Wraps form in `ContentTypeLayout` — passes `renderActions` for locale selector + Publish/Unpublish buttons
- Uses registry `wrapper` override if registered for the slug; falls back to `ContentTypeLayout`
- Delegates all form state to `FormProvider`; does not own `isDirty` locally

---

### 7.5 Collection-Type List Page (`CollectionListPage`)

- Renders a `<table>` driven by `columns` from the registry for the resolved slug
- **Fallback**: if no registry entry defines `columns`, display the first field as a text column + Status column
- Column type rendering:
  - `text` → string value
  - `boolean` → `✓` / `—`
  - `number` → numeric string
  - `image` → `<img>` thumbnail (src = field value)
- "Add new item": creates document (`useCreateDocument`) → navigate to `/admin/content-type/collection-type/:slug/:id`
- Per-row: Edit link → detail page; Delete button → `useDeleteDocument` (with `window.confirm` guard)
- No locale switching on the list view (locale is only relevant on the detail/edit form)

---

### 7.6 Collection-Type Detail Page (`CollectionDetailPage`)

- Navigates by URL (`/admin/content-type/collection-type/:slug/:id`)
- Maintains local `locale` state identical to `SingleTypePage`
- When `useLocales()` returns more than one locale, renders a locale `<select>` in `renderActions`
- Switching locale reloads the document for the new locale; form resets and `isDirty` becomes `false`
- Save mutation sends `{ data, locale: activeLocale }` — identical shape to existing `useUpdateDocument`
- Back link returns to `/admin/content-type/collection-type/:slug`
- `ContentTypeLayout` used with `renderActions` for locale selector + Publish/Unpublish

---

### 7.7 Locale Switching — Shared Process

Applies identically to `SingleTypePage` and `CollectionDetailPage`:

1. `useLocales()` fetches available locales on mount
2. Local `locale` state initialises to `locales[0]` (or `''` before locales load)
3. `<select aria-label="Locale">` rendered in `renderActions` only when `locales.length > 1`
4. On locale change: update local state → React Query re-fetches the document for the new locale → `FormProvider`'s `values` prop (from `useForm({ values: ... })`) syncs the form inputs → `isDirty` resets to `false` automatically
5. Publish/Unpublish/Save mutations always forward `locale: activeLocale`

---

### 7.8 Sidebar (Lazy Loading)

`Sidebar`:
- Eagerly loads `useContentTypes()` (metadata only: `Name`, `Slug`, `Kind`)
- Groups into **Single Types** / **Collection Types** nav sections
- Generates `NavLink` hrefs:
  - single → `/admin/content-type/single-type/:slug`
  - collection → `/admin/content-type/collection-type/:slug`
- **Never imports component source at sidebar mount** — component code is loaded via `React.lazy` when the router renders the target route

---

### 7.9 Acceptance Criteria

**Single-type form:**
- [ ] Navigating to `/admin/content-type/single-type/:slug` renders all fields, pre-filled from draft data
- [ ] Save disabled on initial load; enabled after any field edit
- [ ] Successful save: success toast + form reset to new server data + Save disabled again
- [ ] Failed save: error toast with server message; edited values preserved
- [ ] Locale selector shown only when API returns > 1 locale
- [ ] Switching locale resets form to new locale's data; isDirty becomes false

**Collection-type list:**
- [ ] Table columns come from the registry for the slug; first-field fallback applies when absent
- [ ] `text`, `boolean`, `number`, `image` column types render correctly
- [ ] "Add new item" creates a document and navigates to the detail page
- [ ] Edit navigates to detail page; Delete removes with confirm guard

**Collection-type detail:**
- [ ] Full single-type lifecycle applies (dirty-tracking, toasts, post-save reset)
- [ ] Locale selector shown/hidden under same condition as single-type
- [ ] Switching locale resets form to new locale's data
- [ ] Back link navigates to the list page

**Sidebar:**
- [ ] Groups by kind; links use the new route structure
- [ ] Component source is never imported at sidebar mount time
- [ ] Renders gracefully when API returns an empty list

**Routing:**
- [ ] Old `/admin/content-types/...` routes removed
- [ ] All new routes require authentication (existing `ProtectedRoute` guard)

---

### 7.10 Out of Scope

- Any API changes
- New input types beyond existing six (`TextInput`, `BooleanInput`, `NumberInput`, `MediaInput`, `RichTextInput`, `JsonInput`)
- Publish/Unpublish mutation logic (only layout wrapper changes; mutations are unchanged)

---

## 8. Delete Media Asset in MediaLibrary

### 8.1 Objective

Allow CMS admins to delete a media asset from the MediaLibrary grid. Deletion removes the record from MongoDB **and** the file from the Cloudinary storage provider. The UI provides a single-asset, inline-confirm flow — a trash icon visible on tile hover, requiring a second click to confirm before the delete is sent.

Target users: CMS administrators managing media assets.

---

### 8.2 API Contract

```
DELETE /api/media/{id}
Authorization: Bearer <access_token>
```

| Status | Body | Condition |
|---|---|---|
| `204 No Content` | — | Successfully deleted from storage and DB |
| `404 Not Found` | `{"error": "not found"}` | Asset ID does not exist |
| `500 Internal Server Error` | `{"error": "..."}` | Storage or DB failure |

---

### 8.3 Files Changed

**Backend — modify only, no new files:**

```
apps/api/internal/usecase/media/
  media_usecase.go          ← add Delete(ctx, id string) error method
  media_usecase_test.go     ← add Delete test cases

apps/api/internal/delivery/http/handler/
  media_handler.go          ← extend mediaUseCase interface + add Delete handler + route
  media_handler_test.go     ← add Delete handler test cases
```

Both `MediaAssetRepository.Delete(ctx, id)` and `StorageAdapter.Delete(ctx, publicID)` already exist — no changes to entity, repository interface, or infrastructure layer.

**Frontend — modify only, no new files:**

```
apps/web/src/hooks/
  useMedia.ts               ← add useDeleteMedia() mutation

apps/web/src/components/media/
  MediaLibrary.tsx          ← add hover trash icon + inline confirm UX
  __tests__/MediaLibrary.test.tsx  ← add delete interaction tests
```

---

### 8.4 Delete Flow

#### UseCase (`media_usecase.go`)

```
Delete(ctx, id):
  1. asset ← assetRepo.FindByID(ctx, id)    // propagate not-found as-is
  2. storage.Delete(ctx, asset.PublicID)     // remove from Cloudinary
  3. assetRepo.Delete(ctx, id)              // remove DB record
  return error
```

If `storage.Delete` fails, do **not** call `assetRepo.Delete` — storage is the source of truth; orphaned DB records are harder to clean up than orphaned Cloudinary files.

#### HTTP Handler (`media_handler.go`)

Extend the `mediaUseCase` interface with:
```go
Delete(ctx context.Context, id string) error
```

Handler method:
```
DELETE /api/media/{id}
  id ← r.PathValue("id")
  err ← h.uc.Delete(r.Context(), id)
  nil  → 204 No Content
  err  → writeErr(w, err)
```

#### Frontend Hook (`useMedia.ts`)

```ts
export function useDeleteMedia() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.delete(`/api/media/${id}`),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['media', 'list'] }),
    onError: (err: unknown) => {
      const msg = (err as AxiosError<{ error: string }>).response?.data?.error ?? 'Delete failed'
      toast.error(msg)
    },
  })
}
```

#### Frontend UI (`MediaLibrary.tsx`)

New local state: `pendingDeleteId: string | null = null`

Per-tile behavior:
- On hover: show trash icon button (absolute top-right corner)
- Trash icon — first click: `setPendingDeleteId(asset.ID)` (arm)
- Trash icon — second click (when `pendingDeleteId === asset.ID`): call `deleteMedia(asset.ID)` (fire)
- `onMouseLeave` on the tile: `setPendingDeleteId(null)` (disarm)
- Armed state: trash icon turns red to signal one more click confirms
- While mutation `isPending`: disable the tile

---

### 8.5 Testing

#### Backend unit tests

**UseCase — `media_usecase_test.go`:**
- `TestDelete_CallsStorageAndRepo` — FindByID returns asset, storage.Delete called with correct PublicID, assetRepo.Delete called with id
- `TestDelete_AssetNotFound_ReturnsError` — FindByID returns not-found error; storage.Delete never called
- `TestDelete_StorageError_DoesNotDeleteFromRepo` — storage.Delete fails; assetRepo.Delete never called
- `TestDelete_RepoDeleteError_ReturnsError` — storage succeeds, assetRepo.Delete fails; error propagated

**Handler — `media_handler_test.go`:**
- `TestMediaHandler_Delete_Returns204` — mock usecase Delete returns nil → expect 204
- `TestMediaHandler_Delete_NotFound_Returns404` — mock returns not-found sentinel → expect 404
- `TestMediaHandler_Delete_UseCaseError_Returns500` — mock returns generic error → expect 500

#### Frontend unit tests (`MediaLibrary.test.tsx`)

- Trash icon is not visible without hover; appears on tile hover
- First click arms the confirm state (icon turns red) but does not call the delete API
- Second click on the armed tile fires the delete mutation
- `mouseLeave` on a tile disarms confirm state
- Tile is disabled/non-interactive while delete mutation `isPending`

---

### 8.6 Boundaries

| Rule | Detail |
|---|---|
| **Always** | Call `storage.Delete` before `assetRepo.Delete` |
| **Always** | Return 404 (not 500) when asset ID is not found |
| **Always** | Invalidate `['media', 'list']` query on successful delete |
| **Never** | Bulk-delete — single asset at a time only |
| **Never** | A confirmation modal — inline hover-confirm is the specified UX |
| **Never** | Skip storage delete (no DB-only or soft-delete removal) |
| **Ask first** | Cascade-deleting assets referenced by documents touches `DeleteByDocumentRef` — out of scope here |

---

## 9. Security Hardening

### 9.1 Objective

Remediate vulnerabilities introduced by the per-content-type collection refactor and close pre-existing security gaps across the full stack. All fixes must ship before production deployment.

Target: CMS administrators and public-facing content API consumers.

---

### 9.2 Findings Summary

| # | Finding | Severity | Category |
|---|---------|----------|----------|
| S1 | Unsanitized slug used as MongoDB collection name | Critical | NoSQL injection |
| S2 | No slug format validation on content-type creation | Critical | Input validation |
| S3 | No documentID format validation in handlers | Critical | Input validation |
| S4 | Missing `Secure` flag on refresh token cookie | High | Cookie security |
| S5 | No CORS middleware | High | Cross-origin |
| S6 | No password strength validation | High | Authentication |
| S7 | No email format validation | High | Input validation |
| S8 | No rate limiting on auth endpoints | Medium | Brute force |
| S9 | No request body size limit | Medium | DoS |
| S10 | No security response headers | Medium | Transport |

---

### 9.3 Implementation Plan

#### S1 + S2: Slug Validation

**File:** `apps/api/internal/usecase/content_type/content_type_usecase.go`

Add `validateSlug` — called in `Create` before any DB operation. Pattern: `^[a-z0-9]+(?:-[a-z0-9]+)*$` (lowercase alphanumeric + hyphens, no leading/trailing hyphens, 1–63 chars).

```go
var slugRe = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func validateSlug(slug string) error {
    if len(slug) == 0 || len(slug) > 63 || !slugRe.MatchString(slug) {
        return fmt.Errorf("%w: slug must be 1-63 lowercase alphanumeric chars or hyphens", pkgerrors.ErrValidation)
    }
    return nil
}
```

**File:** `apps/api/internal/delivery/http/handler/document_handler.go`

Add a `validateSlug` guard at the handler level for every method that reads `r.PathValue("slug")`. Return 400 if invalid. This is defense-in-depth — the usecase validates on creation, the handler validates on every request.

---

#### S3: DocumentID Validation

**File:** `apps/api/internal/delivery/http/handler/document_handler.go`

Add a helper to validate documentID is a 24-character hex string (MongoDB ObjectID format):

```go
var objectIDRe = regexp.MustCompile(`^[a-f0-9]{24}$`)

func validObjectIDHex(id string) bool {
    return objectIDRe.MatchString(id)
}
```

Apply in `GetByID`, `Update`, `Delete`, `Publish`, `Unpublish`, `GetPublic`. Return 400 for invalid format.

---

#### S4: Secure Cookie Flag

**File:** `apps/api/internal/delivery/http/handler/auth_handler.go`

Add `Secure: true` to the refresh-token cookie in both `Login` and `Refresh` handlers. Controlled by a config flag `cfg.CookieSecure` (default `true`, set `false` for local HTTP dev):

```go
http.SetCookie(w, &http.Cookie{
    Name:     RefreshCookieName,
    Value:    refresh,
    HttpOnly: true,
    Secure:   h.cookieSecure,
    SameSite: http.SameSiteLaxMode,
    MaxAge:   refreshCookieMaxAge,
    Path:     "/",
})
```

**File:** `apps/api/internal/config/config.go` — add `CookieSecure bool` field (env `COOKIE_SECURE`, default `true`).

---

#### S5: CORS Middleware

**File:** `apps/api/internal/delivery/http/middleware/cors.go` (new)

```go
func CORS(allowedOrigins []string) func(http.Handler) http.Handler
```

- Checks `Origin` header against the whitelist.
- Sets `Access-Control-Allow-Origin`, `Access-Control-Allow-Methods`, `Access-Control-Allow-Headers`, `Access-Control-Allow-Credentials`.
- Handles `OPTIONS` preflight with 204.
- If no origin match, omit CORS headers (browser blocks the request).

**File:** `apps/api/internal/config/config.go` — add `CORSOrigins []string` (env `CORS_ORIGINS`, comma-separated, default `http://localhost:5173`).

**File:** `apps/api/cmd/server/main.go` — wrap `mux` with `middleware.CORS(cfg.CORSOrigins)`.

---

#### S6: Password Validation

**File:** `apps/api/internal/usecase/auth/auth_usecase.go`

Add validation in `Register`:
- Minimum 8 characters
- Maximum 72 characters (bcrypt limit)
- At least one letter and one digit

```go
func validatePassword(password string) error {
    if len(password) < 8 || len(password) > 72 {
        return fmt.Errorf("%w: password must be 8-72 characters", pkgerrors.ErrValidation)
    }
    hasLetter := false
    hasDigit := false
    for _, r := range password {
        if unicode.IsLetter(r) { hasLetter = true }
        if unicode.IsDigit(r) { hasDigit = true }
    }
    if !hasLetter || !hasDigit {
        return fmt.Errorf("%w: password must contain at least one letter and one digit", pkgerrors.ErrValidation)
    }
    return nil
}
```

---

#### S7: Email Validation

**File:** `apps/api/internal/usecase/auth/auth_usecase.go`

Add validation in `Register` using `net/mail.ParseAddress`:

```go
func validateEmail(email string) error {
    if len(email) > 254 {
        return fmt.Errorf("%w: email too long", pkgerrors.ErrValidation)
    }
    if _, err := mail.ParseAddress(email); err != nil {
        return fmt.Errorf("%w: invalid email format", pkgerrors.ErrValidation)
    }
    return nil
}
```

---

#### S8: Rate Limiting

**File:** `apps/api/internal/delivery/http/middleware/ratelimit.go` (new)

In-memory token-bucket rate limiter using `golang.org/x/time/rate`. Keyed by client IP.

```go
func RateLimit(rps float64, burst int) func(http.Handler) http.Handler
```

Applied to auth endpoints only: `/auth/login`, `/auth/register`, `/auth/refresh`.

Config: `RATE_LIMIT_RPS` (default `5`), `RATE_LIMIT_BURST` (default `10`).

---

#### S9: Request Body Size Limit

**File:** `apps/api/internal/delivery/http/middleware/bodylimit.go` (new)

```go
func BodyLimit(maxBytes int64) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
            next.ServeHTTP(w, r)
        })
    }
}
```

Applied globally with 10 MB default. Media upload exempt (uses its own multipart limit).

---

#### S10: Security Response Headers

**File:** `apps/api/internal/delivery/http/middleware/security_headers.go` (new)

```go
func SecurityHeaders(next http.Handler) http.Handler
```

Sets on every response:
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 0` (modern CSP preferred over legacy header)
- `Referrer-Policy: strict-origin-when-cross-origin`
- `Content-Security-Policy: default-src 'self'` (configurable for CDN assets)

---

### 9.4 Files Changed

**New files:**
```
apps/api/internal/delivery/http/middleware/cors.go
apps/api/internal/delivery/http/middleware/cors_test.go
apps/api/internal/delivery/http/middleware/ratelimit.go
apps/api/internal/delivery/http/middleware/ratelimit_test.go
apps/api/internal/delivery/http/middleware/bodylimit.go
apps/api/internal/delivery/http/middleware/bodylimit_test.go
apps/api/internal/delivery/http/middleware/security_headers.go
apps/api/internal/delivery/http/middleware/security_headers_test.go
```

**Modified files:**
```
apps/api/internal/usecase/content_type/content_type_usecase.go      ← slug validation
apps/api/internal/usecase/content_type/content_type_usecase_test.go  ← slug validation tests
apps/api/internal/usecase/auth/auth_usecase.go                      ← email + password validation
apps/api/internal/usecase/auth/auth_usecase_test.go                 ← validation tests
apps/api/internal/delivery/http/handler/document_handler.go         ← slug + documentID guards
apps/api/internal/delivery/http/handler/document_handler_test.go    ← guard tests
apps/api/internal/delivery/http/handler/auth_handler.go             ← Secure cookie flag
apps/api/internal/delivery/http/handler/auth_handler_test.go        ← cookie tests
apps/api/internal/config/config.go                                  ← new config fields
apps/api/cmd/server/main.go                                         ← middleware wiring
SPEC.md                                                             ← this section
```

---

### 9.5 Testing

**Slug validation (`content_type_usecase_test.go`):**
- Valid slugs accepted: `blog-post`, `homepage`, `a`, `my-content-123`
- Invalid slugs rejected: `""`, `Blog Post`, `../admin`, `system.users`, `$cmd`, `-leading`, `trailing-`, `a`.repeat(64)

**DocumentID validation (`document_handler_test.go`):**
- Valid: 24 hex chars → 200
- Invalid: `abc`, `../../../etc/passwd`, `$where`, empty → 400

**Password validation (`auth_usecase_test.go`):**
- Rejected: `""`, `"short"`, `"noletter1234"`, `"nodigithere"`, `"a".repeat(73)`
- Accepted: `"Password1"`, `"my-s3cure-pass!"`

**Email validation (`auth_usecase_test.go`):**
- Rejected: `""`, `"notanemail"`, `"@@@"`, `"a".repeat(255)+"@x.com"`
- Accepted: `"user@example.com"`, `"name+tag@domain.co"`

**CORS (`cors_test.go`):**
- Allowed origin → correct headers set
- Disallowed origin → no CORS headers
- OPTIONS preflight → 204

**Rate limiting (`ratelimit_test.go`):**
- Under limit → 200
- Over limit → 429

**Body limit (`bodylimit_test.go`):**
- Under limit → 200
- Over limit → 413

**Security headers (`security_headers_test.go`):**
- All headers present on every response

---

### 9.6 Config Additions

| Variable | Description | Default |
|----------|-------------|---------|
| `COOKIE_SECURE` | Set `Secure` flag on refresh token cookie | `true` |
| `CORS_ORIGINS` | Comma-separated allowed origins | `http://localhost:5173` |
| `RATE_LIMIT_RPS` | Auth endpoint requests per second per IP | `5` |
| `RATE_LIMIT_BURST` | Auth endpoint burst size | `10` |
| `BODY_LIMIT_BYTES` | Max request body size (bytes) | `10485760` (10 MB) |

---

### 9.7 Out of Scope

- Token blacklist / refresh token invalidation (requires Redis or DB-backed session store)
- Audit logging (separate feature)
- Content Security Policy tuning for CKEditor / CDN assets
- Frontend XSS — React auto-escapes JSX expressions; CKEditor sanitizes by default. No `dangerouslySetInnerHTML` found in codebase.
- HTTPS enforcement (handled by reverse proxy / Render.com)

---

### 9.8 Boundaries

| Rule | Detail |
|---|---|
| **Always** | Validate slug format at both usecase (creation) and handler (every request) levels |
| **Always** | Validate documentID is 24-char hex before passing to usecase |
| **Always** | Set `Secure: true` on cookies in production (`COOKIE_SECURE=true`) |
| **Always** | Return 400 (not 500) for invalid slug or documentID format |
| **Always** | Apply body size limit globally; media upload uses its own multipart limit |
| **Never** | Allow slug characters outside `[a-z0-9-]` |
| **Never** | Allow password shorter than 8 characters or longer than 72 |
| **Never** | Reflect user input in error messages verbatim (prevents information leakage) |
| **Never** | Set `Access-Control-Allow-Origin: *` — always use explicit origin whitelist |
| **Ask first** | Adjusting rate limit thresholds for production traffic patterns |

---

## 10. Document Manager API Restructure & Paginated Collections

### 10.1 Objective

Restructure the document manager API to explicitly separate single-type and collection-type routes, add server-side pagination for collection lists, and introduce field projection so collection list responses only include the fields needed for table display.

**Why:** The current API uses a flat `/api/document-manager/{slug}` prefix for all document operations regardless of content-type kind. This creates several problems:
- Collection lists return **all** documents with no pagination — untenable as content grows
- Single-type endpoints return an array when only one document can exist
- The FE fetches full document data for list views when only a few fields are needed for the table
- No API-level enforcement of single-type vs collection-type semantics

**Target users:** CMS administrators (both single-type content editors and collection-type content managers).

---

### 10.2 Content-Type Schema Change (`listFields`)

Add a `listFields` property to content-type JSON definitions. This array specifies which document data fields the collection list API response includes — the BE uses it for field projection.

**Before:**
```json
{
  "slug": "blog-posts",
  "name": "Blog Posts",
  "kind": "collection",
  "fields": [
    { "name": "title", "type": "text" },
    { "name": "slug", "type": "text" },
    { "name": "coverImage", "type": "media" },
    { "name": "excerpt", "type": "text" },
    { "name": "body", "type": "richtext" },
    { "name": "readingTime", "type": "number" },
    { "name": "featured", "type": "boolean" },
    { "name": "metadata", "type": "json" }
  ]
}
```

**After:**
```json
{
  "slug": "blog-posts",
  "name": "Blog Posts",
  "kind": "collection",
  "listFields": ["title", "slug", "featured"],
  "fields": [
    { "name": "title", "type": "text" },
    { "name": "slug", "type": "text" },
    { "name": "coverImage", "type": "media" },
    { "name": "excerpt", "type": "text" },
    { "name": "body", "type": "richtext" },
    { "name": "readingTime", "type": "number" },
    { "name": "featured", "type": "boolean" },
    { "name": "metadata", "type": "json" }
  ]
}
```

**Rules:**
- `listFields` is optional. If omitted, defaults to the first 3 field names from `fields`.
- Only meaningful for `kind: "collection"` — ignored for single-type.
- Each entry must reference a `name` that exists in the `fields` array. The schema loader validates this on startup and fatally logs any mismatch.
- Synced into the `ContentType` entity's `ListFields` field during schema-as-code sync on API startup.

**Entity change (`entity/content_type.go`):**
```go
type ContentType struct {
    // ... existing fields unchanged
    ListFields []string `json:"listFields,omitempty" bson:"listFields,omitempty"`
}
```

**Schema loader change (`usecase/content_type/schema_loader.go`):**
```go
type ContentTypeDefinition struct {
    Slug       string                   `json:"slug"`
    Name       string                   `json:"name"`
    Kind       string                   `json:"kind"`
    ListFields []string                 `json:"listFields,omitempty"`
    Fields     []entity.FieldDefinition `json:"fields"`
}
```

Validation: each `listFields` entry must match a `fields[].name`. Error on mismatch.

---

### 10.3 API Routes — Single-Type

All single-type document operations use `/api/document-manager/single-type/{slug}`. No `documentId` in the URL — slug + locale uniquely identifies the singleton document.

| Method | Route | Query Params | Response | Description |
|--------|-------|-------------|----------|-------------|
| `GET` | `/api/document-manager/single-type/{slug}` | `locale` | `Document` or `404` | Fetch single-type draft document |
| `PUT` | `/api/document-manager/single-type/{slug}` | `locale` | `Document` | Save (create-if-first, update-if-exists) |
| `POST` | `/api/document-manager/single-type/{slug}/publish` | `locale` | `{"status":"published"}` | Publish |
| `POST` | `/api/document-manager/single-type/{slug}/unpublish` | `locale` | `{"status":"draft"}` | Unpublish |

**GET behavior:**
- Finds the single draft document matching slug + locale.
- If no document exists → **404 Not Found**. The FE interprets 404 as "show empty form".
- If document exists → returns full document with computed `status`.

**PUT behavior:**
- If no document exists for slug + locale → creates a new document (auto-generates `documentId`).
- If document already exists → updates its `data` (identical to current Save semantics — draft only, never touches published).
- Always returns the saved document with computed `status`.

**Publish / Unpublish:**
- Finds the single document by slug + locale.
- Delegates to existing publish/unpublish logic.
- Returns 404 if no document exists to publish/unpublish.

**No DELETE for single-type** — consistent with existing SPEC §6 boundary ("never expose delete actions for single-type content").

---

### 10.4 API Routes — Collection-Type

| Method | Route | Query Params | Response | Description |
|--------|-------|-------------|----------|-------------|
| `GET` | `/api/document-manager/collection-type/{slug}` | `start`, `size`, `locale` | `PaginatedList` | Paginated list (projected fields) |
| `GET` | `/api/document-manager/collection-type/{slug}/{documentId}` | `locale` | `Document` | Fetch single document by ID |
| `POST` | `/api/document-manager/collection-type/{slug}` | `locale` | `Document` | Create new document |
| `PUT` | `/api/document-manager/collection-type/{slug}/{documentId}` | `locale` | `Document` | Update document by ID |
| `DELETE` | `/api/document-manager/collection-type/{slug}/{documentId}` | — | `204` | Delete (cascade) |
| `POST` | `/api/document-manager/collection-type/{slug}/{documentId}/publish` | `locale` | `{"status":"published"}` | Publish |
| `POST` | `/api/document-manager/collection-type/{slug}/{documentId}/unpublish` | `locale` | `{"status":"draft"}` | Unpublish |

**Paginated list response format:**
```json
{
  "items": [
    {
      "documentId": "683abc...",
      "data": { "title": "My Post", "slug": "my-post", "featured": true },
      "status": "draft",
      "locale": "en",
      "createdAt": "2026-06-01T00:00:00Z",
      "updatedAt": "2026-06-15T12:00:00Z"
    }
  ],
  "total": 42,
  "start": 0,
  "size": 20
}
```

**Pagination parameters:**
| Param | Default | Max | Description |
|-------|---------|-----|-------------|
| `start` | `0` | — | Offset from beginning of results |
| `size` | `20` | `100` | Number of items per page |
| `locale` | first supported locale | — | Filter documents by locale |

**Field projection:**
- `items[].data` contains **only** the fields specified in the content-type's `listFields`.
- `documentId`, `status` (computed), `locale`, `createdAt`, `updatedAt` are always included.
- Full `data` is available only via the single-document GET endpoint (`/collection-type/{slug}/{documentId}`).

**Status computation:**
- For each draft in the page, the handler batches a lookup of corresponding published records to compute status efficiently (avoids N+1 queries).

---

### 10.5 Removed Routes

The following flat routes are **removed** — all callers must migrate to the kind-prefixed routes above:

```
GET    /api/document-manager/{slug}                          → REMOVED
GET    /api/document-manager/{slug}/{documentId}             → REMOVED
POST   /api/document-manager/{slug}                          → REMOVED
PUT    /api/document-manager/{slug}/{documentId}             → REMOVED
DELETE /api/document-manager/{slug}/{documentId}             → REMOVED
POST   /api/document-manager/{slug}/{documentId}/publish     → REMOVED
POST   /api/document-manager/{slug}/{documentId}/unpublish   → REMOVED
```

The public read route is **unchanged**:
```
GET /api/public/document-manager/{slug}/{documentId}   → KEPT (no kind prefix needed — public API always resolves published records by documentId)
```

---

### 10.6 Backend Changes

#### Entity (`internal/domain/entity/content_type.go`)

Add `ListFields` field to `ContentType`:
```go
ListFields []string `json:"listFields,omitempty" bson:"listFields,omitempty"`
```

---

#### Repository Interface (`internal/domain/repository/document_repository.go`)

Add two methods:

```go
// FindDraftsByContentTypePaginated returns paginated draft documents filtered
// by locale, sorted by createdAt descending, plus the total count of matching drafts.
FindDraftsByContentTypePaginated(ctx context.Context, contentTypeSlug string, start, size int, locale string) ([]*entity.Document, int64, error)

// FindPublishedByDocumentIDs batch-fetches published records for multiple
// documentIDs in a single query. Used for efficient status computation on list pages.
FindPublishedByDocumentIDs(ctx context.Context, contentTypeSlug string, documentIDs []string, locale string) ([]*entity.Document, error)
```

---

#### MongoDB Implementation (`internal/infrastructure/mongodb/document_repository.go`)

**`FindDraftsByContentTypePaginated`:**
```
Filter: { version: "draft", locale: <locale> }
Sort:   { createdAt: -1 }
Skip:   start
Limit:  size
```
Also runs `CountDocuments` with the same filter to return `total`.

**`FindPublishedByDocumentIDs`:**
```
Filter: { version: "published", locale: <locale>, documentId: { $in: <ids> } }
```
Returns all matching published records in one query.

---

#### Usecase (`internal/usecase/document/document_usecase.go`)

Add new methods to the `UseCase`:

**`GetSingleType(ctx, contentTypeSlug, locale) → (*Document, status string, error)`**
1. Resolve locale.
2. `repo.FindDraftsByContentTypePaginated(ctx, slug, 0, 1, locale)` — at most 1 draft.
3. If count == 0, return `ErrNotFound`.
4. Lookup published record for the single document.
5. Compute and return status.

**`SaveSingleType(ctx, contentTypeSlug, data map[string]any, locale, userID) → (*Document, error)`**
1. Resolve locale.
2. Try to find existing draft via `FindDraftsByContentTypePaginated(ctx, slug, 0, 1, locale)`.
3. If exists → build `Document` with existing `DocumentID`, delegate to `Save`.
4. If not exists → build `Document` with empty `DocumentID` (auto-generated), delegate to `Save`.

**`PublishSingleType(ctx, contentTypeSlug, locale, userID) → error`**
1. Find the single draft (same as GetSingleType step 2).
2. If not found → `ErrNotFound`.
3. Delegate to existing `Publish(ctx, slug, draft.DocumentID, locale, userID)`.

**`UnpublishSingleType(ctx, contentTypeSlug, locale) → error`**
1. Find the single draft.
2. If not found → `ErrNotFound`.
3. Delegate to existing `Unpublish(ctx, slug, draft.DocumentID, locale)`.

**`GetAllPaginated(ctx, contentTypeSlug, start, size int, locale) → (docs []*Document, statuses []string, total int64, error)`**
1. Resolve locale.
2. `repo.FindDraftsByContentTypePaginated(ctx, slug, start, size, locale)` → drafts + total.
3. Collect `documentIDs` from drafts.
4. `repo.FindPublishedByDocumentIDs(ctx, slug, documentIDs, locale)` → published map.
5. Compute status for each draft against the published map.
6. Return drafts, statuses, total.

---

#### Handler (`internal/delivery/http/handler/document_handler.go`)

Update `DocumentHandler` to accept both usecases:

```go
type DocumentHandler struct {
    uc   documentUseCase
    ctUC contentTypeUseCase
}

func NewDocumentHandler(uc documentUseCase, ctUC contentTypeUseCase) *DocumentHandler {
    return &DocumentHandler{uc: uc, ctUC: ctUC}
}
```

Extend the `documentUseCase` interface with the new methods:
```go
type documentUseCase interface {
    // ... existing methods
    GetSingleType(ctx context.Context, contentTypeSlug, locale string) (*entity.Document, string, error)
    SaveSingleType(ctx context.Context, contentTypeSlug string, data map[string]any, locale, userID string) (*entity.Document, error)
    PublishSingleType(ctx context.Context, contentTypeSlug, locale, userID string) error
    UnpublishSingleType(ctx context.Context, contentTypeSlug, locale string) error
    GetAllPaginated(ctx context.Context, contentTypeSlug string, start, size int, locale string) ([]*entity.Document, []string, int64, error)
}
```

New handler methods:

| Method | Handler Function | Notes |
|--------|-----------------|-------|
| Single-type GET | `GetSingleType(w, r)` | Returns document or 404 |
| Single-type PUT | `SaveSingleType(w, r)` | Create-or-update |
| Single-type Publish | `PublishSingleType(w, r)` | No documentId needed |
| Single-type Unpublish | `UnpublishSingleType(w, r)` | No documentId needed |
| Collection list | `ListCollection(w, r)` | Paginated + field projection |
| Collection GET | `GetCollection(w, r)` | Same as current GetByID |
| Collection POST | `CreateCollection(w, r)` | Same as current Create |
| Collection PUT | `UpdateCollection(w, r)` | Same as current Update |
| Collection DELETE | `DeleteCollection(w, r)` | Same as current Delete |
| Collection Publish | `PublishCollection(w, r)` | Same as current Publish |
| Collection Unpublish | `UnpublishCollection(w, r)` | Same as current Unpublish |

**`ListCollection` flow:**
1. Parse `start`, `size`, `locale` from query params (with defaults and max validation).
2. Fetch content-type by slug via `ctUC.FindBySlug(slug)` → get `ListFields`.
3. Fetch paginated documents via `uc.GetAllPaginated(slug, start, size, locale)`.
4. Project each document's `data` — keep only keys in `ListFields`.
5. Build and return the paginated response.

Field projection helper:
```go
func projectData(data map[string]any, fields []string) map[string]any {
    projected := make(map[string]any, len(fields))
    for _, f := range fields {
        if v, ok := data[f]; ok {
            projected[f] = v
        }
    }
    return projected
}
```

Old handler methods (`List`, `Create`, `GetByID`, `Update`, `Delete`, `Publish`, `Unpublish`) are **removed** — replaced by the kind-specific methods above.

---

#### Router (`cmd/server/main.go`)

Remove old flat routes. Add kind-prefixed routes:

```go
// Single-type document routes
mux.Handle("GET /api/document-manager/single-type/{slug}", authRequired(docHandler.GetSingleType))
mux.Handle("PUT /api/document-manager/single-type/{slug}", adminOnly(docHandler.SaveSingleType))
mux.Handle("POST /api/document-manager/single-type/{slug}/publish", adminOnly(docHandler.PublishSingleType))
mux.Handle("POST /api/document-manager/single-type/{slug}/unpublish", adminOnly(docHandler.UnpublishSingleType))

// Collection-type document routes
mux.Handle("GET /api/document-manager/collection-type/{slug}", authRequired(docHandler.ListCollection))
mux.Handle("GET /api/document-manager/collection-type/{slug}/{documentId}", authRequired(docHandler.GetCollection))
mux.Handle("POST /api/document-manager/collection-type/{slug}", adminOnly(docHandler.CreateCollection))
mux.Handle("PUT /api/document-manager/collection-type/{slug}/{documentId}", adminOnly(docHandler.UpdateCollection))
mux.Handle("DELETE /api/document-manager/collection-type/{slug}/{documentId}", adminOnly(docHandler.DeleteCollection))
mux.Handle("POST /api/document-manager/collection-type/{slug}/{documentId}/publish", adminOnly(docHandler.PublishCollection))
mux.Handle("POST /api/document-manager/collection-type/{slug}/{documentId}/unpublish", adminOnly(docHandler.UnpublishCollection))
```

Update `NewDocumentHandler` call to pass both usecases:
```go
docHandler := deliveryhandler.NewDocumentHandler(documentUC, ctUC)
```

---

### 10.7 Frontend Changes

#### Types (`src/types/cms.ts`)

Add `ListFields` to `ContentType` and new response types:

```ts
export interface ContentType extends ContentTypeSummary {
  Fields?: FieldDefinition[]
  ListFields?: string[]
  CreatedAt: string
  UpdatedAt: string
}

export interface PaginatedResponse<T> {
  items: T[]
  total: number
  start: number
  size: number
}
```

---

#### Hooks — Split by Kind

Delete `src/hooks/useDocuments.ts`. Replace with three files:

**`src/hooks/useSingleTypeDocuments.ts`:**
```ts
// useSingleTypeDocument(slug, locale)
//   → GET /api/document-manager/single-type/{slug}?locale={locale}
//   → returns Document | undefined
//   → on 404: returns undefined (not an error — FE shows empty form)

// useSaveSingleType()
//   → PUT /api/document-manager/single-type/{slug}?locale={locale}
//   → invalidates single-type document query key

// usePublishSingleType()
//   → POST /api/document-manager/single-type/{slug}/publish?locale={locale}
//   → invalidates single-type document query key

// useUnpublishSingleType()
//   → POST /api/document-manager/single-type/{slug}/unpublish?locale={locale}
//   → invalidates single-type document query key
```

`useSingleTypeDocument` must intercept 404 responses and return `undefined` instead of throwing — this is the "no document yet" state, not an error.

**`src/hooks/useCollectionDocuments.ts`:**
```ts
// useCollectionDocuments(slug, start, size, locale)
//   → GET /api/document-manager/collection-type/{slug}?start={start}&size={size}&locale={locale}
//   → returns PaginatedResponse<Document>

// useCollectionDocument(slug, documentId, locale)
//   → GET /api/document-manager/collection-type/{slug}/{documentId}?locale={locale}
//   → returns Document

// useCreateCollectionDocument()
//   → POST /api/document-manager/collection-type/{slug}?locale={locale}
//   → invalidates collection list query key

// useUpdateCollectionDocument()
//   → PUT /api/document-manager/collection-type/{slug}/{documentId}?locale={locale}
//   → invalidates collection list + detail query keys

// useDeleteCollectionDocument()
//   → DELETE /api/document-manager/collection-type/{slug}/{documentId}
//   → invalidates collection list query key

// usePublishCollectionDocument()
//   → POST /api/document-manager/collection-type/{slug}/{documentId}/publish?locale={locale}
//   → invalidates collection detail + list query keys

// useUnpublishCollectionDocument()
//   → POST /api/document-manager/collection-type/{slug}/{documentId}/unpublish?locale={locale}
//   → invalidates collection detail + list query keys
```

**`src/hooks/useLocales.ts`:**
Extract the existing `useLocales()` hook into its own file (unchanged logic).

**Query key namespacing (updated):**
```ts
const KEYS = {
  singleType: (slug: string, locale: string) =>
    ['documents', 'single-type', slug, locale] as const,
  collectionList: (slug: string, start: number, size: number, locale: string) =>
    ['documents', 'collection-type', slug, start, size, locale] as const,
  collectionDetail: (slug: string, id: string, locale: string) =>
    ['documents', 'collection-type', 'detail', slug, id, locale] as const,
  locales: ['locales'] as const,
}
```

---

#### Component Changes

**`ContentTypePanel.tsx` (single-type + collection-detail editor):**
- For single-type: uses `useSingleTypeDocument(slug, locale)` — when `undefined`, renders empty form (first Save creates the document via `useSaveSingleType`).
- For collection-detail: uses `useCollectionDocument(slug, documentId, locale)`.
- Save mutation: `useSaveSingleType()` for single-type, `useUpdateCollectionDocument()` for collection.
- Publish/Unpublish: `usePublishSingleType()` / `useUnpublishSingleType()` for single-type; `usePublishCollectionDocument()` / `useUnpublishCollectionDocument()` for collection.
- Receives a `kind: 'single' | 'collection'` prop (or derives from `contentType.Kind`) to select the correct hooks.

**`CollectionListPage.tsx` — paginated + schema-derived columns:**
- Uses `useCollectionDocuments(slug, start, size, locale)` with local pagination state (`start`, `size`).
- **Column derivation:** reads `contentType.ListFields` (from the full content-type schema) and cross-references `contentType.Fields` to determine each column's type. Renders each column according to its field type:
  - `text` → string
  - `boolean` → `✓` / `—`
  - `number` → numeric string
  - `media` → `<img>` thumbnail
  - Other types → string fallback
- **Fallback:** if `ListFields` is empty, uses first 3 fields from `Fields`.
- **FE registry override:** if a content-type-registry entry defines `columns` for the slug, those override the schema-derived columns (backward-compatible escape hatch for custom labels or type overrides).
- **Pagination controls:** Previous/Next buttons + "Showing X–Y of Z" text. Previous disabled when `start === 0`; Next disabled when `start + size >= total`.

**`content-type-registry/index.ts`:**
- `columns` property becomes optional — only needed for custom label or type overrides that differ from the schema-derived defaults.
- Registry continues to support `wrapper` overrides.

**`ContentTypePage.tsx`:**
- No structural changes — already routes by kind to `ContentTypePanel` or `CollectionListPage`.

---

### 10.8 FE Flow — All Entry Points

```
Enter admin page:
└── Sidebar: GET /api/content-types → ContentTypeSummary[] (no fields)
    └── Select content-type in side-menu
        └── Navigate to ContentTypePage → GET /api/content-types/{slug} → full schema
            ├── if single-type → Render ContentTypePanel
            │   └── GET /api/document-manager/single-type/{slug}?locale={locale}
            │       ├── 404 → empty form (first Save creates document)
            │       └── 200 → form pre-filled with data
            └── if collection-type → Render CollectionListPage
                └── GET /api/document-manager/collection-type/{slug}?start=0&size=20&locale={locale}
                    └── Paginated table (projected fields only)
                        └── Click item / Edit / Add new
                            └── Navigate to CollectionDetailPage
                                └── GET /api/document-manager/collection-type/{slug}/{documentId}?locale={locale}
                                    ├── 404 → error state
                                    └── 200 → form pre-filled with full data

Direct enter single-type page (/admin/content-type/single-type/:slug):
└── Sidebar: GET /api/content-types
└── GET /api/content-types/{slug} → full schema
    └── Render ContentTypePanel → same single-type flow above

Direct enter collection-type list (/admin/content-type/collection-type/:slug):
└── Sidebar: GET /api/content-types
└── GET /api/content-types/{slug} → full schema
    └── Render CollectionListPage → same collection list flow above

Direct enter collection-type document (/admin/content-type/collection-type/:slug/:id):
└── Sidebar: GET /api/content-types
└── GET /api/content-types/{slug} → full schema
    └── Render ContentTypePanel (collection mode)
        └── GET /api/document-manager/collection-type/{slug}/{documentId}?locale={locale}
```

---

### 10.9 Files Changed

**Backend — modified files:**
```
apps/api/content-types/blog-posts.json                                  ← add listFields
apps/api/internal/domain/entity/content_type.go                         ← add ListFields field
apps/api/internal/domain/repository/document_repository.go              ← add 2 new interface methods
apps/api/internal/infrastructure/mongodb/document_repository.go         ← implement new methods
apps/api/internal/infrastructure/mongodb/document_repository_test.go    ← test new methods
apps/api/internal/usecase/content_type/schema_loader.go                 ← parse listFields, validate
apps/api/internal/usecase/document/document_usecase.go                  ← add single-type + paginated methods
apps/api/internal/usecase/document/document_usecase_test.go             ← test new methods
apps/api/internal/delivery/http/handler/document_handler.go             ← rewrite: kind-specific handlers + ctUC dep
apps/api/internal/delivery/http/handler/document_handler_test.go        ← rewrite: test new handlers
apps/api/cmd/server/main.go                                             ← new route wiring, pass ctUC to handler
```

**Frontend — new files:**
```
apps/web/src/hooks/useSingleTypeDocuments.ts
apps/web/src/hooks/useCollectionDocuments.ts
apps/web/src/hooks/useLocales.ts
```

**Frontend — modified files:**
```
apps/web/src/types/cms.ts                                                        ← add ListFields, PaginatedResponse
apps/web/src/pages/admin/panels/content-type/ContentTypePanel.tsx                ← use kind-specific hooks
apps/web/src/pages/admin/panels/collection-type/layout/CollectionListPage.tsx    ← pagination + schema-derived columns
apps/web/src/content-type-registry/index.ts                                      ← columns becomes optional
```

**Frontend — deleted files:**
```
apps/web/src/hooks/useDocuments.ts    ← replaced by kind-specific hook files
```

---

### 10.10 Testing

#### Backend — Usecase (`document_usecase_test.go`)

**GetSingleType:**
- Document exists → returns document + computed status
- No document exists → returns `ErrNotFound`
- Invalid locale → returns validation error

**SaveSingleType:**
- First save (no existing document) → creates new document, returns with status `"draft"`
- Subsequent save (document exists) → updates data, preserves `documentId`
- Invalid locale → returns validation error

**PublishSingleType / UnpublishSingleType:**
- Document exists → delegates to Publish/Unpublish correctly
- No document → returns `ErrNotFound`

**GetAllPaginated:**
- Returns correct page (start=0, size=2 out of 5 documents → 2 items, total=5)
- Computes status correctly for each item (batch published lookup)
- Filters by locale
- Empty result → returns empty slice + total=0

#### Backend — Repository (`document_repository_test.go`)

**FindDraftsByContentTypePaginated:**
- Pagination: skip + limit applied correctly
- Locale filter works
- Total count matches all drafts for locale (not just page)
- Sort order: createdAt descending

**FindPublishedByDocumentIDs:**
- Returns matching published records for given IDs
- Filters by locale
- Returns empty slice for no matches

#### Backend — Handler (`document_handler_test.go`)

**Single-type handlers:**
- `GET` returns 200 + document when exists
- `GET` returns 404 when no document
- `PUT` returns 201 on first save, 200 on update
- `Publish` returns 200, `Unpublish` returns 200
- Slug validation: invalid slug → 400

**Collection handlers:**
- `ListCollection` returns paginated response with projected fields
- `ListCollection` defaults: start=0, size=20
- `ListCollection` respects max size=100
- `GetCollection`, `CreateCollection`, `UpdateCollection`, `DeleteCollection` — same behavior as previous flat handlers
- `PublishCollection`, `UnpublishCollection` — same behavior
- Slug and documentId validation: invalid → 400

#### Frontend — Hook tests

**`useSingleTypeDocument`:**
- Returns document data on 200
- Returns `undefined` on 404 (not an error)
- Query key includes slug + locale

**`useCollectionDocuments`:**
- Returns paginated response
- Query key includes slug + start + size + locale

**Mutation hooks:**
- Each mutation invalidates correct query keys on success
- Error handler shows toast

#### Frontend — Component tests

**`CollectionListPage`:**
- Renders columns derived from `ListFields` + `Fields`
- Pagination controls: Previous disabled at start; Next disabled at end
- "Showing X–Y of Z" text correct
- "Add new item" creates document and navigates

---

### 10.11 Acceptance Criteria

**API restructure:**
- [ ] Old flat `/api/document-manager/{slug}` routes return 404
- [ ] Single-type routes work without `documentId` in URL
- [ ] Collection-type routes require `documentId` where specified
- [ ] `?locale=` param works consistently across all endpoints

**Single-type flow:**
- [ ] `GET /api/document-manager/single-type/{slug}` returns 404 when no document exists
- [ ] `PUT` creates document on first save, updates on subsequent saves
- [ ] Publish/Unpublish work without `documentId`
- [ ] FE shows empty form on 404, pre-filled form on 200

**Collection-type list (paginated):**
- [ ] `GET /api/document-manager/collection-type/{slug}` returns paginated response
- [ ] Response `data` contains only `listFields` (not full document data)
- [ ] `start` and `size` params control pagination
- [ ] `total` reflects all matching documents, not just page size
- [ ] Default `size=20`, max `size=100`
- [ ] Each item includes computed `status`

**Content-type schema (`listFields`):**
- [ ] `listFields` in JSON definition syncs to `ContentType.ListFields` on startup
- [ ] Missing `listFields` defaults to first 3 field names
- [ ] Invalid `listFields` entry (not in `fields`) causes startup error

**FE pagination:**
- [ ] CollectionListPage renders Previous/Next pagination controls
- [ ] Previous disabled when `start === 0`
- [ ] Next disabled when `start + size >= total`
- [ ] Page info displays "Showing X–Y of Z"

**FE column derivation:**
- [ ] Columns derived from `ListFields` + field types in schema
- [ ] `text`, `boolean`, `number`, `media` types render correctly
- [ ] FE registry `columns` overrides schema-derived columns when defined

---

### 10.12 Out of Scope

- Public read API restructure — `GET /api/public/document-manager/{slug}/{documentId}` remains unchanged
- GraphQL schema updates — REST restructure only
- Full-text search / filtering on collection lists
- Sorting controls in collection list (always `createdAt` descending)
- Server-side cursor-based pagination (offset-based is sufficient for CMS scale)
- FE infinite scroll (explicit pagination controls only)
- New input types or field types

---

### 10.13 Boundaries

| Rule | Detail |
|---|---|
| **Always** | Use `?locale=` (not `?local=`) as the locale query parameter name across all endpoints |
| **Always** | Return 404 (not empty object) for single-type GET when no document exists |
| **Always** | Include computed `status` in every document response (single GET, list items, collection detail) |
| **Always** | Project `data` in collection list responses — never return full document data in paginated lists |
| **Always** | Validate `listFields` entries against `fields` names during schema sync startup |
| **Always** | Invalidate the correct query keys after each mutation (kind-specific keys, not shared) |
| **Always** | Batch-fetch published records for status computation on list pages (no N+1 queries) |
| **Never** | Expose DELETE for single-type documents (existing boundary, unchanged) |
| **Never** | Allow `size` parameter above 100 on collection list endpoint |
| **Never** | Return draft data through the public read API (existing boundary, unchanged) |
| **Never** | Include `documentId` in single-type URLs — slug + locale is the unique identifier |
| **Ask first** | Changing default `size` from 20 or maximum from 100 |
| **Ask first** | Adding sort parameters to collection list endpoint |
