# Plan — personal-cms (project-abyssoftime-cms-v2)

> Completed plans archived in `tasks/archive/plan-phases-A-Y.md` and `tasks/archive/plan-phases-Z-to-GQ.md`. This file tracks only current and upcoming plans.

---

## Current: Phase BC — Bulk Create + Publish (Collection-Type Documents)

Spec: [specs/bulk-document-create-publish.md](../specs/bulk-document-create-publish.md).
Full plan detail: [/Users/hungnguyenhuy/.claude/plans/clever-juggling-goblet.md](/Users/hungnguyenhuy/.claude/plans/clever-juggling-goblet.md) (approved plan-mode output).

### Context

Seeding N collection-type documents (e.g. `en-vocab` entries) today requires N create requests + N publish requests. Adding one endpoint, `POST /api/document-manager/collection-type/:slug/bulk`, that creates and immediately publishes up to 100 items in a single call, all-or-nothing.

### Architecture Decision

**Rollback, not pre-validation.** The codebase has no "required field" validation concept (`entity.FieldDefinition` has no `Required` flag), and the only existing validation (component shape checking) is interleaved with the actual writes inside `saveTopLevelComponents`/`saveNestedComponents` — it can't be cleanly extracted into a separate validate-first pass without duplicating logic. So `BulkCreateAndPublish` processes items sequentially through the existing `Save()` → `Publish()` methods unchanged; on the first failure, it rolls back everything already committed in this batch via the existing `Delete()` usecase method (confirmed: `Delete()` already removes both draft + published records + components for every locale — a ready-made rollback primitive, no new deletion logic needed).

### Dependency Graph

```
T1 (Usecase: BulkCreateAndPublish + rollback + tests)
    ↓
[Checkpoint 1: usecase tests green]
    ↓
T2 (Handler: BulkCreateCollection + route + dual-permission middleware + tests)
    ↓
[Checkpoint 2: go vet + test-api + build green, manual curl smoke test]
    ↓
T3 (Docs: update rules/document.md §2.4 + §1.4)
    ↓
[Checkpoint 3 (Final): all acceptance criteria met]
```

### T1: Usecase — `BulkCreateAndPublish`

**File:** `apps/api/internal/usecase/document/document_usecase.go`

Add method: resolves locale once, loops `itemsData []map[string]any`, calls `Save` then `Publish` per item in order, rolls back via `Delete` on the first failure (including the failing item's own draft if `Save` succeeded but `Publish` failed), returns `([]*entity.Document, error)` with the error wrapped as `item[N]: ...` (0-based index).

**Verify:** `cd apps/api && go test ./internal/usecase/document/... -run BulkCreateAndPublish -v`

### T2: Handler + route — `BulkCreateCollection`

**Files:** `apps/api/internal/delivery/http/handler/document_handler.go`, `apps/api/internal/delivery/http/router.go`

- `bulkCreateRequest{Items []documentRequest}` / `bulkCreateResponse{Items []documentResponse}` types
- Handler validates `1 <= len(items) <= 100` before calling usecase (400 otherwise)
- Route: `colGroup.POST("/:slug/bulk", middleware.GinRequirePermission(cache, "content:create"), middleware.GinRequirePermission(cache, "content:publish"), cfg.DocHandler.BulkCreateCollection)` — chaining two `GinRequirePermission` calls for an AND of both permissions (confirmed safe: standard Gin middleware, no new permission constant needed)

**Verify:** `cd apps/api && go test ./internal/delivery/http/handler/... -run BulkCreateCollection -v`

### T3: Documentation

**File:** `rules/document.md`

Add the new route to `§2.4 REST Routes — Collection-Type Documents` and a note under `§1.4 Collection-Type Rules` describing bulk semantics (max 100 items, all-or-nothing via rollback, one locale per request) — required in the same task per `rules/GLOBAL.md §12`.

### Checkpoint (Final)

```bash
cd apps/api && go vet ./... && make test-api && go build ./...
```

Manual smoke test: `make dev-api`, then `POST /api/document-manager/collection-type/en-vocab/bulk?locale=en` with 2-3 valid items → 201, all published; then one batch with a malformed `meanings` field (object instead of array) → 400, zero new drafts in the list endpoint.

---

## Upcoming

*(Add new plans here as they are defined.)*
