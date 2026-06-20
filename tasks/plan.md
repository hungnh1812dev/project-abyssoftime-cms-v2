# Plan — personal-cms (project-abyssoftime-cms-v2)

> Completed plans archived in `tasks/archive/plan-phases-A-Y.md`. This file tracks only current and upcoming plans.

---

## Current: Phase Z — Bug Fixes (Auth Flow & Component Naming)

See [bugfix-auth-and-naming.md](bugfix-auth-and-naming.md) for full spec (root causes, code samples).

### Dependency Graph

```
B1 (Register redirect)       — independent, FE only
B2 (Session persistence)     — independent, BE only (config + usecase + handler + tests)
B3 (Component table naming)  — independent, BE only (GORM infra)
    ↓
[Checkpoint Z: all three verified]
```

All three are independent — can be done in any order.

### B1: Register → Login Redirect (FE)

- **File:** `apps/web/src/pages/auth/RegisterPage.tsx`
- **Change:** In register mutation `onSuccess`, call `queryClient.invalidateQueries({ queryKey: ['auth-setup'] })` before `navigate('/login')`
- **Verify:** Register first user → lands on `/login`, not redirected back

### B2: Session Persistence (BE)

**Sub-changes:**

1. **Cookie defaults** (`apps/api/internal/config/config.go`):
   - `COOKIE_SECURE`: default `true` → `false`
   - `COOKIE_SAMESITE`: default `none` → `lax`

2. **RefreshToken signature** (`apps/api/internal/usecase/auth/auth_usecase.go`):
   - Before: `(accessToken string, err)` → After: `(accessToken, newRefreshToken string, err)`
   - Generate new refresh token via `pkgjwt.GenerateRefreshToken(user.ID)`

3. **Refresh handler** (`apps/api/internal/delivery/http/handler/auth_handler.go`):
   - Update `AuthUseCase` interface for 3-return signature
   - Re-set refresh cookie with new token in `Refresh()`

4. **Tests** (`auth_usecase_test.go`, `auth_handler_test.go`): Update for new signature

- **Verify:** `go test ./internal/usecase/auth/... ./internal/delivery/http/handler/...` + login → F5 → stays on admin

### B3: Component Table Naming (BE)

- **File:** `apps/api/internal/infrastructure/gormdb/document_repository.go`
- **Change:** `component_` prefix → `components_` (plural, consistent with `documents_`)
- **Verify:** `go test ./internal/infrastructure/gormdb/...`

### Checkpoint Z

1. `make test-api` — all backend tests green
2. `make test-web` — all frontend tests green
3. Manual smoke: register → login → F5 stays logged in
4. Update `tasks/todo.md` — mark all complete

---

## Upcoming

*(Add new plans here as they are defined.)*
