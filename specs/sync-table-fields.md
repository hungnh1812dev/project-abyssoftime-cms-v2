# SPEC — Sync Table Field Names & Schema Alignment

Aligns the GORM database schema with the intended design: renames primary key columns, standardizes ID formats, removes obsolete columns, replaces the JSON `data` column with per-field columns, fixes the `published_at` nullable type, updates GraphQL response shape for media/components, and fixes the MediaInput aspect ratio bug.

---

## Objective

Bring the PostgreSQL schema produced by GORM AutoMigrate into a consistent, predictable state so that:
1. Every static table uses `gorm_id` as its auto-increment primary key column name.
2. Every `document_id` column uses the same UUID v4 format.
3. Dynamic tables (`documents_{slug}`, `components_{slug}_{component}`) store each content-type field as its own column — no catch-all `data` JSON column.
4. Unused/legacy columns are removed.
5. `published_at` is nullable (NULL when unpublished, not zero-value).
6. The GraphQL API returns media fields as objects (`{ url }`) and components as nested objects.
7. The MediaInput frontend component preserves the original aspect ratio of images.

---

## 1. Rename Primary Key Column: `id` → `gorm_id`

### Affected Tables

| Table | Entity File | Current Tag | New Tag |
|-------|------------|-------------|---------|
| `users` | `entity/user.go` | `gorm:"column:id;primaryKey"` | `gorm:"column:gorm_id;primaryKey;autoIncrement"` |
| `roles` | `entity/role.go` (RoleEntity) | `gorm:"column:id;primaryKey"` | `gorm:"column:gorm_id;primaryKey;autoIncrement"` |
| `media_assets` | `entity/media_asset.go` | `gorm:"column:id;primaryKey"` | `gorm:"column:gorm_id;primaryKey;autoIncrement"` |
| `invites` | `entity/invite.go` | `gorm:"column:id;primaryKey"` | `gorm:"column:gorm_id;primaryKey;autoIncrement"` |
| `content_types` | `entity/content_type.go` | `gorm:"column:id;primaryKey"` | `gorm:"column:gorm_id;primaryKey;autoIncrement"` |
| `access_tokens` | `entity/access_token.go` | `gorm:"column:id;primaryKey"` | `gorm:"column:gorm_id;primaryKey;autoIncrement"` |

### Rules

- **Go field name stays `ID`** — only the GORM column tag changes.
- Field type changes from `string` to `uint` (auto-increment integer PK, matching `Document.GormID`).
- The existing `DocumentID string` field remains the application-level unique identifier (UUID). All repository lookups (`FindByID`, `FindByDocumentID`, etc.) must be audited to use the correct column.
- All repository code that queries by primary key (`WHERE id = ?`) must be updated to `WHERE gorm_id = ?` or switched to query by `document_id`.

### Migration Strategy

Fresh start — drop and recreate tables via `AutoMigrate` after entity struct changes. Acceptable for dev.

---

## 2. Standardize `document_id` Format → UUID v4

### Current State

| Generator | Location | Format |
|-----------|----------|--------|
| `uuid.New().String()` | `auth_usecase.go`, `document_usecase.go`, `content_type_usecase.go` | UUID v4 (`550e8400-e29b-41d4-a716-446655440000`) |
| `generateDocID()` (random hex) | `role_usecase.go` | 24-char hex (`a1b2c3d4e5f6a7b8c9d0e1f2`) |

### Fix

Replace `generateDocID()` in `role_usecase.go` with `uuid.New().String()`. Remove the `generateDocID` function entirely.

### Scope

| File | Change |
|------|--------|
| `apps/api/internal/usecase/role/role_usecase.go` | Replace `generateDocID()` calls with `uuid.New().String()`, remove `generateDocID` func, add `"github.com/google/uuid"` import |
| `apps/api/internal/usecase/role/role_usecase_test.go` | Update assertions to expect UUID format |

---

## 3. Add Missing Columns

After the `id` → `gorm_id` rename (section 1), these tables need additional columns present in the struct but missing from the current DB:

| Table | Missing Columns | Fix |
|-------|----------------|-----|
| `media_assets` | `gorm_id` (new PK), `document_id` | Covered by entity struct change + AutoMigrate |
| `invites` | `gorm_id` (new PK) | Covered by entity struct change + AutoMigrate |
| `access_tokens` | `gorm_id` (new PK) | Covered by entity struct change + AutoMigrate |

No manual migration needed — `AutoMigrate` after struct update handles this.

---

## 4. Remove Obsolete Columns

### 4a. `media_assets`: Remove `content_type_id` and `document_ref`

These columns are legacy from the MongoDB era where media was scoped to a content type and document. Media assets are now linked via their `document_id` field — document/component media columns store this UUID as a foreign key reference.

| File | Change |
|------|--------|
| `apps/api/internal/domain/entity/media_asset.go` | Remove `ContentTypeID` and `DocumentRef` fields |
| `apps/api/internal/infrastructure/gormdb/media_asset_repository.go` | Remove `FindByDocumentRef`, `DeleteByDocumentRef` methods; add `FindByDocumentID(ctx, documentID)` method (lookup by `document_id` column) |
| `apps/api/internal/domain/repository/media_asset_repository.go` | Remove `FindByDocumentRef`, `DeleteByDocumentRef` from interface; add `FindByDocumentID` |
| All callers of removed methods | Update or remove |

### 4b. `documents_{slug}`: Remove `content_type_id`

The content type is already encoded in the table name (`documents_blog_posts`). The column is redundant.

| File | Change |
|------|--------|
| `apps/api/internal/domain/entity/document.go` | Remove `ContentTypeID` field |
| Repository/handler code referencing `ContentTypeID` | Remove |

### 4c. `components_{slug}_{component}`: Remove `order`

`order` is not a default schema field — it was added incorrectly. Component ordering is determined by array position in the parent document's component field.

| File | Change |
|------|--------|
| `apps/api/internal/domain/entity/component.go` | Remove `Order int` field |
| `apps/api/internal/infrastructure/gormdb/component_repository.go` | Remove `Order` assignment in `UpsertAll`; remove `ORDER BY "order"` clause |

---

## 5. Fix `published_at` → Nullable (`*time.Time`)

### Problem

`Document.PublishedAt` is `time.Time` (non-pointer). GORM stores the zero value (`0001-01-01 00:00:00`) instead of `NULL` for unpublished documents.

### Fix

| File | Change |
|------|--------|
| `apps/api/internal/domain/entity/document.go` | Change `PublishedAt time.Time` → `PublishedAt *time.Time` |
| `apps/api/internal/usecase/document/document_usecase.go` | Update Publish logic: `doc.PublishedAt = &now` |
| `apps/api/internal/delivery/grpc/document_service.go` | Update nil check: `if d.PublishedAt != nil` |
| `apps/api/graphql/dynamic/resolver_factory.go` | Update `docToMap`: `if d.PublishedAt != nil { m["publishedAt"] = *d.PublishedAt }` |
| All test files referencing `PublishedAt` | Update to pointer semantics |

---

## 6. Replace `data` JSON Column with Per-Field Columns

### Current Design

`documents_{slug}` and `components_{slug}_{component}` tables store all content in a single `data` JSONB column:

```go
Data map[string]any `gorm:"column:data;serializer:json"`
```

### New Design

Each field defined in the content-type schema becomes its own column on the table. The `data` field is removed from the entity struct.

**Media field storage:** Media fields do NOT store a URL string directly. They store the `document_id` (UUID) of the referenced `media_assets` row. This creates a logical foreign key relationship: `documents_{slug}.{media_field}` → `media_assets.document_id`. The actual URL is resolved at query time by joining/looking up the media asset.

**Example:** Content type `blog-posts` with fields `[{name: "title", type: "text"}, {name: "coverImage", type: "media"}, {name: "body", type: "richtext"}]` produces table `documents_blog_posts`:

| Column | Type | Source |
|--------|------|--------|
| `gorm_id` | SERIAL PK | system |
| `document_id` | VARCHAR | system |
| `version` | VARCHAR(20) | system |
| `locale` | VARCHAR | system |
| `title` | TEXT | field |
| `cover_image` | VARCHAR | field (FK → `media_assets.document_id`) |
| `body` | TEXT | field |
| `created_at` | TIMESTAMP | system |
| `updated_at` | TIMESTAMP | system |
| `published_at` | TIMESTAMP NULL | system |
| `created_by` | VARCHAR | system |
| `updated_by` | VARCHAR | system |
| `published_by` | VARCHAR | system |

**Components** follow the same pattern. Content type `blog-posts` with component field `banner` having sub-fields `[{name: "background", type: "media"}, {name: "caption", type: "text"}]` produces table `components_blog_posts_banner`:

| Column | Type | Source |
|--------|------|--------|
| `gorm_id` | SERIAL PK | system |
| `component_id` | VARCHAR | system |
| `document_id` | VARCHAR | system |
| `version` | VARCHAR(20) | system |
| `locale` | VARCHAR | system |
| `background` | VARCHAR | field (FK → `media_assets.document_id`) |
| `caption` | TEXT | field |
| `created_at` | TIMESTAMP | system |
| `updated_at` | TIMESTAMP | system |

### Implementation Approach

Since dynamic tables have columns determined at runtime by the content-type schema, GORM struct tags can't define them statically. Use raw SQL / GORM Migrator to:

1. **On sync (`EnsureCollection`)**: Create table with system columns + one column per field from the schema definition.
2. **On schema change (field added/removed)**: Drop and recreate the table (fresh start — acceptable for dev).
3. **On read/write**: Use `map[string]any` with GORM's `Table().Create/Find` using column maps instead of struct reflection.

### Entity Changes

```go
// Document — remove Data field, keep system fields only
type Document struct {
    GormID      uint             `gorm:"column:gorm_id;primaryKey;autoIncrement"`
    DocumentID  string           `gorm:"column:document_id"`
    Version     DocumentVersion  `gorm:"column:version;type:varchar(20)"`
    Locale      string           `gorm:"column:locale"`
    CreatedAt   time.Time        `gorm:"column:created_at"`
    UpdatedAt   time.Time        `gorm:"column:updated_at"`
    PublishedAt *time.Time       `gorm:"column:published_at"`
    CreatedBy   string           `gorm:"column:created_by"`
    UpdatedBy   string           `gorm:"column:updated_by"`
    PublishedBy string           `gorm:"column:published_by"`
    // Content fields are stored as dynamic columns, not in this struct.
    // Use Fields map for in-memory representation during read/write.
    Fields      map[string]any   `gorm:"-"`
}

// Component — remove Data and Order fields
type Component struct {
    GormID      uint             `gorm:"column:gorm_id;primaryKey;autoIncrement"`
    ComponentID string           `gorm:"column:component_id"`
    DocumentID  string           `gorm:"column:document_id"`
    Version     DocumentVersion  `gorm:"column:version;type:varchar(20)"`
    Locale      string           `gorm:"column:locale"`
    CreatedAt   time.Time        `gorm:"column:created_at"`
    UpdatedAt   time.Time        `gorm:"column:updated_at"`
    // Content fields stored as dynamic columns.
    Fields      map[string]any   `gorm:"-"`
}
```

### Repository Changes

The document and component repositories need new logic to:

1. **`EnsureCollection`**: Build `CREATE TABLE` SQL from content-type field definitions, mapping field types to Postgres column types:
   - `text`, `richtext` → `TEXT`
   - `media` → `VARCHAR` (stores `document_id` UUID referencing `media_assets`)
   - `number` → `DOUBLE PRECISION`
   - `boolean` → `BOOLEAN`
   - `json` → `JSONB`

2. **Write (Create/Update)**: Build column-value maps from `Document.Fields` + system fields, then use `db.Table(name).Create(map)` or `db.Table(name).Save(map)`.

3. **Read (Find)**: Use `db.Table(name).Find(&results)` scanning into `[]map[string]any`, then hydrate `Document.Fields` from the non-system columns.

### Migration Strategy

**Fresh start** — on first run with new code, drop all `documents_*` and `components_*` tables and recreate from content-type definitions. Data loss is acceptable for dev.

### Scope

| File | Change |
|------|--------|
| `apps/api/internal/domain/entity/document.go` | Remove `Data`, `ContentTypeID`; add `Fields map[string]any` with `gorm:"-"` tag; change `PublishedAt` to `*time.Time` |
| `apps/api/internal/domain/entity/component.go` | Remove `Data`, `Order`; add `Fields map[string]any` with `gorm:"-"` tag |
| `apps/api/internal/infrastructure/gormdb/document_repository.go` | Rewrite `EnsureCollection` to create per-field columns; rewrite CRUD to use column maps |
| `apps/api/internal/infrastructure/gormdb/component_repository.go` | Same as above |
| `apps/api/internal/usecase/document/document_usecase.go` | Update to use `doc.Fields` instead of `doc.Data` |
| `apps/api/internal/usecase/content_type/sync.go` | Drop+recreate tables on schema change |
| `apps/api/graphql/dynamic/resolver_factory.go` | Update `docToMap` to use `doc.Fields` |
| `apps/api/internal/delivery/http/handler/document_handler.go` | Update request/response mapping |
| `apps/api/internal/delivery/grpc/document_service.go` | Update proto mapping |
| All test files | Update for new entity shape |

---

## 7. GraphQL Response Shape — Media as Object, Components as Nested Object

### Current Behavior

- Media fields return a plain `String` (URL).
- Components return a flat object with scalar fields.
- Queries return wrapped responses: `{ data: BlogPost }` or `{ data: [BlogPost!]!, total, start, size }`.

### New Behavior

**Single query** (e.g., `blogPost(blogPostId: "...", locale: "en")`):
```json
{
  "data": {
    "blogPost": {
      "title": "...",
      "coverImage": { "url": "..." },
      "banner": {
        "background": { "url": "..." }
      }
    }
  }
}
```

**List query** (e.g., `blogPostList(start: 0, size: 20)`):
```json
{
  "data": {
    "blogPostList": [
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

### Key Changes

1. **Media fields → `MediaAsset` object type** instead of `String`:
   ```graphql
   type MediaAsset {
     url: String!
     thumbnailUrl: String
     fileName: String
     width: Int
     height: Int
   }
   ```
   The resolver looks up the media asset by `document_id` (stored in the field column) and returns the asset object.

2. **Component fields → nested object** with their sub-fields fully resolved (including media sub-fields as `MediaAsset` objects).

3. **Remove Response wrapper types** — single queries return the object type directly (nullable for "not found"), list queries return an array directly:
   ```graphql
   # Before
   blogPost(blogPostId: ID!): BlogPostResponse  # { data: BlogPost }
   blogPostList(...): BlogPostListResponse!      # { data: [...], total, start, size }

   # After
   blogPost(blogPostId: ID!, locale: String): BlogPost
   blogPostList(start: Int, size: Int, locale: String): [BlogPost!]!
   ```

4. **Pagination metadata** — if needed, provide via a separate `_meta` query or field. The list query itself returns just the array.

### Schema Changes

| File | Change |
|------|--------|
| `apps/api/graphql/dynamic/schema_builder.go` | Add `MediaAsset` type; change `media` field type from `String` to `MediaAsset`; remove `Response`/`ListResponse` wrapper types; update query return types |
| `apps/api/graphql/dynamic/resolver_factory.go` | Remove response wrapping; add media asset lookup in `docToMap`; resolve component sub-fields including nested media |
| `apps/api/graphql/dynamic/resolver_factory.go` | `buildObjectType`: media fields get `MediaAsset` type; component types resolve media sub-fields as objects |

### Resolver Logic for Media Fields

Media fields store a `document_id` reference to `media_assets`. The resolver must look up the actual asset to return the full object:

```go
// For each field in the content-type schema:
if field.Type == "media" {
    assetDocID := doc.Fields[field.Name].(string)  // UUID referencing media_assets.document_id
    if assetDocID != "" {
        asset, _ := mediaRepo.FindByDocumentID(ctx, assetDocID)
        if asset != nil {
            m[field.Name] = map[string]any{
                "url":          asset.URL,
                "thumbnailUrl": asset.ThumbnailURL,
                "fileName":     asset.FileName,
                "width":        asset.Width,
                "height":       asset.Height,
            }
        }
    }
}
```

### Resolver Logic for Component Fields

```go
if field.Type == "component" {
    components, _ := compRepo.FindByDocumentID(ctx, slug, field.Name, docID, locale, version)
    if len(components) == 1 {
        // Single component → nested object
        compMap := map[string]any{}
        for _, subField := range field.Fields {
            if subField.Type == "media" {
                // resolve media asset object
            } else {
                compMap[subField.Name] = components[0].Fields[subField.Name]
            }
        }
        m[field.Name] = compMap
    }
}
```

---

## 8. Fix MediaInput Aspect Ratio Bug + Store Reference

### Problem

1. **Aspect ratio**: The `MediaInput` component (`apps/web/src/components/form/inputs/MediaInput.tsx`) displays selected images inside a fixed-height container (`min-h-28 max-h-28` = 112px). The `object-contain` class preserves aspect ratio within the box, but the container clips images to 112px height regardless of their natural dimensions, making wide/panoramic images appear as thin slivers and tall images fill the box awkwardly.

2. **Value storage**: Currently `field.onChange(asset.thumbnailUrl || asset.url)` stores a direct URL string. It should store `asset.documentId` (the media asset's UUID) so the backend can resolve the relationship. The preview image is displayed using the URL from the selected asset, but the form value submitted to the API is the `document_id` reference.

### Fix

Replace the fixed-height container with an aspect-ratio-aware approach:

```tsx
<div className="cursor-pointer border border-input rounded-md transition-colors hover:border-ring relative overflow-hidden">
  {displayUrl ? (
    <img
      src={displayUrl}
      alt="media preview"
      className="w-full h-auto max-h-40 object-contain"
    />
  ) : (
    <div className="flex flex-col items-center justify-center h-28 gap-2 text-muted-foreground">
      ...
    </div>
  )}
</div>
```

Key changes:
- Remove `min-h-28 max-h-28` from the outer container (let it size to content).
- Use `h-auto` on the `<img>` so it respects natural aspect ratio.
- Keep `max-h-40` as an upper bound so very tall images don't blow out the layout.
- The empty state placeholder keeps `h-28` for consistent sizing when no image is selected.
- Change `field.onChange(asset.thumbnailUrl || asset.url)` → `field.onChange(asset.documentId)` to store the media library reference.
- Display the preview using a local state or the asset's URL (looked up by `documentId`), not the form field value.

### Scope

| File | Change |
|------|--------|
| `apps/web/src/components/form/inputs/MediaInput.tsx` | Update container/img classes for aspect ratio; store `documentId` as form value; display preview via asset URL lookup |
| `apps/web/src/components/form/inputs/__tests__/MediaInput.test.tsx` | Update assertions for new value format and classes |

---

## Implementation Order

| Phase | Task | Depends On |
|-------|------|-----------|
| 1 | Rename PK columns `id` → `gorm_id` (entity structs) | — |
| 2 | Standardize `document_id` to UUID | — |
| 3 | Remove obsolete columns (entity structs) | — |
| 4 | Fix `published_at` → `*time.Time` | — |
| 5 | Replace `data` column with per-field columns (entity + repository rewrite) | Phases 1-4 |
| 6 | Update GraphQL schema + resolvers (media objects, no wrappers) | Phase 5 |
| 7 | Update REST handlers and gRPC service | Phase 5 |
| 8 | Fix MediaInput aspect ratio | Independent |
| 9 | Update all tests | Phases 5-7 |

Phases 1-4 can be done in a single commit (entity struct cleanup).
Phase 5 is the largest — repository rewrite for dynamic columns.
Phase 8 is independent and can be done anytime.

---

## Testing Strategy

- **Repository tests**: Verify dynamic table creation, CRUD with per-field columns, schema changes trigger table recreation.
- **Usecase tests**: Verify `doc.Fields` replaces `doc.Data` in all flows (save, publish, get).
- **GraphQL tests**: Verify media fields return `{ url, ... }` objects; component fields return nested objects; no Response wrappers.
- **Handler tests**: Verify REST/gRPC responses use new field structure.
- **Frontend**: Manual test MediaInput with various aspect ratios (landscape, portrait, square, panoramic).

---

## Boundaries

| Rule | Detail |
|------|--------|
| **Always** | Use `gorm_id` (auto-increment uint) as the PK column for all tables |
| **Always** | Use `uuid.New().String()` for all `document_id` generation |
| **Always** | Store `NULL` for `published_at` when document is not published |
| **Always** | Map each content-type field to its own DB column |
| **Always** | Return media fields as `{ url, ... }` objects in GraphQL |
| **Never** | Store content in a catch-all JSON `data` column |
| **Never** | Use the `order` column on component tables |
| **Never** | Expose `content_type_id` in document tables (redundant with table name) |
| **Ask first** | Whether list queries need pagination metadata (`total`, `start`, `size`) alongside the array or via a separate mechanism |
