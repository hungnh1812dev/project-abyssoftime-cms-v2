# SPEC — content Module

## 1. Overview

The content module manages content types (schema-as-code), documents (draft/publish workflow), paginated collection lists, field projection, and dynamic GraphQL schema generation. It is the largest domain module, covering the core CMS data model: content-type definition files synced on startup, per-content-type document collections, and the complete single-type / collection-type API surface.

---

## 2. File Map

All paths relative to `apps/api/`.

```
content-types/                                               # JSON schema-as-code definition files
internal/domain/entity/content_type.go                       # ContentType entity
internal/domain/entity/document.go                           # Document entity
internal/domain/repository/content_type_repository.go        # ContentTypeRepository interface
internal/domain/repository/document_repository.go            # DocumentRepository interface
internal/domain/repository/mock/content_type_repository.go   # Mock for testing
internal/domain/repository/mock/document_repository.go       # Mock for testing
internal/usecase/content_type/content_type_usecase.go        # ContentType CRUD
internal/usecase/content_type/content_type_usecase_test.go
internal/usecase/content_type/schema_loader.go               # JSON definition file parser
internal/usecase/content_type/schema_loader_test.go
internal/usecase/content_type/sync.go                        # Schema-as-code sync engine
internal/usecase/content_type/sync_test.go
internal/usecase/content_type/testdata/                      # Test fixture JSON files
internal/usecase/document/document_usecase.go                # Document business logic
internal/usecase/document/document_usecase_test.go
internal/delivery/http/handler/content_type_handler.go       # ContentType Gin handlers
internal/delivery/http/handler/content_type_handler_test.go
internal/delivery/http/handler/document_handler.go           # Document Gin handlers
internal/delivery/http/handler/document_handler_test.go
internal/delivery/http/handler/locale_handler.go             # Locale list handler
internal/delivery/http/handler/locale_handler_test.go
internal/delivery/grpc/content_type_service.go               # gRPC ContentTypeService
internal/delivery/grpc/document_service.go                   # gRPC DocumentService
internal/delivery/grpc/document_service_test.go
internal/infrastructure/mongodb/content_type_repository.go   # MongoDB ContentType repo
internal/infrastructure/mongodb/content_type_repository_test.go
internal/infrastructure/mongodb/document_repository.go       # MongoDB Document repo
internal/infrastructure/mongodb/document_repository_test.go
internal/infrastructure/mongodb/document_filter_test.go
internal/infrastructure/gormdb/content_type_repository.go    # GORM ContentType repo
internal/infrastructure/gormdb/content_type_repository_test.go
internal/infrastructure/gormdb/document_repository.go        # GORM Document repo
internal/infrastructure/gormdb/document_repository_test.go
graphql/dynamic/schema_builder.go                            # Dynamic GraphQL SDL generator
graphql/dynamic/schema_builder_test.go
graphql/dynamic/resolver_factory.go                          # Per-content-type resolver factory
graphql/dynamic/resolver_factory_test.go
proto/cms/v1/document.proto                                  # gRPC Document proto
proto/cms/v1/content_type.proto                              # gRPC ContentType proto
proto/cms/v1/document.pb.go                                  # Generated
proto/cms/v1/document_grpc.pb.go                             # Generated
proto/cms/v1/content_type.pb.go                              # Generated
proto/cms/v1/content_type_grpc.pb.go                         # Generated
```

---

## 3. Entities

### ContentType

```go
type ContentType struct {
    ID         string            `bson:"_id,omitempty" gorm:"column:id;primaryKey"   json:"-"`
    Name       string            `bson:"name"          gorm:"column:name"             json:"name"`
    Slug       string            `bson:"slug"          gorm:"column:slug;uniqueIndex" json:"slug"`
    Kind       ContentKind       `bson:"kind"          gorm:"column:kind;type:varchar(20)" json:"kind"`
    Fields     []FieldDefinition `bson:"fields,omitempty"     gorm:"column:fields;serializer:json"      json:"Fields,omitempty"`
    ListFields []string          `bson:"listFields,omitempty" gorm:"column:list_fields;serializer:json"  json:"listFields,omitempty"`
    CreatedAt  time.Time         `bson:"createdAt"     gorm:"column:created_at"       json:"createdAt"`
    UpdatedAt  time.Time         `bson:"updatedAt"     gorm:"column:updated_at"       json:"updatedAt"`
}

type ContentKind string
const (
    KindSingle     ContentKind = "single"
    KindCollection ContentKind = "collection"
)
```

- `Kind`: `"single"` (at most one entry) or `"collection"` (many entries)
- `ListFields`: which fields appear in paginated collection list responses (optional; defaults to first 3 fields)
- `Fields`: array of `FieldDefinition` (name, type, and optional sub-fields for components)

### Document

```go
type Document struct {
    DocumentID    string          `bson:"documentId"     gorm:"column:document_id;index"        json:"documentId"`
    Version       DocumentVersion `bson:"version"        gorm:"column:version;type:varchar(20)"  json:"version"`
    ContentTypeID string          `bson:"contentTypeId"  gorm:"column:content_type_id;index"     json:"contentTypeId"`
    Data          map[string]any  `bson:"data"           gorm:"column:data;serializer:json"      json:"data"`
    Locale        string          `bson:"locale"         gorm:"column:locale"                    json:"locale"`
    CreatedAt     time.Time       `bson:"createdAt"      gorm:"column:created_at"                json:"createdAt"`
    UpdatedAt     time.Time       `bson:"updatedAt"      gorm:"column:updated_at"                json:"updatedAt"`
    PublishedAt   time.Time       `bson:"publishedAt,omitempty"  gorm:"column:published_at"      json:"publishedAt,omitempty"`
    CreatedBy     string          `bson:"createdBy"      gorm:"column:created_by"                json:"createdBy"`
    UpdatedBy     string          `bson:"updatedBy"      gorm:"column:updated_by"                json:"updatedBy"`
    PublishedBy   string          `bson:"publishedBy,omitempty"  gorm:"column:published_by"      json:"publishedBy,omitempty"`
    Slug          string          `bson:"-"              gorm:"column:slug;index"                json:"-"`
}

type DocumentVersion string
const (
    VersionDraft     DocumentVersion = "draft"
    VersionPublished DocumentVersion = "published"
)
```

**Database IDs:** Document entities use `documentId` as the primary domain identifier. Higher layers (usecase, handler, frontend) only work with `documentId` and content-type `slug` — never with MongoDB `_id` or generic `id`. ContentType entities retain their MongoDB `_id` as `ID`.

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

### Content Type Kinds

- **Single-type**: at most one entry per content type. No auto-created singleton — first explicit Save creates it. UI: edit + Save + Publish only. No create/delete.
- **Collection-type**: zero or more entries, each with own `documentId`. List + create/edit/delete, each with independent draft/published pair.

### Content-Type Schema as Code

- Structure defined in JSON files under `content-types/*.json` — never via API or UI.
- On every API startup, `usecase/content_type.Sync` reconciles definitions against DB:
  - **New file** → create ContentType + per-content-type document collection with indexes
  - **Changed file** → update ContentType schema in place
  - **Field removed** → drop from schema, leave stored data untouched
  - **File deleted** → delete ContentType, cascade-delete all entries, drop collection
- Sync is one-directional: JSON definitions are source of truth.
- JSON schemas declare only content fields. System fields (`createdAt`, `updatedAt`, `publishedAt`, `createdBy`, `updatedBy`, `publishedBy`, `locale`) are injected automatically.

### `listFields` (Field Projection)

- Optional array in JSON definitions specifying which fields appear in paginated collection list responses.
- If omitted, defaults to first 3 field names from `fields`.
- Only meaningful for `kind: "collection"`.
- Each entry must reference a valid field name. Schema loader validates on startup.

---

## 5. Repository Interfaces

### ContentTypeRepository

```go
type ContentTypeRepository interface {
    Create(ctx context.Context, ct *entity.ContentType) error
    FindBySlug(ctx context.Context, slug string) (*entity.ContentType, error)
    FindByID(ctx context.Context, id string) (*entity.ContentType, error)
    FindAll(ctx context.Context) ([]*entity.ContentType, error)
    Update(ctx context.Context, ct *entity.ContentType) error
    Delete(ctx context.Context, id string) error
}
```

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
    EnsureCollection(ctx context.Context, contentTypeSlug string) error
    DropCollection(ctx context.Context, contentTypeSlug string) error
}
```

**MongoDB**: Routes by collection name (`documents_<slug>`). Components remain nested in BSON `data`.
**GORM**: Routes by dynamic table name (`documents_<slug_underscored>`), matching MongoDB's per-content-type pattern. `EnsureCollection` creates the table + unique index; `DropCollection` drops the table. Component fields stored in separate `components_<slug_underscored>_<component_name_underscored>` tables (see §9).

---

## 6. Use Cases

### ContentType UseCase (`usecase/content_type/`)

| Method | Description |
|---|---|
| `Create(ctx, ct)` | Create content type; validates slug format |
| `FindBySlug(ctx, slug)` | Get by slug |
| `FindByID(ctx, id)` | Get by ID |
| `FindAll(ctx)` | List all |
| `ListSummary(ctx)` | List with minimal fields (Name, Slug, Kind) |
| `Update(ctx, ct)` | Update content type |
| `Delete(ctx, id)` | Delete content type |

**Slug validation:** `^[a-z0-9]+(?:-[a-z0-9]+)*$`, 1–63 chars. Called in `Create` before any DB operation.

### Schema Loader (`usecase/content_type/schema_loader.go`)

```go
func LoadDefinitions(dir string) ([]ContentTypeDefinition, error)
```

Reads all `*.json` files from the directory. Validates `listFields` entries reference valid field names.

### Syncer (`usecase/content_type/sync.go`)

```go
type Syncer struct { ctUC, docUC, docRepo }
func (s *Syncer) Sync(ctx, defs []ContentTypeDefinition) error
```

Runs on every startup. Reconciles JSON definitions against DB records.

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

---

## 7. API Contracts

### REST — Content-Type Routes

| Method | Route | Permission | Response |
|---|---|---|---|
| `GET` | `/api/content-types` | `content_types:read` | `ContentTypeSummary[]` |
| `GET` | `/api/content-types/:identifier` | `content_types:read` | `ContentType` (full) |

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
| `PUT` | `/api/document-manager/collection-type/:slug/:documentId` | `content:update` | `Document` |
| `DELETE` | `/api/document-manager/collection-type/:slug/:documentId` | `content:delete` | `204` |
| `POST` | `/api/document-manager/collection-type/:slug/:documentId/publish` | `content:publish` | `{"status":"published"}` |
| `POST` | `/api/document-manager/collection-type/:slug/:documentId/unpublish` | `content:unpublish` | `{"status":"draft"}` |

**Pagination parameters:**

| Param | Default | Max | Description |
|-------|---------|-----|-------------|
| `start` | `0` | — | Offset |
| `size` | `20` | `100` | Items per page |
| `locale` | first supported | — | Filter by locale |

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
      "updatedAt": "..."
    }
  ],
  "total": 42,
  "start": 0,
  "size": 20
}
```

`items[].data` contains **only** the fields specified in `listFields`. Full data available via single-document GET.

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

**DocumentID validation:** `^[a-f0-9]{24}$` (MongoDB ObjectID format) — applied on every request that reads `documentId`. 400 on invalid.

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

### gRPC — ContentTypeService

```protobuf
service ContentTypeService {
    rpc GetContentType(GetContentTypeRequest) returns (ContentType);
    rpc ListContentTypes(ListContentTypesRequest) returns (ListContentTypesResponse);
}
```

---

## 8. Dynamic GraphQL Schema Generation

On startup, after content-type sync:
1. Schema builder reads all `ContentTypeDefinition` structs
2. For each content-type, generates GraphQL types, queries, and mutations
3. Resolver factory creates resolvers that delegate to document usecase

### Field Type Mapping

| Content-Type `type` | GraphQL Type |
|---|---|
| `text` | `String` |
| `richtext` | `String` |
| `number` | `Float` |
| `boolean` | `Boolean` |
| `media` | `String` (URL) |
| `json` | `JSON` (scalar) |
| `component` | Nested object type |

### Generated Schema Per Content-Type

**Collection-type** generates:
- `Query.<slug>(Id: ID!, locale: String): <Type>Response` — fetch one document (with component fields resolved as nested objects)
- `Query.<slugList>(where: <Type>Filter, orderBy: <Type>OrderBy, start: Int, size: Int, locale: String): <Type>ListResponse` — paginated list with filtering, sorting, and component data
- `Mutation.create<Type>(data: <Type>Input!): <Type>! @auth`
- `Mutation.update<Type>(Id: ID!, data: <Type>Input!): <Type>! @auth`
- `Mutation.delete<Type>(Id: ID!): Boolean! @auth`
- `Mutation.publish<Type>(Id: ID!, locale: String): <Type>! @auth`
- `Mutation.unpublish<Type>(Id: ID!, locale: String): <Type>! @auth`

**Single-type** generates:
- `Query.<slug>(locale: String): <Type>Response` — fetch singleton (with component fields resolved)
- `Mutation.save<Type>(data: <Type>Input!, locale: String): <Type>! @auth`
- `Mutation.publish<Type>(locale: String): <Type>! @auth`
- `Mutation.unpublish<Type>(locale: String): <Type>! @auth`

### Naming Conventions
- Type: PascalCase of slug (`blog-posts` → `BlogPost`)
- Input: `<Type>Input`
- Response (single): `<Type>Response` — wraps `data: <Type>`
- Response (list): `<Type>ListResponse` — wraps `data: [<Type>]`
- Filter: `<Type>Filter` — generated from document + component fields
- OrderBy: `<Type>OrderBy` — sort by document + component fields
- Query single: camelCase of slug (`blog-posts` → `blogPost`)
- Query list: camelCase of slug + `List` (`blog-posts` → `blogPostList`)
- Component types: PascalCase of `<ContentType><ComponentName>` (e.g., `BlogPostBanner`)

### Query Examples

**Single document (single-type or collection-type by ID):**

```graphql
query GetBlogPost($locale: String!) {
  blogPost(Id: "abc123", locale: $locale) {
    title
    coverImage { url }
    banner {
      background { url }
    }
  }
}
```

```json
{
  "blogPost": {
    "data": {
      "title": "...",
      "coverImage": { "url": "..." },
      "banner": {
        "background": { "url": "..." }
      }
    }
  }
}
```

**List query with filter, sort, and pagination:**

```graphql
query GetBlogPostList($locale: String!) {
  blogPostList(
    where: { featured: { eq: true } }
    orderBy: { createdAt: DESC }
    start: 0
    size: 10
    locale: $locale
  ) {
    title
    coverImage { url }
    banner {
      background { url }
    }
  }
}
```

```json
{
  "blogPostList": {
    "data": [
      {
        "title": "...",
        "coverImage": { "url": "..." },
        "banner": {
          "background": { "url": "..." }
        }
      }
    ]
  }
}
```

### Filtering & Sorting

Filters are generated per content-type based on field types:

| Field Type | Supported Operators |
|---|---|
| `text` | `eq`, `ne`, `contains`, `startsWith`, `endsWith`, `in` |
| `number` | `eq`, `ne`, `gt`, `gte`, `lt`, `lte`, `in` |
| `boolean` | `eq` |
| `component` | Nested filter on component sub-fields |

Top-level logical operators: `AND`, `OR`, `NOT`.

OrderBy supports `ASC` / `DESC` on scalar fields (text, number, boolean) and system fields (`createdAt`, `updatedAt`, `publishedAt`).

---

## 9. Infrastructure — GORM Specifics

### Per-Content-Type Document Tables

GORM creates one table per content type, matching MongoDB's per-collection pattern. Tables are created dynamically during content-type sync, not via `AutoMigrate`.

**Table naming:** `documents_<slug_underscored>` (hyphens replaced with underscores).
Example: content type `blog-posts` → table `documents_blog_posts`.

**Slug sanitization:** Only `[a-z0-9-]` allowed in slugs (validated at content-type creation). Hyphens converted to underscores for PostgreSQL table names.

```sql
CREATE TABLE IF NOT EXISTS documents_<slug_underscored> (
    gorm_id         BIGSERIAL PRIMARY KEY,
    document_id     VARCHAR(255) NOT NULL,
    version         VARCHAR(20)  NOT NULL,
    content_type_id VARCHAR(255),
    data            JSONB,
    locale          VARCHAR(10)  NOT NULL,
    created_at      TIMESTAMPTZ,
    updated_at      TIMESTAMPTZ,
    published_at    TIMESTAMPTZ,
    created_by      VARCHAR(255),
    updated_by      VARCHAR(255),
    published_by    VARCHAR(255),
    UNIQUE(document_id, version, locale)
);
```

**Implementation changes:**
- `EnsureCollection(slug)`: Executes `CREATE TABLE IF NOT EXISTS documents_<slug_underscored>` with the schema above + unique index
- `DropCollection(slug)`: Executes `DROP TABLE IF EXISTS documents_<slug_underscored>`
- All repository queries use `r.db.Table("documents_" + sanitize(slug))` instead of the default model table
- `Slug` field removed from the `Document` entity (or marked `gorm:"-"`) — table name replaces the discriminator
- `&entity.Document{}` removed from `AutoMigrate()` — tables are created dynamically by `EnsureCollection`

**Migration from single table:**
1. For each distinct `slug` in the existing `documents` table, create `documents_<slug_underscored>` and copy rows
2. Drop the old `documents` table after verification

### Component Tables (PostgreSQL Only)

When a content-type field has `type: "component"`, GORM creates a dedicated component table. MongoDB does **not** use component collections — components remain nested in BSON `data`.

**Table naming:** `components_<slug_underscored>_<component_name_snake_case>`
Example: content type `blog-posts` with component field `banner` → table `components_blog_posts_banner`.

```sql
CREATE TABLE IF NOT EXISTS components_<slug_underscored>_<component_name_underscored> (
    gorm_id       BIGSERIAL PRIMARY KEY,
    component_id  VARCHAR(255) NOT NULL,
    document_id   VARCHAR(255) NOT NULL,
    version       VARCHAR(20)  NOT NULL,
    locale        VARCHAR(10)  NOT NULL,
    "order"       INT          NOT NULL DEFAULT 0,
    data          JSONB,
    created_at    TIMESTAMPTZ,
    updated_at    TIMESTAMPTZ,
    UNIQUE(document_id, version, locale, "order")
);
```

**New entity — `Component`:**

```go
type Component struct {
    GormID      uint            `gorm:"column:gorm_id;primaryKey;autoIncrement"`
    ComponentID string          `gorm:"column:component_id"`
    DocumentID  string          `gorm:"column:document_id"`
    Version     DocumentVersion `gorm:"column:version;type:varchar(20)"`
    Locale      string          `gorm:"column:locale"`
    Order       int             `gorm:"column:order"`
    Data        map[string]any  `gorm:"column:data;serializer:json"`
    CreatedAt   time.Time       `gorm:"column:created_at"`
    UpdatedAt   time.Time       `gorm:"column:updated_at"`
}
```

**New repository — `ComponentRepository`:**

```go
type ComponentRepository interface {
    FindByDocumentID(ctx context.Context, contentTypeSlug, componentName, documentID, locale string, version DocumentVersion) ([]*Component, error)
    UpsertAll(ctx context.Context, contentTypeSlug, componentName, documentID, locale string, version DocumentVersion, components []*Component) error
    DeleteByDocumentID(ctx context.Context, contentTypeSlug, componentName, documentID, locale string) error
    DeleteAllByContentType(ctx context.Context, contentTypeSlug, componentName string) error
    EnsureCollection(ctx context.Context, contentTypeSlug, componentName string) error
    DropCollection(ctx context.Context, contentTypeSlug, componentName string) error
}
```

**Sync engine changes:**
- `syncOne()`: After ensuring document table, iterate content-type fields — for each `"type": "component"` field, call `componentRepo.EnsureCollection(slug, field.Name)`
- `removeContentType()`: Drop all component tables before dropping the document table
- `Syncer` receives `ComponentRepository` in constructor

**Document save/read flow (PostgreSQL):**
- **Save:** Extract component fields from `doc.Data` based on content-type field definitions → save scalar data to `documents_<slug_underscored>` → upsert component records to `components_<slug_underscored>_<component_name_underscored>`
- **Read:** Load document from `documents_<slug_underscored>` → for each component field, load from `components_<slug_underscored>_<component_name_underscored>` → merge into `doc.Data` before returning
- **Delete:** Delete component records from all component tables → delete document record
- **Publish:** Copy draft component records to published version → copy draft document to published

**MongoDB (no change):** Components remain nested in the document's BSON `data` field. No component collections are created.

---

## 10. Testing

**Schema sync (`sync_test.go`):**
- New file → creates ContentType + collection
- Changed file → updates schema
- Removed field → drops from schema, data untouched
- Deleted file → cascade-deletes type + entries + collection

**Schema loader (`schema_loader_test.go`):**
- Valid JSON → parses correctly
- `listFields` referencing invalid field → error
- Malformed JSON → error

**Document usecase (`document_usecase_test.go`):**
- Save: upserts draft, never touches published
- Publish: copies draft → published, sets timestamps
- Unpublish: deletes published record
- Status computation: draft / modified / published
- Single-type: GetSingleType returns 404 when empty, SaveSingleType creates on first save
- GetAllPaginated: correct pagination, batch status computation

**Content-type usecase (`content_type_usecase_test.go`):**
- Valid slugs accepted; invalid rejected
- CRUD operations

**Document handler (`document_handler_test.go`):**
- Single-type: GET 200/404, PUT create/update, Publish, Unpublish
- Collection: List paginated with projected fields, CRUD, Publish, Unpublish
- Slug/documentID validation → 400

**GraphQL (`schema_builder_test.go`, `resolver_factory_test.go`):**
- SDL generation for collection and single types
- Field type mapping
- Resolver delegation to usecase methods

---

## 11. Boundaries

| Rule | Detail |
|---|---|
| **Always** | Validate slug format at both usecase (creation) and handler (every request) levels |
| **Always** | Validate documentID is 24-char hex before passing to usecase |
| **Always** | Return 404 (not empty object) for single-type GET when no document exists |
| **Always** | Include computed `status` in every document response |
| **Always** | Project `data` in collection list responses — never return full data in paginated lists |
| **Always** | Validate `listFields` against `fields` during schema sync startup |
| **Always** | Batch-fetch published records for status computation (no N+1) |
| **Always** | Route dynamic GraphQL through the same usecase — no logic in resolvers |
| **Always** | Return 400 (not 500) for invalid slug or documentID format |
| **Never** | Allow slug characters outside `[a-z0-9-]` |
| **Never** | Expose DELETE for single-type documents |
| **Never** | Allow `size` above 100 on collection list |
| **Never** | Return draft data through public read API |
| **Never** | Include `documentId` in single-type URLs |
| **Never** | Let content-type sync write back to JSON definition files |
| **Never** | Add API/UI path to create/edit/delete ContentType structure |
| **Never** | Restrict create/update/delete/save/publish on content data (documents) |
| **Ask first** | Changing default `size` from 20 or max from 100 |
| **Always** | GraphQL list queries support `where`, `orderBy`, `start`, `size` parameters |
| **Always** | GraphQL responses wrap document data in a `data` field |
| **Always** | PostgreSQL component tables created/dropped during content-type sync |
| **Never** | Create component collections in MongoDB — components stay nested in BSON `data` |

---

## 12. Changelog

| Date | Change | Source |
|------|--------|--------|
| v1.0 | Content types, documents, draft/publish workflow | §1, §4 |
| v1.1 | Schema-as-code (JSON definitions, startup sync) | §4 |
| v1.2 | Slug + documentID input validation | §9.1–§9.3 |
| v1.3 | API restructure: single-type/collection-type routes, pagination, field projection | §10 |
| v1.4 | Dynamic GraphQL schema generation | §11.5 |
| v1.5 | PostgreSQL per-content-type document tables (`documents_<slug_underscored>`) replacing single `documents` table | §9 |
| v1.6 | PostgreSQL component tables (`components_<slug_underscored>_<component_name_underscored>`) — MongoDB keeps nested BSON | §9 |
| v1.7 | GraphQL: list query renamed to `<slug>List`, response wrapped in `data`, filter/orderBy/pagination support | §8 |
