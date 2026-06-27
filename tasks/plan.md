# Plan — personal-cms (project-abyssoftime-cms-v2)

> Completed plans archived in `tasks/archive/plan-phases-A-Y.md`. This file tracks only current and upcoming plans.

---

## Current: Phase Z — Bug Fixes (Auth Flow & Component Naming)

See [bugfix-auth-and-naming.md](bugfix-auth-and-naming.md) for full spec (root causes, code samples).

### Dependency Graph

```
B1 (Register redirect)       — independent, FE only
B2 (Session persistence)     — independent, BE only (config + usecase + handler + tests)
B3 (Component table naming)  — independent, BE only (GORM infra)
    ↓
[Checkpoint Z: all three verified]
```

All three are independent — can be done in any order.

### B1: Register → Login Redirect (FE)

- **File:** `apps/web/src/pages/auth/RegisterPage.tsx`
- **Change:** In register mutation `onSuccess`, call `queryClient.invalidateQueries({ queryKey: ['auth-setup'] })` before `navigate('/login')`
- **Verify:** Register first user → lands on `/login`, not redirected back

### B2: Session Persistence (BE)

**Sub-changes:**

1. **Cookie defaults** (`apps/api/internal/config/config.go`):
   - `COOKIE_SECURE`: default `true` → `false`
   - `COOKIE_SAMESITE`: default `none` → `lax`

2. **RefreshToken signature** (`apps/api/internal/usecase/auth/auth_usecase.go`):
   - Before: `(accessToken string, err)` → After: `(accessToken, newRefreshToken string, err)`
   - Generate new refresh token via `pkgjwt.GenerateRefreshToken(user.ID)`

3. **Refresh handler** (`apps/api/internal/delivery/http/handler/auth_handler.go`):
   - Update `AuthUseCase` interface for 3-return signature
   - Re-set refresh cookie with new token in `Refresh()`

4. **Tests** (`auth_usecase_test.go`, `auth_handler_test.go`): Update for new signature

- **Verify:** `go test ./internal/usecase/auth/... ./internal/delivery/http/handler/...` + login → F5 → stays on admin

### B3: Component Table Naming (BE)

- **File:** `apps/api/internal/infrastructure/gormdb/document_repository.go`
- **Change:** `component_` prefix → `components_` (plural, consistent with `documents_`)
- **Verify:** `go test ./internal/infrastructure/gormdb/...`

### Checkpoint Z

1. `make test-api` — all backend tests green
2. `make test-web` — all frontend tests green
3. Manual smoke: register → login → F5 stays logged in
4. Update `tasks/todo.md` — mark all complete

---

## Current: Phase ZZ — Bug Fixes v1.8 (Response Shape, Auth, Inputs, GraphQL)

See [bugfix-v1.8.md](bugfix-v1.8.md) for full plan. Spec: [specs/BUGFIX-SPEC.md](../specs/BUGFIX-SPEC.md).

### Summary

Six bugs across auth, content API, frontend inputs, and response formatting:

| Bug | Scope | Description |
|-----|-------|-------------|
| B1 | BE | User entity missing `id`/`documentId` on registration |
| B2 | FE | Register page re-shown after admin creation |
| B4+B5 | BE+FE | Response shape: merge system+content into `data`, remove `contentTypeId`/`status` from public |
| B3 | FE | JsonInput/RichTextInput data loss on save |
| B6 | BE | GraphQL defaults to draft instead of published |

### Order: B1 → B2 → B4+B5 → B3 → B6

---

## Current: Phase DL — Fix Production Data Loss in EnsureCollection

See [specs/fix-data-loss-ensure-collection.md](../specs/fix-data-loss-ensure-collection.md) for full spec.

### Problem Summary

GORM `EnsureCollection` drops and recreates document/component tables on every API startup. Render.com free tier cold-starts after ~15 min idle, destroying all Supabase data each time. MongoDB's version is non-destructive (index-only), so local dev is unaffected.

### Dependency Graph

```
T1 (shared column introspection helper)
 ├── T2 (document EnsureCollection → non-destructive)
 │    └── T3 (document preservation tests)
 ├── T4 (component EnsureCollection → non-destructive)
 │    └── T5 (component preservation tests)
 └─────── [Checkpoint 1: data safe, all tests green]
              └── T6 (sync startup logging)
                   └── [Checkpoint 2: observability, full suite green]
```

### T1: Shared column introspection helper

**File**: `apps/api/internal/infrastructure/gormdb/document_repository.go`

Add a package-level function `existingColumns(db *gorm.DB, table string) (map[string]bool, error)` that returns a set of column names for a raw table. Uses the cross-database approach:

```go
func existingColumns(db *gorm.DB, table string) (map[string]bool, error) {
    rows, err := db.Table(table).Limit(1).Rows()
    if err != nil { return nil, err }
    defer rows.Close()
    cols, err := rows.Columns()
    if err != nil { return nil, err }
    set := make(map[string]bool, len(cols))
    for _, c := range cols { set[c] = true }
    return set, nil
}
```

Works on both SQLite (tests) and PostgreSQL (production) — no dialect-specific SQL needed.

**Acceptance**: Existing tests still pass. Private function, no interface changes.
**Verify**: `cd apps/api && go test ./internal/infrastructure/gormdb/ -run TestDocument`

### T2: Rewrite document `EnsureCollection` (non-destructive)

**File**: `apps/api/internal/infrastructure/gormdb/document_repository.go`

Replace drop-and-recreate with:
1. If table doesn't exist → `createDocumentTable()` (extract current CREATE logic)
2. If table exists → `addMissingColumns()` using T1's helper, running `ALTER TABLE ADD COLUMN` only for missing field columns

System columns (`gorm_id`, `document_id`, `version`, `locale`, timestamps, `*_by`) are only created with the table, never added after — they're always present from initial creation.

**Acceptance**:
- Data survives EnsureCollection calls
- New fields get columns added
- Removed fields keep their columns (no drops)
- All 13 existing document tests pass unchanged

**Verify**: `cd apps/api && go test ./internal/infrastructure/gormdb/ -run TestDocument`

### T3: Document data-preservation tests

**File**: `apps/api/internal/infrastructure/gormdb/document_repository_test.go`

Add 3 tests:
1. `TestDocumentRepository_EnsureCollection_PreservesData` — insert rows, re-run EnsureCollection, rows queryable
2. `TestDocumentRepository_EnsureCollection_AddsNewColumn` — create with [title,body], insert, re-ensure with [title,body,summary], verify old data + new column
3. `TestDocumentRepository_EnsureCollection_IgnoresRemovedField` — create with [title,body,summary], insert, re-ensure with [title,body], verify summary column still exists

**Acceptance**: All 16 tests pass (13 existing + 3 new)
**Verify**: `cd apps/api && go test ./internal/infrastructure/gormdb/ -run TestDocument -v`

### T4: Rewrite component `EnsureCollection` (non-destructive)

**File**: `apps/api/internal/infrastructure/gormdb/component_repository.go`

Same pattern as T2. Uses the shared `existingColumns` function from T1. Component system columns differ slightly: `gorm_id`, `component_id`, `document_id`, `version`, `locale`, `created_at`, `updated_at`.

**Acceptance**: Same criteria as T2 but for component tables. All 7 existing component tests pass.
**Verify**: `cd apps/api && go test ./internal/infrastructure/gormdb/ -run TestComponent`

### T5: Component data-preservation tests

**File**: `apps/api/internal/infrastructure/gormdb/component_repository_test.go`

Add 3 tests (same pattern as T3):
1. `TestComponentRepository_EnsureCollection_PreservesData`
2. `TestComponentRepository_EnsureCollection_AddsNewColumn`
3. `TestComponentRepository_EnsureCollection_IgnoresRemovedField`

**Acceptance**: All 10 tests pass (7 existing + 3 new)
**Verify**: `cd apps/api && go test ./internal/infrastructure/gormdb/ -run TestComponent -v`

### Checkpoint 1: Data safe

1. `cd apps/api && go test ./internal/infrastructure/gormdb/ -v` — all 26 tests pass
2. `cd apps/api && go test ./...` — full backend suite green
3. Ready to deploy if needed (logging can come later)

### T6: Sync startup logging

**Files**:
- `apps/api/internal/usecase/content_type/sync.go` — add log lines in `syncOne`
- `apps/api/internal/infrastructure/gormdb/document_repository.go` — add `TableInfo` method
- `apps/api/internal/domain/repository/document_repository.go` — add `TableInfo` to interface

Add `TableInfo(ctx, slug) (exists bool, rowCount int64, err error)` to `DocumentRepository`. In `syncOne`, call it before `EnsureCollection` and log:
```
sync: content-type "blog-posts" — table exists=true, rows=42
sync: content-type "about" — table exists=false, creating
```

Update the `fakeDocRepository` mock in `sync_test.go` with a no-op `TableInfo`.

**Acceptance**: Startup shows per-content-type table status. All sync tests pass.
**Verify**: `cd apps/api && go test ./internal/usecase/content_type/ -v` + `make dev-api` (check logs)

### Checkpoint 2 (Final): Full verification

1. `cd apps/api && go test ./...` — all tests green
2. `make dev-api` — startup logs show table status
3. Manual test: create documents → restart API → documents survive
4. Ready for Render.com deploy

---

## Current: Phase AX — Fix Cross-Origin Auth (Production F5 Redirect)

### Context

After logging in with rememberMe=true, refreshing (F5) redirects to login on **production (Render)** with `COOKIE_SECURE=true; COOKIE_SAMESITE=none`. On Render, frontend and API are separate services on different domains. Since `onrender.com` is on the Public Suffix List, these are **different sites**. Browsers (Safari ITP, Chrome third-party cookie deprecation) block the refresh token cookie as a third-party cookie. The fix from Phase Z (cookie defaults + re-issue) only works for same-origin deployments (docker-compose with nginx proxy).

### Fix: Token-based refresh (body transport)

Return refresh token in the response body AND accept it from the request body. Frontend stores in localStorage (rememberMe) or sessionStorage (!rememberMe). Cookie still set for backward compat.

### Dependency Graph

```
T1 (BE: Refresh accepts body token)      ←── independent
T2 (BE: Login+Refresh return refreshToken in body) ← T1
T3 (BE: Update handler tests)            ← T1+T2
    ↓
[Checkpoint 1: go test ./... green]
    ↓
T4 (FE: Token storage helpers in api.ts) ←── independent
T5 (FE: AuthContext — login/mount/logout) ← T4
T6 (FE: LoginPage — pass rememberMe+refreshToken) ← T5
T7 (FE: api.ts — interceptor sends body token) ← T4
T8 (FE: Update tests)                    ← T5+T6+T7
    ↓
[Checkpoint 2: vitest run + tsc --noEmit green]
    ↓
[Checkpoint 3: Manual E2E cross-origin test]
```

### T1: Backend — Refresh handler accepts token from request body

**File:** `apps/api/internal/delivery/http/handler/auth_handler.go`

Change `Refresh()` to read refresh token from JSON body first, falling back to cookie. Use `ShouldBindJSON` (ignore error — body is optional for backward compat).

**Verify:** Existing tests still pass + new body-based flow works.

### T2: Backend — Return refreshToken in response body

**File:** `apps/api/internal/delivery/http/handler/auth_handler.go`

Add `"refreshToken": refresh` to Login and Refresh JSON responses. Cookie still set alongside.

### T3: Backend — Update handler tests

**File:** `apps/api/internal/delivery/http/handler/auth_handler_test.go`

- Login success: assert `refreshToken` in body
- Refresh: add "success with body token" test case (JSON body, no cookie)
- Refresh: assert `refreshToken` in response body

**Checkpoint 1:** `cd apps/api && go test ./... -count=1`

### T4: Frontend — Refresh token storage helpers

**File:** `apps/web/src/lib/api.ts`

Add `storeRefreshToken(token, remember?)`, `getRefreshToken()`, `clearRefreshToken()`. Uses localStorage for rememberMe=true, sessionStorage for false. When `remember` is omitted (during refresh), preserves current storage location.

### T5: Frontend — AuthContext updates

**File:** `apps/web/src/context/AuthContext.tsx`

- `login()` accepts `(accessToken, refreshToken, rememberMe)` — stores refresh token
- Mount effect: read stored refresh token → if absent, set LOGGED_OUT; if present, send in body to `/auth/refresh`
- `logout()`: call `clearRefreshToken()`

### T6: Frontend — LoginPage passes rememberMe + refreshToken

**File:** `apps/web/src/pages/auth/LoginPage.tsx`

- Update `LoginResponse` to include `refreshToken`
- `onSuccess`: call `login(data.accessToken, data.refreshToken, variables.rememberMe)`

### T7: Frontend — api.ts interceptor sends body token

**File:** `apps/web/src/lib/api.ts`

- `refreshAccessToken()`: send stored token in body, store new token from response
- 401 catch: call `clearRefreshToken()`

### T8: Frontend — Update tests

Update `AuthContext.test.tsx`, `LoginPage.test.tsx` for new `login()` signature and `refreshToken` in responses.

**Checkpoint 2:** `cd apps/web && bun run test && npx tsc --noEmit`

### Checkpoint 3: Manual E2E

Cross-origin test: run API on :8080, web on :5173 with `VITE_API_URL=http://localhost:8080` (bypasses vite proxy). Login → F5 → stays on admin. Close browser → reopen → stays if rememberMe, redirects if not.

---

## Current: Phase HP — Background Health Ping Service

Spec: [specs/health-ping.md](../specs/health-ping.md)
Scope: Frontend only (`apps/web/`)
New files: 2 | Modified files: 1

### Dependency Graph

```
ConnectionOverlay.tsx (pure UI, no deps)
       ↑
HealthContext.tsx (imports ConnectionOverlay, uses fetch)
       ↑
main.tsx (wraps app with HealthProvider)
```

The overlay is a leaf node (no project dependencies beyond Tailwind). The context imports the overlay and contains all logic. `main.tsx` is the final integration point.

---

### Phase 1: ConnectionOverlay Component

**Goal:** Build and test the pure UI overlay component in isolation.

#### Task 1.1 — Create `ConnectionOverlay.tsx`

**File:** `apps/web/src/components/ConnectionOverlay.tsx`

**What to build:**
- Named export `ConnectionOverlay` accepting `{ visible: boolean }`
- Fixed full-screen overlay: `z-50`, `bg-background/80 backdrop-blur-sm`
- Centered content: spinner (`animate-spin`), heading "Connecting to service...", subtext about server startup
- Fade in/out transition using Tailwind (`transition-opacity`, conditional `opacity-0 pointer-events-none`)
- Accessibility: `role="alert"`, `aria-live="assertive"`, `aria-busy="true"`

**Acceptance criteria:**
- `visible={true}` → overlay visible, blocks interaction
- `visible={false}` → overlay hidden (`opacity-0`, `pointer-events-none`)
- ARIA attributes present
- No `any` types, named export only

#### Task 1.2 — Test `ConnectionOverlay`

**File:** `apps/web/src/components/__tests__/ConnectionOverlay.test.tsx`

**Test cases:**
1. Renders spinner + text when `visible={true}`
2. Has `opacity-0` / `pointer-events-none` when `visible={false}`
3. Has `role="alert"`, `aria-live="assertive"`, `aria-busy="true"`

**Verify:** `npx vitest run src/components/__tests__/ConnectionOverlay.test.tsx`

#### Checkpoint 1

```bash
cd apps/web && npx tsc --noEmit && npx vitest run src/components/__tests__/ConnectionOverlay.test.tsx
```

---

### Phase 2: HealthContext Provider

**Goal:** Implement the ping loop, state management, and visibility API integration.

#### Task 2.1 — Create `HealthContext.tsx`

**File:** `apps/web/src/context/HealthContext.tsx`

**What to build:**
- `HealthProvider` component wrapping children + `<ConnectionOverlay>`
- State: `isApiHealthy` (initial `true` — optimistic)
- Ref: `timerRef` (stores `setTimeout` ID)
- `pingHealth()` function:
  - Uses `fetch` with `AbortController` (5s timeout) to `GET ${VITE_API_URL}/health`
  - On success (HTTP 200): set healthy, schedule next in 14 min
  - On failure: set unhealthy, schedule retry in 10s
- Effect 1 (mount): call `pingHealth()` immediately
- Effect 2 (visibility): add `visibilitychange` listener
  - `hidden` → clear timer
  - `visible` → immediate `pingHealth()`
- Cleanup: clear timer, remove listener
- Export `useHealthStatus()` hook

**Key constraints:**
- Do NOT use the `api` axios instance (no auth interceptor)
- Do NOT use TanStack Query
- All names 3+ characters
- Named export only

**Acceptance criteria:**
- On mount, fires GET /health immediately
- Healthy → 14m interval; Unhealthy → 10s interval
- State transition: unhealthy→healthy auto-dismisses overlay
- Tab hidden pauses, tab visible resumes
- Uses standalone `fetch`, not `api` client
- Timer cleaned up on unmount

#### Task 2.2 — Test `HealthContext`

**File:** `apps/web/src/context/__tests__/HealthContext.test.tsx`

**Test cases (using `vi.useFakeTimers()` + `vi.stubGlobal('fetch', ...)`):**

1. **Initial healthy state** — renders children, no overlay
2. **Ping failure shows overlay** — mock fetch rejects → overlay visible
3. **Recovery hides overlay** — fail then succeed → overlay disappears
4. **Retry interval on failure** — after failure, next ping at ~10s (advance timers)
5. **Success interval** — after success, next ping at ~14m (advance timers)
6. **Cleanup on unmount** — timer cleared, no dangling effects
7. **Visibility pause** — dispatch `visibilitychange` with `hidden` → timer cleared
8. **Visibility resume** — dispatch `visibilitychange` with `visible` → immediate ping

**Verify:** `npx vitest run src/context/__tests__/HealthContext.test.tsx`

#### Checkpoint 2

```bash
cd apps/web && npx tsc --noEmit && npx vitest run src/context/__tests__/HealthContext.test.tsx
```

---

### Phase 3: Integration

**Goal:** Wire `HealthProvider` into the app tree.

#### Task 3.1 — Update `main.tsx`

**File:** `apps/web/src/main.tsx`

**Changes:**
- Import `HealthProvider` from `@/context/HealthContext`
- Wrap inside `QueryClientProvider`, outside `BrowserRouter`:

```tsx
<QueryClientProvider client={queryClient}>
  <HealthProvider>
    <BrowserRouter>
      <AuthProvider>
        <AppRouter />
      </AuthProvider>
    </BrowserRouter>
  </HealthProvider>
  <Toaster position="top-right" />
  <ReactQueryDevtools initialIsOpen={false} />
</QueryClientProvider>
```

**Rationale for position:**
- Outside `BrowserRouter` because it's route-independent
- Inside `QueryClientProvider` in case future features need query client access from health context
- Overlay renders inside `HealthProvider`, so it covers the entire app including auth pages

**Acceptance criteria:**
- App compiles without errors
- Overlay appears on API down, disappears on recovery
- Existing functionality (auth, routing) unaffected

#### Checkpoint 3 (Final)

```bash
make test-web        # all frontend tests pass
cd apps/web && npx tsc --noEmit   # no type errors
make dev-web         # visual verification in browser
```

Manual browser test:
1. Start only frontend (`make dev-web`) without API → overlay should appear
2. Start API (`make dev-api`) → overlay should auto-dismiss
3. Stop API → overlay reappears within 14 minutes (or 10s if already unhealthy)
4. Switch tab away and back → ping fires immediately on return

---

### Risk Assessment

| Risk | Mitigation |
|---|---|
| Overlay flash on fast connections | Optimistic initial `true` — only show overlay after confirmed failure |
| Timer leak on fast navigation | `useEffect` cleanup clears `setTimeout` + removes event listener |
| Auth interceptor interference | Standalone `fetch` with no headers |
| Render.com CORS blocking `/health` | `/health` is already public and CORS-configured in router |
| Tests flaky with real timers | Use `vi.useFakeTimers()` exclusively |

---

## Current: Phase GQ — Migrate GraphQL to gqlgen (Full Codegen Pipeline)

Spec: [specs/graphql-library-evaluation.md](../specs/graphql-library-evaluation.md)
Scope: Backend only (`apps/api/`). No frontend changes.

### Summary

Replace `graphql-go/graphql` (dormant) with **gqlgen** (build-time codegen). A custom `gqlcodegen` tool reads `content-types/*.json` and generates `.graphql` schema files + resolver implementations. Then gqlgen generates the Go execution runtime. One `make graphql-generate` command does everything.

**Key metrics:** ~1,089 lines of runtime code → ~345 (-68%). 2.7x performance improvement.

### Dependency Graph

```
T1 (gqlgen dep + dirs)
 └─ T2 (gqlgen.yml + model/types.go)
     └─ T3 (gqlcodegen: SDL generation)
         └─ T4 (gqlcodegen: resolver + yml generation + tests)
             ─── [Checkpoint 1: gqlcodegen produces correct outputs]
                  └─ T5 (first codegen run + resolver.go, content_types.go)
                      ├─ T6 (media.go)           ── independent ──┐
                      ├─ T7 (filter.go + orderBy) ── independent ──┤
                      └─ T8 (document_helpers.go) ── independent ──┘
                          └─ T9 (handler.go)
                              ─── [Checkpoint 2: go build passes]
                                   └─ T10 (main.go integration)
                                       └─ T11 (remove old code)
                                           └─ T12 (.gitignore + Makefile)
                                               ─── [Checkpoint 3: go vet + go test + go build]
                                                    ├─ T13 (resolver tests)  ── independent ──┐
                                                    ├─ T14 (filter tests)    ── independent ──┤
                                                    └─ T15 (rules + docs)    ── independent ──┘
                                                        ─── [Checkpoint 4: full E2E]
```

---

### Phase 1: Codegen Pipeline (Build-Time)

#### T1: Add gqlgen dependency + directory structure

**Files:** `apps/api/tools.go` (new), directories `graphql/{schema,generated,model,resolver}/`
**Commands:** `cd apps/api && go get github.com/99designs/gqlgen@latest`
**Acceptance:** `go build ./...` still passes.
**Verify:** `cd apps/api && go build ./...`

#### T2: Create gqlgen configuration + model types

**Files:**
- `apps/api/graphql/gqlgen.yml` — schema/exec/model/resolver paths, JSON→Map, Time→Time scalars
- `apps/api/graphql/model/types.go` — `DocumentMap`, `MediaAssetMap`, `ContentTypeMap` type aliases

**Acceptance:** Valid YAML/Go.
**Verify:** `cd apps/api && go vet ./graphql/model/...`

#### T3: Build gqlcodegen tool — SDL generation

**File:** `apps/api/cmd/gqlcodegen/main.go` (new)

Adapt `schema_builder.go` logic to write individual `.graphql` files. Key additions vs current:
- `base.graphql` includes `directive @auth on FIELD_DEFINITION` + `type Mutation { _empty: Boolean }`
- Per-type `.graphql` adds `status: String` to ALL query args (missing in current schema_builder)
- Per-type `.graphql` adds `@auth` to ALL mutation fields

**Reuse:** `slugToPascalCase`, `slugToCamelCase`, `fieldTypeToGraphQL`, `writeComponentType`, `writeFilterType`, `writeOrderByType` from `schema_builder.go`.

**Acceptance:** `go run ./cmd/gqlcodegen --phase=schema` produces correct `.graphql` files.

#### T4: Build gqlcodegen tool — resolver + config generation + tests

**File:** `apps/api/cmd/gqlcodegen/main.go` (extend) + `main_test.go` (new)

- Generate `graphql/resolver/content_gen.go` — field definitions + thin resolver methods
- Inject models section into `gqlgen.yml`
- `--phase=schema` / `--phase=resolvers` flags
- Tests: SDL output, resolver output, models injection

**Acceptance:** `go run ./cmd/gqlcodegen` produces all 3 artifacts.
**Verify:** `cd apps/api && go test ./cmd/gqlcodegen/...`

#### Checkpoint 1

```bash
cd apps/api && go run ./cmd/gqlcodegen --phase=schema  # verify .graphql files
cd apps/api && go test ./cmd/gqlcodegen/...
```

---

### Phase 2: gqlgen Generation + Hand-Written Runtime

#### T5: First codegen run + root resolver + content_types

Bootstrap pipeline: `gqlcodegen --phase=schema` → `gqlgen generate` → remove stubs → `gqlcodegen --phase=resolvers`

**Hand-written files:**
- `graphql/resolver/resolver.go` — Resolver struct, NewResolver, Query()/Mutation() methods
- `graphql/resolver/content_types.go` — contentTypes query + _empty mutation stub

#### T6: Media field resolution

**File:** `graphql/resolver/media.go` — adapt from `resolver_factory.go:571-660`
- `docToMap`, `resolveMediaField`, `resolveComponentMedia`, `resolveComponentMap`
- Receiver changes from `ResolverFactory` to `Resolver` (methods, not closures)

#### T7: Filter + OrderBy conversion

**File:** `graphql/resolver/filter.go`
- `convertFilterStructs[T any](filters []*T) []entity.FilterNode` — reflection-based
- `extractOrderBy(orderByArg any) (string, int)` — first non-nil field, default `("createdAt", -1)`

#### T8: Document CRUD helpers

**File:** `graphql/resolver/document_helpers.go`
- Collection: `getDocument`, `getDocumentList`, `createDocument`, `updateDocument`, `deleteDocument`, `publishDocument`, `unpublishDocument`
- Single: `getSingleType`, `saveSingleType`, `publishSingleType`, `unpublishSingleType`
- Helper: `structToMap(v any) map[string]any` (json roundtrip, same as current `inputToMap`)

#### T9: HTTP handler

**File:** `graphql/handler.go`
- `NewHandler(resolver, tokenValidator) http.Handler`
- Handler-level auth (JWT + access token fallback) — same as current
- `@auth` directive: `skip_runtime: true` in gqlgen.yml (auth enforced at handler level)
- Error presenter: maps domain errors to GraphQL error codes

#### Checkpoint 2

```bash
cd apps/api && make graphql-generate && go build ./...
```

---

### Phase 3: Server Integration + Cleanup

#### T10: Update main.go

Replace `dynamic.NewResolverFactory` + `BuildHandler` with `resolver.NewResolver` + `graphqlhandler.NewHandler`. Update imports.

#### T11: Remove old code + dependencies

Delete `graphql/dynamic/` directory (5 files, ~1,760 lines). Run `go mod tidy`.

#### T12: Update .gitignore + Makefile

`.gitignore`: add generated file patterns. Makefile: update `graphql-generate` to 4-step pipeline.

#### Checkpoint 3

```bash
cd apps/api && go vet ./... && go test ./... && go build ./...
```

---

### Phase 4: Tests + Documentation

#### T13: Handler + resolver tests

Adapt `resolver_factory_test.go` patterns: contentTypes query, single-type query, collection list, mutation auth.

#### T14: Filter conversion tests

New tests for `convertFilterStructs` and `extractOrderBy` — nil, eq, combinators, in/notIn, field name conversion.

#### T15: Update rules + CLAUDE.md

- `rules/content.md` §6: update for gqlgen architecture
- `CLAUDE.md` commands table: add `make graphql-generate`
- `SPEC.md`: mark GraphQL migration complete

#### Checkpoint 4 (Final)

```bash
cd apps/api && go vet ./... && go test ./... && go build ./...
make graphql-generate && cd apps/api && go build ./...  # idempotency
make dev-api  # manual E2E testing
```

---

## Upcoming

*(Add new plans here as they are defined.)*
