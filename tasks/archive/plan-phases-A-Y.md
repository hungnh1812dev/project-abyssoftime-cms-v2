# Archive — Plan: Phases A–Y Migration

> All phases completed. Archived from `tasks/plan.md`.

---

# Plan — Migrate personal-cms to Draft/Publish + Schema-as-Code + Multi-Storage

## Context

`SPEC.md` was extended with three features that this codebase does not implement yet:

1. Draft/publish as two records per entry (`entryId` + computed status), with audit
   fields (`createdBy`/`updatedBy`/`publishedBy`/`locale`).
2. Content-type **structure** defined as JSON files synced into Mongo on boot — never
   created via API/UI.
3. Dual media storage (S3 + Cloudinary) behind the existing `StorageAdapter` interface.

> **Status: All phases complete (A–D, W, X, Y).**

---

## Phase Y — Document Manager API Restructure & Paginated Collections (SPEC §10)

### Context

The document manager API uses flat routes (`/api/document-manager/{slug}`) for all content
types. This causes: no pagination on collection lists, single-type endpoints returning arrays,
and list views fetching full document data. Phase Y restructures routes by kind, adds
server-side pagination, and introduces field projection via `listFields` in content-type
JSON schemas.

**Key de-risking insight**: Go 1.22 ServeMux literal segments (`single-type`,
`collection-type`) take precedence over wildcards (`{slug}`), so new routes can coexist
with old routes during migration.

### Dependency Graph

```
Task Y1 (entity/schema/JSON)
    ↓
Task Y2 (repository methods)
    ↓
Task Y3 (single-type usecase) ←→ Task Y4 (collection usecase)
    ↓                                  ↓
    └────────→ Task Y5 (handler+routes) ←─┘
                    ↓
            [Checkpoint A: BE complete]
                    ↓
              Task Y6 (FE types+hooks)
                    ↓
              Task Y7 (FE components)
                    ↓
            [Checkpoint B: full stack]
                    ↓
              Task Y8 (cleanup+verify)
```

### Y1: BE — Entity + Schema + JSON (`listFields` foundation)

**Files:** `entity/content_type.go`, `usecase/content_type/schema_loader.go`,
`usecase/content_type/sync.go`, `content-types/blog-posts.json`

- Add `ListFields []string` to `ContentType` entity (json + bson tags)
- Add `ListFields` to `ContentTypeDefinition` in schema loader
- Validate each `listFields` entry exists in `fields` — fatal on mismatch
- Carry `ListFields` through sync (`syncOne`)
- Add `"listFields": ["title", "slug", "featured"]` to `blog-posts.json`

### Y2: BE — Repository (paginated + batch methods)

**Files:** `repository/document_repository.go`, `repository/mock/document_repository.go`,
`infrastructure/mongodb/document_repository.go`

- `FindDraftsByContentTypePaginated(ctx, slug, start, size, locale) → (docs, total, err)`
- `FindPublishedByDocumentIDs(ctx, slug, documentIDs, locale) → (docs, err)`
- MongoDB: paginated Find + CountDocuments; batch $in query for published

### Y3: BE — Single-type usecase + tests

**Files:** `usecase/document/document_usecase.go`, `document_usecase_test.go`

- `GetSingleType(ctx, slug, locale)` → find single draft, compute status, ErrNotFound if none
- `SaveSingleType(ctx, slug, data, locale, userID)` → find-or-create, delegate to Save
- `PublishSingleType(ctx, slug, locale, userID)` → find draft, delegate to Publish
- `UnpublishSingleType(ctx, slug, locale)` → find draft, delegate to Unpublish

### Y4: BE — Collection-type paginated usecase + tests

**Files:** `usecase/document/document_usecase.go`, `document_usecase_test.go`

- `GetAllPaginated(ctx, slug, start, size, locale)` → paginated drafts + batch status computation

### Y5: BE — Handler rewrite + route migration + tests

**Files:** `handler/document_handler.go`, `document_handler_test.go`, `cmd/server/main.go`

- Add `ctUC` dependency to `DocumentHandler` (for ListFields lookup)
- 11 new handler methods (4 single-type + 7 collection-type)
- `ListCollection`: fetch content-type for ListFields, paginate, project data
- `projectData(data, fields)` helper
- Replace all flat routes with kind-prefixed routes in main.go
- Keep public route unchanged

### Checkpoint A: BE Complete

`go test ./...` all pass. Curl: single-type 404/200, collection paginated, old routes 404.

### Y6: FE — Types + hooks

**Files:** `types/cms.ts`, new `hooks/useSingleTypeDocuments.ts`,
`hooks/useCollectionDocuments.ts`, `hooks/useLocales.ts`, delete `hooks/useDocuments.ts`

- `ContentType.ListFields`, `PaginatedResponse<T>` types
- Single-type hooks: query (404→undefined), save, publish, unpublish
- Collection hooks: paginated list, detail, create, update, delete, publish, unpublish

### Y7: FE — Components (ContentTypePanel + CollectionListPage)

**Files:** `ContentTypePanel.tsx`, `CollectionListPage.tsx`, `content-type-registry/index.ts`

- ContentTypePanel: kind-aware hook selection (single vs collection-detail)
- CollectionListPage: paginated query, schema-derived columns from ListFields + Fields,
  Previous/Next controls, "Showing X–Y of Z"
- Registry `columns` becomes optional override

### Checkpoint B: Full Stack Integration

Browser test all 4 entry points. Dev server + manual verification.

### Y8: Final cleanup + verification

- Grep for stale `/api/document-manager/{slug}` refs in FE
- Grep for stale `useDocuments` imports
- Full test suite: `go test ./...` + `npm run build` + `npm run lint`
