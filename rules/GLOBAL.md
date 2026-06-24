# RULES — Global (Project-Wide)

These rules apply to **every module, every file, every spec, every plan** in this project. Module-specific rules extend these — they never override them. If a module rule conflicts with a global rule, **ask the user** before proceeding.

---

## 1. Architecture — Clean Architecture (Strict)

### 1.1 Layer Dependencies (Inward Only)
- `delivery` (handlers, middleware, resolvers) → imports `usecase` + `domain`
- `usecase` → imports only `domain` (entity, repository interfaces, pkg/errors, pkg/jwt)
- `domain` → imports nothing from the project (only stdlib + third-party types)
- `infrastructure` → imports `domain` (implements repository interfaces)
- **NEVER** let `usecase` import `delivery`, `infrastructure`, or any DB-specific package
- **NEVER** let `domain` import any other project package

### 1.2 Cross-Module Communication
- All cross-module communication goes through interfaces defined in `domain/repository/`
- **NEVER** import another module's usecase package directly
- Dependencies between modules are injected in `cmd/server/main.go`
- If module A needs module B's logic, A receives B's usecase as a constructor parameter (interface-typed)

### 1.3 Multi-Protocol Delivery
- REST (Gin), gRPC, and GraphQL all delegate to the **same** usecase methods
- **NEVER** duplicate business logic across delivery layers
- Handlers/resolvers are thin adapters: parse input → call usecase → format output

---

## 2. Code Style — Go Backend

### 2.1 Naming
- `PascalCase` for exported identifiers; `camelCase` for unexported
- Repository interfaces: `<Entity>Repository` (e.g., `DocumentRepository`, `UserRepository`)
- Usecase structs: `UseCase` inside their package (e.g., `auth.UseCase`, `document.UseCase`)
- Handler structs: `<Entity>Handler` (e.g., `AuthHandler`, `DocumentHandler`)
- All variable/function names must be **3+ characters** — never use 1-2 letter names (except `i` in for-loops and `ok` in type assertions)

### 2.2 Error Handling
- Wrap errors at the usecase boundary with context
- Handlers map domain errors (`pkg/errors`) to protocol-specific status codes
- **NEVER** return naked `error` strings in HTTP responses
- **NEVER** reflect user input verbatim in error messages (information leakage)
- Error mapping table:

| Domain Error | HTTP | gRPC |
|---|---|---|
| `ErrNotFound` | 404 | `codes.NotFound` |
| `ErrValidation` | 400 | `codes.InvalidArgument` |
| `ErrConflict` | 409 | `codes.AlreadyExists` |
| Other | 500 | `codes.Internal` |

### 2.3 Database Isolation
- MongoDB-specific code (ObjectID, bson primitives) → **only** in `infrastructure/mongodb/`
- GORM-specific code (gorm.DB, gorm.Model) → **only** in `infrastructure/gormdb/`
- **NEVER** use DB-specific types in entity structs beyond tags
- **NEVER** hard-code SQL dialect-specific queries in GORM repositories

### 2.4 General
- No `init()` functions in business logic packages
- Cascade deletion implemented in usecase layer, **not** as DB-level triggers
- Configuration loaded once at boot into typed `Config` struct — no `os.Getenv` anywhere else
- All entity fields use dual tags: `bson` + `gorm` + `json`

---

## 3. Code Style — React Frontend

### 3.1 TypeScript
- Strict mode enabled — no `any` type
- Named exports only (no `export default`)
- Path aliases: `@/components`, `@/lib`, `@/hooks`, `@/types`, `@/pages`

### 3.2 Component Library & Styling
- TailwindCSS + Shadcn UI for all components
- Shadcn components are copied into `src/components/ui/` — owned, not imported
- Follow Radix UI patterns for accessibility (ARIA, keyboard nav, focus management)

### 3.3 Server State
- TanStack Query for **all** server state — no raw `useEffect` + `useState` for API calls
- Every `useMutation` **must** call `queryClient.invalidateQueries` on success with affected query keys
- Query keys follow the pattern: `['entity', 'action', ...params]`

### 3.4 Forms
- `FormProvider` manages loading, submitting, isDirty states automatically
- Dot-notation field names auto-deserialize to nested JSON on submit
- **NEVER** use `React.Children.map` or recursive child scanning in `FormProvider`
- **NEVER** use drag-and-drop or dynamic form engine

### 3.5 Routing
- React Router for client-side routing (SPA)
- Panels are hard-coded React pages — no dynamic form engine
- Custom panels registered as routes **before** generic catch-all routes

---

## 4. Testing Strategy

### 4.1 Backend Tests
- **Usecase tests**: Mock repository interfaces (from `domain/repository/mock/`), ≥80% coverage
- **Handler tests**: `httptest` with `gin.SetMode(gin.TestMode)` — **always** set test mode
- **Infrastructure tests**: Real DB in Docker (MongoDB) or in-memory SQLite (GORM)
- **GORM repos must pass the same test cases as MongoDB counterparts**
- **NEVER** mock the DB directly in usecase tests — mock the repository interface instead

### 4.2 Frontend Tests
- Vitest + React Testing Library + MSW for API mocking
- Test user behavior, not implementation details
- Every new component needs a test file in a co-located `__tests__/` directory

### 4.3 CI Pipeline
- `go vet ./...` → `go test ./...` → `npm run lint` → `npm run build`
- All four must pass before merge

---

## 5. Draft/Publish Workflow (Cross-Module)

- Every content entry = two separate records: `draft` + `published` (version field)
- **Save** upserts `draft` only — **NEVER** touches `published`
- **Publish** copies `draft.data` → `published` record
- **Unpublish** deletes `published` record
- Entry `status` is **computed, never stored**: `draft` | `modified` | `published`
- Public read API returns `published` only — **NEVER** expose draft data publicly
- System fields (`createdAt`, `updatedAt`, `publishedAt`, `createdBy`, `updatedBy`, `publishedBy`, `locale`) are injected automatically

---

## 6. Schema-as-Code

- Content-type structure defined in JSON files under `apps/api/content-types/*.json`
- **NEVER** create/edit/delete ContentType structure via API or UI
- Sync is one-directional: JSON → DB (never DB → JSON)
- Sync runs on every API startup
- JSON schemas declare only content fields — system fields injected automatically
- `listFields` is **NOT** part of JSON schemas — it is managed exclusively via the UI

---

## 7. Database — Per-Entity Selection

- Each entity can use MongoDB or GORM (PostgreSQL/MySQL), configured independently
- Both database clients initialized at startup if needed
- Repository factory pattern in `main.go` selects implementation based on env var
- **NEVER** assume a specific database engine in usecase layer

---

## 8. Security

- JWT access token (15 min) in memory + refresh token (7 days) in HttpOnly cookie
- **NEVER** use `Access-Control-Allow-Origin: *` — always explicit origin whitelist
- **NEVER** set `COOKIE_SAMESITE=lax` for cross-origin deployment
- Refresh token cookie: `HttpOnly: true`, `Secure: true` (production), `SameSite: None` (cross-origin)
- All auth routes rate-limited
- Security headers on every response: `X-Content-Type-Options`, `X-Frame-Options`, `X-XSS-Protection`, `Referrer-Policy`, `CSP`

---

## 9. Environment & Configuration

- **NEVER** read, edit, create, or expose `.env` files
- All env vars documented in `specs/core.md §7`
- `VITE_API_URL` set at build time for frontend
- `CORS_ORIGINS` must be explicit, never wildcard

---

## 10. Conflict Resolution Protocol

When a conflict arises between:
- **Module rule vs Global rule** → Ask user
- **Spec instruction vs Existing rule** → Ask user
- **Multiple valid approaches** → Present options, don't choose autonomously
- **New feature vs Existing boundary** → Ask user before proceeding

**Always** ask before starting any new coding task — confirm scope and approach.

---

## 11. Spec/Plan Reference Protocol

Before writing any spec, plan, or code for a module:
1. Read `rules/GLOBAL.md` (this file)
2. Read the module-specific rule file (`rules/<module>.md`)
3. Read the module spec (`specs/<module>.md`)
4. Cross-check for conflicts between the three
5. If conflict found → ask user before proceeding
6. If no rule covers the situation → check related module rules, then ask user
