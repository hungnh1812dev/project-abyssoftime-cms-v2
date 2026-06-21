# Bug Fix Plan — v1.8

Full spec: [specs/BUGFIX-SPEC.md](../specs/BUGFIX-SPEC.md)

---

## Dependency Graph

```
B1 (User id/documentId)     — independent, BE only
B2 (Register page guard)    — independent, FE only
    ↓
[Checkpoint 1: B1+B2 verified]
    ↓
B4+B5 (Response shape)      — BE + FE, combined
    ↓
[Checkpoint 2: response shape verified end-to-end]
    ↓
B3 (Input data loss)         — FE, depends on B4+B5 (may be resolved by it)
    ↓
[Checkpoint 3: input persistence verified]
    ↓
B6 (GraphQL default published) — BE + GraphQL
    ↓
[Checkpoint 4: all tests green, full smoke test]
```

---

## B1: User entity missing `id` and `documentId`

**Root cause:** `Register()` in `auth_usecase.go:51` never sets `user.ID` or `user.DocumentID`. GORM doesn't auto-generate string UUIDs.

**Changes:**

| File | Change |
|------|--------|
| `apps/api/internal/usecase/auth/auth_usecase.go` | In `Register()`: set `user.ID = uuid.New().String()`, `user.DocumentID = uuid.New().String()`, `user.CreatedAt = time.Now().UTC()` |
| `apps/api/internal/usecase/auth/auth_usecase_test.go` | Assert returned user has non-empty `ID`, `DocumentID`, and `CreatedAt` |

**Verify:** `cd apps/api && go test ./internal/usecase/auth/...`

---

## B2: Register page re-shown after admin creation

**Root cause:** `/register` has no guard — always accessible. After admin registration, query invalidation re-renders page with `adminExists=true` before navigation.

**Changes:**

| File | Change |
|------|--------|
| `apps/web/src/pages/auth/RegisterPage.tsx` | After `setupLoading` check: if `adminExists`, return `<Navigate to="/login" replace />` |
| `apps/web/src/pages/auth/__tests__/RegisterPage.test.tsx` | Add test: when `adminExists=true`, redirects to `/login` |

**Verify:** `cd apps/web && npx vitest run src/pages/auth`

---

## Checkpoint 1

- `cd apps/api && go test ./internal/usecase/auth/...` — green
- `cd apps/web && npx vitest run src/pages/auth` — green

---

## B4+B5: Response shape restructure (combined)

**Decisions:**
- `status`: Keep in admin responses (top-level alongside `data`), remove from public API + GraphQL
- `contentTypeId`: Remove from ALL responses
- Response shape: Strapi-style — system fields + content fields merged flat inside `data`

### New response shapes

**Admin single document:**
```json
{
  "data": {
    "documentId": "abc", "locale": "en",
    "createdAt": "...", "updatedAt": "...", "createdBy": "...", "updatedBy": "...",
    "title": "Hello"
  },
  "status": "draft"
}
```

**Admin paginated list:**
```json
{
  "items": [
    {
      "data": {
        "documentId": "abc", "locale": "en",
        "createdAt": "...", "updatedAt": "...",
        "title": "Hello"
      },
      "status": "draft"
    }
  ],
  "total": 42, "start": 0, "size": 20
}
```

**Public single document (no status):**
```json
{
  "data": {
    "documentId": "abc", "locale": "en",
    "createdAt": "...", "updatedAt": "...", "createdBy": "...", "updatedBy": "...",
    "title": "Hello"
  }
}
```

### Sub-task B4+B5a: Backend response restructure

| File | Change |
|------|--------|
| `apps/api/internal/delivery/http/handler/document_handler.go` | Replace `entrySummary` with `documentResponse { Data map[string]any; Status string }`. Replace `paginatedListItem` similarly. New `toDocResponse(doc, status)` merges system+content into `Data`. `GetPublic` uses a struct without `Status`. Remove `contentTypeId` from all response paths. |
| `apps/api/internal/delivery/http/handler/document_handler_test.go` | Update all response assertions: access fields via `resp["data"].(map[string]any)["documentId"]` etc. Verify no `contentTypeId`. |
| `apps/api/graphql/dynamic/schema_builder.go` | Remove `status: String!` from generated type (line 71). |
| `apps/api/graphql/dynamic/schema_builder_test.go` | Remove assertions checking for `status: String!`. |
| `apps/api/graphql/dynamic/resolver_factory.go` | Remove `"status"` from `docToMap()`. Remove `"status"` field from `buildObjectType()`. |
| `apps/api/graphql/dynamic/resolver_factory_test.go` | Remove assertions checking `status` in response maps. |
| `apps/api/internal/delivery/grpc/document_service.go` | Remove `ContentTypeId` and `Status` from `toProtoDoc()`. |
| `apps/api/internal/delivery/grpc/document_service_test.go` | Update assertions. |

**Verify:** `cd apps/api && go test ./internal/delivery/... ./graphql/...`

### Sub-task B4+B5b: Frontend adaptation

| File | Change |
|------|--------|
| `apps/web/src/types/cms.ts` | Update `Document` type: top-level has `data: Record<string, unknown>` and optional `status: string`. Remove `documentId`, `contentTypeId`, `locale`, etc. from top level. |
| `apps/web/src/hooks/useSingleTypeDocuments.ts` | Hooks return the response object. Callers access `.data.documentId` instead of `.documentId`. |
| `apps/web/src/hooks/useCollectionDocuments.ts` | Update cache invalidation: `result.data.documentId` instead of `result.documentId`. Same for `result.data.locale`. |
| `apps/web/src/pages/admin/panels/content-type/ContentTypePanel.tsx` | `doc.status` still works (top-level). `doc.documentId` → `doc.data.documentId`. Form queryFn: strip system fields from `doc.data` before passing to form (define `SYSTEM_FIELDS = ['documentId', 'locale', 'createdAt', 'updatedAt', 'createdBy', 'updatedBy']`). |
| `apps/web/src/pages/admin/panels/collection-type/CollectionDetailPage.tsx` | Same pattern — access data via `doc.data.*`. |
| `apps/web/src/pages/admin/panels/collection-type/layout/CollectionListPage.tsx` | Update list item access pattern. |

**Verify:** `cd apps/web && npx vitest run`

---

## Checkpoint 2

- `cd apps/api && go test ./...` — all green
- `cd apps/web && npx vitest run` — all green
- Manual: save a document → response has correct shape
- Manual: paginated list → items have correct shape

---

## B3: JsonInput/RichTextInput data loss on save

**Root cause hypothesis:** After B4+B5, the form receives `doc.data` which now has system fields merged in. If system fields are properly stripped (done in B4+B5b), the issue may be resolved. If not:

**Remaining fix areas:**

| File | Change |
|------|--------|
| `apps/web/src/components/form/inputs/JsonInput.tsx` | Fix value comparison: use `JSON.stringify` deep equality instead of `!==` reference equality for `fieldValue !== prevFieldValue`. Add fallback: if `field.value` is `undefined`, keep previous `rawValue`. |
| `apps/web/src/components/form/inputs/RichTextInput.tsx` | Add fallback: `data={field.value ?? ''}` to prevent CKEditor receiving `undefined`. |
| `apps/web/src/components/form/FormProvider.tsx` | Consider using `resetOptions: { keepDirtyValues: true }` in `useForm({ values })` to prevent value flicker during refetch. |

**Verify:** Manual — save document with JSON + richtext fields → values persist after save. Edit again → re-save → values persist.

---

## Checkpoint 3

- Manual: create document with JSON field, save → value retained
- Manual: create document with richtext field, save → value retained
- Manual: edit and re-save → values retained
- `cd apps/web && npx vitest run` — all green

---

## B6: GraphQL + REST public default to published

**Root cause:** GraphQL resolvers call `GetForEdit`/`GetAllPaginated` (draft-oriented) instead of published-oriented methods.

### Sub-task B6a: New repository + usecase methods

| File | Change |
|------|--------|
| `apps/api/internal/domain/repository/document_repository.go` | Add `FindPublishedByContentTypePaginated(ctx, slug string, start, size int, locale string) ([]*entity.Document, int64, error)` |
| `apps/api/internal/domain/repository/mock/document_repository.go` | Add `FindPublishedByContentTypePaginatedFn` field + delegating method |
| `apps/api/internal/infrastructure/gormdb/document_repository.go` | Implement: copy `FindDraftsByContentTypePaginated`, change `VersionDraft` → `VersionPublished` |
| `apps/api/internal/infrastructure/gormdb/document_repository_test.go` | Test: insert published docs, query paginated, verify correct results |
| `apps/api/internal/usecase/document/document_usecase.go` | Add `GetPublishedPaginated(ctx, slug, start, size, locale, fields)` and `GetPublishedSingleType(ctx, slug, locale, fields)` |
| `apps/api/internal/usecase/document/document_usecase_test.go` | Test both new methods |

**Verify:** `cd apps/api && go test ./internal/...`

### Sub-task B6b: GraphQL resolver changes

| File | Change |
|------|--------|
| `apps/api/graphql/dynamic/resolver_factory.go` | Collection single query: default to `GetPublished()`, add `status` arg with auth check for draft. Collection list query: default to `GetPublishedPaginated()`, add `status` arg. Single-type query: default to `GetPublishedSingleType()`, add `status` arg. |
| `apps/api/graphql/dynamic/resolver_factory_test.go` | Test: unauthenticated query returns published. Test: authenticated with `status: "draft"` returns draft. |

**Verify:** `cd apps/api && go test ./graphql/...`

---

## Checkpoint 4 (Final)

1. `cd apps/api && go test ./...` — all green
2. `cd apps/web && npx vitest run` — all green
3. Manual smoke test:
   - Register → redirected to login
   - Login → create document with json + richtext fields → save → values persist
   - Publish → GraphQL query returns published data (not draft)
   - Unpublish → GraphQL returns null/empty
   - Check response shape matches spec
