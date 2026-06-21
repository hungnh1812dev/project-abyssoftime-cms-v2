# Plan — Sync Table Field Names & Schema Alignment

Spec: [specs/sync-table-fields.md](../specs/sync-table-fields.md)

---

## Dependency Graph

```
Task 1 (column rename + UUID)
  ├─→ Task 2 (media cleanup)
  │     ├─→ Task 5 (GraphQL media objects)
  │     └─→ Task 6 (frontend MediaInput)
  └─→ Task 3 (entity cleanup)
        └─→ Task 4 (per-field columns)
              └─→ Task 5 (GraphQL)
```

---

## Task 1: Static Table Column Rename (`id` → `gorm_id`) + UUID Standardization

Rename PK column in 6 entity gorm tags. Fix raw SQL queries. Replace `generateDocID()` with UUID.

**Entity tags** (`gorm:"column:id;primaryKey"` → `gorm:"column:gorm_id;primaryKey"`):
- `entity/user.go`, `entity/role.go`, `entity/media_asset.go`, `entity/invite.go`, `entity/content_type.go`, `entity/access_token.go`

**Repository SQL** (`"id = ?"` → `"gorm_id = ?"`):
- `gormdb/user_repository.go` (FindByID, FindByIDs, Delete)
- `gormdb/content_type_repository.go` (FindByID, Delete)
- `gormdb/media_asset_repository.go` (FindByID, Delete)
- `gormdb/invite_repository.go` (Delete)
- `gormdb/access_token_repository.go` (Delete, UpdateLastUsed)

**UUID fix**: `usecase/role/role_usecase.go` — replace `generateDocID()` with `uuid.New().String()`

**Verify**: `cd apps/api && go build ./... && go test ./...`

---

## Task 2: MediaAsset Cleanup + FindByDocumentID

Remove `ContentTypeID`, `DocumentRef` from entity. Remove `FindByDocumentRef`/`DeleteByDocumentRef`. Add `FindByDocumentID`. Remove `DeleteByDocumentRef` call in document usecase Delete(). Clean up media upload params. Update frontend types.

**Verify**: `cd apps/api && go build ./... && go test ./...` + `cd apps/web && npx tsc --noEmit`

---

## Task 3: Document & Component Entity Cleanup

- Document: remove `ContentTypeID`, rename `Data` → `Fields`, `PublishedAt` → `*time.Time`
- Component: remove `Order`, rename `Data` → `Fields`
- Update all references: usecase, handlers, GraphQL resolver, gRPC service
- Component repo: `gorm_id ASC` ordering instead of `"order" ASC`

**Verify**: `cd apps/api && go build ./... && go test ./...`

---

## CHECKPOINT A

```bash
cd apps/api && go build ./... && go test ./...
cd apps/web && npx tsc --noEmit
```

---

## Task 4: Per-Field Dynamic Columns (Repository Rewrite)

`EnsureCollection` takes `[]FieldDefinition`, drops+creates table with per-field columns via raw SQL. CRUD uses column maps instead of struct serialization.

- Document repo: CREATE TABLE with system + field columns, map-based insert/query
- Component repo: same pattern
- Syncer passes field defs to EnsureCollection
- Column types: `text`/`richtext` → TEXT, `media` → VARCHAR, `number` → REAL, `boolean` → BOOLEAN, `json` → TEXT
- Entity tags: `Fields` becomes `gorm:"-"`

**Verify**: `cd apps/api && go build ./... && go test ./...`

---

## Task 5: GraphQL — Media as Object + Remove Response Wrappers

- Add `MediaAsset` object type to schema
- Media fields resolve via `mediaRepo.FindByDocumentID` → return object
- Remove `*Response`/`*ListResponse` wrapper types
- Single queries return type directly, list returns `[Type!]!`
- Component media sub-fields also resolve to objects
- Add `MediaAssetRepository` dependency to `ResolverFactory`

**Verify**: `cd apps/api && go build ./... && go test ./graphql/...`

---

## Task 6: Frontend — MediaInput documentId + Aspect Ratio

- `handleSelect`: store `asset.documentId` instead of URL
- Local state for preview URL display
- Fix container: remove fixed height, use `h-auto max-h-40 object-contain`

**Verify**: `cd apps/web && npx tsc --noEmit && npx vitest run`

---

## CHECKPOINT B (Final)

```bash
cd apps/api && go build ./... && go test ./...
cd apps/web && npx tsc --noEmit && npx vitest run
```
