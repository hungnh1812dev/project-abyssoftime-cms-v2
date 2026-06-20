# Todo — personal-cms (project-abyssoftime-cms-v2)

> Completed phases archived in `tasks/archive/`. This file tracks only current and upcoming work.

---

## Phase Z — Bug Fixes: Auth Flow & Naming

See [bugfix-auth-and-naming.md](bugfix-auth-and-naming.md) for full details.

- [x] B1 Register → Login redirect (invalidate `['auth-setup']` query before navigate)
- [x] B2 Session persistence: fix cookie defaults (`COOKIE_SECURE=false`, `COOKIE_SAMESITE=lax` for dev) + re-issue refresh token on `/auth/refresh`
- [x] B3 Component table naming: spec-only fix — no code exists yet; `components_` prefix documented for future implementation
- [x] ✅ Checkpoint Z: all tests pass (`make test-api` + `make test-web`)

---

## Archive Index

| Archive | Phases | Status |
|---------|--------|--------|
| [phases-0-5-foundation.md](archive/phases-0-5-foundation.md) | 0–5 | ✅ Complete |
| [phases-A-D-core-migrations.md](archive/phases-A-D-core-migrations.md) | A–D | ✅ Complete |
| [phase-M-media-forms.md](archive/phase-M-media-forms.md) | M | ✅ Complete |
| [phases-W-X-Y-web-api.md](archive/phases-W-X-Y-web-api.md) | W, X, Y | ✅ Complete |
