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
