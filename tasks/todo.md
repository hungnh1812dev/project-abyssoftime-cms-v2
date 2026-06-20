# Todo — personal-cms (project-abyssoftime-cms-v2)

> Completed phases archived in `tasks/archive/`. This file tracks only current and upcoming work.

---

## Phase Z — Bug Fixes: Auth Flow & Naming

See [bugfix-auth-and-naming.md](bugfix-auth-and-naming.md) for full details.

- [ ] B1 Register → Login redirect (invalidate `['auth-setup']` query before navigate)
- [ ] B2 Session persistence: fix cookie defaults (`COOKIE_SECURE=false`, `COOKIE_SAMESITE=lax` for dev) + re-issue refresh token on `/auth/refresh`
- [ ] B3 Component table naming: `component_` → `components_` plural prefix in GORM adapter
- [ ] ✅ Checkpoint Z: register→login works, F5 stays logged in, component tables named correctly

---

## Archive Index

| Archive | Phases | Status |
|---------|--------|--------|
| [phases-0-5-foundation.md](archive/phases-0-5-foundation.md) | 0–5 | ✅ Complete |
| [phases-A-D-core-migrations.md](archive/phases-A-D-core-migrations.md) | A–D | ✅ Complete |
| [phase-M-media-forms.md](archive/phase-M-media-forms.md) | M | ✅ Complete |
| [phases-W-X-Y-web-api.md](archive/phases-W-X-Y-web-api.md) | W, X, Y | ✅ Complete |
