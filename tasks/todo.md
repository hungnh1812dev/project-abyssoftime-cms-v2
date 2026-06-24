# Todo — personal-cms (project-abyssoftime-cms-v2)

> Completed phases archived in `tasks/archive/`. This file tracks only current and upcoming work.

---

## Phase Z — Bug Fixes: Auth Flow & Naming (Complete)

- [x] B1 Register → Login redirect
- [x] B2 Session persistence
- [x] B3 Component table naming
- [x] ✅ Checkpoint Z

---

## Phase ZZ — Bug Fixes v1.8 (Response Shape, Auth, Inputs, GraphQL)

See [bugfix-v1.8.md](bugfix-v1.8.md) for full plan. Spec: [specs/BUGFIX-SPEC.md](../specs/BUGFIX-SPEC.md).

- [x] B1 User entity: add UUID generation for `id` + `documentId` in `Register()`
- [x] B2 Register page: add `adminExists` guard → redirect to `/login`
- [x] ✅ Checkpoint 1: B1+B2 verified
- [x] B4+B5a Backend: restructure response shape (merge system+content into `data`, remove `contentTypeId`/`status` from public)
- [x] B4+B5b Frontend: adapt hooks + panels to new response shape
- [x] ✅ Checkpoint 2: response shape verified end-to-end
- [x] B3 JsonInput/RichTextInput: fix data loss on save (deep comparison in JsonInput, fallback in RichTextInput)
- [x] ✅ Checkpoint 3: input persistence verified
- [x] B6a Backend: new repo+usecase methods for published document queries
- [x] B6b GraphQL: default queries to published, add `status` filter with auth
- [x] ✅ Checkpoint 4 (Final): all tests green (`go test ./...` + `vitest run`)

---

## Phase UI — Design System: Strapi-Inspired UI Overhaul

See [ui-design-system.md](ui-design-system.md) for full plan. Spec: [specs/ui-design-system.md](../specs/ui-design-system.md).

- [x] T1 Color tokens: migrate to indigo primary + add success/warning/sidebar-muted tokens
- [x] T2 Button: add `success` variant, `loading` prop, update hover/active states
- [x] T3 Badge: add `draft`/`published`/`modified` semantic variants
- [x] T4 SidebarContext: collapsed state + localStorage + mobile detection
- [x] T5 Sidebar components: Brand, Group, SubGroup, Item, CollapseToggle, rail popover
- [x] T6 Sidebar responsive: mobile overlay with backdrop
- [x] T7 AdminLayout: wire new sidebar, remove old Sidebar.tsx
- [x] ✅ Checkpoint 1: foundation + sidebar complete
- [x] T8 Breadcrumbs hook + TopBar rebuild with hamburger
- [x] T9 StickyActionBar: glassmorphism sticky header for content pages
- [x] T10 Card component + page-level spacing polish
- [x] T11 Dark mode verification pass
- [x] ✅ Checkpoint 2 (Final): full design system complete

---

## Phase S — Sync Table Fields & Schema Alignment

See [sync-table-fields.md](sync-table-fields.md) for full plan. Spec: [specs/sync-table-fields.md](../specs/sync-table-fields.md).

- [x] T1 Static table column rename (`id` → `gorm_id`) + UUID standardization
- [x] T2 MediaAsset cleanup + add `FindByDocumentID`
- [x] T3 Document & Component entity cleanup (`Data` → `Fields`, `*time.Time`, remove obsolete)
- [x] ✅ Checkpoint A: backend compiles + tests pass
- [x] T4 Per-field dynamic columns (repository rewrite)
- [x] T5 GraphQL: media as object + remove response wrappers
- [x] T6 Frontend: MediaInput stores `documentId` + aspect ratio fix
- [x] ✅ Checkpoint B (Final): full system verification

---

## Phase DL — Fix Production Data Loss in EnsureCollection

Spec: [fix-data-loss-ensure-collection.md](../specs/fix-data-loss-ensure-collection.md). Plan: [plan.md](plan.md) (Phase DL section).

- [x] T1 Shared `existingColumns` helper (cross-DB column introspection)
- [x] T2 Document `EnsureCollection` → non-destructive (create-if-missing + add-columns)
- [x] T3 Document data-preservation tests (3 new tests)
- [x] ✅ Checkpoint 1a: `go test ./internal/infrastructure/gormdb/ -run TestDocument -v` — 16 pass
- [x] T4 Component `EnsureCollection` → non-destructive
- [x] T5 Component data-preservation tests (3 new tests)
- [x] ✅ Checkpoint 1: `go test ./internal/infrastructure/gormdb/` — all 26 pass + `go test ./...` green
- [x] T6 Sync startup logging (`TableInfo` + log in `syncOne`)
- [x] ✅ Checkpoint 2 (Final): full test suite + startup logs verified

---

## Phase RC — Repeatable Components

Spec: [repeatable-components.md](../specs/repeatable-components.md). Plan: [plan.md](plan.md) (Phase RC section).

- [x] T1 Entity: `Repeatable bool` on `FieldDefinition`, `SortOrder int` on `Component`
- [x] T2 TS type: `repeatable?: boolean` on `FieldDefinition`
- [x] T3 Component repo: `sort_order` column (create + migrate + serialize + order)
- [x] T4 Sync: `Repeatable` in `fieldsEqual` comparison
- [x] ✅ Checkpoint 1: `go test ./...` green, no behavior change
- [x] T5 Usecase: `extractAndSaveComponents` validates shape + assigns `SortOrder`
- [x] T6 Usecase: `mergeComponents` uses `Repeatable` flag (array vs object)
- [x] T7 Usecase tests: 6 new tests (validation + merge shape)
- [x] ✅ Checkpoint 2: `go test ./...` green, shape enforcement active
- [x] T8 GraphQL SDL: `[Type!]` for repeatable component fields
- [x] T9 GraphQL resolver: `NewList(NewNonNull(compType))` for repeatable
- [x] T10 GraphQL test: repeatable component SDL generation
- [x] ✅ Checkpoint 3: `go test ./...` green
- [x] T11 `RepeatableComponentField` React component (useFieldArray + add/remove/reorder)
- [x] T12 `renderSchemaField`: branch on `field.repeatable`
- [x] T13 Frontend tests: 5 tests for repeatable component UI
- [x] T14 Barrel export from `@/components/form`
- [x] ✅ Checkpoint 4 (Final): `vitest run` 186 pass + `tsc --noEmit` clean

---

## Phase AX — Fix Cross-Origin Auth (Production F5 Redirect)

Spec: [bugfix-auth-and-naming.md](bugfix-auth-and-naming.md) (B2 follow-up). Plan: [plan.md](plan.md) (Phase AX section).

- [x] T1 BE: Refresh handler accepts token from request body (fallback to cookie)
- [x] T2 BE: Login + Refresh return `refreshToken` in response body
- [x] T3 BE: Update auth handler tests (body-based refresh, refreshToken in response)
- [x] ✅ Checkpoint 1: `go test ./... -count=1` green
- [x] T4 FE: Refresh token storage helpers (localStorage/sessionStorage)
- [x] T5 FE: AuthContext — login/mount/logout use stored refresh token
- [x] T6 FE: LoginPage passes rememberMe + refreshToken to login()
- [x] T7 FE: api.ts interceptor sends stored refresh token in body
- [x] T8 FE: Update frontend tests
- [x] ✅ Checkpoint 2: `vitest run` (187 pass) + `tsc --noEmit` + lint green
- [ ] ✅ Checkpoint 3: Manual E2E cross-origin test

---

## Phase HP — Background Health Ping Service

Spec: [specs/health-ping.md](../specs/health-ping.md). Plan: [plan.md](plan.md) (Phase HP section).

- [ ] T1.1 Create `ConnectionOverlay.tsx` — full-screen overlay with spinner, text, ARIA, fade transition
- [ ] T1.2 Create `ConnectionOverlay.test.tsx` — visible/hidden states, accessibility
- [ ] Checkpoint 1: `tsc --noEmit` + ConnectionOverlay tests pass
- [ ] T2.1 Create `HealthContext.tsx` — ping loop with fetch, 10s/14m intervals, visibility API, HealthProvider + useHealthStatus
- [ ] T2.2 Create `HealthContext.test.tsx` — 8 test cases (ping lifecycle, timers, visibility, cleanup)
- [ ] Checkpoint 2: `tsc --noEmit` + HealthContext tests pass
- [ ] T3.1 Update `main.tsx` — wrap app with HealthProvider
- [ ] Checkpoint 3 (Final): `make test-web` + `tsc --noEmit` + manual browser verification

---

## Archive Index

| Archive | Phases | Status |
|---------|--------|--------|
| [phases-0-5-foundation.md](archive/phases-0-5-foundation.md) | 0–5 | ✅ Complete |
| [phases-A-D-core-migrations.md](archive/phases-A-D-core-migrations.md) | A–D | ✅ Complete |
| [phase-M-media-forms.md](archive/phase-M-media-forms.md) | M | ✅ Complete |
| [phases-W-X-Y-web-api.md](archive/phases-W-X-Y-web-api.md) | W, X, Y | ✅ Complete |
