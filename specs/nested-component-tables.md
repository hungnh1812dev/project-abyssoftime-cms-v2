# SPEC — Nested Component Tables

## 1. Overview

Changes how nested components (components inside other components) are stored in PostgreSQL. Top-level components (direct children of documents) keep `document_id` as their foreign key. Nested components (children of other components) use `parent_component_id` instead — they have **no** `document_id`. The maximum nesting depth increases from 2 to 3 levels. MongoDB is unchanged — components remain nested in BSON `data`.

---

## 2. Objective

Normalize the relational model for component storage so that:
- A top-level component belongs to a **document** (`document_id` FK)
- A nested component belongs to a **parent component** (`parent_component_id` FK)
- The relationship chain is explicit and queryable at each level

**Target users:** Backend developers maintaining the CMS data layer; content authors are unaffected (API response shapes do not change).

---

## 3. Current Behavior

All component tables — regardless of nesting depth — share the same schema:

```sql
CREATE TABLE components_{slug}_{path} (
    gorm_id       SERIAL PRIMARY KEY,
    component_id  TEXT,
    document_id   TEXT,         -- always present, even for nested components
    version       TEXT,
    locale        TEXT,
    sort_order    INTEGER DEFAULT 0,
    <field_columns>,
    created_at    TIMESTAMP,
    updated_at    TIMESTAMP
);
```

- All components reference the **document** via `document_id`, even if they are nested inside another component.
- Maximum nesting depth: 2 levels (document -> comp -> child comp).
- Publish/delete/read operations use `document_id` at every level.

---

## 4. Nesting Levels & Table Naming

### 4.1 Level Definitions

| Level | Relationship | Table Pattern | FK Column | Chain Key |
|-------|-------------|---------------|-----------|-----------|
| 1 | Component inside document | `components_{slug}_{comp}` | `document_id` | `(locale, document_id)` |
| 2 | Component inside level-1 component | `components_{slug}_{parent}_{child}` | `parent_component_id` | `(locale, parent_component_id)` |
| 3 | Component inside level-2 component | `components_{slug}_{p1}_{p2}_{grandchild}` | `parent_component_id` | `(locale, parent_component_id)` |
| 4+ | **ERROR on startup** | — | — | — |

### 4.2 Examples

Content type `blog-posts` with this schema:
```json
{
  "slug": "blog-posts",
  "fields": [
    {
      "name": "seo",
      "type": "component",
      "fields": [
        { "name": "title", "type": "text" },
        {
          "name": "openGraph",
          "type": "component",
          "fields": [
            { "name": "image", "type": "media" },
            {
              "name": "metadata",
              "type": "component",
              "fields": [
                { "name": "author", "type": "text" }
              ]
            }
          ]
        }
      ]
    }
  ]
}
```

Creates these tables:

| Table | Level | FK Column |
|-------|-------|-----------|
| `components_blog_posts_seo` | 1 | `document_id` |
| `components_blog_posts_seo_open_graph` | 2 | `parent_component_id` |
| `components_blog_posts_seo_open_graph_metadata` | 3 | `parent_component_id` |

### 4.3 Multi-Locale Chain Key

Every chain traversal query uses the **combined key of (locale + FK ID)** — never the FK ID alone. This prevents cross-locale contamination when the same `component_id` exists across multiple locales.

**Lookup keys per level:**

| Level | Chain Key | Query |
|-------|-----------|-------|
| 1 (top-level) | `(locale, document_id)` | `WHERE document_id = ? AND locale = ? AND version = ?` |
| 2+ (nested) | `(locale, parent_component_id)` | `WHERE parent_component_id = ? AND locale = ? AND version = ?` |

**Why this matters:** A parent component with `component_id=ABC` may have rows in multiple locales (en, vi). When a child component queries by `parent_component_id=ABC`, it must also filter by `locale` to get only the children belonging to the same locale as the parent. Without locale in the chain key, a read/publish/delete for locale "en" would incorrectly affect locale "vi" children.

**Invariant:** At every level of the nesting chain — find, upsert, delete, publish — `locale` is always part of the WHERE clause alongside the FK column. No chain traversal operation ever queries by FK ID alone.

### 4.4 Depth Validation

Schema loader validates at startup. If a 4th level is encountered, the API **refuses to start** with a fatal error:

```
"blog-posts.json": component "deepChild" exceeds maximum nesting depth of 3
```

---

## 5. Database Schema Changes

### 5.1 Top-Level Component Table (Level 1 — Unchanged)

```sql
CREATE TABLE components_{slug}_{component} (
    gorm_id       SERIAL PRIMARY KEY,
    component_id  TEXT NOT NULL,
    document_id   TEXT NOT NULL,            -- FK to documents_{slug}.document_id
    version       TEXT NOT NULL,
    locale        TEXT NOT NULL,
    sort_order    INTEGER DEFAULT 0,
    <field_columns>,
    created_at    TIMESTAMP,
    updated_at    TIMESTAMP
);
```

### 5.2 Nested Component Table (Level 2 and 3 — New)

```sql
CREATE TABLE components_{slug}_{parent_path}_{component} (
    gorm_id              SERIAL PRIMARY KEY,
    component_id         TEXT NOT NULL,
    parent_component_id  TEXT NOT NULL,      -- FK to parent component table's component_id
    version              TEXT NOT NULL,
    locale               TEXT NOT NULL,
    sort_order           INTEGER DEFAULT 0,
    <field_columns>,
    created_at           TIMESTAMP,
    updated_at           TIMESTAMP
);
```

**`document_id` and `parent_component_id` are mutually exclusive per table.** A top-level table has `document_id` and no `parent_component_id`. A nested table has `parent_component_id` and no `document_id`. No component table ever contains both columns.

---

## 6. Entity Changes

### 6.1 Component Entity

```go
type Component struct {
    GormID            uint            `gorm:"column:gorm_id;primaryKey;autoIncrement"`
    ComponentID       string          `gorm:"column:component_id"`
    DocumentID        string          `gorm:"column:document_id"`           // level-1 only; column absent on nested tables
    ParentComponentID string          `gorm:"column:parent_component_id"`   // level-2+ only; column absent on top-level tables
    Version           DocumentVersion `gorm:"column:version;type:varchar(20)"`
    Locale            string          `gorm:"column:locale"`
    SortOrder         int             `gorm:"column:sort_order"`
    Fields            map[string]any  `gorm:"-"`
    CreatedAt         time.Time       `gorm:"column:created_at"`
    UpdatedAt         time.Time       `gorm:"column:updated_at"`
}
```

The entity carries both FK fields for convenience (one struct for both levels), but they are **mutually exclusive** — exactly one is populated, and the other's **column does not exist** in the table:

| Level | Populated field | Column in table | Absent column |
|-------|----------------|-----------------|---------------|
| 1 (top-level) | `DocumentID` | `document_id` | `parent_component_id` does NOT exist |
| 2+ (nested) | `ParentComponentID` | `parent_component_id` | `document_id` does NOT exist |

Since all queries use raw SQL via `compToRow`/`rowToComp` (not GORM auto-mapping), the row functions must **never** include the absent column in the row map — inserting a column that doesn't exist in the table would cause a SQL error.

---

## 7. Repository Interface Changes

```go
type ComponentRepository interface {
    // --- Top-level component operations (by document_id) ---
    FindByDocumentID(ctx context.Context, contentTypeSlug, componentName, documentID, locale string, version entity.DocumentVersion) ([]*entity.Component, error)
    UpsertAll(ctx context.Context, contentTypeSlug, componentName, documentID, locale string, version entity.DocumentVersion, components []*entity.Component) error
    DeleteByDocumentID(ctx context.Context, contentTypeSlug, componentName, documentID, locale string) error
    DeleteAllByContentType(ctx context.Context, contentTypeSlug, componentName string) error

    // --- Nested component operations (by parent_component_id) ---
    FindByParentComponentID(ctx context.Context, contentTypeSlug, componentPath, parentComponentID, locale string, version entity.DocumentVersion) ([]*entity.Component, error)
    UpsertAllByParent(ctx context.Context, contentTypeSlug, componentPath, parentComponentID, locale string, version entity.DocumentVersion, components []*entity.Component) error
    DeleteByParentComponentID(ctx context.Context, contentTypeSlug, componentPath, parentComponentID, locale string) error

    // --- Table management ---
    EnsureCollection(ctx context.Context, contentTypeSlug, componentName string, fields []entity.FieldDefinition, isNested bool) error
    DropCollection(ctx context.Context, contentTypeSlug, componentName string) error
}
```

### 7.1 New Methods

All chain traversal methods use the combined key `(locale, parent_component_id)` — never `parent_component_id` alone.

| Method | Description |
|--------|-------------|
| `FindByParentComponentID` | Queries by combined key `(locale, parent_component_id, version)`. Returns results ordered by `sort_order ASC, gorm_id ASC`. |
| `UpsertAllByParent` | Deletes existing rows matching `(locale, parent_component_id, version)`, then inserts new components with locale and parent_component_id set. |
| `DeleteByParentComponentID` | Deletes all rows matching `(locale, parent_component_id)` — used for cascade cleanup. |

### 7.2 Changed Methods

| Method | Change |
|--------|--------|
| `EnsureCollection` | Added `isNested bool` parameter. When `true`, creates table with `parent_component_id` column instead of `document_id`. |

---

## 8. Schema Loader Changes

### 8.1 Depth Validation

Change the depth threshold from 2 to 3:

```go
// Before
if depth > 2 {
    return fmt.Errorf("%q: component %q exceeds maximum nesting depth of 3", path, f.Name)
}

// After
if depth > 3 {
    return fmt.Errorf("%q: component %q exceeds maximum nesting depth of 3", path, f.Name)
}
```

Since `validateFields` starts at `depth=1` and increments on each component level, this allows:
- depth=1: top-level fields (component here = level 1) -> OK
- depth=2: fields inside level-1 component (component here = level 2) -> OK
- depth=3: fields inside level-2 component (component here = level 3) -> OK
- depth=4: would be level 4 -> ERROR

---

## 9. Sync Engine Changes

### 9.1 `ensureComponentTablesRecursive`

Pass `isNested` based on depth:

```go
func (s *Syncer) ensureComponentTablesRecursive(ctx context.Context, slug, prefix string, fields []entity.FieldDefinition, depth int) error {
    if s.compRepo == nil {
        return nil
    }
    for _, f := range fields {
        if f.Type == "component" {
            path := f.Name
            if prefix != "" {
                path = prefix + "_" + f.Name
            }
            isNested := depth > 0
            if err := s.compRepo.EnsureCollection(ctx, slug, path, f.Fields, isNested); err != nil {
                return err
            }
            if err := s.ensureComponentTablesRecursive(ctx, slug, path, f.Fields, depth+1); err != nil {
                return err
            }
        }
    }
    return nil
}
```

`depth=0` for the initial call (fields directly in the content type) means level-1 components pass `isNested=false`. Level-2+ pass `isNested=true`.

### 9.2 `dropComponentTablesRecursive`

No change needed — drop order (deepest first) and logic remains the same. Table names are path-based and independent of the FK column.

---

## 10. Component Repository Implementation Changes

### 10.1 `createComponentTable`

```go
func (r *componentRepository) createComponentTable(ctx context.Context, table string, fields []entity.FieldDefinition, isNested bool) error {
    var cols []string
    if r.isPostgres() {
        cols = append(cols, "gorm_id SERIAL PRIMARY KEY")
    } else {
        cols = append(cols, "gorm_id INTEGER PRIMARY KEY AUTOINCREMENT")
    }
    cols = append(cols, "component_id TEXT")
    if isNested {
        cols = append(cols, "parent_component_id TEXT")
    } else {
        cols = append(cols, "document_id TEXT")
    }
    cols = append(cols, "version TEXT")
    cols = append(cols, "locale TEXT")
    cols = append(cols, "sort_order INTEGER DEFAULT 0")
    // ... field columns and timestamps
}
```

### 10.2 `addMissingComponentColumns`

When called on an existing table:
- If `isNested`: ensure `parent_component_id` exists; **never** add `document_id`
- If `!isNested`: ensure `document_id` exists; **never** add `parent_component_id`

Each table type only gets its own FK column. The other FK column must remain absent.

### 10.3 `compToRow` / `rowToComp`

`compToRow` writes **exactly one** FK column — the one that exists in the target table. Writing the other would cause a SQL error (column doesn't exist).

```go
func compToRow(comp *entity.Component) map[string]any {
    row := map[string]any{
        "component_id": comp.ComponentID,
        "version":      string(comp.Version),
        "locale":       comp.Locale,
        "sort_order":   comp.SortOrder,
        "created_at":   comp.CreatedAt,
        "updated_at":   comp.UpdatedAt,
    }
    // Exactly one of these is populated; the other's column does not exist in the table.
    if comp.ParentComponentID != "" {
        row["parent_component_id"] = comp.ParentComponentID
        // Do NOT add "document_id" — column absent on nested tables
    } else {
        row["document_id"] = comp.DocumentID
        // Do NOT add "parent_component_id" — column absent on top-level tables
    }
    for k, v := range comp.Fields {
        row[toSnakeCase(k)] = serializeFieldValue(v)
    }
    return row
}
```

`rowToComp` reads whichever FK column exists in the row (the absent column simply won't appear in the row map). Add both `"document_id"` and `"parent_component_id"` to the `systemCols` set so neither leaks into `Fields`.

### 10.4 New Repository Methods

**`FindByParentComponentID`:**
```go
func (r *componentRepository) FindByParentComponentID(ctx context.Context, contentTypeSlug, componentPath, parentComponentID, locale string, version entity.DocumentVersion) ([]*entity.Component, error) {
    var rows []map[string]any
    err := r.table(contentTypeSlug, componentPath).WithContext(ctx).
        Where("parent_component_id = ? AND version = ? AND locale = ?", parentComponentID, version, locale).
        Order("sort_order ASC, gorm_id ASC").
        Find(&rows).Error
    // ... map rows to components
}
```

**`UpsertAllByParent`:**
```go
func (r *componentRepository) UpsertAllByParent(ctx context.Context, contentTypeSlug, componentPath, parentComponentID, locale string, version entity.DocumentVersion, components []*entity.Component) error {
    tbl := r.table(contentTypeSlug, componentPath).WithContext(ctx)
    if err := tbl.Where("parent_component_id = ? AND version = ? AND locale = ?", parentComponentID, version, locale).Delete(map[string]any{}).Error; err != nil {
        return err
    }
    for _, comp := range components {
        comp.ParentComponentID = parentComponentID
        comp.Version = version
        comp.Locale = locale
        row := compToRow(comp)
        if err := r.table(contentTypeSlug, componentPath).WithContext(ctx).Create(row).Error; err != nil {
            return err
        }
    }
    return nil
}
```

**`DeleteByParentComponentID`:**
```go
func (r *componentRepository) DeleteByParentComponentID(ctx context.Context, contentTypeSlug, componentPath, parentComponentID, locale string) error {
    return r.table(contentTypeSlug, componentPath).WithContext(ctx).
        Where("parent_component_id = ? AND locale = ?", parentComponentID, locale).
        Delete(map[string]any{}).Error
}
```

---

## 11. Document Usecase Changes

### 11.1 Save Flow

The save flow must be restructured. Previously, recursion processed nested components first (bottom-up), saving all via `document_id`. Now, the flow is **top-down**: save parent components first (to generate their `component_id`), then save children using `parent_component_id`.

**New flow:**

```
extractAndSaveComponents(slug, doc, fields)
  // Phase 1: Clean up old nested components (bottom-up)
  cleanupNestedComponents(slug, doc.DocumentID, doc.Locale, doc.Version, fields)

  // Phase 2: Save new components (top-down)
  saveTopLevelComponents(slug, doc.DocumentID, doc.Locale, doc.Version, doc.Fields, fields)

saveTopLevelComponents(slug, documentID, locale, version, data, fields)
  for each component field:
    extract data[fieldName]
    delete data[fieldName]
    generate Component with new ComponentID for each item
    compRepo.UpsertAll(slug, fieldName, documentID, locale, version, components)
    for each component:
      saveNestedComponents(slug, fieldName, component.ComponentID, locale, version, component.Fields, field.Fields)

saveNestedComponents(slug, parentPath, parentComponentID, locale, version, data, childFields)
  for each component field in childFields:
    extract data[fieldName]
    delete data[fieldName]
    path = parentPath + "_" + fieldName
    generate Component with new ComponentID for each item
    compRepo.UpsertAllByParent(slug, path, parentComponentID, locale, version, components)
    for each component:
      saveNestedComponents(slug, path, component.ComponentID, locale, version, component.Fields, field.Fields)
```

**`cleanupNestedComponents`** removes orphaned nested rows before new parents are saved:

```
cleanupNestedComponents(slug, documentID, locale, version, fields)
  for each top-level component field:
    oldParents = compRepo.FindByDocumentID(slug, field.Name, documentID, locale, version)
    for each oldParent:
      cleanupChildren(slug, field.Name, oldParent.ComponentID, locale, field.Fields)

cleanupChildren(slug, parentPath, parentComponentID, locale, childFields)
  for each component field in childFields:
    path = parentPath + "_" + field.Name
    oldChildren = compRepo.FindByParentComponentID(slug, path, parentComponentID, locale, version)
    for each oldChild:
      cleanupChildren(slug, path, oldChild.ComponentID, locale, field.Fields)
    compRepo.DeleteByParentComponentID(slug, path, parentComponentID, locale)
```

### 11.2 Read (Merge) Flow

Chain traversal from top to bottom:

```
mergeComponents(slug, doc, fields)
  mergeTopLevel(slug, doc.DocumentID, doc.Locale, doc.Version, doc.Fields, fields)

mergeTopLevel(slug, documentID, locale, version, data, fields)
  for each component field:
    components = compRepo.FindByDocumentID(slug, field.Name, documentID, locale, version)
    for each component:
      mergeNested(slug, field.Name, component.ComponentID, locale, version, component.Fields, field.Fields)
    if repeatable: data[field.Name] = [comp.Fields for each comp]
    else: data[field.Name] = components[0].Fields

mergeNested(slug, parentPath, parentComponentID, locale, version, parentData, childFields)
  for each component field in childFields:
    path = parentPath + "_" + field.Name
    children = compRepo.FindByParentComponentID(slug, path, parentComponentID, locale, version)
    for each child:
      mergeNested(slug, path, child.ComponentID, locale, version, child.Fields, field.Fields)
    if repeatable: parentData[field.Name] = [child.Fields for each child]
    else: parentData[field.Name] = children[0].Fields
```

### 11.3 Publish Flow

Chain traversal — copy draft components to published version at each level:

```
publishComponents(slug, documentID, locale, fields)
  for each top-level component field:
    draftParents = compRepo.FindByDocumentID(slug, field.Name, documentID, locale, draft)
    compRepo.UpsertAll(slug, field.Name, documentID, locale, published, draftParents)
    for each draftParent:
      publishNested(slug, field.Name, draftParent.ComponentID, locale, field.Fields)

publishNested(slug, parentPath, parentComponentID, locale, childFields)
  for each component field in childFields:
    path = parentPath + "_" + field.Name
    draftChildren = compRepo.FindByParentComponentID(slug, path, parentComponentID, locale, draft)
    compRepo.UpsertAllByParent(slug, path, parentComponentID, locale, published, draftChildren)
    for each draftChild:
      publishNested(slug, path, draftChild.ComponentID, locale, field.Fields)
```

This works because both draft and published versions of a parent component share the same `component_id`. The `version` column on child rows distinguishes draft from published.

### 11.4 Delete Flow

Chain traversal bottom-up — delete deepest children first, then parents:

```
deleteComponents(slug, documentID, fields)
  for each locale in supportedLocales:
    for each top-level component field:
      parents = compRepo.FindByDocumentID(slug, field.Name, documentID, locale, draft)
      for each parent:
        deleteChildren(slug, field.Name, parent.ComponentID, locale, field.Fields)
      compRepo.DeleteByDocumentID(slug, field.Name, documentID, locale)

deleteChildren(slug, parentPath, parentComponentID, locale, childFields)
  for each component field in childFields:
    path = parentPath + "_" + field.Name
    children = compRepo.FindByParentComponentID(slug, path, parentComponentID, locale, draft)
    for each child:
      deleteChildren(slug, path, child.ComponentID, locale, field.Fields)
    compRepo.DeleteByParentComponentID(slug, path, parentComponentID, locale)
```

---

## 12. GraphQL Changes

No changes to the GraphQL schema generation. The API response shape for components (nested objects and arrays) remains identical. The storage-level FK change is transparent to consumers.

---

## 13. Frontend Changes

None. The REST and GraphQL response shapes are unchanged. The frontend continues to send/receive components as nested objects/arrays.

---

## 14. File Map (Changes)

All paths relative to project root.

### Backend (`apps/api/`)

| File | Change |
|------|--------|
| `internal/domain/entity/component.go` | Add `ParentComponentID string` field |
| `internal/domain/repository/component_repository.go` | Add `FindByParentComponentID`, `UpsertAllByParent`, `DeleteByParentComponentID`; change `EnsureCollection` signature to add `isNested bool` |
| `internal/domain/repository/mock/component_repository.go` | Add mock implementations for new methods |
| `internal/infrastructure/gormdb/component_repository.go` | Implement new methods; update `createComponentTable`, `addMissingComponentColumns`, `compToRow`, `rowToComp` |
| `internal/usecase/content_type/schema_loader.go` | Change depth threshold from `> 2` to `> 3` |
| `internal/usecase/content_type/sync.go` | Update `ensureComponentTablesRecursive` to pass `isNested` and depth |
| `internal/usecase/document/document_usecase.go` | Restructure `extractAndSaveComponents`, `mergeComponents`, `publishComponentsRecursive`, `deleteComponentsRecursive` for chain traversal |

---

## 15. Testing

### Schema Loader (`schema_loader_test.go`)

- 3 levels of component nesting -> parses correctly, no error
- 4 levels of component nesting -> fatal error with clear message
- Existing tests for 2-level nesting continue to pass

### Sync (`sync_test.go`)

- Content type with 3-level components -> creates all 3 component tables
- Level-1 table has `document_id` column
- Level-2 and level-3 tables have `parent_component_id` column (no `document_id`)
- Drop content type -> drops all 3 component tables

### Component Repository (`component_repository_test.go`)

- `EnsureCollection(isNested=false)` -> creates table with `document_id` column, no `parent_component_id` column
- `EnsureCollection(isNested=true)` -> creates table with `parent_component_id` column, no `document_id` column
- `FindByParentComponentID` -> returns components matching parent_component_id, ordered by sort_order
- `UpsertAllByParent` -> deletes old, inserts new with parent_component_id
- `DeleteByParentComponentID` -> removes all rows for given parent
- `addMissingComponentColumns` for nested table adds `parent_component_id` if missing

### Document Usecase (`document_usecase_test.go`)

**Save:**
- Save document with level-1 component -> saved with `document_id`
- Save document with level-2 nested component -> level-1 saved with `document_id`, level-2 saved with `parent_component_id` = level-1's `component_id`
- Save document with level-3 nested component -> chain of parent references is correct
- Re-save document -> old nested components cleaned up, no orphans

**Read:**
- Read document with 3-level components -> chain traversal correctly merges all levels
- Repeatable nested components -> returned as arrays at each level

**Publish:**
- Publish copies draft components at all levels to published version
- Published children reference same `parent_component_id` as draft children (component_id is preserved)

**Delete:**
- Delete removes components at all levels (bottom-up chain traversal)
- No orphaned rows remain after delete

**Multi-locale chain key isolation:**
- Save same document in locale "en" and "vi" -> each locale gets independent component_ids and parent_component_id chains
- Read locale "en" -> returns only "en" components at every nesting level; "vi" data never leaks
- Publish locale "en" -> only "en" components copied to published; "vi" draft components untouched
- Delete locale "en" -> only "en" component chain removed; "vi" component chain intact

---

## 16. Migration

### Existing Data

Existing nested component tables (level 2, currently stored with `document_id`) need migration:

1. **Schema migration**: `EnsureCollection(isNested=true)` adds `parent_component_id` column to existing level-2 tables via `addMissingComponentColumns`.

2. **Data migration**: A one-time startup migration step:
   - For each content type with nested components:
     - For each level-2 component table:
       - For each row: look up the parent component's `component_id` from the level-1 table (matching `document_id`, `version`, `locale`), and set `parent_component_id`.
   - After migration, `document_id` column remains on nested tables but is unused.

3. **Trigger**: The migration runs as part of the sync engine on first startup after deploy. A flag or check ensures it runs only once (e.g., `parent_component_id IS NULL` rows exist).

### New Installations

No migration needed. Tables are created with the correct schema from the start.

---

## 17. Boundaries

| Rule | Detail |
|------|--------|
| **Always** | Chain key is `(locale, FK_ID)` — every find/upsert/delete/publish query combines locale with the FK column |
| **Always** | Top-level component tables use `document_id` FK; chain key = `(locale, document_id)` |
| **Always** | Nested component tables (level 2+) use `parent_component_id` FK; chain key = `(locale, parent_component_id)` |
| **Always** | Maximum 3 levels of component nesting; fatal error on startup if exceeded |
| **Always** | Publish preserves `component_id` so child references remain valid across draft/published |
| **Always** | Clean up old nested components before saving new parent components (no orphans) |
| **Always** | Delete traverses bottom-up: deepest children first, then parents |
| **Never** | Query by FK ID alone without locale — always use the combined chain key |
| **Never** | Have both `document_id` and `parent_component_id` columns in the same table — they are mutually exclusive per table |
| **Never** | Write `document_id` to a nested component row, or `parent_component_id` to a top-level component row |
| **Never** | Create nested component collections in MongoDB — components stay nested in BSON `data` |
| **Never** | Allow more than 3 levels of component nesting |
| **Ask first** | Increasing the max nesting depth beyond 3 |
| **Ask first** | Adding indexes on `parent_component_id` for performance |
