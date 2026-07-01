# SPEC — Document Module

## 1. Overview

The document module manages the content entries (documents) within the CMS. Each document belongs to a content type and follows a draft/publish workflow where every entry is stored as two separate records (draft and published). Documents live in per-content-type storage (MongoDB collections or PostgreSQL tables) and support paginated collection lists with field projection. This module owns the Document entity, its repository, the document usecase, all REST and gRPC API contracts for document CRUD, and the draft/publish lifecycle.

---

## 2. File Map

All paths relative to `apps/api/`.

```
internal/domain/entity/document.go                           # Document entity
internal/domain/repository/document_repository.go            # DocumentRepository interface
internal/domain/repository/mock/document_repository.go       # Mock for testing
internal/usecase/document/document_usecase.go                # Document business logic
internal/usecase/document/document_usecase_test.go
internal/delivery/http/handler/document_handler.go           # Document Gin handlers
internal/delivery/http/handler/document_handler_test.go
internal/delivery/grpc/document_service.go                   # gRPC DocumentService
internal/delivery/grpc/document_service_test.go
internal/infrastructure/mongodb/document_repository.go       # MongoDB Document repo
internal/infrastructure/mongodb/document_repository_test.go
internal/infrastructure/mongodb/document_filter_test.go
internal/infrastructure/gormdb/document_repository.go        # GORM Document repo
internal/infrastructure/gormdb/document_repository_test.go
proto/cms/v1/document.proto                                  # gRPC Document proto
proto/cms/v1/document.pb.go                                  # Generated
proto/cms/v1/document_grpc.pb.go                             # Generated
```

---

## 3. Entities

### Document

```go
type Document struct {
    GormID      uint            `gorm:"column:gorm_id;primaryKey;autoIncrement"`
    DocumentID  string          `bson:"documentId"     gorm:"column:document_id;index"        json:"documentId"`
    Version     DocumentVersion `bson:"version"        gorm:"column:version;type:varchar(20)"  json:"version"`
    Fields      map[string]any  `bson:"data"           gorm:"-"                               json:"data"`
    Locale      string          `bson:"locale"         gorm:"column:locale"                    json:"locale"`
    CreatedAt   time.Time       `bson:"createdAt"      gorm:"column:created_at"                json:"createdAt"`
    UpdatedAt   time.Time       `bson:"updatedAt"      gorm:"column:updated_at"                json:"updatedAt"`
    PublishedAt *time.Time      `bson:"publishedAt,omitempty"  gorm:"column:published_at"      json:"publishedAt,omitempty"`
    CreatedBy   string          `bson:"createdBy"      gorm:"column:created_by"                json:"createdBy"`
    UpdatedBy   string          `bson:"updatedBy"      gorm:"column:updated_by"                json:"updatedBy"`
    PublishedBy string          `bson:"publishedBy,omitempty"  gorm:"column:published_by"      json:"publishedBy,omitempty"`
    Slug        string          `bson:"-"              gorm:"column:slug;index"                json:"-"`
}

type DocumentVersion string
const (
    VersionDraft     DocumentVersion = "draft"
    VersionPublished DocumentVersion = "published"
)
```

**Entity changes (v1.8):**
- Removed `ContentTypeID` field (content type is implicit from the per-content-type table/collection)
- Renamed `Data` → `Fields` (tagged `gorm:"-"` — GORM uses per-field columns, not a JSON blob)
- `PublishedAt` changed to `*time.Time` (nullable pointer)
- Added `GormID` (auto-increment uint) for display ordering in list queries
- `document_id` standardized to UUID v4

**Database IDs:** Document entities use `documentId` (UUID v4) as the primary domain identifier. Higher layers (usecase, handler, frontend) only work with `documentId` and content-type `slug` — never with MongoDB `_id` or generic `id`. ContentType entities retain their MongoDB `_id` as `ID`.

---

## 4. Domain Rules

### Draft & Publish Workflow

- Every content entry is stored as **two separate records**: a `draft` and a `published` record (`version: "draft" | "published"`).
- Each content type's documents live in their own standalone storage: MongoDB collection (`documents_<slug>`) or PostgreSQL table (`documents_<slug_underscored>`), created during sync.
- `draft` record: holds latest edits, `createdAt`/`updatedAt`/`createdBy`/`updatedBy`, **no** `publishedAt`/`publishedBy`.
- `published` record: only exists after first publish; carries `publishedAt`/`publishedBy`.
- Every record carries `locale` (defaults to `"en"`).
- **Documents are only created on explicit save.** No document exists until the user saves.
- **Save**: upserts `draft` record's `data`, `updatedAt`, `updatedBy`. Never touches `published`.
- **Publish**: copies `draft.data` → `published` record, sets `publishedAt = now()`, `publishedBy = user`.
- Entry `status` is **computed, never stored**: `draft` (no published), `modified` (`draft.updatedAt > published.updatedAt`), `published` (timestamps match).
- Public read API resolves `published` record only. If no `published` exists → 404.
- Admin edit screens read `draft` record + computed `status`.

---

## 5. Repository Interfaces

### DocumentRepository

```go
type DocumentRepository interface {
    FindDraftByDocumentID(ctx context.Context, contentTypeSlug, documentID, locale string) (*entity.Document, error)
    FindPublishedByDocumentID(ctx context.Context, contentTypeSlug, documentID, locale string) (*entity.Document, error)
    UpsertDraft(ctx context.Context, contentTypeSlug string, doc *entity.Document) error
    UpsertPublished(ctx context.Context, contentTypeSlug string, doc *entity.Document) error
    DeleteByDocumentID(ctx context.Context, contentTypeSlug, documentID string) error
    DeletePublished(ctx context.Context, contentTypeSlug, documentID, locale string) error
    DeleteAllByContentType(ctx context.Context, contentTypeSlug string) error
    FindDraftsByContentTypePaginated(ctx context.Context, contentTypeSlug string, start, size int, locale string) ([]*entity.Document, int64, error)
    FindPublishedByDocumentIDs(ctx context.Context, contentTypeSlug string, documentIDs []string, locale string) ([]*entity.Document, error)
    EnsureCollection(ctx context.Context, contentTypeSlug string, fields []entity.FieldDefinition) error
    DropCollection(ctx context.Context, contentTypeSlug string) error
}
```

**MongoDB**: Routes by collection name (`documents_<slug>`). Components remain nested in BSON `data`.
**GORM**: Routes by dynamic table name (`documents_<slug_underscored>`), matching MongoDB's per-content-type pattern. `EnsureCollection` creates the table + unique index; `DropCollection` drops the table. Component fields stored in separate `components_<slug_underscored>_<component_name_underscored>` tables (see `specs/component.md`).

---

## 6. Use Cases

### Document UseCase (`usecase/document/`)

| Method | Signature | Description |
|---|---|---|
| `Save` | `(ctx, slug, doc, locale, userID) → (*Document, err)` | Upsert draft record |
| `Publish` | `(ctx, slug, documentID, locale, userID) → err` | Copy draft → published |
| `Unpublish` | `(ctx, slug, documentID, locale) → err` | Delete published record |
| `GetForEdit` | `(ctx, slug, documentID, locale) → (*Document, status, err)` | Get draft + computed status |
| `GetPublic` | `(ctx, slug, documentID, locale) → (*Document, err)` | Get published only |
| `Delete` | `(ctx, slug, documentID) → err` | Delete draft + published + cascade media |
| `GetSingleType` | `(ctx, slug, locale) → (*Document, status, err)` | Get single-type draft |
| `SaveSingleType` | `(ctx, slug, data, locale, userID) → (*Document, err)` | Create-or-update single-type |
| `PublishSingleType` | `(ctx, slug, locale, userID) → err` | Publish single-type |
| `UnpublishSingleType` | `(ctx, slug, locale) → err` | Unpublish single-type |
| `GetAllPaginated` | `(ctx, slug, start, size, locale) → (docs, statuses, total, err)` | Paginated collection list |
| `Duplicate` | `(ctx, slug, sourceDocumentID, locale, userID) → (*Document, err)` | Copy a document into a new draft (fresh `documentId`, media refs shared) |
| `BulkCreateAndPublish` | `(ctx, slug, itemsData []map[string]any, locale, userID) → ([]*Document, err)` | Create + publish up to 100 collection-type documents in one call; sequential Save→Publish per item, rollback via `Delete` on the first failure (all-or-nothing) |

---

## 7. API Contracts

### REST — Single-Type Document Routes

| Method | Route | Permission | Response |
|---|---|---|---|
| `GET` | `/api/document-manager/single-type/:slug` | `content:read` | `Document` or `404` |
| `PUT` | `/api/document-manager/single-type/:slug` | `content:update` | `Document` |
| `POST` | `/api/document-manager/single-type/:slug/publish` | `content:publish` | `{"status":"published"}` |
| `POST` | `/api/document-manager/single-type/:slug/unpublish` | `content:unpublish` | `{"status":"draft"}` |

Query param: `?locale=` (defaults to first supported locale).

**GET behavior:** 404 when no document exists (FE interprets as "show empty form").
**PUT behavior:** Creates on first save (auto-generates `documentId`), updates on subsequent saves.

### REST — Collection-Type Document Routes

| Method | Route | Permission | Response |
|---|---|---|---|
| `GET` | `/api/document-manager/collection-type/:slug` | `content:read` | `PaginatedList` |
| `GET` | `/api/document-manager/collection-type/:slug/:documentId` | `content:read` | `Document` |
| `POST` | `/api/document-manager/collection-type/:slug` | `content:create` | `Document` (201) |
| `POST` | `/api/document-manager/collection-type/:slug/bulk` | `content:create` **and** `content:publish` | `{"items":[Document, ...]}` (201) |
| `PUT` | `/api/document-manager/collection-type/:slug/:documentId` | `content:update` | `Document` |
| `DELETE` | `/api/document-manager/collection-type/:slug/:documentId` | `content:delete` | `204` |
| `POST` | `/api/document-manager/collection-type/:slug/:documentId/publish` | `content:publish` | `{"status":"published"}` |
| `POST` | `/api/document-manager/collection-type/:slug/:documentId/unpublish` | `content:unpublish` | `{"status":"draft"}` |
| `POST` | `/api/document-manager/collection-type/:slug/:documentId/duplicate` | `content:create` | `Document` (201) |

### Bulk Create + Publish (Collection-Type Only)

`POST /api/document-manager/collection-type/:slug/bulk?locale=` — creates and immediately publishes multiple documents in one request. Added to support bulk content import (e.g. seeding vocabulary entries) without N create + N publish round trips.

**Request:**
```json
{ "items": [ { "data": { "...": "same field shape as single-item create" } }, ... ] }
```
- `items`: required, 1–100 entries. Each item's `data` follows the same contract as the single-item `documentRequest` — it is **not** optional; an item with no `data` key is *not* rejected by extra validation today (see boundary note below) and will create a document with empty fields if sent that way.
- One `?locale=` for the entire request — no per-item locale.

**Response (all succeeded):**
```json
{ "items": [ { "data": { "documentId": "...", "...": "..." }, "status": "published" }, ... ] }
```

**Semantics — all-or-nothing via rollback, not a DB transaction:** items are processed sequentially through the existing `Save` → `Publish` methods, unchanged. If any item fails, every document already committed earlier in the batch is deleted (via the existing `Delete` usecase method, which removes draft + published + components for every locale) before the error is returned. This is a compensating rollback rather than an atomic multi-document transaction — MongoDB standalone deployments don't support that, and this codebase has no transaction abstraction for either DB backend (Mongo or GORM). Requires both `content:create` and `content:publish` permissions (enforced via two chained `GinRequirePermission` middleware calls on the route).

**Pagination parameters:**

| Param | Default | Max | Description |
|-------|---------|-----|-------------|
| `start` | `0` | — | Offset |
| `size` | `20` | `100` | Items per page |
| `locale` | first supported | — | Filter by locale |
| `orderBy` | `gorm_id` | — | Field to sort by |
| `sortDir` | `desc` | — | Sort direction (`asc` or `desc`) |

**Paginated list response:**
```json
{
  "items": [
    {
      "documentId": "...",
      "data": { /* only listFields */ },
      "status": "draft",
      "locale": "en",
      "createdAt": "...",
      "updatedAt": "...",
      "updatedByName": "John Doe"
    }
  ],
  "total": 42,
  "start": 0,
  "size": 20
}
```

`updatedByName` is resolved server-side from User entity's `DisplayName` field.

`items[].data` contains **only** the fields specified in `listFields`. Full data available via single-document GET.

**REST document response shape (v1.8):**
```json
{
  "data": {
    "documentId": "...",
    "status": "draft",
    "locale": "en",
    "createdAt": "...",
    "updatedAt": "...",
    "fieldName1": "value1",
    "fieldName2": "value2"
  }
}
```

System fields and content fields are merged flat inside `data`. Fields `contentTypeId` and `status` are excluded from public API responses.

### REST — Public Read Route

| Method | Route | Auth | Response |
|---|---|---|---|
| `GET` | `/api/public/document-manager/:slug/:documentId` | None | Published `Document` or `404` |

### REST — Locales

| Method | Route | Auth | Response |
|---|---|---|---|
| `GET` | `/api/locales` | None | `string[]` |

### Input Validation (Handler-Level)

**Slug validation:** `^[a-z0-9]+(?:-[a-z0-9]+)*$` — applied on every request that reads `slug` from URL. 400 on invalid.

**DocumentID validation:** UUID v4 format — applied on every request that reads `documentId`. 400 on invalid.

### gRPC — DocumentService

```protobuf
service DocumentService {
    rpc GetDocument(GetDocumentRequest) returns (Document);
    rpc ListDocuments(ListDocumentsRequest) returns (ListDocumentsResponse);
    rpc SaveDocument(SaveDocumentRequest) returns (Document);
    rpc PublishDocument(PublishDocumentRequest) returns (Document);
    rpc UnpublishDocument(PublishDocumentRequest) returns (Document);
    rpc DeleteDocument(DeleteDocumentRequest) returns (DeleteDocumentResponse);
    rpc GetSingleType(GetSingleTypeRequest) returns (Document);
    rpc SaveSingleType(SaveSingleTypeRequest) returns (Document);
    rpc PublishSingleType(GetSingleTypeRequest) returns (Document);
    rpc UnpublishSingleType(GetSingleTypeRequest) returns (Document);
}
```

---

## 8. Testing

**Document usecase (`document_usecase_test.go`):**
- Save: upserts draft, never touches published
- Publish: copies draft → published, sets timestamps
- Unpublish: deletes published record
- Status computation: draft / modified / published
- Single-type: GetSingleType returns 404 when empty, SaveSingleType creates on first save
- GetAllPaginated: correct pagination, batch status computation
- BulkCreateAndPublish: all-valid batch creates+publishes in order; a mid-batch `Save` failure rolls back all prior items and reports the failing index; a `Publish` failure rolls back the current item too (not just prior ones); rejects unsupported locale before any repo call

**Document handler (`document_handler_test.go`):**
- Single-type: GET 200/404, PUT create/update, Publish, Unpublish
- Collection: List paginated with projected fields, CRUD, Publish, Unpublish
- Slug/documentID validation → 400
- BulkCreateCollection: 201 on valid batch; 400 on empty/over-100-item batch or malformed body (checked before the usecase is called); usecase errors map through the existing error mapping (e.g. `ErrValidation` → 422)

---

## 9. Boundaries

| Rule | Detail |
|---|---|
| **Always** | Validate slug format at both usecase (creation) and handler (every request) levels |
| **Always** | Validate documentID is UUID v4 format before passing to usecase |
| **Always** | Return 404 (not empty object) for single-type GET when no document exists |
| **Always** | Include computed `status` in every document response |
| **Always** | Project `data` in collection list responses — never return full data in paginated lists |
| **Always** | Batch-fetch published records for status computation (no N+1) |
| **Always** | Return 400 (not 500) for invalid slug or documentID format |
| **Never** | Expose DELETE for single-type documents |
| **Never** | Allow `size` above 100 on collection list |
| **Never** | Return draft data through public read API |
| **Never** | Include `documentId` in single-type URLs |
| **Never** | Restrict create/update/delete/save/publish on content data (documents) |
| **Never** | Allow more than 100 items in a single bulk create+publish request |
| **Never** | Accept per-item locale on bulk create+publish — one `?locale=` for the whole request |
| **Ask first** | Changing default `size` from 20 or max from 100 |
| **Known gap** | Bulk create+publish does not reject items with a missing/empty `data` field — it will silently create a document with empty fields rather than erroring (confirmed and intentionally left as-is on 2026-07-01; revisit if it causes repeated confusion) |

---

## 10. Changelog

| Date | Change | Source |
|------|--------|--------|
| v1.0 | Content types, documents, draft/publish workflow | §1, §4 |
| v1.3 | API restructure: single-type/collection-type routes, pagination, field projection | §10 |
| v1.8 | Document entity: removed `ContentTypeID`, renamed `Data` → `Fields` (gorm:"-"), `PublishedAt` → `*time.Time`, added `GormID` for ordering | sync-table-fields |
| v1.11 | Document list supports `orderBy`/`sortDir` query params; responses include `updatedByName` | collection-list-enhancements |
| v1.12 | REST responses wrap content in `{ data: { ...systemFields, ...contentFields } }`; `contentTypeId`/`status` removed from public responses | bugfix-v1.8 |
| v1.13 | Bulk create+publish for collection-type documents: `POST /:slug/bulk`, up to 100 items, all-or-nothing via rollback (not a DB transaction) | bulk-document-create-publish |
