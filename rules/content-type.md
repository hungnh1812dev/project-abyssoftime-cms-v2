# RULES — Content Type

**Scope:** Content type entity, schema-as-code, schema sync, JSON field definitions, configurable list columns.

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

## 2. Schema Sync Rules

### 2.1 Sync Engine (`usecase/content_type/sync.go`)
- New file → create ContentType + per-content-type document collection/table
- Changed file → update ContentType schema in place
- Field removed → drop from schema, leave stored data untouched
- File deleted → delete ContentType, cascade-delete all entries, drop collection/table
- **NEVER** let sync write back to JSON definition files
- **NEVER** overwrite user-configured `ListFields` — sync only seeds when empty

### 2.2 Schema Loader (`usecase/content_type/schema_loader.go`)
- Reads all `*.json` files from `CONTENT_TYPES_DIR`
- Validates field definitions (name, type, nesting depth)
- `listFields` validation removed — no longer in JSON schemas
- `validateFields` starts at depth=1, checks component nesting ≤ 3

### 2.3 EnsureCollection (Content-Specific)
- Document tables: `documents_<slug_underscored>` (hyphens → underscores)
- Component tables: `components_<slug_underscored>_<component_path>`
- Non-destructive: create if missing, add columns if existing
- **NEVER** drop and recreate tables

---

## 3. Configurable List Columns Rules

### 3.1 Source of Truth
- `ListFields` stored in content_types DB table — UI-managed
- **NEVER** defined in JSON schema files (permanently removed)
- Startup sync only seeds when DB value is empty/nil — never overwrites

### 3.2 Column Layout
```
| Id | [selected content fields] | [selected system fields] | Status | Actions |
```
- Locked columns: Id (first), Status (before Actions), Actions (last) — not in popup
- Content fields: from `Fields` definition, excluding `component` type
- System fields: CreatedAt, UpdatedAt, UpdatedBy
- Default (empty listFields): first 3 content fields + all system fields

### 3.3 Validation
- Each entry must be a known content field name OR a known system field
- Component-type fields rejected
- Empty array = revert to defaults

---

## 4. Boundaries Summary

| Rule | Detail |
|---|---|
| **Always** | Validate slug format at both usecase and handler levels |
| **Always** | Default `repeatable` to `false` when omitted |
| **Always** | Non-destructive `EnsureCollection` — never DROP+CREATE |
| **Always** | Max 3 levels of component nesting; fatal error if exceeded |
| **Never** | Add API/UI to create/edit/delete ContentType structure |
| **Never** | Let sync write back to JSON definition files |
| **Never** | Define `listFields` in JSON schema files |
| **Never** | Let sync overwrite user-configured `listFields` |
| **Ask first** | Increasing max nesting depth beyond 3 |
