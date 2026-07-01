# Todo — personal-cms (project-abyssoftime-cms-v2)

> Completed phases archived in `tasks/archive/`. This file tracks only current and upcoming work.

---

## Phase BC — Bulk Create + Publish (Collection-Type Documents)

Spec: [specs/bulk-document-create-publish.md](../specs/bulk-document-create-publish.md). Plan: [plan.md](plan.md) (Phase BC section).

- [x] T1 Usecase: `BulkCreateAndPublish` (sequential Save+Publish per item, rollback via `Delete` on first failure) + tests (4 new tests: all-valid, save-failure rollback, publish-failure rollback, invalid locale)
- [x] ✅ Checkpoint 1: `go test ./internal/usecase/document/...` passes (all 4 new tests green)
- [x] T2 Handler: `BulkCreateCollection` + `bulkCreateRequest`/`bulkCreateResponse` types + route (`POST /:slug/bulk`, dual `content:create`+`content:publish` permission) + tests (5 new subtests: 201 success, 400 empty, 400 over-cap, 400 invalid body, 422 usecase error)
- [x] ✅ Checkpoint 2: `go vet ./...` + `make test-api` + `go build ./...` all green
- [x] T3 Update `rules/document.md` (§2.4 route table + §1.4 bulk semantics note)
- [x] ✅ Checkpoint 3 (Final): user ran a live test against `en-vocab` on Postgres — surfaced that items without the `data` wrapper silently create empty published documents (no validation error). Diagnosed as a request-shape issue, not a save/publish bug (confirmed via direct DB inspection — 0 components, empty text columns). User declined adding stricter per-item validation and declined cleanup of the 2 test documents; documented as an accepted "Known gap" in `specs/document.md` §9

---

## Archive Index

| Archive | Phases | Status |
|---------|--------|--------|
| [phases-0-5-foundation.md](archive/phases-0-5-foundation.md) | 0–5 | ✅ Complete |
| [phases-A-D-core-migrations.md](archive/phases-A-D-core-migrations.md) | A–D | ✅ Complete |
| [phase-M-media-forms.md](archive/phase-M-media-forms.md) | M | ✅ Complete |
| [phases-W-X-Y-web-api.md](archive/phases-W-X-Y-web-api.md) | W, X, Y | ✅ Complete |
| [phases-Z-to-GQ.md](archive/phases-Z-to-GQ.md) | Z, ZZ, UI, S, DL, RC, AX, HP, LW, GF, GQ | ✅ Complete (2 stray manual-verification checkboxes left unchecked from source) |
