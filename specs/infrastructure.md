# SPEC — Infrastructure Module

## 1. Overview

The infrastructure module covers the application's routing setup, gRPC client management for external microservices, development and build commands, and deployment configuration. It provides the glue that connects delivery layers (HTTP, gRPC) to the rest of the application, manages external service connections, and documents the operational aspects of running the CMS in development and production environments.

---

## 2. File Map

All paths relative to `apps/api/`.

```
internal/delivery/http/router.go                     # Gin router setup (route registration)
internal/delivery/http/handler/response.go           # Shared response helpers (writeErr, writeJSON)
internal/delivery/http/handler/response_test.go
internal/delivery/grpc/server.go                     # gRPC server setup
internal/delivery/grpc/errors.go                     # gRPC error mapping
internal/grpcclient/client.go                        # gRPC client connection manager
internal/grpcclient/client_test.go
cmd/server/main.go                                   # Application entry point, dependency wiring
Dockerfile
```

---

## 3. Router Setup

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

## 4. gRPC Client (`internal/grpcclient/`)

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

## 5. Commands

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

## 6. Deployment (Render.com)

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

## 7. Boundaries

| Rule | Detail |
|---|---|
| **Always** | Keep REST endpoint paths and response shapes identical after any framework migration |
| **Always** | Map domain errors to protocol-appropriate codes in every delivery layer |
| **Ask first** | Adding new gRPC client service connections |
| **Ask first** | Changing the gRPC port default |

---

## 8. Changelog

| Date | Change | Source |
|------|--------|--------|
| v1.0 | Initial project setup with MongoDB + net/http | §1–§6 |
| v1.1 | Gin migration (Phase A) | §11.4 |
| v1.3 | gRPC server + client (Phase D) | §11.7 |
| v1.4 | Deployment to Render.com | §13 |
