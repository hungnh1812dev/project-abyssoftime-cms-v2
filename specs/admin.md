# SPEC — admin Module

## 1. Overview

The admin module handles user management (list, update role, delete), the invite system (create invite links, accept invitations), and access token management (create/revoke API tokens). It provides the administrative surface for team management within the CMS.

---

## 2. File Map

All paths relative to `apps/api/`.

```
internal/domain/entity/invite.go                            # Invite entity
internal/domain/entity/access_token.go                      # AccessToken entity
internal/domain/repository/invite_repository.go             # InviteRepository interface
internal/domain/repository/access_token_repository.go       # AccessTokenRepository interface
internal/domain/repository/mock/invite_repository.go        # Mock for testing
internal/domain/repository/mock/access_token_repository.go  # Mock for testing
internal/usecase/user/user_usecase.go                       # User management logic
internal/usecase/user/user_usecase_test.go
internal/usecase/invite/invite_usecase.go                   # Invite business logic
internal/usecase/invite/invite_usecase_test.go
internal/usecase/access_token/access_token_usecase.go       # Access token logic
internal/usecase/access_token/access_token_usecase_test.go
internal/delivery/http/handler/user_handler.go              # Gin user handlers
internal/delivery/http/handler/invite_handler.go            # Gin invite handlers
internal/delivery/http/handler/access_token_handler.go      # Gin access token handlers
internal/infrastructure/mongodb/invite_repository.go        # MongoDB Invite repo
internal/infrastructure/mongodb/access_token_repository.go  # MongoDB AccessToken repo
internal/infrastructure/gormdb/invite_repository.go         # GORM Invite repo
internal/infrastructure/gormdb/access_token_repository.go   # GORM AccessToken repo
```

---

## 3. Entities

### Invite

```go
type Invite struct {
    ID        string    `bson:"_id,omitempty" gorm:"column:id;primaryKey"     json:"-"`
    Token     string    `bson:"token"        gorm:"column:token;uniqueIndex" json:"token"`
    Email     string    `bson:"email"        gorm:"column:email"             json:"email"`
    RoleID    string    `bson:"roleId"       gorm:"column:role_id"           json:"roleId"`
    CreatedBy string    `bson:"createdBy"    gorm:"column:created_by"        json:"createdBy"`
    CreatedAt time.Time `bson:"createdAt"    gorm:"column:created_at"        json:"createdAt"`
    ExpiresAt time.Time `bson:"expiresAt"    gorm:"column:expires_at"        json:"expiresAt"`
    Used      bool      `bson:"used"         gorm:"column:used"              json:"used"`
}
```

### AccessToken

```go
type AccessToken struct {
    ID          string    `bson:"_id,omitempty" gorm:"column:id;primaryKey"           json:"-"`
    DocumentID  string    `bson:"documentId"    gorm:"column:document_id;uniqueIndex" json:"documentId"`
    Name        string    `bson:"name"          gorm:"column:name"                    json:"name"`
    TokenHash   string    `bson:"tokenHash"     gorm:"column:token_hash"              json:"-"`
    Prefix      string    `bson:"prefix"        gorm:"column:prefix"                  json:"prefix"`
    CreatedBy   string    `bson:"createdBy"     gorm:"column:created_by"              json:"createdBy"`
    CreatedAt   time.Time `bson:"createdAt"     gorm:"column:created_at"              json:"createdAt"`
    LastUsedAt  time.Time `bson:"lastUsedAt"    gorm:"column:last_used_at"            json:"lastUsedAt,omitempty"`
}
```

---

## 4. Repository Interfaces

### InviteRepository

```go
type InviteRepository interface {
    Create(ctx context.Context, invite *entity.Invite) error
    FindByToken(ctx context.Context, token string) (*entity.Invite, error)
    FindAll(ctx context.Context) ([]*entity.Invite, error)
    Delete(ctx context.Context, id string) error
    MarkUsed(ctx context.Context, id string) error
}
```

### AccessTokenRepository

```go
type AccessTokenRepository interface {
    Create(ctx context.Context, token *entity.AccessToken) error
    FindByID(ctx context.Context, documentID string) (*entity.AccessToken, error)
    FindAll(ctx context.Context) ([]*entity.AccessToken, error)
    Delete(ctx context.Context, documentID string) error
    FindByTokenHash(ctx context.Context, hash string) (*entity.AccessToken, error)
}
```

---

## 5. Use Cases

### User UseCase (`usecase/user/`)

| Method | Description |
|---|---|
| `List(ctx)` | List all users |
| `Get(ctx, id)` | Get user by ID |
| `UpdateRole(ctx, userID, roleID)` | Change user's role |
| `Delete(ctx, userID)` | Delete user |

### Invite UseCase (`usecase/invite/`)

| Method | Description |
|---|---|
| `Create(ctx, email, roleID, createdBy)` | Create invite with unique token |
| `List(ctx)` | List all invites |
| `Accept(ctx, token, password)` | Accept invite — create user with the invited role |
| `Revoke(ctx, id)` | Delete/revoke an invite |

### AccessToken UseCase (`usecase/access_token/`)

| Method | Description |
|---|---|
| `Create(ctx, name, createdBy)` | Generate token, store hash, return plaintext once |
| `List(ctx)` | List all tokens (no plaintext — prefix only) |
| `Delete(ctx, documentID)` | Revoke/delete token |

---

## 6. API Contracts

### REST — User Management Routes

| Method | Route | Permission | Response |
|---|---|---|---|
| `GET` | `/api/users` | `users:manage` | `User[]` |
| `GET` | `/api/users/:id` | `users:manage` | `User` |
| `PUT` | `/api/users/:id/role` | `users:manage` | `User` |
| `DELETE` | `/api/users/:id` | `users:manage` | `204` |

### REST — Invite Routes

| Method | Route | Permission | Response |
|---|---|---|---|
| `POST` | `/api/invites` | `users:manage` | `Invite` (201) |
| `GET` | `/api/invites` | `users:manage` | `Invite[]` |
| `DELETE` | `/api/invites/:id` | `users:manage` | `204` |
| `POST` | `/auth/invite/:token` | Public | `{ accessToken, user }` |

The invite accept route is under `/auth/` (public — no auth required).

### REST — Access Token Routes

| Method | Route | Permission | Response |
|---|---|---|---|
| `POST` | `/api/access-tokens` | `access_tokens:manage` | `{ token, ...metadata }` (201) |
| `GET` | `/api/access-tokens` | `access_tokens:manage` | `AccessToken[]` |
| `DELETE` | `/api/access-tokens/:id` | `access_tokens:manage` | `204` |

**Create response:** Returns the plaintext token **once**. Subsequent list calls show only the prefix (first 8 chars).

---

## 7. Testing

**User usecase (`user_usecase_test.go`):**
- List returns all users
- UpdateRole changes roleID
- Delete removes user

**Invite usecase (`invite_usecase_test.go`):**
- Create generates unique token with expiry
- Accept: valid token + password → creates user with correct role; expired → error; already used → error
- Revoke deletes invite

**Access token usecase (`access_token_usecase_test.go`):**
- Create generates token, stores hash, returns plaintext
- List returns tokens without plaintext
- Delete revokes token

---

## 8. Boundaries

| Rule | Detail |
|---|---|
| **Always** | Return plaintext access token only on creation — never in list/get responses |
| **Always** | Mark invite as `used` after successful acceptance — prevent reuse |
| **Always** | Cascade deletion of user removes their associated data per module conventions |
| **Never** | Allow a user to delete themselves |
| **Never** | Allow a user to assign a role with level >= their own |
| **Ask first** | Adding bulk user management operations |

---

## 9. Changelog

| Date | Change | Source |
|------|--------|--------|
| v1.0 | User list + role update + delete | §1 |
| v1.1 | Invite system (create, accept, revoke) | §1 |
| v1.2 | Access token management | §1 |
| v1.3 | Permission-based access (`users:manage`, `access_tokens:manage`) | §12 |
