# SPEC: Remove Layout Field Type + Add Field Width Property

**Status:** Draft
**Modules:** content, frontend
**Rules to update:** `rules/content.md`, `rules/content-type-parsing.md`, `rules/frontend.md`

---

## 1. Objective

Remove the `layout` field type from content-type JSON schemas and all backend/frontend code that supports it. Replace with a new `width` property on each field definition that controls how wide a field renders in the CMS form editor.

**Why:** The `layout` type was a structural wrapper used solely for UI grouping (2-column grid). It added complexity throughout the stack — schema validation, flattening logic in 6+ backend locations, and recursive handling in GraphQL. A per-field `width` property is simpler, more flexible (supports 3 widths instead of just 50/50), and eliminates the need for wrapper nodes in the schema.

---

## 2. Changes — Backend (apps/api)

### 2.1 FieldDefinition Entity

**File:** `internal/domain/entity/content_type.go`

Add a `Width` field:

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

- `Width` valid values: `"100%"`, `"50%"`, `"1/3"`. Empty string = defaults to `"100%"`.
- `Width` has **no effect on storage** — it is purely a UI hint passed through to the frontend.
- `Width` is only meaningful on non-component, non-layout fields. Components always render full-width.

### 2.2 Schema Loader (`usecase/content_type/schema_loader.go`)

- **Remove** the `case "layout":` validation block entirely.
- If a JSON file still contains `type: "layout"` fields, the loader should silently ignore them (unknown types pass through — `json.Unmarshal` will still parse them but no validation code runs for them).
- **No new validation** for `width` — any string is accepted; the frontend interprets it.

### 2.3 Remove All `flattenLayoutFields` Functions

These functions exist in 3 locations and must all be removed:

| File | Function | Action |
|------|----------|--------|
| `internal/delivery/http/handler/document_handler.go:49-59` | `flattenLayoutFields()` | Remove function + all call sites |
| `internal/infrastructure/gormdb/document_repository.go:67-77` | `flattenLayoutFields()` | Remove function + all call sites |
| `internal/infrastructure/gormdb/component_repository.go` | Uses `flattenLayoutFields` from document_repository.go (same package) | Update call sites |

After removal, all places that called `flattenLayoutFields(fields)` should just use `fields` directly (since layout fields no longer exist in the schema).

### 2.4 Content-Type Handler (`content_type_handler.go`)

- **Remove** the `if field.Type == "layout"` branch in `UpdateListFields()` (line ~98).
- The loop that builds allowed field names should iterate `ct.Fields` directly.

### 2.5 Document Usecase (`usecase/document/document_usecase.go`)

- **Remove** the `if f.Type == "layout"` branch in `resolveMediaFields()` (line ~456).
- No longer needed since layout fields won't exist.

### 2.6 GraphQL Schema Builder (`graphql/dynamic/schema_builder.go`)

- **Remove** `flattenLayoutFieldsDef()` function (lines 160-170).
- Replace all call sites with direct field iteration.
- In `BuildContentTypeSDL()` and `writeComponentType()`, remove flattening calls.

### 2.7 GraphQL Resolver Factory (`graphql/dynamic/resolver_factory.go`)

- **Remove** the `if fd.Type == "layout" { continue }` checks (lines 218-219, 245-246).
- **Remove** the `else if fd.Type != "layout"` condition (line 535) — simplify to just the remaining logic.
- In `buildComponentType()`, remove `flattenLayoutFieldsDef()` call (line 188).

### 2.8 Sync Engine (`usecase/content_type/sync.go`)

- `fieldsEqual()` — add `Width` to the comparison:
  ```go
  if a[i].Width != b[i].Width { return false }
  ```
- No other sync changes needed — layout was never stored as a DB concept.

### 2.9 Test Updates

| Test File | Changes |
|-----------|---------|
| `schema_loader_test.go` | Remove `TestLoadDefinitions_LayoutEmptyFields` and `TestLoadDefinitions_LayoutContainsComponent` |
| `document_repository_test.go` | Remove `TestDocumentRepository_LayoutFieldsRoundTrip` and `TestDocumentRepository_LayoutFieldsMigrateExistingTable` |
| `schema_builder_test.go` | Remove `TestBuildContentTypeSDL_TopLevelLayout` |
| Test fixtures `testdata/invalid/layout-empty-fields/` | Delete directory |
| Test fixtures `testdata/invalid/layout-contains-component/` | Delete directory |

### 2.10 Content-Type JSON Files — Migration

All JSON files must have their `layout` wrappers unwrapped. Each child of a layout becomes a direct sibling with the appropriate `width`.

**Mapping rule:** A former layout with N children means each child was displayed in a 2-column grid. Convert to:
- 2 children in a layout → each child gets `"width": "50%"`
- 1 child in a layout → no width needed (defaults to 100%)

---

## 3. Changes — Frontend (apps/web)

### 3.1 TypeScript Type (`types/cms.ts`)

Update `FieldDefinition`:

```typescript
export interface FieldDefinition {
  name: string;
  type: string;
  ext?: string[];
  width?: '100%' | '50%' | '1/3';
  repeatable?: boolean;
  fields?: FieldDefinition[];
}
```

**Remove** the `flattenFields()` function entirely.

### 3.2 Remove `flattenFields` Usages

| File | Current Usage | Replacement |
|------|---------------|-------------|
| `components/collection/ColumnChooserDialog.tsx` | `flattenFields(contentType.Fields ?? []).filter(...)` | `(contentType.Fields ?? []).filter(...)` |
| `pages/admin/panels/collection-type/layout/CollectionListPage.tsx` | `flattenFields(contentType.Fields ?? [])` | `(contentType.Fields ?? [])` |

### 3.3 Form Renderer (`renderSchemaField.tsx`)

**Current:** Layout fields render as `<div className="grid gap-4 md:grid-cols-2">` wrapping their children.

**New:** Remove layout handling. Instead, the top-level container wraps **all** fields in a responsive 6-column grid. Each field applies its own column span based on `width`.

```
Grid container: "grid grid-cols-1 md:grid-cols-6 gap-4"

Width mapping:
  - "100%" (or undefined) → "md:col-span-6"  (full width)
  - "50%"                 → "md:col-span-3"  (half width)
  - "1/3"                 → "md:col-span-2"  (one-third width)

Mobile: all fields are grid-cols-1 → full width automatically
```

**Changes to `renderField()`:**

1. Remove the `if (field.type === 'layout')` branch entirely.
2. Each field wrapper `<div>` gets a width class: `md:col-span-{span}` based on `field.width`.
3. Component fields (`type: "component"`) always span full width (`md:col-span-6`) regardless of `width`.

**Changes to `renderSchemaField()` / parent caller:**

The parent that calls `renderSchemaField` in a loop must wrap the fields in:
```tsx
<div className="grid grid-cols-1 md:grid-cols-6 gap-4">
  {fields.map((field, index) => renderSchemaField(field, prefix, keyPrefix, index))}
</div>
```

**Inside components:** The same 6-column grid applies to component children, so fields inside a component also respect `width`.

### 3.4 RepeatableComponentField

Each repeatable component entry's children should also be wrapped in the 6-column grid, so `width` works at all nesting levels.

### 3.5 Test Updates

| Test File | Changes |
|-----------|---------|
| `ContentTypeBuilder.test.tsx` | Remove `ContentTypeBuilder — layout` test suite. Add new test for width-based column spans. |

---

## 4. Content-Type JSON Migration

### 4.1 `cv-contact.json` (Before → After)

**Before:**
```json
{
  "type": "layout",
  "fields": [
    { "name": "name", "type": "text" },
    { "name": "address", "type": "text" }
  ]
}
```

**After:**
```json
{ "name": "name", "type": "text", "width": "50%" },
{ "name": "address", "type": "text", "width": "50%" }
```

### 4.2 `en-vocab-pack.json` (Before → After)

**Before:**
```json
{
  "type": "layout",
  "fields": [
    { "name": "packName", "type": "text" },
    { "name": "packTitle", "type": "text" }
  ]
}
```

**After:**
```json
{ "name": "packName", "type": "text", "width": "50%" },
{ "name": "packTitle", "type": "text", "width": "50%" }
```

### 4.3 `cv-page.json`

All layout wrappers unwrapped. Each pair of fields formerly in a layout gets `"width": "50%"`. Fields inside components follow the same rule.

### 4.4 `common-text.json`

No change needed — no layout fields.

---

## 5. What Does NOT Change

- **Database storage:** No migration needed. Layout was never stored as data — children were already promoted. The `width` field is stored in the `fields` JSON column of `content_types` table but has no effect on document tables.
- **API contracts:** No REST/GraphQL response shape changes. `width` is metadata on the content type, not on documents.
- **Draft/Publish workflow:** Unaffected.
- **Component behavior:** Components still work identically. Layout inside components was just UI grouping — now replaced by `width` on individual fields.

---

## 6. Rule Updates Required

### 6.1 `rules/content-type-parsing.md`

- Remove §2.2 row for `layout` from Field Types table
- Remove §2.4 (Layout Fields section) entirely
- Remove §2.5 note about "Layout does NOT count as a nesting level"
- Remove §4.2 layout validation rules
- Add `Width` to §2.1 FieldDefinition struct
- Update §8 examples to show `width` instead of layout

### 6.2 `rules/content.md`

- Remove any reference to layout flattening in §1.2 FieldDefinition
- Add `width` to the FieldDefinition struct

### 6.3 `rules/frontend.md`

- Update §3.2 to note that fields support `width` property
- Update form rendering description to describe 6-column grid

---

## 7. Acceptance Criteria

### Backend
- [ ] `type: "layout"` no longer validated or special-cased anywhere in Go code
- [ ] All `flattenLayoutFields` / `flattenLayoutFieldsDef` functions removed
- [ ] `Width` field added to `FieldDefinition` entity
- [ ] `fieldsEqual()` compares `Width`
- [ ] All JSON content-type files have layouts unwrapped with `width` set
- [ ] `go vet ./...` passes
- [ ] `go test ./...` passes
- [ ] `go build ./...` passes

### Frontend
- [ ] `flattenFields()` removed from `types/cms.ts`
- [ ] All call sites updated to use fields directly
- [ ] `renderSchemaField` no longer handles `type === 'layout'`
- [ ] Form fields render in a 6-column responsive grid
- [ ] `width: "100%"` → full width (span 6)
- [ ] `width: "50%"` → half width (span 3)
- [ ] `width: "1/3"` → one-third width (span 2)
- [ ] Mobile: all fields full width (grid-cols-1)
- [ ] Fields inside components also respect `width`
- [ ] `npm run lint` passes
- [ ] `npm run build` passes

---

## 8. Boundaries

| Rule | Detail |
|------|--------|
| **Always** | Each field in the JSON schema is a data field with `name` — no structural wrappers |
| **Always** | `width` defaults to `"100%"` when omitted |
| **Always** | Components render full-width regardless of `width` |
| **Always** | Mobile layout is single-column (grid-cols-1) |
| **Never** | Use `type: "layout"` in content-type schemas |
| **Never** | Add flatten/promote logic for field types |
| **Never** | Let `width` affect database column creation or storage |
