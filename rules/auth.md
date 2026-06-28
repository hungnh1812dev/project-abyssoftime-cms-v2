# RULES — auth Module

**Scope:** Authentication (register/login/logout), JWT token lifecycle, password/email validation, role CRUD, permission enforcement, rate limiting.

---

## 1. Entity Rules

### 1.1 User Entity
- `ID` + `DocumentID`: both generated at registration time (UUID v4)
- `PasswordHash`: bcrypt — **NEVER** store plaintext
- `RoleID`: references `Role.DocumentID` (application-level FK, no DB constraint)
- `DisplayName`: **required** on register and invite accept — reject if empty
- `json:"-"` on `ID` and `PasswordHash` — **NEVER** expose in API responses

### 1.2 Role Entity
- `Slug`: unique, validated format `^[a-z0-9]+(?:-[a-z0-9]+)*$`, 1-63 chars
- `Name`: non-empty, max 100 chars
- `Permissions`: array of strings, each must be a valid `Permission` constant
- `Level`: integer, 1-99 for custom roles, fixed values for defaults
- `IsDefault`: marks the 4 built-in roles (super_admin, admin, editor, guest)

### 1.3 Permission Constants
- Format: `<module>:<action>` (e.g., `content:read`, `media:upload`)
- Adding new permissions: define constant in `entity/role.go` → **ask user first**
- Full list: `content:read`, `content:create`, `content:update`, `content:delete`, `content:publish`, `content:unpublish`, `media:read`, `media:upload`, `media:delete`, `users:manage`, `roles:manage`, `access_tokens:manage`, `content_types:read`

---

## 2. Password Validation Rules

### 2.1 Requirements (enforced in Register)
- Length: 8-72 characters (72 = bcrypt limit)
- Must contain at least one letter AND at least one digit
- Validated via `net/mail.ParseAddress` for email

### 2.2 Rejected Inputs
- Empty string
- Shorter than 8 characters
- Longer than 72 characters
- Only letters (no digit)
- Only digits (no letter)

### 2.3 Accepted Inputs
- `"Password1"`, `"my-s3cure-pass!"`, `"complexP@ss1"` etc.

---

## 3. Email Validation Rules

- Max 254 characters
- Must pass `net/mail.ParseAddress`
- Return 400 (not 500) for invalid format
- **NEVER** reflect the invalid email back in the error message

---

## 4. JWT Token Lifecycle

### 4.1 Access Token
- Short-lived: 15 minutes
- Stored in memory (React state/context)
- Sent via `Authorization: Bearer <token>` header
- Claims: `userId`, `role` (slug)

### 4.2 Refresh Token
- Long-lived: 7 days
- Stored in `HttpOnly` cookie
- Cookie settings: `HttpOnly: true`, `Secure: COOKIE_SECURE`, `SameSite: COOKIE_SAMESITE`
- Path: `/`

### 4.3 Token Operations
- `Register` → returns (accessToken, refreshToken)
- `Login` → returns (accessToken, refreshToken), sets cookie
- `Refresh` → validates refresh token, returns new access token
- `Logout` → clears refresh token cookie

---

## 5. Role System Rules

### 5.1 Default Roles (Seeded on First Startup)
| Role | Slug | Level | IsDefault |
|---|---|---|---|
| Super Admin | `super_admin` | 100 | true |
| Admin | `admin` | 80 | true |
| Editor | `editor` | 60 | true |
| Guest | `guest` | 20 | true |

- Seeded only when `HasAny()` returns false — idempotent
- **NEVER** delete default roles (`isDefault: true`)

### 5.2 Role Hierarchy
- `Level` determines hierarchy — higher level = more authority
- A user can **only** manage users/roles with a **strictly lower** level
- Custom roles use levels 1-99
- **NEVER** allow creating/editing a role with `level >= caller's role level`

### 5.3 Default Role Modification
- Only `permissions` can be changed on default roles
- `slug`, `name`, `level` are immutable on default roles
- Custom roles: all fields editable (within validation rules)

### 5.4 Role Deletion
- **NEVER** delete default roles
- **NEVER** delete roles assigned to any user (return `ErrConflict`)
- Only custom, unassigned roles can be deleted

---

## 6. Middleware Rules

### 6.1 GinAuth (`middleware/gin_auth.go`)
- Extracts JWT from `Authorization: Bearer <token>`
- Validates token, sets `userID` and `roleSlug` in Gin context
- Returns 401 for missing/invalid token

### 6.2 GinRequirePermission
- Checks user's role against required permission via RoleCache
- Returns 403 if role lacks permission
- Applied per-route in `router.go`

### 6.3 RoleCache (`middleware/role_cache.go`)
- In-memory cache: `slug → Role`
- Loaded on startup from all roles
- **MUST** be invalidated and reloaded on any role CRUD operation
- Methods: `Load`, `HasPermission`, `GetLevel`, `Get`

### 6.4 Rate Limiting (`middleware/ratelimit.go`)
- Token-bucket algorithm (golang.org/x/time/rate)
- Keyed by client IP
- Applied to auth endpoints only
- Config: `RATE_LIMIT_RPS` (default 5), `RATE_LIMIT_BURST` (default 10)
- **Ask first** before adjusting thresholds

---

## 7. API Contract Rules

### 7.1 Auth Routes (Public, Rate-Limited)
| Method | Route | Description |
|---|---|---|
| `GET` | `/auth/setup` | Check if admin exists |
| `POST` | `/auth/register` | Register new user |
| `POST` | `/auth/login` | Login, returns JWT + sets cookie |
| `POST` | `/auth/refresh` | Refresh access token |
| `POST` | `/auth/logout` | Clear refresh token cookie |

### 7.2 Role Routes
| Method | Route | Permission |
|---|---|---|
| `GET` | `/api/roles` | Any authenticated |
| `GET` | `/api/roles/:id` | Any authenticated |
| `POST` | `/api/roles` | `roles:manage` |
| `PUT` | `/api/roles/:id` | `roles:manage` |
| `DELETE` | `/api/roles/:id` | `roles:manage` |

### 7.3 Response Shapes
- Login: `{ "accessToken": "...", "user": { "documentId", "email", "roleId" } }`
- Refresh: `{ "accessToken": "..." }`
- Role CRUD: standard `Role` object
- **NEVER** include `passwordHash` in any response
- **NEVER** include internal `ID` (MongoDB `_id`) in responses

### 7.4 Cookie Handling
```go
c.SetSameSite(h.cookieSameSite)
c.SetCookie(RefreshCookieName, refresh, maxAge, "/", "", h.cookieSecure, true)
```
- `HttpOnly: true` always
- `Secure` and `SameSite` from config

---

## 8. Testing Rules (Auth-Specific)

### 8.1 Auth Usecase Tests
- Register: valid → success; duplicate email → conflict; weak password → validation; invalid email → validation
- Login: correct password → tokens; wrong password → error; unknown email → not found
- Refresh: valid → new access token; expired → error; invalid → error
- SetupStatus: no users → false; users → true

### 8.2 Role Usecase Tests
- SeedDefaults: seeds 4 when empty; skips when roles exist
- Create: valid → success; duplicate slug → conflict; level too high → error; invalid permission → error
- Update default role: only permissions changeable; slug/name/level rejected
- Delete: default → error; assigned → conflict; custom unassigned → success

### 8.3 Middleware Tests
- GinAuth: valid token → sets context; invalid → 401
- GinRequirePermission: has permission → next; lacks → 403
- RoleCache: Load → populated; HasPermission → correct; reload after update
- Rate limiting: under limit → 200; over limit → 429

---

## 9. Boundaries Summary

| Rule | Detail |
|---|---|
| **Always** | Validate password (8-72 chars, letter + digit) on Register |
| **Always** | Validate email via `net/mail.ParseAddress` on Register |
| **Always** | Set `Secure: true` on cookies in production |
| **Always** | Seed default roles on first startup if `HasAny()` is false |
| **Always** | Validate permissions against the allowed set — reject unknown strings |
| **Always** | Use `GinRequirePermission` for content/media/admin routes |
| **Always** | Cache role permissions in memory; invalidate on any role CRUD |
| **Always** | Return 400 (not 500) for invalid password or email format |
| **Never** | Allow deletion of default roles |
| **Never** | Allow creating/editing role with `level >= caller's level` |
| **Never** | Allow password shorter than 8 or longer than 72 |
| **Never** | Reflect user input in error messages verbatim |
| **Never** | Set `COOKIE_SAMESITE=lax` for cross-origin deployment |
| **Ask first** | Adding new permission actions to the allowed set |
| **Ask first** | Changing default role level values or range |
| **Ask first** | Adjusting rate limit thresholds |
