# RULES — Component

**Scope:** Component entity, repeatable/non-repeatable components, nested component tables (PostgreSQL), MongoDB components, component CRUD operations.

---

## 1. Component Rules

### 1.1 Component Entity
```go
type Component struct {
    GormID            uint
    ComponentID       string
    DocumentID        string           // level-1 only
    ParentComponentID string           // level-2+ only
    Version           DocumentVersion
    Locale            string
    SortOrder         int
    Fields            map[string]any   // gorm:"-"
}
```

### 1.2 Non-Repeatable Components
- Single object: `{ "banner": { "title": "...", "background": "..." } }`
- Exactly one row per `(document_id, version, locale)` tuple
- `sort_order = 0` always
- Validation: reject if client sends array for non-repeatable field

### 1.3 Repeatable Components
- Array: `{ "skills": [ { "name": "..." }, { "name": "..." } ] }`
- Zero or more rows per `(document_id, version, locale)` tuple
- `sort_order` from array index (0, 1, 2, ...)
- On save: delete-all then insert (UpsertAll pattern)
- Validation: reject if client sends object for repeatable field

### 1.4 Nested Component Tables (PostgreSQL Only)
- Level 1 (document child): `components_{slug}_{comp}` — FK = `document_id`
- Level 2 (component child): `components_{slug}_{parent}_{child}` — FK = `parent_component_id`
- Level 3 (grandchild): `components_{slug}_{p1}_{p2}_{grandchild}` — FK = `parent_component_id`
- Level 4+ → **FATAL ERROR** on startup
- `document_id` and `parent_component_id` are **mutually exclusive** per table
- **NEVER** have both FK columns in the same table
- **NEVER** write `document_id` to a nested table, or `parent_component_id` to a top-level table

### 1.5 Chain Key Invariant (Multi-Locale)
- Every chain traversal query uses `(locale, FK_ID)` — **NEVER** FK ID alone
- This prevents cross-locale contamination
- Applies to: find, upsert, delete, publish operations at every level

### 1.6 MongoDB Components
- Components remain nested in BSON `data` field — no separate collections
- Non-repeatable: object; Repeatable: array
- **NEVER** create component collections in MongoDB

---

## 2. GORM Infrastructure Rules (Content-Specific)

### 2.1 Dynamic Table Naming
- Document tables: `documents_<slug_underscored>` (hyphens → underscores)
- Component tables: `components_<slug_underscored>_<component_path_underscored>`
- All queries use `r.db.Table("documents_" + sanitize(slug))`
- Document entity removed from `AutoMigrate()` — tables created by `EnsureCollection`

### 2.2 Per-Field Column Mapping
| Content Type | SQL Type |
|---|---|
| `text` / `richtext` | TEXT |
| `media` | VARCHAR (stores documentId FK) |
| `number` | REAL |
| `boolean` | BOOLEAN |
| `json` | TEXT |

### 2.3 Document Table Schema
```sql
CREATE TABLE documents_<slug> (
    gorm_id BIGSERIAL PRIMARY KEY,
    document_id UUID NOT NULL,
    version VARCHAR(20) NOT NULL,
    locale VARCHAR(10) NOT NULL,
    <per_field_columns>,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    published_at TIMESTAMPTZ,
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    published_by VARCHAR(255),
    UNIQUE(document_id, version, locale)
);
```

### 2.4 Component Table Schema (Top-Level)
```sql
-- Has document_id, NO parent_component_id
gorm_id, component_id, document_id, version, locale, sort_order, <fields>, created_at, updated_at
```

### 2.5 Component Table Schema (Nested)
```sql
-- Has parent_component_id, NO document_id
gorm_id, component_id, parent_component_id, version, locale, sort_order, <fields>, created_at, updated_at
```

### 2.6 `compToRow` / `rowToComp`
- Write **exactly one** FK column based on which is populated
- Writing the absent column causes SQL error
- Both `document_id` and `parent_component_id` in `systemCols` set

### 2.7 Save/Read/Publish/Delete Flow (PostgreSQL)
- **Save**: Top-down. Save parents first (generate `component_id`), then children with `parent_component_id`. Cleanup old nested rows before saving new.
- **Read**: Top-down chain traversal. Load parent components → for each, load children → merge into `Fields`.
- **Publish**: Chain traversal. Copy draft components to published at each level. `component_id` preserved across versions.
- **Delete**: Bottom-up. Delete deepest children first, then parents. Traverse all locales.

---

## 3. Testing Rules (Content-Specific)

### 3.1 Schema Sync Tests
- New file → creates ContentType + collection
- Changed file → updates schema
- Removed field → drops from schema, data untouched
- Deleted file → cascade-deletes type + entries + collection
- Sync does NOT overwrite user-configured ListFields
- 3 levels of nesting → OK; 4 levels → fatal error

### 3.2 Document Usecase Tests
- Save: upserts draft, never touches published
- Publish: copies draft → published, sets timestamps
- Unpublish: deletes published
- Status computation: draft / modified / published
- Duplicate: new documentId, same data, draft only
- Repeatable/non-repeatable validation: correct shape enforced
- Component chain: parent references correct at all levels
- Multi-locale isolation: no cross-locale contamination

### 3.3 Handler Tests
- Slug validation → 400
- DocumentID validation → 400
- Permission checks → 403
- Not found → 404
- CRUD operations → correct status codes

---

## 4. Boundaries Summary

| Rule | Detail |
|---|---|
| **Always** | Validate data shape at usecase (object vs array based on repeatable) |
| **Always** | Preserve `sort_order` through save→publish→read cycle |
| **Always** | Chain key is `(locale, FK_ID)` for all component operations |
| **Always** | Max 3 levels of component nesting; fatal error if exceeded |
| **Always** | Clean up old nested components before saving new parents |
| **Always** | Delete components bottom-up: deepest children first |
| **Always** | Non-destructive `EnsureCollection` — never DROP+CREATE |
| **Never** | Create component collections in MongoDB |
| **Never** | Have both `document_id` and `parent_component_id` in same table |
| **Never** | Query by FK ID alone without locale |
| **Never** | Allow more than 3 levels of component nesting |
| **Ask first** | Increasing max nesting depth beyond 3 |
| **Ask first** | Adding indexes on `parent_component_id` |
