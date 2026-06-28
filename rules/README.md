# Module Rules Index

These rules are the single source of truth for project conventions. They apply to **every** plan, code change, feature addition, and code style decision within their scope.

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
| [content-type.md](content-type.md) | content | Content type entity, schema-as-code, schema sync, JSON field definitions, configurable list columns |
| [document.md](document.md) | content | Document entity, draft/publish workflow, single-type and collection-type CRUD, pagination, API contracts, duplicate documents |
| [graphql.md](graphql.md) | content | Build-time GraphQL schema generation (gqlgen), resolvers, filters, ordering, codegen pipeline |
| [component.md](component.md) | content | Component entity, repeatable/non-repeatable components, nested component tables, component CRUD |
| [media.md](media.md) | media | Media assets, upload/delete, S3/Cloudinary adapters |
| [admin.md](admin.md) | admin | User management, invites, access tokens |
| [i18n.md](i18n.md) | i18n | Locale management, locale selector, settings page |
| [frontend.md](frontend.md) | frontend | Project structure, TypeScript conventions, component library (Shadcn/Tailwind), repeatable component UI |
| [frontend-forms.md](frontend-forms.md) | frontend | FormProvider, field rendering, dot-notation deserialization, form state management |
| [frontend-data.md](frontend-data.md) | frontend | TanStack Query, query keys, mutations, invalidation, health ping, locale switching |
| [frontend-routing.md](frontend-routing.md) | frontend | React Router routes, content-type registry, sidebar navigation, testing |

### Infrastructure Rules (Granular)
| File | Scope |
|------|-------|
| [mongodb.md](mongodb.md) | MongoDB client, BSON filters, CRUD patterns, index management, cursor handling, component storage |
| [postgresql.md](postgresql.md) | GORM client, raw SQL DDL, row↔entity conversion, EnsureCollection, field serialization, dialect detection |
| [content-type-parsing.md](content-type-parsing.md) | JSON schema loading, FieldDefinition model, validation rules, sync engine, component table paths, layout flattening |

## Rule Priority

1. **Global rules** (`GLOBAL.md`) — always apply, never overridden
2. **Module rules** (`<module>.md`) — extend global rules for specific modules
3. **If conflict** between any of the above → **ask user**

## Cross-Module Dependencies

```
GLOBAL ← core ← auth ← admin
                     ← content-type ← document ← media
                     ← component
                     ← graphql
                     ← i18n
         frontend ← frontend-forms
                  ← frontend-data
                  ← frontend-routing
```

Every module depends on `core`. Cross-module communication goes through `domain/repository/` interfaces only.
