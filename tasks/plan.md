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

## Upcoming

*(Add new plans here as they are defined.)*
