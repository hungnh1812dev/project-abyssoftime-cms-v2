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
3. **Render deployment**: ~~Single Render service running `docker-compose up`.~~ **Updated (§13):** Two separate Render services — web-api as Web Service (native Go), web-ui as Static Site. Deploy hooks from CI with path-based change detection. Database on Supabase PostgreSQL.

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
| `/admin/content-type/collection-type/:slug/new` | `CollectionDetailPage` | New item — empty form, no `documentId` |
| `/admin/content-type/collection-type/:slug/:id` | `CollectionDetailPage` | Existing item edit form |

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
- "Add new item": navigates to `/admin/content-type/collection-type/:slug/new` (does **not** create a document — the empty form is rendered without a `documentId`)
- Per-row: Edit link → detail page (`/:slug/:id`); Delete button → `useDeleteDocument` (with `window.confirm` guard)
- No locale switching on the list view (locale is only relevant on the detail/edit form)

---

### 7.6 Collection-Type Detail Page (`CollectionDetailPage`)

- Two URL forms:
  - **New item**: `/admin/content-type/collection-type/:slug/new` — no `documentId` in URL
  - **Existing item**: `/admin/content-type/collection-type/:slug/:id` — `documentId` in URL
- **New item flow** (`/new`):
  1. Renders an empty form based on content-type fields (no API fetch)
  2. Publish/Unpublish buttons disabled until first Save
  3. On Save → `useCreateCollectionDocument` → API returns the created document with `documentId`
  4. After successful save → `navigate(`/admin/content-type/collection-type/${slug}/${documentId}?locale=${locale}`)` — replaces the `/new` URL with the real document URL
  5. Page reloads as the existing-item form with full lifecycle (dirty tracking, publish/unpublish)
- **Existing item flow** (`/:id`):
  - Fetches document by `documentId` → pre-fills form
  - Save mutation sends `{ data, locale: activeLocale }` via `useUpdateCollectionDocument`
  - Full lifecycle: dirty tracking, toasts, post-save reset
- Maintains local `locale` state identical to `SingleTypePage`
- When `useLocales()` returns more than one locale, renders a locale `<select>` in `renderActions`
- Switching locale reloads the document for the new locale; form resets and `isDirty` becomes `false`
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
                        ├── Click "Add new item"
                        │   └── Navigate to /admin/content-type/collection-type/{slug}/new
                        │       └── Empty form (no API fetch, no documentId)
                        │           └── Save → POST /api/document-manager/collection-type/{slug}
                        │               └── 201 → navigate to /{slug}/{documentId}?locale={locale}
                        └── Click item / Edit
                            └── Navigate to /admin/content-type/collection-type/{slug}/{documentId}
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

Direct enter new item (/admin/content-type/collection-type/:slug/new):
└── Sidebar: GET /api/content-types
└── GET /api/content-types/{slug} → full schema
    └── Render CollectionDetailPage (new mode) → empty form, Save creates document

Direct enter collection-type document (/admin/content-type/collection-type/:slug/:id):
└── Sidebar: GET /api/content-types
└── GET /api/content-types/{slug} → full schema
    └── Render CollectionDetailPage (edit mode)
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

---

## 11. Backend Tech Stack Refactoring

### 11.1 Objective

Refactor the API backend from a single-service, MongoDB-only architecture to a multi-protocol, multi-database architecture that supports:

1. **Gin** replaces `net/http` + `http.ServeMux` for REST API routing, middleware, and request handling
2. **gqlgen** enhanced with **dynamic schema generation** — each content-type JSON definition auto-generates GraphQL types and query/mutation fields at startup (no static schema file)
3. **gRPC** for inter-service communication — the CMS exposes a gRPC server mirroring the full REST surface AND acts as a gRPC client to call external microservices
4. **GORM** for MySQL/PostgreSQL as an alternative to MongoDB — every entity supports both a GORM and a MongoDB adapter, selectable per-entity via config
5. **mongo-driver** continues as the MongoDB adapter (already in place)

**Target users:** CMS developers and administrators. External services consuming CMS content via gRPC.

**Constraint:** All existing REST API endpoints retain the same URL paths and response shapes. The frontend requires zero changes from this refactoring.

---

### 11.2 Tech Stack Summary

| Layer | Current | After Refactor |
|-------|---------|----------------|
| REST framework | `net/http` + `http.ServeMux` | **Gin** (`github.com/gin-gonic/gin`) |
| GraphQL | `gqlgen` (static schema) | `gqlgen` + **dynamic schema generation** per content-type |
| gRPC | *(none)* | **gRPC server + client** (`google.golang.org/grpc`) |
| SQL database | *(none)* | **GORM** (`gorm.io/gorm`) with MySQL + PostgreSQL drivers |
| MongoDB | `go.mongodb.org/mongo-driver` | *(unchanged)* |
| DB selection | MongoDB only | **Configurable per entity** via env vars |

---

### 11.3 Project Structure (After Refactor)

```
apps/api/
├── cmd/
│   └── server/
│       └── main.go                           # Wires all protocols: Gin, gRPC, GraphQL
├── content-types/                            # JSON schema-as-code definitions (unchanged)
├── proto/                                    # gRPC protocol buffer definitions
│   └── cms/
│       └── v1/
│           ├── document.proto
│           ├── content_type.proto
│           ├── media.proto
│           └── auth.proto
├── graphql/
│   ├── dynamic/                              # Dynamic schema builder (new)
│   │   ├── schema_builder.go                 # Builds SDL from content-type definitions
│   │   ├── schema_builder_test.go
│   │   ├── resolver_factory.go               # Creates resolvers per content-type
│   │   └── resolver_factory_test.go
│   ├── codegen/                              # Retained for base types codegen
│   ├── generated/                            # gqlgen-generated code (base types only)
│   ├── model/
│   └── resolver/
│       ├── resolver.go
│       ├── directive.go
│       └── schema.resolvers.go               # Base resolvers (contentTypes query, etc.)
├── internal/
│   ├── config/
│   │   └── config.go                         # Extended with GORM, gRPC, per-entity DB config
│   ├── domain/
│   │   ├── entity/                           # Entities gain GORM struct tags alongside bson
│   │   │   ├── content_type.go
│   │   │   ├── document.go
│   │   │   ├── user.go
│   │   │   └── media_asset.go
│   │   └── repository/                       # Interfaces unchanged
│   │       ├── document_repository.go
│   │       ├── content_type_repository.go
│   │       ├── user_repository.go
│   │       ├── media_asset_repository.go
│   │       └── storage_adapter.go
│   ├── usecase/                              # Business logic unchanged — DB-agnostic
│   │   ├── auth/
│   │   ├── document/
│   │   ├── content_type/
│   │   └── media/
│   ├── infrastructure/
│   │   ├── mongodb/                          # Existing MongoDB adapters (unchanged)
│   │   ├── gormdb/                           # NEW: GORM-based adapters
│   │   │   ├── client.go                     # GORM connection + auto-migrate
│   │   │   ├── client_test.go
│   │   │   ├── user_repository.go
│   │   │   ├── user_repository_test.go
│   │   │   ├── content_type_repository.go
│   │   │   ├── content_type_repository_test.go
│   │   │   ├── document_repository.go        # Uses JSONB for flexible data
│   │   │   ├── document_repository_test.go
│   │   │   ├── media_asset_repository.go
│   │   │   └── media_asset_repository_test.go
│   │   ├── cloudinary/
│   │   └── s3/
│   ├── delivery/
│   │   ├── http/                             # Gin-based handlers (refactored)
│   │   │   ├── handler/
│   │   │   │   ├── auth_handler.go           # Gin handler signatures
│   │   │   │   ├── document_handler.go
│   │   │   │   ├── content_type_handler.go
│   │   │   │   ├── media_handler.go
│   │   │   │   └── locale_handler.go
│   │   │   ├── middleware/
│   │   │   │   ├── auth.go                   # Gin middleware
│   │   │   │   ├── cors.go
│   │   │   │   ├── ratelimit.go
│   │   │   │   ├── bodylimit.go
│   │   │   │   └── security_headers.go
│   │   │   └── router.go                     # Gin router setup
│   │   └── grpc/                             # NEW: gRPC delivery layer
│   │       ├── server.go                     # gRPC server setup
│   │       ├── document_service.go           # DocumentService implementation
│   │       ├── content_type_service.go
│   │       ├── media_service.go
│   │       ├── auth_service.go
│   │       └── interceptor/
│   │           ├── auth.go                   # JWT auth interceptor
│   │           └── logging.go
│   └── grpcclient/                           # NEW: gRPC client adapters
│       ├── client.go                         # Connection manager
│       └── registry.go                       # Service discovery/registry
├── pkg/
│   ├── jwt/
│   └── errors/
├── go.mod
└── go.sum
```

---

### 11.4 Phase A — Gin Migration (REST Framework)

Replace `net/http` + `http.ServeMux` with Gin. All existing endpoints keep the same paths and response shapes.

#### 11.4.1 Dependencies

```
github.com/gin-gonic/gin v1.10+
```

#### 11.4.2 Handler Signature Changes

All handlers change from `func(w http.ResponseWriter, r *http.Request)` to `func(c *gin.Context)`.

**Before:**
```go
func (h *DocumentHandler) GetSingleType(w http.ResponseWriter, r *http.Request) {
    slug := r.PathValue("slug")
    doc, status, err := h.uc.GetSingleType(r.Context(), slug, localeParam(r))
    if err != nil {
        writeErr(w, err)
        return
    }
    writeJSON(w, http.StatusOK, toSummary(doc, status))
}
```

**After:**
```go
func (h *DocumentHandler) GetSingleType(c *gin.Context) {
    slug := c.Param("slug")
    doc, status, err := h.uc.GetSingleType(c.Request.Context(), slug, c.Query("locale"))
    if err != nil {
        writeErr(c, err)
        return
    }
    c.JSON(http.StatusOK, toSummary(doc, status))
}
```

#### 11.4.3 Middleware Migration

| Current (`net/http`) | Gin equivalent |
|---|---|
| `middleware.Auth(handler)` | `gin.HandlerFunc` — extracts JWT from `Authorization` header, sets `userID` in `c.Set("userID", id)` |
| `middleware.RequireRole("admin", h)` | `gin.HandlerFunc` — checks role from JWT claims in context |
| `middleware.CORS(origins)` | `gin.HandlerFunc` or use `github.com/gin-contrib/cors` |
| `middleware.RateLimit(rps, burst)` | `gin.HandlerFunc` — same token-bucket logic |
| `middleware.BodyLimit(maxBytes)` | `c.Request.Body = http.MaxBytesReader(...)` in middleware |
| `middleware.SecurityHeaders` | `gin.HandlerFunc` — sets response headers |

The `middleware.UserID(ctx)` helper changes to read from Gin context:
```go
func UserID(c *gin.Context) string {
    id, _ := c.Get("userID")
    return id.(string)
}
```

#### 11.4.4 Router Setup (`internal/delivery/http/router.go`)

Extract route registration from `main.go` into a dedicated router setup function:

```go
func SetupRouter(
    authHandler *AuthHandler,
    ctHandler *ContentTypeHandler,
    docHandler *DocumentHandler,
    mediaHandler *MediaHandler,
    localeHandler *LocaleHandler,
    gqlHandler http.Handler,
    cfg *config.Config,
) *gin.Engine {
    r := gin.New()
    r.Use(gin.Recovery())
    r.Use(middleware.CORS(cfg.CORSOrigins))
    r.Use(middleware.SecurityHeaders())
    r.Use(middleware.BodyLimit(cfg.BodyLimitBytes))

    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    // Auth routes (public)
    auth := r.Group("/auth")
    auth.Use(middleware.RateLimit(cfg.RateLimitRPS, cfg.RateLimitBurst))
    {
        auth.GET("/setup", authHandler.SetupStatus)
        auth.POST("/register", authHandler.Register)
        auth.POST("/login", authHandler.Login)
        auth.POST("/refresh", authHandler.Refresh)
        auth.POST("/logout", authHandler.Logout)
    }

    // Protected API routes
    api := r.Group("/api")
    api.Use(middleware.Auth())
    {
        // Content types
        admin := api.Group("")
        admin.Use(middleware.RequireRole("admin"))
        {
            admin.GET("/content-types", ctHandler.ListSummary)
            admin.GET("/content-types/:identifier", ctHandler.Get)
        }

        // Single-type document routes
        st := api.Group("/document-manager/single-type")
        {
            st.GET("/:slug", docHandler.GetSingleType)
            st.PUT("/:slug", middleware.RequireRole("admin"), docHandler.SaveSingleType)
            st.POST("/:slug/publish", middleware.RequireRole("admin"), docHandler.PublishSingleType)
            st.POST("/:slug/unpublish", middleware.RequireRole("admin"), docHandler.UnpublishSingleType)
        }

        // Collection-type document routes
        ct := api.Group("/document-manager/collection-type")
        {
            ct.GET("/:slug", docHandler.ListCollection)
            ct.GET("/:slug/:documentId", docHandler.GetCollection)
            ct.POST("/:slug", middleware.RequireRole("admin"), docHandler.CreateCollection)
            ct.PUT("/:slug/:documentId", middleware.RequireRole("admin"), docHandler.UpdateCollection)
            ct.DELETE("/:slug/:documentId", middleware.RequireRole("admin"), docHandler.DeleteCollection)
            ct.POST("/:slug/:documentId/publish", middleware.RequireRole("admin"), docHandler.PublishCollection)
            ct.POST("/:slug/:documentId/unpublish", middleware.RequireRole("admin"), docHandler.UnpublishCollection)
        }

        // Media routes
        media := api.Group("/media")
        media.Use(middleware.RequireRole("admin"))
        {
            media.GET("", mediaHandler.List)
            media.POST("/upload", mediaHandler.Upload)
            media.DELETE("/:id", mediaHandler.Delete)
        }

        api.GET("/locales", localeHandler.List)
    }

    // Public document route (no auth)
    r.GET("/api/public/document-manager/:slug/:documentId", docHandler.GetPublic)

    // GraphQL — mounted as a Gin-wrapped net/http handler
    r.Any(cfg.GraphQL.Path, gin.WrapH(gqlHandler))

    return r
}
```

#### 11.4.5 Error Handling

Replace `writeErr(w, err)` with a Gin-compatible version:

```go
func writeErr(c *gin.Context, err error) {
    switch {
    case pkgerrors.Is(err, pkgerrors.ErrNotFound):
        c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
    case pkgerrors.Is(err, pkgerrors.ErrValidation):
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    case pkgerrors.Is(err, pkgerrors.ErrConflict):
        c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
    default:
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
    }
}
```

#### 11.4.6 `main.go` Changes

```go
func main() {
    // ... config, DB, repos, usecases, sync (unchanged)

    // Build handlers
    authHandler := handler.NewAuthHandler(authUC, cfg.CookieSecure)
    ctHandler := handler.NewContentTypeHandler(ctUC)
    docHandler := handler.NewDocumentHandler(documentUC, ctUC)
    mediaHandler := handler.NewMediaHandler(mediaUC)
    localeHandler := handler.NewLocaleHandler(cfg.SupportedLocales)

    // GraphQL
    gqlSrv := buildGraphQLServer(documentUC, ctUC, contentTypeDefs)

    // Gin router
    router := handler.SetupRouter(authHandler, ctHandler, docHandler, mediaHandler, localeHandler, gqlSrv, cfg)

    // Start Gin server
    addr := ":" + cfg.Port
    log.Printf("REST server listening on %s", addr)
    if err := router.Run(addr); err != nil {
        log.Fatal(err)
    }
}
```

---

### 11.5 Phase B — Dynamic GraphQL Schema Generation

Replace the single static `schema.graphqls` with a dynamic schema builder that auto-generates GraphQL types and query/mutation fields from content-type JSON definitions at startup.

#### 11.5.1 How It Works

On startup, after content-type JSON definitions are loaded and synced:

1. **Schema builder** reads all `ContentTypeDefinition` structs
2. For each content-type, generates:
   - A **GraphQL type** named after the content-type (PascalCase of slug), e.g. `BlogPost`, `AboutPage`
   - Fields mapped from the content-type's field definitions
3. For each **collection-type**, generates:
   - `Query.<slug>(<slug>Id: ID!, locale: String): <Type>` — fetch one document
   - `Query.<slugPlural>(start: Int, size: Int, locale: String): <Type>Connection` — paginated list
   - `Mutation.create<Type>(data: <Type>Input!): <Type>! @auth`
   - `Mutation.update<Type>(<slug>Id: ID!, data: <Type>Input!): <Type>! @auth`
   - `Mutation.delete<Type>(<slug>Id: ID!): Boolean! @auth`
   - `Mutation.publish<Type>(<slug>Id: ID!, locale: String): <Type>! @auth`
   - `Mutation.unpublish<Type>(<slug>Id: ID!, locale: String): <Type>! @auth`
4. For each **single-type**, generates:
   - `Query.<slug>(locale: String): <Type>` — fetch the singleton
   - `Mutation.save<Type>(data: <Type>Input!, locale: String): <Type>! @auth`
   - `Mutation.publish<Type>(locale: String): <Type>! @auth`
   - `Mutation.unpublish<Type>(locale: String): <Type>! @auth`

#### 11.5.2 Field Type Mapping

| Content-Type Field `type` | GraphQL Type |
|---|---|
| `text` | `String` |
| `richtext` | `String` |
| `number` | `Float` |
| `boolean` | `Boolean` |
| `media` | `String` (URL) |
| `json` | `JSON` (scalar) |
| `component` | Nested object type (recursive) |

#### 11.5.3 Generated Schema Example

Given `content-types/blog-posts.json`:
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

Generated SDL (merged into runtime schema):
```graphql
type BlogPost {
  documentId: ID!
  title: String
  slug: String
  coverImage: String
  excerpt: String
  body: String
  readingTime: Float
  featured: Boolean
  metadata: JSON
  locale: String!
  status: String!
  createdAt: Time!
  updatedAt: Time!
  publishedAt: Time
}

type BlogPostConnection {
  items: [BlogPost!]!
  total: Int!
  start: Int!
  size: Int!
}

input BlogPostInput {
  title: String
  slug: String
  coverImage: String
  excerpt: String
  body: String
  readingTime: Float
  featured: Boolean
  metadata: JSON
}

extend type Query {
  blogPost(blogPostId: ID!, locale: String): BlogPost
  blogPosts(start: Int, size: Int, locale: String): BlogPostConnection!
}

extend type Mutation {
  createBlogPost(data: BlogPostInput!): BlogPost! @auth
  updateBlogPost(blogPostId: ID!, data: BlogPostInput!): BlogPost! @auth
  deleteBlogPost(blogPostId: ID!): Boolean! @auth
  publishBlogPost(blogPostId: ID!, locale: String): BlogPost! @auth
  unpublishBlogPost(blogPostId: ID!, locale: String): BlogPost! @auth
}
```

#### 11.5.4 Schema Builder (`graphql/dynamic/schema_builder.go`)

```go
type SchemaBuilder struct {
    defs []contenttype.ContentTypeDefinition
}

func NewSchemaBuilder(defs []contenttype.ContentTypeDefinition) *SchemaBuilder

// BuildSDL generates the complete SDL string for all content-type-derived
// types, queries, and mutations. This SDL is merged with the base schema
// at runtime.
func (b *SchemaBuilder) BuildSDL() (string, error)
```

Naming conventions:
- Type name: PascalCase of slug (`blog-posts` → `BlogPost`)
- Input name: `<Type>Input`
- Connection name: `<Type>Connection`
- Query single: camelCase of slug (`blogPost`)
- Query list: camelCase plural (`blogPosts`)
- Mutations: `create<Type>`, `update<Type>`, `delete<Type>`, `publish<Type>`, `unpublish<Type>`
- For single-type: `save<Type>` instead of create/update

#### 11.5.5 Resolver Factory (`graphql/dynamic/resolver_factory.go`)

Creates resolver functions per content-type that delegate to the existing document usecase:

```go
type ResolverFactory struct {
    documentUC  *document.UseCase
    contentTypeUC *contenttype.UseCase
}

func NewResolverFactory(docUC *document.UseCase, ctUC *contenttype.UseCase) *ResolverFactory

// BuildFieldResolvers returns a map of field name → resolver function
// for all dynamically generated query and mutation fields.
func (f *ResolverFactory) BuildFieldResolvers(defs []contenttype.ContentTypeDefinition) map[string]graphql.FieldResolveFn
```

Each generated resolver internally calls the same usecase methods as the REST handlers — no business logic duplication.

#### 11.5.6 Runtime Schema Assembly

In `main.go`, after content-type sync:

```go
// Build dynamic GraphQL schema from content-type definitions
schemaBuilder := dynamic.NewSchemaBuilder(defs)
dynamicSDL, err := schemaBuilder.BuildSDL()
if err != nil {
    log.Fatalf("graphql dynamic schema: %v", err)
}

// Merge base schema + dynamic SDL
// Use gqlgen's runtime schema merging or graphql-go/graphql for dynamic execution
resolverFactory := dynamic.NewResolverFactory(documentUC, ctUC)
fieldResolvers := resolverFactory.BuildFieldResolvers(defs)

gqlSrv := buildGraphQLServer(dynamicSDL, fieldResolvers, documentUC, ctUC)
```

#### 11.5.7 Base Schema (Static)

The static schema retains only base types and non-content-type queries:

```graphql
scalar JSON
scalar Time

directive @auth on FIELD_DEFINITION

type ContentType {
  id: ID!
  name: String!
  slug: String!
  kind: String!
  createdAt: Time!
  updatedAt: Time!
}

type Query {
  contentTypes: [ContentType!]!
}
```

All `Document`-related types and resolvers are generated dynamically.

---

### 11.6 Phase C — GORM Adapters (Multi-Database Support)

Every entity gets a GORM-based repository adapter alongside the existing MongoDB adapter. A config flag per entity selects which adapter to use at startup.

#### 11.6.1 Dependencies

```
gorm.io/gorm v1.26+
gorm.io/driver/mysql v1.5+
gorm.io/driver/postgres v1.5+
```

#### 11.6.2 Entity Struct Tag Changes

Entities gain GORM struct tags alongside existing `bson` and `json` tags:

```go
type User struct {
    ID           string    `bson:"_id,omitempty"  gorm:"column:id;primaryKey"   json:"-"`
    DocumentID   string    `bson:"documentId"     gorm:"column:document_id;uniqueIndex" json:"documentId"`
    Email        string    `bson:"email"          gorm:"column:email;uniqueIndex"        json:"email"`
    PasswordHash string    `bson:"passwordHash"   gorm:"column:password_hash"            json:"-"`
    Role         Role      `bson:"role"           gorm:"column:role;type:varchar(20)"     json:"role"`
    CreatedAt    time.Time `bson:"createdAt"      gorm:"column:created_at"               json:"createdAt"`
}

type ContentType struct {
    ID         string            `bson:"_id,omitempty" gorm:"column:id;primaryKey"   json:"-"`
    Name       string            `bson:"name"          gorm:"column:name"             json:"name"`
    Slug       string            `bson:"slug"          gorm:"column:slug;uniqueIndex" json:"slug"`
    Kind       ContentKind       `bson:"kind"          gorm:"column:kind;type:varchar(20)" json:"kind"`
    Fields     []FieldDefinition `bson:"fields,omitempty"     gorm:"column:fields;serializer:json"      json:"Fields,omitempty"`
    ListFields []string          `bson:"listFields,omitempty" gorm:"column:list_fields;serializer:json"  json:"listFields,omitempty"`
    CreatedAt  time.Time         `bson:"createdAt"     gorm:"column:created_at"       json:"createdAt"`
    UpdatedAt  time.Time         `bson:"updatedAt"     gorm:"column:updated_at"       json:"updatedAt"`
}

type Document struct {
    DocumentID    string          `bson:"documentId"     gorm:"column:document_id;index"        json:"documentId"`
    Version       DocumentVersion `bson:"version"        gorm:"column:version;type:varchar(20)"  json:"version"`
    ContentTypeID string          `bson:"contentTypeId"  gorm:"column:content_type_id;index"     json:"contentTypeId"`
    Data          map[string]any  `bson:"data"           gorm:"column:data;serializer:json"      json:"data"`
    Locale        string          `bson:"locale"         gorm:"column:locale"                    json:"locale"`
    CreatedAt     time.Time       `bson:"createdAt"      gorm:"column:created_at"                json:"createdAt"`
    UpdatedAt     time.Time       `bson:"updatedAt"      gorm:"column:updated_at"                json:"updatedAt"`
    PublishedAt   time.Time       `bson:"publishedAt,omitempty"  gorm:"column:published_at"      json:"publishedAt,omitempty"`
    CreatedBy     string          `bson:"createdBy"      gorm:"column:created_by"                json:"createdBy"`
    UpdatedBy     string          `bson:"updatedBy"      gorm:"column:updated_by"                json:"updatedBy"`
    PublishedBy   string          `bson:"publishedBy,omitempty"  gorm:"column:published_by"      json:"publishedBy,omitempty"`
    Slug          string          `bson:"-"              gorm:"column:slug;index"                json:"-"`
}

type MediaAsset struct {
    ID            string    `bson:"_id,omitempty"  gorm:"column:id;primaryKey"        json:"ID"`
    DocumentID    string    `bson:"documentId"     gorm:"column:document_id"          json:"documentId"`
    URL           string    `bson:"url"            gorm:"column:url"                  json:"url"`
    ThumbnailURL  string    `bson:"thumbnailUrl"   gorm:"column:thumbnail_url"        json:"thumbnailUrl"`
    PublicID      string    `bson:"publicId"       gorm:"column:public_id"            json:"publicId"`
    FileName      string    `bson:"fileName"       gorm:"column:file_name"            json:"fileName"`
    FileExt       string    `bson:"fileExt"        gorm:"column:file_ext"             json:"fileExt"`
    Hash          string    `bson:"hash"           gorm:"column:hash"                 json:"hash"`
    ContentTypeID string    `bson:"contentTypeId"  gorm:"column:content_type_id"      json:"contentTypeId"`
    DocumentRef   string    `bson:"documentRef"    gorm:"column:document_ref;index"   json:"documentRef"`
    CreatedAt     time.Time `bson:"createdAt"      gorm:"column:created_at"           json:"createdAt"`
}
```

#### 11.6.3 GORM Client (`internal/infrastructure/gormdb/client.go`)

```go
func NewClient(driver string, dsn string) (*gorm.DB, error) {
    var dialector gorm.Dialector
    switch driver {
    case "mysql":
        dialector = mysql.Open(dsn)
    case "postgres":
        dialector = postgres.Open(dsn)
    default:
        return nil, fmt.Errorf("unsupported GORM driver: %s", driver)
    }
    db, err := gorm.Open(dialector, &gorm.Config{})
    if err != nil {
        return nil, fmt.Errorf("gorm connect (%s): %w", driver, err)
    }
    return db, nil
}

func AutoMigrate(db *gorm.DB) error {
    return db.AutoMigrate(
        &entity.User{},
        &entity.ContentType{},
        &entity.Document{},
        &entity.MediaAsset{},
    )
}
```

#### 11.6.4 GORM Document Repository

The MongoDB adapter uses per-content-type collections (`documents_<slug>`). The GORM adapter uses a **single `documents` table** with a `slug` column for content-type routing:

```sql
CREATE TABLE documents (
    document_id VARCHAR(24) NOT NULL,
    version VARCHAR(20) NOT NULL,
    content_type_id VARCHAR(255),
    slug VARCHAR(63) NOT NULL,           -- content-type slug, replaces per-collection routing
    data JSON NOT NULL,                  -- JSONB in PostgreSQL, JSON in MySQL
    locale VARCHAR(10) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    published_at TIMESTAMP,
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    published_by VARCHAR(255),
    PRIMARY KEY (document_id, version, locale, slug),
    INDEX idx_slug_version_locale (slug, version, locale),
    INDEX idx_document_ref (document_id)
);
```

Key difference: MongoDB routes by collection name, GORM routes by `WHERE slug = ?`.

**`FindDraftByDocumentID`:**
```go
func (r *DocumentRepository) FindDraftByDocumentID(ctx context.Context, contentTypeSlug, documentID, locale string) (*entity.Document, error) {
    var doc entity.Document
    result := r.db.WithContext(ctx).
        Where("slug = ? AND document_id = ? AND version = ? AND locale = ?",
            contentTypeSlug, documentID, entity.VersionDraft, locale).
        First(&doc)
    if result.Error != nil {
        if errors.Is(result.Error, gorm.ErrRecordNotFound) {
            return nil, pkgerrors.ErrNotFound
        }
        return nil, result.Error
    }
    return &doc, nil
}
```

**`UpsertDraft`:**
```go
func (r *DocumentRepository) UpsertDraft(ctx context.Context, contentTypeSlug string, doc *entity.Document) error {
    doc.Version = entity.VersionDraft
    doc.Slug = contentTypeSlug
    result := r.db.WithContext(ctx).
        Where("slug = ? AND document_id = ? AND version = ? AND locale = ?",
            contentTypeSlug, doc.DocumentID, entity.VersionDraft, doc.Locale).
        Assign(doc).
        FirstOrCreate(doc)
    return result.Error
}
```

**`EnsureCollection` / `DropCollection`:**
These are no-ops in the GORM adapter — the single `documents` table is auto-migrated on startup, so per-content-type collection management doesn't apply.

```go
func (r *DocumentRepository) EnsureCollection(ctx context.Context, contentTypeSlug string) error {
    return nil // single table, no per-slug collection needed
}

func (r *DocumentRepository) DropCollection(ctx context.Context, contentTypeSlug string) error {
    return nil // DeleteAllByContentType handles data cleanup
}
```

#### 11.6.5 Configuration — Per-Entity DB Selection

```go
type DBConfig struct {
    Driver     string // DB_DRIVER: "mongo" | "mysql" | "postgres"
    Mongo      MongoConfig
    SQL        SQLConfig
    EntityDB   EntityDBConfig
}

type SQLConfig struct {
    Driver string // SQL_DRIVER: "mysql" | "postgres"
    DSN    string // SQL_DSN: full connection string
}

type EntityDBConfig struct {
    User        string // DB_USER: "mongo" | "sql" (default: value of DB_DRIVER)
    ContentType string // DB_CONTENT_TYPE: "mongo" | "sql"
    Document    string // DB_DOCUMENT: "mongo" | "sql"
    Media       string // DB_MEDIA: "mongo" | "sql"
}
```

Environment variables:
| Variable | Description | Default |
|----------|-------------|---------|
| `DB_DRIVER` | Default database driver | `mongo` |
| `SQL_DRIVER` | SQL dialect when using GORM | `postgres` |
| `SQL_DSN` | SQL connection string | *(required when any entity uses SQL)* |
| `DB_USER` | DB adapter for User entity | value of `DB_DRIVER` |
| `DB_CONTENT_TYPE` | DB adapter for ContentType entity | value of `DB_DRIVER` |
| `DB_DOCUMENT` | DB adapter for Document entity | value of `DB_DRIVER` |
| `DB_MEDIA` | DB adapter for MediaAsset entity | value of `DB_DRIVER` |

#### 11.6.6 Repository Factory (`main.go`)

```go
func newUserRepository(cfg *config.Config, mongoDB *mongo.Database, gormDB *gorm.DB) repository.UserRepository {
    switch cfg.DB.EntityDB.User {
    case "sql":
        return gormdb.NewUserRepository(gormDB)
    default:
        return mongodb.NewUserRepository(mongoDB)
    }
}
// Repeat for ContentType, Document, MediaAsset
```

Both `mongoDB` and `gormDB` are initialized at startup (if configured). Only the adapters selected by config are actually used.

---

### 11.7 Phase D — gRPC Server & Client

#### 11.7.1 Dependencies

```
google.golang.org/grpc v1.70+
google.golang.org/protobuf v1.36+
```

#### 11.7.2 Proto Definitions

**`proto/cms/v1/document.proto`:**
```protobuf
syntax = "proto3";
package cms.v1;
option go_package = "project-abyssoftime-cms-v2/api/proto/cms/v1";

import "google/protobuf/timestamp.proto";

message Document {
  string document_id = 1;
  string version = 2;
  string content_type_id = 3;
  bytes data = 4;                     // JSON-encoded map[string]any
  string locale = 5;
  google.protobuf.Timestamp created_at = 6;
  google.protobuf.Timestamp updated_at = 7;
  google.protobuf.Timestamp published_at = 8;
  string created_by = 9;
  string updated_by = 10;
  string published_by = 11;
  string status = 12;                // computed: draft/modified/published
}

message GetDocumentRequest {
  string content_type_slug = 1;
  string document_id = 2;
  string locale = 3;
}

message ListDocumentsRequest {
  string content_type_slug = 1;
  int32 start = 2;
  int32 size = 3;
  string locale = 4;
}

message ListDocumentsResponse {
  repeated Document items = 1;
  int64 total = 2;
  int32 start = 3;
  int32 size = 4;
}

message SaveDocumentRequest {
  string content_type_slug = 1;
  string document_id = 2;
  bytes data = 3;
  string locale = 4;
}

message PublishDocumentRequest {
  string content_type_slug = 1;
  string document_id = 2;
  string locale = 3;
}

message DeleteDocumentRequest {
  string content_type_slug = 1;
  string document_id = 2;
}

message DeleteDocumentResponse {
  bool success = 1;
}

// Single-type operations
message GetSingleTypeRequest {
  string content_type_slug = 1;
  string locale = 2;
}

message SaveSingleTypeRequest {
  string content_type_slug = 1;
  bytes data = 2;
  string locale = 3;
}

service DocumentService {
  // Collection-type operations
  rpc GetDocument(GetDocumentRequest) returns (Document);
  rpc ListDocuments(ListDocumentsRequest) returns (ListDocumentsResponse);
  rpc SaveDocument(SaveDocumentRequest) returns (Document);
  rpc PublishDocument(PublishDocumentRequest) returns (Document);
  rpc UnpublishDocument(PublishDocumentRequest) returns (Document);
  rpc DeleteDocument(DeleteDocumentRequest) returns (DeleteDocumentResponse);

  // Single-type operations
  rpc GetSingleType(GetSingleTypeRequest) returns (Document);
  rpc SaveSingleType(SaveSingleTypeRequest) returns (Document);
  rpc PublishSingleType(GetSingleTypeRequest) returns (Document);
  rpc UnpublishSingleType(GetSingleTypeRequest) returns (Document);
}
```

**`proto/cms/v1/content_type.proto`:**
```protobuf
syntax = "proto3";
package cms.v1;
option go_package = "project-abyssoftime-cms-v2/api/proto/cms/v1";

import "google/protobuf/timestamp.proto";

message FieldDefinition {
  string name = 1;
  string type = 2;
  repeated string ext = 3;
  repeated FieldDefinition fields = 4;
}

message ContentType {
  string id = 1;
  string name = 2;
  string slug = 3;
  string kind = 4;
  repeated FieldDefinition fields = 5;
  repeated string list_fields = 6;
  google.protobuf.Timestamp created_at = 7;
  google.protobuf.Timestamp updated_at = 8;
}

message GetContentTypeRequest {
  string slug = 1;
}

message ListContentTypesRequest {}

message ListContentTypesResponse {
  repeated ContentType content_types = 1;
}

service ContentTypeService {
  rpc GetContentType(GetContentTypeRequest) returns (ContentType);
  rpc ListContentTypes(ListContentTypesRequest) returns (ListContentTypesResponse);
}
```

**`proto/cms/v1/media.proto`:**
```protobuf
syntax = "proto3";
package cms.v1;
option go_package = "project-abyssoftime-cms-v2/api/proto/cms/v1";

import "google/protobuf/timestamp.proto";

message MediaAsset {
  string id = 1;
  string document_id = 2;
  string url = 3;
  string thumbnail_url = 4;
  string public_id = 5;
  string file_name = 6;
  string file_ext = 7;
  string hash = 8;
  string content_type_id = 9;
  string document_ref = 10;
  google.protobuf.Timestamp created_at = 11;
}

message ListMediaRequest {
  int32 page = 1;
  int32 limit = 2;
}

message ListMediaResponse {
  repeated MediaAsset assets = 1;
  int64 total = 2;
}

message DeleteMediaRequest {
  string id = 1;
}

message DeleteMediaResponse {
  bool success = 1;
}

service MediaService {
  rpc ListMedia(ListMediaRequest) returns (ListMediaResponse);
  rpc DeleteMedia(DeleteMediaRequest) returns (DeleteMediaResponse);
}
```

**`proto/cms/v1/auth.proto`:**
```protobuf
syntax = "proto3";
package cms.v1;
option go_package = "project-abyssoftime-cms-v2/api/proto/cms/v1";

message LoginRequest {
  string email = 1;
  string password = 2;
}

message LoginResponse {
  string access_token = 1;
  string refresh_token = 2;
}

message RegisterRequest {
  string email = 1;
  string password = 2;
}

message RegisterResponse {
  string access_token = 1;
  string refresh_token = 2;
}

message RefreshRequest {
  string refresh_token = 1;
}

message RefreshResponse {
  string access_token = 1;
}

message SetupStatusRequest {}

message SetupStatusResponse {
  bool has_admin = 1;
}

service AuthService {
  rpc Login(LoginRequest) returns (LoginResponse);
  rpc Register(RegisterRequest) returns (RegisterResponse);
  rpc Refresh(RefreshRequest) returns (RefreshResponse);
  rpc SetupStatus(SetupStatusRequest) returns (SetupStatusResponse);
}
```

#### 11.7.3 gRPC Server Implementation (`internal/delivery/grpc/`)

Each gRPC service implementation wraps the same usecase layer used by REST handlers:

```go
type DocumentServiceServer struct {
    pb.UnimplementedDocumentServiceServer
    documentUC *document.UseCase
}

func NewDocumentServiceServer(docUC *document.UseCase) *DocumentServiceServer {
    return &DocumentServiceServer{documentUC: docUC}
}

func (s *DocumentServiceServer) GetDocument(ctx context.Context, req *pb.GetDocumentRequest) (*pb.Document, error) {
    doc, status, err := s.documentUC.GetForEdit(ctx, req.ContentTypeSlug, req.DocumentId, req.Locale)
    if err != nil {
        return nil, toGRPCError(err)
    }
    return toProtoDocument(doc, status), nil
}
```

**Error mapping:**
```go
func toGRPCError(err error) error {
    switch {
    case pkgerrors.Is(err, pkgerrors.ErrNotFound):
        return status.Error(codes.NotFound, "not found")
    case pkgerrors.Is(err, pkgerrors.ErrValidation):
        return status.Error(codes.InvalidArgument, err.Error())
    case pkgerrors.Is(err, pkgerrors.ErrConflict):
        return status.Error(codes.AlreadyExists, err.Error())
    default:
        return status.Error(codes.Internal, "internal error")
    }
}
```

#### 11.7.4 gRPC Auth Interceptor

```go
func AuthUnaryInterceptor(jwtSecret string) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
        // Skip auth for public methods
        if isPublicMethod(info.FullMethod) {
            return handler(ctx, req)
        }
        md, ok := metadata.FromIncomingContext(ctx)
        if !ok {
            return nil, status.Error(codes.Unauthenticated, "missing metadata")
        }
        token := extractBearerToken(md)
        claims, err := pkgjwt.Validate(token)
        if err != nil {
            return nil, status.Error(codes.Unauthenticated, "invalid token")
        }
        ctx = context.WithValue(ctx, userIDKey, claims.UserID)
        return handler(ctx, req)
    }
}
```

#### 11.7.5 gRPC Client (`internal/grpcclient/`)

For calling external microservices:

```go
type ClientManager struct {
    conns map[string]*grpc.ClientConn
}

func NewClientManager() *ClientManager

// Connect establishes a gRPC connection to an external service.
// Service addresses are configured via env vars.
func (m *ClientManager) Connect(ctx context.Context, serviceName, address string, opts ...grpc.DialOption) error

// GetConnection returns an established connection by service name.
func (m *ClientManager) GetConnection(serviceName string) (*grpc.ClientConn, error)

// Close closes all connections.
func (m *ClientManager) Close() error
```

Configuration:
| Variable | Description | Default |
|----------|-------------|---------|
| `GRPC_PORT` | gRPC server listen port | `9090` |
| `GRPC_SERVICES` | Comma-separated list of `name=address` pairs for client connections | *(empty)* |

Example: `GRPC_SERVICES=search=search-svc:9091,notification=notify-svc:9092`

#### 11.7.6 `main.go` — Dual Server Startup

```go
func main() {
    // ... config, DB, repos, usecases, sync

    // --- REST (Gin) ---
    router := handler.SetupRouter(...)
    go func() {
        addr := ":" + cfg.Port
        log.Printf("REST server listening on %s", addr)
        if err := router.Run(addr); err != nil {
            log.Fatal(err)
        }
    }()

    // --- gRPC Server ---
    grpcServer := grpc.NewServer(
        grpc.UnaryInterceptor(grpcdelivery.AuthUnaryInterceptor(cfg.JWTSecret)),
    )
    pb.RegisterDocumentServiceServer(grpcServer, grpcdelivery.NewDocumentServiceServer(documentUC))
    pb.RegisterContentTypeServiceServer(grpcServer, grpcdelivery.NewContentTypeServiceServer(ctUC))
    pb.RegisterMediaServiceServer(grpcServer, grpcdelivery.NewMediaServiceServer(mediaUC))
    pb.RegisterAuthServiceServer(grpcServer, grpcdelivery.NewAuthServiceServer(authUC))

    lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
    if err != nil {
        log.Fatalf("grpc listen: %v", err)
    }
    log.Printf("gRPC server listening on :%s", cfg.GRPCPort)

    // --- gRPC Client connections ---
    clientMgr := grpcclient.NewClientManager()
    defer clientMgr.Close()
    for name, addr := range cfg.GRPCServices {
        if err := clientMgr.Connect(ctx, name, addr); err != nil {
            log.Printf("grpc client %s: %v (non-fatal)", name, err)
        }
    }

    // Block on gRPC server
    if err := grpcServer.Serve(lis); err != nil {
        log.Fatal(err)
    }
}
```

---

### 11.8 Configuration (Full)

All new environment variables introduced by this refactoring:

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_DRIVER` | Default database driver for all entities | `mongo` |
| `SQL_DRIVER` | SQL dialect (`mysql` or `postgres`) | `postgres` |
| `SQL_DSN` | SQL connection string | *(required when any entity uses SQL)* |
| `DB_USER` | DB adapter override for User entity | value of `DB_DRIVER` |
| `DB_CONTENT_TYPE` | DB adapter override for ContentType entity | value of `DB_DRIVER` |
| `DB_DOCUMENT` | DB adapter override for Document entity | value of `DB_DRIVER` |
| `DB_MEDIA` | DB adapter override for MediaAsset entity | value of `DB_DRIVER` |
| `GRPC_PORT` | gRPC server listen port | `9090` |
| `GRPC_SERVICES` | External gRPC services (`name=address,...`) | *(empty)* |

Existing env vars are unchanged.

---

### 11.9 Testing

#### Phase A — Gin Migration

- **Handler tests**: Replace `httptest.NewRequest` + `httptest.ResponseRecorder` with Gin's test mode (`gin.SetMode(gin.TestMode)`) and `httptest`.
- **Middleware tests**: Test each Gin middleware in isolation with `gin.CreateTestContext`.
- **Integration test**: Start Gin server, hit every endpoint, verify same response shapes as before migration.
- **Regression**: Response bodies must be byte-for-byte identical (same JSON key ordering, same status codes) to prove backward compatibility.

#### Phase B — Dynamic GraphQL

- **SchemaBuilder unit tests**: Given content-type definitions, assert generated SDL contains correct types, queries, mutations.
- **Single-type SDL test**: Single-type definitions produce `save<Type>` mutations (not create/update/delete).
- **Collection-type SDL test**: Collection definitions produce full CRUD mutations + `<Type>Connection` for pagination.
- **Field type mapping test**: Each field type maps to correct GraphQL type.
- **ResolverFactory tests**: Mock usecase, call generated resolvers, verify correct usecase methods invoked.
- **End-to-end**: Boot server with test content-type definitions, execute GraphQL queries against generated schema.

#### Phase C — GORM Adapters

- **Each GORM repository tested with same test cases as MongoDB counterpart** — identical test functions, different setup (GORM with SQLite in-memory for unit tests, PostgreSQL in Docker for integration tests).
- **GORM Document repository**:
  - `FindDraftByDocumentID`: filters by `slug`, `document_id`, `version`, `locale`
  - `UpsertDraft`: creates or updates correctly
  - `FindDraftsByContentTypePaginated`: pagination, locale filter, sort order
  - `FindPublishedByDocumentIDs`: batch lookup with `$in`-equivalent (`WHERE document_id IN (?)`)
  - `EnsureCollection` / `DropCollection`: no-ops that don't error
  - `DeleteAllByContentType`: `DELETE WHERE slug = ?`
- **Auto-migration test**: `AutoMigrate` creates all expected tables with correct columns.
- **Mixed-mode test**: User on MongoDB, Document on PostgreSQL — verify both work simultaneously.

#### Phase D — gRPC

- **Service tests**: Mock usecase, call each gRPC method, verify correct response.
- **Error mapping tests**: Domain errors map to correct gRPC status codes.
- **Auth interceptor test**: Valid token → passes; invalid/missing → `Unauthenticated`.
- **Integration test**: Start gRPC server, connect client, execute full CRUD flow.

---

### 11.10 Implementation Order

The four phases have dependencies and should be implemented in order:

```
Phase A (Gin)           → Foundation: replaces the HTTP framework
Phase B (Dynamic GQL)   → Depends on A: GraphQL mounted on Gin router
Phase C (GORM)          → Independent of A/B: new infrastructure adapters
Phase D (gRPC)          → Depends on A: runs alongside Gin server
```

Recommended sequence: **A → C → B → D**

- A first: all subsequent work builds on Gin
- C next: purely additive (new adapters), no risk to existing code
- B next: replaces static GraphQL schema with dynamic — needs careful testing
- D last: new protocol, most isolated, lowest risk to existing functionality

Each phase should be a separate PR with its own tests passing before merging.

---

### 11.11 Files Changed (Summary)

**New files:**
```
apps/api/proto/cms/v1/document.proto
apps/api/proto/cms/v1/content_type.proto
apps/api/proto/cms/v1/media.proto
apps/api/proto/cms/v1/auth.proto
apps/api/graphql/dynamic/schema_builder.go
apps/api/graphql/dynamic/schema_builder_test.go
apps/api/graphql/dynamic/resolver_factory.go
apps/api/graphql/dynamic/resolver_factory_test.go
apps/api/internal/infrastructure/gormdb/client.go
apps/api/internal/infrastructure/gormdb/client_test.go
apps/api/internal/infrastructure/gormdb/user_repository.go
apps/api/internal/infrastructure/gormdb/user_repository_test.go
apps/api/internal/infrastructure/gormdb/content_type_repository.go
apps/api/internal/infrastructure/gormdb/content_type_repository_test.go
apps/api/internal/infrastructure/gormdb/document_repository.go
apps/api/internal/infrastructure/gormdb/document_repository_test.go
apps/api/internal/infrastructure/gormdb/media_asset_repository.go
apps/api/internal/infrastructure/gormdb/media_asset_repository_test.go
apps/api/internal/delivery/grpc/server.go
apps/api/internal/delivery/grpc/document_service.go
apps/api/internal/delivery/grpc/document_service_test.go
apps/api/internal/delivery/grpc/content_type_service.go
apps/api/internal/delivery/grpc/content_type_service_test.go
apps/api/internal/delivery/grpc/media_service.go
apps/api/internal/delivery/grpc/media_service_test.go
apps/api/internal/delivery/grpc/auth_service.go
apps/api/internal/delivery/grpc/auth_service_test.go
apps/api/internal/delivery/grpc/interceptor/auth.go
apps/api/internal/delivery/grpc/interceptor/auth_test.go
apps/api/internal/delivery/grpc/interceptor/logging.go
apps/api/internal/delivery/http/router.go
apps/api/internal/grpcclient/client.go
apps/api/internal/grpcclient/registry.go
```

**Modified files:**
```
apps/api/go.mod                                                    ← add gin, gorm, grpc dependencies
apps/api/cmd/server/main.go                                        ← dual server startup, repository factory
apps/api/internal/config/config.go                                 ← SQL, gRPC, per-entity DB config
apps/api/internal/domain/entity/content_type.go                    ← add gorm struct tags
apps/api/internal/domain/entity/document.go                        ← add gorm struct tags + Slug field
apps/api/internal/domain/entity/user.go                            ← add gorm struct tags
apps/api/internal/domain/entity/media_asset.go                     ← add gorm struct tags
apps/api/internal/delivery/http/handler/auth_handler.go            ← gin.Context signatures
apps/api/internal/delivery/http/handler/auth_handler_test.go       ← gin test mode
apps/api/internal/delivery/http/handler/document_handler.go        ← gin.Context signatures
apps/api/internal/delivery/http/handler/document_handler_test.go   ← gin test mode
apps/api/internal/delivery/http/handler/content_type_handler.go    ← gin.Context signatures
apps/api/internal/delivery/http/handler/content_type_handler_test.go
apps/api/internal/delivery/http/handler/media_handler.go           ← gin.Context signatures
apps/api/internal/delivery/http/handler/media_handler_test.go
apps/api/internal/delivery/http/handler/locale_handler.go          ← gin.Context signatures
apps/api/internal/delivery/http/handler/locale_handler_test.go
apps/api/internal/delivery/http/middleware/auth.go                 ← gin middleware
apps/api/internal/delivery/http/middleware/auth_test.go
apps/api/internal/delivery/http/middleware/cors.go                 ← gin middleware
apps/api/internal/delivery/http/middleware/ratelimit.go            ← gin middleware
apps/api/internal/delivery/http/middleware/bodylimit.go            ← gin middleware
apps/api/internal/delivery/http/middleware/security_headers.go     ← gin middleware
apps/api/graphql/resolver/resolver.go                              ← integrate dynamic resolvers
apps/api/graphql/resolver/schema.resolvers.go                      ← remove static Document queries
apps/api/Dockerfile                                                ← install protoc for proto compilation
```

**Removed files:**
```
apps/api/graphql/schema.graphqls                                   ← replaced by dynamic generation
apps/api/graphql/generated/generated.go                            ← regenerated for base types only
apps/api/graphql/model/models_gen.go                               ← regenerated for base types only
```

---

### 11.12 Acceptance Criteria

**Phase A — Gin:**
- [ ] All existing REST endpoints return identical responses (same status codes, same JSON shapes)
- [ ] All middleware (auth, CORS, rate limit, body limit, security headers) works under Gin
- [ ] `gin.SetMode(gin.ReleaseMode)` in production; `gin.TestMode` in tests
- [ ] Route registration moved from `main.go` to `router.go`
- [ ] Frontend works without any changes

**Phase B — Dynamic GraphQL:**
- [ ] Each content-type JSON definition auto-generates a typed GraphQL type at startup
- [ ] Collection-types get: single query, list query (with `Connection` type), create/update/delete/publish/unpublish mutations
- [ ] Single-types get: single query, save/publish/unpublish mutations (no delete)
- [ ] Field types map correctly: text→String, number→Float, boolean→Boolean, media→String, json→JSON, richtext→String
- [ ] Dynamic resolvers call existing usecase methods — no business logic duplication
- [ ] `contentTypes` base query continues to work
- [ ] `@auth` directive enforced on all mutations
- [ ] Schema regenerates on restart when content-type definitions change

**Phase C — GORM:**
- [ ] Every GORM repository passes the same test cases as its MongoDB counterpart
- [ ] `DB_DRIVER=postgres` with `SQL_DSN` starts the app with all entities on PostgreSQL
- [ ] `DB_DRIVER=mongo` (default) behavior is unchanged
- [ ] Per-entity overrides work: e.g. `DB_DRIVER=mongo DB_DOCUMENT=sql` uses MongoDB for everything except documents
- [ ] GORM auto-migration creates correct tables and indexes
- [ ] Document flexible `data` field stored as JSONB (PostgreSQL) / JSON (MySQL)
- [ ] `EnsureCollection` / `DropCollection` are safe no-ops in GORM adapter

**Phase D — gRPC:**
- [ ] gRPC server starts on `GRPC_PORT` alongside Gin on `PORT`
- [ ] All proto services compile and register
- [ ] DocumentService mirrors full REST document API surface
- [ ] ContentTypeService exposes read-only content-type queries
- [ ] MediaService exposes list + delete
- [ ] AuthService exposes login, register, refresh, setup-status
- [ ] JWT auth interceptor protects mutations; public queries pass without auth
- [ ] gRPC client manager connects to external services configured via `GRPC_SERVICES`
- [ ] Domain errors map to correct gRPC status codes (NotFound, InvalidArgument, etc.)

---

### 11.13 Out of Scope

- Frontend changes — the web app is not modified in this refactoring
- Database data migration tooling (SQL ↔ MongoDB data sync)
- gRPC streaming (server-streaming or bidirectional) — unary RPCs only
- gRPC-Gateway (HTTP-to-gRPC proxy) — REST and gRPC coexist as separate protocols
- GraphQL subscriptions
- Connection pooling tuning for GORM
- Read replicas or multi-primary database topology
- Service mesh / load balancing for gRPC
- Media upload via gRPC (multipart upload remains REST-only)
- gRPC health checking protocol (future enhancement)

---

### 11.14 Boundaries

| Rule | Detail |
|---|---|
| **Always** | Keep REST endpoint paths and response shapes identical after Gin migration — zero frontend changes |
| **Always** | Route dynamic GraphQL queries/mutations through the same usecase layer as REST — no business logic in resolvers |
| **Always** | Use GORM's `serializer:json` tag for flexible data fields (Fields, ListFields, Data) — never raw SQL JSON functions |
| **Always** | Map domain errors to protocol-appropriate codes in every delivery layer (HTTP status, gRPC status) |
| **Always** | Initialize both MongoDB and SQL connections at startup when configured — lazy init risks runtime failures |
| **Always** | Run `AutoMigrate` for GORM entities on startup (matches MongoDB's `EnsureIndexes` pattern) |
| **Always** | Use `gin.SetMode(gin.TestMode)` in all test files to suppress Gin debug logging |
| **Always** | Proto files live under `proto/cms/v1/` with `option go_package` matching the module path |
| **Never** | Duplicate business logic across REST handlers, GraphQL resolvers, and gRPC services — all three must call usecase methods |
| **Never** | Use MongoDB-specific logic (ObjectID, bson primitives) in usecase or delivery layers — only in `infrastructure/mongodb/` |
| **Never** | Use GORM-specific logic (gorm.DB, gorm.Model) outside `infrastructure/gormdb/` |
| **Never** | Allow `EnsureCollection` / `DropCollection` to fail in the GORM adapter — they must be safe no-ops |
| **Never** | Expose media upload via gRPC — multipart upload remains REST-only |
| **Never** | Hard-code SQL dialect-specific queries in GORM repositories — use GORM's abstraction layer |
| **Ask first** | Choosing between PostgreSQL and MySQL for production deployment |
| **Ask first** | Adding new gRPC client service connections beyond the initial configured set |
| **Ask first** | Changing the gRPC port default from 9090 |

---

## 12. Role-Based Permission System & PostgreSQL Sub-Component Tables

### 12.1 Objective

Replace the hardcoded role system (string constants on the User entity) with a database-backed Role table. Each role defines a set of per-action permissions applied globally to all content types. Roles are fully managed via the API and admin UI (CRUD). Default roles are seeded on first startup; admins can create additional custom roles at runtime.

Additionally, define the PostgreSQL table strategy for content-type sub-components: when a content-type field has `type: "component"`, the GORM adapter creates a dedicated relational table for the component's sub-fields instead of storing them in the parent document's JSONB `data` column. MongoDB behavior is unchanged — components remain nested objects in the document's BSON `data` field.

**Target users:** CMS administrators managing team access and content permissions.

---

### 12.2 Permission Model

Permissions are flat action strings applied globally (not scoped to specific content types).

**Permission actions:**

| Permission | Description |
|---|---|
| `content:read` | View content entries (draft + published) |
| `content:create` | Create new content entries |
| `content:update` | Edit existing content entries (save draft) |
| `content:delete` | Delete content entries |
| `content:publish` | Publish content entries |
| `content:unpublish` | Unpublish content entries |
| `media:read` | View media library |
| `media:upload` | Upload media assets |
| `media:delete` | Delete media assets |
| `users:manage` | View, invite, update role, delete users |
| `roles:manage` | Create, edit, delete roles |
| `access_tokens:manage` | Create, view, delete API access tokens |
| `content_types:read` | View content-type schemas |

**Default roles (seeded on first startup if no roles exist):**

| Role | Slug | Level | Permissions |
|---|---|---|---|
| Super Admin | `super_admin` | 100 | All permissions |
| Admin | `admin` | 80 | All except `roles:manage`, `access_tokens:manage` |
| Editor | `editor` | 60 | `content:*`, `media:*`, `content_types:read` |
| Guest | `guest` | 20 | `content:read`, `media:read`, `content_types:read` |

**Level** is an integer used for hierarchy comparison — a user can only manage users/roles with a lower level than their own. Custom roles can use any level value between 1–99.

---

### 12.3 Role Entity

**File:** `apps/api/internal/domain/entity/role.go` (new)

```go
type Permission string

const (
    PermContentRead       Permission = "content:read"
    PermContentCreate     Permission = "content:create"
    PermContentUpdate     Permission = "content:update"
    PermContentDelete     Permission = "content:delete"
    PermContentPublish    Permission = "content:publish"
    PermContentUnpublish  Permission = "content:unpublish"
    PermMediaRead         Permission = "media:read"
    PermMediaUpload       Permission = "media:upload"
    PermMediaDelete       Permission = "media:delete"
    PermUsersManage       Permission = "users:manage"
    PermRolesManage       Permission = "roles:manage"
    PermAccessTokenManage Permission = "access_tokens:manage"
    PermContentTypesRead  Permission = "content_types:read"
)

type Role struct {
    ID          string    `bson:"_id,omitempty"   gorm:"column:id;primaryKey"              json:"-"`
    DocumentID  string    `bson:"documentId"      gorm:"column:document_id;uniqueIndex"    json:"documentId"`
    Name        string    `bson:"name"            gorm:"column:name"                       json:"name"`
    Slug        string    `bson:"slug"            gorm:"column:slug;uniqueIndex"           json:"slug"`
    Permissions []string  `bson:"permissions"     gorm:"column:permissions;serializer:json" json:"permissions"`
    Level       int       `bson:"level"           gorm:"column:level"                      json:"level"`
    IsDefault   bool      `bson:"isDefault"       gorm:"column:is_default"                 json:"isDefault"`
    CreatedAt   time.Time `bson:"createdAt"       gorm:"column:created_at"                 json:"createdAt"`
    UpdatedAt   time.Time `bson:"updatedAt"       gorm:"column:updated_at"                 json:"updatedAt"`
}
```

**PostgreSQL table: `roles`**
```sql
CREATE TABLE roles (
    id            VARCHAR(255) PRIMARY KEY,
    document_id   VARCHAR(255) UNIQUE NOT NULL,
    name          VARCHAR(255) NOT NULL,
    slug          VARCHAR(63)  UNIQUE NOT NULL,
    permissions   JSONB        NOT NULL,
    level         INTEGER      NOT NULL,
    is_default    BOOLEAN      NOT NULL DEFAULT false,
    created_at    TIMESTAMP    NOT NULL,
    updated_at    TIMESTAMP    NOT NULL
);
```

**MongoDB collection: `roles`** — same fields as entity struct.

`IsDefault` marks the four seeded roles. Default roles cannot be deleted via the API.

---

### 12.4 User ↔ Role Relationship

Replace the `Role Role` string field on User with `RoleID string` referencing the Role's `DocumentID`:

**Before:**
```go
type User struct {
    ID           string    `bson:"_id,omitempty" gorm:"column:id;primaryKey"`
    DocumentID   string    `bson:"documentId"    gorm:"column:document_id;uniqueIndex"`
    Email        string    `bson:"email"         gorm:"column:email;uniqueIndex"`
    PasswordHash string    `bson:"passwordHash"  gorm:"column:password_hash"`
    Role         Role      `bson:"role"          gorm:"column:role;type:varchar(20)"`
    CreatedAt    time.Time `bson:"createdAt"     gorm:"column:created_at"`
}
```

**After:**
```go
type User struct {
    ID           string    `bson:"_id,omitempty" gorm:"column:id;primaryKey"`
    DocumentID   string    `bson:"documentId"    gorm:"column:document_id;uniqueIndex"`
    Email        string    `bson:"email"         gorm:"column:email;uniqueIndex"`
    PasswordHash string    `bson:"passwordHash"  gorm:"column:password_hash"`
    RoleID       string    `bson:"roleId"        gorm:"column:role_id;index"`
    CreatedAt    time.Time `bson:"createdAt"     gorm:"column:created_at"`
}
```

**PostgreSQL:** `users.role_id` references `roles.document_id` (application-level FK — not DB-level constraint, to stay consistent with MongoDB which has no FK enforcement).

**MongoDB:** `users.roleId` is a string reference to the role's `documentId`.

**Startup migration (one-time):**
On startup, if the `roles` collection/table has no records:
1. Seed the four default roles
2. For each existing user with the legacy `role` string field, look up the matching default role by slug and set `roleId` to its `documentId`

The old `Role` type (`type Role string` with constants `RoleSuperAdmin`, `RoleAdmin`, etc.) and `RoleLevel()` function are **removed** from `entity/user.go`. The `Role` entity in `entity/role.go` replaces them entirely.

---

### 12.5 JWT & Auth Changes

JWT access token claims are unchanged structurally:

```go
type Claims struct {
    UserID string `json:"userId"`
    Role   string `json:"role"`   // role slug
    jwt.RegisteredClaims
}
```

The `Role` claim stores the role **slug** (e.g., `"editor"`). The change is that the slug now references a database record rather than a hardcoded constant.

**Permission resolution flow:**
1. On startup, load all roles into an in-memory cache (`slug → Role` with permissions set)
2. When a role is created/updated/deleted via the API, invalidate and reload the cache
3. Auth middleware reads the role slug from JWT → looks up permissions from the cache
4. Route-level middleware checks if the user's role has the required permission(s)

**New middleware:**

```go
// GinRequirePermission aborts with 403 if the authenticated user's role
// lacks the specified permission.
func GinRequirePermission(cache *RoleCache, permission string) gin.HandlerFunc
```

This replaces `GinRequireMinRole` for content, media, and admin routes. `GinRequireMinRole` is retained for level-based hierarchy checks (e.g., "can only manage users with a lower role level").

**RoleCache** (`internal/delivery/http/middleware/role_cache.go` — new):

```go
type RoleCache struct {
    mu    sync.RWMutex
    roles map[string]*entity.Role // slug → Role
}

func NewRoleCache() *RoleCache
func (c *RoleCache) Load(roles []*entity.Role)
func (c *RoleCache) HasPermission(roleSlug, permission string) bool
func (c *RoleCache) GetLevel(roleSlug string) int
func (c *RoleCache) Get(roleSlug string) *entity.Role
```

---

### 12.6 API Routes — Role CRUD

Base path: `/api/roles`

| Method | Route | Required Permission | Response | Description |
|---|---|---|---|---|
| `GET` | `/api/roles` | Any authenticated user | `Role[]` | List all roles |
| `GET` | `/api/roles/:id` | Any authenticated user | `Role` | Get role by documentId |
| `POST` | `/api/roles` | `roles:manage` | `Role` (201) | Create custom role |
| `PUT` | `/api/roles/:id` | `roles:manage` | `Role` | Update role |
| `DELETE` | `/api/roles/:id` | `roles:manage` | `204` | Delete custom role |

**Create request body:**
```json
{
  "name": "Content Manager",
  "slug": "content-manager",
  "permissions": ["content:read", "content:create", "content:update", "content:publish"],
  "level": 50
}
```

**Validation rules:**
- `slug`: same pattern as content-type slugs (`^[a-z0-9]+(?:-[a-z0-9]+)*$`, 1–63 chars)
- `name`: non-empty, max 100 chars
- `permissions`: each entry must be a valid permission string from the defined set
- `level`: 1–99 (level 100 reserved for super_admin; 0 is invalid)
- Cannot create a role with `level >= caller's role level`
- Cannot delete a default role (`isDefault: true`)
- Cannot delete a role that is currently assigned to any user
- Updating a default role: only `permissions` can be changed (`slug`, `name`, `level` are immutable for defaults)

---

### 12.7 Repository Interface

**New file: `internal/domain/repository/role_repository.go`**

```go
type RoleRepository interface {
    Create(ctx context.Context, role *entity.Role) error
    FindByID(ctx context.Context, documentID string) (*entity.Role, error)
    FindBySlug(ctx context.Context, slug string) (*entity.Role, error)
    FindAll(ctx context.Context) ([]*entity.Role, error)
    Update(ctx context.Context, role *entity.Role) error
    Delete(ctx context.Context, documentID string) error
    HasAny(ctx context.Context) (bool, error)
}
```

Implemented by both `infrastructure/mongodb/role_repository.go` and `infrastructure/gormdb/role_repository.go`.

---

### 12.8 PostgreSQL Sub-Component Tables

When a content-type field has `type: "component"`, the GORM adapter creates a separate relational table for the component's sub-fields instead of storing them in the parent document's JSONB `data` column.

#### 12.8.1 Naming Convention

Table name pattern: `component_{content_type_slug}_{component_field_name}`

- Slug hyphens → underscores (SQL-safe)
- CamelCase field names → snake_case

| Content-type slug | Component field name | Table name |
|---|---|---|
| `blog-posts` | `seo` | `component_blog_posts_seo` |
| `about-page` | `heroSection` | `component_about_page_hero_section` |
| `en-vocab-pack` | `wordList` | `component_en_vocab_pack_word_list` |

#### 12.8.2 Component Table Structure

```sql
-- Example: blog-posts has a component field "seo" with sub-fields metaTitle, metaDescription
CREATE TABLE component_blog_posts_seo (
    id          SERIAL PRIMARY KEY,
    parent_id   INTEGER NOT NULL REFERENCES documents(gorm_id) ON DELETE CASCADE,
    -- sub-field columns derived from the component's field definitions:
    meta_title       TEXT,
    meta_description TEXT
);
```

- `id`: auto-increment PK — no `documentId` (sub-components can't be directly queried)
- `parent_id`: FK to the parent document's `gorm_id`, with `ON DELETE CASCADE`
- Sub-field columns: one column per sub-field, using the field-type → SQL-type mapping below

#### 12.8.3 Field Type → SQL Column Mapping

| Content-type field type | PostgreSQL column type |
|---|---|
| `text` | `TEXT` |
| `richtext` | `TEXT` |
| `number` | `DOUBLE PRECISION` |
| `boolean` | `BOOLEAN` |
| `media` | `TEXT` (stores URL string) |
| `json` | `JSONB` |

#### 12.8.4 Nested Components

Not supported in v1. If a component field contains another component sub-field, the inner component is stored as `JSONB` in the parent component table's column — not extracted into yet another table.

#### 12.8.5 MongoDB Behavior

Unchanged. Components remain as nested objects inside the document's BSON `data` field. No additional collections are created for component fields.

---

### 12.9 GORM Document Repository Changes

The GORM document repository transparently handles component fields for the usecase layer — the `map[string]any` data interface is unchanged.

#### On Write (UpsertDraft / UpsertPublished)

1. Look up the content-type schema (via `ContentTypeProvider`)
2. Separate `doc.Data` into scalar fields and component fields
3. Write scalar fields to the `documents` table's `data` JSONB column
4. For each component field, upsert rows in the corresponding component table (keyed by `parent_id`)

#### On Read (FindDraftByDocumentID, etc.)

1. Read the document row from the `documents` table
2. For each component field in the content-type schema, query the component table by `parent_id`
3. Merge component data back into `doc.Data` before returning to the caller

The repository constructor accepts a `ContentTypeProvider` interface so it can look up schemas:

```go
type ContentTypeProvider interface {
    FindBySlug(ctx context.Context, slug string) (*entity.ContentType, error)
}
```

#### EnsureCollection (GORM — no longer a no-op for component types)

```go
func (r *documentRepository) EnsureCollection(ctx context.Context, contentTypeSlug string) error {
    ct, err := r.ctProvider.FindBySlug(ctx, contentTypeSlug)
    if err != nil { return err }

    for _, field := range ct.Fields {
        if field.Type == "component" {
            if err := r.ensureComponentTable(contentTypeSlug, field); err != nil {
                return err
            }
        }
    }
    return nil
}
```

`ensureComponentTable` uses GORM's `Migrator` to create or update the component table with the correct columns.

#### DropCollection (GORM — drops component tables)

When a content-type is deleted, `DropCollection` drops all `component_{slug}_*` tables associated with it, plus deletes all document rows for that slug.

---

### 12.10 Backend Changes — Files

**New files:**
```
apps/api/internal/domain/entity/role.go                          ← Role entity + Permission constants
apps/api/internal/domain/repository/role_repository.go           ← RoleRepository interface
apps/api/internal/infrastructure/mongodb/role_repository.go      ← MongoDB implementation
apps/api/internal/infrastructure/mongodb/role_repository_test.go
apps/api/internal/infrastructure/gormdb/role_repository.go       ← GORM implementation
apps/api/internal/infrastructure/gormdb/role_repository_test.go
apps/api/internal/usecase/role/role_usecase.go                   ← Role CRUD + seed + permission validation
apps/api/internal/usecase/role/role_usecase_test.go
apps/api/internal/delivery/http/handler/role_handler.go          ← Gin handlers for role CRUD
apps/api/internal/delivery/http/handler/role_handler_test.go
apps/api/internal/delivery/http/middleware/role_cache.go          ← In-memory role permission cache
apps/api/internal/delivery/http/middleware/role_cache_test.go
```

**Modified files:**
```
apps/api/internal/domain/entity/user.go                          ← Role string → RoleID string; remove Role type + RoleLevel()
apps/api/internal/domain/entity/user_test.go                     ← Update tests for RoleID
apps/api/internal/infrastructure/gormdb/client.go                ← AutoMigrate includes Role entity
apps/api/internal/infrastructure/gormdb/document_repository.go   ← Component table read/write + ContentTypeProvider dependency
apps/api/internal/infrastructure/gormdb/document_repository_test.go
apps/api/internal/infrastructure/gormdb/user_repository.go       ← Update for RoleID field
apps/api/internal/infrastructure/gormdb/user_repository_test.go
apps/api/internal/infrastructure/mongodb/user_repository.go      ← Update for roleId field
apps/api/internal/infrastructure/mongodb/user_repository_test.go
apps/api/internal/delivery/http/router.go                        ← Add role routes, migrate to GinRequirePermission
apps/api/internal/delivery/http/middleware/gin_auth.go           ← Add GinRequirePermission; keep GinRequireMinRole for hierarchy
apps/api/internal/delivery/http/middleware/gin_auth_test.go
apps/api/internal/usecase/auth/auth_usecase.go                   ← Use RoleID instead of Role string; look up role slug for JWT
apps/api/internal/usecase/auth/auth_usecase_test.go
apps/api/cmd/server/main.go                                      ← Role repo/usecase init, seed on startup, role cache init
```

---

### 12.11 Router Changes

**New route group:**
```go
// Role routes
roleGroup := r.Group("/api/roles", middleware.GinAuth())
{
    roleGroup.GET("", cfg.RoleHandler.List)
    roleGroup.GET("/:id", cfg.RoleHandler.Get)
    roleGroup.POST("", middleware.GinRequirePermission(cache, "roles:manage"), cfg.RoleHandler.Create)
    roleGroup.PUT("/:id", middleware.GinRequirePermission(cache, "roles:manage"), cfg.RoleHandler.Update)
    roleGroup.DELETE("/:id", middleware.GinRequirePermission(cache, "roles:manage"), cfg.RoleHandler.Delete)
}
```

**Existing routes migrated to permission-based middleware:**

| Route Group | Current Middleware | New Middleware |
|---|---|---|
| Content-type routes | `GinRequireMinRole("admin")` | `GinRequirePermission(cache, "content_types:read")` |
| Single-type PUT/POST | `GinRequireMinRole("editor")` | `GinRequirePermission(cache, "content:update")` |
| Collection POST (create) | `GinRequireMinRole("editor")` | `GinRequirePermission(cache, "content:create")` |
| Collection PUT (update) | `GinRequireMinRole("editor")` | `GinRequirePermission(cache, "content:update")` |
| Collection DELETE | `GinRequireMinRole("editor")` | `GinRequirePermission(cache, "content:delete")` |
| Publish routes | `GinRequireMinRole("editor")` | `GinRequirePermission(cache, "content:publish")` |
| Unpublish routes | `GinRequireMinRole("editor")` | `GinRequirePermission(cache, "content:unpublish")` |
| Media GET | `GinRequireMinRole("editor")` | `GinRequirePermission(cache, "media:read")` |
| Media upload | `GinRequireMinRole("editor")` | `GinRequirePermission(cache, "media:upload")` |
| Media DELETE | `GinRequireMinRole("editor")` | `GinRequirePermission(cache, "media:delete")` |
| User routes | `GinRequireMinRole("admin")` | `GinRequirePermission(cache, "users:manage")` |
| Access token routes | `GinRequireMinRole("super_admin")` | `GinRequirePermission(cache, "access_tokens:manage")` |

---

### 12.12 Frontend Changes

#### Types (`src/types/cms.ts`)

```ts
export interface RoleResponse {
  documentId: string
  name: string
  slug: string
  permissions: string[]
  level: number
  isDefault: boolean
  createdAt: string
  updatedAt: string
}

export const ALL_PERMISSIONS = [
  { value: 'content:read',          label: 'View Content',       group: 'Content' },
  { value: 'content:create',        label: 'Create Content',     group: 'Content' },
  { value: 'content:update',        label: 'Edit Content',       group: 'Content' },
  { value: 'content:delete',        label: 'Delete Content',     group: 'Content' },
  { value: 'content:publish',       label: 'Publish Content',    group: 'Content' },
  { value: 'content:unpublish',     label: 'Unpublish Content',  group: 'Content' },
  { value: 'media:read',            label: 'View Media',         group: 'Media' },
  { value: 'media:upload',          label: 'Upload Media',       group: 'Media' },
  { value: 'media:delete',          label: 'Delete Media',       group: 'Media' },
  { value: 'users:manage',          label: 'Manage Users',       group: 'Admin' },
  { value: 'roles:manage',          label: 'Manage Roles',       group: 'Admin' },
  { value: 'access_tokens:manage',  label: 'Manage Tokens',      group: 'Admin' },
  { value: 'content_types:read',    label: 'View Content Types', group: 'Admin' },
] as const
```

#### Hooks (`src/hooks/useRoles.ts` — new)

```ts
// useRoles()
//   → GET /api/roles → returns RoleResponse[]
//   → query key: ['roles']

// useRole(id: string)
//   → GET /api/roles/:id → returns RoleResponse
//   → query key: ['roles', id]

// useCreateRole()
//   → POST /api/roles
//   → invalidates ['roles']

// useUpdateRole()
//   → PUT /api/roles/:id
//   → invalidates ['roles']

// useDeleteRole()
//   → DELETE /api/roles/:id
//   → invalidates ['roles']
```

#### Auth Context Changes (`src/context/AuthContext.tsx`)

- On login/token refresh, fetch the user's role permissions via `GET /api/roles` (or embed in login response)
- Store `permissions: string[]` alongside `role: string` in auth context
- Expose `hasPermission(action: string): boolean` from `useAuth()`
- Replace all `isSuperAdmin` / `isAdminOrAbove` checks with `hasPermission(...)` calls

#### RolesPage.tsx — Full Rewrite

Replace the static hardcoded permission matrix with a dynamic CRUD page:

**List view:**
- Table columns: Name, Slug, Level, Permissions (grouped badges), Actions
- "Create Role" button (visible when user has `roles:manage`)
- Each row: Edit button, Delete button (hidden for default roles)
- Default roles show a locked badge — cannot be deleted, slug/name/level read-only on edit

**Create/Edit form (Dialog):**
- Name: text input
- Slug: text input, auto-derived from name on create; read-only for default roles on edit
- Level: number input (1–99)
- Permissions: checkbox grid grouped by category (Content, Media, Admin)
- Save button → `useCreateRole()` or `useUpdateRole()`

**Delete:**
- Confirm dialog: "Delete role {name}?"
- API returns error if role is assigned to any user → show toast
- Cannot delete default roles (button not rendered)

#### UsersPage.tsx — Changes

- Replace hardcoded `ALL_ROLES` array with data from `useRoles()`
- Role `<Select>` dropdown shows role names fetched from the API
- `rolesBelow()` uses the role's `level` field from the API instead of the hardcoded `roleLevel()` function
- Remove import of `roleLevel` from `@/lib/roles`
- Role assignment sends `roleId` (documentId) instead of role slug string

#### Sidebar.tsx — Changes

- Replace `isSuperAdmin` / `isAdminOrAbove` boolean checks with `hasPermission()`:
  - Users nav item → `hasPermission('users:manage')`
  - Access Tokens nav item → `hasPermission('access_tokens:manage')`
  - Roles nav item → `hasPermission('roles:manage')`
- Remove import of `roleLevel` from `@/lib/roles`

#### `@/lib/roles.ts` — Removed

This file contains the hardcoded `roleLevel()` function. It is replaced by the `useAuth().hasPermission()` hook and role data from the API. Delete the file; remove all imports.

---

### 12.13 Testing

#### Backend — Role Usecase (`role_usecase_test.go`)

- `TestSeedDefaultRoles` — seeds 4 default roles when `HasAny()` returns false; skips when roles already exist
- `TestCreate_Valid` — creates role with valid permissions and level < caller's level
- `TestCreate_DuplicateSlug` — returns conflict error
- `TestCreate_LevelTooHigh` — rejects when `level >= caller's level`
- `TestCreate_InvalidPermission` — rejects unknown permission strings
- `TestUpdate_DefaultRole_OnlyPermissions` — permits permission changes; rejects slug/name/level changes
- `TestUpdate_CustomRole_AllFields` — all fields changeable
- `TestDelete_DefaultRole` — returns validation error
- `TestDelete_AssignedRole` — returns conflict error (role in use by user(s))
- `TestDelete_CustomRole` — succeeds
- `TestFindAll` — returns all roles sorted by level descending

#### Backend — Role Repository (GORM + MongoDB)

Same test cases for both implementations:
- CRUD operations (create, read, update, delete)
- `HasAny` returns false when empty, true after seeding
- `FindBySlug` returns correct role
- Unique constraint violation on duplicate slug → returns conflict error

#### Backend — Role Handler (`role_handler_test.go`)

- `GET /api/roles` → 200 with role list
- `GET /api/roles/:id` → 200 with role; 404 for unknown ID
- `POST /api/roles` → 201 with created role; 400 for invalid data; 403 without `roles:manage`
- `PUT /api/roles/:id` → 200 with updated role; 400 for invalid changes to default role
- `DELETE /api/roles/:id` → 204 for custom role; 400 for default role; 409 for assigned role; 403 without `roles:manage`

#### Backend — Permission Middleware (`role_cache_test.go`, `gin_auth_test.go`)

- `GinRequirePermission("content:create")` — user with permission → next handler called; without → 403
- `RoleCache.Load` — populates cache; `HasPermission` returns correct results
- Cache reload after role update — stale permissions not served

#### Backend — Component Tables (GORM) (`document_repository_test.go`)

- `EnsureCollection` with component fields → creates component table with correct columns and FK
- Write document with component data → scalar data in `documents.data`, component data in component table
- Read document → component data assembled back into `doc.Data`
- Delete document → component table rows cascade-deleted
- `DropCollection` → drops associated component tables
- Content-type without components → `EnsureCollection` behaves as current no-op

#### Frontend — RolesPage

- Renders role list from `useRoles()` hook
- Create form validates: name required, slug format, level range, at least one permission
- Edit form loads existing role data; default role slug/name/level fields disabled
- Delete button hidden for default roles
- Delete shows confirm dialog; toast on API error

#### Frontend — UsersPage

- Role `<Select>` populated from `useRoles()` data (not hardcoded array)
- Only roles with level < current user's role level are shown in dropdown

#### Frontend — Sidebar

- Nav items shown/hidden based on `hasPermission()` not hardcoded role checks

---

### 12.14 Acceptance Criteria

**Role CRUD:**
- [ ] `GET /api/roles` returns all roles including default ones
- [ ] `POST /api/roles` creates a custom role with valid permissions and level
- [ ] `PUT /api/roles/:id` updates role; default roles only allow permission changes
- [ ] `DELETE /api/roles/:id` deletes custom roles only; rejects if role is assigned to a user
- [ ] Default roles (super_admin, admin, editor, guest) seeded automatically on first startup
- [ ] Seeding is idempotent — skipped if any roles already exist

**Permission enforcement:**
- [ ] All content/media/admin routes use `GinRequirePermission` instead of `GinRequireMinRole`
- [ ] Users with custom roles get exactly the permissions defined on their role
- [ ] Permission cache reloads when a role is created, updated, or deleted
- [ ] JWT continues to embed role slug for backward compatibility

**User ↔ Role relationship:**
- [ ] User entity stores `roleId` referencing Role's `documentId`
- [ ] Existing users auto-migrated to role references on startup (one-time)
- [ ] Auth usecase resolves role slug from roleId for JWT generation
- [ ] UsersPage shows dynamic role list from API

**PostgreSQL sub-component tables:**
- [ ] Content-type with `component` field → GORM creates `component_{slug}_{field_name}` table
- [ ] Component table has auto-increment PK + FK to `documents.gorm_id` (cascade delete)
- [ ] No `documentId` column in component tables
- [ ] Write: component data extracted from `doc.Data`, written to component table
- [ ] Read: component data assembled from component table back into `doc.Data`
- [ ] Drop content-type → drops associated component tables
- [ ] MongoDB: no change — components remain nested in BSON `data`

**Frontend:**
- [ ] RolesPage: dynamic role CRUD with permission checkbox grid
- [ ] UsersPage: role assignment uses database roles (not hardcoded list)
- [ ] Sidebar: nav item visibility driven by `hasPermission()`, not role name checks
- [ ] `@/lib/roles.ts` removed; all imports replaced with `useAuth().hasPermission()`

---

### 12.15 Out of Scope

- Per-content-type permission scoping (e.g., "can only edit blog-posts but not about-page")
- Nested component tables (component within component → stored as JSONB in parent component table)
- Role versioning or audit trail
- GraphQL schema changes for roles (REST-only in this phase)
- gRPC service for roles
- Bulk role assignment
- Role-based field-level visibility (hiding specific fields from the form based on role)
- Frontend permission-gated form fields (Save/Publish buttons already controlled; field-level is out of scope)

---

### 12.16 Boundaries

| Rule | Detail |
|---|---|
| **Always** | Seed default roles on first startup; skip seeding if `HasAny()` returns true |
| **Always** | Validate permissions against the allowed set — reject unknown permission strings |
| **Always** | Use `GinRequirePermission(cache, perm)` for content/media/admin routes |
| **Always** | Keep `GinRequireMinRole` only for hierarchy-based checks (managing users/roles with lower level) |
| **Always** | Cache role permissions in memory; invalidate and reload on any role CRUD operation |
| **Always** | Convert content-type slug hyphens to underscores in PostgreSQL component table names |
| **Always** | Convert camelCase component field names to snake_case for PostgreSQL column names |
| **Always** | Cascade-delete component table rows when parent document is deleted (FK constraint) |
| **Always** | Component read/write is transparent to the usecase layer — `map[string]any` data interface unchanged |
| **Never** | Allow deletion of default roles (`isDefault: true`) |
| **Never** | Allow a user to create/edit a role with `level >= their own role's level` |
| **Never** | Store component data in both JSONB `data` and the component table — mutually exclusive per field |
| **Never** | Add `documentId` to component tables — they are not directly queryable |
| **Never** | Change MongoDB document storage behavior for components (always nested in `data`) |
| **Never** | Create nested component tables (component-in-component → JSONB fallback) |
| **Ask first** | Adding new permission actions to the allowed set |
| **Ask first** | Adding per-content-type permission scoping |
| **Ask first** | Changing default role level values or the level range (1–99) |

---

## 13. Deployment — Render.com (Separate Services)

### 13.1 Objective

Deploy web-api and web-ui as two independent Render.com services on the free tier (no Blueprint). Each service redeploys only when its own code changes, triggered by deploy hooks from GitHub Actions CI.

**Supersedes** Resolved Decision #3 (single docker-compose service).

**Target users:** Project maintainers deploying and operating the CMS.

---

### 13.2 Architecture

```
┌──────────────────────────────────────────────────────────┐
│                    Render.com (Free Tier)                 │
│                                                          │
│  ┌─────────────────────────┐  ┌────────────────────────┐ │
│  │ Web Service (Go)        │  │ Static Site             │ │
│  │ abyssoftime-cms-api     │  │ abyssoftime-cms-web     │ │
│  │ Native Go runtime       │  │ React/Vite SPA          │ │
│  │ Port 8080               │  │ Served from Render CDN  │ │
│  └───────────┬─────────────┘  └──────────┬─────────────┘ │
│              │                           │               │
└──────────────┼───────────────────────────┼───────────────┘
               │                           │
               │  HTTPS API calls          │  Browser loads
               │  (CORS-enabled)           │  static assets
               │                           │
       ┌───────┴───────┐           ┌───────┴───────┐
       │ Supabase      │           │ Cloudinary    │
       │ PostgreSQL    │           │ Media storage │
       └───────────────┘           └───────────────┘
```

| Component | Service Type | Name | Source |
|---|---|---|---|
| web-api | Render Web Service (Go) | `abyssoftime-cms-api` | `apps/api` |
| web-ui | Render Static Site | `abyssoftime-cms-web` | `apps/web` |
| Database | Supabase PostgreSQL (external) | — | — |
| Media | Cloudinary (external) | — | — |

**Free tier constraints:**
- Web Services spin down after 15 minutes of inactivity (cold start ~30s for Go). The existing `keep-alive.yml` GitHub Actions workflow pings `/health` every 14 minutes to prevent this.
- Static Sites are served from Render CDN — always available, no cold starts.
- 750 free Web Service hours/month. One service running 24/7 uses ~730 hours — fits within the limit.

---

### 13.3 Prerequisites

Before Render setup, ensure the following external services are ready:

1. **Supabase PostgreSQL** — project created at supabase.com. Note down:
   - Host: `db.<project-ref>.supabase.co`
   - Port: `5432` (default, use `6543` for connection pooling via Supavisor)
   - Database name: `postgres`
   - Username: `postgres` (or a dedicated user)
   - Password: the database password set during project creation

2. **Cloudinary** — account created. Note down:
   - Cloud name, API key, API secret (from Cloudinary dashboard → Settings → API Keys)

3. **GitHub repository** — the monorepo is pushed to GitHub and GitHub Actions CI is passing.

---

### 13.4 Code Changes Required

Three code changes are required before deployment. These ensure the frontend can call the API cross-origin and the backend accepts cross-origin requests.

#### 13.4.1 Frontend — API Base URL (`apps/web/src/lib/api.ts`)

The frontend uses relative URLs (`/api/*`, `/auth/*`) that rely on Vite's dev proxy. Render Static Site has no server-side proxy — the frontend must call the API directly via absolute URL.

**Change:**
```ts
export const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || '',
  withCredentials: true,
})
```

- **Dev:** `VITE_API_URL` is unset → `baseURL` is `''` → relative paths → Vite proxy handles routing to `localhost:8080`.
- **Production:** `VITE_API_URL=https://abyssoftime-cms-api.onrender.com` → all requests go directly to the API service.

`VITE_API_URL` is a build-time env var (Vite embeds it during `vite build`). Set it in Render's Static Site environment.

#### 13.4.2 Backend — Cross-Origin Cookie Configuration (`apps/api/internal/delivery/http/handler/auth_handler.go`)

With web-ui and web-api on different origins (`*.onrender.com` subdomains), the refresh token cookie requires `SameSite=None` + `Secure=true` for the browser to include it in cross-origin requests.

**Current:**
```go
c.SetCookie(RefreshCookieName, refresh, maxAge, "/", "", false, true)
```

**After:**
```go
c.SetSameSite(h.cookieSameSite)
c.SetCookie(RefreshCookieName, refresh, maxAge, "/", "", h.cookieSecure, true)
```

Apply this change in all three locations where the refresh cookie is set: `Login`, `Refresh`, and `Logout`.

**AuthHandler struct changes:**
```go
type AuthHandler struct {
    uc             authUseCase
    cookieSecure   bool
    cookieSameSite http.SameSite
}

func NewAuthHandler(uc authUseCase, cookieSecure bool, cookieSameSite http.SameSite) *AuthHandler {
    return &AuthHandler{uc: uc, cookieSecure: cookieSecure, cookieSameSite: cookieSameSite}
}
```

#### 13.4.3 Backend — Config Additions (`apps/api/internal/config/config.go`)

Add to `Config` struct and `Load()`:

```go
type Config struct {
    // ... existing fields
    CookieSecure   bool          // COOKIE_SECURE, default true
    CookieSameSite http.SameSite // COOKIE_SAMESITE: "none"|"lax"|"strict", default "none"
    CORSOrigins    []string      // CORS_ORIGINS, comma-separated, default "http://localhost:5173"
}
```

Parsing `COOKIE_SAMESITE`:
```go
func parseSameSite(raw string) http.SameSite {
    switch strings.ToLower(raw) {
    case "lax":
        return http.SameSiteLaxMode
    case "strict":
        return http.SameSiteStrictMode
    case "none":
        return http.SameSiteNoneMode
    default:
        return http.SameSiteNoneMode
    }
}
```

#### 13.4.4 Backend — CORS Middleware (Prerequisite from §9.5)

Cross-origin deployment requires the CORS middleware defined in §9.5. If not yet implemented, it must be completed before deployment. The middleware must:

- Accept `CORS_ORIGINS` (comma-separated allowed origins)
- Set `Access-Control-Allow-Credentials: true` (required for cookie-based auth)
- Handle `OPTIONS` preflight requests with `204`
- Never use `Access-Control-Allow-Origin: *` — always explicit origin whitelist

#### 13.4.5 CI — Path-Based Deploy Triggers (`.github/workflows/ci.yml`)

Update deploy jobs to only trigger when the relevant service's files change:

```yaml
  # Detect which directories changed
  changes:
    name: Detect changes
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main' && github.event_name == 'push'
    outputs:
      api: ${{ steps.filter.outputs.api }}
      web: ${{ steps.filter.outputs.web }}
    steps:
      - uses: actions/checkout@v5
      - uses: dorny/paths-filter@v3
        id: filter
        with:
          filters: |
            api:
              - 'apps/api/**'
            web:
              - 'apps/web/**'

  deploy-api:
    name: Deploy API to Render
    runs-on: ubuntu-latest
    needs: [api-test, web-lint, changes]
    environment: Production
    if: needs.changes.outputs.api == 'true'
    steps:
      - name: Trigger Render deploy — API
        run: curl -fsSL "${{ vars.RENDER_DEPLOY_HOOK_API }}"

  deploy-web:
    name: Deploy Web to Render
    runs-on: ubuntu-latest
    needs: [api-test, web-lint, changes]
    environment: Production
    if: needs.changes.outputs.web == 'true'
    steps:
      - name: Trigger Render deploy — Web
        run: curl -fsSL "${{ vars.RENDER_DEPLOY_HOOK_WEB }}"
```

#### 13.4.6 Keep-Alive Workflow Update (`.github/workflows/keep-alive.yml`)

Update the health endpoint URL to the new API service name:

```yaml
- name: Ping API Health Endpoint
  run: curl -fsSL "https://abyssoftime-cms-api.onrender.com/health" > /dev/null
```

---

### 13.5 Render Dashboard Setup — Step by Step

#### Step 1: Create the API Web Service

1. Go to [dashboard.render.com](https://dashboard.render.com) → **New** → **Web Service**
2. Connect your GitHub repository (`project-abyssoftime-cms-v2`)
3. Configure:

| Setting | Value |
|---|---|
| **Name** | `abyssoftime-cms-api` |
| **Region** | Oregon (US West) or closest to your users |
| **Branch** | `main` |
| **Root Directory** | `apps/api` |
| **Runtime** | Go |
| **Build Command** | `go build -trimpath -ldflags="-s -w" -o bin/server ./cmd/server` |
| **Start Command** | `bin/server` |
| **Instance Type** | Free |
| **Auto-Deploy** | No (we use deploy hooks from CI) |

Native Go builds directly on Render's Ubuntu environment — no Docker layer overhead, smaller footprint, faster builds (~1-2 minutes vs ~3-5 minutes with Docker).

4. Add environment variables (see §13.6 below)
5. Click **Create Web Service**
6. Wait for first build to complete (~1-2 minutes)
7. Copy the service URL: `https://abyssoftime-cms-api.onrender.com`

#### Step 2: Create the Web Static Site

1. Go to Render dashboard → **New** → **Static Site**
2. Connect the same GitHub repository
3. Configure:

| Setting | Value |
|---|---|
| **Name** | `abyssoftime-cms-web` |
| **Branch** | `main` |
| **Root Directory** | `apps/web` |
| **Build Command** | `bun install --frozen-lockfile && bun run build` |
| **Publish Directory** | `dist` |
| **Auto-Deploy** | No (we use deploy hooks from CI) |

4. Add environment variables:

| Variable | Value |
|---|---|
| `VITE_API_URL` | `https://abyssoftime-cms-api.onrender.com` |

5. Under **Redirects/Rewrites**, add an SPA rewrite rule:

| Source | Destination | Action |
|---|---|---|
| `/*` | `/index.html` | Rewrite |

This ensures React Router works for all client-side routes (e.g., `/admin/content-type/single-type/homepage` serves `index.html` instead of a 404).

6. Click **Create Static Site**
7. Wait for first build to complete (~1-2 minutes)
8. Copy the site URL: `https://abyssoftime-cms-web.onrender.com`

#### Step 3: Copy Deploy Hook URLs

For each service:

1. Go to the service's **Settings** page on Render
2. Scroll to **Build & Deploy** → **Deploy Hook**
3. Copy the deploy hook URL (format: `https://api.render.com/deploy/srv-...?key=...`)

#### Step 4: Configure GitHub Repository Secrets

1. Go to your GitHub repo → **Settings** → **Environments** → **Production**
2. Add these **environment variables** (not secrets — the CI uses `vars.*`):

| Variable | Value |
|---|---|
| `RENDER_DEPLOY_HOOK_API` | The API service deploy hook URL from Step 3 |
| `RENDER_DEPLOY_HOOK_WEB` | The Static Site deploy hook URL from Step 3 |

#### Step 5: Update CORS Origin on API

After the Static Site is created and you have the URL:

1. Go to the API Web Service → **Environment**
2. Set `CORS_ORIGINS` to the Static Site URL:
   ```
   CORS_ORIGINS=https://abyssoftime-cms-web.onrender.com
   ```
3. Save and manually deploy the API to pick up the new env var

#### Step 6: Verify

1. Open `https://abyssoftime-cms-web.onrender.com` in a browser
2. Verify the login page loads
3. Open browser DevTools → Network tab
4. Attempt login — verify the request goes to `https://abyssoftime-cms-api.onrender.com/auth/login`
5. Verify the response includes:
   - `Access-Control-Allow-Origin: https://abyssoftime-cms-web.onrender.com`
   - `Access-Control-Allow-Credentials: true`
   - `Set-Cookie: refresh_token=...; SameSite=None; Secure; HttpOnly`
6. Verify navigation works (React Router routes should not 404)
7. Verify content operations: create, save, publish a content entry

---

### 13.6 Environment Variables

#### API Web Service (`abyssoftime-cms-api`)

| Variable | Value | Notes |
|---|---|---|
| `PORT` | `8080` | Render also sets `PORT` automatically; explicit value ensures consistency |
| `DB_DRIVER` | `postgres` | Use PostgreSQL via Supabase |
| `DB_HOST` | `db.<project-ref>.supabase.co` | From Supabase dashboard → Settings → Database → Host |
| `DB_PORT` | `5432` | Default PostgreSQL port (use `6543` for Supavisor pooling) |
| `DB_NAME` | `postgres` | Supabase default database name |
| `DB_USERNAME` | `postgres` | Or a dedicated DB user |
| `DB_PASSWORD` | `<supabase-db-password>` | From Supabase dashboard → Settings → Database → Password |
| `DB_SSL_MODE` | `require` | Supabase requires SSL |
| `JWT_SECRET` | `<random-64-char-string>` | Generate: `openssl rand -hex 32` |
| `CLOUDINARY_CLOUD_NAME` | `<your-cloud-name>` | From Cloudinary dashboard |
| `CLOUDINARY_API_KEY` | `<your-api-key>` | From Cloudinary dashboard |
| `CLOUDINARY_API_SECRET` | `<your-api-secret>` | From Cloudinary dashboard |
| `STORAGE_PROVIDER` | `cloudinary` | Active media storage adapter |
| `CONTENT_TYPES_DIR` | `content-types` | Relative to root directory (`apps/api/content-types/`) — native Go runs from the root directory |
| `CORS_ORIGINS` | `https://abyssoftime-cms-web.onrender.com` | The Static Site URL (set after Step 2) |
| `COOKIE_SECURE` | `true` | Required for `SameSite=None` |
| `COOKIE_SAMESITE` | `none` | Required for cross-origin cookie delivery |
| `SUPPORTED_LOCALES` | `en,vi` | Comma-separated locale codes |

#### Static Site (`abyssoftime-cms-web`)

| Variable | Value | Notes |
|---|---|---|
| `VITE_API_URL` | `https://abyssoftime-cms-api.onrender.com` | Build-time env var; Vite embeds it during `vite build` |

---

### 13.7 Files Changed

**Modified files:**
```
apps/web/src/lib/api.ts                                    ← add baseURL from VITE_API_URL
apps/api/internal/delivery/http/handler/auth_handler.go    ← configurable Secure + SameSite on cookies
apps/api/internal/config/config.go                         ← add CookieSecure, CookieSameSite, CORSOrigins
apps/api/cmd/server/main.go                                ← pass cookie config to auth handler
.github/workflows/ci.yml                                   ← path-based change detection for deploy jobs
.github/workflows/keep-alive.yml                           ← update health endpoint URL
```

**New files (if CORS middleware from §9.5 not yet implemented):**
```
apps/api/internal/delivery/http/middleware/cors.go          ← CORS middleware
apps/api/internal/delivery/http/middleware/cors_test.go     ← CORS tests
```

---

### 13.8 Testing

**Local cross-origin simulation:**
1. Start API on `localhost:8080`
2. Build frontend with `VITE_API_URL=http://localhost:8080 bun run build`
3. Serve `dist/` on a different port: `npx serve dist -l 3000`
4. Open `http://localhost:3000` — verify login, CORS headers, cookie behavior

**Post-deploy verification:**
- Health check: `curl https://abyssoftime-cms-api.onrender.com/health` → `{"status":"ok"}`
- CORS preflight: `curl -X OPTIONS -H "Origin: https://abyssoftime-cms-web.onrender.com" -H "Access-Control-Request-Method: POST" https://abyssoftime-cms-api.onrender.com/auth/login -v` → verify CORS headers in response
- Static site: `curl -I https://abyssoftime-cms-web.onrender.com` → 200 with HTML content
- Deep link: `curl -I https://abyssoftime-cms-web.onrender.com/admin/content-type/single-type/test` → 200 (SPA rewrite working)

---

### 13.9 Acceptance Criteria

**Infrastructure:**
- [ ] API Web Service (`abyssoftime-cms-api`) is running on Render and responds to `/health`
- [ ] Static Site (`abyssoftime-cms-web`) is deployed and serves the React SPA
- [ ] SPA client-side routes work (no 404 on direct navigation to `/admin/...`)
- [ ] API connects to Supabase PostgreSQL successfully
- [ ] Media upload to Cloudinary works from the deployed API

**Cross-origin:**
- [ ] Frontend at `abyssoftime-cms-web.onrender.com` successfully calls API at `abyssoftime-cms-api.onrender.com`
- [ ] CORS headers are present on API responses (`Access-Control-Allow-Origin`, `Access-Control-Allow-Credentials`)
- [ ] Refresh token cookie is set with `SameSite=None; Secure; HttpOnly`
- [ ] Login, token refresh, and authenticated API calls work cross-origin

**CI/CD:**
- [ ] Push to `main` with changes only in `apps/api/` triggers API deploy only
- [ ] Push to `main` with changes only in `apps/web/` triggers web deploy only
- [ ] Push to `main` with changes in both triggers both deploys
- [ ] Push to `main` with no changes in `apps/` skips both deploy jobs
- [ ] Deploy hooks are configured in GitHub environment `Production`

**Keep-alive:**
- [ ] `keep-alive.yml` pings the correct API URL every 14 minutes
- [ ] API does not spin down due to inactivity

---

### 13.10 Boundaries

| Rule | Detail |
|---|---|
| **Always** | Set `COOKIE_SAMESITE=none` and `COOKIE_SECURE=true` for cross-origin deployment |
| **Always** | Set `DB_SSL_MODE=require` when connecting to Supabase |
| **Always** | Use explicit origin whitelist in `CORS_ORIGINS` — never `*` |
| **Always** | Set `VITE_API_URL` at build time on the Static Site — it's embedded by Vite, not read at runtime |
| **Always** | Add SPA rewrite rule (`/* → /index.html`) on the Static Site |
| **Always** | Use deploy hooks (not auto-deploy) so deployments only happen after CI passes |
| **Never** | Store deploy hook URLs in public GitHub repository settings — use the `Production` environment |
| **Never** | Set `COOKIE_SAMESITE=lax` for cross-origin deployment — cookies won't be sent cross-origin |
| **Never** | Use `Access-Control-Allow-Origin: *` with `Access-Control-Allow-Credentials: true` — browsers reject this |
| **Ask first** | Adding custom domains (requires DNS configuration + Render TLS provisioning) |

---

### 13.11 Out of Scope

- Custom domain setup (both services use `*.onrender.com` subdomains)
- Render Blueprint / Infrastructure-as-Code (`render.yaml`)
- Render paid tier features (persistent disk, private networking, preview environments)
- Database migration tooling (Supabase handles schema via GORM auto-migrate on API startup)
- CDN caching configuration for the Static Site
- Monitoring and alerting beyond the health check ping
- Multi-environment deployment (staging/production) — single production environment only
