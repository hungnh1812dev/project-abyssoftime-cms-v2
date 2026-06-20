# SPEC — auth Module

## 1. Overview

The auth module handles user authentication (register, login, logout), JWT token lifecycle (access/refresh), password and email validation, role-based permission enforcement, and role CRUD management. It provides the security boundary for all protected API routes across other modules.

---

## 2. File Map

All paths relative to `apps/api/`.

```
internal/domain/entity/user.go                            # User entity
internal/domain/entity/user_test.go
internal/domain/entity/role.go                            # Role entity + Permission constants
internal/domain/repository/user_repository.go             # UserRepository interface
internal/domain/repository/role_repository.go             # RoleRepository interface
internal/domain/repository/mock/user_repository.go        # Mock for testing
internal/domain/repository/mock/role_repository.go        # Mock for testing
internal/usecase/auth/auth_usecase.go                     # Auth business logic
internal/usecase/auth/auth_usecase_test.go
internal/usecase/role/role_usecase.go                     # Role CRUD + seed + permissions
internal/usecase/role/role_usecase_test.go
internal/delivery/http/handler/auth_handler.go            # Gin auth handlers
internal/delivery/http/handler/auth_handler_test.go
internal/delivery/http/handler/role_handler.go            # Gin role CRUD handlers
internal/delivery/http/handler/role_handler_test.go
internal/delivery/http/middleware/auth.go                  # JWT auth middleware (net/http)
internal/delivery/http/middleware/auth_test.go
internal/delivery/http/middleware/gin_auth.go              # GinAuth + GinRequirePermission
internal/delivery/http/middleware/gin_auth_test.go
internal/delivery/http/middleware/role_cache.go            # In-memory role permission cache
internal/delivery/http/middleware/role_cache_test.go
internal/delivery/http/middleware/ratelimit.go             # Rate limiting (auth endpoints)
internal/delivery/http/middleware/ratelimit_test.go
internal/delivery/grpc/auth_service.go                    # gRPC auth service
internal/delivery/grpc/interceptor/auth.go                # gRPC JWT auth interceptor
internal/delivery/grpc/interceptor/auth_test.go
internal/infrastructure/mongodb/user_repository.go        # MongoDB User repo
internal/infrastructure/mongodb/user_repository_test.go
internal/infrastructure/mongodb/role_repository.go        # MongoDB Role repo
internal/infrastructure/mongodb/role_repository_test.go
internal/infrastructure/gormdb/user_repository.go         # GORM User repo
internal/infrastructure/gormdb/user_repository_test.go
internal/infrastructure/gormdb/role_repository.go         # GORM Role repo
internal/infrastructure/gormdb/role_repository_test.go
proto/cms/v1/auth.proto                                   # gRPC auth proto definition
proto/cms/v1/auth.pb.go                                   # Generated
proto/cms/v1/auth_grpc.pb.go                              # Generated
```

---

## 3. Entities

### User

```go
type User struct {
    ID           string    `bson:"_id,omitempty" gorm:"column:id;primaryKey"           json:"-"`
    DocumentID   string    `bson:"documentId"    gorm:"column:document_id;uniqueIndex" json:"documentId"`
    Email        string    `bson:"email"         gorm:"column:email;uniqueIndex"       json:"email"`
    PasswordHash string    `bson:"passwordHash"  gorm:"column:password_hash"           json:"-"`
    RoleID       string    `bson:"roleId"        gorm:"column:role_id;index"           json:"roleId"`
    CreatedAt    time.Time `bson:"createdAt"     gorm:"column:created_at"              json:"createdAt"`
}
```

- `RoleID` references Role's `DocumentID` (application-level FK, no DB constraint)
- Password stored as bcrypt hash

### Role

```go
type Permission string

const (
    PermContentRead       Permission = "content:read"
    PermContentCreate     Permission = "content:create"
    PermContentUpdate     Permission = "content:update"
    PermContentDelete     Permission = "content:delete"
    PermContentPublish    Permission = "content:publish"
    PermContentUnpublish  Permission = "content:unpublish"
    PermMediaRead         Permission = "media:read"
    PermMediaUpload       Permission = "media:upload"
    PermMediaDelete       Permission = "media:delete"
    PermUsersManage       Permission = "users:manage"
    PermRolesManage       Permission = "roles:manage"
    PermAccessTokenManage Permission = "access_tokens:manage"
    PermContentTypesRead  Permission = "content_types:read"
)

type Role struct {
    ID          string    `bson:"_id,omitempty"   gorm:"column:id;primaryKey"               json:"-"`
    DocumentID  string    `bson:"documentId"      gorm:"column:document_id;uniqueIndex"     json:"documentId"`
    Name        string    `bson:"name"            gorm:"column:name"                        json:"name"`
    Slug        string    `bson:"slug"            gorm:"column:slug;uniqueIndex"            json:"slug"`
    Permissions []string  `bson:"permissions"     gorm:"column:permissions;serializer:json"  json:"permissions"`
    Level       int       `bson:"level"           gorm:"column:level"                       json:"level"`
    IsDefault   bool      `bson:"isDefault"       gorm:"column:is_default"                  json:"isDefault"`
    CreatedAt   time.Time `bson:"createdAt"       gorm:"column:created_at"                  json:"createdAt"`
    UpdatedAt   time.Time `bson:"updatedAt"       gorm:"column:updated_at"                  json:"updatedAt"`
}
```

**Default roles (seeded on first startup if no roles exist):**

| Role | Slug | Level | Permissions |
|---|---|---|---|
| Super Admin | `super_admin` | 100 | All permissions |
| Admin | `admin` | 80 | All except `roles:manage`, `access_tokens:manage` |
| Editor | `editor` | 60 | `content:*`, `media:*`, `content_types:read` |
| Guest | `guest` | 20 | `content:read`, `media:read`, `content_types:read` |

`Level` is used for hierarchy comparison — a user can only manage users/roles with a lower level. Custom roles use values 1–99.

---

## 4. Repository Interfaces

### UserRepository

```go
type UserRepository interface {
    Create(ctx context.Context, user *entity.User) error
    FindByID(ctx context.Context, id string) (*entity.User, error)
    FindByEmail(ctx context.Context, email string) (*entity.User, error)
    FindAll(ctx context.Context) ([]*entity.User, error)
    Update(ctx context.Context, user *entity.User) error
    Delete(ctx context.Context, id string) error
    Count(ctx context.Context) (int64, error)
}
```

### RoleRepository

```go
type RoleRepository interface {
    Create(ctx context.Context, role *entity.Role) error
    FindByID(ctx context.Context, documentID string) (*entity.Role, error)
    FindBySlug(ctx context.Context, slug string) (*entity.Role, error)
    FindAll(ctx context.Context) ([]*entity.Role, error)
    Update(ctx context.Context, role *entity.Role) error
    Delete(ctx context.Context, documentID string) error
    HasAny(ctx context.Context) (bool, error)
}
```

---

## 5. Use Cases

### Auth UseCase (`usecase/auth/`)

| Method | Signature | Description |
|---|---|---|
| `Register` | `(ctx, email, password string) → (accessToken, refreshToken string, err)` | Create new user; validates email format, password strength; hashes password; assigns default role |
| `Login` | `(ctx, email, password string) → (accessToken, refreshToken string, err)` | Verify credentials; return JWT pair |
| `Refresh` | `(ctx, refreshToken string) → (accessToken string, err)` | Validate refresh token; issue new access token |
| `SetupStatus` | `(ctx) → (hasAdmin bool, err)` | Check if any user exists (for first-time setup) |

**Password validation (in Register):**
- 8–72 characters (72 = bcrypt limit)
- At least one letter and one digit

**Email validation (in Register):**
- Max 254 characters
- Must pass `net/mail.ParseAddress`

### Role UseCase (`usecase/role/`)

| Method | Signature | Description |
|---|---|---|
| `SeedDefaults` | `(ctx) → err` | Seed 4 default roles if `HasAny()` returns false. Idempotent. |
| `Create` | `(ctx, role, callerRoleSlug) → (*Role, err)` | Create custom role; validate permissions, level < caller's level |
| `FindAll` | `(ctx) → ([]*Role, err)` | Return all roles sorted by level descending |
| `FindByID` | `(ctx, documentID) → (*Role, err)` | Get role by documentId |
| `Update` | `(ctx, role, callerRoleSlug) → (*Role, err)` | Update role; default roles only allow permission changes |
| `Delete` | `(ctx, documentID, callerRoleSlug) → err` | Delete custom role; reject if default or assigned to any user |

**Validation rules for Create/Update:**
- `slug`: `^[a-z0-9]+(?:-[a-z0-9]+)*$`, 1–63 chars
- `name`: non-empty, max 100 chars
- `permissions`: each entry must be a valid Permission constant
- `level`: 1–99
- Cannot create/edit a role with `level >= caller's role level`
- Default roles: only `permissions` can be changed (slug, name, level immutable)

---

## 6. API Contracts

### REST — Auth Routes (Public)

| Method | Route | Description |
|---|---|---|
| `GET` | `/auth/setup` | Check if admin exists (first-time setup) |
| `POST` | `/auth/register` | Register new user |
| `POST` | `/auth/login` | Login, returns JWT pair + sets refresh cookie |
| `POST` | `/auth/refresh` | Refresh access token via cookie |
| `POST` | `/auth/logout` | Clear refresh token cookie |

Rate limited: all auth routes subject to `RATE_LIMIT_RPS` / `RATE_LIMIT_BURST`.

**Login response:**
```json
{
  "accessToken": "eyJ...",
  "user": { "documentId": "...", "email": "...", "roleId": "..." }
}
```

**Cookie handling:**
```go
c.SetSameSite(h.cookieSameSite)
c.SetCookie(RefreshCookieName, refresh, maxAge, "/", "", h.cookieSecure, true)
```

- `HttpOnly: true` (always)
- `Secure`: from `COOKIE_SECURE` env (default `true`)
- `SameSite`: from `COOKIE_SAMESITE` env (default `none` for cross-origin deployment)

### REST — Role Routes

| Method | Route | Required Permission | Response |
|---|---|---|---|
| `GET` | `/api/roles` | Any authenticated user | `Role[]` |
| `GET` | `/api/roles/:id` | Any authenticated user | `Role` |
| `POST` | `/api/roles` | `roles:manage` | `Role` (201) |
| `PUT` | `/api/roles/:id` | `roles:manage` | `Role` |
| `DELETE` | `/api/roles/:id` | `roles:manage` | `204` |

### gRPC — AuthService

```protobuf
service AuthService {
    rpc Login(LoginRequest) returns (LoginResponse);
    rpc Register(RegisterRequest) returns (RegisterResponse);
    rpc Refresh(RefreshRequest) returns (RefreshResponse);
    rpc SetupStatus(SetupStatusRequest) returns (SetupStatusResponse);
}
```

---

## 7. Middleware

### GinAuth (`middleware/gin_auth.go`)

Extracts JWT from `Authorization: Bearer <token>` header, validates it, sets `userID` and `roleSlug` in Gin context.

```go
func GinAuth() gin.HandlerFunc
```

### GinRequirePermission (`middleware/gin_auth.go`)

Checks if the authenticated user's role has the required permission via the role cache.

```go
func GinRequirePermission(cache *RoleCache, permission string) gin.HandlerFunc
```

Returns 403 if the user's role lacks the permission.

### RoleCache (`middleware/role_cache.go`)

In-memory cache of role permissions, loaded on startup and invalidated on role CRUD.

```go
type RoleCache struct { ... }
func NewRoleCache() *RoleCache
func (c *RoleCache) Load(roles []*entity.Role)
func (c *RoleCache) HasPermission(roleSlug, permission string) bool
func (c *RoleCache) GetLevel(roleSlug string) int
func (c *RoleCache) Get(roleSlug string) *entity.Role
```

### Rate Limiting (`middleware/ratelimit.go`)

In-memory token-bucket rate limiter using `golang.org/x/time/rate`, keyed by client IP. Applied to auth endpoints only.

---

## 8. JWT Token Lifecycle

- **Access token**: short-lived (15 min), stored in memory (React state/context)
- **Refresh token**: long-lived (7 days), stored in `HttpOnly` cookie
- All protected routes require `Authorization: Bearer <access_token>`
- JWT claims include `userId` and `role` (slug)

**Permission resolution flow:**
1. On startup, load all roles into RoleCache (`slug → Role`)
2. On role create/update/delete, invalidate and reload cache
3. Auth middleware reads role slug from JWT → looks up permissions from cache
4. Route-level `GinRequirePermission` checks the permission

---

## 9. gRPC Auth Interceptor

```go
func AuthUnaryInterceptor(jwtSecret string) grpc.UnaryServerInterceptor
```

- Skips auth for public methods (configurable)
- Reads bearer token from gRPC metadata
- Validates JWT and injects `userID` into context

---

## 10. Startup Migration (One-Time)

When `roles` table/collection is empty on startup:
1. Seed the 4 default roles
2. For each existing user with legacy `role` string field, look up matching default role by slug and set `roleId`

---

## 11. Testing

**Auth usecase (`auth_usecase_test.go`):**
- Register: valid credentials → success; duplicate email → conflict; weak password → validation error; invalid email → validation error
- Login: correct password → tokens; wrong password → error; unknown email → not found
- Refresh: valid token → new access token; expired/invalid → error
- SetupStatus: no users → false; users exist → true

**Password validation:**
- Rejected: `""`, `"short"`, `"noletter1234"`, `"nodigithere"`, `"a".repeat(73)`
- Accepted: `"Password1"`, `"my-s3cure-pass!"`

**Email validation:**
- Rejected: `""`, `"notanemail"`, `"@@@"`, `"a".repeat(255)+"@x.com"`
- Accepted: `"user@example.com"`, `"name+tag@domain.co"`

**Role usecase (`role_usecase_test.go`):**
- SeedDefaults: seeds 4 when empty, skips when roles exist
- Create: valid → success; duplicate slug → conflict; level too high → error; invalid permission → error
- Update default role: only permissions changeable; slug/name/level changes rejected
- Delete: default role → error; assigned role → conflict; custom unassigned → success

**Role handler (`role_handler_test.go`):**
- Full CRUD status codes and permission checks

**Middleware (`role_cache_test.go`, `gin_auth_test.go`):**
- GinRequirePermission: user with permission → next; without → 403
- RoleCache: Load populates; HasPermission returns correct results; cache reload after update

**Rate limiting (`ratelimit_test.go`):**
- Under limit → 200; over limit → 429

---

## 12. Boundaries

| Rule | Detail |
|---|---|
| **Always** | Validate password (8–72 chars, at least one letter + one digit) on Register |
| **Always** | Validate email format via `net/mail.ParseAddress` on Register |
| **Always** | Set `Secure: true` on cookies in production (`COOKIE_SECURE=true`) |
| **Always** | Seed default roles on first startup; skip if `HasAny()` returns true |
| **Always** | Validate permissions against the allowed set — reject unknown strings |
| **Always** | Use `GinRequirePermission` for content/media/admin routes |
| **Always** | Cache role permissions in memory; invalidate and reload on any role CRUD |
| **Always** | Return 400 (not 500) for invalid password or email format |
| **Never** | Allow deletion of default roles (`isDefault: true`) |
| **Never** | Allow a user to create/edit a role with `level >= their own role's level` |
| **Never** | Allow password shorter than 8 or longer than 72 characters |
| **Never** | Reflect user input in error messages verbatim (prevents information leakage) |
| **Never** | Set `COOKIE_SAMESITE=lax` for cross-origin deployment |
| **Ask first** | Adding new permission actions to the allowed set |
| **Ask first** | Changing default role level values or the level range (1–99) |
| **Ask first** | Adjusting rate limit thresholds for production traffic patterns |

---

## 13. Changelog

| Date | Change | Source |
|------|--------|--------|
| v1.0 | Basic auth (login/register/refresh) with hardcoded admin/guest roles | §4 Auth |
| v1.1 | Password + email validation, secure cookie flag | §9.4, §9.6, §9.7 |
| v1.2 | Rate limiting on auth endpoints | §9.8 |
| v1.3 | Gin middleware migration (GinAuth, GinRequirePermission) | §11.4 |
| v1.4 | Dynamic role system with permission-based access control | §12 |
| v1.5 | Cross-origin cookie config (SameSite=None) for Render deployment | §13.4.2 |
