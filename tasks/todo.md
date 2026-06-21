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
- [ ] T2 Button: add `success` variant, `loading` prop, update hover/active states
- [ ] T3 Badge: add `draft`/`published`/`modified` semantic variants
- [ ] T4 SidebarContext: collapsed state + localStorage + mobile detection
- [ ] T5 Sidebar components: Brand, Group, SubGroup, Item, CollapseToggle, rail popover
- [ ] T6 Sidebar responsive: mobile overlay with backdrop
- [ ] T7 AdminLayout: wire new sidebar, remove old Sidebar.tsx
- [ ] ✅ Checkpoint 1: foundation + sidebar complete
- [ ] T8 Breadcrumbs hook + TopBar rebuild with hamburger
- [ ] T9 StickyActionBar: glassmorphism sticky header for content pages
- [ ] T10 Card component + page-level spacing polish
- [ ] T11 Dark mode verification pass
- [ ] ✅ Checkpoint 2 (Final): full design system complete

---

## Archive Index

| Archive | Phases | Status |
|---------|--------|--------|
| [phases-0-5-foundation.md](archive/phases-0-5-foundation.md) | 0–5 | ✅ Complete |
| [phases-A-D-core-migrations.md](archive/phases-A-D-core-migrations.md) | A–D | ✅ Complete |
| [phase-M-media-forms.md](archive/phase-M-media-forms.md) | M | ✅ Complete |
| [phases-W-X-Y-web-api.md](archive/phases-W-X-Y-web-api.md) | W, X, Y | ✅ Complete |
