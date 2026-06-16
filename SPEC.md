# SPEC — personal-cms

A lightweight, code-first Personal Headless CMS. Developers define content panels manually in React; the Go backend enforces strict data contracts via Clean Architecture.

---

## 1. Objective

Build a self-hosted headless CMS for developers and administrators who prefer code over drag-and-drop builders. Content panels are hard-coded React pages. The backend exposes a REST API (GraphQL later) backed by MongoDB, deployed as two Docker services on Render.com.

**Not in scope (v1):** GraphQL, PostgreSQL support, Localization/i18n.

---

## 2. Commands

### Native development (recommended)

MongoDB runs as a container; API and web run natively. See `docs/local-dev.md` for full setup.

| Command | Description |
|---|---|
| `make mongo-start` | Start MongoDB container (`cms-mongo`, port 27017) |
| `make mongo-stop` | Stop the MongoDB container |
| `make dev` | Start API + web in parallel (Ctrl-C stops both) |
| `make dev-api` | Start Go API server only |
| `make dev-web` | Start Vite dev server only |
| `make test-api` | `go test ./...` inside `apps/api` |
| `make test-web` | `vitest run` inside `apps/web` |

Podman users: prefix any `make` target with `CONTAINER_CLI=podman`.

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
│   │   │   └── server/               # main.go — entry point
│   │   ├── internal/
│   │   │   ├── domain/               # Entities, value objects, repository interfaces
│   │   │   │   ├── entity/           # User, Document, ContentType, MediaAsset
│   │   │   │   └── repository/       # Pure interfaces (no DB code)
│   │   │   ├── usecase/              # Application business logic (DB-agnostic)
│   │   │   │   ├── auth/
│   │   │   │   ├── document/
│   │   │   │   ├── content_type/
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
│   │   └── go.mod                    # module: personal-cms/api
│   │
│   └── web/                          # React frontend
│       ├── src/
│       │   ├── pages/
│       │   │   ├── auth/             # Login, Register pages
│       │   │   └── admin/
│       │   │       ├── panels/       # Hard-coded content panels (one file per panel)
│       │   │       └── layout/       # Admin shell, sidebar, nav
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
- **Panels**: Each content panel is a standalone page file in `src/pages/admin/panels/`. No dynamic panel engine.
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

---

## 5. Testing Strategy

### Backend
- **Unit tests**: Every usecase tested in isolation with mock repositories (interface-based). Located in `internal/usecase/<name>/<name>_test.go`.
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

### Ask before (require explicit approval)
- Starting any new coding task — confirm scope and approach with the user first.
- Any feature or implementation with multiple valid approaches — present options, do not choose autonomously.
- `MediaInput` opens the OS file picker before uploading to storage.
- Switching storage providers (S3 vs Cloudinary) — confirm which one to implement.

### Never
- Never read, edit, create, or expose `.env` or any environment variable files.
- Never use a drag-and-drop or dynamic form engine — panels are hard-coded React pages only.
- Never use `React.Children.map` or recursive child scanning in `FormProvider`.
- Never couple usecase logic to a specific database — all DB access is behind repository interfaces.
- Never auto-choose an implementation path when multiple options exist — always ask.
- Never add GraphQL or PostgreSQL support until REST + MongoDB are fully complete and the user authorizes the next phase.

---

## Open Questions (resolve before implementing)

1. **Media storage**: AWS S3 or Cloudinary for v1? (Or support both via an interface from day one?)
2. **Go module path**: `personal-cms/api` or a full GitHub path like `github.com/<username>/personal-cms/apps/api`?
3. **Render deployment**: One Render service per Docker service, or a single service running docker-compose?
