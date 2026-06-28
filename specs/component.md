# SPEC — Component Module

## 1. Overview

The component module manages the storage and lifecycle of component fields within documents. In PostgreSQL (GORM), component fields are stored in dedicated per-content-type component tables, separate from the main document table. In MongoDB, components remain nested in the document's BSON `data` field and require no separate storage. This module defines the Component entity, ComponentRepository interface, the component table creation/teardown during content-type sync, and the document save/read flow for extracting and merging component data in PostgreSQL.

---

## 2. File Map

All paths relative to `apps/api/`.

```
internal/domain/entity/component.go                          # Component entity (if separate file)
internal/domain/repository/component_repository.go           # ComponentRepository interface
internal/infrastructure/gormdb/component_repository.go       # GORM component repo (PostgreSQL only)
internal/infrastructure/gormdb/component_repository_test.go
```

---

## 3. Infrastructure — GORM Specifics

### Per-Content-Type Document Tables

GORM creates one table per content type, matching MongoDB's per-collection pattern. Tables are created dynamically during content-type sync, not via `AutoMigrate`.

**Table naming:** `documents_<slug_underscored>` (hyphens replaced with underscores).
Example: content type `blog-posts` → table `documents_blog_posts`.

**Slug sanitization:** Only `[a-z0-9-]` allowed in slugs (validated at content-type creation). Hyphens converted to underscores for PostgreSQL table names.

```sql
CREATE TABLE IF NOT EXISTS documents_<slug_underscored> (
    gorm_id         BIGSERIAL PRIMARY KEY,
    document_id     UUID         NOT NULL,
    version         VARCHAR(20)  NOT NULL,
    locale          VARCHAR(10)  NOT NULL,
    -- per-field columns (type depends on FieldDefinition):
    --   text/richtext → TEXT
    --   media         → VARCHAR (stores documentId FK to media_assets)
    --   number        → REAL
    --   boolean       → BOOLEAN
    --   json          → TEXT
    <field_name_1>  <mapped_type>,
    <field_name_2>  <mapped_type>,
    ...
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
- `EnsureCollection(slug, fields []FieldDefinition)`: DROP+CREATE strategy — drops existing table and recreates with per-field columns based on `FieldDefinition` type mapping
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
    document_id   UUID         NOT NULL,
    version       VARCHAR(20)  NOT NULL,
    locale        VARCHAR(10)  NOT NULL,
    -- per-field columns based on component's FieldDefinition (same type mapping as documents)
    <field_name_1>  <mapped_type>,
    <field_name_2>  <mapped_type>,
    ...
    created_at    TIMESTAMPTZ,
    updated_at    TIMESTAMPTZ,
    UNIQUE(document_id, version, locale)
);
```

---

## 4. Entities

### Component

```go
type Component struct {
    GormID      uint            `gorm:"column:gorm_id;primaryKey;autoIncrement"`
    ComponentID string          `gorm:"column:component_id"`
    DocumentID  string          `gorm:"column:document_id"`
    Version     DocumentVersion `gorm:"column:version;type:varchar(20)"`
    Locale      string          `gorm:"column:locale"`
    Fields      map[string]any  `gorm:"-"`
    CreatedAt   time.Time       `gorm:"column:created_at"`
    UpdatedAt   time.Time       `gorm:"column:updated_at"`
}
```

**Component entity changes (v1.8):** Removed `Order` field; renamed `Data` → `Fields` (tagged `gorm:"-"` — per-field columns used instead of JSON blob, matching the document table pattern).

---

## 5. Repository Interfaces

### ComponentRepository

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

---

## 6. Sync Engine Changes

- `syncOne()`: After ensuring document table, iterate content-type fields — for each `"type": "component"` field, call `componentRepo.EnsureCollection(slug, field.Name)`
- `removeContentType()`: Drop all component tables before dropping the document table
- `Syncer` receives `ComponentRepository` in constructor

---

## 7. Document Save/Read Flow (PostgreSQL)

- **Save:** Extract component fields from `doc.Data` based on content-type field definitions → save scalar data to `documents_<slug_underscored>` → upsert component records to `components_<slug_underscored>_<component_name_underscored>`
- **Read:** Load document from `documents_<slug_underscored>` → for each component field, load from `components_<slug_underscored>_<component_name_underscored>` → merge into `doc.Data` before returning
- **Delete:** Delete component records from all component tables → delete document record
- **Publish:** Copy draft component records to published version → copy draft document to published

**MongoDB (no change):** Components remain nested in the document's BSON `data` field. No component collections are created.

---

## 8. Boundaries

| Rule | Detail |
|---|---|
| **Always** | PostgreSQL component tables created/dropped during content-type sync |
| **Never** | Create component collections in MongoDB — components stay nested in BSON `data` |

---

## 9. Changelog

| Date | Change | Source |
|------|--------|--------|
| v1.6 | PostgreSQL component tables (`components_<slug_underscored>_<component_name_underscored>`) — MongoDB keeps nested BSON | §9 |
| v1.9 | Component entity: removed `Order`, renamed `Data` → `Fields` (gorm:"-"); per-field columns in dynamic tables | sync-table-fields |
| v1.10 | `EnsureCollection` accepts `[]FieldDefinition`, uses DROP+CREATE with per-field columns | sync-table-fields |
