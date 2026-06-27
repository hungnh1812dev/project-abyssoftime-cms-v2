# SPEC: Migrate GraphQL to gqlgen (Full Codegen Pipeline)

**Status:** Draft — awaiting user review
**Module:** content (GraphQL subsystem)
**Related rules:** `rules/content.md §6`, `rules/GLOBAL.md §1.3`

---

## 1. Objective

Replace `graphql-go/graphql` (dormant since 2023) with **gqlgen** — the most popular, fastest, and actively maintained Go GraphQL library.

Since gqlgen requires schemas at build time, introduce a **full codegen pipeline**: a custom tool reads `content-types/*.json` and generates both `.graphql` schema files and resolver implementations. Then gqlgen generates the execution runtime. One `make graphql-generate` command does everything.

**Target users:** API consumers (frontend app, third-party integrations via access tokens).

---

## 2. Architecture Change

### 2.1 Before (Current)

```
Startup:
  content-types/*.json → schema_loader → sync DB
                        → schema_builder (SDL, unused)
                        → resolver_factory (programmatic types + closures) → serve
```

- Schema built **at runtime** using `graphql-go/graphql` programmatic API
- `schema_builder.go` generates SDL but it's not used for execution
- `resolver_factory.go` constructs types, inputs, filters, and closures manually (~817 lines)

### 2.2 After (gqlgen)

```
Build time (make graphql-generate):
  content-types/*.json → [gqlcodegen tool] → graphql/schema/*.graphql
                                            → graphql/resolver/content_gen.go
                                            → graphql/gqlgen.yml (models section)
                       → [gqlgen generate]  → graphql/generated/generated.go
                                            → graphql/model/models_gen.go

Startup:
  content-types/*.json → schema_loader → sync DB (unchanged)
  Pre-compiled gqlgen schema → serve
```

- Schema defined in `.graphql` files (generated from JSON)
- Resolver implementations auto-generated (all delegate to usecase)
- gqlgen generates execution runtime + model types
- Startup sync still runs (DB schema management unchanged)

### 2.3 What Changes

| Aspect | Before | After |
|--------|--------|-------|
| Schema source | Programmatic Go types at runtime | `.graphql` files at build time |
| Resolver style | Closures in `resolver_factory.go` | Struct methods (generated + hand-written) |
| Library | `graphql-go/graphql` v0.8.1 (dormant) | `gqlgen` (active, ~10.7k stars) |
| Performance | ~19k req/sec | ~52.8k req/sec (2.7x faster) |
| Adding content type | Add JSON → restart | Add JSON → `make graphql-generate` → rebuild |
| Type safety | `map[string]any` everywhere | Generated filter/input structs; document maps |

### 2.4 What Does NOT Change

- `content-types/*.json` format — unchanged
- Startup DB sync (`schema_loader.go`, `sync.go`) — unchanged
- `DocumentUseCase` interface — unchanged
- REST API — unchanged
- Filter domain model (`entity.FilterNode`) — unchanged
- Media resolution logic — same approach, moved to `resolver/media.go`
- JWT/access-token validation — HTTP middleware unchanged; mutation auth via `@auth` directive

---

## 3. Code Elimination: What gqlgen Replaces

### 3.1 Line-by-Line Breakdown of `resolver_factory.go` (817 lines)

| Lines | Code | Status | gqlgen Feature |
|-------|------|--------|----------------|
| 22-44 | `DocumentUseCase`, `ContentTypeUseCase`, `AccessTokenValidator` interfaces | **KEEP** — move to `resolver/` package | — |
| 46-55 | `ResolverFactory` struct + constructor | **REMOVE** | gqlgen `Resolver` struct pattern |
| 57-69 | JSON scalar definition (serialize/parse/parseLiteral) | **REMOVE** | `gqlgen/graphql.Map` built-in scalar |
| 70 | Time scalar (`graphql.DateTime`) | **REMOVE** | `gqlgen/graphql.Time` built-in scalar |
| 72-118 | Filter input type construction (IDFilter, StringFilter, NumberFilter, BooleanFilter, TimeFilter + mapping) | **REMOVE** | gqlgen generates Go structs from `.graphql` schema |
| 121-131 | MediaAsset object type construction | **REMOVE** | gqlgen generates from schema |
| 133-143 | ContentType object type construction | **REMOVE** | gqlgen generates from schema |
| 145-163 | Query fields map + contentTypes resolver | **REMOVE** | gqlgen generates; resolver in `content_types.go` |
| 165-198 | Loop: build object/input/filter types per content type + add fields | **REMOVE** | gqlgen generates types; resolvers in `content_gen.go` |
| 200-209 | `graphql.NewSchema()` creation | **REMOVE** | `generated.NewExecutableSchema()` |
| 211-236 | HTTP handler + inline JWT/access-token auth | **REWRITE** | gqlgen `handler.NewDefaultServer()` + `@auth` directive |
| 238-259 | `buildComponentType()` — recursive component type construction | **REMOVE** | gqlgen generates from `.graphql` nested types |
| 261-290 | `buildObjectType()` — per-content-type output type | **REMOVE** | gqlgen generates from schema |
| 292-304 | `buildInputType()` — per-content-type input type | **REMOVE** | gqlgen generates from schema |
| 306-467 | `addCollectionFields()` — 7 query/mutation fields with closure resolvers | **REMOVE** | Generated thin resolvers in `content_gen.go` |
| 469-560 | `addSingleFields()` — 4 query/mutation fields with closure resolvers | **REMOVE** | Generated thin resolvers in `content_gen.go` |
| 562-569 | `authRequired()` wrapper function | **REMOVE** | gqlgen `@auth` directive (one implementation) |
| 571-593 | `docToMap()` — document to response map | **KEEP** — move to `resolver/media.go` | — |
| 595-628 | `resolveMediaField()` — media ID → MediaAsset lookup | **KEEP** — move to `resolver/media.go` | — |
| 630-660 | `resolveComponentMedia()` + `resolveComponentMap()` | **KEEP** — move to `resolver/media.go` | — |
| 662-672 | `inputToMap()` — manual input deserialization | **REMOVE** | gqlgen auto-deserializes inputs into typed structs |
| 674-707 | `buildFilterInputType()` — recursive filter type construction | **REMOVE** | gqlgen generates from schema |
| 709-798 | `parseFilters()` + `parseFilterMap()` + `parseFilterList()` | **REWRITE** | New `convertFilterStructs[T]()` with typed gqlgen structs |
| 800-815 | `gqlScalarFor()` — field type → GraphQL type mapping | **REMOVE** | gqlgen handles type mapping from schema |

### 3.2 Line-by-Line Breakdown of `schema_builder.go` (272 lines)

| Lines | Code | Status | Replacement |
|-------|------|--------|-------------|
| 11-17 | `SchemaBuilder` struct + constructor | **REMOVE** from runtime | Logic moves to `cmd/gqlcodegen` (build tool) |
| 19-81 | `BuildBaseSchema()` — base SDL (scalars, enums, filters, types) | **REMOVE** from runtime | `gqlcodegen` generates `base.graphql` |
| 83-164 | `BuildContentTypeSDL()` — per-type SDL generation | **REMOVE** from runtime | `gqlcodegen` generates `<slug>.graphql` |
| 166-189 | `writeComponentType()` — recursive component SDL | **REMOVE** from runtime | `gqlcodegen` handles nested types |
| 192-226 | `writeFilterType()` + `writeOrderByType()` | **REMOVE** from runtime | `gqlcodegen` generates filter/orderBy inputs |
| 228-236 | `BuildSDL()` — concatenate all SDL | **REMOVE** from runtime | `gqlcodegen` writes individual files |
| 238-271 | `slugToPascalCase()`, `slugToCamelCase()`, `fieldTypeToGraphQL()` | **REMOVE** from runtime | Move to `cmd/gqlcodegen` (build tool only) |

### 3.3 Summary: Code Impact

| Category | Current Lines | After Migration | Change |
|----------|--------------|-----------------|--------|
| **REMOVED** (gqlgen handles natively) | ~580 | 0 | -580 |
| **REWRITTEN** (auth directive, filter conversion) | ~115 | ~60 | -55 |
| **KEPT** (media resolution, docToMap, interfaces) | ~115 | ~115 | 0 |
| **NEW** hand-written (resolver, helpers, handler) | 0 | ~170 | +170 |
| | | | |
| **Total runtime code** | **~1,089** (817+272) | **~345** | **-68%** |
| **Codegen tool** (build-time only) | 0 | ~350 | +350 (not runtime) |

### 3.4 Specific gqlgen Features That Replace Manual Code

#### 3.4.1 `@auth` Directive — Replaces `authRequired()` Wrapper

**Before** (manual wrapper per mutation):
```go
mf["create"+pascal] = &graphql.Field{
    Resolve: f.authRequired(func(p graphql.ResolveParams) (any, error) {
        // ... mutation logic
    }),
}
```

**After** (single directive implementation):

Schema:
```graphql
directive @auth on FIELD_DEFINITION

extend type Mutation {
  createCvPage(data: CvPageInput!): CvPage! @auth
  updateCvPage(cvPageId: ID!, data: CvPageInput!): CvPage! @auth
  deleteCvPage(cvPageId: ID!): Boolean! @auth
  publishCvPage(cvPageId: ID!, locale: String): CvPage! @auth
  unpublishCvPage(cvPageId: ID!, locale: String): CvPage! @auth
}
```

Implementation (once, in `handler.go`):
```go
cfg := generated.Config{Resolvers: resolver}
cfg.Directives.Auth = func(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
    if middleware.UserID(ctx) == "" {
        return nil, fmt.Errorf("unauthorized")
    }
    return next(ctx)
}
srv := handler.NewDefaultServer(generated.NewExecutableSchema(cfg))
```

**Impact:** Eliminates per-mutation auth wrapping. Auth is declarative in the schema. The `gqlcodegen` tool adds `@auth` to all mutation fields automatically.

#### 3.4.2 Automatic Input Deserialization — Replaces `inputToMap()`

**Before:**
```go
func inputToMap(v any) map[string]any {
    switch val := v.(type) {
    case map[string]any:
        return val
    default:
        b, _ := json.Marshal(v)
        var m map[string]any
        _ = json.Unmarshal(b, &m)
        return m
    }
}
```

**After:** gqlgen automatically deserializes GraphQL input into the generated `model.<Type>Input` struct. The resolver receives a typed struct directly:

```go
func (r *mutationResolver) CreateCvPage(ctx context.Context, data model.CvPageInput) (map[string]interface{}, error) {
    // data is already a typed struct — no manual deserialization
    return r.createDocument(ctx, "cv-page", data, cvPageFields)
}
```

The `createDocument` helper converts the typed struct to `map[string]any` for the usecase using `structToMap()`.

#### 3.4.3 Generated Type System — Replaces All Manual Type Construction

**Before** (~300 lines per content type):
```go
objType := graphql.NewObject(graphql.ObjectConfig{
    Name: typeName,
    Fields: graphql.Fields{
        "documentId":  &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
        "locale":      &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
        // ... per field
    },
})
inputType := graphql.NewInputObject(graphql.InputObjectConfig{ /* ... */ })
filterType := buildFilterInputType(/* ... */)
```

**After:** gqlgen reads `.graphql` schema files and generates all Go types + execution code automatically. Zero manual type construction at runtime.

#### 3.4.4 gqlgen Error Presenter — Cleaner Error Handling

**Before:** Manual error wrapping:
```go
return nil, fmt.Errorf("unauthorized")
```

**After:** gqlgen's error presenter provides structured GraphQL errors with extensions:
```go
srv.SetErrorPresenter(func(ctx context.Context, err error) *gqlerror.Error {
    // Map domain errors to GraphQL error codes
    // Add extensions for error classification
})
```

#### 3.4.5 Map Field Resolution — Eliminates Manual Field Mapping

gqlgen natively resolves `map[string]interface{}` fields by key name. A document map like:
```go
map[string]interface{}{
    "documentId": "abc-123",
    "title":      "Hello",
    "createdAt":  time.Now(),
}
```
automatically satisfies GraphQL fields `documentId`, `title`, `createdAt` without any manual field extraction code.

---

## 4. Why gqlgen Despite the Build-Step Tradeoff

### 4.1 The Tradeoff

Content type changes now require `make graphql-generate` + rebuild instead of just restarting. This is acceptable because:

1. **Content types are code** — defined in JSON files checked into the repo. Adding one already requires a commit + deploy.
2. **Same pattern as protobuf** — `.proto` → `protoc` → generated Go. The Go ecosystem normalizes build-time code generation.
3. **The Makefile target already exists** — `graphql-generate` is pre-planned in the project's `Makefile`.

### 4.2 Why Not graph-gophers/graphql-go

| Concern | Detail |
|---------|--------|
| **4.7k stars** | Minimalist by design; smaller market. No code-gen = less visibility. |
| **No public adopters** | Zero named companies using it in production (vs gqlgen: Sky Bet, LitmusChaos, etc.) |
| **17% slower than gqlgen** | 44.3k vs 52.8k req/sec in HTTP benchmarks |
| **Missing enterprise features** | No federation, no query complexity, no batch querying, no plugin system |
| **Resolver binding problem** | Method names must match SDL fields at compile time — same dynamic-schema problem, just at a different layer |
| **Smaller ecosystem** | Fewer tutorials, fewer StackOverflow answers, fewer community tools |

### 4.3 Why gqlgen Wins

| Advantage | Impact |
|-----------|--------|
| **Most popular** (10.7k stars) | Largest community, most resources, best hiring signal |
| **Fastest** (52.8k req/sec) | 2.7x faster than current library |
| **Active maintenance** | Regular releases through 2026 |
| **Type-safe generated code** | Compile-time errors instead of runtime panics |
| **Federation-ready** | Apollo Federation v1 & v2 if needed later |
| **Subscription support** | WebSocket-based, channel-driven |
| **Directive support** | Custom schema directives for auth, caching, validation |
| **Query complexity analysis** | Built-in protection against expensive queries |
| **Plugin system** | Extensible code generation pipeline |
| **Standard tooling** | Works with GraphQL Playground, Apollo Studio, Altair |

---

## 5. Code Generation Pipeline

### 5.1 Custom Tool: `gqlcodegen`

**Location:** `apps/api/cmd/gqlcodegen/main.go`

**Input:** Reads `content-types/*.json` using the existing `contenttype.LoadDefinitions()` function.

**Output (3 artifacts):**

#### Artifact 1: `.graphql` Schema Files

`graphql/schema/base.graphql` — shared types:
```graphql
scalar JSON
scalar Time

enum SortOrder {
  ASC
  DESC
}

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

type MediaAsset {
  documentId: ID!
  url: String!
  thumbnailUrl: String
  fileName: String
  width: Int
  height: Int
}

type ContentType {
  id: ID!
  name: String!
  slug: String!
  kind: String!
  createdAt: Time!
  updatedAt: Time!
}

type Query {
  contentTypes: [ContentType!]!
}

type Mutation {
  _empty: Boolean
}
```

`graphql/schema/cv-page.graphql` — per content type (example):
```graphql
type CvPageSkills {
  level: String
  skill: String
}

type CvPageExperiencesRoles {
  position: String
  period: String
  teamSize: Float
  projects: String
  techStack: JSON
  responsibilities: String
}

type CvPageExperiences {
  company: String
  location: String
  roles: [CvPageExperiencesRoles!]
}

type CvPage {
  documentId: ID!
  position: String
  isMain: Boolean
  company: String
  summary: String
  skills: [CvPageSkills!]
  experiences: [CvPageExperiences!]
  projects: [CvPageProjects!]
  educations: [CvPageEducations!]
  languages: [CvPageLanguages!]
  references: [CvPageReferences!]
  locale: String!
  createdAt: Time!
  updatedAt: Time!
  publishedAt: Time
}

input CvPageInput {
  position: String
  isMain: Boolean
  company: String
  summary: String
}

input CvPageFilter {
  documentId: IDFilter
  position: StringFilter
  isMain: BooleanFilter
  company: StringFilter
  summary: StringFilter
  createdAt: TimeFilter
  updatedAt: TimeFilter
  publishedAt: TimeFilter
  and: [CvPageFilter!]
  or: [CvPageFilter!]
  not: CvPageFilter
}

input CvPageOrderBy {
  position: SortOrder
  isMain: SortOrder
  company: SortOrder
  summary: SortOrder
  createdAt: SortOrder
  updatedAt: SortOrder
  publishedAt: SortOrder
}

extend type Query {
  cvPage(cvPageId: ID!, locale: String, status: String): CvPage
  cvPageList(
    filters: [CvPageFilter!]
    orderBy: CvPageOrderBy
    start: Int
    size: Int
    locale: String
    status: String
  ): [CvPage!]!
}

extend type Mutation {
  createCvPage(data: CvPageInput!): CvPage!
  updateCvPage(cvPageId: ID!, data: CvPageInput!): CvPage!
  deleteCvPage(cvPageId: ID!): Boolean!
  publishCvPage(cvPageId: ID!, locale: String): CvPage!
  unpublishCvPage(cvPageId: ID!, locale: String): CvPage!
}
```

`graphql/schema/cv-contact.graphql` — single type (example):
```graphql
type CvContact {
  documentId: ID!
  name: String
  address: String
  phone: String
  email: String
  linkedin: String
  github: String
  avatar: MediaAsset
  locale: String!
  createdAt: Time!
  updatedAt: Time!
  publishedAt: Time
}

input CvContactInput {
  name: String
  address: String
  phone: String
  email: String
  linkedin: String
  github: String
  avatar: String
}

extend type Query {
  cvContact(locale: String, status: String): CvContact
}

extend type Mutation {
  saveCvContact(data: CvContactInput!, locale: String): CvContact!
  publishCvContact(locale: String): CvContact!
  unpublishCvContact(locale: String): CvContact!
}
```

#### Artifact 2: Generated Resolver Implementations

`graphql/resolver/content_gen.go` — all content-type resolvers:

```go
// Code generated by gqlcodegen from content-types/*.json. DO NOT EDIT.
package resolver

import (
    "context"
    "project-abyssoftime-cms-v2/api/internal/domain/entity"
)

// ── Field definitions (embedded from JSON at codegen time) ──

var cvPageFields = []entity.FieldDefinition{
    {Name: "position", Type: "text"},
    {Name: "isMain", Type: "boolean"},
    {Name: "company", Type: "text"},
    {Name: "summary", Type: "richtext"},
    {Name: "skills", Type: "component", Repeatable: true, Fields: []entity.FieldDefinition{
        {Name: "level", Type: "text"},
        {Name: "skill", Type: "text"},
    }},
    // ... full field tree
}

// ── Collection-type resolvers: cv-page ──

func (r *queryResolver) CvPage(ctx context.Context, cvPageID string, locale *string, status *string) (map[string]interface{}, error) {
    return r.getDocument(ctx, "cv-page", cvPageID, locale, status, cvPageFields)
}

func (r *queryResolver) CvPageList(ctx context.Context, filters []*model.CvPageFilter, orderBy *model.CvPageOrderBy, start *int, size *int, locale *string, status *string) ([]map[string]interface{}, error) {
    return r.getDocumentList(ctx, "cv-page", filters, orderBy, start, size, locale, status, cvPageFields)
}

func (r *mutationResolver) CreateCvPage(ctx context.Context, data model.CvPageInput) (map[string]interface{}, error) {
    return r.createDocument(ctx, "cv-page", data, cvPageFields)
}

func (r *mutationResolver) UpdateCvPage(ctx context.Context, cvPageID string, data model.CvPageInput) (map[string]interface{}, error) {
    return r.updateDocument(ctx, "cv-page", cvPageID, data, cvPageFields)
}

func (r *mutationResolver) DeleteCvPage(ctx context.Context, cvPageID string) (bool, error) {
    return r.deleteDocument(ctx, "cv-page", cvPageID, cvPageFields)
}

func (r *mutationResolver) PublishCvPage(ctx context.Context, cvPageID string, locale *string) (map[string]interface{}, error) {
    return r.publishDocument(ctx, "cv-page", cvPageID, locale, cvPageFields)
}

func (r *mutationResolver) UnpublishCvPage(ctx context.Context, cvPageID string, locale *string) (map[string]interface{}, error) {
    return r.unpublishDocument(ctx, "cv-page", cvPageID, locale, cvPageFields)
}

// ── Single-type resolvers: cv-contact ──

func (r *queryResolver) CvContact(ctx context.Context, locale *string, status *string) (map[string]interface{}, error) {
    return r.getSingleType(ctx, "cv-contact", locale, status, cvContactFields)
}

func (r *mutationResolver) SaveCvContact(ctx context.Context, data model.CvContactInput, locale *string) (map[string]interface{}, error) {
    return r.saveSingleType(ctx, "cv-contact", data, locale, cvContactFields)
}

func (r *mutationResolver) PublishCvContact(ctx context.Context, locale *string) (map[string]interface{}, error) {
    return r.publishSingleType(ctx, "cv-contact", locale, cvContactFields)
}

func (r *mutationResolver) UnpublishCvContact(ctx context.Context, locale *string) (map[string]interface{}, error) {
    return r.unpublishSingleType(ctx, "cv-contact", locale, cvContactFields)
}
```

#### Artifact 3: gqlgen Model Mappings

The tool appends to `graphql/gqlgen.yml` models section, mapping all content-type output types to `map[string]interface{}`:

```yaml
models:
  # Content-type document models → map (fields are dynamic)
  CvPage:
    model: "project-abyssoftime-cms-v2/api/graphql/model.DocumentMap"
  CvContact:
    model: "project-abyssoftime-cms-v2/api/graphql/model.DocumentMap"
  EnVocabPack:
    model: "project-abyssoftime-cms-v2/api/graphql/model.DocumentMap"
  CommonText:
    model: "project-abyssoftime-cms-v2/api/graphql/model.DocumentMap"
  # Component types → map
  CvPageSkills:
    model: "project-abyssoftime-cms-v2/api/graphql/model.DocumentMap"
  CvPageExperiences:
    model: "project-abyssoftime-cms-v2/api/graphql/model.DocumentMap"
  CvPageExperiencesRoles:
    model: "project-abyssoftime-cms-v2/api/graphql/model.DocumentMap"
  # ... all component types
  # Static types → existing entities
  MediaAsset:
    model: "project-abyssoftime-cms-v2/api/graphql/model.MediaAssetMap"
  ContentType:
    model: "project-abyssoftime-cms-v2/api/graphql/model.ContentTypeMap"
```

### 5.2 gqlgen Configuration

`graphql/gqlgen.yml`:
```yaml
schema:
  - schema/*.graphql

exec:
  filename: generated/generated.go
  package: generated

model:
  filename: model/models_gen.go
  package: model

resolver:
  layout: follow-schema
  dir: resolver
  package: resolver
  filename_template: "{name}.resolvers.go"

autobind: []

models:
  JSON:
    model: "github.com/99designs/gqlgen/graphql.Map"
  Time:
    model: "github.com/99designs/gqlgen/graphql.Time"
  # Content-type models injected by gqlcodegen (see §5.1 Artifact 3)
```

### 5.3 Build Command

```makefile
graphql-generate:
	cd apps/api && go run ./cmd/gqlcodegen
	cd apps/api && go run github.com/99designs/gqlgen generate
```

Single command: `make graphql-generate`

---

## 6. File Structure (Post-Migration)

```
apps/api/
├── cmd/
│   ├── server/main.go              # Startup (updated: load gqlgen schema)
│   └── gqlcodegen/main.go          # Custom codegen tool (NEW)
│
├── graphql/                         # NEW — replaces graphql/dynamic/
│   ├── schema/                      # Generated .graphql files (DO NOT EDIT)
│   │   ├── base.graphql
│   │   ├── cv-page.graphql
│   │   ├── cv-contact.graphql
│   │   ├── common-text.graphql
│   │   └── en-vocab-pack.graphql
│   │
│   ├── generated/                   # gqlgen output (DO NOT EDIT)
│   │   └── generated.go
│   │
│   ├── model/                       # Model types
│   │   ├── models_gen.go            # gqlgen generated (DO NOT EDIT)
│   │   └── types.go                 # DocumentMap, MediaAssetMap aliases (hand-written)
│   │
│   ├── resolver/                    # Resolver implementations
│   │   ├── resolver.go              # Root resolver struct + constructor (hand-written)
│   │   ├── content_gen.go           # Content-type resolvers (generated by gqlcodegen)
│   │   ├── content_types.go         # contentTypes query resolver (hand-written)
│   │   ├── document_helpers.go      # Generic CRUD helpers: getDocument, createDocument, etc. (hand-written)
│   │   ├── media.go                 # Media field resolution (hand-written, adapted from current)
│   │   └── filter.go               # Filter conversion: gqlgen model → FilterNode (hand-written)
│   │
│   ├── handler.go                   # HTTP handler setup + auth middleware (hand-written)
│   └── gqlgen.yml                   # gqlgen configuration
│
├── graphql/dynamic/                 # REMOVED (replaced by graphql/)
│   ├── schema_builder.go           # REMOVED (replaced by gqlcodegen)
│   ├── resolver_factory.go         # REMOVED (replaced by resolver/)
│   └── *_test.go                   # REMOVED (replaced by new tests)
│
└── content-types/*.json             # Unchanged
```

### 6.1 Generated vs Hand-Written

| File | Source | Editable? |
|------|--------|-----------|
| `schema/*.graphql` | gqlcodegen | No — regenerated from JSON |
| `generated/generated.go` | gqlgen | No — regenerated from .graphql |
| `model/models_gen.go` | gqlgen | No — regenerated |
| `resolver/content_gen.go` | gqlcodegen | No — regenerated from JSON |
| `model/types.go` | Hand-written | Yes |
| `resolver/resolver.go` | Hand-written | Yes |
| `resolver/document_helpers.go` | Hand-written | Yes |
| `resolver/content_types.go` | Hand-written | Yes |
| `resolver/media.go` | Hand-written | Yes |
| `resolver/filter.go` | Hand-written | Yes |
| `handler.go` | Hand-written | Yes |
| `gqlgen.yml` | Partially generated | Models section regenerated by gqlcodegen |

### 6.2 .gitignore

Generated files are **.gitignored** — codegen runs at build time on Render.com (see §14.3). Add `DO NOT EDIT` headers to all generated files for clarity.

```gitignore
# gqlgen generated files (regenerated at build time)
apps/api/graphql/schema/
apps/api/graphql/generated/
apps/api/graphql/model/models_gen.go
apps/api/graphql/resolver/content_gen.go
```

---

## 7. Implementation Details

### 7.1 Model Types (`model/types.go`)

```go
package model

type DocumentMap = map[string]interface{}
type MediaAssetMap = map[string]interface{}
type ContentTypeMap = map[string]interface{}
```

gqlgen resolves fields from maps automatically by key name. A `DocumentMap` with key `"documentId"` satisfies the `documentId: ID!` field in the schema.

### 7.2 Root Resolver (`resolver/resolver.go`)

```go
package resolver

import (
    "project-abyssoftime-cms-v2/api/graphql/generated"
    "project-abyssoftime-cms-v2/api/internal/domain/entity"
    "project-abyssoftime-cms-v2/api/internal/domain/repository"
)

type Resolver struct {
    docUC     DocumentUseCase
    ctUC      ContentTypeUseCase
    mediaRepo repository.MediaAssetRepository
}

func NewResolver(docUC DocumentUseCase, ctUC ContentTypeUseCase, mediaRepo repository.MediaAssetRepository) *Resolver {
    return &Resolver{docUC: docUC, ctUC: ctUC, mediaRepo: mediaRepo}
}

func (r *Resolver) Query() generated.QueryResolver       { return &queryResolver{r} }
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

type queryResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
```

### 7.3 Generic Document Helpers (`resolver/document_helpers.go`)

These are the shared methods that all generated resolvers delegate to:

```go
package resolver

// Collection-type helpers
func (r *queryResolver) getDocument(ctx, slug, docID string, locale, status *string, fields []entity.FieldDefinition) (map[string]interface{}, error)
func (r *queryResolver) getDocumentList(ctx, slug string, filters any, orderBy any, start, size *int, locale, status *string, fields []entity.FieldDefinition) ([]map[string]interface{}, error)
func (r *mutationResolver) createDocument(ctx, slug string, data any, fields []entity.FieldDefinition) (map[string]interface{}, error)
func (r *mutationResolver) updateDocument(ctx, slug, docID string, data any, fields []entity.FieldDefinition) (map[string]interface{}, error)
func (r *mutationResolver) deleteDocument(ctx, slug, docID string, fields []entity.FieldDefinition) (bool, error)
func (r *mutationResolver) publishDocument(ctx, slug, docID string, locale *string, fields []entity.FieldDefinition) (map[string]interface{}, error)
func (r *mutationResolver) unpublishDocument(ctx, slug, docID string, locale *string, fields []entity.FieldDefinition) (map[string]interface{}, error)

// Single-type helpers
func (r *queryResolver) getSingleType(ctx, slug string, locale, status *string, fields []entity.FieldDefinition) (map[string]interface{}, error)
func (r *mutationResolver) saveSingleType(ctx, slug string, data any, locale *string, fields []entity.FieldDefinition) (map[string]interface{}, error)
func (r *mutationResolver) publishSingleType(ctx, slug string, locale *string, fields []entity.FieldDefinition) (map[string]interface{}, error)
func (r *mutationResolver) unpublishSingleType(ctx, slug string, locale *string, fields []entity.FieldDefinition) (map[string]interface{}, error)
```

Each helper follows the same pattern as the current resolver logic:
1. Extract auth from context (mutations require auth)
2. Check status arg (draft vs published)
3. Delegate to `DocumentUseCase`
4. Convert result via `docToMap()` (with media resolution)

### 7.4 Filter Conversion (`resolver/filter.go`)

gqlgen generates **typed filter structs** (e.g., `model.CvPageFilter`). These are converted to `[]entity.FilterNode` using reflection, since all filter structs share the same operator pattern (see §14.2).

```go
func convertFilterStructs[T any](filters []*T) []entity.FilterNode {
    // Reflect over struct fields
    // For each non-nil field:
    //   - "And"/"Or"/"Not" → recurse into combinator
    //   - Other fields → extract operator values into entity.FieldFilter
    // PascalCase struct field names → camelCase for DB column mapping
}
```

This replaces the current `parseFilters()` / `parseFilterMap()` functions. The `entity.FilterNode` domain type remains unchanged.

### 7.5 OrderBy Conversion (`resolver/document_helpers.go`)

gqlgen generates **typed OrderBy structs** (e.g., `model.CvPageOrderBy`). The first non-nil field determines sort column and direction (see §14.1).

```go
func extractOrderBy(orderByArg any) (string, int) {
    // Reflect to find first non-nil *SortOrder field
    // Convert PascalCase field name → camelCase column name
    // Convert SortOrder enum → int direction (ASC=1, DESC=-1)
    // Default: ("createdAt", -1)
}
```

The existing `DocumentUseCase.GetAllPaginated(orderBy string, sortDir int)` signature is unchanged.

### 7.6 Auth & Directives (`handler.go`)

Two layers of auth, each handled by a gqlgen feature:

**Layer 1: HTTP middleware** — JWT/access-token validation. Extracts credentials and injects `UserID` + `Role` into context. All requests pass through (queries may be public).

**Layer 2: `@auth` directive** — Enforces authentication on mutations. Replaces the old `authRequired()` wrapper (see §3.4.1).

```go
func NewHandler(resolver *resolver.Resolver, tokenValidator AccessTokenValidator) http.Handler {
    cfg := generated.Config{Resolvers: resolver}

    // @auth directive — replaces per-mutation authRequired() wrapper
    cfg.Directives.Auth = func(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
        if middleware.UserID(ctx) == "" {
            return nil, fmt.Errorf("unauthorized")
        }
        return next(ctx)
    }

    srv := handler.NewDefaultServer(generated.NewExecutableSchema(cfg))

    // Error presenter — maps domain errors to GraphQL error codes
    srv.SetErrorPresenter(func(ctx context.Context, err error) *gqlerror.Error {
        // Map ErrNotFound → NOT_FOUND, ErrValidation → BAD_REQUEST, etc.
        return graphql.DefaultErrorPresenter(ctx, err)
    })

    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()
        // JWT / access-token validation (same as current)
        // Inject UserID and Role into context — even for unauthenticated requests
        srv.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

The `gqlcodegen` tool adds `@auth` to all mutation fields in generated `.graphql` files automatically. Queries remain public by default (with `status: "draft"` opt-in for authenticated users).
```

### 7.7 Server Integration (`cmd/server/main.go`)

```go
// Replace:
//   gqlFactory := dynamic.NewResolverFactory(documentUC, ctUC, mediaRepo, accessTokenUC)
//   gqlHandler, err := gqlFactory.BuildHandler(defs)

// With:
gqlResolver := resolver.NewResolver(documentUC, ctUC, mediaRepo)
gqlHandler := graphqlpkg.NewHandler(gqlResolver, accessTokenUC)
```

The handler no longer needs `defs` — schema is pre-compiled by gqlgen.

---

## 8. Migration Plan

### Phase 1: Setup (Low Risk)

1. Add gqlgen dependency: `go get github.com/99designs/gqlgen`
2. Add gqlgen tools dependency: `tools.go` with `_ "github.com/99designs/gqlgen"`
3. Create `graphql/gqlgen.yml` configuration
4. Create `graphql/model/types.go` with map type aliases

### Phase 2: Custom Codegen Tool (Medium Risk)

1. Create `cmd/gqlcodegen/main.go`
2. Implement SDL generation (adapt logic from current `schema_builder.go`)
3. Implement resolver generation (template-based)
4. Implement gqlgen.yml model injection
5. Test: `go run ./cmd/gqlcodegen` produces correct `.graphql` + `content_gen.go`

### Phase 3: gqlgen Integration (Medium Risk)

1. Run `make graphql-generate` — gqlgen generates `generated.go` + `models_gen.go`
2. Implement hand-written resolvers:
   - `resolver/resolver.go` — root resolver
   - `resolver/document_helpers.go` — generic CRUD helpers
   - `resolver/content_types.go` — contentTypes query
   - `resolver/media.go` — media resolution
   - `resolver/filter.go` — filter conversion
3. Implement `handler.go` — HTTP handler with auth

### Phase 4: Server Integration (Low Risk)

1. Update `cmd/server/main.go` to use new gqlgen handler
2. Remove old `graphql/dynamic/` directory
3. Remove `graphql-go/graphql` and `graphql-go/handler` dependencies
4. Update `Makefile` if needed

### Phase 5: Verification

1. Run `go vet ./...` and `go build ./...`
2. Run all existing tests
3. Manual testing: all queries and mutations via GraphQL playground
4. Verify filter, pagination, media resolution, auth

---

## 9. Acceptance Criteria

- [ ] `make graphql-generate` produces all `.graphql` files, `content_gen.go`, and gqlgen output
- [ ] All existing GraphQL queries and mutations work identically
- [ ] All existing tests pass (adapted to new structure)
- [ ] Filters work with typed structs (reflection-based conversion to FilterNode)
- [ ] OrderBy wired up properly (not hardcoded to createdAt DESC)
- [ ] Pagination works as before
- [ ] Media field resolution works as before (including nested component media)
- [ ] JWT and access-token authentication work as before
- [ ] `go vet ./...` and `go build ./...` pass
- [ ] No regression in response shapes or error formats
- [ ] `graphql-go/graphql` dependency fully removed
- [ ] Generated files have `DO NOT EDIT` headers
- [ ] Generated files are .gitignored; codegen runs at build time
- [ ] Adding a new content-type JSON + running `make graphql-generate` produces correct schema + resolvers
- [ ] Render.com build command works: codegen → compile → deploy
- [ ] CI pipeline runs codegen before build/test

---

## 10. Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| `map[string]interface{}` field resolution may not work for all types in gqlgen | High | Prototype with one content type first; verify nested components, media, JSON scalar |
| Reflection-based filter/orderBy conversion is fragile | Medium | Comprehensive test coverage for all filter operator + field type combinations |
| gqlgen generated code conflicts with custom code | Medium | Strict separation: generated files in `generated/` and `model/`; hand-written in `resolver/` |
| Build step forgotten locally after JSON change | Low | Document in CLAUDE.md; `go build` fails if generated code is missing |
| Render.com build timeout with codegen step | Low | Codegen is fast (~1s); total build well within free plan limits |
| gqlgen version upgrades break generated code | Low | Pin gqlgen version in `go.mod`; upgrade deliberately |

---

## 11. Performance Comparison

| Library | Ops/sec (unit bench) | Req/sec (HTTP load) | Relative |
|---------|---------------------|---------------------|----------|
| **gqlgen** (target) | 383,760 | **52,854** | **2.7x faster** |
| graph-gophers | 165,276 | 44,308 | 2.3x faster |
| graphql-go (current) | 12,556 | 19,004 | 1.0x (baseline) |

Source: [appleboy/golang-graphql-benchmark](https://github.com/appleboy/golang-graphql-benchmark)

---

## 12. Deployment (Render.com Native Go — Free Plan)

### 12.1 Build Command

Render.com native Go services support custom build commands. Chain codegen before compilation:

```
cd apps/api && go run ./cmd/gqlcodegen && go run github.com/99designs/gqlgen generate && go build -o server ./cmd/server
```

### 12.2 Start Command

```
./apps/api/server
```

### 12.3 Why Not Docker

The project uses Render.com **free plan with native Go**, not Docker. The existing `Dockerfile` is available but not used for deployment. Native Go on Render.com:
- Installs Go automatically
- Runs the build command in the repo root
- Supports `go run` for tools during build
- No Docker image size or layer concerns

### 12.4 CI Pipeline (GitHub Actions)

Update CI to run codegen before tests/build:

```yaml
- name: Generate GraphQL
  run: cd apps/api && go run ./cmd/gqlcodegen && go run github.com/99designs/gqlgen generate

- name: Build
  run: cd apps/api && go build ./...

- name: Test
  run: cd apps/api && go test ./...
```

---

## 13. Out of Scope

- Subscriptions (no current need; gqlgen supports adding later)
- Federation (single-service CMS; gqlgen supports adding later)
- Schema hot-reload (schema rebuilds only via `make graphql-generate`)
- Changes to content-type JSON schema format
- Changes to document usecase or repository layers
- Changes to REST API
- GraphQL playground configuration (gqlgen provides built-in; configure later)
- Switching from native Go to Docker on Render.com

---

## 14. Resolved Decisions

### 14.1 OrderBy — Wire Up Properly

gqlgen has **no built-in orderBy support** — it is schema-first, so all types including OrderBy must be defined in `.graphql` files manually. Since our `gqlcodegen` tool already generates `<Type>OrderBy` input types in the schema, this migration **will wire up orderBy properly** instead of hardcoding `createdAt DESC`.

**Implementation:**
- gqlgen generates typed `model.<Type>OrderBy` structs (e.g., `model.CvPageOrderBy`)
- Each struct has `*SortOrder` fields for sortable columns
- `document_helpers.go` converts the typed struct to `(orderBy string, sortDir int)` for the usecase:

```go
func extractOrderBy(orderByArg any) (string, int) {
    // Use reflection to find the first non-nil field in the OrderBy struct
    // Return (fieldName, direction) where direction is 1 (ASC) or -1 (DESC)
    // Default: ("createdAt", -1) if no orderBy provided
}
```

- The existing `DocumentUseCase.GetAllPaginated()` already accepts `orderBy string, sortDir int` — no usecase changes needed.

### 14.2 Filters — Typed Structs (gqlgen-generated)

Use **gqlgen-generated typed structs** for filter types (not `map[string]interface{}`). This gives compile-time type safety consistent with choosing gqlgen.

**Implementation:**
- gqlgen generates structs like `model.CvPageFilter` with typed fields
- `resolver/filter.go` converts these to `[]entity.FilterNode` using **reflection**, since all filter structs follow the same pattern:
  - Operator fields: `*<Type>Filter` (eq, ne, in, notIn)
  - Combinator fields: `And []*<Self>`, `Or []*<Self>`, `Not *<Self>`

```go
func convertFilterStructs[T any](filters []*T) []entity.FilterNode {
    // Reflect over struct fields
    // For each non-nil field:
    //   - "And"/"Or"/"Not" → recurse into combinator
    //   - Other fields → extract operator values into entity.FieldFilter
    // Field names are PascalCase in struct → convert to camelCase for DB
}
```

- The existing `parseFilters()` and `parseFilterMap()` functions are **replaced** by `convertFilterStructs()`.
- `entity.FilterNode` and `entity.FieldFilter` domain types remain unchanged.

### 14.3 Generated Files — .gitignore (Codegen at Build Time)

Generated files are **.gitignored** — codegen runs during the Render.com build step.

**Render.com free plan (native Go)** supports custom build commands. The build command chains codegen before compilation:

```
cd apps/api && go run ./cmd/gqlcodegen && go run github.com/99designs/gqlgen generate && go build -o server ./cmd/server
```

**Render.com configuration:**
- **Build Command:** `cd apps/api && go run ./cmd/gqlcodegen && go run github.com/99designs/gqlgen generate && go build -o server ./cmd/server`
- **Start Command:** `./apps/api/server`

**.gitignore additions:**
```gitignore
# gqlgen generated files (regenerated at build time)
apps/api/graphql/schema/
apps/api/graphql/generated/
apps/api/graphql/model/models_gen.go
apps/api/graphql/resolver/content_gen.go
```

**CI check:** GitHub Actions runs `make graphql-generate && go build ./...` to verify codegen + compilation passes. No `git diff --exit-code` check needed since files aren't committed.

**Local development:** Developers must run `make graphql-generate` after changing `content-types/*.json`. Add a note to `CLAUDE.md` commands table.
