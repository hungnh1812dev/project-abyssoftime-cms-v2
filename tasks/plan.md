# Plan — personal-cms (project-abyssoftime-cms-v2)

> Completed plans archived in `tasks/archive/plan-phases-A-Y.md`. This file tracks only current and upcoming plans.

---

## Current: Phase Z — Bug Fixes

See [bugfix-auth-and-naming.md](bugfix-auth-and-naming.md) for implementation details.

### Dependency Graph

```
B1 (Register redirect)     — independent
B2 (Session persistence)   — independent
B3 (Component table naming) — independent
    ↓
[Checkpoint Z: all fixes verified]
```

All three bugs are independent — can be fixed in any order or in parallel.

### B1: Register → Login Redirect
- **File:** `apps/web/src/pages/auth/RegisterPage.tsx`
- **Fix:** Invalidate `['auth-setup']` query in `onSuccess` before `navigate('/login')`

### B2: Session Persistence
- **Files:** `config.go`, `auth_usecase.go`, `auth_handler.go` + tests
- **Fix 1:** Change cookie defaults: `COOKIE_SECURE=false`, `COOKIE_SAMESITE=lax`
- **Fix 2:** `RefreshToken` returns new refresh token; handler re-sets cookie

### B3: Component Table Naming
- **File:** `apps/api/internal/infrastructure/gormdb/document_repository.go`
- **Fix:** `component_` prefix → `components_` (plural)

---

## Upcoming

*(Add new plans here as they are defined.)*
