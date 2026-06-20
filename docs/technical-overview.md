# Technical Overview — personal-cms

## 1. Tech Stack

### Backend

| Technology | Role | Version Context |
|---|---|---|
| **Go** | Primary backend language | Statically typed, compiled, fast cold starts |
| **Gin** | HTTP framework | REST API routing, middleware pipeline |
| **GORM** | SQL ORM | PostgreSQL/MySQL abstraction, auto-migration |
| **MongoDB Go Driver** | NoSQL client | Document storage, per-content-type collections |
| **gRPC + Protocol Buffers** | RPC framework | Typed inter-service communication |
| **gqlgen** | GraphQL server | Dynamic schema generation per content-type |
| **bcrypt** | Password hashing | Industry-standard adaptive hashing |
| **golang.org/x/time/rate** | Rate limiting | Token-bucket algorithm for auth endpoints |

### Frontend

| Technology | Role |
|---|---|
| **React 18** | UI library |
| **TypeScript (strict mode)** | Type-safe frontend code |
| **Vite** | Build tool and dev server |
| **TanStack Query (React Query)** | Server state management, caching, mutations |
| **Shadcn UI** | Accessible, composable component library |
| **TailwindCSS** | Utility-first CSS framework |
| **React Hook Form** | Form state management and validation |
| **React Router** | Client-side routing (SPA) |

### Infrastructure & Deployment

| Technology | Role |
|---|---|
| **Render.com** | Hosting — Go Web Service (API) + Static Site (SPA) |
| **Supabase PostgreSQL** | Production relational database |
| **MongoDB** | Development/alternative document database |
| **Cloudinary** | Media asset storage and transformation (production) |
| **AWS S3** | Alternative media storage backend |
| **GitHub Actions** | CI pipeline (`go vet` + `go test` + `npm lint` + `npm build`) |
| **Docker / Docker Compose** | Local development environment |

---

## 2. System Architecture

### Deployment Topology

```
                    Internet
                       |
        +--------------+--------------+
        |                             |
  [Render Web Service]         [Render Static Site]
   Go binary (API)              React SPA (Vite)
   Port 8080                    CDN-served
        |                             |
        +-------- CORS + JWT ---------+
        |
   +----+----+
   |         |
[Supabase] [Cloudinary]
PostgreSQL  Media CDN
```

- **Two-service split**: API and web are independently deployed and scaled. The SPA communicates with the API via REST over HTTPS with JWT authentication.
- **Cross-origin auth**: Refresh tokens are stored in `HttpOnly` cookies with `SameSite=None; Secure` for cross-origin deployment. Access tokens are held in memory (React state).
- **Stateless API**: No server-side sessions. JWT tokens carry user identity and role claims. The API can be horizontally scaled without shared session state.

### Software Architecture (Backend)

```
┌─────────────────────────────────────────────────────┐
│                   Delivery Layer                     │
│  ┌─────────┐  ┌──────────┐  ┌────────────────────┐  │
│  │  Gin    │  │  gRPC    │  │  GraphQL (gqlgen)  │  │
│  │ Handlers│  │ Services │  │  Dynamic Resolvers │  │
│  └────┬────┘  └────┬─────┘  └────────┬───────────┘  │
│       │            │                 │               │
│  ┌────┴────────────┴─────────────────┴────┐          │
│  │          Middleware (Gin)               │          │
│  │  Auth · CORS · RateLimit · BodyLimit   │          │
│  │  SecurityHeaders · RequirePermission   │          │
│  └────────────────────────────────────────┘          │
└───────────────────────┬─────────────────────────────┘
                        │ calls
┌───────────────────────┴─────────────────────────────┐
│                   Use Case Layer                     │
│  auth · role · document · content_type              │
│  media · user · invite · access_token               │
│                                                      │
│  Business logic only. Depends on domain interfaces.  │
│  No database-specific code. No HTTP/gRPC awareness.  │
└───────────────────────┬─────────────────────────────┘
                        │ depends on
┌───────────────────────┴─────────────────────────────┐
│                   Domain Layer                       │
│  entity/     → User, Role, Document, ContentType,   │
│                MediaAsset, Invite, AccessToken       │
│  repository/ → Interfaces (UserRepository,          │
│                DocumentRepository, StorageAdapter…)  │
│  pkg/errors  → ErrNotFound, ErrValidation…          │
│  pkg/jwt     → Token generation and validation      │
└───────────────────────┬─────────────────────────────┘
                        │ implemented by
┌───────────────────────┴─────────────────────────────┐
│               Infrastructure Layer                   │
│  mongodb/     → MongoDB repository implementations  │
│  gormdb/      → GORM (PostgreSQL) implementations   │
│  cloudinary/  → Cloudinary storage adapter           │
│  s3/          → S3 storage adapter                   │
└─────────────────────────────────────────────────────┘
```

**Clean Architecture** — strict dependency rule: inner layers never import outer layers. All cross-layer communication flows through interfaces defined in `domain/repository/`.

### Module Decomposition

```
core ← auth     (user/role repos, JWT utilities)
core ← content  (content-type/document repos, config)
core ← media    (media-asset repo, storage adapter interface)
core ← admin    (invite/access-token repos)

auth ← admin    (user management uses role usecase for assignment)
content ← media (document delete cascades to media assets)
content ← auth  (auth middleware protects content routes)
```

All cross-module communication goes through interfaces in `domain/repository/`. No module imports another module's usecase directly — dependencies are injected in `main.go`.

---

## 3. Key Technical Patterns & Skills

### Schema-as-Code

Content types are defined as JSON files in `content-types/`. On every API startup, a sync engine reconciles these definitions against the database — creating, updating, or deleting content types and their associated collections/tables. This makes the CMS structure version-controlled and reproducible.

### Draft/Publish Workflow

Every content entry exists as two separate records — a `draft` and a `published` version. Save only touches draft; publish copies draft data to the published record. Entry status (`draft`, `modified`, `published`) is computed at read time, never stored. Public API only returns published records.

### Dual-Database Repository Pattern

Each entity can independently use either MongoDB or PostgreSQL (via GORM), configured per-entity through environment variables. Both database clients initialize at startup. Repository interfaces in `domain/repository/` are database-agnostic — MongoDB and GORM implementations pass the same test suites.

### Multi-Protocol API Surface

The same business logic is exposed through three protocols:
- **REST (Gin)** — primary API for the frontend SPA
- **gRPC** — typed RPC for potential inter-service communication
- **GraphQL** — dynamically generated per content-type for flexible public queries

All three protocols delegate to the same usecase methods — no business logic is duplicated across delivery layers.

### Dynamic GraphQL Schema Generation

GraphQL types, queries, and mutations are generated at runtime from content-type definitions. Each content-type produces its own GraphQL type and CRUD operations. Resolvers are thin wrappers that delegate to the document usecase.

### JWT Authentication with Refresh Token Rotation

- Short-lived access tokens (15 min) held in memory
- Long-lived refresh tokens (7–30 days) stored in `HttpOnly` cookies
- On each `/auth/refresh` call, a new refresh token is issued (rotation), preventing token reuse attacks
- JWT claims carry `userId` and `role` (slug) for stateless permission checking

### Role-Based Access Control (RBAC)

Dynamic role system with granular permissions (`content:read`, `media:upload`, `roles:manage`, etc.). Roles are cached in memory at startup and invalidated on CRUD operations. Permission checks happen at the middleware level via `GinRequirePermission`.

### Per-Entity Database Selection

```
DB_ENTITY_USER=postgres
DB_ENTITY_DOCUMENT=mongo
DB_ENTITY_MEDIA=postgres
```

Each domain entity can target a different database engine. The repository factory in `main.go` wires the correct implementation based on configuration. This allows gradual migration between databases without all-or-nothing switches.

### Content-Type Field Projection

Paginated collection lists return only the fields specified in `listFields` (or the first 3 fields by default), reducing payload size. Full document data is available through single-document endpoints.

---

## 4. Why These Choices

### Go over Node.js/Python

- **Cold start performance**: Go compiles to a single binary with sub-second startup — critical for Render.com's free-tier spin-down behavior and schema sync on every boot.
- **Type safety without runtime overhead**: Compile-time type checking catches interface mismatches (e.g., repository implementations) without reflection-heavy frameworks.
- **Concurrency model**: Goroutines handle concurrent HTTP/gRPC/GraphQL serving with minimal boilerplate compared to async/await patterns.
- **Single binary deployment**: No runtime dependencies, no `node_modules`, no interpreter — simplifies Docker images and Render.com deployment.

### Clean Architecture over MVC

- **Database independence**: The dual MongoDB/PostgreSQL support requires business logic that doesn't know which database it's talking to. Clean Architecture enforces this through interface boundaries.
- **Testability**: Usecases are tested with mock repositories — no database setup needed. Infrastructure tests run against real databases. This separation gives fast unit tests and thorough integration tests.
- **Multi-protocol delivery**: Adding gRPC and GraphQL alongside REST was straightforward because handlers are thin adapters over usecases, not the home of business logic.

### Gin over net/http

- **Middleware chaining**: Auth, CORS, rate limiting, body limits, security headers, and permission checks compose cleanly in Gin's middleware pipeline.
- **Route grouping**: Permission-protected routes are grouped under authenticated middleware, reducing per-route boilerplate.
- **Context helpers**: `c.ShouldBindJSON`, `c.SetCookie`, `c.JSON` reduce handler verbosity compared to raw `net/http`.

### TanStack Query over Redux/Zustand

- **Server state is not client state**: Content entries, media assets, and user lists are server-owned data. TanStack Query treats them as cached server state with automatic invalidation, refetching, and stale-while-revalidate — patterns that would require manual implementation in Redux.
- **Mutation-invalidation coupling**: Every `useMutation` invalidates affected query keys on success. This eliminates an entire category of state-sync bugs where the UI shows stale data after a write.
- **No boilerplate**: No reducers, no action creators, no selectors. A query hook is 5 lines; a mutation hook is 10.

### Shadcn UI over Material UI/Ant Design

- **Copy-paste ownership**: Components are copied into the project, not imported from a package. This means full control over styling and behavior without fighting a component library's opinion.
- **TailwindCSS native**: No CSS-in-JS runtime, no theme provider overhead. Components use Tailwind utilities directly.
- **Accessible by default**: Built on Radix UI primitives with proper ARIA attributes, keyboard navigation, and focus management.

### MongoDB + PostgreSQL (dual) over single database

- **Development speed**: MongoDB's schemaless nature makes rapid prototyping fast — no migrations needed when content-type schemas change.
- **Production reliability**: PostgreSQL (via Supabase) provides ACID transactions, relational integrity, and managed backups for production deployment.
- **Gradual migration path**: Per-entity database selection allows moving entities one at a time from MongoDB to PostgreSQL without a big-bang migration.

### Schema-as-Code over Admin UI schema builder

- **Version control**: Content-type definitions live in JSON files tracked by git. Schema changes go through PRs with code review, not through a UI that's hard to audit.
- **Reproducibility**: Any environment can be bootstrapped from the same JSON definitions. No database seeding scripts, no manual setup.
- **Simplicity**: No need for a complex schema builder UI with drag-and-drop, validation rules, and field type selectors. The JSON format is simple and well-documented.

### Separate API + SPA over SSR (Next.js)

- **Headless CMS pattern**: The API serves any client — React SPA, mobile app, static site generator, or third-party integration. Coupling to Next.js would limit this flexibility.
- **Independent deployment**: API and web scale and deploy independently. A frontend-only change doesn't require restarting the Go server.
- **Technology independence**: The frontend can be replaced (e.g., with a mobile app or different framework) without touching the backend.

### JWT over Session-based Auth

- **Stateless scaling**: No session store to share across API instances. The JWT itself carries all claims needed for authentication and authorization.
- **Cross-origin friendly**: Works naturally with the separate API + SPA deployment model, where cookies carry the refresh token and the access token travels in the `Authorization` header.
- **gRPC compatibility**: JWT tokens work identically across REST and gRPC via metadata headers, unlike cookie-based sessions that are HTTP-specific.

### Cloudinary + S3 behind StorageAdapter over single provider

- **Provider flexibility**: Cloudinary offers automatic image transformations and thumbnails; S3 offers raw storage at lower cost. The adapter interface allows choosing per environment.
- **No vendor lock-in**: Switching providers requires implementing one interface (`Upload` + `Delete`), not rewriting media handling logic.
- **Environment-appropriate**: Cloudinary for production (transformations, CDN); S3 or local storage for development/testing.
