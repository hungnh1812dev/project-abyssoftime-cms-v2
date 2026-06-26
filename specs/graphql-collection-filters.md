# SPEC — GraphQL Collection-Type Filters

## 1. Overview

Add a `filters` argument to every collection-type `*List` GraphQL query so consumers can filter documents by field values, `documentId`, and timestamps. Filters support `eq`, `ne`, `in`, `notIn` operators, with `and` / `or` / `not` logical combinators. The top-level `filters` array is implicitly ANDed.

**Target users:** Frontend apps and third-party consumers querying collection-type documents through the GraphQL API.

**Example query:**
```graphql
query CvPage {
  cvPageList(
    filters: [
      { documentId: { eq: "A" } },
      { documentId: { eq: "B" } },
      { or: [
        { position: { eq: "C" } },
        { position: { eq: "D" } }
      ]}
    ]
  ) {
    documentId
    position
  }
}
```

---

## 2. Decisions

| Question | Decision |
|----------|----------|
| Filter operators | `eq`, `ne`, `in`, `notIn` — no comparison operators (gt/lt) in this iteration |
| Top-level array semantics | Implicit AND — each element in the `filters` array is ANDed together |
| Filterable system fields | `documentId` (IDFilter) + `createdAt`, `updatedAt`, `publishedAt` (TimeFilter) |
| Filterable content fields | `text`, `richtext` (StringFilter), `number` (NumberFilter), `boolean` (BooleanFilter) |
| Non-filterable fields | `component`, `media`, `json` — excluded from filter types |
| Logical operators | `and: [Filter!]`, `or: [Filter!]`, `not: Filter` — nestable to arbitrary depth |
| DB adapters | Both PostgreSQL (GORM) and MongoDB |
| Existing `where` argument | Replaced by `filters: [<Type>Filter!]` |
| Column name mapping | PostgreSQL uses `snake_case` columns; MongoDB uses `camelCase` BSON keys |

---

## 3. Excluded from Scope

| Item | Reason |
|------|--------|
| String matching operators (`contains`, `startsWith`) | Future iteration |
| Comparison operators (`gt`, `lt`, `gte`, `lte`) | Future iteration |
| Filtering on component sub-fields | Rule: "NEVER filter on repeatable component sub-fields in GraphQL" |
| Filtering on `media` or `json` fields | Not meaningful for equality/set ops |
| REST API filtering | Out of scope — GraphQL only |
| Locale filtering via `filters` | Already handled by the existing `locale` argument |

---

## 4. GraphQL Schema Changes

### 4.1 Base Filter Input Types (in `BuildBaseSchema`)

Add these scalar filter inputs to the base schema:

```graphql
input IDFilter {
  eq: ID
  ne: ID
  in: [ID!]
  notIn: [ID!]
}

input StringFilter {
  eq: String
  ne: String
  in: [String!]
  notIn: [String!]
}

input NumberFilter {
  eq: Float
  ne: Float
  in: [Float!]
  notIn: [Float!]
}

input BooleanFilter {
  eq: Boolean
  ne: Boolean
}

input TimeFilter {
  eq: Time
  ne: Time
}
```

### 4.2 Per-Type Filter (in `writeFilterType`)

Generated dynamically per content type. Example for a `cv-page` with fields `position: text`, `active: boolean`:

```graphql
input CvPageFilter {
  # System fields
  documentId: IDFilter
  createdAt: TimeFilter
  updatedAt: TimeFilter
  publishedAt: TimeFilter

  # Content fields (text, richtext → StringFilter; number → NumberFilter; boolean → BooleanFilter)
  position: StringFilter
  active: BooleanFilter

  # Logical operators
  and: [CvPageFilter!]
  or: [CvPageFilter!]
  not: CvPageFilter
}
```

### 4.3 List Query Argument Change

**Before:** `cvPageList(where: CvPageFilter, orderBy: CvPageOrderBy, start: Int, size: Int, locale: String)`

**After:** `cvPageList(filters: [CvPageFilter!], orderBy: CvPageOrderBy, start: Int, size: Int, locale: String)`

---

## 5. Changes — Domain Layer

### 5.1 Filter Entity

**File:** `apps/api/internal/domain/entity/filter.go` (new file)

```go
type FilterOperator string

const (
    FilterOpEq    FilterOperator = "eq"
    FilterOpNe    FilterOperator = "ne"
    FilterOpIn    FilterOperator = "in"
    FilterOpNotIn FilterOperator = "notIn"
)

type FieldFilter struct {
    Field    string
    Operator FilterOperator
    Value    any
}

type FilterNode struct {
    Field *FieldFilter
    And   []FilterNode
    Or    []FilterNode
    Not   *FilterNode
}
```

`FilterNode` is a recursive tree. A leaf node has `Field` set; a branch node has `And`, `Or`, or `Not` set. The resolver converts the GraphQL input into a `[]FilterNode` (implicit AND at the top level), then passes it through the usecase to the repository.

---

## 6. Changes — GraphQL Layer

### 6.1 Schema Builder (`schema_builder.go`)

1. **`BuildBaseSchema`** — append the `IDFilter`, `StringFilter`, `NumberFilter`, `BooleanFilter`, `TimeFilter` input type definitions.

2. **`writeFilterType`** — rewrite to:
   - Add `documentId: IDFilter` as the first field
   - Add `createdAt: TimeFilter`, `updatedAt: TimeFilter`, `publishedAt: TimeFilter`
   - Keep content field filters (`text`/`richtext` → StringFilter, `number` → NumberFilter, `boolean` → BooleanFilter)
   - Rename logical operators to lowercase: `and`, `or`, `not` (instead of current `AND`, `OR`, `NOT`)

3. **`BuildContentTypeSDL`** (collection type query) — change the list query argument from `where: <Type>Filter` to `filters: [<Type>Filter!]`.

### 6.2 Resolver Factory (`resolver_factory.go`)

1. **`addCollectionFields`** — add `filters` argument to the `*List` field config using `graphql.NewList(filterInputType)`. Parse the `filters` arg from `p.Args["filters"]` in the resolver.

2. **New function `parseFilters`** — converts `[]any` (GraphQL input) into `[]entity.FilterNode`:
   - For each map in the array, iterate keys
   - If key is `and` → recurse into `FilterNode.And`
   - If key is `or` → recurse into `FilterNode.Or`
   - If key is `not` → recurse into `FilterNode.Not`
   - Otherwise → key is a field name, value is a map of `{operator: value}` → produce `FieldFilter`

3. **Build programmatic `graphql.InputObject`** for each content type's filter (like `buildInputType` does for inputs). This is needed because the `graphql-go` library requires programmatic type construction, not just SDL strings. The SDL filter type in `schema_builder.go` serves documentation; the resolver factory builds the actual `graphql.InputObject` used at runtime.

---

## 7. Changes — UseCase Layer

### 7.1 Document UseCase

**File:** `apps/api/internal/usecase/document/document_usecase.go`

Add `filters []entity.FilterNode` parameter to:

```go
func (uc *UseCase) GetAllPaginated(
    ctx context.Context, contentTypeSlug string,
    start, size int, locale string,
    fields []entity.FieldDefinition,
    orderBy string, sortDir int,
    filters []entity.FilterNode,   // NEW
) ([]*entity.Document, []string, int64, error)

func (uc *UseCase) GetPublishedPaginated(
    ctx context.Context, contentTypeSlug string,
    start, size int, locale string,
    fields []entity.FieldDefinition,
    filters []entity.FilterNode,   // NEW
) ([]*entity.Document, int64, error)
```

Pass `filters` through to the repository. No business logic in the usecase — pure delegation.

### 7.2 DocumentUseCase Interface (in `resolver_factory.go`)

Update the `DocumentUseCase` interface to match the new signatures.

---

## 8. Changes — Repository Layer

### 8.1 DocumentRepository Interface

**File:** `apps/api/internal/domain/repository/document_repository.go`

Add `filters []entity.FilterNode` parameter to:

```go
FindDraftsByContentTypePaginated(ctx context.Context, contentTypeSlug string, start, size int, locale, orderBy string, sortDir int, filters []entity.FilterNode) ([]*entity.Document, int64, error)
FindPublishedByContentTypePaginated(ctx context.Context, contentTypeSlug string, start, size int, locale, orderBy string, sortDir int, filters []entity.FilterNode) ([]*entity.Document, int64, error)
```

### 8.2 PostgreSQL (GORM) Implementation

**File:** `apps/api/internal/infrastructure/gormdb/document_repository.go`

New helper function `applyFilters(query *gorm.DB, filters []entity.FilterNode) *gorm.DB`:

- Wraps the top-level `[]FilterNode` as an implicit AND
- For each `FilterNode`:
  - If `Field` is set → apply single field condition
  - If `And` → recursive AND group: `query.Where(subQuery)` for each child
  - If `Or` → build OR group: combine children with `db.Or(...)` 
  - If `Not` → wrap child in `NOT (...)`

**Column name mapping** (field name → SQL column):
- `documentId` → `document_id`
- `createdAt` → `created_at`
- `updatedAt` → `updated_at`
- `publishedAt` → `published_at`
- Content fields → use field name as-is (columns use the field name directly)

**Operator mapping:**
| FilterOperator | SQL |
|---|---|
| `eq` | `column = ?` |
| `ne` | `column != ?` |
| `in` | `column IN (?)` |
| `notIn` | `column NOT IN (?)` |

**SQL injection prevention:** Never interpolate field names directly. Validate that the field name matches a known content field or system field before using it in a query. Use a whitelist approach — reject unknown field names.

### 8.3 MongoDB Implementation

**File:** `apps/api/internal/infrastructure/mongodb/document_repository.go`

New helper function `buildMongoFilter(filters []entity.FilterNode) bson.M`:

- Converts `[]FilterNode` into a `bson.M` with `$and` at the top level
- For each `FilterNode`:
  - If `Field` is set → `bson.M{fieldName: bson.M{operator: value}}`
  - If `And` → `$and: [...]`
  - If `Or` → `$or: [...]`
  - If `Not` → wrap in `$nor: [...]` (single element)

**Field name mapping** (field name → BSON key):
- `documentId` → `documentId` (MongoDB uses camelCase)
- `createdAt` → `createdAt`
- Content fields → `data.<fieldName>` (MongoDB stores content in a nested `data` field)

**Operator mapping:**
| FilterOperator | MongoDB |
|---|---|
| `eq` | `$eq` |
| `ne` | `$ne` |
| `in` | `$in` |
| `notIn` | `$nin` |

### 8.4 Mock Repository

**File:** `apps/api/internal/domain/repository/mock/document_repository.go`

Update `FindDraftsByContentTypePaginatedFn` and `FindPublishedByContentTypePaginatedFn` function types to include the `filters` parameter.

---

## 9. Changes — REST Callers

The REST handler calls `GetAllPaginated` and `GetPublishedPaginated`. These callers must pass `nil` for the new `filters` parameter (REST does not support filtering in this iteration).

**Files to update:**
- `apps/api/internal/delivery/http/handler/document_handler.go` — pass `nil` for filters
- Any other callers of these usecase methods

---

## 10. Testing Strategy

### 10.1 Unit Tests — Filter Parsing (`resolver_factory_test.go` or new file)

- Parse empty filters → `nil` slice
- Parse single field filter `{documentId: {eq: "A"}}` → correct `FilterNode`
- Parse multiple filters (implicit AND) → `[]FilterNode` with 2+ elements
- Parse `or` combinator → `FilterNode` with `Or` populated
- Parse nested `and` inside `or` → correct tree structure
- Parse `not` → `FilterNode` with `Not` set
- Unknown field → ignored or error (decide: ignore silently)
- Invalid operator → error

### 10.2 Unit Tests — GORM Filter Builder

- Single `eq` filter → `WHERE column = ?`
- `ne` filter → `WHERE column != ?`
- `in` filter → `WHERE column IN (?)`
- `notIn` filter → `WHERE column NOT IN (?)`
- OR combinator → `WHERE (a = ? OR b = ?)`
- AND combinator → `WHERE (a = ? AND b = ?)`
- NOT combinator → `WHERE NOT (a = ?)`
- Nested `or` inside `and` → correct SQL grouping
- `nil` filters → no additional WHERE clauses
- System field mapping: `documentId` → `document_id`

### 10.3 Unit Tests — MongoDB Filter Builder

- Single `eq` filter → `bson.M{field: bson.M{"$eq": value}}`
- OR combinator → `$or: [...]`
- Content field mapping → `data.<fieldName>`
- `nil` filters → empty/no additional filter

### 10.4 Integration Tests (if applicable)

- GraphQL query with `filters: [{documentId: {eq: "known-id"}}]` → returns only that document
- GraphQL query with `filters: [{or: [...]}]` → returns union of matches
- GraphQL query with no filters → returns all (existing behavior preserved)

---

## 11. Boundaries

| Rule | Detail |
|------|--------|
| **Always** | Validate filter field names against known content fields + system fields |
| **Always** | Use parameterized queries — never interpolate user values into SQL |
| **Always** | Map field names to correct column/key names per adapter |
| **Always** | Pass `nil` filters from REST callers — no behavior change for REST |
| **Always** | Preserve existing behavior when `filters` argument is omitted |
| **Never** | Allow filtering on `component`, `media`, or `json` fields |
| **Never** | Allow filtering on fields not defined in the content type |
| **Never** | Build SQL with string concatenation of user-provided values |
| **Ask first** | Adding new operators beyond `eq`, `ne`, `in`, `notIn` |
| **Ask first** | Adding filter support to the REST API |
