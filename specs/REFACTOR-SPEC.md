# SPEC — BE Module Restructuring & Spec Splitting

## 1. Objective

Split the monolithic `SPEC.md` (~4,250 lines, ~53k tokens) into standalone per-module spec files and reorganize the BE documentation into 5 logical modules: **core**, **auth**, **content**, **media**, **admin**. Each module spec is self-contained so AI coding agents only read the relevant module when working on a domain — reducing token consumption by ~80% per task.

The global `SPEC.md` becomes a thin navigation index (~200 lines) that helps AI agents locate the right module spec, track planned/in-progress changes, and understand cross-module dependencies.

**Scope:** BE (api server) spec and documentation only. FE (web UI) stays unchanged — the existing FE spec sections (§7) remain in the global spec.

**Architecture constraint:** Keep the existing Clean Architecture horizontal layers (`domain/entity`, `domain/repository`, `usecase/`, `infrastructure/`, `delivery/`). No directory moves or package restructuring — the code is already well-organized. This refactoring is purely about documentation structure.

---

## 2. Module Definitions

### 2.1 core

Shared foundation used by all modules: entities, repository interfaces, error types, config, JWT utilities, database clients, and cross-cutting middleware.

**Owns:**
- `internal/config/` — application configuration
- `internal/domain/entity/` — all entity definitions (shared types)
- `internal/domain/repository/` — all repository interfaces
- `internal/infrastructure/mongodb/client.go` — MongoDB client setup
- `internal/infrastructure/mongodb/indexes.go` — index management
- `internal/infrastructure/gormdb/client.go` — GORM client setup
- `internal/infrastructure/gormdb/driver_*.go` — build-tag drivers
- `internal/delivery/http/handler/response.go` — shared response helpers
- `internal/delivery/http/middleware/cors.go` — CORS middleware
- `internal/delivery/http/middleware/bodylimit.go` — body size limit
- `internal/delivery/http/middleware/security_headers.go` — security headers
- `internal/delivery/http/router.go` — Gin router setup
- `internal/delivery/grpc/server.go` — gRPC server setup
- `internal/delivery/grpc/errors.go` — gRPC error mapping
- `internal/grpcclient/` — gRPC client manager
- `pkg/errors/` — domain error types
- `pkg/jwt/` — JWT sign/validate
- `cmd/server/main.go` — application entry point, wiring

**Spec content from current SPEC.md:**
- §2 Commands (all build/test/run commands)
- §3 Project Structure
- §4 Code Style (Go conventions, Clean Architecture rules)
- §5 Testing Strategy (shared rules)
- §6 Boundaries (global always/ask/never)
- §9.5 CORS middleware
- §9.9 Request body size limit
- §9.10 Security response headers
- §11.2 Tech stack summary
- §11.3 Project structure (after refactor)
- §11.4 Gin migration (framework-level)
- §11.6 GORM client + per-entity DB selection
- §11.7.5 gRPC client
- §11.8 Configuration (full env var table)
- Resolved Decisions
- §13 Deployment (Render.com setup)

---

### 2.2 auth

User authentication, registration, JWT token lifecycle, password/email validation, role-based access control, and permission enforcement.

**Owns:**
- `internal/domain/entity/user.go`
- `internal/domain/entity/role.go`
- `internal/domain/repository/user_repository.go`
- `internal/domain/repository/role_repository.go`
- `internal/usecase/auth/`
- `internal/usecase/role/`
- `internal/delivery/http/handler/auth_handler.go`
- `internal/delivery/http/handler/role_handler.go`
- `internal/delivery/http/middleware/auth.go`
- `internal/delivery/http/middleware/gin_auth.go`
- `internal/delivery/http/middleware/role_cache.go`
- `internal/delivery/http/middleware/ratelimit.go` (auth endpoint rate limiting)
- `internal/delivery/grpc/auth_service.go`
- `internal/delivery/grpc/interceptor/auth.go`
- `internal/infrastructure/mongodb/user_repository.go`
- `internal/infrastructure/mongodb/role_repository.go`
- `internal/infrastructure/gormdb/user_repository.go`
- `internal/infrastructure/gormdb/role_repository.go`
- `proto/cms/v1/auth.proto`

**Spec content from current SPEC.md:**
- §4 Auth section (access/refresh tokens, roles, JWT)
- §9.4 Secure cookie flag
- §9.6 Password validation
- §9.7 Email validation
- §9.8 Rate limiting (auth endpoints)
- §11.7.2 auth.proto definition
- §11.7.4 gRPC auth interceptor
- §12 Dynamic Roles & Permission-Based Access (entire section)
- §13.4.2 Cross-origin cookie configuration

---

### 2.3 content

Content types (schema-as-code), document management (draft/publish workflow), schema sync, paginated collections, and dynamic GraphQL schema generation.

**Owns:**
- `internal/domain/entity/content_type.go`
- `internal/domain/entity/document.go`
- `internal/domain/repository/content_type_repository.go`
- `internal/domain/repository/document_repository.go`
- `internal/usecase/content_type/` (including schema_loader, sync)
- `internal/usecase/document/`
- `internal/delivery/http/handler/content_type_handler.go`
- `internal/delivery/http/handler/document_handler.go`
- `internal/delivery/http/handler/locale_handler.go`
- `internal/delivery/grpc/content_type_service.go`
- `internal/delivery/grpc/document_service.go`
- `internal/infrastructure/mongodb/content_type_repository.go`
- `internal/infrastructure/mongodb/document_repository.go`
- `internal/infrastructure/gormdb/content_type_repository.go`
- `internal/infrastructure/gormdb/document_repository.go`
- `graphql/dynamic/` (schema builder + resolver factory)
- `content-types/` (JSON schema definitions)
- `proto/cms/v1/document.proto`
- `proto/cms/v1/content_type.proto`

**Spec content from current SPEC.md:**
- §4 Domain Rules: Draft & Publish
- §4 Domain Rules: Content Type Kinds
- §4 Domain Rules: Content-Type Schema as Code
- §4 Database IDs
- §9.1–§9.3 Slug validation, documentID validation
- §10 Document Manager API Restructure & Paginated Collections (entire section)
- §11.5 Dynamic GraphQL Schema Generation (entire section)
- §11.7.2 document.proto, content_type.proto definitions
- §11.7.3 gRPC document service implementation

---

### 2.4 media

Media asset management, file upload/delete, storage provider adapters (S3, Cloudinary).

**Owns:**
- `internal/domain/entity/media_asset.go`
- `internal/domain/repository/media_asset_repository.go`
- `internal/domain/repository/storage_adapter.go`
- `internal/usecase/media/`
- `internal/delivery/http/handler/media_handler.go`
- `internal/delivery/grpc/media_service.go`
- `internal/infrastructure/mongodb/media_asset_repository.go`
- `internal/infrastructure/gormdb/media_asset_repository.go`
- `internal/infrastructure/cloudinary/`
- `internal/infrastructure/s3/`
- `proto/cms/v1/media.proto`

**Spec content from current SPEC.md:**
- §8 Delete Media Asset (entire section)
- §11.7.2 media.proto definition
- Resolved Decision #1 (S3 + Cloudinary behind storage interface)

---

### 2.5 admin

User management, invite system, and access token management for API consumers.

**Owns:**
- `internal/domain/entity/invite.go`
- `internal/domain/entity/access_token.go`
- `internal/domain/repository/invite_repository.go`
- `internal/domain/repository/access_token_repository.go`
- `internal/usecase/user/`
- `internal/usecase/invite/`
- `internal/usecase/access_token/`
- `internal/delivery/http/handler/user_handler.go`
- `internal/delivery/http/handler/invite_handler.go`
- `internal/delivery/http/handler/access_token_handler.go`

**Spec content from current SPEC.md:**
- User management API contracts (from §10 routes)
- Invite system (from §12 / current routes)
- Access token management (from current routes)

---

## 3. Cross-Module Dependencies

```
core ← auth     (auth uses user/role repos from core interfaces)
core ← content  (content uses content-type/document repos from core interfaces)
core ← media    (media uses media-asset repo + storage adapter from core interfaces)
core ← admin    (admin uses user/invite/access-token repos from core interfaces)

auth ← admin    (admin.user_usecase uses auth.role_usecase for role assignment)
content ← media (content.document_usecase cascade-deletes media assets)
content ← auth  (content uses auth middleware for permission checks)
```

**Rule:** All cross-module communication goes through interfaces defined in `core` (i.e., `domain/repository/`). No module imports another module's usecase or handler directly — only through interfaces injected via constructor in `main.go`.

---

## 4. Spec File Structure

```
SPEC.md                          ← Global navigation index (~200 lines)
specs/
├── core.md                      ← Core module spec
├── auth.md                      ← Auth module spec
├── content.md                   ← Content module spec
├── media.md                     ← Media module spec
├── admin.md                     ← Admin module spec
└── REFACTOR-SPEC.md             ← This file (refactoring plan)
```

---

### 4.1 Global SPEC.md Structure (New)

The slimmed-down global SPEC.md serves three purposes:
1. **Navigate** — point AI agents to the right module spec
2. **Track** — list planned and in-progress changes before they're finalized into module specs
3. **Share** — hold truly global rules (code style, testing strategy, boundaries)

```markdown
# SPEC — personal-cms

One-paragraph project objective.

## Commands
Build/test/run commands (unchanged from current §2).

## Module Index
| Module | Spec | Description |
|--------|------|-------------|
| core   | [specs/core.md](specs/core.md) | Shared entities, interfaces, config, middleware, DB clients |
| auth   | [specs/auth.md](specs/auth.md) | Authentication, JWT, roles, permissions |
| content| [specs/content.md](specs/content.md) | Content types, documents, draft/publish, GraphQL |
| media  | [specs/media.md](specs/media.md) | Media assets, upload/delete, storage adapters |
| admin  | [specs/admin.md](specs/admin.md) | User mgmt, invites, access tokens |

## Cross-Module Dependency Graph
(Diagram from §3 above)

## Code Style
(Condensed from current §4 — Go conventions, Clean Architecture rules, naming)

## Testing Strategy
(Condensed from current §5 — shared rules only, module-specific testing in module specs)

## Global Boundaries
(Only truly global always/ask/never rules from §6)

## FE Spec (Unchanged)
(§7 Web — Content-Type Management System stays here since FE is not modularized)

## Deployment
(§13 or link to it — deployment is cross-cutting)

## Pending Changes
(Section for tracking planned/in-progress work before it's finalized into module specs)

## Resolved Decisions
(Condensed list)
```

---

### 4.2 Per-Module Spec Template

Each module spec is standalone — an AI agent reads only this file when working on the module.

```markdown
# SPEC — {module-name} Module

## 1. Overview
One paragraph: what this module does, who it serves.

## 2. File Map
All files owned by this module (paths relative to apps/api/).

## 3. Entities
Entity definitions with field descriptions and invariants.
GORM + BSON struct tags documented.

## 4. Repository Interfaces
Interface methods with behavior contracts, error returns, and DB-agnostic guarantees.

## 5. Use Cases
Business logic methods: signatures, input/output, error conditions, side effects.

## 6. API Contracts
### REST (Gin)
Route table with method, path, query params, request/response shapes.

### gRPC
Proto service definition with method descriptions.

### GraphQL (content module only)
Generated queries/mutations per content-type.

## 7. Infrastructure
MongoDB and GORM repository implementation notes.
Storage adapter details (media module only).

## 8. Testing
Module-specific test strategy, fixture patterns, coverage targets.

## 9. Boundaries
Always/Ask/Never rules specific to this module.

## 10. Changelog
Completed features/changes tracked here (moved from global spec after implementation).
```

---

## 5. Content Mapping — Current SPEC Sections → Modules

| Current Section | Target Module | Notes |
|---|---|---|
| §1 Objective | Global | Condensed to 1 paragraph |
| §2 Commands | Global | Unchanged |
| §3 Project Structure | Global | Updated to reference module specs |
| §4 Code Style — Go/Architecture | Global | Shared rules |
| §4 Code Style — Auth | auth | JWT, tokens, cookies |
| §4 Code Style — React/TanStack/Forms | Global (FE) | Unchanged |
| §4 Domain Rules: Draft & Publish | content | Core business logic |
| §4 Domain Rules: Content Type Kinds | content | Single/collection semantics |
| §4 Domain Rules: Schema as Code | content | JSON definitions, sync |
| §4 Database IDs | content | documentId conventions |
| §5 Testing Strategy — Backend | Global + per-module | Shared rules in global, specific in modules |
| §5 Testing Strategy — Frontend | Global (FE) | Unchanged |
| §5 CI | Global | Unchanged |
| §6 Boundaries | Global + per-module | Global rules stay, module-specific rules move |
| Resolved Decisions | Global | Keep as reference |
| §7 Web (FE Refactor) | Global (FE) | Unchanged |
| §8 Delete Media Asset | media | Full section |
| §9.1–§9.3 Slug/DocID validation | content | Input validation |
| §9.4 Secure cookie flag | auth | Cookie security |
| §9.5 CORS middleware | core | Cross-cutting |
| §9.6 Password validation | auth | Registration |
| §9.7 Email validation | auth | Registration |
| §9.8 Rate limiting | auth | Auth endpoint protection |
| §9.9 Body size limit | core | Cross-cutting |
| §9.10 Security headers | core | Cross-cutting |
| §10 Document Manager API | content | Full section |
| §11.2–11.3 Tech stack/structure | core | Foundation |
| §11.4 Gin migration | core | Framework-level |
| §11.5 Dynamic GraphQL | content | Schema generation |
| §11.6 GORM adapters | core + per-module infra | Client in core, repos in modules |
| §11.7 gRPC server | Split per service | auth.proto→auth, document.proto→content, etc. |
| §11.8 Configuration | core | Full env var table |
| §12 Roles & Permissions | auth | Full section |
| §13 Deployment | Global | Cross-cutting |

---

## 6. Migration Plan

### Phase 1: Create per-module spec files

For each module (core, auth, content, media, admin):
1. Create `specs/{module}.md` using the template from §4.2
2. Extract relevant content from monolithic SPEC.md
3. Ensure each spec is self-contained (can be read without the global spec)
4. Add module-specific boundaries, testing rules, and file maps

### Phase 2: Slim down global SPEC.md

1. Replace detailed sections with the navigation index structure from §4.1
2. Keep only: objective, commands, module index, code style, testing strategy (shared), global boundaries, FE spec, deployment, resolved decisions, pending changes
3. Remove all content that has been moved to module specs
4. Target: ~200 lines

### Phase 3: Validate completeness

1. Diff the total content: old monolithic SPEC vs (new global + all module specs)
2. Ensure no spec content was lost — every requirement maps to exactly one location
3. Check cross-references: if module A's spec references module B's concept, it links to `specs/b.md`

### Phase 4: Update CLAUDE.md (if needed)

If the project has a CLAUDE.md that references SPEC.md sections by number, update those references to point to the correct module spec file.

---

## 7. Token Budget Impact (Estimated)

| Scenario | Before | After |
|---|---|---|
| Read full spec | ~53k tokens | ~53k total (unchanged aggregate) |
| Work on auth feature | ~53k (read all) | ~8k (read specs/auth.md) |
| Work on content feature | ~53k (read all) | ~15k (read specs/content.md) |
| Work on media feature | ~53k (read all) | ~5k (read specs/media.md) |
| Work on admin feature | ~53k (read all) | ~4k (read specs/admin.md) |
| Navigate to right module | N/A | ~3k (read global SPEC.md) |
| Typical task total | ~53k | ~6k–18k (global index + one module) |

**Net savings: 60-90% token reduction per task.**

---

## 8. Acceptance Criteria

- [ ] 5 standalone module specs created in `specs/` directory
- [ ] Global SPEC.md slimmed to ~200 lines (navigation index)
- [ ] Every section from the original SPEC.md mapped to exactly one module spec
- [ ] No spec content lost during migration
- [ ] Each module spec includes: overview, file map, entities, interfaces, usecases, API contracts, testing, boundaries
- [ ] Cross-module references use relative links (`specs/auth.md`)
- [ ] An AI agent can work on a single module by reading only `SPEC.md` (for navigation) + `specs/{module}.md` (for details)
- [ ] FE spec sections remain in global SPEC.md (unchanged)

---

## 9. Boundaries

| Rule | Detail |
|---|---|
| **Always** | Keep each module spec self-contained — readable without the global spec |
| **Always** | Include file maps in every module spec so AI agents know which files to read |
| **Always** | Update the global SPEC.md pending changes section when planning new work |
| **Always** | Move completed changes from global pending → relevant module spec changelog |
| **Always** | Link cross-module references with relative paths |
| **Never** | Duplicate spec content across module specs — reference instead |
| **Never** | Put FE spec content in module specs (FE stays in global) |
| **Never** | Change Go package structure or move files — this is a documentation-only refactor |
| **Ask first** | Adding a new module beyond the 5 defined |
| **Ask first** | Moving FE spec into its own module spec (future scope) |

---

## 10. Out of Scope

- Go code restructuring (no file moves, no package changes)
- FE spec modularization
- New features or bug fixes
- Test changes
- CI/CD changes
