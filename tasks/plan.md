# Plan — Migrate personal-cms to Draft/Publish + Schema-as-Code + Multi-Storage

## Context

`SPEC.md` was extended with three features that this codebase does not implement yet:

1. Draft/publish as two records per entry (`entryId` + computed status), with audit
   fields (`createdBy`/`updatedBy`/`publishedBy`/`locale`).
2. Content-type **structure** defined as JSON files synced into Mongo on boot — never
   created via API/UI.
3. Dual media storage (S3 + Cloudinary) behind the existing `StorageAdapter` interface.

Phases 0–5 of the original plan (see "Archived: Original Plan" below) are already
built end-to-end against the **old** model:

- `entity.Document{ID, ContentTypeID, Status, Data}` — one record, `Status: draft|published`, toggled via `Publish`/`Unpublish`.
- `entity.ContentType{ID, Name, Slug, Kind}` — created/updated/deleted via `POST/PUT/DELETE /api/content-types` (`content_type_handler.go`) and matching FE hooks (`useCreateContentType` etc.), though no UI currently calls those mutation hooks.
- Single storage adapter wired (Cloudinary only); `StorageAdapter` interface already exists, so adding S3 is additive.
- `Sidebar.tsx` lists content types flat, not grouped by `kind`.
- `SingleTypePanel.tsx` requires a document to already exist; nothing auto-creates the singleton.

This plan is a **migration**: close the gap between the already-built code and the
current `SPEC.md`, reusing existing patterns (mock-repository unit tests, TanStack
Query hook conventions, `FormProvider` usage) rather than rewriting them.

## Decisions (confirmed)

- **Unpublish stays** as a CMS convenience beyond what `SPEC.md` defines (clears the
  published record, reverting status to `draft`). Documented divergence — follow up
  with a `/spec` pass once built so spec and code don't drift apart again.
- **Routes**: keep `/api/documents` prefix; `:id` now means `entryId`.
- **S3 adapter**: build it now, alongside Cloudinary, config-selectable.

## Dependency Graph

```
Phase A (Schema-as-Code)         — independent, foundational
  A1 JSON schema loader (pure parsing)
  A2 Sync usecase (reconcile defs → ContentType + cascade) ── depends on A1
  A3 Wire Sync into cmd/server/main.go boot ── depends on A2
  A4 Remove ContentType Create/Update/Delete (API + unused FE hooks) ── depends on A2/A3

Phase B (Draft/Publish remodel)  — independent of A, A4 routes through it
  B1 Document entity/repo: entryId + version (draft/published) records
  B2 Document usecase: Save/Publish/computed status/audit fields ── depends on B1
  B3 Document handlers: GET/PUT/.../publish addressed by entryId; public read = published-only ── depends on B2
  B4 FE: useDocuments hooks updated to entryId + status field ── depends on B3
  B5 FE: panels — status badge + Save/Publish/Unpublish wiring ── depends on B4

Phase C (Content-type kind UX) — depends on existing ContentType.Kind (already present)
  C1 Single-type auto-singleton creation, triggered from Sync (A2) ── depends on A2, B2
  C2 Sidebar grouping by kind ── verified after B5

Phase D (Storage)                — independent, parallel with A/B/C
  D1 S3 adapter (implements StorageAdapter)
  D2 Config-driven adapter selection

Checkpoints: after A, after B, after C+D (final)
```

## Phase A — Schema-as-Code

### A1 — JSON Schema Loader
New file: `apps/api/internal/usecase/content_type/schema_loader.go`
- Reads `apps/api/content-types/*.json`, unmarshals into `ContentTypeDefinition{slug, name, kind, fields: []FieldDef{name, type}}`.
- Pure function, no DB. Unit-testable with fixture files in `testdata/`.

**Acceptance:** loads fixture JSON files → parsed definitions; malformed JSON returns a clear error, never panics.

### A2 — Sync Usecase
`Sync(ctx, definitions []ContentTypeDefinition) error` in `usecase/content_type`:
- New definition → `Create`. Changed definition → `Update`. Definition missing → `Delete` + cascade (new `DeleteByContentType` on `DocumentRepository`, since current `Document.Delete` only deletes by single id).
- Field removal: schema-only change, no document data mutation.

**Acceptance:** table-driven tests with mock repos — new/changed/missing/unchanged definitions each trigger the right repo calls.

### A3 — Wire Into Startup
`cmd/server/main.go`: load JSON defs, call `Sync` before the HTTP server starts listening. Log created/updated/removed counts.

**Acceptance:** new fixture file → content type appears in Mongo on restart; removed file → content type and its documents are gone on restart.

### A4 — Remove ContentType Mutation API/UI
- Drop `Create`/`Update`/`Delete` from `content_type_handler.go` + router (keep `GET` list/detail only).
- Drop `useCreateContentType`/`useUpdateContentType`/`useDeleteContentType` from `useContentTypes.ts`.
- Remove now-stale handler tests for the removed endpoints.

**Acceptance:** `go build ./...` and `npm run build` succeed; remaining handler tests pass.

---
## ✅ Checkpoint A
- Server boot syncs `content-types/*.json` → Mongo (create/update/delete all verified).
- No HTTP route or FE call exists to create/edit/delete a `ContentType`'s structure.

---

## Phase B — Draft/Publish Remodel

### B1 — Entity + Repository
Document entity gains `EntryID`, `Version` (`draft`|`published`), `CreatedBy`, `UpdatedBy`, `PublishedBy`, `Locale` (default `"en"`).

`DocumentRepository` becomes entry-aware:
- `FindDraftByEntryID`, `FindPublishedByEntryID`
- `UpsertDraft`, `UpsertPublished`
- `FindEntriesByContentType` — one row per `entryID` for list views/status computation
- `DeleteByEntryID`, `DeleteByContentType` (for A2 cascade)

Update Mongo implementation + mocks.

**Acceptance:** `go build ./internal/domain/...` succeeds; mocks match new interface.

### B2 — Document Usecase
- `Save(ctx, entryID, data, userID)` → upserts draft only.
- `Publish(ctx, entryID, userID)` → copies draft → published, syncs `UpdatedAt`/`PublishedAt`/`PublishedBy`.
- `Unpublish(ctx, entryID)` → deletes published record (confirmed extra beyond `SPEC.md`).
- `Status(draft, published)` → pure helper: `draft` / `modified` / `published`.
- `GetForEdit(ctx, entryID)` → draft + computed status (admin).
- `GetPublished(ctx, entryID)` → published only, `ErrNotFound` if absent (public read).
- `Delete(ctx, entryID)` → cascades to media + both draft/published records.

**Acceptance:** tests cover all three statuses, Save never touching published, Publish syncing timestamps, `GetPublished` 404 when unpublished.

### B3 — Handlers
Keep `/api/documents` prefix, `:id` = `entryID`:
- `GET /api/documents/:id` (admin) → `GetForEdit`
- `PUT /api/documents/:id` → `Save`
- `POST /api/documents/:id/publish` / `/unpublish`
- `DELETE /api/documents/:id`
- `GET /api/documents?contentType=:slug` → entry summaries with computed status
- New public/content read path resolving `GetPublished` only.

**Open item:** exact gating for the public read path (separate route prefix vs. existing role check) isn't pinned by `SPEC.md` — confirm before building this task.

**Acceptance:** admin GET returns draft+status; unpublished entry on public path → 404; after publish, public path returns new data.

### B4 — FE Hooks
`useDocuments.ts` updated to entry-shaped responses (`status: 'draft'|'modified'|'published'`). Mutation payload shape (`{ id, contentTypeId }`) unchanged.

**Acceptance:** `useDocuments.test.ts` passes against new response shape.

### B5 — Panels
`SingleTypePanel.tsx`, `CollectionDetailPanel.tsx`, `CollectionListPage.tsx`: tri-state status badge; Publish shows whenever `status !== 'published'`; Unpublish shows when a published record exists.

**Acceptance:** component tests updated; manual check via `make dev` — edit → save → "modified" badge → publish → "published" badge.

---
## ✅ Checkpoint B
- Save never affects the public/published read; Publish makes it catch up.
- Tri-state status badge correct in both panel types.
- `go test ./...` and `vitest run` pass.

---

## Phase C — Content-Type Kind UX

### C1 — Single-Type Auto-Singleton
`content_type.Sync`: on creating a `single`-kind `ContentType`, immediately create its singleton draft entry (`entryID = contentTypeID`) so `SingleTypePanel` never shows "No document found" for a synced single type.

**Acceptance:** sync test — new single-kind definition results in both `ContentType` and singleton document existing.

### C2 — Sidebar Grouping
`Sidebar.tsx`: split content types by `Kind` into two labeled sections — "Single Types" / "Collection Types".

**Acceptance:** test asserts both section headers render with content types in the correct group.

---
## ✅ Checkpoint C
- New single-type definition → singleton auto-created, editable immediately, no manual creation step.
- Sidebar visually grouped by kind.

---

## Phase D — Storage: S3 Adapter

### D1 — S3 Adapter
`apps/api/internal/infrastructure/storage/s3_adapter.go` implements `StorageAdapter` using AWS SDK v2; credentials from env vars only.

**Acceptance:** unit test against a mocked S3 client — `Upload`/`Delete` behave correctly.

### D2 — Config-Driven Selection
`cmd/server/main.go`: env var (`STORAGE_PROVIDER=s3|cloudinary`) selects which adapter is injected into `media.New(...)`.

**Acceptance:** flipping the env var changes which adapter `media.Upload` calls; covered by `media_usecase_test.go`-style test.

---
## ✅ Checkpoint D (Final)
- Both storage adapters exist, pass unit tests, selectable via env var.
- Full suite green: `go test ./...`, `go vet ./...`, `npm run lint`, `npm run build`, `vitest run`.
- Manual smoke test: `make dev` → log in → edit single-type entry → Save (still draft, public read unchanged) → Publish (public read updates) → Sidebar shows two grouped sections.

---

## After This Plan Lands
Run a short `/spec` follow-up to record the kept "Unpublish" behavior in `SPEC.md`,
since this plan intentionally keeps it as an extra beyond the written spec.

---

## Archived: Original Plan (Phases 0–5, already built)

See git history / existing code for full detail — summarized here for context only:

- **Phase 0 (Foundation):** monorepo scaffold, domain entities, Mongo client, FE base. ✅ Done.
- **Phase 1 (Auth):** register/login/refresh/logout, JWT, route guards. ✅ Done.
- **Phase 2 (Form System):** FormProvider/FormField, Text/Number/Boolean/RichText/Json inputs. ✅ Done.
- **Phase 3 (Core CMS):** ContentType + Document CRUD (old single-record model), admin layout, single-type/collection-type panels. ✅ Done — being migrated by Phases A–C above.
- **Phase 4 (Media Upload):** MediaAsset + StorageAdapter + Cloudinary adapter, MediaInput. ✅ Done — extended by Phase D above.
- **Phase 5 (CI/CD + Docker):** Dockerfiles, docker-compose, GitHub Actions. ✅ Done (assumed stable; not affected by this migration).

---

---

# Plan — Web Content-Type Management System (Refactor)

## Context

`SPEC.md §7` specifies a refactor of the web admin panel. The API is untouched.
All work is in `apps/web/src/`.

**Current state (what already exists):**
- `FormProvider` — works, but missing: `isDirty` exposure, success toast, post-save `reset()`
- `FormStateContext` — has `loading` + `submitting`; missing `isDirty`
- `SingleTypePanel` — has locale state + Publish/Unpublish; needs `ContentTypeLayout` wrapper
- `CollectionListPage` — has a table, but columns are hard-coded (first-field + Status)
- `CollectionDetailPanel` / `CollectionDetailPage` — same locale/layout issues as SingleTypePanel
- `Sidebar` — grouped by kind (correct); links use old `/admin/content-types/...` paths
- `router.tsx` — old routes; `SiteHomepagePanelWrapper` is a one-off inline wrapper
- `SiteHomepagePanel` — custom panel; must keep working after migration

**What is new:**
- `ContentTypeLayout` component (render-prop wrapper)
- Content-type registry (`src/content-type-registry/index.ts`)
- `SingleTypePage` (replaces `SingleTypePanel`; uses `ContentTypeLayout`)
- Updated `CollectionListPage` (registry-driven columns)
- Updated `CollectionDetailPage` / `CollectionDetailPanel` (uses `ContentTypeLayout`)
- Updated `Sidebar` (new route paths)
- Updated `router.tsx` (new routes, remove old, re-wire `SiteHomepagePanel`)

---

## Dependency Graph

```
FormStateContext (+ isDirty)
        ↓
FormProvider (success toast, reset, isDirty forwarded)
        ↓
ContentTypeLayout (reads isDirty from FormStateContext for Save button state)

Registry (standalone, no component imports)
        ↓
CollectionListPage (reads columns from registry)

ContentTypeLayout + FormProvider + Registry
        ↓
SingleTypePage
CollectionDetailPanel (refactored)

Sidebar (standalone: just route string changes)

Router (depends on: SingleTypePage, CollectionListPage, CollectionDetailPage, Sidebar)
```

---

## Slices

Work is sliced **vertically** — each task delivers a complete, testable unit of user-visible behaviour.

---

### W1 — FormProvider Lifecycle Hardening

**What**: Extend `FormProvider` and `FormStateContext` to enforce the full form lifecycle.

**Changes:**
- Add `isDirty: boolean` to `FormStateContext` value (sourced from `react-hook-form`'s `formState.isDirty`)
- On successful save: `toast.success('Saved')` + `queryClient.invalidateQueries(queryKey)` + `reset(newServerData)` so form syncs to fresh server data and `isDirty` resets to `false`
- On failed save: `toast.error(message)` — already fired in mutation hooks; `FormProvider` must not suppress it
- Save button (`type="submit"`) disabled when `submitting || !isDirty`

**Files changed:**
- `src/components/form/FormStateContext.tsx` — add `isDirty` field
- `src/components/form/FormProvider.tsx` — wire isDirty, onSuccess reset, toast
- `src/components/form/__tests__/FormProvider.test.tsx` — add lifecycle test cases

**Acceptance criteria:**
- [ ] `FormStateContext` exports `isDirty`
- [ ] On mount with server data, `isDirty === false`
- [ ] After editing any field, `isDirty === true`
- [ ] Successful mutation: success toast fired, `reset()` called with new data, `isDirty` becomes `false`
- [ ] Failed mutation: error toast fired, form values unchanged
- [ ] TypeScript strict: no `any`, no type errors

**Verification:** `npm run test -- FormProvider` passes; `npm run build` clean.

---

### W2 — ContentTypeLayout Component

**What**: New layout shell for any content-type page.

```ts
interface ContentTypeLayoutProps {
  title: string
  status?: string
  renderHeader?: (defaultHeader: ReactNode) => ReactNode
  renderActions?: () => ReactNode
  children: ReactNode
}
```

Default render: left side = title + status badge; right side = `renderActions()` output.
`renderHeader(defaultHeader)` replaces the entire header row when provided.

**Files changed:**
- `src/components/content-type/ContentTypeLayout.tsx` — new
- `src/components/content-type/__tests__/ContentTypeLayout.test.tsx` — new

**Acceptance criteria:**
- [ ] Default header renders title and status badge
- [ ] `renderActions` output appears to the right of the default header
- [ ] `renderHeader` completely replaces the header row
- [ ] `children` renders below the header
- [ ] When `status` is omitted, no badge rendered
- [ ] TypeScript strict

**Verification:** `npm run test -- ContentTypeLayout` passes.

---

### W3 — Content-Type Registry

**What**: Metadata-only module; no component imports at this level.

```ts
// src/content-type-registry/index.ts
export interface CollectionColumnDef {
  key: string
  label: string
  type: 'text' | 'boolean' | 'number' | 'image'
}

export interface ContentTypeRegistration {
  slug: string
  kind: 'single' | 'collection'
  columns?: CollectionColumnDef[]
  wrapper?: React.ComponentType<ContentTypeLayoutProps>
}

export const contentTypeRegistry: ContentTypeRegistration[] = []
```

**Files changed:**
- `src/content-type-registry/index.ts` — new

**Acceptance criteria:**
- [ ] Module exports `ContentTypeRegistration`, `CollectionColumnDef`, `contentTypeRegistry`
- [ ] No component imports in this file
- [ ] TypeScript strict; no `any`

**Verification:** `npm run build` compiles without error; no circular imports.

Registry starts empty; content types registered in W4–W7 as needed.

---

### ✅ Checkpoint W-Alpha

Foundation complete. Verify before proceeding:
- `FormProvider` lifecycle tests all pass
- `ContentTypeLayout` renders correctly in isolation
- Registry types compile
- `npm run build` clean

---

### W4 — SingleTypePage with Locale + Layout

**What**: New `SingleTypePage` replaces `SingleTypePanel`. Complete single-type edit with new layout and form lifecycle.

**Locale process:**
1. `useLocales()` fetches available locales
2. Local `locale` state initialises to `locales[0]`
3. `<select aria-label="Locale">` in `renderActions` only when `locales.length > 1`
4. Switching locale → `useDocuments` re-fetches for new locale → `FormProvider` `values` prop syncs inputs → `isDirty` auto-resets to `false`
5. Save/Publish/Unpublish mutations forward `locale: activeLocale`

**Files changed:**
- `src/pages/admin/panels/SingleTypePage.tsx` — new
- `src/pages/admin/panels/__tests__/SingleTypePage.test.tsx` — new

**Acceptance criteria:**
- [ ] All fields pre-filled from draft data on load
- [ ] Save disabled on initial load; enabled after any field edit
- [ ] Successful save: success toast + form reset to server data + Save disabled
- [ ] Failed save: error toast; edited values preserved
- [ ] Locale selector hidden when ≤ 1 locale
- [ ] Locale selector shown when > 1 locale
- [ ] Switching locale resets form and disables Save

**Verification:** `npm run test -- SingleTypePage` passes.

---

### W5 — CollectionListPage with Column Registry

**What**: Registry-driven columns replace hard-coded first-field display.

**Column rendering rules:**
- `text` → string value
- `boolean` → `✓` or `—`
- `number` → numeric string
- `image` → `<img>` thumbnail (src = field value as string)
- **Fallback**: when no registry entry for the slug defines `columns`, render first field as text + Status column

**Files changed:**
- `src/pages/admin/panels/CollectionListPage.tsx` — update
- `src/pages/admin/panels/__tests__/CollectionListPage.test.tsx` — update

**Acceptance criteria:**
- [ ] Registry `columns` drive table headers and cell rendering
- [ ] Each type renders its correct cell format
- [ ] Fallback to first-field + Status when no registry entry
- [ ] "Add new item" navigates to `/admin/content-type/collection-type/:slug/:id`
- [ ] Edit link navigates to detail page
- [ ] Delete fires `window.confirm` then `useDeleteDocument`

**Verification:** `npm run test -- CollectionListPage` passes.

---

### W6 — CollectionDetailPage/Panel with Locale + Layout

**What**: Refactor `CollectionDetailPanel` to use `ContentTypeLayout` and full form lifecycle. Locale process identical to `SingleTypePage`.

**Files changed:**
- `src/pages/admin/panels/CollectionDetailPanel.tsx` — refactor in-place
- `src/pages/admin/panels/CollectionDetailPage.tsx` — minor (slug-based registry lookup)
- `src/pages/admin/panels/__tests__/CollectionDetailPanel.test.tsx` — update

**Acceptance criteria:**
- [ ] Full single-type form lifecycle applies (dirty, toasts, reset)
- [ ] Locale selector shown/hidden under same condition as `SingleTypePage`
- [ ] Switching locale resets form and disables Save
- [ ] Back link navigates to `/admin/content-type/collection-type/:slug`
- [ ] Publish/Unpublish buttons behave as before

**Verification:** `npm run test -- CollectionDetailPanel` passes.

---

### ✅ Checkpoint W-Beta

All content-type pages functionally complete. Verify before proceeding:
- Single-type lifecycle: save → toast → reset; locale switching
- Collection list: registry columns + fallback; column type rendering
- Collection detail: same lifecycle as single-type; locale switching
- `npm run build` clean; full test suite passes

---

### W7 — Sidebar + Router Migration

**What**: Update routes and sidebar links to the new path structure. Remove old routes. Wire `SiteHomepagePanel` through the registry.

**Route changes:**

| Old | New |
|---|---|
| `/admin/content-types/:slug` (single) | `/admin/content-type/single-type/:slug` |
| `/admin/content-types/:slug` (collection) | `/admin/content-type/collection-type/:slug` |
| `/admin/content-types/:slug/:id` | `/admin/content-type/collection-type/:slug/:id` |

`SiteHomepagePanel` currently uses a one-off `SiteSiteHomepagePanelWrapper` in `router.tsx`. After this task: register the `site-settings` slug in `contentTypeRegistry` with `SiteHomepagePanel` as the `wrapper` component so the generic `SingleTypePage` route handles it without a bespoke wrapper.

**Files changed:**
- `src/pages/admin/layout/Sidebar.tsx` — update `NavLink` hrefs
- `src/router.tsx` — new routes with `React.lazy`, remove old routes, remove `SiteSiteHomepagePanelWrapper`
- `src/content-type-registry/index.ts` — register `site-settings` slug
- `src/pages/admin/layout/__tests__/AdminLayout.test.tsx` — update route assertions if any

**Acceptance criteria:**
- [ ] Sidebar single-type links → `/admin/content-type/single-type/:slug`
- [ ] Sidebar collection-type links → `/admin/content-type/collection-type/:slug`
- [ ] Old `/admin/content-types/...` routes removed
- [ ] `SiteHomepagePanel` still renders when navigating to its slug via new path
- [ ] Component code loaded via `React.lazy` (not eagerly imported)
- [ ] `ProtectedRoute` auth guard still applies to all new routes
- [ ] `npm run build` clean; no dead imports

**Verification:** `npm run test` full suite passes; `npm run build` clean.

---

### ✅ Checkpoint W-Final

Full migration complete. Verify:
- All acceptance criteria from SPEC.md §7.9 checked off
- `SiteHomepagePanel` works end-to-end via new routing
- Sidebar links and routing are consistent
- No old `/admin/content-types/...` routes remain
- `npm run lint && npm run build && npm run test` all green

---

## File Map

```
src/
  components/
    content-type/
      ContentTypeLayout.tsx            ← NEW (W2)
      __tests__/
        ContentTypeLayout.test.tsx     ← NEW (W2)
    form/
      FormStateContext.tsx             ← UPDATED (W1: adds isDirty)
      FormProvider.tsx                 ← UPDATED (W1: success/reset/error)
      __tests__/
        FormProvider.test.tsx          ← UPDATED (W1)
  content-type-registry/
    index.ts                           ← NEW (W3), UPDATED (W7: registers site-settings)
  pages/admin/
    layout/
      Sidebar.tsx                      ← UPDATED (W7: new hrefs)
    panels/
      SingleTypePage.tsx               ← NEW (W4)
      CollectionListPage.tsx           ← UPDATED (W5: registry columns)
      CollectionDetailPanel.tsx        ← UPDATED (W6: ContentTypeLayout)
      CollectionDetailPage.tsx         ← UPDATED (W6)
      __tests__/
        SingleTypePage.test.tsx        ← NEW (W4)
        CollectionListPage.test.tsx    ← UPDATED (W5)
        CollectionDetailPanel.test.tsx ← UPDATED (W6)
  router.tsx                           ← UPDATED (W7: new routes)
```

## Notes

- `SingleTypePanel.tsx` stays on disk until W7 is complete. Do not delete it before `SingleTypePage` is wired into the router.
- `CollectionDetailPanel` is refactored in-place (not replaced) since `CollectionDetailPage` delegates to it.
- Use `axios-mock-adapter` (already in devDependencies) for API mocking in component tests — no MSW setup exists in this project.
