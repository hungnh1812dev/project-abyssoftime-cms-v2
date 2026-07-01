# RULES — Document

**Scope:** Document entity, draft/publish workflow, single-type and collection-type CRUD, pagination, API contracts, duplicate documents.

---

## 1. Document Rules

### 1.1 Document Entity
- `DocumentID`: UUID v4 — the primary domain identifier
- `Version`: `"draft"` or `"published"` — two separate records per entry
- `Fields`: `map[string]any` — content data (tagged `gorm:"-"` for GORM)
- `GormID`: auto-increment uint for display ordering
- `Locale`: defaults to `"en"` (or default locale)
- Higher layers only use `documentId` + content-type `slug` — **NEVER** MongoDB `_id`

### 1.2 Draft/Publish Workflow
- Every entry = two separate records (draft + published) sharing the same `documentId`
- **Save** → upsert draft record only. **NEVER** touch published.
- **Publish** → copy `draft.data` to published record, set `publishedAt = now()`
- **Unpublish** → delete published record
- **Status** computed, never stored:
  - `draft`: no published record exists
  - `modified`: `draft.updatedAt > published.updatedAt`
  - `published`: timestamps match
- Documents only created on explicit Save — no auto-creation

### 1.3 Single-Type Rules
- At most one entry per content type
- No auto-created singleton — first Save creates it
- UI: edit + Save + Publish only. No create/delete.
- **NEVER** expose DELETE for single-type documents
- **NEVER** include `documentId` in single-type URLs
- GET returns 404 when no document exists (FE shows empty form)
- PUT creates on first save, updates on subsequent saves

### 1.4 Collection-Type Rules
- Zero or more entries, each with own `documentId`
- List + create/edit/delete, each with independent draft/published pair
- Pagination via `PaginationInput` supporting two modes:
  - Offset mode: `{start: Int, limit: Int}` — both optional, defaults `start=0, limit=10`
  - Page mode: `{page: Int!, pageSize: Int!}` — both required if one is provided
  - `limit: -1` returns all documents; response `pageSize = total`
  - Default (no input): `page=1, pageSize=10`
- **NEVER** allow `limit` or `pageSize` above 100 (except `limit: -1` = no limit)
- **NEVER** mix offset and page modes in the same request (validation error)
- Support `orderBy` and `sortDir` query params
- **Bulk create + publish** (`POST .../:slug/bulk`): accepts `{ "items": [{ "data": {...} }, ...] }`, 1–100 items, one `?locale=` for the whole request (no per-item locale). Each item is created as a draft and immediately published, in submission order.
  - **All-or-nothing via rollback, not a DB transaction**: items are processed sequentially through the existing `Save`/`Publish` flow. If any item fails, every document already committed earlier in that batch is deleted (via the existing `Delete` usecase method, which removes draft + published + components for all locales) before the error is returned. This is a compensating rollback, not an atomic multi-document transaction — MongoDB standalone deployments (the assumed default here) don't support that, and this codebase has no transaction abstraction for either DB backend.
  - Requires **both** `content:create` and `content:publish` (two chained `GinRequirePermission` calls on the route)

---

## 2. API Contract Rules

### 2.1 Input Validation (Handler-Level)
- Slug: `^[a-z0-9]+(?:-[a-z0-9]+)*$` — applied on every request. 400 on invalid.
- DocumentID: UUID v4 format — applied on every request. 400 on invalid.
- Return 400 (not 500) for invalid slug or documentID format

### 2.2 REST Routes — Content Types
| Method | Route | Permission |
|---|---|---|
| `GET` | `/api/content-types` | `content_types:read` |
| `GET` | `/api/content-types/:identifier` | `content_types:read` |
| `PATCH` | `/api/content-types/:slug/list-fields` | `content_types:read` |

### 2.3 REST Routes — Single-Type Documents
| Method | Route | Permission |
|---|---|---|
| `GET` | `/api/document-manager/single-type/:slug` | `content:read` |
| `PUT` | `/api/document-manager/single-type/:slug` | `content:update` |
| `POST` | `/api/document-manager/single-type/:slug/publish` | `content:publish` |
| `POST` | `/api/document-manager/single-type/:slug/unpublish` | `content:unpublish` |

### 2.4 REST Routes — Collection-Type Documents
| Method | Route | Permission |
|---|---|---|
| `GET` | `/api/document-manager/collection-type/:slug` | `content:read` |
| `GET` | `/api/document-manager/collection-type/:slug/:documentId` | `content:read` |
| `POST` | `/api/document-manager/collection-type/:slug` | `content:create` |
| `POST` | `/api/document-manager/collection-type/:slug/bulk` | `content:create` **and** `content:publish` |
| `PUT` | `/api/document-manager/collection-type/:slug/:documentId` | `content:update` |
| `DELETE` | `/api/document-manager/collection-type/:slug/:documentId` | `content:delete` |
| `POST` | `.../:documentId/publish` | `content:publish` |
| `POST` | `.../:documentId/unpublish` | `content:unpublish` |
| `POST` | `.../:documentId/duplicate` | `content:create` |

### 2.5 REST Response Shapes
- Document response: `{ "data": { "documentId", "status", "locale", ...systemFields, ...contentFields } }`
- Paginated list: `{ "items": [...], "total", "start", "size", "listFields" }`
- List items contain **only** selected fields (from `listFields`) — **NEVER** return full data in paginated lists
- `updatedByName` resolved server-side from User's `DisplayName`
- `status` and `contentTypeId` excluded from public API responses

### 2.6 Public Read Route
- `GET /api/public/document-manager/:slug/:documentId` — no auth required
- Returns published record **ONLY** — 404 if not published
- **NEVER** return draft data through public API

---

## 3. Duplicate Document Rules

### 3.1 Behavior
- Creates full copy as new draft with fresh `documentId`
- Media references shared (same documentIds) — no file re-upload
- All component data fully copied; new `componentId` per component
- Active locale only — does not duplicate other locales
- Always draft-only — **NEVER** create published version for duplicate
- Navigate to new document's edit page after duplication

### 3.2 Fields NOT Copied
- `documentId` (new UUID generated)
- `gormId` (auto-incremented)
- `createdAt`, `updatedAt` (set to now)
- `createdBy`, `updatedBy` (set to current user)
- `publishedAt`, `publishedBy` (not set)
- Published version record (not created)

---

## 4. Boundaries Summary

| Rule | Detail |
|---|---|
| **Always** | Validate slug format at both usecase and handler levels |
| **Always** | Validate documentID as UUID v4 before passing to usecase |
| **Always** | Return 404 for single-type GET when no document exists |
| **Always** | Include computed `status` in every document response |
| **Always** | Project `data` in collection lists — never full data |
| **Always** | Batch-fetch published records for status computation (no N+1) |
| **Always** | Validate data shape at usecase (object vs array based on repeatable) |
| **Always** | Sanitize field values before save: coerce `""` to nil for number, boolean, media fields |
| **Never** | Expose DELETE for single-type documents |
| **Never** | Allow `size` above 100 on collection list |
| **Never** | Return draft data through public API |
| **Never** | Include `documentId` in single-type URLs |
| **Ask first** | Changing default `size` or max from 100 |
