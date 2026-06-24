# RULES — PostgreSQL/GORM Communication Patterns

**Scope:** Every `infrastructure/gormdb/*.go` file. Governs how Go code reads, writes, queries, and manages tables via GORM + raw SQL for PostgreSQL/SQLite.
**Package:** `project-abyssoftime-cms-v2/api/internal/infrastructure/gormdb`

---

## 1. Client Connection

### 1.1 Initialization (`client.go`)
```go
func NewClient(driver, dsn string) (*gorm.DB, error)
func AutoMigrate(db *gorm.DB) error
```
- Driver from `DB_DRIVER` env var (`postgres` or `sqlite` for tests)
- DSN from env vars: `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USERNAME`, `DB_PASSWORD`, `DB_SSL_MODE`
- Logger: `logger.Silent` — no GORM query logging by default
- Ping after connect for PostgreSQL (verify connection)
- `resolveDialector(driver, dsn)` — build-tag-based driver selection (CGO for SQLite in tests)

### 1.2 AutoMigrate — Static Tables Only
```go
func AutoMigrate(db *gorm.DB) error {
    return db.AutoMigrate(
        &entity.User{},
        &entity.ContentType{},
        &entity.MediaAsset{},
        &entity.RoleEntity{},
        &entity.Invite{},
        &entity.AccessToken{},
        &entity.Locale{},
    )
}
```
- Called once at startup
- **Only** static entity tables — **NEVER** include Document or Component
- Dynamic tables (`documents_*`, `components_*`) created by `EnsureCollection`

---

## 2. Repository Struct Patterns

### 2.1 Static Entity Repositories (User, Role, ContentType, MediaAsset, etc.)
```go
type <entity>Repository struct {
    database *gorm.DB        // field name is "database"
}

func New<Entity>Repository(database *gorm.DB) repository.XRepository {
    return &<entity>Repository{database: database}
}
```
- Field name: `database` (not `db`)
- Receives `*gorm.DB` — the GORM database handle
- Queries use `r.database.WithContext(ctx)` chaining

### 2.2 Dynamic Table Repositories (Document, Component)
```go
type documentRepository struct {
    database *gorm.DB
}

func (r *documentRepository) table(slug string) *gorm.DB {
    return r.database.Table(documentTableName(slug))
}
```
- Uses `r.database.Table(name)` to target dynamic tables
- Table name resolved via helper functions

---

## 3. Slug-to-Table Name Conversion (`slug.go`)

```go
func sanitizeSlug(slug string) string {
    return strings.ReplaceAll(slug, "-", "_")
}

func documentTableName(slug string) string {
    return "documents_" + sanitizeSlug(slug)
}

func componentTableName(slug, componentName string) string {
    return "components_" + sanitizeSlug(slug) + "_" + sanitizeSlug(componentName)
}
```

| Content-Type Slug | Document Table | Component Table (field `banner`) |
|---|---|---|
| `blog-posts` | `documents_blog_posts` | `components_blog_posts_banner` |
| `cv-page` | `documents_cv_page` | `components_cv_page_skills` |
| `en-vocab-pack` | `documents_en_vocab_pack` | — |

- Hyphens → underscores (PostgreSQL table name convention)
- Component path appended with underscore separator
- Nested component paths: `components_cv_page_experiences_roles`

---

## 4. Field Name Conversion Helpers

### 4.1 camelCase ↔ snake_case
```go
func toSnakeCase(s string) string    // "siteName" → "site_name"
func toCamelCase(s string) string    // "site_name" → "siteName"
```
- `toSnakeCase`: inserts `_` before uppercase letters, lowercases all
- `toCamelCase`: splits on `_`, capitalizes first letter of each part after first
- Used in `docToRow` / `rowToDoc` to convert field names between Go (camelCase) and SQL (snake_case)

### 4.2 Type Conversion Helpers
```go
func toString(v any) string       // nil-safe string extraction
func toUint(v any) uint           // handles int32, int64, uint, float64, int
func toInt(v any) int             // handles int, int32, int64, float64, uint
func toTime(v any) time.Time      // handles time.Time and RFC3339 string
```
- All are nil-safe — return zero value on nil
- Used in `rowToDoc` / `rowToComp` to convert `map[string]any` values
- `toTime` parses `time.RFC3339Nano` format from string

---

## 5. Field Value Serialization

### 5.1 Serialize (Go → SQL)
```go
func serializeFieldValue(v any) any {
    switch v.(type) {
    case map[string]any, []any:
        b, _ := json.Marshal(v)
        return string(b)        // complex types → JSON string
    default:
        return v                // scalars pass through
    }
}
```
- `map[string]any` or `[]any` → JSON string (`TEXT` column)
- Scalars (string, int, float, bool) → pass through directly

### 5.2 Deserialize (SQL → Go)
```go
func deserializeFieldValue(v any) any {
    s, ok := v.(string)
    if !ok || len(s) == 0 { return v }
    if (s[0] == '{' && s[len(s)-1] == '}') || (s[0] == '[' && s[len(s)-1] == ']') {
        var parsed any
        if json.Unmarshal([]byte(s), &parsed) == nil {
            return parsed
        }
    }
    return v
}
```
- Detects JSON strings by first/last character: `{}` or `[]`
- Parses back to `map[string]any` or `[]any`
- Non-JSON strings pass through as-is

---

## 6. Row Conversion — Documents

### 6.1 `docToRow` (Entity → SQL row map)
```go
func docToRow(doc *entity.Document) map[string]any
```
- System columns explicitly mapped: `document_id`, `version`, `locale`, `created_at`, `updated_at`, `published_at`, `created_by`, `updated_by`, `published_by`
- `published_at`: set to `nil` if pointer is nil; dereference if non-nil
- Content fields: iterate `doc.Fields`, convert keys with `toSnakeCase`, values with `serializeFieldValue`
- **NEVER** include `gorm_id` in insert row — auto-incremented

### 6.2 `rowToDoc` (SQL row map → Entity)
```go
func rowToDoc(row map[string]any) *entity.Document
```
- Extract system columns by known keys, convert with type helpers
- `published_at`: check nil AND zero time — only set pointer if non-zero
- Remaining columns (not in `systemCols` set) → content `Fields`, keys converted with `toCamelCase`, values with `deserializeFieldValue`
- `systemCols` set:
```go
map[string]bool{
    "gorm_id": true, "document_id": true, "version": true, "locale": true,
    "created_at": true, "updated_at": true, "published_at": true,
    "created_by": true, "updated_by": true, "published_by": true,
}
```

### 6.3 Key Rule
- System columns are hardcoded in the set — **NEVER** leak into `Fields`
- Content columns are everything else — **NEVER** hardcode content field names

---

## 7. Row Conversion — Components

### 7.1 `compToRow` (Component → SQL row map)
```go
func compToRow(comp *entity.Component) map[string]any
```
- Writes **exactly one** FK column based on which is populated:
  - `comp.ParentComponentID != ""` → write `parent_component_id`, skip `document_id`
  - Otherwise → write `document_id`, skip `parent_component_id`
- **CRITICAL**: Writing the absent FK column causes SQL error (column doesn't exist in table)
- Content fields: iterate `comp.Fields`, `toSnakeCase` keys, `serializeFieldValue` values

### 7.2 `rowToComp` (SQL row map → Component)
```go
func rowToComp(row map[string]any) *entity.Component
```
- Reads whichever FK column exists (absent column just won't appear in row)
- Both `document_id` and `parent_component_id` in `systemCols`:
```go
map[string]bool{
    "gorm_id": true, "component_id": true, "document_id": true, "parent_component_id": true,
    "version": true, "locale": true, "sort_order": true, "created_at": true, "updated_at": true,
}
```

---

## 8. CRUD Operation Patterns — Static Entities

### 8.1 Create
```go
func (r *xRepository) Create(ctx context.Context, entity *entity.X) error {
    return r.database.WithContext(ctx).Create(entity).Error
}
```
- GORM auto-handles tags (`gorm:"column:..."`)
- Auto-increment PK filled by DB
- Unique constraint violation → GORM returns error (no specific check like MongoDB)

### 8.2 FindOne
```go
func (r *xRepository) FindByX(ctx context.Context, value string) (*entity.X, error) {
    var result entity.X
    if err := r.database.WithContext(ctx).Where("column = ?", value).First(&result).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, pkgerrors.ErrNotFound
        }
        return nil, err
    }
    return &result, nil
}
```
- **ALWAYS** use `errors.Is(err, gorm.ErrRecordNotFound)` — NOT `==`
- Use `First` for single result (not `Find`)
- Use parameterized queries `Where("col = ?", val)` — **NEVER** string interpolation

### 8.3 FindMany
```go
func (r *xRepository) FindAll(ctx context.Context) ([]*entity.X, error) {
    var results []*entity.X
    if err := r.database.WithContext(ctx).Order("created_at DESC").Find(&results).Error; err != nil {
        return nil, err
    }
    return results, nil
}
```
- Use `Find` for multiple results
- `Order("column_name DESC")` — string-based sort
- Empty results return `[]` (not nil)

### 8.4 FindMany Paginated
```go
var total int64
r.database.WithContext(ctx).Model(&entity.X{}).Count(&total)
r.database.WithContext(ctx).Order("...").Offset(offset).Limit(limit).Find(&results)
```
- Count with `.Model(&entity.X{}).Count(&total)`
- Paginate with `.Offset(offset).Limit(limit)`

### 8.5 Update
```go
func (r *xRepository) Update(ctx context.Context, entity *entity.X) error {
    result := r.database.WithContext(ctx).Save(entity)
    if result.Error != nil { return result.Error }
    if result.RowsAffected == 0 { return pkgerrors.ErrNotFound }
    return nil
}
```
- Use `.Save(entity)` — updates all fields based on primary key
- Check `RowsAffected == 0` → `pkgerrors.ErrNotFound`

### 8.6 Partial Update (Locale ClearDefault)
```go
r.database.WithContext(ctx).Model(&entity.Locale{}).
    Where("is_default = ?", true).
    Update("is_default", false).Error
```
- Use `.Model(&entity.X{}).Where(...).Update("col", val)` for single-field updates
- Only used for `ClearDefault` — most updates use `.Save()`

### 8.7 Delete
```go
func (r *xRepository) Delete(ctx context.Context, id string) error {
    result := r.database.WithContext(ctx).Where("document_id = ?", id).Delete(&entity.X{})
    if result.Error != nil { return result.Error }
    if result.RowsAffected == 0 { return pkgerrors.ErrNotFound }
    return nil
}
```
- Use `.Where(...).Delete(&entity.X{})` with typed model
- Check `RowsAffected == 0` → `pkgerrors.ErrNotFound`

### 8.8 Delete All (Dangerous — Dynamic Tables Only)
```go
r.table(slug).WithContext(ctx).
    Session(&gorm.Session{AllowGlobalUpdate: true}).
    Delete(map[string]any{}).Error
```
- `AllowGlobalUpdate: true` required to delete without WHERE
- Used only in `DeleteAllByContentType`
- Use `map[string]any{}` as Delete argument for dynamic tables (no model)

---

## 9. CRUD Operation Patterns — Dynamic Tables (Documents)

### 9.1 FindOne
```go
var result map[string]any
err := r.table(slug).WithContext(ctx).
    Where("document_id = ? AND version = ? AND locale = ?", docID, version, locale).
    Take(&result).Error
```
- Use `Take` (not `First`) — avoids default ordering
- Scan into `map[string]any` — **NOT** entity struct (dynamic columns)
- Convert with `rowToDoc(result)`

### 9.2 Upsert (Document-Specific)
```go
func (r *documentRepository) upsert(ctx context.Context, slug string, doc *entity.Document) error {
    var existing map[string]any
    err := r.table(slug).WithContext(ctx).
        Where("document_id = ? AND version = ? AND locale = ?", ...).
        Take(&existing).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            row := docToRow(doc)
            return r.table(slug).WithContext(ctx).Create(row).Error  // INSERT
        }
        return err
    }
    doc.GormID = toUint(existing["gorm_id"])
    row := docToRow(doc)
    row["gorm_id"] = doc.GormID
    return r.table(slug).WithContext(ctx).Where("gorm_id = ?", doc.GormID).Updates(row).Error  // UPDATE
}
```
- Try `Take` first → if `ErrRecordNotFound` → `Create` (INSERT)
- If found → extract `gorm_id` from existing → `Updates` (UPDATE by PK)
- Use `Updates(row)` — partial update with map
- **NEVER** use GORM's built-in upsert (`Clauses(clause.OnConflict{})`) — this manual approach is the project pattern

### 9.3 FindMany Paginated
```go
var rows []map[string]any
q := r.table(slug).WithContext(ctx).
    Where("version = ? AND locale = ?", version, locale)
q.Count(&total)
q.Order(resolveGormSortClause(orderBy, sortDir)).Offset(start).Limit(size).Find(&rows)
```
- Scan into `[]map[string]any`
- Convert each with `rowToDoc(row)`

### 9.4 Delete
```go
r.table(slug).WithContext(ctx).
    Where("document_id = ? AND locale = ?", docID, locale).
    Delete(map[string]any{}).Error
```
- Use `map[string]any{}` as model — dynamic tables have no Go model

---

## 10. CRUD Operation Patterns — Component Tables

### 10.1 UpsertAll (Delete + Insert)
```go
func (r *componentRepository) UpsertAll(...) error {
    // Step 1: Delete existing rows for this (documentID, version, locale)
    tbl.Where("document_id = ? AND version = ? AND locale = ?", ...).Delete(map[string]any{})
    // Step 2: Insert new rows one by one
    for _, comp := range components {
        comp.DocumentID = documentID
        comp.ParentComponentID = ""     // CLEAR the other FK
        row := compToRow(comp)
        r.table(...).Create(row)
    }
}
```

### 10.2 UpsertAllByParent (Nested Components)
```go
func (r *componentRepository) UpsertAllByParent(...) error {
    // Step 1: Delete by parent_component_id
    tbl.Where("parent_component_id = ? AND version = ? AND locale = ?", ...).Delete(map[string]any{})
    // Step 2: Insert with ParentComponentID set, DocumentID cleared
    for _, comp := range components {
        comp.ParentComponentID = parentID
        comp.DocumentID = ""            // CLEAR the other FK
        row := compToRow(comp)
        r.table(...).Create(row)
    }
}
```
- **CRITICAL**: Clear the opposite FK before `compToRow` — prevents writing absent column

### 10.3 Ordering
```go
.Order("sort_order ASC, gorm_id ASC")
```
- Primary sort: `sort_order` (client-controlled array index)
- Secondary sort: `gorm_id` (insertion order, deterministic tiebreaker)

---

## 11. EnsureCollection — Non-Destructive Schema Sync

### 11.1 Document Tables
```go
func (r *documentRepository) EnsureCollection(ctx, slug string, fields []FieldDefinition) error {
    table := documentTableName(slug)
    if !r.database.Migrator().HasTable(table) {
        return r.createDocumentTable(ctx, table, fields)    // CREATE TABLE
    }
    return r.addMissingDocumentColumns(ctx, table, fields)  // ALTER TABLE ADD COLUMN
}
```

### 11.2 Component Tables
```go
func (r *componentRepository) EnsureCollection(ctx, slug, name string, fields []FieldDefinition, isNested bool) error {
    table := componentTableName(slug, name)
    if !r.database.Migrator().HasTable(table) {
        return r.createComponentTable(ctx, table, fields, isNested)
    }
    return r.addMissingComponentColumns(ctx, table, fields, isNested)
}
```

### 11.3 Create Table — Raw SQL
```go
var cols []string
if r.isPostgres() {
    cols = append(cols, "gorm_id SERIAL PRIMARY KEY")
} else {
    cols = append(cols, "gorm_id INTEGER PRIMARY KEY AUTOINCREMENT")
}
// ... add columns ...
sql := fmt.Sprintf("CREATE TABLE %s (%s)", table, strings.Join(cols, ", "))
r.database.WithContext(ctx).Exec(sql)
```
- Use raw SQL `CREATE TABLE` — **NOT** GORM's `AutoMigrate`
- PK syntax: `SERIAL PRIMARY KEY` (PostgreSQL) vs `INTEGER PRIMARY KEY AUTOINCREMENT` (SQLite)
- Check `r.isPostgres()` for dialect differences

### 11.4 Column Order in CREATE TABLE

**Document table:**
```
gorm_id, document_id, version, locale, <field_columns...>, created_at, updated_at, published_at, created_by, updated_by, published_by
```

**Top-level component table (isNested=false):**
```
gorm_id, component_id, document_id, version, locale, sort_order, <field_columns...>, created_at, updated_at
```

**Nested component table (isNested=true):**
```
gorm_id, component_id, parent_component_id, version, locale, sort_order, <field_columns...>, created_at, updated_at
```

### 11.5 Field Column Type Mapping
```go
func fieldColumnType(fieldType string) string {
    switch fieldType {
    case "number":  return "REAL"
    case "boolean": return "BOOLEAN"
    case "json":    return "TEXT"
    default:        return "TEXT"    // text, richtext, media → TEXT
    }
}
```

| Content Field Type | SQL Column Type | Notes |
|---|---|---|
| `text` | `TEXT` | Standard string |
| `richtext` | `TEXT` | HTML content |
| `media` | `TEXT` | Stores media asset's `document_id` (UUID) |
| `number` | `REAL` | Float64 |
| `boolean` | `BOOLEAN` | true/false |
| `json` | `TEXT` | JSON string (serialized by `serializeFieldValue`) |
| `component` | — | **SKIPPED** — stored in separate table |

### 11.6 Layout Field Flattening
```go
func flattenLayoutFields(fields []FieldDefinition) []FieldDefinition {
    var result []FieldDefinition
    for _, field := range fields {
        if field.Type == "layout" {
            result = append(result, field.Fields...)
        } else {
            result = append(result, field)
        }
    }
    return result
}
```
- Layout fields are UI groupings — their children are the real columns
- **ALWAYS** call `flattenLayoutFields` before iterating fields for column creation
- Layout children are promoted to top-level columns
- Component fields skipped (stored in separate tables)

### 11.7 Add Missing Columns — Non-Destructive
```go
func (r *documentRepository) addMissingDocumentColumns(ctx, table string, fields []FieldDefinition) error {
    cols, err := existingColumns(r.database, table)  // get current columns
    for _, f := range flattenLayoutFields(fields) {
        if f.Type == "component" { continue }
        col := toSnakeCase(f.Name)
        if cols[col] { continue }                    // already exists → skip
        stmt := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, col, fieldColumnType(f.Type))
        r.database.WithContext(ctx).Exec(stmt)
    }
}
```
- `existingColumns()`: queries `LIMIT 1` row to get column names from `rows.Columns()`
- **Only ADD** missing columns — **NEVER** drop or alter existing columns
- Component fields skipped — they have their own tables

### 11.8 Component-Specific Missing Columns
```go
func addMissingComponentColumns(ctx, table string, fields, isNested) {
    // System columns first:
    if !cols["sort_order"] { ALTER TABLE ADD COLUMN sort_order INTEGER DEFAULT 0 }
    if isNested && !cols["parent_component_id"] { ALTER TABLE ADD COLUMN parent_component_id TEXT }
    if !isNested && !cols["document_id"] { ALTER TABLE ADD COLUMN document_id TEXT }
    // Then field columns (same as document)
}
```
- **NEVER** add both `document_id` and `parent_component_id` — they're mutually exclusive

---

## 12. Sort Key Mapping

```go
var gormSortColumn = map[string]string{
    "id":        "gorm_id",
    "createdAt": "created_at",
    "updatedAt": "updated_at",
}

func resolveGormSortClause(orderBy string, sortDir int) string {
    col := gormSortColumn[orderBy]  // fallback: "created_at"
    dir := "DESC"
    if sortDir == 1 { dir = "ASC" }
    return fmt.Sprintf("%s %s", col, dir)
}
```
- Maps API sort field names to SQL column names
- Default column: `created_at`
- Sort direction: `1` = ASC, `-1` = DESC (converted to string)

---

## 13. Dialect Detection

```go
func (r *documentRepository) isPostgres() bool {
    return r.database.Dialector.Name() == "postgres"
}
```
- Used for PK syntax differences (`SERIAL` vs `INTEGER AUTOINCREMENT`)
- **NEVER** use dialect-specific SQL in query logic — only in DDL

---

## 14. Drop Table

```go
func (r *documentRepository) DropCollection(ctx, slug string) error {
    return r.database.WithContext(ctx).Migrator().DropTable(documentTableName(slug))
}

func (r *componentRepository) DropCollection(ctx, slug, name string) error {
    return r.database.WithContext(ctx).Migrator().DropTable(componentTableName(slug, name))
}
```
- Uses GORM Migrator's `DropTable`
- Only called from sync engine when content-type JSON file is deleted
- Component tables dropped before document table (bottom-up in recursive traversal)

---

## 15. Boundaries Summary

| Rule | Detail |
|---|---|
| **Always** | Use `.WithContext(ctx)` on every GORM query |
| **Always** | Use `errors.Is(err, gorm.ErrRecordNotFound)` — NOT `==` |
| **Always** | Use parameterized queries `Where("col = ?", val)` — never interpolation |
| **Always** | Scan dynamic tables into `map[string]any` — not entity structs |
| **Always** | Call `flattenLayoutFields` before iterating for column creation |
| **Always** | Clear the opposite FK before `compToRow` (DocumentID vs ParentComponentID) |
| **Always** | Use `Take` (not `First`) for dynamic table single-row lookups |
| **Always** | Include interface compliance: `var _ repo.X = (*x)(nil)` |
| **Always** | Non-destructive `EnsureCollection`: create if missing, add columns if existing |
| **Never** | Use GORM `AutoMigrate` for dynamic tables |
| **Never** | Drop columns in `addMissingColumns` |
| **Never** | Write both `document_id` and `parent_component_id` in same `compToRow` call |
| **Never** | Use string interpolation in SQL WHERE clauses |
| **Never** | Use GORM `Clauses(OnConflict)` — use manual Take→Create/Updates pattern |
| **Never** | Hard-code SQL dialect-specific queries outside DDL |
