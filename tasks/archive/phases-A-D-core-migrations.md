# Archive — Phases A–D: Core Feature Migrations

> All phases completed. Archived from `tasks/todo.md`.

---

## Phase A — Schema-as-Code
- [x] A1 JSON schema loader (`content-types/*.json` → `ContentTypeDefinition`) + fixture tests
- [x] A2 Sync usecase (create/update/delete reconciliation + cascade) + tests
- [x] A3 Wire Sync into `cmd/server/main.go` startup
- [x] A4 Remove ContentType Create/Update/Delete (handlers, routes, unused FE hooks, stale tests)
- [x] ✅ Checkpoint A: boot syncs JSON defs → Mongo; no API/UI path to mutate ContentType structure

## Phase B — Draft/Publish Remodel
- [x] B1 Document entity + repository: `entryId`/`version` records, audit fields, mocks updated
- [x] B2 Document usecase: Save/Publish/Unpublish/computed Status/GetForEdit/GetPublished + tests
- [x] B3 Document handlers: entryId-addressed routes + public read path (GET /api/public/documents/{id}, no auth) + tests
- [x] B4 FE: `useDocuments.ts` hooks updated to tri-state status + tests
- [x] B5 FE: panels (SingleTypePanel, CollectionDetailPanel, CollectionListPage) tri-state badge + tests
- [x] ✅ Checkpoint B: Save never affects public read; Publish syncs it; tri-state badge correct; full test suite passes

## Phase C — Content-Type Kind UX
- [x] C1 Single-type auto-singleton creation wired into Sync + test
- [x] C2 Sidebar grouped by kind (Single Types / Collection Types) + test
- [x] ✅ Checkpoint C: new single-type definition auto-creates its singleton; sidebar visually grouped

## Phase D — Storage: S3 Adapter
- [x] D1 S3 adapter implementing `StorageAdapter` + test
- [x] D2 Config-driven adapter selection (env var) in `cmd/server/main.go` + test
- [x] ✅ Checkpoint D (Final): both adapters pass tests, selectable via env var (verified live, both providers boot); full suite green; full E2E verified live against running server (register/promote test admin → draft save → public 404 → publish → public 200 synced → new draft save → status "modified" while public read still serves old data) → test artifacts cleaned up

## Follow-up
- [x] `/spec` pass to record the kept "Unpublish" behavior in SPEC.md (done as part of §7 refactor spec)
