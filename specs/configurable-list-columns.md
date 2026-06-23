# SPEC — Configurable List Columns (CollectionListPage)

## 1. Overview

Allow users to choose which columns are visible on the CollectionListPage table via a popup dialog, instead of manually editing `listFields` in content-type JSON schema files. The selection is persisted globally (per content type) in the `list_fields` column of the `content_types` DB table and affects all users.

**Target users:** CMS admins who want to customize which fields appear in collection list views without touching JSON config files.

---

## 2. Decisions

| Question | Decision |
|----------|----------|
| Column source | Content-type `Fields` definition (user-defined schema fields), not physical DB columns |
| Locked columns | **Id** (first), **Status** (before Actions), **Actions** (last) — always visible, not in the popup |
| Scope | **Global** — saved on the content-type record, affects all users |
| Sync behavior | **Seed once, never overwrite** — startup sync only sets `ListFields` when the DB value is empty/nil; once a user configures it, sync leaves it alone |
| Default (empty listFields) | First 3 content fields + CreatedAt + UpdatedAt + UpdatedBy (same as current) |
| Content-type registry | Registry overrides still take precedence; column chooser button hidden when a registry entry defines columns |

---

## 3. Current Behavior

1. `listFields` is **never** defined in content-type JSON schema files — it has been permanently removed from the schema format
2. On startup, `sync.go` loads JSON definitions and seeds `list_fields` in the DB **only when the DB value is empty/nil** — it never overwrites an existing user-configured value
3. Backend handler `ListCollection` reads `ct.ListFields`, falls back to first 3 content fields if empty
4. Frontend `CollectionListPage` renders a fixed layout:

```
| Id | [dynamic content fields from listFields] | Status | CreatedAt | UpdatedAt | UpdatedBy | Actions |
```

System columns (CreatedAt, UpdatedAt, UpdatedBy) are always rendered regardless of `listFields`.

---

## 4. Target Behavior

### Column layout after this feature

```
| Id | [selected content fields in Fields order] | [selected system fields] | Status | Actions |
```

- **Content fields** — only the selected ones from the popup; order follows the `Fields` definition order
- **System fields** — CreatedAt, UpdatedAt, UpdatedBy; only rendered if selected (or if listFields is empty = defaults)
- **Status** — always shown, always positioned immediately before Actions
- **Actions** — always shown, always last

### Popup behavior

- Accessible via a gear/settings icon button in the page header
- Two sections: "Content fields" and "System fields"
- Content fields: all fields from the content-type `Fields` definition, excluding `component` type fields
- System fields: CreatedAt, UpdatedAt, UpdatedBy
- Each field has a checkbox; checked = visible in the table
- Save persists to DB via new PATCH endpoint; Cancel discards
- Empty selection = revert to defaults

---

## 5. Changes — Backend (Go API)

### 5.1 `listFields` removed from JSON schemas

**Files:** `apps/api/content-types/*.json`

The `"listFields"` key no longer exists in any content-type JSON schema file. It has been permanently removed from the schema format. `listFields` is now exclusively managed via the UI and persisted in the DB.

If any JSON file still contains a `listFields` key, the schema loader ignores it (backward-compatible, no error).

### 5.2 Sync: seed-only, never overwrite ListFields

**File:** `apps/api/internal/usecase/content_type/sync.go`

Since `listFields` is no longer in JSON schemas, sync must stop treating it as a synced field. The DB value is the sole source of truth — sync only touches it when creating a brand-new content type (where it starts as `nil`).

In `syncOne()`:

1. **New content type (line 74):** Set `ListFields` to `nil` (no JSON source):
```go
ct := &entity.ContentType{Name: def.Name, Slug: def.Slug, Kind: kind, Fields: def.Fields}
```

2. **Change detection (line 91):** Remove `stringSliceEqual(current.ListFields, def.ListFields)` from the comparison — `ListFields` changes should never trigger a sync update:
```go
if current.Name == def.Name && current.Kind == kind && fieldsEqual(current.Fields, def.Fields) {
    return nil
}
```

3. **Update (line 97):** Never overwrite `ListFields` — preserve whatever the user configured:
```go
current.Name = def.Name
current.Kind = kind
current.Fields = def.Fields
// current.ListFields is NOT touched — it's user-managed via the UI
```

### 5.3 Schema loader: remove listFields validation

**File:** `apps/api/internal/usecase/content_type/schema_loader.go`

- Keep `ListFields` field in `ContentTypeDefinition` struct for backward compatibility (won't fail if someone still has it in JSON)
- Remove the `validateListFields()` call from `validateDefinition()` — listFields validation moves to the new PATCH handler
- Delete the `validateListFields()` function

```go
func validateDefinition(def ContentTypeDefinition, path string) error {
    return validateFields(def.Fields, path, 1)
}
```

### 5.4 New handler: UpdateListFields

**File:** `apps/api/internal/delivery/http/handler/content_type_handler.go`

Expand the `contentTypeUseCase` interface:
```go
type contentTypeUseCase interface {
    FindByID(ctx context.Context, id string) (*entity.ContentType, error)
    FindBySlug(ctx context.Context, slug string) (*entity.ContentType, error)
    FindAll(ctx context.Context) ([]*entity.ContentType, error)
    Update(ctx context.Context, ct *entity.ContentType) error
}
```

Add handler method:
```go
type updateListFieldsRequest struct {
    ListFields []string `json:"listFields"`
}

var knownSystemFields = map[string]bool{
    "createdAt":     true,
    "updatedAt":     true,
    "updatedByName": true,
}

func (h *ContentTypeHandler) UpdateListFields(ginCtx *gin.Context) {
    slug := ginCtx.Param("slug")

    var req updateListFieldsRequest
    if err := ginCtx.ShouldBindJSON(&req); err != nil {
        ginWriteError(ginCtx, http.StatusBadRequest, "invalid request body")
        return
    }

    contentType, err := h.usecase.FindBySlug(ginCtx.Request.Context(), slug)
    if err != nil {
        ginWriteErr(ginCtx, err)
        return
    }

    // Validate: each entry must be a known content field or system field
    fieldNames := make(map[string]bool, len(contentType.Fields))
    for _, field := range contentType.Fields {
        if field.Type != "component" {
            fieldNames[field.Name] = true
        }
    }
    for _, entry := range req.ListFields {
        if !fieldNames[entry] && !knownSystemFields[entry] {
            ginWriteError(ginCtx, http.StatusBadRequest, "invalid field: "+entry)
            return
        }
    }

    contentType.ListFields = req.ListFields
    if err := h.usecase.Update(ginCtx.Request.Context(), contentType); err != nil {
        ginWriteErr(ginCtx, err)
        return
    }

    ginCtx.JSON(http.StatusOK, gin.H{"listFields": contentType.ListFields})
}
```

### 5.5 Register the new route

**File:** `apps/api/internal/delivery/http/router.go`

Add to the `ctGroup`:
```go
ctGroup.PATCH("/:slug/list-fields", cfg.CTHandler.UpdateListFields)
```

This reuses the existing `content_types:read` permission on the group. If stricter gating is needed later, swap to a dedicated permission.

### 5.6 Update ListCollection response

**File:** `apps/api/internal/delivery/http/handler/document_handler.go`

In `ListCollection()`, after the existing fallback logic (lines 261-271):

1. Separate `listFields` into content fields and system field flags:
```go
contentFields := []string{}
showCreatedAt := false
showUpdatedAt := false
showUpdatedByName := false

if len(listFields) == 0 {
    // Default: first 3 content fields + all system fields
    // (existing fallback logic for content fields stays)
    showCreatedAt = true
    showUpdatedAt = true
    showUpdatedByName = true
} else {
    for _, field := range listFields {
        switch field {
        case "createdAt":
            showCreatedAt = true
        case "updatedAt":
            showUpdatedAt = true
        case "updatedByName":
            showUpdatedByName = true
        default:
            contentFields = append(contentFields, field)
        }
    }
}
```

2. Use `contentFields` for `projectData()` and conditionally include system fields in `mergeListItemData`:
- Only include `createdAt` if `showCreatedAt`
- Only include `updatedAt` if `showUpdatedAt`
- Only include `updatedByName` if `showUpdatedByName`

3. Add `listFields` to the paginated response so the frontend knows the saved config:
```go
ginCtx.JSON(http.StatusOK, gin.H{
    "items":      items,
    "total":      total,
    "start":      start,
    "size":       size,
    "listFields": ct.ListFields,
})
```

---

## 6. Changes — Frontend (React)

### 6.1 New hook: useUpdateListFields

**File:** `apps/web/src/hooks/useContentTypes.ts`

```typescript
export function useUpdateListFields() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ slug, listFields }: { slug: string; listFields: string[] }) =>
      api.patch<{ listFields: string[] }>(`/api/content-types/${slug}/list-fields`, { listFields }).then((res) => res.data),
    onSuccess: (_, { slug }) => {
      queryClient.invalidateQueries({ queryKey: KEYS.bySlug(slug) });
      queryClient.invalidateQueries({ queryKey: KEYS.all });
    },
  });
}
```

### 6.2 New component: ColumnChooserDialog

**File:** `apps/web/src/components/collection/ColumnChooserDialog.tsx`

Uses Shadcn UI `Dialog`, `Checkbox`, `Button`, `Label` components.

**Props:**
```typescript
interface ColumnChooserDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  contentType: ContentType;
  currentListFields: string[];
  onSave: (selectedFields: string[]) => void;
  isSaving: boolean;
}
```

**Behavior:**
- Renders two sections: "Content fields" and "System fields"
- Content fields: all from `contentType.Fields` where `type !== "component"`, listed in definition order
- System fields: `createdAt` (label: "Created At"), `updatedAt` (label: "Updated At"), `updatedByName` (label: "Updated By")
- Checkboxes initialized from `currentListFields`; if `currentListFields` is empty, default selection = first 3 content fields + all system fields
- Save button calls `onSave(selectedFields)` where `selectedFields` is the array of checked field names (content fields first in definition order, then system fields)
- Cancel closes the dialog without saving

### 6.3 Update CollectionListPage

**File:** `apps/web/src/pages/admin/panels/collection-type/layout/CollectionListPage.tsx`

1. **Add gear icon button** in the header area, next to "Add new item":
```tsx
<div className="flex items-center justify-between">
  <h1 className="text-xl font-semibold">{contentType.Name}</h1>
  <div className="flex items-center gap-2">
    {!getRegistration(contentType.Slug)?.columns && (
      <Button variant="outline" size="icon" onClick={() => setColumnChooserOpen(true)}>
        <Settings2 className="h-4 w-4" />
      </Button>
    )}
    <Button onClick={handleCreate}>Add new item</Button>
  </div>
</div>
```

The gear button is hidden when a content-type registry override exists (registry takes full control).

2. **Render ColumnChooserDialog:**
```tsx
<ColumnChooserDialog
  open={columnChooserOpen}
  onOpenChange={setColumnChooserOpen}
  contentType={contentType}
  currentListFields={contentType.listFields ?? []}
  onSave={handleSaveListFields}
  isSaving={updateListFields.isPending}
/>
```

3. **Update `deriveColumns()`** to separate content columns from system columns:
- Content columns: derived from `listFields` entries that match a content field
- System columns: derived from `listFields` entries that match system field keys
- If `listFields` is empty, use defaults (first 3 content fields + all system fields)

4. **Update table rendering** to conditionally show system columns:
- CreatedAt `<TableHead>` and `<TableCell>` only if selected
- UpdatedAt only if selected
- UpdatedBy only if selected

5. **Add `useUpdateListFields` hook** for the save handler:
```typescript
const updateListFields = useUpdateListFields();

function handleSaveListFields(selectedFields: string[]) {
  updateListFields.mutate(
    { slug: contentType.Slug, listFields: selectedFields },
    { onSuccess: () => setColumnChooserOpen(false) },
  );
}
```

---

## 7. Excluded from Scope

| Item | Reason |
|------|--------|
| Drag-and-drop column reordering | Keep it simple; columns follow Fields definition order |
| Custom column labels/aliases | Column labels = field names (current behavior) |
| Column width persistence | Not part of this feature |
| Per-user preferences | Decided: global per content type |
| Sort preferences in listFields | Sort state is ephemeral (session only) |

---

## 8. Testing Strategy

### Backend unit tests

| Test | File |
|------|------|
| Sync does NOT overwrite user-configured ListFields | `sync_test.go` |
| Sync creates new content type with nil ListFields | `sync_test.go` |
| UpdateListFields: valid content fields | `content_type_handler_test.go` |
| UpdateListFields: valid system fields | `content_type_handler_test.go` |
| UpdateListFields: rejects invalid field name | `content_type_handler_test.go` |
| UpdateListFields: rejects component-type fields | `content_type_handler_test.go` |
| UpdateListFields: empty array (reset to defaults) | `content_type_handler_test.go` |
| ListCollection: system fields conditionally included | `document_handler_test.go` |
| Schema loader: no error when listFields absent from JSON | `schema_loader_test.go` |

### Frontend unit tests

| Test | File |
|------|------|
| ColumnChooserDialog renders content fields (excludes components) | `ColumnChooserDialog.test.tsx` |
| ColumnChooserDialog renders system fields section | `ColumnChooserDialog.test.tsx` |
| ColumnChooserDialog initializes checkboxes from currentListFields | `ColumnChooserDialog.test.tsx` |
| ColumnChooserDialog defaults when currentListFields is empty | `ColumnChooserDialog.test.tsx` |
| CollectionListPage shows gear icon (no registry override) | `CollectionListPage.test.tsx` |
| CollectionListPage hides gear icon (registry override exists) | `CollectionListPage.test.tsx` |
| CollectionListPage conditionally renders system columns | `CollectionListPage.test.tsx` |

### Manual verification

1. Open a collection list page — see default columns (first 3 fields + system columns)
2. Click gear icon — dialog opens with all content fields and system fields
3. Uncheck some content fields and a system field — save
4. Table updates to show only selected columns; Status and Actions still visible
5. Refresh page — columns persist
6. Restart API server — columns still persist (sync did not overwrite)
7. Save empty selection — reverts to default behavior (first 3 + all system)

---

## 9. Boundaries

### Always
- Preserve backward compatibility: if `listFields` is null/empty, show defaults
- Validate field names against the content-type's Fields definition before saving
- Keep the content-type registry override working for hardcoded column layouts
- Status column always appears before Actions; Id always first

### Never
- Allow removing locked columns (Id, Status, Actions) via the popup
- Store column width or sort preferences in `listFields`
- Make this per-user — it's a global content-type setting
- Let the startup sync overwrite user-configured `listFields` — sync only seeds when empty, never overwrites
- Define `listFields` in content-type JSON schema files — it is permanently removed from the schema format

### Ask First
- Adding drag-and-drop column reordering
- Supporting custom column labels or aliases
- Changing the permission from `content_types:read` to something stricter

---

## 10. File Change Summary

| File | Change |
|------|--------|
| `apps/api/content-types/*.json` | Remove `listFields` key |
| `apps/api/internal/usecase/content_type/sync.go` | Stop overwriting ListFields from JSON |
| `apps/api/internal/usecase/content_type/schema_loader.go` | Remove `validateListFields` call/function |
| `apps/api/internal/delivery/http/handler/content_type_handler.go` | Add `UpdateListFields` handler, expand interface |
| `apps/api/internal/delivery/http/handler/document_handler.go` | Conditional system fields in ListCollection |
| `apps/api/internal/delivery/http/router.go` | Register `PATCH /:slug/list-fields` route |
| `apps/web/src/hooks/useContentTypes.ts` | Add `useUpdateListFields` mutation hook |
| `apps/web/src/components/collection/ColumnChooserDialog.tsx` | **New** — column chooser dialog component |
| `apps/web/src/pages/admin/panels/collection-type/layout/CollectionListPage.tsx` | Gear button, dynamic system columns, dialog integration |
| `apps/web/src/content-type-registry/index.ts` | No changes (registry override preserved) |
| `apps/web/src/types/cms.ts` | No changes (`listFields?: string[]` already supports this) |
