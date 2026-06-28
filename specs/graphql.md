# SPEC — GraphQL Module

## 1. Overview

The GraphQL module provides dynamic schema generation for content types. On startup, after content-type sync completes, the schema builder reads all `ContentTypeDefinition` structs and generates GraphQL types, queries, and mutations for each content type. A resolver factory creates per-content-type resolvers that delegate to the document usecase, ensuring no business logic lives in the resolvers themselves. The generated schema supports filtering, sorting, and pagination for collection-type queries, and handles media and component field resolution recursively.

---

## 2. File Map

All paths relative to `apps/api/`.

```
graphql/dynamic/schema_builder.go                            # Dynamic GraphQL SDL generator
graphql/dynamic/schema_builder_test.go
graphql/dynamic/resolver_factory.go                          # Per-content-type resolver factory
graphql/dynamic/resolver_factory_test.go
```

---

## 3. Dynamic GraphQL Schema Generation

On startup, after content-type sync:
1. Schema builder reads all `ContentTypeDefinition` structs
2. For each content-type, generates GraphQL types, queries, and mutations
3. Resolver factory creates resolvers that delegate to document usecase

### Field Type Mapping

| Content-Type `type` | GraphQL Type |
|---|---|
| `text` | `String` |
| `richtext` | `String` |
| `number` | `Float` |
| `boolean` | `Boolean` |
| `media` | `MediaAsset` object type (`{ documentId, url, thumbnailUrl, fileName, width, height }`) |
| `json` | `JSON` (scalar) |
| `component` | Nested object type (with media sub-fields resolved recursively) |

### Generated Schema Per Content-Type

**Collection-type** generates:
- `Query.<slug>(Id: ID!, locale: String, status: String): <Type>` — fetch one document (nullable; returns `null` if not found). Defaults to published; `status: "draft"` opt-in for authenticated users.
- `Query.<slugList>(where: <Type>Filter, orderBy: <Type>OrderBy, start: Int, size: Int, locale: String): [<Type>!]!` — paginated list with filtering, sorting, and component data
- `Mutation.create<Type>(data: <Type>Input!): <Type>! @auth`
- `Mutation.update<Type>(Id: ID!, data: <Type>Input!): <Type>! @auth`
- `Mutation.delete<Type>(Id: ID!): Boolean! @auth`
- `Mutation.publish<Type>(Id: ID!, locale: String): <Type>! @auth`
- `Mutation.unpublish<Type>(Id: ID!, locale: String): <Type>! @auth`

**Single-type** generates:
- `Query.<slug>(locale: String, status: String): <Type>` — fetch singleton (nullable). Defaults to published; `status: "draft"` opt-in for authenticated users.
- `Mutation.save<Type>(data: <Type>Input!, locale: String): <Type>! @auth`
- `Mutation.publish<Type>(locale: String): <Type>! @auth`
- `Mutation.unpublish<Type>(locale: String): <Type>! @auth`

**Response shape changes (v1.8):** Response wrappers removed — single queries return the type directly (nullable), list queries return `[Type!]!`. The `ResolverFactory` now takes `MediaAssetRepository` as a dependency to resolve media fields into full `MediaAsset` objects.

### Naming Conventions
- Type: PascalCase of slug (`blog-posts` → `BlogPost`)
- Input: `<Type>Input`
- Response (single): `<Type>Response` — wraps `data: <Type>`
- Response (list): `<Type>ListResponse` — wraps `data: [<Type>]`
- Filter: `<Type>Filter` — generated from document + component fields
- OrderBy: `<Type>OrderBy` — sort by document + component fields
- Query single: camelCase of slug (`blog-posts` → `blogPost`)
- Query list: camelCase of slug + `List` (`blog-posts` → `blogPostList`)
- Component types: PascalCase of `<ContentType><ComponentName>` (e.g., `BlogPostBanner`)

### Query Examples

**Single document (single-type or collection-type by ID):**

```graphql
query GetBlogPost($locale: String!) {
  blogPost(Id: "abc123", locale: $locale) {
    title
    coverImage { url }
    banner {
      background { url }
    }
  }
}
```

```json
{
  "blogPost": {
    "data": {
      "title": "...",
      "coverImage": { "url": "..." },
      "banner": {
        "background": { "url": "..." }
      }
    }
  }
}
```

**List query with filter, sort, and pagination:**

```graphql
query GetBlogPostList($locale: String!) {
  blogPostList(
    where: { featured: { eq: true } }
    orderBy: { createdAt: DESC }
    start: 0
    size: 10
    locale: $locale
  ) {
    title
    coverImage { url }
    banner {
      background { url }
    }
  }
}
```

```json
{
  "blogPostList": {
    "data": [
      {
        "title": "...",
        "coverImage": { "url": "..." },
        "banner": {
          "background": { "url": "..." }
        }
      }
    ]
  }
}
```

### Filtering & Sorting

Filters are generated per content-type based on field types:

| Field Type | Supported Operators |
|---|---|
| `text` | `eq`, `ne`, `contains`, `startsWith`, `endsWith`, `in` |
| `number` | `eq`, `ne`, `gt`, `gte`, `lt`, `lte`, `in` |
| `boolean` | `eq` |
| `component` | Nested filter on component sub-fields |

Top-level logical operators: `AND`, `OR`, `NOT`.

OrderBy supports `ASC` / `DESC` on scalar fields (text, number, boolean) and system fields (`createdAt`, `updatedAt`, `publishedAt`).

---

## 4. Testing

**GraphQL (`schema_builder_test.go`, `resolver_factory_test.go`):**
- SDL generation for collection and single types
- Field type mapping
- Resolver delegation to usecase methods

---

## 5. Boundaries

| Rule | Detail |
|---|---|
| **Always** | Route dynamic GraphQL through the same usecase — no logic in resolvers |
| **Always** | GraphQL list queries support `where`, `orderBy`, `start`, `size` parameters |
| **Always** | GraphQL responses wrap document data in a `data` field |
| **Never** | Duplicate business logic across REST, GraphQL, and gRPC — all call usecase methods |

---

## 6. Changelog

| Date | Change | Source |
|------|--------|--------|
| v1.4 | Dynamic GraphQL schema generation | §11.5 |
| v1.7 | GraphQL: list query renamed to `<slug>List`, response wrapped in `data`, filter/orderBy/pagination support | §8 |
| v1.13 | GraphQL: queries default to published (status: "draft" opt-in for auth'd); response wrappers removed (nullable single, `[Type!]!` list) | graphql-overhaul |
| v1.14 | GraphQL: media fields return `MediaAsset` object type; component sub-fields resolve media recursively; `ResolverFactory` takes `MediaAssetRepository` | graphql-overhaul |
