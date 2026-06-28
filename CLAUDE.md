# AbyssOfTime CMS v2

A lightweight, code-first Personal Headless CMS. Go backend (Clean Architecture) + React frontend (TypeScript strict, TanStack Query, Shadcn UI). Content types defined as JSON schema files; the API syncs them on startup. Every entry follows a draft → publish workflow.

## Module Rules (MANDATORY)

**Before writing any spec, plan, or code**, read the relevant rule files from `rules/`:

| Rule File | When to Read |
|-----------|-------------|
| `rules/GLOBAL.md` | **Always** — applies to every task |
| `rules/core.md` | Touching entities, repos, errors, config, middleware, router, DB clients |
| `rules/auth.md` | Touching auth, JWT, passwords, roles, permissions, rate limiting |
| `rules/content-type.md` | Touching content type entity, schema-as-code, schema sync, field definitions, list columns |
| `rules/document.md` | Touching documents, draft/publish workflow, single/collection CRUD, pagination, API contracts |
| `rules/graphql.md` | Touching GraphQL schema generation, resolvers, filters, ordering, codegen pipeline |
| `rules/component.md` | Touching component entity, repeatable/non-repeatable components, nested tables, component CRUD |
| `rules/media.md` | Touching media assets, upload/delete, storage adapters |
| `rules/admin.md` | Touching user management, invites, access tokens |
| `rules/i18n.md` | Touching locales, language settings, locale selector |
| `rules/frontend.md` | Touching project structure, TypeScript conventions, Shadcn/Tailwind, repeatable component UI |
| `rules/frontend-forms.md` | Touching FormProvider, field rendering, dot-notation, form state management |
| `rules/frontend-data.md` | Touching TanStack Query, query keys, mutations, invalidation, locale switching |
| `rules/frontend-routing.md` | Touching React Router routes, content-type registry, sidebar, frontend testing |
| `rules/mongodb.md` | Touching any MongoDB repo in `infrastructure/mongodb/` |
| `rules/postgresql.md` | Touching any GORM repo in `infrastructure/gormdb/` |
| `rules/content-type-parsing.md` | Touching schema loader, sync engine, or content-type JSON files |

**Conflict resolution:** If a rule conflicts with a spec, plan, or instruction → **ask the user** before proceeding.

See `rules/README.md` for the full index and priority system.

## Project Structure

```
apps/api/          → Go backend (Clean Architecture)
apps/web/          → React frontend (Vite + TypeScript)
rules/             → Per-module rules (source of truth for conventions)
docs/              → Technical overview, guide, local dev
content-types/     → JSON schema-as-code definitions (→ apps/api/content-types/)
```

## Commands

| Command | Description |
|---|---|
| `make dev` | Start API + web in parallel |
| `make dev-api` | Start Go API server only |
| `make dev-web` | Start Vite dev server only |
| `make test-api` | `go test ./...` inside `apps/api` |
| `make test-web` | `vitest run` inside `apps/web` |
| `make mongo-start` | Start MongoDB container |
| `make graphql-generate` | Regenerate GraphQL schema + resolvers from `content-types/*.json` |

## Boundaries

### Always
- Read `rules/GLOBAL.md` + relevant module rule before any work
- Confirm scope and approach before starting new coding tasks
- Present options when multiple valid approaches exist
- Follow Clean Architecture: usecase imports only domain; no cross-layer leakage
- Use TanStack Query for all server state (frontend)
- Every `useMutation` must invalidate affected query keys on success

### Ask Before
- Starting any new coding task — confirm scope and approach
- Multiple valid approaches — present options
- Choosing storage adapter per environment
- Adding new permission actions or roles
- Any change that crosses module boundaries

### Never
- Read, edit, create, or expose `.env` files
- Use drag-and-drop or dynamic form engine
- Use `React.Children.map` or recursive child scanning in `FormProvider`
- Couple usecase logic to a specific database
- Auto-choose implementation path when multiple options exist
- Let public read API return draft data
- Use `any` type in TypeScript
- Use `export default` in frontend code
- Use `Access-Control-Allow-Origin: *`
- Drop tables in `EnsureCollection` (causes production data loss)
