# SPEC — core Module

## 1. Overview

The core module provides the shared foundation for all other modules: entity type definitions, repository interfaces, error types, configuration, JWT utilities, database clients, cross-cutting middleware, and the application entry point. Every other module depends on core but core never depends on any domain module.

---

## 2. File Map

All paths relative to `apps/api/`.

```
cmd/server/main.go                                  # Application entry point, dependency wiring
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
internal/delivery/http/router.go                     # Gin router setup (route registration)
internal/delivery/http/handler/response.go           # Shared response helpers (writeErr, writeJSON)
internal/delivery/http/handler/response_test.go
internal/delivery/http/middleware/cors.go             # CORS middleware
internal/delivery/http/middleware/cors_test.go
internal/delivery/http/middleware/bodylimit.go        # Request body size limit
internal/delivery/http/middleware/bodylimit_test.go
internal/delivery/http/middleware/security_headers.go # Security response headers
internal/delivery/http/middleware/security_headers_test.go
internal/delivery/grpc/server.go                     # gRPC server setup
internal/delivery/grpc/errors.go                     # gRPC error mapping
internal/grpcclient/client.go                        # gRPC client connection manager
internal/grpcclient/client_test.go
pkg/errors/errors.go                                 # Domain error types (ErrNotFound, ErrValidation, etc.)
pkg/errors/errors_test.go
pkg/jwt/jwt.go                                       # JWT sign/validate utilities
pkg/jwt/jwt_test.go
go.mod
go.sum
Dockerfile
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
├── content-types/                        # JSON schema-as-code definitions → see specs/content.md
├── proto/cms/v1/                         # gRPC protocol buffer definitions
├── graphql/dynamic/                      # Dynamic GraphQL schema builder → see specs/content.md
├── internal/
│   ├── config/                           # App configuration
│   ├── domain/
│   │   ├── entity/                       # All entity structs
│   │   └── repository/                   # All repository interfaces (+ mock/)
│   ├── usecase/                          # Business logic (DB-agnostic)
│   │   ├── auth/                         # → see specs/auth.md
│   │   ├── role/                         # → see specs/auth.md
│   │   ├── content_type/                 # → see specs/content.md
│   │   ├── document/                     # → see specs/content.md
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
│   │   │   ├── middleware/               # Gin middleware
│   │   │   └── router.go                 # Route registration
│   │   └── grpc/                         # gRPC service implementations
│   └── grpcclient/                       # gRPC client adapters
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

## 9. Middleware (Cross-Cutting)

### CORS (`middleware/cors.go`)
- Checks `Origin` header against `CORS_ORIGINS` whitelist
- Sets `Access-Control-Allow-Origin`, `Allow-Methods`, `Allow-Headers`, `Allow-Credentials`
- Handles `OPTIONS` preflight with 204
- Never uses `Access-Control-Allow-Origin: *`

### Body Size Limit (`middleware/bodylimit.go`)
- Applied globally with `BODY_LIMIT_BYTES` default (10 MB)
- Media upload exempt (uses its own multipart limit)
- Uses `http.MaxBytesReader`

### Security Headers (`middleware/security_headers.go`)
Sets on every response:
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 0`
- `Referrer-Policy: strict-origin-when-cross-origin`
- `Content-Security-Policy: default-src 'self'`

---

## 10. Router Setup

Route registration is centralized in `internal/delivery/http/router.go`. The `SetupRouter` function accepts a `RouterConfig` struct with all handler instances and returns a configured `*gin.Engine`.

```go
type RouterConfig struct {
    AuthHandler        *handler.AuthHandler
    CTHandler          *handler.ContentTypeHandler
    DocHandler         *handler.DocumentHandler
    MediaHandler       *handler.MediaHandler
    LocaleHandler      *handler.LocaleHandler
    UserHandler        *handler.UserHandler
    InviteHandler      *handler.InviteHandler
    AccessTokenHandler *handler.AccessTokenHandler
    RoleHandler        *handler.RoleHandler
    RoleCache          *middleware.RoleCache
    GraphQLHandler     http.Handler
    GraphQLPath        string
    CORSOrigins        []string
}
```

Route groups and permission requirements are documented in each module's spec.

---

## 11. gRPC Client (`internal/grpcclient/`)

For calling external microservices:

```go
type ClientManager struct { conns map[string]*grpc.ClientConn }
func NewClientManager() *ClientManager
func (m *ClientManager) Connect(ctx, serviceName, address string, opts ...grpc.DialOption) error
func (m *ClientManager) GetConnection(serviceName string) (*grpc.ClientConn, error)
func (m *ClientManager) Close() error
```

Configured via `GRPC_SERVICES` env var: `name=address,name=address`.

---

## 12. Commands

### Native Development
| Command | Description |
|---|---|
| `make mongo-start` | Start MongoDB container (port 27017, persistent volume) |
| `make mongo-stop` | Stop the MongoDB container |
| `make dev` | Start API + web in parallel |
| `make dev-api` | Start Go API server only |
| `make dev-web` | Start Vite dev server only |
| `make test-api` | `go test ./...` inside `apps/api` |
| `make test-web` | `vitest run` inside `apps/web` |

### Backend Only
| Command | Description |
|---|---|
| `go run ./cmd/server` | Start the API server |
| `go test ./...` | Run all tests |
| `go build -o bin/server ./cmd/server` | Compile production binary |
| `go vet ./...` | Static analysis |

### Docker Compose
| Command | Description |
|---|---|
| `docker-compose -f docker-compose.yml -f docker-compose.dev.yml up --build` | Full stack with hot reload |
| `docker-compose down` | Stop and remove all containers |

---

## 13. Testing Strategy (Shared Rules)

- **Unit tests**: Every usecase tested in isolation with mock repositories (interface-based).
- **Integration tests**: Repository implementations tested against real databases (MongoDB in Docker, PostgreSQL via GORM/SQLite in-memory for unit tests).
- **HTTP handler tests**: Use `httptest` with `gin.SetMode(gin.TestMode)`.
- **Coverage target**: ≥ 80% on `internal/usecase/` packages.
- **No mocking the DB in usecase tests** — mock the repository interface instead.
- **GORM repositories must pass the same test cases as their MongoDB counterparts.**

Module-specific testing rules are in each module's spec.

---

## 14. Deployment (Render.com)

Two separate Render.com services:
- **API**: Web Service (native Go), `abyssoftime-cms-api`, port 8080
- **Web**: Static Site, `abyssoftime-cms-web`, React/Vite SPA
- **Database**: Supabase PostgreSQL (external)
- **Media**: Cloudinary (external)

Key deployment rules:
- `COOKIE_SAMESITE=none` + `COOKIE_SECURE=true` for cross-origin deployment
- `DB_SSL_MODE=require` for Supabase
- `CORS_ORIGINS` set to the Static Site URL (explicit, never `*`)
- `VITE_API_URL` set at build time on Static Site
- SPA rewrite rule (`/* → /index.html`) on Static Site
- Deploy hooks from CI (not auto-deploy)
- Keep-alive workflow pings `/health` every 14 minutes

Full deployment walkthrough: see SPEC.md §13.

---

## 15. Boundaries

| Rule | Detail |
|---|---|
| **Always** | Keep REST endpoint paths and response shapes identical after any framework migration |
| **Always** | Map domain errors to protocol-appropriate codes in every delivery layer |
| **Always** | Initialize both MongoDB and SQL connections at startup when configured |
| **Always** | Run `AutoMigrate` for GORM entities on startup |
| **Always** | Use `gin.SetMode(gin.TestMode)` in all test files |
| **Always** | Apply body size limit globally; media upload uses its own multipart limit |
| **Always** | Set security headers on every response |
| **Never** | Use MongoDB-specific logic outside `infrastructure/mongodb/` |
| **Never** | Use GORM-specific logic outside `infrastructure/gormdb/` |
| **Never** | Duplicate business logic across REST, GraphQL, and gRPC — all call usecase methods |
| **Never** | Use `Access-Control-Allow-Origin: *` — always explicit origin whitelist |
| **Never** | Hard-code SQL dialect-specific queries in GORM repositories |
| **Ask first** | Choosing between PostgreSQL and MySQL for production |
| **Ask first** | Adding new gRPC client service connections |
| **Ask first** | Changing the gRPC port default |

---

## 16. Changelog

| Date | Change | Source |
|------|--------|--------|
| v1.0 | Initial project setup with MongoDB + net/http | §1–§6 |
| v1.1 | Gin migration (Phase A) | §11.4 |
| v1.2 | GORM adapters + per-entity DB selection (Phase C) | §11.6 |
| v1.3 | gRPC server + client (Phase D) | §11.7 |
| v1.4 | Deployment to Render.com | §13 |
