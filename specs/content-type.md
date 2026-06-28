# SPEC — Content Type Module

## 1. Overview

The content-type module manages the schema-as-code content type definitions that form the structural backbone of the CMS. Content types are defined as JSON files under `content-types/*.json` and synced to the database on every API startup. Each content type has a slug, a kind (single or collection), a list of field definitions, and optional `listFields` for controlling paginated list projections. This module owns the ContentType entity, its repository, the schema loader that parses JSON definitions, and the sync engine that reconciles definitions against the database.

---

## 2. File Map

All paths relative to `apps/api/`.

```
content-types/                                               # JSON schema-as-code definition files
internal/domain/entity/content_type.go                       # ContentType entity
internal/domain/repository/content_type_repository.go        # ContentTypeRepository interface
internal/domain/repository/mock/content_type_repository.go   # Mock for testing
internal/usecase/content_type/content_type_usecase.go        # ContentType CRUD
internal/usecase/content_type/content_type_usecase_test.go
internal/usecase/content_type/schema_loader.go               # JSON definition file parser
internal/usecase/content_type/schema_loader_test.go
internal/usecase/content_type/sync.go                        # Schema-as-code sync engine
internal/usecase/content_type/sync_test.go
internal/usecase/content_type/testdata/                      # Test fixture JSON files
internal/delivery/http/handler/content_type_handler.go       # ContentType Gin handlers
internal/delivery/http/handler/content_type_handler_test.go
internal/delivery/grpc/content_type_service.go               # gRPC ContentTypeService
internal/infrastructure/mongodb/content_type_repository.go   # MongoDB ContentType repo
internal/infrastructure/mongodb/content_type_repository_test.go
internal/infrastructure/gormdb/content_type_repository.go    # GORM ContentType repo
internal/infrastructure/gormdb/content_type_repository_test.go
proto/cms/v1/content_type.proto                              # gRPC ContentType proto
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

---

## 4. Domain Rules

### Content-Type Schema as Code

- Structure defined in JSON files under `content-types/*.json` — never via API or UI.
- On every API startup, `usecase/content_type.Sync` reconciles definitions against DB:
  - **New file** → create ContentType + per-content-type document collection with indexes
  - **Changed file** → update ContentType schema in place
  - **Field removed** → drop from schema, leave stored data untouched
  - **File deleted** → delete ContentType, cascade-delete all entries, drop collection
- Sync is one-directional: JSON definitions are source of truth.
- JSON schemas declare only content fields. System fields (`createdAt`, `updatedAt`, `publishedAt`, `createdBy`, `updatedBy`, `publishedBy`, `locale`) are injected automatically.

### Content Type Kinds

- **Single-type**: at most one entry per content type. No auto-created singleton — first explicit Save creates it. UI: edit + Save + Publish only. No create/delete.
- **Collection-type**: zero or more entries, each with own `documentId`. List + create/edit/delete, each with independent draft/published pair.

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

---

## 7. API Contracts

### REST — Content-Type Routes

| Method | Route | Permission | Response |
|---|---|---|---|
| `GET` | `/api/content-types` | `content_types:read` | `ContentTypeSummary[]` |
| `GET` | `/api/content-types/:identifier` | `content_types:read` | `ContentType` (full) |

### gRPC — ContentTypeService

```protobuf
service ContentTypeService {
    rpc GetContentType(GetContentTypeRequest) returns (ContentType);
    rpc ListContentTypes(ListContentTypesRequest) returns (ListContentTypesResponse);
}
```

---

## 8. Testing

**Schema sync (`sync_test.go`):**
- New file → creates ContentType + collection
- Changed file → updates schema
- Removed field → drops from schema, data untouched
- Deleted file → cascade-deletes type + entries + collection

**Schema loader (`schema_loader_test.go`):**
- Valid JSON → parses correctly
- `listFields` referencing invalid field → error
- Malformed JSON → error

**Content-type usecase (`content_type_usecase_test.go`):**
- Valid slugs accepted; invalid rejected
- CRUD operations

---

## 9. Boundaries

| Rule | Detail |
|---|---|
| **Always** | Validate slug format at both usecase (creation) and handler (every request) levels |
| **Always** | Validate `listFields` against `fields` during schema sync startup |
| **Always** | Return 400 (not 500) for invalid slug format |
| **Never** | Allow slug characters outside `[a-z0-9-]` |
| **Never** | Let content-type sync write back to JSON definition files |
| **Never** | Add API/UI path to create/edit/delete ContentType structure |

---

## 10. Changelog

| Date | Change | Source |
|------|--------|--------|
| v1.0 | Content types, draft/publish workflow | §1, §4 |
| v1.1 | Schema-as-code (JSON definitions, startup sync) | §4 |
| v1.2 | Slug + documentID input validation | §9.1–§9.3 |
| v1.5 | PostgreSQL per-content-type document tables (`documents_<slug_underscored>`) replacing single `documents` table | §9 |
| v1.6 | PostgreSQL component tables (`components_<slug_underscored>_<component_name_underscored>`) — MongoDB keeps nested BSON | §9 |
| v1.10 | `EnsureCollection` accepts `[]FieldDefinition`, uses DROP+CREATE with per-field columns | sync-table-fields |
