# RULES — GraphQL

**Scope:** Build-time GraphQL schema generation (gqlgen), resolvers, filters, ordering, codegen pipeline.

---

## 1. GraphQL Rules

### 1.1 Build-Time Schema (gqlgen)
- Schema defined in `.graphql` files generated at build time by `cmd/gqlcodegen`
- Library: `github.com/99designs/gqlgen` (replaces `graphql-go/graphql`)
- **Must run `make graphql-generate` after changing `content-types/*.json`**
- Generated files (`.gitignored`): `graphql/schema/`, `graphql/generated/`, `graphql/model/models_gen.go`, `graphql/resolver/content_gen.go`
- Hand-written files: `graphql/resolver/resolver.go`, `document_helpers.go`, `media.go`, `filter.go`, `content_types.go`, `graphql/handler.go`

### 1.2 Schema Shape
- Collection type → `Query.<slug>(Id: ID!, locale: String, status: String)` + `Query.<slugList>(pagination: PaginationInput, filters, orderBy, locale, status): <Type>List!`
- `<Type>List` wrapper: `{ items: [<Type>!]!, meta: ListMeta! }`
- `ListMeta`: `{ pagination: PaginationMeta! }`
- `PaginationMeta`: `{ page: Int!, pageSize: Int!, total: Int! }`
- `PaginationInput`: `{ start, limit, page, pageSize }` — two modes, validated at resolver level
- Single type → `Query.<slug>(locale: String, status: String)`
- Queries default to published; `status: "draft"` opt-in for authenticated users
- All mutations have `@auth` directive (handler-level auth enforced)
- Base filter input types: `IDFilter`, `StringFilter`, `NumberFilter`, `BooleanFilter`, `TimeFilter`
- Per-type `<Type>Filter` includes: `documentId` (IDFilter), `createdAt`/`updatedAt`/`publishedAt` (TimeFilter), content fields, `and`/`or`/`not` combinators
- `filters` argument is an array — items are implicitly ANDed
- Supported operators: `eq`, `ne`, `in`, `notIn`
- `orderBy` uses typed `<Type>OrderBy` struct (first non-nil field determines sort)

### 1.3 Field Type Mapping
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

### 1.4 Naming Conventions
- Type: PascalCase of slug (`blog-posts` → `BlogPosts`)
- Input: `<Type>Input`
- Filter: `<Type>Filter`
- Base filters: `IDFilter`, `StringFilter`, `NumberFilter`, `BooleanFilter`, `TimeFilter`
- OrderBy: `<Type>OrderBy`
- Component: `<ContentType><ComponentName>` (e.g., `CvPageSkills`)
- Query single: camelCase (`cvPage`)
- Query list: camelCase + `List` (`cvPageList`)

### 1.5 Resolvers
- All delegate to generic helpers in `document_helpers.go` — **NO** business logic in resolvers
- `Resolver` struct takes `DocumentUseCase`, `ContentTypeUseCase`, `MediaAssetRepository`
- Generated resolvers in `content_gen.go` are thin wrappers calling generic helpers
- Media fields resolved into full `MediaAsset` objects recursively via `media.go`
- Filter conversion: gqlgen typed structs → `entity.FilterNode` via reflection (`filter.go`)
- **NEVER** duplicate filtering on repeatable component sub-fields (defer to future)
- **NEVER** add business logic to generated or hand-written resolvers

### 1.6 Codegen Pipeline (`make graphql-generate`)
1. `gqlcodegen --phase=schema` → generates `.graphql` files + updates `gqlgen.yml` models
2. `gqlgen generate` → generates `generated.go` + `models_gen.go`
3. Remove gqlgen resolver stubs (`*.resolvers.go`)
4. `gqlcodegen --phase=resolvers` → generates `content_gen.go`
- **NEVER** edit generated files — they are overwritten on every run
- **NEVER** commit generated files to git

---

## 2. Boundaries Summary

| Rule | Detail |
|---|---|
| **Always** | Route GraphQL through the same usecase — no logic in resolvers |
| **Always** | Validate filter field names against known content fields + system fields |
| **Always** | Use parameterized queries for all filter conditions |
| **Always** | Map filter field names to correct column/key names per DB adapter |
| **Never** | Filter on repeatable component sub-fields in GraphQL |
| **Never** | Allow filtering on `component`, `media`, or `json` fields in GraphQL |
| **Never** | Build SQL with string concatenation of user-provided filter values |
| **Ask first** | Adding new filter operators beyond `eq`, `ne`, `in`, `notIn` |
