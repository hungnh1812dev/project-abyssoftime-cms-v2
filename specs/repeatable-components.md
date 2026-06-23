# SPEC — Repeatable Components

## 1. Overview

Extends the component field type to support two modes: **non-repeatable** (single item, current behavior) and **repeatable** (ordered array of items). A `"repeatable": true` flag in the content-type JSON schema controls the mode. The backend stores, validates, and serves component data in the correct shape; the frontend renders add/remove/reorder controls for repeatable components.

---

## 2. Objective

Allow content-type authors to declare whether a component field holds exactly one item or an ordered list of items. This enables use cases like:
- A CV page with multiple `technicalSkills` entries (repeatable)
- A blog post with a single `banner` component (non-repeatable, current default)

**Target users:** Content authors using the admin panel; developers defining content-type schemas.

---

## 3. JSON Schema Declaration

A component field gains an optional `repeatable` boolean. Default is `false` (backward-compatible).

**Non-repeatable (default):**
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

**Repeatable:**
```json
{
  "name": "technicalSkills",
  "type": "component",
  "repeatable": true,
  "fields": [
    { "name": "category", "type": "text" },
    { "name": "skills", "type": "text" }
  ]
}
```

`"repeatable"` is only valid on `type: "component"` fields. The schema loader ignores it on other field types.

---

## 4. Entity Changes

### Go — `FieldDefinition`

```go
type FieldDefinition struct {
    Name       string            `json:"name"             bson:"name"`
    Type       string            `json:"type"             bson:"type"`
    Ext        []string          `json:"ext,omitempty"    bson:"ext,omitempty"`
    Repeatable bool              `json:"repeatable,omitempty" bson:"repeatable,omitempty"`
    Fields     []FieldDefinition `json:"fields,omitempty" bson:"fields,omitempty"`
}
```

Add `Repeatable bool` with `omitempty` tags. Defaults to `false` (zero value).

### Go — `Component`

```go
type Component struct {
    GormID      uint            `gorm:"column:gorm_id;primaryKey;autoIncrement"`
    ComponentID string          `gorm:"column:component_id"`
    DocumentID  string          `gorm:"column:document_id"`
    Version     DocumentVersion `gorm:"column:version;type:varchar(20)"`
    Locale      string          `gorm:"column:locale"`
    SortOrder   int             `gorm:"column:sort_order"`
    Fields      map[string]any  `gorm:"-"`
    CreatedAt   time.Time       `gorm:"column:created_at"`
    UpdatedAt   time.Time       `gorm:"column:updated_at"`
}
```

Add `SortOrder int` column. For non-repeatable components, `sort_order` is always `0`. For repeatable, it carries the client-supplied order (0-indexed).

### TypeScript — `FieldDefinition`

```typescript
export interface FieldDefinition {
  name: string;
  type: string;
  ext?: string[];
  repeatable?: boolean;
  fields?: FieldDefinition[];
}
```

Add optional `repeatable` boolean.

---

## 5. Database Changes

### Component Table Schema

The `sort_order` column is added to every component table (both new and existing):

```sql
CREATE TABLE components_<slug>_<component> (
    gorm_id       SERIAL PRIMARY KEY,       -- or INTEGER AUTOINCREMENT for SQLite
    component_id  TEXT,
    document_id   TEXT,
    version       TEXT,
    locale        TEXT,
    sort_order    INTEGER DEFAULT 0,          -- NEW
    <field_columns...>,
    created_at    TIMESTAMP,
    updated_at    TIMESTAMP
);
```

**`EnsureCollection` change:** When a component table already exists, `addMissingComponentColumns` adds `sort_order INTEGER DEFAULT 0` if the column is missing (treated as a system column, not derived from `FieldDefinition`).

### Non-repeatable Storage

- Exactly **one** row per `(document_id, version, locale)` tuple.
- `sort_order = 0` always.
- The `UNIQUE(document_id, version, locale)` constraint is **not** added — the constraint would prevent repeatable components from storing multiple rows. Instead, uniqueness for non-repeatable components is enforced at the **usecase layer**.

### Repeatable Storage

- **Zero or more** rows per `(document_id, version, locale)` tuple.
- `sort_order` is set from the array index (0, 1, 2, ...).
- Ordering query: `ORDER BY sort_order ASC`.
- On save: delete-all then insert (existing `UpsertAll` pattern), preserving `sort_order` from client data.

### MongoDB (No Change)

MongoDB stores components nested in the document's BSON `data` field:
- Non-repeatable: `{ "banner": { "title": "...", "background": "..." } }` (object)
- Repeatable: `{ "technicalSkills": [ { "category": "...", "skills": "..." }, ... ] }` (array)

No schema change needed — BSON handles both shapes natively.

---

## 6. Schema Loader Changes

### Validation (`validateFields`)

- `"repeatable"` is only meaningful on `type: "component"` fields. If set on a non-component field, the loader **ignores it** (no error — forward-compatible).
- No new validation rules for the `repeatable` flag itself (it is a simple boolean).

### `fieldsEqual` in sync.go

Add `Repeatable` comparison:

```go
if a[i].Repeatable != b[i].Repeatable {
    return false
}
```

This ensures toggling `repeatable` triggers a content-type update on sync.

---

## 7. API Data Shape

### REST — Document Data

**Non-repeatable component** (unchanged):
```json
{
  "data": {
    "documentId": "...",
    "banner": {
      "title": "Hello",
      "background": "media-doc-id"
    }
  }
}
```

**Repeatable component:**
```json
{
  "data": {
    "documentId": "...",
    "technicalSkills": [
      { "category": "Frontend", "skills": "React, TypeScript" },
      { "category": "Backend", "skills": "Go, PostgreSQL" }
    ]
  }
}
```

### Save (PUT/POST) Input

- Non-repeatable: client sends `{ "banner": { ... } }` — a single object.
- Repeatable: client sends `{ "technicalSkills": [ { ... }, { ... } ] }` — an array of objects.

### Usecase Validation

The document usecase uses the content-type's `FieldDefinition` to determine expected shape:
- If `repeatable == false` and the field value is an array → **400 error**: "field X expects a single component, not an array"
- If `repeatable == true` and the field value is a plain object → **400 error**: "field X expects an array of components"

---

## 8. GORM Document Repository Changes

### Save Flow (PostgreSQL)

Current flow extracts component fields from `doc.Fields` and saves them via `ComponentRepository.UpsertAll`. Changes:

1. For each component `FieldDefinition` in the content type:
   - If `repeatable == false`:
     - Extract `doc.Fields[fieldName]` as `map[string]any` (single object).
     - Wrap in a single-element `[]*Component` with `SortOrder = 0`.
     - Call `UpsertAll(...)`.
   - If `repeatable == true`:
     - Extract `doc.Fields[fieldName]` as `[]any` (array of objects).
     - Map each element to `*Component` with `SortOrder = index`.
     - Call `UpsertAll(...)`.
2. Remove component fields from `doc.Fields` before writing the document row (existing behavior).

### Read Flow (PostgreSQL)

1. Load document from `documents_<slug>`.
2. For each component `FieldDefinition`:
   - Call `FindByDocumentID(...)` — returns `[]*Component` ordered by `sort_order ASC`.
   - If `repeatable == false`:
     - Take the first element (or nil). Merge as `doc.Fields[fieldName] = comp.Fields` (single object).
   - If `repeatable == true`:
     - Map all elements to `[]map[string]any`. Merge as `doc.Fields[fieldName] = [...]` (array).

### Publish Flow

Copy draft component rows to published version. `UpsertAll` already handles this — `sort_order` is preserved in the copy.

### Delete Flow

No change — `DeleteByDocumentID` already deletes all rows for a `(document_id, locale)` tuple regardless of count.

---

## 9. Component Repository Changes

### `compToRow` / `rowToComp`

Add `sort_order` to the row map:

```go
// compToRow
row["sort_order"] = c.SortOrder

// rowToComp
if v, ok := row["sort_order"]; ok {
    comp.SortOrder = toInt(v)
}
```

Add `"sort_order"` to the `systemCols` set in `rowToComp`.

### `FindByDocumentID`

Already orders by `gorm_id ASC`. Change to `ORDER BY sort_order ASC, gorm_id ASC` for deterministic ordering.

### `createComponentTable`

Add `"sort_order INTEGER DEFAULT 0"` to the column list (between `locale` and field columns).

### `addMissingComponentColumns`

Add `sort_order` to the set of system columns that are always ensured:
```go
if !cols["sort_order"] {
    sql := fmt.Sprintf("ALTER TABLE %s ADD COLUMN sort_order INTEGER DEFAULT 0", table)
    // execute
}
```

---

## 10. GraphQL Changes

### Schema Generation

For a content type with component fields, the schema builder generates:

**Non-repeatable:**
```graphql
type BlogPost {
  title: String
  banner: BlogPostBanner
}
```

**Repeatable:**
```graphql
type CvPage {
  position: String
  technicalSkills: [CvPageTechnicalSkill!]
}
```

The component object type (`CvPageTechnicalSkill`) is generated the same way regardless of `repeatable`. Only the parent field type changes: `Type` vs `[Type!]`.

### Input Generation

**Non-repeatable:**
```graphql
input BlogPostInput {
  banner: BlogPostBannerInput
}
```

**Repeatable:**
```graphql
input CvPageInput {
  technicalSkills: [CvPageTechnicalSkillInput!]
}
```

### Filter Generation

- Non-repeatable components: nested filter on component sub-fields (existing behavior).
- Repeatable components: **no filter support** (filtering on array elements is complex; defer to a future spec). The field is excluded from the generated `<Type>Filter`.

---

## 11. Frontend — Admin Panel UI

### Non-Repeatable Component (No Change)

Rendered as a bordered `<fieldset>` with the component's fields inside. Existing `renderSchemaField` behavior.

### Repeatable Component

Rendered as a list of component entries with controls:

```
┌─ technicalSkills ───────────────────────────────┐
│                                                 │
│  ┌─ #1 ──────────────────────── [↑] [↓] [✕] ─┐ │
│  │  category: [___________]                    │ │
│  │  skills:   [___________]                    │ │
│  └─────────────────────────────────────────────┘ │
│                                                 │
│  ┌─ #2 ──────────────────────── [↑] [↓] [✕] ─┐ │
│  │  category: [___________]                    │ │
│  │  skills:   [___________]                    │ │
│  └─────────────────────────────────────────────┘ │
│                                                 │
│  [ + Add entry ]                                │
└─────────────────────────────────────────────────┘
```

**Controls per entry:**
- **Move up** (disabled on first item)
- **Move down** (disabled on last item)
- **Remove** (always enabled; no confirmation for now)

**Footer:**
- **Add entry** button appends a new empty component at the end

### Form State

Repeatable component data is stored in the form as an array under `fieldName`:

```typescript
// Form values shape
{
  technicalSkills: [
    { category: "Frontend", skills: "React" },
    { category: "Backend", skills: "Go" }
  ]
}
```

The `FormProvider` already uses dot-notation names auto-deserialized to nested JSON. For repeatable components, the field names use array indexing:

- `technicalSkills.0.category`
- `technicalSkills.0.skills`
- `technicalSkills.1.category`
- `technicalSkills.1.skills`

### Implementation Approach

Add a `RepeatableComponentField` React component in `apps/web/src/components/form/inputs/RepeatableComponentField.tsx`:

```typescript
interface RepeatableComponentFieldProps {
  name: string;
  fields: FieldDefinition[];
}
```

This component:
1. Reads the current array value from the form context.
2. Renders each entry as a bordered card with move/remove controls.
3. Provides an "Add entry" button.
4. On reorder: swaps array elements and updates form state.
5. On remove: splices the element and re-indexes.
6. On add: appends an empty object `{}` to the array.

### `renderSchemaField` Changes

Update the `component` branch:

```typescript
if (field.type === 'component') {
  if (field.repeatable) {
    return (
      <RepeatableComponentField
        key={fieldName}
        name={fieldName}
        fields={field.fields ?? []}
      />
    );
  }
  // existing non-repeatable rendering...
}
```

---

## 12. File Map (Changes)

All paths relative to project root.

### Backend (`apps/api/`)

| File | Change |
|------|--------|
| `internal/domain/entity/content_type.go` | Add `Repeatable bool` to `FieldDefinition` |
| `internal/domain/entity/component.go` | Add `SortOrder int` to `Component` |
| `internal/usecase/content_type/schema_loader.go` | No validation changes (boolean needs no extra validation) |
| `internal/usecase/content_type/sync.go` | Add `Repeatable` to `fieldsEqual` comparison |
| `internal/infrastructure/gormdb/component_repository.go` | Add `sort_order` to table creation, row serialization, ordering, column migration |
| `internal/infrastructure/gormdb/document_repository.go` | Handle repeatable vs non-repeatable in save/read flows |
| `internal/usecase/document/document_usecase.go` | Validate component data shape (object vs array) based on `repeatable` |
| `graphql/dynamic/schema_builder.go` | Emit `[Type!]` for repeatable component fields |

### Frontend (`apps/web/`)

| File | Change |
|------|--------|
| `src/types/cms.ts` | Add `repeatable?: boolean` to `FieldDefinition` |
| `src/components/form/inputs/RepeatableComponentField.tsx` | **New** — repeatable component UI with add/remove/reorder |
| `src/pages/admin/panels/content-type/renderSchemaField.tsx` | Branch on `field.repeatable` for component fields |

### Content Types

| File | Change |
|------|--------|
| `content-types/cv-page.json` | Add `"repeatable": true` to `technicalSkills` |

---

## 13. Testing

### Backend

**Schema loader (`schema_loader_test.go`):**
- Valid JSON with `"repeatable": true` on component → parses correctly, `Repeatable == true`
- Valid JSON without `"repeatable"` → `Repeatable == false` (default)
- `"repeatable"` on non-component field → ignored (no error)

**Sync (`sync_test.go`):**
- Toggling `repeatable` on a component field triggers content-type update

**Component repository (`component_repository_test.go`):**
- `EnsureCollection` creates table with `sort_order` column
- `addMissingComponentColumns` adds `sort_order` to existing table
- `UpsertAll` with multiple components preserves `sort_order`
- `FindByDocumentID` returns components ordered by `sort_order ASC`

**Document usecase (`document_usecase_test.go`):**
- Save with repeatable component: array data → stored as multiple component rows with correct `sort_order`
- Save with non-repeatable component: object data → stored as single component row with `sort_order = 0`
- Save with repeatable component but object data → 400 error
- Save with non-repeatable component but array data → 400 error
- Read with repeatable component: returns array in `doc.Fields`
- Read with non-repeatable component: returns object in `doc.Fields`
- Publish copies component rows with `sort_order` preserved

**GraphQL (`schema_builder_test.go`):**
- Non-repeatable component field → generates `Type` field
- Repeatable component field → generates `[Type!]` field

### Frontend

**`RepeatableComponentField` tests:**
- Renders all entries from form state
- "Add entry" appends new item, form becomes dirty
- "Remove" splices item, re-indexes remaining entries
- "Move up/down" swaps entries, preserves data
- Move up disabled on first item, move down disabled on last

**`renderSchemaField` tests:**
- `repeatable: true` → renders `RepeatableComponentField`
- `repeatable: false` or missing → renders existing `<fieldset>`

---

## 14. Boundaries

| Rule | Detail |
|------|--------|
| **Always** | Default `repeatable` to `false` when omitted — backward-compatible |
| **Always** | Validate data shape at usecase layer (object for non-repeatable, array for repeatable) |
| **Always** | Preserve `sort_order` through save → publish → read cycle |
| **Always** | Order repeatable components by `sort_order ASC` in all queries |
| **Always** | Include `sort_order` in component table creation and migration |
| **Never** | Add unique constraint on `(document_id, version, locale)` in component tables |
| **Never** | Support filtering on repeatable component sub-fields in GraphQL (defer to future spec) |
| **Never** | Allow `repeatable` on non-component field types (ignore silently) |
| **Never** | Allow nested repeatable components (nesting depth limit of 2 still applies) |
| **Ask first** | Adding a max item count for repeatable components |
| **Ask first** | Drag-and-drop reordering (start with button-based move up/down) |

---

## 15. Migration Notes

- **Existing component data**: All existing component fields are non-repeatable. Adding `sort_order DEFAULT 0` to existing tables is additive and non-breaking.
- **Existing JSON schemas**: No `repeatable` field means `false` — existing schemas continue to work without modification.
- **Content-type sync**: On next startup after deploying, `EnsureCollection` adds the `sort_order` column to all existing component tables via `addMissingComponentColumns`.
