# SPEC — personal-cms

A lightweight, code-first Personal Headless CMS. Developers define content panels manually in React; the Go backend enforces strict data contracts via Clean Architecture. Content types are defined as JSON schema files (schema-as-code); the API syncs them on startup. Every entry follows a draft → publish workflow.

---

## Module Index

| Module | Spec | Scope |
|--------|------|-------|
| **core** | [specs/core.md](specs/core.md) | Shared entities, repository interfaces, config, middleware, DB clients, deployment |
| **auth** | [specs/auth.md](specs/auth.md) | Authentication (login/register/JWT), roles, permissions, rate limiting |
| **content** | [specs/content.md](specs/content.md) | Content types, documents, draft/publish, schema sync, pagination, GraphQL |
| **media** | [specs/media.md](specs/media.md) | Media assets, upload/delete, S3/Cloudinary storage adapters |
| **admin** | [specs/admin.md](specs/admin.md) | User management, invites, access tokens |

---

## Cross-Module Dependency Graph

```
core ← auth     (user/role repos, JWT utilities)
core ← content  (content-type/document repos, config)
core ← media    (media-asset repo, storage adapter interface)
core ← admin    (invite/access-token repos)

auth ← admin    (user usecase uses role usecase for assignment)
content ← media (document delete cascades to media assets)
content ← auth  (auth middleware protects content routes)
```

All cross-module communication goes through interfaces defined in `core` (`domain/repository/`). No module imports another module's usecase directly.

---

## Commands

| Command | Description |
|---|---|
| `make dev` | Start API + web in parallel |
| `make dev-api` | Start Go API server only |
| `make dev-web` | Start Vite dev server only |
| `make test-api` | `go test ./...` inside `apps/api` |
| `make test-web` | `vitest run` inside `apps/web` |
| `make mongo-start` | Start MongoDB container |
| `go run ./cmd/server` | Start API server (from `apps/api/`) |
| `go test ./...` | Run all backend tests |

Quick start: `cd apps/web && npm install && cd ../..` → `make mongo-start` → `make dev`

---

## Code Style

### Go (Backend)
- **Clean Architecture**: `usecase` imports only `domain`; `delivery` and `infrastructure` import `usecase` and `domain`. Zero cross-layer leakage.
- **Error handling**: Wrap at usecase boundary; handlers map to HTTP/gRPC status codes.
- **Naming**: `PascalCase` exported, `camelCase` unexported. `<Entity>Repository` for interfaces.
- **Cascade deletion**: Usecase layer only, not DB-level triggers.
- **No `init()` functions** in business logic.
- **DB isolation**: MongoDB-specific code only in `infrastructure/mongodb/`; GORM only in `infrastructure/gormdb/`.

### React (Frontend)
- TypeScript strict mode, no `any`. Named exports only.
- TailwindCSS + Shadcn UI. Path aliases (`@/components`, `@/lib`).
- TanStack Query for all server state. No raw `useEffect` + `useState` for API calls.
- Every `useMutation` must call `queryClient.invalidateQueries` on success.
- `FormProvider` manages loading, submitting, isDirty states.
- Panels are hard-coded React pages — no dynamic form engine.

---

## Testing Strategy

- **Usecase tests**: Mock repository interfaces, ≥ 80% coverage.
- **Handler tests**: `httptest` with `gin.SetMode(gin.TestMode)`.
- **Infrastructure tests**: Real DB in Docker (MongoDB) or in-memory (GORM/SQLite).
- **GORM repos must pass the same test cases as MongoDB counterparts.**
- **Frontend**: Vitest + React Testing Library + MSW.
- **CI**: `go vet` → `go test ./...` → `npm run lint` → `npm run build`.

Module-specific testing rules are in each module's spec.

---

## Global Boundaries

### Always
- `FormProvider` manages loading/submitting/isDirty automatically.
- Dot-notation names auto-deserialized to nested JSON on submit.
- JWT access token auto-refreshed via refresh token.
- Every `useMutation` invalidates affected query keys on success.
- Save writes draft only; never publishes implicitly.
- Entry status (draft/modified/published) always computed, never stored.
- Schema sync runs on every API startup.
- System fields (`createdAt`, `updatedAt`, etc.) injected automatically.

### Ask Before
- Starting any new coding task — confirm scope and approach.
- Multiple valid approaches — present options, don't choose autonomously.
- Choosing storage adapter per environment.

### Never
- Read, edit, create, or expose `.env` files.
- Use drag-and-drop or dynamic form engine.
- Use `React.Children.map` or recursive child scanning in `FormProvider`.
- Couple usecase logic to a specific database.
- Auto-choose implementation path when multiple options exist.
- Let public read API return draft data.

---

## FE Spec (Unchanged)

The frontend spec is not modularized. Key sections:

### Routing
| Route | Component |
|---|---|
| `/admin/content-type/single-type/:slug` | SingleTypePage |
| `/admin/content-type/collection-type/:slug` | CollectionListPage |
| `/admin/content-type/collection-type/:slug/new` | CollectionDetailPage |
| `/admin/content-type/collection-type/:slug/:id` | CollectionDetailPage |

### Components
- **ContentTypeLayout**: Layout shell with `renderHeader` / `renderActions` slots
- **Content-Type Registry**: Metadata-only module at `src/content-type-registry/index.ts`
- **FormProvider**: Enhanced lifecycle (dirty tracking, toasts, post-save reset)
- **Sidebar**: Eager metadata load, lazy component loading via React.lazy

### Form Lifecycle
| Moment | Behavior |
|---|---|
| Initial load | Fields rendered, pre-filled from server data |
| Clean state | Save button disabled (`isDirty === false`) |
| After edit | Save button enabled (`isDirty === true`) |
| Failed save | `toast.error(msg)`, form stays edited |
| Successful save | `toast.success('Saved')` → invalidate → reset → Save disabled |

### Locale Switching
- `useLocales()` fetches available locales on mount
- `<select>` in `renderActions` only when `locales.length > 1`
- Locale change → re-fetch document → form reset → `isDirty = false`
- All mutations forward `locale: activeLocale`

### Data Fetching (TanStack Query)
- `useSingleTypeDocument(slug, locale)` — 404 returns `undefined` (not error)
- `useCollectionDocuments(slug, start, size, locale)` — paginated
- `useCollectionDocument(slug, documentId, locale)` — single document
- All mutations invalidate kind-specific query keys

---

## Resolved Decisions

1. **Media storage**: S3 + Cloudinary behind `StorageAdapter` interface, selectable via config.
2. **Go module path**: `project-abyssoftime-cms-v2/api`.
3. **Deployment**: Two separate Render.com services — API as Web Service (native Go), web as Static Site. Database on Supabase PostgreSQL.

---

## Completed Milestones

| Milestone | Description |
|-----------|-------------|
| Schema alignment | PK column `id` → `gorm_id`, per-field dynamic columns, `Fields` replaces JSON `data` blob |
| GraphQL overhaul | Media fields return `MediaAsset` objects, response wrappers removed, queries return types directly |
| MediaInput | Stores `documentId` reference (not URL), aspect ratio fix |
| Collection list | Sortable table, server-side ordering, user display names, icon actions |
| UI design system | Indigo tokens, sidebar, sticky action bar, dark mode |
| Bug fixes v1.8 | Auth UUID, register guard, response shape, input persistence, published-by-default GraphQL |

## Pending Changes

*(Track planned/in-progress work here.)*

| Status | Module | Description |
|--------|--------|-------------|
| — | — | No pending specs |
