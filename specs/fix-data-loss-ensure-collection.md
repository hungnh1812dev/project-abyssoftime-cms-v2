# SPEC — Fix Production Data Loss in EnsureCollection

**Status**: Draft
**Module**: core (infrastructure/gormdb)
**Priority**: Critical — data loss in production on every cold start

---

## Problem

All document and component data is wiped from the Supabase PostgreSQL database every time the Render.com API service restarts (cold start after ~15 min idle on free tier, or on every deploy).

### Root Cause

`EnsureCollection` in both `gormdb/document_repository.go:67-99` and `gormdb/component_repository.go:32-58` unconditionally **drops and recreates** the table on every call:

```go
if r.db.Migrator().HasTable(table) {
    r.db.Migrator().DropTable(table)  // ← destroys all data
}
// CREATE TABLE ...
```

This is called on every API startup via the sync chain:
`main.go:192` → `Syncer.Sync()` → `syncOne()` → `docRepo.EnsureCollection()`.

MongoDB's `EnsureCollection` is non-destructive (just ensures an index), so local development with MongoDB never triggers this bug.

### Why Production-Only

| Environment | DB Driver | EnsureCollection behavior | Cold starts |
|---|---|---|---|
| Local dev | MongoDB | Non-destructive (index only) | Rare (air hot-reload) |
| Render.com | PostgreSQL (Supabase) | **Destructive (DROP + CREATE)** | Every ~15 min idle |

---

## Objective

Make `EnsureCollection` non-destructive for PostgreSQL: preserve existing data while ensuring the table schema matches the content-type definition. Add startup logging to detect future schema-sync issues.

---

## Acceptance Criteria

1. **No data loss on restart**: Existing rows survive API cold starts.
2. **New fields are added**: When a field is added to a content-type JSON definition, the corresponding column is added to the table (nullable, no default required).
3. **Removed fields are ignored**: Old columns stay in the table when a field is removed from the definition. No column drops.
4. **Existing tests still pass**: All `document_repository_test.go` and `component_repository_test.go` tests pass with no changes (they exercise EnsureCollection).
5. **Startup logging**: On startup, log each content-type table's existence and row count before sync runs.
6. **MongoDB unaffected**: The MongoDB `EnsureCollection` implementation is already correct; no changes needed.

---

## Approach

### 1. Document Repository — `EnsureCollection`

**File**: `apps/api/internal/infrastructure/gormdb/document_repository.go`

Change from drop-and-recreate to:

```
IF table does NOT exist → CREATE TABLE (same as current)
IF table exists → for each field in the definition:
    IF column does NOT exist → ALTER TABLE ADD COLUMN <name> <type>
    (skip columns that already exist — no type changes, no drops)
```

Pseudocode:
```go
func (r *documentRepository) EnsureCollection(ctx context.Context, contentTypeSlug string, fields []entity.FieldDefinition) error {
    table := documentTableName(contentTypeSlug)
    if !r.db.Migrator().HasTable(table) {
        // Create table with all system columns + field columns (same as current)
        return r.createTable(ctx, table, fields)
    }
    // Table exists — only add missing columns
    return r.addMissingColumns(ctx, table, fields)
}
```

`addMissingColumns` checks `r.db.Migrator().HasColumn()` (or queries `information_schema.columns` for PostgreSQL) for each field and runs `ALTER TABLE ADD COLUMN` only for missing ones.

### 2. Component Repository — `EnsureCollection`

**File**: `apps/api/internal/infrastructure/gormdb/component_repository.go`

Same pattern as document repository: create if missing, add columns if existing.

### 3. Startup Sync Logging

**File**: `apps/api/internal/usecase/content_type/sync.go`

Add a `log.Printf` in `syncOne` before calling `EnsureCollection`, reporting:
- Content type slug
- Whether the table already exists
- Row count (if table exists)

This provides visibility into what the sync is doing on each startup.

### 4. Helper: Column Existence Check

**File**: `apps/api/internal/infrastructure/gormdb/document_repository.go` (or shared helper)

For PostgreSQL, query `information_schema.columns` to check if a column exists:
```sql
SELECT column_name FROM information_schema.columns
WHERE table_name = $1 AND table_schema = 'public'
```

Compare against the expected columns from the content-type definition. Only `ADD COLUMN` for missing ones.

For SQLite (tests), use `PRAGMA table_info(<table>)`.

---

## Files to Change

| File | Change |
|------|--------|
| `apps/api/internal/infrastructure/gormdb/document_repository.go` | Rewrite `EnsureCollection` to be non-destructive |
| `apps/api/internal/infrastructure/gormdb/component_repository.go` | Rewrite `EnsureCollection` to be non-destructive |
| `apps/api/internal/infrastructure/gormdb/document_repository_test.go` | Add test: EnsureCollection preserves existing rows |
| `apps/api/internal/infrastructure/gormdb/component_repository_test.go` | Add test: EnsureCollection preserves existing rows |
| `apps/api/internal/usecase/content_type/sync.go` | Add startup logging (table existence + row count) |

---

## Test Plan

### Unit Tests (SQLite in-memory)

1. **EnsureCollection — create from scratch**: Table doesn't exist → created with correct columns. *(existing test, should still pass)*
2. **EnsureCollection — preserves data**: Insert rows → call EnsureCollection again → rows still exist.
3. **EnsureCollection — adds new column**: Create table with fields A, B → call EnsureCollection with fields A, B, C → column C exists, rows intact.
4. **EnsureCollection — ignores removed field**: Create table with fields A, B, C → call EnsureCollection with fields A, B → column C still exists, rows intact.
5. **Component EnsureCollection**: Same 4 test cases as above for component tables.
6. **DropCollection**: Still works as before (only called when a content type is removed from definitions).

### Manual Verification on Render

1. Deploy the fix to Render
2. Create documents via the CMS UI
3. Wait for the service to spin down (~15 min idle)
4. Access the CMS again (triggers cold start)
5. Verify documents are still present
6. Check Render logs for the new sync logging output

---

## Boundaries

### Always
- Use `CREATE TABLE IF NOT EXISTS` or check `HasTable` before creating
- Use `ALTER TABLE ADD COLUMN` for new fields only
- Preserve all existing data on startup
- Log sync operations for observability

### Never
- Drop existing columns (even for removed fields)
- Change column types on existing columns
- Modify the MongoDB `EnsureCollection` implementation
- Add migration versioning or tracking (overkill for this CMS)

---

## Out of Scope

- Column type changes (field type changed in definition) — not needed yet
- Column renames — not needed yet
- Migration versioning system — overkill for this use case
- Render tier upgrade — the fix should work regardless of hosting tier
