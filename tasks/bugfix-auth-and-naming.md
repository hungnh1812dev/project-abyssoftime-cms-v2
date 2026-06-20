# Bug Fixes — Auth Flow & Content-Type Document Naming

> Backup of SPEC §14 changes. Apply to SPEC.md after project structure refactoring is complete, then use for planning and build.

---

## B1: Register → Login Redirect

**Symptom:** Register → success → stays on register page. Expected: redirect to login page.

**Root cause:** After first-user registration, `RegisterPage` navigates to `/login`. `LoginPage` queries `['auth-setup']` with `staleTime: 30_000`. Since `RegisterPage` already fetched it (returning `adminExists: false`) < 30s ago, TanStack Query returns stale cache. `LoginPage` checks `if (!setupData.adminExists)` → redirects back to `/register`.

**Fix — `apps/web/src/pages/auth/RegisterPage.tsx`:**

Invalidate `['auth-setup']` query before navigating:

```tsx
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

const queryClient = useQueryClient()

const mutation = useMutation({
  mutationFn: (data: RegisterFields) =>
    api.post('/auth/register', data).then((r) => r.data),
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['auth-setup'] })
    navigate('/login')
  },
  onError: () => {
    setErrorMsg('Registration failed. The email may already be in use.')
  },
})
```

---

## B2: Session Persistence (Stay Logged In)

**Symptom:** Login (check "Stay logged in") → F5 → login page shown. Expected: admin page.

**Root cause (primary):** Default cookie config `COOKIE_SECURE=true` + `COOKIE_SAMESITE=none` prevents the browser from saving the refresh token cookie on HTTP (local dev). `SameSite=None; Secure` requires HTTPS. On F5, `/auth/refresh` finds no cookie → 401 → redirect to login.

**Root cause (secondary):** Refresh handler never re-issues the cookie. Session expires from original login time regardless of activity.

### Fix 1 — Change cookie defaults (`apps/api/internal/config/config.go`)

| Setting | Before (default) | After (default) |
|---|---|---|
| `COOKIE_SECURE` | `true` | `false` |
| `COOKIE_SAMESITE` | `none` | `lax` |

```go
CookieSecure:   parseBoolDefault(os.Getenv("COOKIE_SECURE"), false),
CookieSameSite: parseSameSite(getenv("COOKIE_SAMESITE", "lax")),
```

Production (Render) must set `COOKIE_SECURE=true` and `COOKIE_SAMESITE=none`.

### Fix 2 — Re-issue refresh token on refresh

**`apps/api/internal/usecase/auth/auth_usecase.go`** — change `RefreshToken` signature:

```go
// Before:
func (uc *UseCase) RefreshToken(ctx context.Context, refreshToken string) (string, error)

// After:
func (uc *UseCase) RefreshToken(ctx context.Context, refreshToken string) (accessToken, newRefreshToken string, err error)
```

Generate new refresh token via `pkgjwt.GenerateRefreshToken(user.ID)` and return alongside access token.

**`apps/api/internal/delivery/http/handler/auth_handler.go`** — update interface + handler:

```go
// Interface change:
RefreshToken(ctx context.Context, refreshToken string) (string, string, error)

// Handler:
func (h *AuthHandler) Refresh(c *gin.Context) {
    cookieVal, err := c.Cookie(RefreshCookieName)
    if err != nil {
        ginWriteError(c, http.StatusUnauthorized, "missing refresh token")
        return
    }

    access, refresh, err := h.uc.RefreshToken(c.Request.Context(), cookieVal)
    if err != nil {
        ginWriteErr(c, err)
        return
    }

    c.SetSameSite(h.cookieSameSite)
    c.SetCookie(RefreshCookieName, refresh, refreshCookieMaxAgeRemember, "/", "", h.cookieSecure, true)

    c.JSON(http.StatusOK, gin.H{
        "accessToken": access,
    })
}
```

---

## B3: Content-Type Document Table Naming (PostgreSQL)

**Symptom:** SPEC §12.8 uses singular `component_` prefix, inconsistent with plural `documents_`.

**Fix — update naming convention:**

| Data type | Table naming |
|---|---|
| Single-type, Collection-type | `documents_<slug>` (unchanged) |
| Component | `components_{slug}_{component-name}` (plural prefix) |

- `{component-name}` = the `name` field from the component's `FieldDefinition` in the JSON schema
- Slug hyphens → underscores, camelCase → snake_case
- **MongoDB: no change** — components stay nested in BSON `data`

**Examples:**

| Content-type slug | Component field name | Table name |
|---|---|---|
| `blog-posts` | `seo` | `components_blog_posts_seo` |
| `about-page` | `heroSection` | `components_about_page_hero_section` |

**GORM helper (`apps/api/internal/infrastructure/gormdb/document_repository.go`):**

```go
func componentTableName(contentTypeSlug, componentFieldName string) string {
    slug := strings.ReplaceAll(contentTypeSlug, "-", "_")
    field := toSnakeCase(componentFieldName)
    return "components_" + slug + "_" + field
}
```

---

## Files to Change

**Frontend:**
```
apps/web/src/pages/auth/RegisterPage.tsx                    ← B1: invalidate ['auth-setup'] query
```

**Backend:**
```
apps/api/internal/config/config.go                          ← B2: cookie defaults
apps/api/internal/usecase/auth/auth_usecase.go              ← B2: RefreshToken new signature
apps/api/internal/usecase/auth/auth_usecase_test.go         ← B2: update tests
apps/api/internal/delivery/http/handler/auth_handler.go     ← B2: re-set cookie in Refresh
apps/api/internal/delivery/http/handler/auth_handler_test.go ← B2: update tests
apps/api/internal/infrastructure/gormdb/document_repository.go ← B3: components_ prefix
```

**Spec:**
```
SPEC.md §12.8  ← component_ → components_
SPEC.md §13.6  ← note cookie default changes
SPEC.md §14    ← add this bug-fix section
```

---

## Acceptance Criteria

- [ ] B1: After first-user registration, user lands on `/login` (not redirected back to `/register`)
- [ ] B2: Login + F5 → admin page loads (no redirect to login)
- [ ] B2: Login with "Stay logged in" + F5 → admin page loads
- [ ] B2: Default cookie config works for local HTTP dev
- [ ] B2: Refresh handler re-issues refresh token cookie
- [ ] B3: GORM component tables use `components_` plural prefix
- [ ] B3: MongoDB unchanged — components nested in BSON
