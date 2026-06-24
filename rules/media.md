# RULES — media Module

**Scope:** Media asset uploads, deletions, storage provider adapters (S3/Cloudinary), media library API.
**Spec:** [specs/media.md](../specs/media.md)

---

## 1. Entity Rules

### 1.1 MediaAsset Entity
```go
type MediaAsset struct {
    ID           string    // internal PK
    DocumentID   string    // domain identifier (UUID)
    URL          string    // full-size asset URL
    ThumbnailURL string    // thumbnail URL
    PublicID     string    // storage provider's identifier (for deletion)
    FileName     string    // original upload filename
    FileExt      string    // file extension
    Hash         string    // content hash
    Width        int       // image width in pixels
    Height       int       // image height in pixels
    CreatedAt    time.Time
}
```

### 1.2 Removed Fields
- `ContentTypeID` — removed (media not back-referenced to content types)
- `DocumentRef` — removed (documents reference media by `documentId`, not vice versa)

### 1.3 Media References in Documents
- Media fields in documents store the media asset's `document_id` (UUID reference)
- **NOT** URL strings
- GraphQL resolves `document_id` → full `MediaAsset` object via `MediaAssetRepository.FindByDocumentID`

---

## 2. Repository Rules

### 2.1 MediaAssetRepository
```go
Create(ctx, asset) error
FindByID(ctx, id) (*MediaAsset, error)
FindByDocumentID(ctx, documentID) (*MediaAsset, error)
FindAll(ctx) ([]*MediaAsset, error)
Delete(ctx, id) error
```

### 2.2 StorageAdapter Interface
```go
Upload(ctx, file, header) (*UploadResult, error)
Delete(ctx, publicID) error
```
- Both S3 and Cloudinary implement this
- Selection via `STORAGE_PROVIDER` env var (`s3` | `cloudinary`)
- **Ask first** before choosing which adapter is active per environment

---

## 3. Delete Flow (Critical Ordering)

### 3.1 Correct Order
```
1. FindByID(id)           → get asset (propagate not-found as-is)
2. storage.Delete(publicID) → remove from storage provider
3. assetRepo.Delete(id)    → remove DB record
```

### 3.2 Failure Handling
- If `storage.Delete` fails → do **NOT** call `assetRepo.Delete`
- Storage is the source of truth
- Orphaned DB records are harder to clean up than orphaned storage files
- If `assetRepo.Delete` fails after storage success → return error (orphaned DB record)

### 3.3 Invariants
- **Always** call `storage.Delete` **before** `assetRepo.Delete`
- **NEVER** skip storage delete (no DB-only or soft-delete removal)
- **NEVER** bulk-delete — single asset at a time only
- Return 404 (not 500) when asset ID not found

---

## 4. Upload Rules

### 4.1 Upload Flow
1. Receive multipart file
2. Call `storage.Upload(file, header)` → returns `UploadResult` (URL, ThumbnailURL, PublicID)
3. Create `MediaAsset` entity with upload result + metadata
4. Call `assetRepo.Create(asset)`
5. Return created asset

### 4.2 Storage Adapters
- **Cloudinary**: real eager thumbnail generation; returns distinct `thumbnailURL`
- **S3**: no native thumbnail; `thumbnailURL == URL`
- Both implement same `StorageAdapter` interface — no adapter-specific branching in usecase

### 4.3 Upload Constraints
- Media upload is REST-only — **NEVER** expose via gRPC (multipart not supported)
- Media upload exempt from global body size limit (uses own multipart limit)

---

## 5. API Contract Rules

### 5.1 REST Routes
| Method | Route | Permission | Response |
|---|---|---|---|
| `GET` | `/api/media` | `media:read` | `MediaAsset[]` |
| `POST` | `/api/media/upload` | `media:upload` | `MediaAsset` (201) |
| `DELETE` | `/api/media/:id` | `media:delete` | 204 |

### 5.2 Delete Responses
| Status | Condition |
|---|---|
| 204 | Deleted from storage and DB |
| 404 | Asset ID doesn't exist |
| 500 | Storage or DB failure |

### 5.3 gRPC Routes
- `ListMedia` and `DeleteMedia` only
- **NO** upload via gRPC

---

## 6. Frontend Integration Rules

### 6.1 Media Library
- `['media', 'list']` query key for listing
- Invalidate on successful delete
- Confirmation dialog before delete

### 6.2 MediaInput Component
- Renders original + thumbnail preview
- Stores `documentId` reference (not URL) in form data
- Used via `<FormField name="coverImage"><MediaInput /></FormField>`

### 6.3 GraphQL Media Resolution
- Media fields in GraphQL return `MediaAsset` object type:
  ```graphql
  type MediaAsset {
    documentId: ID!
    url: String!
    thumbnailUrl: String
    fileName: String!
    width: Int
    height: Int
  }
  ```
- Resolved recursively in component sub-fields

---

## 7. Testing Rules (Media-Specific)

### 7.1 Usecase Tests
- `TestDelete_CallsStorageAndRepo` — correct order: FindByID → storage.Delete → assetRepo.Delete
- `TestDelete_AssetNotFound_ReturnsError` — storage.Delete never called
- `TestDelete_StorageError_DoesNotDeleteFromRepo` — assetRepo.Delete never called
- `TestDelete_RepoDeleteError_ReturnsError` — error propagated

### 7.2 Handler Tests
- `DELETE /api/media/:id` → 204 on success
- `DELETE /api/media/:id` → 404 for unknown ID
- `DELETE /api/media/:id` → 500 on usecase error

### 7.3 Storage Adapter Tests
- Test with mock/stub external services
- Verify `UploadResult` shape
- Verify `Delete` uses correct `publicID`

---

## 8. Boundaries Summary

| Rule | Detail |
|---|---|
| **Always** | Call `storage.Delete` before `assetRepo.Delete` — storage is source of truth |
| **Always** | Return 404 (not 500) when asset ID not found |
| **Always** | Invalidate `['media', 'list']` query on successful delete (FE) |
| **Always** | Store media references as `document_id` (UUID) in document fields |
| **Never** | Bulk-delete — single asset at a time |
| **Never** | Skip storage delete (no DB-only removal) |
| **Never** | Expose media upload via gRPC |
| **Never** | Store URL strings in document media fields (use documentId) |
| **Ask first** | Choosing which storage adapter is active per environment |
| **Ask first** | Cascade-deleting assets referenced by documents |
