# RULES — Content-Type Parsing (Schema Loader + Sync Engine)

**Scope:** JSON schema loading, field validation, content-type sync on startup, and the FieldDefinition data model.
**Files:** `usecase/content_type/schema_loader.go`, `usecase/content_type/sync.go`, `domain/entity/content_type.go`

---

## 1. JSON Schema File Format

### 1.1 File Location
- Directory: `apps/api/content-types/*.json`
- One file per content type
- File name = slug (convention, not enforced)
- Path configurable via `CONTENT_TYPES_DIR` env var (default: `content-types`)

### 1.2 Required Top-Level Fields
```json
{
  "slug": "blog-posts",          // REQUIRED — unique, validated format
  "name": "Blog Posts",           // REQUIRED — display name
  "kind": "collection",          // REQUIRED — "single" or "collection"
  "fields": [...]                 // REQUIRED — array of FieldDefinition
}
```

### 1.3 Removed/Ignored Fields
- `"listFields"` — permanently removed from schema format; ignored if present (backward-compatible)
- `"description"` — not part of the schema; ignored
- Any unknown keys — silently ignored by `json.Unmarshal`

---

## 2. FieldDefinition Data Model

### 2.1 Go Struct
```go
type FieldDefinition struct {
    Name       string            `json:"name"                    bson:"name"`
    Type       string            `json:"type"                    bson:"type"`
    Ext        []string          `json:"ext,omitempty"           bson:"ext,omitempty"`
    Width      string            `json:"width,omitempty"         bson:"width,omitempty"`
    Repeatable bool              `json:"repeatable,omitempty"    bson:"repeatable,omitempty"`
    Fields     []FieldDefinition `json:"fields,omitempty"        bson:"fields,omitempty"`
}
```
- `Width`: UI hint for form column span. Values: `"100%"` (default), `"50%"`, `"1/3"`. No effect on storage.

### 2.2 Field Types

| Type | Description | Has Sub-Fields | Stored in |
|---|---|---|---|
| `text` | Short text, URLs | No | Document table column (TEXT) |
| `richtext` | HTML content | No | Document table column (TEXT) |
| `number` | Integer/float | No | Document table column (REAL) |
| `boolean` | True/false | No | Document table column (BOOLEAN) |
| `media` | Media reference | No | Document table column (TEXT — stores documentId) |
| `json` | Arbitrary JSON | No | Document table column (TEXT — serialized JSON string) |
| `component` | Nested component | **Yes** — has `fields` | PostgreSQL: separate table; MongoDB: nested BSON |

### 2.3 Component Fields
```json
{
  "name": "banner",
  "type": "component",
  "repeatable": false,
  "fields": [
    { "name": "title", "type": "text" },
    { "name": "background", "type": "media" }
  ]
}
```
- `name`: required, non-empty — used as component table name suffix
- `repeatable`: optional, defaults to `false`
  - `false` → single object
  - `true` → ordered array of objects
- `fields`: sub-field definitions (recursive)
- May have additional `"component"` key in JSON — ignored by Go (not in struct)
- Component fields can contain other components (nesting)

### 2.4 Nesting Depth Limits
- Maximum: **3 levels** of component nesting
- Counting starts at depth=1 for top-level fields

| Depth | Example | Status |
|---|---|---|
| 1 | `document → component` | OK |
| 2 | `document → component → component` | OK |
| 3 | `document → component → component → component` | OK |
| 4 | `document → comp → comp → comp → comp` | **FATAL ERROR** |

---

## 3. Schema Loader (`schema_loader.go`)

### 3.1 `LoadDefinitions(dir string) ([]ContentTypeDefinition, error)`
```go
type ContentTypeDefinition struct {
    Slug       string                   `json:"slug"`
    Name       string                   `json:"name"`
    Kind       string                   `json:"kind"`
    ListFields []string                 `json:"listFields,omitempty"`
    Fields     []entity.FieldDefinition `json:"fields"`
}
```
- Reads all `*.json` files in `dir` (skips directories, non-JSON files)
- For each file: `os.ReadFile` → `json.Unmarshal` → `validateDefinition`
- Returns all definitions or first error
- `ListFields` kept in struct for backward compatibility (parsed but never used)

### 3.2 Loading Steps
1. `os.ReadDir(dir)` — list directory entries
2. Filter: skip directories, keep only `*.json` suffix
3. `os.ReadFile(path)` — read file bytes
4. `json.Unmarshal(data, &def)` — parse into `ContentTypeDefinition`
5. `validateDefinition(def, path)` — validate fields recursively
6. Append to results

### 3.3 Error Messages
- Directory read error: `"read content-type definitions dir %q: %w"`
- File read error: `"read %q: %w"`
- Parse error: `"parse %q: %w"`
- Validation errors: include file path and field name

---

## 4. Field Validation (`validateFields`)

### 4.1 Signature
```go
func validateFields(fields []entity.FieldDefinition, path string, depth int) error
```

### 4.2 Validation Rules Per Type

**`component` fields:**
- `Name` must be non-empty → error: `"component field must have a non-empty name"`
- `depth > 3` → FATAL error: `"component %q exceeds maximum nesting depth of 3"`
- Recursively validates sub-fields with `depth+1`

**Other types (text, richtext, number, boolean, media, json):**
- No validation — pass through

### 4.3 What Is NOT Validated
- Slug format (validated at usecase level, not loader)
- `Kind` value (no check for "single"/"collection" in loader)
- `Repeatable` on non-component fields (ignored, no error)
- `listFields` (removed — no validation)
- Field name uniqueness (not checked)
- Field name format (not checked — any string accepted)

---

## 5. Sync Engine (`sync.go`)

### 5.1 Architecture
```go
type Syncer struct {
    *UseCase                              // inherits ContentType CRUD
    entries  EntryManager                  // document usecase subset
    docRepo  repository.DocumentRepository // for EnsureCollection/DropCollection
    compRepo repository.ComponentRepository // for component tables (may be nil)
}
```

### 5.2 `EntryManager` Interface
```go
type EntryManager interface {
    GetAll(ctx context.Context, contentTypeSlug string) ([]*entity.Document, error)
    Delete(ctx context.Context, contentTypeSlug, documentID string, fields []entity.FieldDefinition) error
}
```
- Subset of document usecase needed by sync
- `Delete` takes `documentID` (not MongoDB `_id`)
- `fields` parameter needed for cascade component deletion

### 5.3 `Sync(ctx, defs)` — Main Entry Point
```
1. Fetch all existing content types from DB
2. For each JSON definition:
   a. Call syncOne(ctx, def)
3. For each existing content type NOT in definitions:
   a. Call removeContentType(ctx, ct)
```
- Builds `defSlugs` set from definitions for O(1) lookup
- Removals are ordered: existing types not in defs get removed

### 5.4 `syncOne(ctx, def)` — Per-Definition Logic
```
1. Log table info (exists? row count?)
2. Try FindBySlug(ctx, def.Slug)
   a. If NOT FOUND:
      - Create ContentType entity (Name, Slug, Kind, Fields)
      - EnsureCollection(slug, fields) → create document table
      - ensureComponentTables(slug, fields) → create component tables
   b. If FOUND:
      - EnsureCollection(slug, fields) → add missing columns
      - ensureComponentTables(slug, fields) → add missing component tables/columns
      - Compare: Name, Kind, Fields (NOT ListFields)
      - If changed: update ContentType (Name, Kind, Fields only)
      - ListFields is NEVER overwritten
```

### 5.5 Change Detection — `fieldsEqual`
```go
func fieldsEqual(a, b []entity.FieldDefinition) bool {
    if len(a) != len(b) { return false }
    for i := range a {
        if a[i].Name != b[i].Name || a[i].Type != b[i].Type || a[i].Width != b[i].Width || a[i].Repeatable != b[i].Repeatable {
            return false
        }
        if !fieldsEqual(a[i].Fields, b[i].Fields) { return false }
    }
    return true
}
```
- Compares: `Name`, `Type`, `Width`, `Repeatable`, and sub-`Fields` (recursive)
- Does NOT compare: `Ext`, `component` key (JSON-only alias), `ListFields`
- Order-sensitive — field order matters

### 5.6 What Triggers a DB Update
| Change | Triggers Update? |
|---|---|
| Name changed | Yes |
| Kind changed | Yes |
| Field added | Yes |
| Field removed | Yes |
| Field type changed | Yes |
| Field `repeatable` toggled | Yes |
| Field `width` changed | Yes |
| Field order changed | Yes |
| `listFields` changed | **No** — never compared, never overwritten |
| `Ext` changed | **No** — not compared |

### 5.7 `ensureComponentTables` — Recursive
```go
func ensureComponentTablesRecursive(ctx, slug, prefix string, fields []FieldDefinition, depth int) error
```
- Iterates fields, finds `type == "component"`
- Builds component path: `fieldName` (depth 0) or `prefix_fieldName` (nested)
- `isNested = depth > 0` — determines FK column (document_id vs parent_component_id)
- Calls `compRepo.EnsureCollection(slug, path, field.Fields, isNested)`
- Recurses into component's sub-fields with `depth+1`
- **Skipped entirely** if `compRepo == nil` (MongoDB mode)

### 5.8 Component Table Path Building

Given `cv-page` with:
```
experiences (component, level 1)
  → roles (component, level 2)
    → (no level 3 here)
projects (component, level 1)
  → roles (component, level 2)
```

| depth=0 call | prefix="" | Component | Path | isNested |
|---|---|---|---|---|
| experiences | "" | experiences | `experiences` | false |
| projects | "" | projects | `projects` | false |

| depth=1 call | prefix | Component | Path | isNested |
|---|---|---|---|---|
| experiences/roles | "experiences" | roles | `experiences_roles` | true |
| projects/roles | "projects" | roles | `projects_roles` | true |

Resulting tables:
- `components_cv_page_experiences` (document_id FK)
- `components_cv_page_experiences_roles` (parent_component_id FK)
- `components_cv_page_projects` (document_id FK)
- `components_cv_page_projects_roles` (parent_component_id FK)

### 5.9 `removeContentType` — Cascade Deletion
```
1. GetAll(ctx, slug) → fetch all draft documents
2. For each document: Delete(ctx, slug, doc.DocumentID, ct.Fields)
3. dropComponentTables(ctx, slug, ct.Fields) → deepest first
4. DropCollection(ctx, slug) → drop document table
5. Delete content type record from DB
```
- Component tables dropped **before** document table
- Drop order: deepest children first (recursive bottom-up)
- Documents deleted through `EntryManager.Delete` (handles component cascade)

### 5.10 `dropComponentTablesRecursive` — Bottom-Up
```go
func dropComponentTablesRecursive(ctx, slug, prefix string, fields []FieldDefinition) error {
    for _, f := range fields {
        if f.Type == "component" {
            path := f.Name; if prefix != "" { path = prefix + "_" + f.Name }
            // First: recurse into children (deeper tables dropped first)
            dropComponentTablesRecursive(ctx, slug, path, f.Fields)
            // Then: drop this component's table
            compRepo.DropCollection(ctx, slug, path)
        }
    }
}
```
- Recursive descent → drop deepest tables first
- Prevents FK reference issues (children removed before parents)

---

## 6. Startup Logging

```go
if exists, count, err := s.docRepo.TableInfo(ctx, def.Slug); err == nil {
    if exists {
        log.Printf("sync: %q — table exists, %d rows", def.Slug, count)
    } else {
        log.Printf("sync: %q — table not found, will create", def.Slug)
    }
}
```
- Logs each content-type's table status before sync
- Uses `TableInfo(ctx, slug) (bool, int64, error)` from DocumentRepository
- Provides observability for production cold starts

---

## 7. ContentType Entity — Database Storage

### 7.1 Struct Tags
```go
type ContentType struct {
    ID         uint              `bson:"_id,omitempty"  gorm:"column:gorm_id;primaryKey;autoIncrement"`
    DocumentID string            `bson:"documentId"     gorm:"column:document_id;uniqueIndex"`
    Name       string            `bson:"name"           gorm:"column:name"`
    Slug       string            `bson:"slug"           gorm:"column:slug;uniqueIndex"`
    Kind       ContentKind       `bson:"kind"           gorm:"column:kind;type:varchar(20)"`
    Fields     []FieldDefinition `json:"Fields,omitempty"  bson:"fields,omitempty"  gorm:"column:fields;serializer:json"`
    ListFields []string          `json:"listFields,omitempty" bson:"listFields,omitempty" gorm:"column:list_fields;serializer:json"`
    CreatedAt  time.Time         `bson:"createdAt"      gorm:"column:created_at"`
    UpdatedAt  time.Time         `bson:"updatedAt"      gorm:"column:updated_at"`
}
```

### 7.2 GORM Serialization
- `Fields []FieldDefinition` → stored as JSON string in `fields` column (`serializer:json`)
- `ListFields []string` → stored as JSON string in `list_fields` column (`serializer:json`)
- GORM handles serialize/deserialize automatically via `serializer:json` tag

### 7.3 MongoDB Storage
- `Fields` → stored as nested BSON array (native MongoDB structure)
- `ListFields` → stored as BSON string array
- No serialization needed — BSON handles natively

### 7.4 ContentKind
```go
type ContentKind string
const (
    KindSingle     ContentKind = "single"
    KindCollection ContentKind = "collection"
)
```
- Stored as `varchar(20)` in PostgreSQL, string in MongoDB
- Only two valid values — **NEVER** add new kinds without spec

---

## 8. Real-World Schema Examples

### 8.1 Simple Collection (no components)
```json
{
  "slug": "en-vocab-pack",
  "name": "English Vocabulary",
  "kind": "collection",
  "fields": [
    { "name": "packName", "type": "text", "width": "50%" },
    { "name": "packTitle", "type": "text", "width": "50%" },
    { "name": "words", "type": "json" }
  ]
}
```
- Creates table `documents_en_vocab_pack` with columns: `pack_name TEXT`, `pack_title TEXT`, `words TEXT`
- No component tables
- `width` is a UI hint only — does not affect column creation

### 8.2 Complex Collection (3-level nesting)
```json
{
  "slug": "cv-page",
  "fields": [
    { "name": "position", "type": "text", "width": "50%" },
    { "name": "isMain", "type": "boolean", "width": "50%" },
    { "name": "company", "type": "text" },
    { "name": "experiences", "type": "component", "repeatable": true,
      "fields": [
        { "name": "company", "type": "text", "width": "50%" },
        { "name": "location", "type": "text", "width": "50%" },
        { "name": "roles", "type": "component", "repeatable": true,
          "fields": [...]
        }
      ]
    }
  ]
}
```
- Document table: `documents_cv_page` — columns: `position`, `is_main`, `company`
- Component tables:
  - `components_cv_page_skills` (document_id FK)
  - `components_cv_page_experiences` (document_id FK)
  - `components_cv_page_experiences_roles` (parent_component_id FK)
  - `components_cv_page_projects` (document_id FK)
  - `components_cv_page_projects_roles` (parent_component_id FK)
  - `components_cv_page_educations` (document_id FK)
  - `components_cv_page_languages` (document_id FK)
  - `components_cv_page_references` (document_id FK)

---

## 9. Boundaries Summary

| Rule | Detail |
|---|---|
| **Always** | Validate fields recursively with depth tracking |
| **Always** | Reject component nesting > 3 levels with fatal error |
| **Always** | Skip component fields when creating document table columns |
| **Always** | Build component table path by concatenating names with `_` |
| **Always** | Set `isNested = depth > 0` for component table FK selection |
| **Always** | Drop component tables bottom-up (deepest first) |
| **Always** | Log table existence and row count during sync |
| **Always** | Compare Name, Kind, Width, Fields (not ListFields) for change detection |
| **Always** | Every field in JSON schema must have a `name` — no structural wrappers |
| **Always** | `width` defaults to `"100%"` when omitted — it is a UI hint only |
| **Never** | Overwrite `ListFields` during sync — it's user-managed |
| **Never** | Let sync write back to JSON definition files |
| **Never** | Create content types via API or UI — JSON-only |
| **Never** | Validate `listFields` in schema loader — removed |
| **Never** | Drop tables in `EnsureCollection` — additive only |
| **Never** | Use `type: "layout"` in content-type schemas — removed |
| **Never** | Let `width` affect database column creation or storage |
| **Ask first** | Increasing max nesting depth beyond 3 |
| **Ask first** | Adding new field types |
