# Archive — Phases W, X, Y: Web Refactor + Media Delete + API Restructure

> All phases completed. Archived from `tasks/todo.md`.

---

## Phase W — Web Content-Type Management System (Refactor)

### Foundation
- [x] W1 FormProvider lifecycle: `isDirty` on context, success toast + `reset(newData)`, error toast, Save disabled when clean
- [x] W2 `ContentTypeLayout` component: `renderHeader`, `renderActions`, title + status badge
- [x] W3 Content-type registry: `src/content-type-registry/index.ts`, types only, no component imports
- [x] ✅ Checkpoint W-Alpha: FormProvider tests pass, ContentTypeLayout renders, registry compiles, `npm run build` clean

### Content-Type Pages
- [x] W4 `SingleTypePage`: locale state + selector, `ContentTypeLayout` wrapper, full form lifecycle
- [x] W5 `CollectionListPage`: registry-driven columns, column type rendering (`text`/`boolean`/`number`/`image`), first-field fallback
- [x] W6 `CollectionDetailPanel` / `CollectionDetailPage`: `ContentTypeLayout`, locale + full form lifecycle, back link
- [x] ✅ Checkpoint W-Beta: single-type and collection-type pages complete, `npm run build` + tests green

### Routing Migration
- [x] W7 Sidebar + Router: new route paths, `SiteHomepagePanel` via registry, remove old routes, `React.lazy` for all panels
- [x] ✅ Checkpoint W-Final: all SPEC.md §7.9 criteria met, full suite green (126/126)

---

## Phase X — Delete Media Asset in MediaLibrary

- [x] X1 Backend: `UseCase.Delete` (storage-first) + `MediaHandler.Delete` + `DELETE /api/media/{id}` route + 4 usecase tests + 3 handler tests
- [x] X2 Frontend: `useDeleteMedia()` mutation hook in `useMedia.ts` (mirrors `useUploadMedia` pattern)
- [x] X3 Frontend: MediaLibrary delete UX (hover trash + `pendingDeleteId` inline confirm) + 4 component tests
- [x] ✅ Checkpoint X: `go test ./...` all green (14 packages), `vitest run` 156/156 green

---

## Phase Y — Document Manager API Restructure & Paginated Collections (SPEC §10)

### BE Foundation
- [x] Y1 Entity + Schema + JSON: `ListFields` on ContentType entity, schema loader parse + validate, sync carry-through, blog-posts.json listFields
- [x] Y2 Repository: `FindDraftsByContentTypePaginated` + `FindPublishedByDocumentIDs` interface + mock + MongoDB impl

### BE Single-Type + Collection-Type Flow
- [x] Y3+Y4 Usecase: single-type methods + `GetAllPaginated` (paginated drafts + batch status) + tests

### BE Handler + Routes
- [x] Y5 Handler rewrite (11 kind-specific methods + `ctUC` dep + `projectData`) + route migration in main.go + tests
- [x] ✅ Checkpoint A: all Go tests pass, new routes respond correctly, old flat routes removed

### FE Migration
- [x] Y6+Y7 Types + hooks + components: kind-specific hooks, paginated CollectionListPage, schema-derived columns, delete old useDocuments.ts
- [x] ✅ Checkpoint B: FE build passes, stale refs cleaned

### Cleanup
- [x] Y8 Final cleanup: grep stale refs, full test suite green, clean build
