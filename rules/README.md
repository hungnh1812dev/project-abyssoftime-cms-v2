# Module Rules Index

These rules are derived from the project's specs (`specs/*.md`) and documentation (`docs/*.md`). They apply to **every** spec, plan, code change, feature addition, and code style decision within their scope.

## How to Use

1. **Before any work**: Read `GLOBAL.md` + the relevant module rule file
2. **Before any spec/plan**: Cross-check against rules for conflicts
3. **If conflict found**: Ask the user before proceeding
4. **If no rule covers the situation**: Check related module rules, then ask user

## Rule Files

### Module Rules
| File | Module | Scope |
|------|--------|-------|
| [GLOBAL.md](GLOBAL.md) | All | Clean Architecture, code style, testing, security, deployment |
| [core.md](core.md) | core | Entities, repos, errors, config, middleware, router, DB clients, gRPC |
| [auth.md](auth.md) | auth | Authentication, JWT, passwords, roles, permissions, rate limiting |
| [content.md](content.md) | content | Content types, documents, draft/publish, schema sync, GraphQL, components |
| [media.md](media.md) | media | Media assets, upload/delete, S3/Cloudinary adapters |
| [admin.md](admin.md) | admin | User management, invites, access tokens |
| [i18n.md](i18n.md) | i18n | Locale management, locale selector, settings page |
| [frontend.md](frontend.md) | frontend | React, TypeScript, TanStack Query, forms, routing, UI |

### Infrastructure Rules (Granular)
| File | Scope |
|------|-------|
| [mongodb.md](mongodb.md) | MongoDB client, BSON filters, CRUD patterns, index management, cursor handling, component storage |
| [postgresql.md](postgresql.md) | GORM client, raw SQL DDL, row↔entity conversion, EnsureCollection, field serialization, dialect detection |
| [content-type-parsing.md](content-type-parsing.md) | JSON schema loading, FieldDefinition model, validation rules, sync engine, component table paths, layout flattening |

## Rule Priority

1. **Global rules** (`GLOBAL.md`) — always apply, never overridden
2. **Module rules** (`<module>.md`) — extend global rules for specific modules
3. **Spec instructions** (`specs/*.md`) — implementation details within rule boundaries
4. **If conflict** between any of the above → **ask user**

## Cross-Module Dependencies

```
GLOBAL ← core ← auth ← admin
                     ← content ← media
                     ← i18n
         frontend (cross-cutting)
```

Every module depends on `core`. Cross-module communication goes through `domain/repository/` interfaces only.
