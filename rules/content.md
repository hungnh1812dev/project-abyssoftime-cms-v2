# RULES — content Module

**Scope:** Content types, documents, draft/publish workflow, schema sync, pagination, field projection, GraphQL schema generation, component tables.
**Spec:** [specs/content.md](../specs/content.md), [specs/repeatable-components.md](../specs/repeatable-components.md), [specs/nested-component-tables.md](../specs/nested-component-tables.md), [specs/duplicate-document.md](../specs/duplicate-document.md), [specs/configurable-list-columns.md](../specs/configurable-list-columns.md)

---

## 1. Content Type Rules

### 1.1 ContentType Entity
- `Kind`: `"single"` (at most one entry) or `"collection"` (many entries)
- `Slug`: validated format `^[a-z0-9]+(?:-[a-z0-9]+)*$`, 1-63 chars
- `Fields`: array of `FieldDefinition` (name, type, repeatable, nested fields)
- `ListFields`: managed via UI only — **NEVER** defined in JSON schema files
- Slug characters: only `[a-z0-9-]` allowed

### 1.2 FieldDefinition
```go
type FieldDefinition struct {
    Name       string            `json:"name"`
    Type       string            `json:"type"`
    Ext        []string          `json:"ext,omitempty"`
    Width      string            `json:"width,omitempty"`
    Repeatable bool              `json:"repeatable,omitempty"`
    Fields     []FieldDefinition `json:"fields,omitempty"`
}
```
- `Width`: UI hint for form column span (`"100%"`, `"50%"`, `"1/3"`). Defaults to `"100%"` when omitted. No effect on storage.
- `type` values: `text`, `richtext`, `number`, `boolean`, `media`, `json`, `component`
- `repeatable` only valid on `type: "component"` — ignored on other types
- Maximum nesting depth: 3 levels — fatal error on startup if exceeded

### 1.3 Schema-as-Code
- JSON files in `apps/api/content-types/*.json` are source of truth
- **NEVER** create/edit/delete ContentType structure via API or UI
- Sync is one-directional: JSON → DB
- Sync runs on every API startup
- `listFields` NOT part of JSON schemas — permanently removed

---

## 2. Document Rules

### 2.1 Document Entity
- `DocumentID`: UUID v4 — the primary domain identifier
- `Version`: `"draft"` or `"published"` — two separate records per entry
- `Fields`: `map[string]any` — content data (tagged `gorm:"-"` for GORM)
- `GormID`: auto-increment uint for display ordering
- `Locale`: defaults to `"en"` (or default locale)
- Higher layers only use `documentId` + content-type `slug` — **NEVER** MongoDB `_id`

### 2.2 Draft/Publish Workflow
- Every entry = two separate records (draft + published) sharing the same `documentId`
- **Save** → upsert draft record only. **NEVER** touch published.
- **Publish** → copy `draft.data` to published record, set `publishedAt = now()`
- **Unpublish** → delete published record
- **Status** computed, never stored:
  - `draft`: no published record exists
  - `modified`: `draft.updatedAt > published.updatedAt`
  - `published`: timestamps match
- Documents only created on explicit Save — no auto-creation

### 2.3 Single-Type Rules
- At most one entry per content type
- No auto-created singleton — first Save creates it
- UI: edit + Save + Publish only. No create/delete.
- **NEVER** expose DELETE for single-type documents
- **NEVER** include `documentId` in single-type URLs
- GET returns 404 when no document exists (FE shows empty form)
- PUT creates on first save, updates on subsequent saves

### 2.4 Collection-Type Rules
- Zero or more entries, each with own `documentId`
- List + create/edit/delete, each with independent draft/published pair
- Pagination: `start` (offset), `size` (default 20, max 100)
- **NEVER** allow `size` above 100
- Support `orderBy` and `sortDir` query params

---

## 3. Schema Sync Rules

### 3.1 Sync Engine (`usecase/content_type/sync.go`)
- New file → create ContentType + per-content-type document collection/table
- Changed file → update ContentType schema in place
- Field removed → drop from schema, leave stored data untouched
- File deleted → delete ContentType, cascade-delete all entries, drop collection/table
- **NEVER** let sync write back to JSON definition files
- **NEVER** overwrite user-configured `ListFields` — sync only seeds when empty

### 3.2 Schema Loader (`usecase/content_type/schema_loader.go`)
- Reads all `*.json` files from `CONTENT_TYPES_DIR`
- Validates field definitions (name, type, nesting depth)
- `listFields` validation removed — no longer in JSON schemas
- `validateFields` starts at depth=1, checks component nesting ≤ 3

### 3.3 EnsureCollection (Content-Specific)
- Document tables: `documents_<slug_underscored>` (hyphens → underscores)
- Component tables: `components_<slug_underscored>_<component_path>`
- Non-destructive: create if missing, add columns if existing
- **NEVER** drop and recreate tables

---

## 4. Component Rules

### 4.1 Component Entity
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

### 4.2 Non-Repeatable Components
- Single object: `{ "banner": { "title": "...", "background": "..." } }`
- Exactly one row per `(document_id, version, locale)` tuple
- `sort_order = 0` always
- Validation: reject if client sends array for non-repeatable field

### 4.3 Repeatable Components
- Array: `{ "skills": [ { "name": "..." }, { "name": "..." } ] }`
- Zero or more rows per `(document_id, version, locale)` tuple
- `sort_order` from array index (0, 1, 2, ...)
- On save: delete-all then insert (UpsertAll pattern)
- Validation: reject if client sends object for repeatable field

### 4.4 Nested Component Tables (PostgreSQL Only)
- Level 1 (document child): `components_{slug}_{comp}` — FK = `document_id`
- Level 2 (component child): `components_{slug}_{parent}_{child}` — FK = `parent_component_id`
- Level 3 (grandchild): `components_{slug}_{p1}_{p2}_{grandchild}` — FK = `parent_component_id`
- Level 4+ → **FATAL ERROR** on startup
- `document_id` and `parent_component_id` are **mutually exclusive** per table
- **NEVER** have both FK columns in the same table
- **NEVER** write `document_id` to a nested table, or `parent_component_id` to a top-level table

### 4.5 Chain Key Invariant (Multi-Locale)
- Every chain traversal query uses `(locale, FK_ID)` — **NEVER** FK ID alone
- This prevents cross-locale contamination
- Applies to: find, upsert, delete, publish operations at every level

### 4.6 MongoDB Components
- Components remain nested in BSON `data` field — no separate collections
- Non-repeatable: object; Repeatable: array
- **NEVER** create component collections in MongoDB

---

## 5. API Contract Rules

### 5.1 Input Validation (Handler-Level)
- Slug: `^[a-z0-9]+(?:-[a-z0-9]+)*$` — applied on every request. 400 on invalid.
- DocumentID: UUID v4 format — applied on every request. 400 on invalid.
- Return 400 (not 500) for invalid slug or documentID format

### 5.2 REST Routes — Content Types
| Method | Route | Permission |
|---|---|---|
| `GET` | `/api/content-types` | `content_types:read` |
| `GET` | `/api/content-types/:identifier` | `content_types:read` |
| `PATCH` | `/api/content-types/:slug/list-fields` | `content_types:read` |

### 5.3 REST Routes — Single-Type Documents
| Method | Route | Permission |
|---|---|---|
| `GET` | `/api/document-manager/single-type/:slug` | `content:read` |
| `PUT` | `/api/document-manager/single-type/:slug` | `content:update` |
| `POST` | `/api/document-manager/single-type/:slug/publish` | `content:publish` |
| `POST` | `/api/document-manager/single-type/:slug/unpublish` | `content:unpublish` |

### 5.4 REST Routes — Collection-Type Documents
| Method | Route | Permission |
|---|---|---|
| `GET` | `/api/document-manager/collection-type/:slug` | `content:read` |
| `GET` | `/api/document-manager/collection-type/:slug/:documentId` | `content:read` |
| `POST` | `/api/document-manager/collection-type/:slug` | `content:create` |
| `PUT` | `/api/document-manager/collection-type/:slug/:documentId` | `content:update` |
| `DELETE` | `/api/document-manager/collection-type/:slug/:documentId` | `content:delete` |
| `POST` | `.../:documentId/publish` | `content:publish` |
| `POST` | `.../:documentId/unpublish` | `content:unpublish` |
| `POST` | `.../:documentId/duplicate` | `content:create` |

### 5.5 REST Response Shapes
- Document response: `{ "data": { "documentId", "status", "locale", ...systemFields, ...contentFields } }`
- Paginated list: `{ "items": [...], "total", "start", "size", "listFields" }`
- List items contain **only** selected fields (from `listFields`) — **NEVER** return full data in paginated lists
- `updatedByName` resolved server-side from User's `DisplayName`
- `status` and `contentTypeId` excluded from public API responses

### 5.6 Public Read Route
- `GET /api/public/document-manager/:slug/:documentId` — no auth required
- Returns published record **ONLY** — 404 if not published
- **NEVER** return draft data through public API

---

## 6. GraphQL Rules

### 6.1 Build-Time Schema (gqlgen)
- Schema defined in `.graphql` files generated at build time by `cmd/gqlcodegen`
- Library: `github.com/99designs/gqlgen` (replaces `graphql-go/graphql`)
- **Must run `make graphql-generate` after changing `content-types/*.json`**
- Generated files (`.gitignored`): `graphql/schema/`, `graphql/generated/`, `graphql/model/models_gen.go`, `graphql/resolver/content_gen.go`
- Hand-written files: `graphql/resolver/resolver.go`, `document_helpers.go`, `media.go`, `filter.go`, `content_types.go`, `graphql/handler.go`

### 6.2 Schema Shape
- Collection type → `Query.<slug>(Id: ID!, locale: String, status: String)` + `Query.<slugList>(filters, orderBy, start, size, locale, status)`
- Single type → `Query.<slug>(locale: String, status: String)`
- Queries default to published; `status: "draft"` opt-in for authenticated users
- All mutations have `@auth` directive (handler-level auth enforced)
- Base filter input types: `IDFilter`, `StringFilter`, `NumberFilter`, `BooleanFilter`, `TimeFilter`
- Per-type `<Type>Filter` includes: `documentId` (IDFilter), `createdAt`/`updatedAt`/`publishedAt` (TimeFilter), content fields, `and`/`or`/`not` combinators
- `filters` argument is an array — items are implicitly ANDed
- Supported operators: `eq`, `ne`, `in`, `notIn`
- `orderBy` uses typed `<Type>OrderBy` struct (first non-nil field determines sort)

### 6.3 Field Type Mapping
| Content Type | GraphQL Output Type | GraphQL Input Type |
|---|---|---|
| `text` | `String` | `String` |
| `richtext` | `String` | `String` |
| `number` | `Float` | `Float` |
| `boolean` | `Boolean` | `Boolean` |
| `media` | `MediaAsset` object | `String` (documentId) |
| `json` | `JSON` scalar | `JSON` |
| `component` (non-repeatable) | Nested object type | *(excluded from input)* |
| `component` (repeatable) | `[NestedType!]` | *(excluded from input)* |

### 6.4 Naming Conventions
- Type: PascalCase of slug (`blog-posts` → `BlogPosts`)
- Input: `<Type>Input`
- Filter: `<Type>Filter`
- Base filters: `IDFilter`, `StringFilter`, `NumberFilter`, `BooleanFilter`, `TimeFilter`
- OrderBy: `<Type>OrderBy`
- Component: `<ContentType><ComponentName>` (e.g., `CvPageSkills`)
- Query single: camelCase (`cvPage`)
- Query list: camelCase + `List` (`cvPageList`)

### 6.5 Resolvers
- All delegate to generic helpers in `document_helpers.go` — **NO** business logic in resolvers
- `Resolver` struct takes `DocumentUseCase`, `ContentTypeUseCase`, `MediaAssetRepository`
- Generated resolvers in `content_gen.go` are thin wrappers calling generic helpers
- Media fields resolved into full `MediaAsset` objects recursively via `media.go`
- Filter conversion: gqlgen typed structs → `entity.FilterNode` via reflection (`filter.go`)
- **NEVER** duplicate filtering on repeatable component sub-fields (defer to future)
- **NEVER** add business logic to generated or hand-written resolvers

### 6.6 Codegen Pipeline (`make graphql-generate`)
1. `gqlcodegen --phase=schema` → generates `.graphql` files + updates `gqlgen.yml` models
2. `gqlgen generate` → generates `generated.go` + `models_gen.go`
3. Remove gqlgen resolver stubs (`*.resolvers.go`)
4. `gqlcodegen --phase=resolvers` → generates `content_gen.go`
- **NEVER** edit generated files — they are overwritten on every run
- **NEVER** commit generated files to git

---

## 7. Duplicate Document Rules

### 7.1 Behavior
- Creates full copy as new draft with fresh `documentId`
- Media references shared (same documentIds) — no file re-upload
- All component data fully copied; new `componentId` per component
- Active locale only — does not duplicate other locales
- Always draft-only — **NEVER** create published version for duplicate
- Navigate to new document's edit page after duplication

### 7.2 Fields NOT Copied
- `documentId` (new UUID generated)
- `gormId` (auto-incremented)
- `createdAt`, `updatedAt` (set to now)
- `createdBy`, `updatedBy` (set to current user)
- `publishedAt`, `publishedBy` (not set)
- Published version record (not created)

---

## 8. Configurable List Columns Rules

### 8.1 Source of Truth
- `ListFields` stored in content_types DB table — UI-managed
- **NEVER** defined in JSON schema files (permanently removed)
- Startup sync only seeds when DB value is empty/nil — never overwrites

### 8.2 Column Layout
```
| Id | [selected content fields] | [selected system fields] | Status | Actions |
```
- Locked columns: Id (first), Status (before Actions), Actions (last) — not in popup
- Content fields: from `Fields` definition, excluding `component` type
- System fields: CreatedAt, UpdatedAt, UpdatedBy
- Default (empty listFields): first 3 content fields + all system fields

### 8.3 Validation
- Each entry must be a known content field name OR a known system field
- Component-type fields rejected
- Empty array = revert to defaults

---

## 9. GORM Infrastructure Rules (Content-Specific)

### 9.1 Dynamic Table Naming
- Document tables: `documents_<slug_underscored>` (hyphens → underscores)
- Component tables: `components_<slug_underscored>_<component_path_underscored>`
- All queries use `r.db.Table("documents_" + sanitize(slug))`
- Document entity removed from `AutoMigrate()` — tables created by `EnsureCollection`

### 9.2 Per-Field Column Mapping
| Content Type | SQL Type |
|---|---|
| `text` / `richtext` | TEXT |
| `media` | VARCHAR (stores documentId FK) |
| `number` | REAL |
| `boolean` | BOOLEAN |
| `json` | TEXT |

### 9.3 Document Table Schema
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

### 9.4 Component Table Schema (Top-Level)
```sql
-- Has document_id, NO parent_component_id
gorm_id, component_id, document_id, version, locale, sort_order, <fields>, created_at, updated_at
```

### 9.5 Component Table Schema (Nested)
```sql
-- Has parent_component_id, NO document_id
gorm_id, component_id, parent_component_id, version, locale, sort_order, <fields>, created_at, updated_at
```

### 9.6 `compToRow` / `rowToComp`
- Write **exactly one** FK column based on which is populated
- Writing the absent column causes SQL error
- Both `document_id` and `parent_component_id` in `systemCols` set

### 9.7 Save/Read/Publish/Delete Flow (PostgreSQL)
- **Save**: Top-down. Save parents first (generate `component_id`), then children with `parent_component_id`. Cleanup old nested rows before saving new.
- **Read**: Top-down chain traversal. Load parent components → for each, load children → merge into `Fields`.
- **Publish**: Chain traversal. Copy draft components to published at each level. `component_id` preserved across versions.
- **Delete**: Bottom-up. Delete deepest children first, then parents. Traverse all locales.

---

## 10. Testing Rules (Content-Specific)

### 10.1 Schema Sync Tests
- New file → creates ContentType + collection
- Changed file → updates schema
- Removed field → drops from schema, data untouched
- Deleted file → cascade-deletes type + entries + collection
- Sync does NOT overwrite user-configured ListFields
- 3 levels of nesting → OK; 4 levels → fatal error

### 10.2 Document Usecase Tests
- Save: upserts draft, never touches published
- Publish: copies draft → published, sets timestamps
- Unpublish: deletes published
- Status computation: draft / modified / published
- Duplicate: new documentId, same data, draft only
- Repeatable/non-repeatable validation: correct shape enforced
- Component chain: parent references correct at all levels
- Multi-locale isolation: no cross-locale contamination

### 10.3 Handler Tests
- Slug validation → 400
- DocumentID validation → 400
- Permission checks → 403
- Not found → 404
- CRUD operations → correct status codes

---

## 11. Boundaries Summary

| Rule | Detail |
|---|---|
| **Always** | Validate slug format at both usecase and handler levels |
| **Always** | Validate documentID as UUID v4 before passing to usecase |
| **Always** | Return 404 for single-type GET when no document exists |
| **Always** | Include computed `status` in every document response |
| **Always** | Project `data` in collection lists — never full data |
| **Always** | Batch-fetch published records for status computation (no N+1) |
| **Always** | Route GraphQL through the same usecase — no logic in resolvers |
| **Always** | Default `repeatable` to `false` when omitted |
| **Always** | Validate data shape at usecase (object vs array based on repeatable) |
| **Always** | Sanitize field values before save: coerce `""` to nil for number, boolean, media fields |
| **Always** | Preserve `sort_order` through save→publish→read cycle |
| **Always** | Chain key is `(locale, FK_ID)` for all component operations |
| **Always** | Max 3 levels of component nesting; fatal error if exceeded |
| **Always** | Clean up old nested components before saving new parents |
| **Always** | Delete components bottom-up: deepest children first |
| **Always** | Non-destructive `EnsureCollection` — never DROP+CREATE |
| **Never** | Expose DELETE for single-type documents |
| **Never** | Allow `size` above 100 on collection list |
| **Never** | Return draft data through public API |
| **Never** | Include `documentId` in single-type URLs |
| **Never** | Let sync write back to JSON definition files |
| **Never** | Add API/UI to create/edit/delete ContentType structure |
| **Never** | Create component collections in MongoDB |
| **Never** | Have both `document_id` and `parent_component_id` in same table |
| **Never** | Query by FK ID alone without locale |
| **Never** | Allow more than 3 levels of component nesting |
| **Never** | Define `listFields` in JSON schema files |
| **Never** | Let sync overwrite user-configured `listFields` |
| **Never** | Filter on repeatable component sub-fields in GraphQL |
| **Never** | Allow filtering on `component`, `media`, or `json` fields in GraphQL |
| **Never** | Build SQL with string concatenation of user-provided filter values |
| **Always** | Validate filter field names against known content fields + system fields |
| **Always** | Use parameterized queries for all filter conditions |
| **Always** | Map filter field names to correct column/key names per DB adapter |
| **Ask first** | Changing default `size` or max from 100 |
| **Ask first** | Increasing max nesting depth beyond 3 |
| **Ask first** | Adding indexes on `parent_component_id` |
| **Ask first** | Adding new filter operators beyond `eq`, `ne`, `in`, `notIn` |
