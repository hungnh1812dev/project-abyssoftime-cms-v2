# RULES — core Module

**Scope:** Shared entities, repository interfaces, error types, configuration, JWT utilities, database clients, middleware, router, gRPC, deployment.
**Spec:** [specs/core.md](../specs/core.md)

---

## 1. Entity Rules

### 1.1 Entity Struct Tags
- Every entity must have triple tags: `bson` + `gorm` + `json`
- `json:"-"` for internal IDs (e.g., MongoDB `_id`, GORM `id`)
- `json:"camelCase"` for all API-exposed fields
- `gorm:"column:snake_case"` for all GORM-mapped fields
- `bson:"camelCase"` for MongoDB fields

### 1.2 ID Conventions
- `ID` (internal) → `bson:"_id,omitempty" gorm:"column:id;primaryKey"`
- `DocumentID` (domain identifier) → UUID v4, used by all higher layers
- Higher layers (usecase, handler, frontend) only work with `documentId` — **NEVER** with MongoDB `_id`
- Static tables use `id` (string UUID) as PK column name
- Dynamic tables (`documents_*`, `components_*`) use `gorm_id` (auto-increment uint) as PK

### 1.3 Timestamps
- `CreatedAt time.Time` — set on creation, never modified
- `UpdatedAt time.Time` — updated on every modification
- `PublishedAt *time.Time` — nullable pointer, set on publish

### 1.4 Adding New Entities
- Define struct in `internal/domain/entity/`
- Define repository interface in `internal/domain/repository/`
- Create mock in `internal/domain/repository/mock/`
- Implement for both MongoDB (`infrastructure/mongodb/`) and GORM (`infrastructure/gormdb/`)
- Wire in `cmd/server/main.go`

---

## 2. Repository Interface Rules

### 2.1 Interface Location
- All interfaces in `internal/domain/repository/`
- All mocks in `internal/domain/repository/mock/`
- Interface naming: `<Entity>Repository` (e.g., `ContentTypeRepository`)

### 2.2 Method Signatures
- First parameter is always `ctx context.Context`
- Return errors as the last return value
- `Find*` methods return `(*entity.X, error)` or `([]*entity.X, error)`
- `Create`/`Update` take `*entity.X`
- `Delete` takes the ID string

### 2.3 Dual Implementation Requirement
- Every repository interface must have both MongoDB and GORM implementations
- Both implementations must pass the **same** test cases
- Use interface-based testing to ensure behavioral parity

---

## 3. Configuration Rules

### 3.1 Config Loading
- All env vars loaded once at boot into `internal/config/config.go`
- **NEVER** call `os.Getenv` outside `config.go`
- Every new env var must be documented in `specs/core.md §7`

### 3.2 Per-Entity DB Selection
- `DB_<ENTITY>` env var overrides `DB_DRIVER` for that entity
- Supported values: `mongo`, `postgres`
- Repository factory in `main.go` wires the correct implementation

---

## 4. Error Type Rules

### 4.1 Domain Errors (`pkg/errors/`)
- Use sentinel errors: `ErrNotFound`, `ErrValidation`, `ErrConflict`
- Wrap with `fmt.Errorf("context: %w", err)` for additional context
- **NEVER** create new sentinel error types without adding them to the mapping table
- **NEVER** return raw `errors.New()` from usecase — always wrap a sentinel

### 4.2 Error Propagation
- Repository → returns domain error (e.g., `ErrNotFound`)
- Usecase → wraps with context if needed, propagates domain error
- Handler → maps domain error to HTTP/gRPC status code via `writeErr`/`ginWriteErr`

---

## 5. Middleware Rules

### 5.1 CORS (`middleware/cors.go`)
- Check `Origin` against `CORS_ORIGINS` whitelist
- **NEVER** use `Access-Control-Allow-Origin: *`
- Handle `OPTIONS` preflight with 204
- Set `Allow-Credentials: true`

### 5.2 Body Size Limit (`middleware/bodylimit.go`)
- Applied globally with `BODY_LIMIT_BYTES` (default 10 MB)
- Media upload exempt — uses its own multipart limit
- Uses `http.MaxBytesReader`

### 5.3 Security Headers (`middleware/security_headers.go`)
- Set on **every** response: `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `X-XSS-Protection: 0`, `Referrer-Policy: strict-origin-when-cross-origin`, `CSP: default-src 'self'`

### 5.4 Adding New Middleware
- Create file in `internal/delivery/http/middleware/`
- Create corresponding test file
- Register in `router.go` at the appropriate scope (global vs group)

---

## 6. Router Rules

### 6.1 Route Registration
- Centralized in `internal/delivery/http/router.go`
- `SetupRouter` accepts `RouterConfig` struct with all handler instances
- Routes grouped by module with appropriate middleware
- Permission-protected routes use `GinRequirePermission`

### 6.2 Adding New Routes
- Add handler reference to `RouterConfig` struct
- Register in the appropriate group in `SetupRouter`
- Document the route in the module's spec (Method, Route, Permission, Response)

---

## 7. gRPC Rules

### 7.1 Service Implementation
- Proto definitions in `proto/cms/v1/`
- Generated code: `*.pb.go`, `*_grpc.pb.go` — **NEVER** hand-edit
- Service implementations in `internal/delivery/grpc/`
- Auth interceptor in `internal/delivery/grpc/interceptor/`

### 7.2 gRPC Client
- Client manager in `internal/grpcclient/`
- Configured via `GRPC_SERVICES` env var
- **Ask first** before adding new gRPC client service connections

---

## 8. Database Client Rules

### 8.1 MongoDB Client (`infrastructure/mongodb/client.go`)
- Connection via `MONGODB_URI`
- Index management in `infrastructure/mongodb/indexes.go`
- Per-content-type collections: `documents_<slug>`

### 8.2 GORM Client (`infrastructure/gormdb/client.go`)
- Connection via `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USERNAME`, `DB_PASSWORD`
- `AutoMigrate` for static entity tables on startup
- Dynamic tables (`documents_*`, `components_*`) created by `EnsureCollection`, **NOT** by AutoMigrate
- `DB_SSL_MODE=require` for Supabase production

### 8.3 EnsureCollection (Critical)
- **MUST be non-destructive** — preserve existing data on startup
- If table does NOT exist → CREATE TABLE
- If table exists → ADD missing columns only (ALTER TABLE ADD COLUMN)
- **NEVER** drop existing columns (even for removed fields)
- **NEVER** change column types on existing columns
- **NEVER** use DROP TABLE + CREATE TABLE pattern (causes production data loss)

---

## 9. Deployment Rules

### 9.1 Render.com
- API: Web Service (native Go binary), port 8080
- Web: Static Site (React/Vite SPA)
- `COOKIE_SAMESITE=none` + `COOKIE_SECURE=true` for cross-origin
- `CORS_ORIGINS` set to Static Site URL (explicit, never `*`)
- SPA rewrite rule: `/* → /index.html`
- Keep-alive pings `/health` every 14 min

### 9.2 Build & Deploy
- Deploy hooks from CI (not auto-deploy)
- `go build -o bin/server ./cmd/server` for production binary
- `npm run build` for frontend static assets

---

## 10. Testing Rules (Core-Specific)

### 10.1 What to Test
- `pkg/errors/` — error wrapping and type checking
- `pkg/jwt/` — sign/validate, expiry, invalid tokens
- Middleware — CORS, body limit, security headers, rate limiting
- Response helpers — `writeErr`, `writeJSON`
- DB clients — connection, migration, index creation

### 10.2 Test Patterns
- Middleware tests use `httptest.NewRecorder` + `gin.SetMode(gin.TestMode)`
- DB client tests use Docker (MongoDB) or in-memory SQLite (GORM)
- JWT tests cover valid, expired, malformed, and missing tokens

---

## 11. Boundaries Summary

| Rule | Detail |
|---|---|
| **Always** | Keep REST endpoint paths and response shapes identical after any framework migration |
| **Always** | Map domain errors to protocol-appropriate codes in every delivery layer |
| **Always** | Initialize both MongoDB and SQL connections at startup when configured |
| **Always** | Run `AutoMigrate` for GORM static entities on startup |
| **Always** | Use `gin.SetMode(gin.TestMode)` in all test files |
| **Always** | Apply body size limit globally; media upload uses its own multipart limit |
| **Always** | Set security headers on every response |
| **Always** | Make `EnsureCollection` non-destructive (no DROP TABLE) |
| **Never** | Use MongoDB-specific logic outside `infrastructure/mongodb/` |
| **Never** | Use GORM-specific logic outside `infrastructure/gormdb/` |
| **Never** | Duplicate business logic across REST, GraphQL, and gRPC |
| **Never** | Use `Access-Control-Allow-Origin: *` |
| **Never** | Hard-code SQL dialect-specific queries in GORM repos |
| **Never** | Call `os.Getenv` outside `config.go` |
| **Ask first** | Choosing between PostgreSQL and MySQL for production |
| **Ask first** | Adding new gRPC client service connections |
| **Ask first** | Changing the gRPC port default |
