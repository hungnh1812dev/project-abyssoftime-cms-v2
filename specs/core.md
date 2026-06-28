# SPEC — Core Module

## 1. Overview

The core module provides the shared foundation for all other modules: entity type definitions, repository interfaces, error types, configuration, JWT utilities, and database clients. Every other module depends on core but core never depends on any domain module. Middleware, routing, gRPC client, commands, and deployment are covered in `specs/middleware.md` and `specs/infrastructure.md`.

---

## 2. File Map

All paths relative to `apps/api/`.

```
internal/config/config.go                            # Configuration loading from env vars
internal/domain/entity/                              # All entity structs (shared types)
internal/domain/repository/                          # All repository interfaces
internal/domain/repository/mock/                     # Mock implementations for testing
internal/infrastructure/mongodb/client.go            # MongoDB client setup
internal/infrastructure/mongodb/client_test.go
internal/infrastructure/mongodb/indexes.go           # MongoDB index management
internal/infrastructure/mongodb/indexes_test.go
internal/infrastructure/gormdb/client.go             # GORM client setup + AutoMigrate
internal/infrastructure/gormdb/client_test.go
internal/infrastructure/gormdb/driver_cgo.go         # CGO-enabled SQLite driver (testing)
internal/infrastructure/gormdb/driver_default.go     # Default driver selection
pkg/errors/errors.go                                 # Domain error types (ErrNotFound, ErrValidation, etc.)
pkg/errors/errors_test.go
pkg/jwt/jwt.go                                       # JWT sign/validate utilities
pkg/jwt/jwt_test.go
go.mod
go.sum
```

---

## 3. Tech Stack

| Layer | Technology |
|-------|------------|
| REST framework | Gin (`github.com/gin-gonic/gin`) |
| GraphQL | gqlgen + dynamic schema generation per content-type |
| gRPC | `google.golang.org/grpc` (server + client) |
| SQL database | GORM (`gorm.io/gorm`) with PostgreSQL/MySQL drivers |
| MongoDB | `go.mongodb.org/mongo-driver` |
| DB selection | Configurable per entity via env vars |

---

## 4. Project Structure

```
apps/api/
├── cmd/server/main.go                    # Wires all protocols: Gin, gRPC, GraphQL
├── content-types/                        # JSON schema-as-code definitions → see specs/content-type.md
├── proto/cms/v1/                         # gRPC protocol buffer definitions
├── graphql/dynamic/                      # Dynamic GraphQL schema builder → see specs/graphql.md
├── internal/
│   ├── config/                           # App configuration
│   ├── domain/
│   │   ├── entity/                       # All entity structs
│   │   └── repository/                   # All repository interfaces (+ mock/)
│   ├── usecase/                          # Business logic (DB-agnostic)
│   │   ├── auth/                         # → see specs/auth.md
│   │   ├── role/                         # → see specs/auth.md
│   │   ├── content_type/                 # → see specs/content-type.md
│   │   ├── document/                     # → see specs/document.md
│   │   ├── media/                        # → see specs/media.md
│   │   ├── user/                         # → see specs/admin.md
│   │   ├── invite/                       # → see specs/admin.md
│   │   └── access_token/                 # → see specs/admin.md
│   ├── infrastructure/
│   │   ├── mongodb/                      # MongoDB repository implementations
│   │   ├── gormdb/                       # GORM repository implementations
│   │   ├── cloudinary/                   # → see specs/media.md
│   │   └── s3/                           # → see specs/media.md
│   ├── delivery/
│   │   ├── http/
│   │   │   ├── handler/                  # Gin handler functions
│   │   │   ├── middleware/               # Gin middleware → see specs/middleware.md
│   │   │   └── router.go                 # Route registration → see specs/infrastructure.md
│   │   └── grpc/                         # gRPC service implementations
│   └── grpcclient/                       # gRPC client adapters → see specs/infrastructure.md
├── pkg/
│   ├── errors/                           # Domain error types
│   └── jwt/                              # JWT utilities
└── go.mod
```

---

## 5. Code Style — Go Conventions

- **Architecture**: Strict Clean Architecture — `usecase` imports only `domain`; `delivery` and `infrastructure` import `usecase` and `domain`. Zero cross-layer leakage.
- **Error handling**: Wrap errors at the usecase boundary; handlers map domain errors to HTTP/gRPC status codes. No naked `error` strings in HTTP responses.
- **Naming**: `PascalCase` for exported identifiers; `camelCase` for unexported. Repository interfaces use `<Entity>Repository` (e.g., `DocumentRepository`).
- **Cascade deletion**: Implemented in the usecase layer, not as DB-level triggers, to remain DB-agnostic.
- **No `init()` functions** in business logic packages.
- **Never** use MongoDB-specific logic (ObjectID, bson primitives) outside `infrastructure/mongodb/`.
- **Never** use GORM-specific logic (gorm.DB, gorm.Model) outside `infrastructure/gormdb/`.
- **Never** duplicate business logic across REST handlers, GraphQL resolvers, and gRPC services — all three call usecase methods.

---

## 6. Error Types (`pkg/errors/`)

```go
var (
    ErrNotFound   = errors.New("not found")
    ErrValidation = errors.New("validation error")
    ErrConflict   = errors.New("conflict")
)
```

All domain errors wrap these sentinels. Delivery layers map them to protocol-appropriate codes:

| Domain Error | HTTP Status | gRPC Code |
|---|---|---|
| `ErrNotFound` | 404 | `codes.NotFound` |
| `ErrValidation` | 400 | `codes.InvalidArgument` |
| `ErrConflict` | 409 | `codes.AlreadyExists` |
| Other | 500 | `codes.Internal` |

---

## 7. Configuration

All environment variables (full table):

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | API listen port | `8080` |
| `MONGODB_URI` | MongoDB connection string | `mongodb://localhost:27017/cms` |
| `JWT_SECRET` | Secret for signing JWTs | *(required)* |
| `CLOUDINARY_CLOUD_NAME` | Cloudinary credentials | *(required for media)* |
| `CLOUDINARY_API_KEY` | Cloudinary credentials | *(required for media)* |
| `CLOUDINARY_API_SECRET` | Cloudinary credentials | *(required for media)* |
| `CONTENT_TYPES_DIR` | JSON content-type definitions dir | `content-types` |
| `STORAGE_PROVIDER` | Active media adapter (`s3` \| `cloudinary`) | `s3` |
| `COOKIE_SECURE` | Set `Secure` flag on refresh token cookie | `true` |
| `COOKIE_SAMESITE` | SameSite cookie policy (`none` \| `lax` \| `strict`) | `none` |
| `CORS_ORIGINS` | Comma-separated allowed origins | `http://localhost:5173` |
| `RATE_LIMIT_RPS` | Auth endpoint requests/sec per IP | `5` |
| `RATE_LIMIT_BURST` | Auth endpoint burst size | `10` |
| `BODY_LIMIT_BYTES` | Max request body size (bytes) | `10485760` (10 MB) |
| `DB_DRIVER` | Default database driver | `mongo` |
| `DB_HOST` | Database host | `localhost` |
| `DB_PORT` | Database port | `5432` |
| `DB_NAME` | Database name | `cms` |
| `DB_USERNAME` | Database username | `postgres` |
| `DB_PASSWORD` | Database password | *(required for postgres)* |
| `DB_SSL_MODE` | SSL mode for PostgreSQL | `disable` |
| `DB_USER` | DB adapter override for User entity | value of `DB_DRIVER` |
| `DB_CONTENT_TYPE` | DB adapter override for ContentType | value of `DB_DRIVER` |
| `DB_DOCUMENT` | DB adapter override for Document | value of `DB_DRIVER` |
| `DB_MEDIA` | DB adapter override for MediaAsset | value of `DB_DRIVER` |
| `GRPC_PORT` | gRPC server listen port | `9090` |
| `GRPC_SERVICES` | External gRPC services (`name=address,...`) | *(empty)* |
| `SUPPORTED_LOCALES` | Comma-separated locale codes | `en` |
| `GRAPHQL_PATH` | GraphQL endpoint path | `/graphql` |
| `VITE_API_URL` | API base URL for frontend (build-time) | *(empty for dev)* |

---

## 8. Database Selection — Per-Entity

Each entity can use either MongoDB or GORM (PostgreSQL/MySQL), configured independently via env vars. Both database clients are initialized at startup if needed.

```go
// main.go — repository factory pattern
var userRepo repository.UserRepository
if isPostgres(cfg.DB.EntityDB.User) {
    userRepo = gormdb.NewUserRepository(sqlDB)
} else {
    userRepo = mongodb.NewUserRepository(mongoDB)
}
// Repeat for ContentType, Document, MediaAsset, Role, Invite, AccessToken
```

---

## 9. Testing Strategy (Shared Rules)

- **Unit tests**: Every usecase tested in isolation with mock repositories (interface-based).
- **Integration tests**: Repository implementations tested against real databases (MongoDB in Docker, PostgreSQL via GORM/SQLite in-memory for unit tests).
- **HTTP handler tests**: Use `httptest` with `gin.SetMode(gin.TestMode)`.
- **Coverage target**: ≥ 80% on `internal/usecase/` packages.
- **No mocking the DB in usecase tests** — mock the repository interface instead.
- **GORM repositories must pass the same test cases as their MongoDB counterparts.**

Module-specific testing rules are in each module's spec.

---

## 10. Boundaries

| Rule | Detail |
|---|---|
| **Always** | Initialize both MongoDB and SQL connections at startup when configured |
| **Always** | Run `AutoMigrate` for GORM entities on startup |
| **Always** | Use `gin.SetMode(gin.TestMode)` in all test files |
| **Never** | Use MongoDB-specific logic outside `infrastructure/mongodb/` |
| **Never** | Use GORM-specific logic outside `infrastructure/gormdb/` |
| **Never** | Duplicate business logic across REST, GraphQL, and gRPC — all call usecase methods |
| **Never** | Hard-code SQL dialect-specific queries in GORM repositories |
| **Ask first** | Choosing between PostgreSQL and MySQL for production |

---

## 11. Changelog

| Date | Change | Source |
|------|--------|--------|
| v1.0 | Initial project setup with MongoDB + net/http | §1–§6 |
| v1.1 | Gin migration (Phase A) | §11.4 |
| v1.2 | GORM adapters + per-entity DB selection (Phase C) | §11.6 |
| v1.3 | gRPC server + client (Phase D) | §11.7 |
| v1.4 | Deployment to Render.com | §13 |
| v1.5 | All static tables use `gorm_id` (string UUID) as PK column name; `document_id` standardized to UUID v4 everywhere | sync-table-fields |
| v1.6 | Dynamic tables (`documents_{slug}`, `components_{slug}_{component}`) use per-field columns instead of JSON `data` blob | sync-table-fields |
| v1.7 | `EnsureCollection` accepts `[]FieldDefinition` and creates columns per field (DROP+CREATE strategy) | sync-table-fields |
| v1.8 | Field type mapping for dynamic columns: text/richtext → TEXT, media → VARCHAR (documentId FK), number → REAL, boolean → BOOLEAN, json → TEXT | sync-table-fields |
