# SPEC — personal-cms

A lightweight, code-first Personal Headless CMS. Developers define content panels manually in React; the Go backend enforces strict data contracts via Clean Architecture.

---

## 1. Objective

Build a self-hosted headless CMS for developers and administrators who prefer code over drag-and-drop builders. Content panels are hard-coded React pages. The backend exposes a REST API (GraphQL later) backed by MongoDB, deployed as a single Render.com service running `docker-compose`.

Every content entry follows a **draft → publish** workflow: editors save changes as a draft at any time; the public read API keeps serving the last published version until an explicit Publish action promotes the draft. Content types are grouped as **single-type** (one auto-created singleton entry, e.g. Homepage settings) or **collection-type** (many entries, e.g. Blog posts).

Content-**type structure** (the schema: which fields exist, their types) is defined as code — JSON definition files checked into the repo — never through the API or UI. On API startup, a sync step reads every definition file and creates, updates, or removes the corresponding `ContentType` records in MongoDB. The web UI is for content **data** only (create/edit/delete entries); it never defines or edits the structure itself.

**Not in scope (v1):** GraphQL, PostgreSQL support, Localization/i18n.

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
- **Database IDs**: All entities use a `documentId` string field (mirrors MongoDB `_id`) for future PostgreSQL compatibility.
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
- Query keys are namespaced by resource: `['documents', panelId]`, `['documents', panelId, documentId]`.

**Refetch after mutation — mandatory rule:**
Every `useMutation` that changes data **must** call `queryClient.invalidateQueries` in its `onSuccess` callback to invalidate all affected query keys. The frontend must never display stale data after a successful write.

```ts
// Example pattern
const { mutate } = useMutation({
  mutationFn: (data) => api.updateDocument(panelId, documentId, data),
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['documents', panelId] });
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
- Every content entry is identified by a shared `entryId` and stored as **two separate MongoDB documents in the same collection**: a `draft` record and a `published` record (`version: "draft" | "published"`).
- `draft` record: holds the latest edits, `createdAt`/`updatedAt`/`createdBy`/`updatedBy`, **no** `publishedAt`/`publishedBy`.
- `published` record: only exists once the entry has been published at least once; additionally carries `publishedAt`/`publishedBy`.
- Every record (draft and published) also carries `locale` (defaults to `"en"`) — set automatically by the system, not authored by editors in v1.
- **Save** (`usecase/document.Save`): upserts the `draft` record's `data`, `updatedAt`, and `updatedBy` (current user). Never touches the `published` record. Never returned by the public read API.
- **Publish** (`usecase/document.Publish`): copies `draft.data` → `published` record (creating it if absent), sets `published.updatedAt = draft.updatedAt`, `published.publishedAt = now()`, and `published.publishedBy = current user`.
- Entry `status` is **computed, never stored**: `draft` (no published record exists), `modified` (`draft.updatedAt > published.updatedAt`), `published` (timestamps match).
- The public/content read API resolves an entry to its `published` record only. If no `published` record exists, the entry does not exist from the API's perspective (404) — the draft is invisible to readers, however recent.
- Admin edit screens read the `draft` record (so editors always see their latest unpublished work) plus the computed `status` for a draft/modified/published badge.
- Applies uniformly to **every** content type — there is no per-type opt-out in v1.

### Domain Rules: Content Type Kinds
- `ContentType.kind: "single" | "collection"`.
- **Single-type**: exactly one entry per content type, auto-created as a singleton (`entryId = contentTypeId`) when the content type is created. No create/delete UI — only edit, Save (draft), and Publish.
- **Collection-type**: zero or more entries, each with its own `entryId`. Standard list + create/edit/delete, each entry carrying its own independent draft/published pair and status.
- The admin sidebar groups navigation into two sections — **Single Types** and **Collection Types** — driven by `ContentType.kind`, not by folder location.

### Domain Rules: Content-Type Schema as Code
- Content-type **structure** (fields, types, `kind`) is defined in JSON files under `apps/api/content-types/*.json` — never created or edited via the API or UI. The UI only manages content **data** (entries).
- On every API startup, `usecase/content_type.Sync` reads all definition files and reconciles them against the `ContentType` records in MongoDB:
  - **New file** → create the `ContentType` (and, if `kind: "single"`, auto-create its singleton entry per the Single-Type rule above).
  - **Changed file** (fields added/changed) → update the `ContentType` record's schema in place.
  - **Field removed from a file** → drop the field from the `ContentType` schema, but leave already-stored entry data untouched (the orphaned key stays in MongoDB, simply no longer shown or editable).
  - **File deleted** → delete the `ContentType` and cascade-delete all its entries (draft + published), per the existing cascade-deletion rule.
- Sync is one-directional: JSON definitions are always the source of truth; nothing the UI or API does ever writes back to the definition files.
- JSON schema files declare only content fields. The system fields (`createdAt`, `updatedAt`, `publishedAt`, `createdBy`, `updatedBy`, `publishedBy`, `locale` — see Draft & Publish rules above) are never declared in the schema; they're injected automatically on every record.
- **This restriction applies only to content-type *structure*, never to content *data*.** Creating, updating, deleting, saving, and publishing entries (documents) stays fully available through the existing API and UI for every content type — schema-as-code only removes the ability to create/edit/delete the *type itself* (its name, slug, kind, field list).

---

## 5. Testing Strategy

### Backend
- **Unit tests**: Every usecase tested in isolation with mock repositories (interface-based). Located in `internal/usecase/<name>/<name>_test.go`.
- **Draft/Publish tests**: cover all three computed statuses (`draft`, `modified`, `published`), Save never mutating the published record, Publish syncing `updatedAt`/`publishedAt`, and the public read API returning 404 for entries with no published record yet.
- **Single-type tests**: auto-creation of the singleton entry on content type creation, and that create/delete operations are rejected for single-type content.
- **Schema sync tests**: new definition file creates a `ContentType`; changed file updates its schema in place; removed field drops from schema but leaves stored entry data untouched; deleted file cascades-deletes the `ContentType` and all its entries. Run against a mock repository, not real files, except for one integration test reading actual fixture JSON files.
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
- A single-type content type's singleton entry is auto-created when the content type itself is created — no manual "create" step for editors.
- Entry `status` (draft/modified/published) is always computed from timestamps, never persisted as a stored field.
- Content-type schema sync runs automatically on every API startup — no manual trigger needed.
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
- Never expose create/delete actions for single-type content in the UI or API — only edit, Save, and Publish.
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
- Fetches document via `useDocuments(contentType.ID)`
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
