# RULES — admin Module

**Scope:** User management (list/update/delete), invite system (create/accept/revoke), access token management (create/list/revoke).
**Spec:** [specs/admin.md](../specs/admin.md)

---

## 1. Entity Rules

### 1.1 Invite Entity
```go
type Invite struct {
    ID        string    // internal PK
    Token     string    // unique invite token (32-byte random)
    Email     string    // invitee email
    RoleID    string    // pre-assigned role (references Role.DocumentID)
    CreatedBy string    // user who created the invite
    CreatedAt time.Time
    ExpiresAt time.Time // 7 days from creation
    Used      bool      // true after acceptance
}
```
- Token: unique, random 32-byte string
- Expires after 7 days
- Single-use: marked `used = true` after acceptance

### 1.2 AccessToken Entity
```go
type AccessToken struct {
    ID         string    // internal PK
    DocumentID string    // domain identifier (UUID)
    Name       string    // human-readable label
    TokenHash  string    // SHA-256 hash — NEVER store plaintext
    Prefix     string    // first 8 chars for display
    CreatedBy  string    // user who created
    CreatedAt  time.Time
    LastUsedAt time.Time // updated on every validation
}
```
- Plaintext token returned **only** on creation — **NEVER** in list/get
- `TokenHash`: SHA-256, stored permanently
- `Prefix`: first 8 characters, for display identification
- Tokens cannot be edited — delete and recreate to change

---

## 2. User Management Rules

### 2.1 List/Get
- Returns all users with their profile info
- `DisplayName` included in responses
- Accessible to `users:manage` permission holders

### 2.2 Role Update
- A user can only change role of someone with **strictly lower** level
- A user can only assign a role that is **strictly lower** than their own
- **NEVER** allow a user to change their own role
- **NEVER** allow assigning a role with `level >= caller's level`

### 2.3 User Deletion
- **NEVER** allow a user to delete themselves
- Deleting a user does **NOT** cascade-delete their content contributions
- `createdBy`/`updatedBy` fields on documents retain the deleted user's ID

---

## 3. Invite System Rules

### 3.1 Create Invite
- Generates unique token with 7-day expiry
- Email and RoleID required
- Role must be strictly lower than creator's role
- CreatedBy set to current user

### 3.2 Accept Invite (`POST /auth/invite/:token`)
- Public route — no auth required
- Validates: token exists, not expired, not used
- Requires: `password` + `displayName` in request body
- Creates user with the pre-assigned role
- Marks invite as `used = true` — prevents reuse
- Returns `{ accessToken, user }` (same as login)

### 3.3 Revoke Invite
- Deletes the invite record
- Permission: `users:manage`

### 3.4 Invariants
- **Always** mark invite as `used` after successful acceptance
- **NEVER** allow reuse of an accepted invite
- **NEVER** allow acceptance of expired invite
- `displayName` is **required** on invite accept

---

## 4. Access Token Rules

### 4.1 Create Token
- Generates random token, stores SHA-256 hash
- Returns plaintext **once** in creation response
- Subsequent list/get shows only prefix (first 8 chars)
- Permission: `access_tokens:manage`

### 4.2 List Tokens
- Returns all tokens with metadata (name, prefix, createdBy, createdAt, lastUsedAt)
- **NEVER** include plaintext or hash in list responses

### 4.3 Delete/Revoke Token
- Removes token record
- Token immediately becomes invalid
- Permission: `access_tokens:manage`

### 4.4 Token Validation (at auth layer)
- Hash incoming token with SHA-256
- Look up by hash
- Check expiration
- Update `lastUsedAt` on success
- Reject expired tokens

---

## 5. API Contract Rules

### 5.1 User Routes
| Method | Route | Permission |
|---|---|---|
| `GET` | `/api/users` | `users:manage` |
| `GET` | `/api/users/:id` | `users:manage` |
| `PUT` | `/api/users/:id/role` | `users:manage` |
| `DELETE` | `/api/users/:id` | `users:manage` |

### 5.2 Invite Routes
| Method | Route | Permission |
|---|---|---|
| `POST` | `/api/invites` | `users:manage` |
| `GET` | `/api/invites` | `users:manage` |
| `DELETE` | `/api/invites/:id` | `users:manage` |
| `POST` | `/auth/invite/:token` | Public |

### 5.3 Access Token Routes
| Method | Route | Permission |
|---|---|---|
| `POST` | `/api/access-tokens` | `access_tokens:manage` |
| `GET` | `/api/access-tokens` | `access_tokens:manage` |
| `DELETE` | `/api/access-tokens/:id` | `access_tokens:manage` |

---

## 6. Testing Rules (Admin-Specific)

### 6.1 User Usecase Tests
- List returns all users
- UpdateRole changes roleID
- Delete removes user
- Self-delete → rejected
- Role upgrade beyond own level → rejected

### 6.2 Invite Usecase Tests
- Create generates unique token with expiry
- Accept: valid token → creates user with correct role
- Accept: expired token → error
- Accept: used token → error
- Accept: missing displayName → error
- Revoke deletes invite

### 6.3 Access Token Usecase Tests
- Create generates token, stores hash, returns plaintext
- List returns tokens without plaintext
- Delete revokes token
- Validation: correct hash → success, wrong hash → error, expired → error

---

## 7. Boundaries Summary

| Rule | Detail |
|---|---|
| **Always** | Return plaintext access token **only** on creation |
| **Always** | Mark invite as `used` after successful acceptance |
| **Always** | Require `displayName` on invite accept |
| **Always** | Cascade deletion of user removes account, NOT content contributions |
| **Always** | Store access tokens as SHA-256 hashes |
| **Always** | Update `lastUsedAt` on every successful token validation |
| **Never** | Allow a user to delete themselves |
| **Never** | Allow assigning role with `level >= caller's level` |
| **Never** | Include plaintext token in list/get responses |
| **Never** | Allow reuse of accepted invite |
| **Never** | Allow editing access tokens (delete + recreate) |
| **Ask first** | Adding bulk user management operations |
