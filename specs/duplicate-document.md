# SPEC — Duplicate Document (Collection List Row Action)

## 1. Overview

Add a "Duplicate" button to each row in the Collection List page. Clicking it creates a full copy of the source document (all fields, components, and data) as a new draft entry with a fresh `documentId` and auto-generated `gormId`. After duplication, the user is navigated to the new document's edit page.

**Target users:** CMS admins who want to quickly create a new entry based on an existing one (e.g., cloning a blog post template, duplicating a product with similar fields).

---

## 2. Decisions

| Question | Decision |
|----------|----------|
| Publish state | Always create as **draft only** — no published version, regardless of the original's status |
| Media assets | **Share references** — duplicate keeps the same media asset document IDs (no file re-upload) |
| Post-action | **Navigate to new document** — open the duplicated document's edit page immediately |
| Locales | **Active locale only** — only duplicate the document in the currently active locale |
| Components | **Fully duplicated** — all component data is copied; each component gets a new `componentId` |

---

## 3. Excluded from Scope

| Item | Reason |
|------|--------|
| `documentId` | Auto-generated (UUID v4) by the backend `Save` method |
| `gormId` | Auto-incremented by the database |
| `createdAt` / `updatedAt` | Set to current time by the backend `Save` method |
| `createdBy` / `updatedBy` | Set to the current authenticated user by the backend |
| `publishedAt` / `publishedBy` | Not set (new document is always draft) |
| Published version record | Not created — only a draft record is created |
| Other locales | Not duplicated — only the active locale is copied |

---

## 4. Changes — Backend (Go API)

### 4.1 Document UseCase — New Method

**File:** `apps/api/internal/usecase/document/document_usecase.go`

Add a `Duplicate` method:

```go
func (uc *UseCase) Duplicate(ctx context.Context, contentTypeSlug, sourceDocumentID, locale string, fields []entity.FieldDefinition, userID string) (*entity.Document, error)
```

**Logic:**
1. Resolve locale (same as other methods)
2. Fetch the source draft via `repo.FindDraftByDocumentID(ctx, contentTypeSlug, sourceDocumentID, locale)`
3. Merge components into the source document via `mergeComponents(ctx, contentTypeSlug, source, fields)` — this loads component data into `source.Fields`
4. Create a new `entity.Document` with:
   - `DocumentID`: empty string (the existing `Save` method generates a new UUID)
   - `Fields`: deep copy of `source.Fields`
   - `Locale`: same as source
5. Call `uc.Save(ctx, contentTypeSlug, newDoc, fields, userID)` — this handles:
   - Generating a new `documentId` (UUID v4)
   - Setting `createdAt`, `updatedAt`, `createdBy`, `updatedBy`
   - Extracting and saving components with new `componentId` values
   - Upserting the draft record
6. Return the saved document

**Deep copy of Fields:** Use a JSON marshal/unmarshal round-trip (`encoding/json`) to produce a true deep copy of `map[string]any`. This ensures nested maps and slices are independent of the source.

### 4.2 Document Handler — New Endpoint

**File:** `apps/api/internal/delivery/http/handler/document_handler.go`

Add `DuplicateCollection` handler:

```go
func (h *DocumentHandler) DuplicateCollection(c *gin.Context) {
    slug := c.Param("slug")
    documentID := c.Param("documentId")
    fields := h.resolveFields(c, slug)
    userID := middleware.UserID(c.Request.Context())

    saved, err := h.uc.Duplicate(ctx, slug, documentID, c.Query("locale"), fields, userID)
    if err != nil { ginWriteErr(c, err); return }

    c.JSON(http.StatusCreated, toDocResponse(saved, "draft"))
}
```

Add `Duplicate` to the `documentUseCase` interface:

```go
Duplicate(ctx context.Context, contentTypeSlug, sourceDocumentID, locale string, fields []entity.FieldDefinition, userID string) (*entity.Document, error)
```

### 4.3 Router — New Route

**File:** `apps/api/internal/delivery/http/router.go`

Add inside the `colGroup` block:

```go
colGroup.POST("/:slug/:documentId/duplicate", middleware.GinRequirePermission(cache, "content:create"), cfg.DocHandler.DuplicateCollection)
```

Permission: `content:create` (duplicating creates a new entry).

### 4.4 API Contract

| Method | Route | Permission | Response |
|--------|-------|------------|----------|
| `POST` | `/api/document-manager/collection-type/:slug/:documentId/duplicate` | `content:create` | `Document` (201) |

**Request body:** None (all data is read from the source document).

**Response:** Same shape as `CreateCollection` — the new document wrapped in `{ data: { ... }, status: "draft" }`.

**Error cases:**
- Source document not found → 404
- Invalid slug or documentId → 400
- Unauthorized → 401/403

---

## 5. Changes — Frontend (React)

### 5.1 New Hook — `useDuplicateCollectionDocument`

**File:** `apps/web/src/hooks/useCollectionDocuments.ts`

Add a new mutation hook:

```typescript
export function useDuplicateCollectionDocument() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({
      contentTypeSlug,
      id,
      locale,
    }: {
      contentTypeSlug: string
      id: string
      locale?: string
    }) =>
      api
        .post<Document>(
          `/api/document-manager/collection-type/${contentTypeSlug}/${id}/duplicate`,
          undefined,
          { params: { locale } },
        )
        .then((r) => r.data),
    onSuccess: (_, { contentTypeSlug }) =>
      qc.invalidateQueries({ queryKey: KEYS.list(contentTypeSlug) }),
    onError: onMutationError,
  })
}
```

### 5.2 Collection List Page — Duplicate Button

**File:** `apps/web/src/pages/admin/panels/collection-type/layout/CollectionListPage.tsx`

**Changes:**
1. Import `Copy` icon from `lucide-react` (alongside existing `Pencil`, `Trash2`)
2. Import `useDuplicateCollectionDocument` hook
3. Add `handleDuplicate` function:
   - Calls `duplicateDoc` mutation with `contentTypeSlug`, `documentId`, and `locale`
   - On success: navigate to `/admin/content-type/collection-type/:slug/:newDocumentId`
4. Add a new icon button in the Actions column between Edit and Delete:

```tsx
<Button variant="ghost" size="icon" aria-label="Duplicate" onClick={(e) => handleDuplicate(e, doc)}>
  <Copy className="h-4 w-4" />
</Button>
```

**Button order in Actions column:** Edit | Duplicate | Delete

---

## 6. Testing

### 6.1 Backend — Usecase Test

**File:** `apps/api/internal/usecase/document/document_usecase_test.go`

| Test Case | Description |
|-----------|-------------|
| `TestDuplicate_Success` | Source exists → new doc created with new documentId, same field data, draft version |
| `TestDuplicate_CopiesComponents` | Source has component fields → components are duplicated with new componentIds |
| `TestDuplicate_SharesMediaRefs` | Source has media field references → duplicate keeps same media documentIds |
| `TestDuplicate_SourceNotFound` | Source documentId doesn't exist → returns `ErrNotFound` |
| `TestDuplicate_InvalidLocale` | Unsupported locale → returns `ErrValidation` |
| `TestDuplicate_NeverPublished` | Source is published → duplicate is still draft-only, no published record created |

### 6.2 Backend — Handler Test

**File:** `apps/api/internal/delivery/http/handler/document_handler_test.go`

| Test Case | Description |
|-----------|-------------|
| `TestDuplicateCollection_201` | Valid request → 201 with new document response |
| `TestDuplicateCollection_404` | Source not found → 404 |
| `TestDuplicateCollection_Permission` | Missing `content:create` → 403 |

### 6.3 Frontend — Component Test

**File:** `apps/web/src/pages/admin/panels/collection-type/layout/__tests__/CollectionListPage.test.tsx`

| Test Case | Description |
|-----------|-------------|
| Duplicate button renders | Each row has a button with aria-label "Duplicate" |
| Duplicate navigates to new doc | After successful mutation, `navigate` called with new documentId |

---

## 7. Boundaries

| Rule | Detail |
|------|--------|
| **Always** | Duplicate creates a draft-only document — never a published version |
| **Always** | New `documentId` (UUID v4) and `gormId` (auto-increment) are system-generated |
| **Always** | `createdAt`, `updatedAt`, `createdBy`, `updatedBy` reflect the duplicating user and current time |
| **Always** | Component data is fully copied; each component gets a new `componentId` |
| **Always** | Media asset references are shared (same document IDs) — no file re-upload |
| **Always** | After duplication, navigate to the new document's edit page |
| **Always** | Require `content:create` permission for the duplicate endpoint |
| **Never** | Copy `publishedAt`, `publishedBy`, or the published version record |
| **Never** | Duplicate across multiple locales in a single request |
| **Never** | Deep-clone media files (share references only) |

---

## 8. Changelog

| Version | Change |
|---------|--------|
| v1.0 | Initial spec for Duplicate Document feature |
