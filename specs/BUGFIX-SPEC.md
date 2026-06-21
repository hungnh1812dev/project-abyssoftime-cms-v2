# SPEC — Bug Fixes (v1.8)

Six bugs across auth, content API, frontend inputs, and response formatting.

---

## Bug 1 — User entity missing `id` and `documentId` on creation

### Problem

When registering the first super_admin user, the `User` entity is created without `ID` or `DocumentID`. The `Register` usecase (`apps/api/internal/usecase/auth/auth_usecase.go:51`) builds the user struct with only `Email`, `PasswordHash`, `Role`, and `RoleID` — no UUID generation. The GORM repository (`apps/api/internal/infrastructure/gormdb/user_repository.go:24`) calls `db.Create(user)` directly, and GORM does **not** auto-generate UUIDs for string primary keys. Result: empty `id` and `document_id` columns in the database.

### Root Cause

`auth_usecase.go` → `Register()` never assigns `user.ID` or `user.DocumentID` before calling `repo.Create()`.

### Fix

Generate UUIDs for both fields in the `Register` method before calling `repo.Create`:

```go
// auth_usecase.go — Register()
user := &entity.User{
    ID:           uuid.New().String(),   // ADD
    DocumentID:   uuid.New().String(),   // ADD
    Email:        email,
    PasswordHash: string(hash),
    Role:         entity.RoleSuperAdmin,
    RoleID:       saRole.DocumentID,
    CreatedAt:    time.Now().UTC(),       // ADD — explicit timestamp
}
```

### Scope

| File | Change |
|------|--------|
| `apps/api/internal/usecase/auth/auth_usecase.go` | Add UUID generation for `ID`, `DocumentID`, and `CreatedAt` in `Register()` |
| `apps/api/internal/usecase/auth/auth_usecase_test.go` | Assert returned user has non-empty `ID` and `DocumentID` |

### Acceptance Criteria

- [ ] After register, `User.ID` is a valid UUID string
- [ ] After register, `User.DocumentID` is a valid UUID string
- [ ] After register, `User.CreatedAt` is set
- [ ] Existing tests still pass
- [ ] `Login` can find the user by `ID` (used in JWT → `RefreshToken` → `FindByID`)

---

## Bug 2 — Register page re-shown after successful admin registration

### Problem

Flow: visit `/register` (no admin) → register super_admin → success → **register page re-appears** showing "Create guest account" instead of navigating to `/login`.

### Root Cause

Two issues:

1. **Race between invalidation and navigation:** In `RegisterPage.tsx:39`, `onSuccess` calls `queryClient.invalidateQueries({ queryKey: ['auth-setup'] })` then `navigate('/login')`. The invalidation triggers an async refetch of `auth-setup`. When the refetch completes (`adminExists: true`), React re-renders the page with "Create guest account" heading. Depending on timing, the user sees this flash before (or instead of) the navigation to `/login`.

2. **No route guard:** The `/register` route is always accessible. Even after an admin exists, a user can visit `/register` and see the "Create guest account" form — but the API will reject it with 403 because `Register()` blocks when `hasSA` is true.

### Fix

**RegisterPage (`apps/web/src/pages/auth/RegisterPage.tsx`):**

Add a redirect when `adminExists === true` — if an admin already exists, the register page should not be accessible:

```tsx
// After setupLoading check
if (adminExists) {
  return <Navigate to="/login" replace />
}
```

**onSuccess:** Remove the `queryClient.invalidateQueries` call (unnecessary since we navigate away), or keep it for ProtectedRoute cache freshness but ensure `navigate('/login')` runs first.

### Scope

| File | Change |
|------|--------|
| `apps/web/src/pages/auth/RegisterPage.tsx` | Add `adminExists` redirect guard; simplify `onSuccess` |
| `apps/web/src/pages/auth/__tests__/RegisterPage.test.tsx` | Test: redirects to `/login` when `adminExists === true` |

### Acceptance Criteria

- [ ] First visit to `/register` (no admin) shows "Set up admin account"
- [ ] After successful registration, user is redirected to `/login`
- [ ] User never sees "Create guest account" after registering super_admin
- [ ] Visiting `/register` when admin already exists redirects to `/login`

---

## Bug 3 — JsonInput and RichTextInput lose data on save

### Problem

After saving a document that contains `json` or `richtext` fields, the input values are lost/reset to empty.

### Root Cause (Investigation Required)

The `FormProvider` (`apps/web/src/components/form/FormProvider.tsx`) uses `useForm({ values })` to sync server data into the form. After a successful save:

1. `methods.reset(methods.getValues())` resets the form with current values
2. `queryClient.invalidateQueries()` triggers a refetch
3. The refetch returns new `values`, which `useForm({ values })` syncs into the form

The likely root cause involves how `Controller`-based inputs (JsonInput, RichTextInput) interact with `useForm({ values })`:

- **JsonInput** (`apps/web/src/components/form/inputs/JsonInput.tsx`): Uses internal `useState` to track raw text (`rawValue`). When `field.value` changes via form reset, the comparison `fieldValue !== prevFieldValue` uses reference equality. If the server returns a structurally identical but referentially different object, the sync triggers. But if the server response structure differs from what was sent (e.g., the `data` wrapper is different), `field.value` may come back as `undefined`, causing `serialize(undefined)` → `''`.

- **RichTextInput** (`apps/web/src/components/form/inputs/RichTextInput.tsx`): CKEditor receives `data={field.value as string}`. If `field.value` becomes `undefined` after form reset/re-sync, CKEditor renders empty content.

**Primary suspect:** The form field names use dot-notation (e.g., `data.myJsonField`). After save, when the query refetches and returns the document, the `values` object structure may not align with the dot-notation paths, causing `field.value` to resolve to `undefined`.

### Fix

Investigate the exact data flow by:
1. Checking how `mutationFn` sends data vs. how the query response structures it
2. Verifying that `useForm({ values })` correctly maps nested `data.*` paths after refetch
3. If the issue is value-sync timing: use `useForm({ values, resetOptions: { keepDirtyValues: true } })` or defer the reset until after refetch completes
4. If the issue is reference equality in JsonInput: use deep comparison instead of `!==`

### Scope

| File | Change |
|------|--------|
| `apps/web/src/components/form/FormProvider.tsx` | Fix value sync after mutation success |
| `apps/web/src/components/form/inputs/JsonInput.tsx` | Fix value comparison / re-sync logic |
| `apps/web/src/components/form/inputs/RichTextInput.tsx` | Ensure `field.value` fallback to `""` not `undefined` |

### Acceptance Criteria

- [ ] Save a document with a `json` field → field retains its value after save
- [ ] Save a document with a `richtext` field → field retains its value after save
- [ ] Editing after save works correctly (isDirty triggers, re-save works)
- [ ] Form reset after save does not flash empty inputs

---

## Bug 4 — `contentTypeId` and `status` exposed in API response

### Problem

The REST API document responses include `contentTypeId` and `status` as top-level fields. These are internal/computed fields that should not be exposed to API consumers:
- `contentTypeId` is an internal DB foreign key — consumers use the `slug` from the URL
- `status` is a computed value (draft/modified/published) that is only meaningful in the admin context

### Root Cause

- `entrySummary` struct (`apps/api/internal/delivery/http/handler/document_handler.go:47`) explicitly includes `ContentTypeID` and `Status` fields with JSON tags
- `toSummary()` maps both fields from the document entity
- The GraphQL `docToMap()` (`apps/api/graphql/dynamic/resolver_factory.go:457`) also includes `status` in the map
- The GraphQL `buildObjectType()` includes `status` in the generated type fields

### Fix

**REST API:** Remove `ContentTypeID` and `Status` from `entrySummary`. Create a separate admin-only response type if status is needed for admin endpoints, or keep status only in admin endpoints via a different response struct.

**GraphQL:** Remove `status` from `buildObjectType` fields. If needed for admin queries, add an optional `status` field only when the request is authenticated.

**gRPC:** `toProtoDoc()` currently includes both `ContentTypeId` and `Status` — remove or deprecate these proto fields.

### Scope

| File | Change |
|------|--------|
| `apps/api/internal/delivery/http/handler/document_handler.go` | Remove `contentTypeId` and `status` from public response structs; keep for admin if needed |
| `apps/api/graphql/dynamic/resolver_factory.go` | Remove `status` from generated object types and `docToMap` |
| `apps/api/internal/delivery/grpc/document_service.go` | Remove `contentTypeId` and `status` from `toProtoDoc` |
| Test files for all three | Update assertions |

### Acceptance Criteria

- [ ] REST document responses do not contain `contentTypeId` or `status`
- [ ] GraphQL document types do not expose `status`
- [ ] gRPC `Document` message omits `contentTypeId` and `status`
- [ ] Admin edit screens still function (status may be needed internally — verify FE usage)

---

## Bug 5 — System fields should be inside `data` in API response

### Problem

Current REST response for a single document:
```json
{
  "documentId": "abc",
  "contentTypeId": "xyz",
  "data": { "title": "Hello" },
  "status": "draft",
  "locale": "en",
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-01T00:00:00Z",
  "createdBy": "user1",
  "updatedBy": "user1"
}
```

Expected response — system fields merged into `data`:
```json
{
  "data": {
    "documentId": "abc",
    "locale": "en",
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-01T00:00:00Z",
    "createdBy": "user1",
    "updatedBy": "user1",
    "title": "Hello"
  }
}
```

This aligns with the GraphQL response format where queries return `{ data: <Type> }`.

### Root Cause

`entrySummary` and `paginatedListItem` structs in `document_handler.go` place system fields (`documentId`, `locale`, `createdAt`, `updatedAt`, `createdBy`, `updatedBy`) at the same level as `data`, instead of merging them into `data` and wrapping in a `{ data: ... }` envelope.

### Fix

Restructure the REST API response to wrap document data in a `data` envelope, merging system fields with content data:

```go
type documentResponse struct {
    Data map[string]any `json:"data"`
}

func toResponse(doc *entity.Document) documentResponse {
    merged := make(map[string]any, len(doc.Data)+6)
    for k, v := range doc.Data {
        merged[k] = v
    }
    merged["documentId"] = doc.DocumentID
    merged["locale"] = doc.Locale
    merged["createdAt"] = doc.CreatedAt
    merged["updatedAt"] = doc.UpdatedAt
    merged["createdBy"] = doc.CreatedBy
    merged["updatedBy"] = doc.UpdatedBy
    return documentResponse{Data: merged}
}
```

For paginated list responses, each item follows the same structure:
```json
{
  "items": [
    {
      "data": {
        "documentId": "abc",
        "locale": "en",
        "createdAt": "...",
        "updatedAt": "...",
        "title": "..."
      }
    }
  ],
  "total": 42,
  "start": 0,
  "size": 20
}
```

### Scope

| File | Change |
|------|--------|
| `apps/api/internal/delivery/http/handler/document_handler.go` | Restructure response types to wrap in `data` envelope |
| `apps/api/internal/delivery/http/handler/document_handler_test.go` | Update response assertions |
| `apps/api/graphql/dynamic/resolver_factory.go` | `docToMap` already merges — verify consistency |
| `apps/web/src/**` (hooks, pages) | Update frontend to read from `data` wrapper if response shape changes |

### Acceptance Criteria

- [ ] Single document GET returns `{ data: { documentId, locale, createdAt, updatedAt, createdBy, updatedBy, ...contentFields } }`
- [ ] Paginated list returns items with same `data` structure
- [ ] No `contentTypeId` or `status` in the response (combines with Bug 4)
- [ ] Frontend reads new response shape correctly
- [ ] GraphQL response format remains consistent

---

## Bug 6 — API + GraphQL should default to returning published documents

### Problem

GraphQL queries currently return **draft** documents by default:
- `Query.<slug>()` calls `GetForEdit()` which fetches the draft record
- `Query.<slugList>()` calls `GetAllPaginated()` which fetches draft records

REST public route (`/api/public/document-manager/:slug/:documentId`) already correctly returns published documents via `GetPublished()`.

Per the spec boundary rule: **"Never let public read API return draft data"** — GraphQL queries violate this.

### Root Cause

- `resolver_factory.go` line 234: collection single query calls `f.docUC.GetForEdit(...)` — returns draft
- `resolver_factory.go` line 253: collection list query calls `f.docUC.GetAllPaginated(...)` — returns drafts
- `resolver_factory.go` line 380: single-type query calls `f.docUC.GetSingleType(...)` — returns draft

These should default to published documents for unauthenticated requests (public API consumers), with an opt-in parameter to request drafts for authenticated admin users.

### Fix

**GraphQL queries (default: published):**

1. Single document query (`Query.<slug>`): Call `GetPublished()` instead of `GetForEdit()` by default. Add optional `status: "draft"` argument that requires authentication.

2. List query (`Query.<slugList>`): Add a new `FindPublishedByContentTypePaginated` method to `DocumentRepository` and a corresponding `GetAllPublishedPaginated` usecase method. Default to published records. Add optional `status: "draft"` filter that requires authentication.

3. Single-type query (`Query.<slug>`): Call `GetPublishedSingleType()` (new method) by default. Add `status: "draft"` opt-in.

**New repository method:**

```go
FindPublishedByContentTypePaginated(ctx context.Context, contentTypeSlug string, start, size int, locale string) ([]*entity.Document, int64, error)
```

**New usecase methods:**

```go
GetPublishedPaginated(ctx context.Context, slug string, start, size int, locale string, fields []entity.FieldDefinition) ([]*entity.Document, int64, error)
GetPublishedSingleType(ctx context.Context, slug, locale string, fields []entity.FieldDefinition) (*entity.Document, error)
```

**REST public route:** Already correct — no changes needed.

### Scope

| File | Change |
|------|--------|
| `apps/api/internal/domain/repository/document_repository.go` | Add `FindPublishedByContentTypePaginated` |
| `apps/api/internal/domain/repository/mock/document_repository.go` | Add mock |
| `apps/api/internal/infrastructure/gormdb/document_repository.go` | Implement `FindPublishedByContentTypePaginated` |
| `apps/api/internal/infrastructure/gormdb/document_repository_test.go` | Test new method |
| `apps/api/internal/usecase/document/document_usecase.go` | Add `GetPublishedPaginated`, `GetPublishedSingleType` |
| `apps/api/internal/usecase/document/document_usecase_test.go` | Test new methods |
| `apps/api/graphql/dynamic/resolver_factory.go` | Default to published; add `status` filter with auth check |
| `apps/api/graphql/dynamic/resolver_factory_test.go` | Test default-published behavior |

### Acceptance Criteria

- [ ] Unauthenticated GraphQL `Query.<slug>` returns published document (or null if not published)
- [ ] Unauthenticated GraphQL `Query.<slugList>` returns only published documents
- [ ] Unauthenticated GraphQL `Query.<slug>` for single-type returns published document
- [ ] Authenticated request with `status: "draft"` returns draft documents
- [ ] REST public route behavior unchanged (already correct)
- [ ] Admin REST routes unchanged (they serve draft for editing)

---

## Implementation Order

Dependencies between bugs determine the order:

| Order | Bug | Rationale |
|-------|-----|-----------|
| 1 | Bug 1 | Self-contained auth fix, no dependencies |
| 2 | Bug 2 | Self-contained FE fix, no dependencies |
| 3 | Bug 4 + Bug 5 | Together — both change the response shape. Do them as one change to avoid double-migration of frontend code |
| 4 | Bug 3 | Depends on Bug 4+5 — input data loss may be partly caused by response shape; fix response first, then verify if input issue persists |
| 5 | Bug 6 | Largest scope — new repo/usecase methods + GraphQL resolver changes |

---

## Testing Strategy

- **Backend:** Unit tests for each changed usecase method; handler tests with `httptest` for response shape; GORM repository tests for new queries
- **Frontend:** Component tests for RegisterPage redirect; manual testing for JsonInput/RichTextInput data persistence
- **Integration:** End-to-end manual test: register → login → create document with json/richtext fields → save → verify data persists → publish → verify GraphQL returns published

---

## Boundaries

| Rule | Detail |
|------|--------|
| **Always** | Generate UUID for `User.ID` and `User.DocumentID` on registration |
| **Always** | Redirect `/register` to `/login` when admin already exists |
| **Always** | Wrap REST document responses in `{ data: { ... } }` envelope |
| **Always** | GraphQL queries return published documents by default |
| **Never** | Expose `contentTypeId` or `status` in public API responses |
| **Never** | Return draft data from unauthenticated GraphQL queries |
| **Ask first** | Whether admin GraphQL queries need a `status: "draft"` filter parameter |
