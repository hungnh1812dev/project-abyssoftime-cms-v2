# SPEC — media Module

## 1. Overview

The media module manages media asset uploads, deletions, and storage provider adapters. It supports both AWS S3 and Cloudinary as storage backends, selectable via configuration. The module provides a media library API for listing assets and a delete flow that removes files from both storage and database.

---

## 2. File Map

All paths relative to `apps/api/`.

```
internal/domain/entity/media_asset.go                        # MediaAsset entity
internal/domain/repository/media_asset_repository.go         # MediaAssetRepository interface
internal/domain/repository/storage_adapter.go                # StorageAdapter interface
internal/domain/repository/mock/media_asset_repository.go    # Mock for testing
internal/domain/repository/mock/storage_adapter.go           # Mock for testing
internal/usecase/media/media_usecase.go                      # Media business logic
internal/usecase/media/media_usecase_test.go
internal/delivery/http/handler/media_handler.go              # Gin media handlers
internal/delivery/http/handler/media_handler_test.go
internal/delivery/grpc/media_service.go                      # gRPC MediaService
internal/delivery/grpc/media_service_test.go
internal/infrastructure/mongodb/media_asset_repository.go    # MongoDB MediaAsset repo
internal/infrastructure/mongodb/media_asset_repository_test.go
internal/infrastructure/gormdb/media_asset_repository.go     # GORM MediaAsset repo
internal/infrastructure/gormdb/media_asset_repository_test.go
internal/infrastructure/cloudinary/cloudinary_adapter.go     # Cloudinary storage adapter
internal/infrastructure/cloudinary/cloudinary_adapter_test.go
internal/infrastructure/s3/s3_adapter.go                     # S3 storage adapter
internal/infrastructure/s3/s3_adapter_test.go
proto/cms/v1/media.proto                                     # gRPC Media proto
proto/cms/v1/media.pb.go                                     # Generated
proto/cms/v1/media_grpc.pb.go                                # Generated
```

---

## 3. Entities

### MediaAsset

```go
type MediaAsset struct {
    ID            string    `bson:"_id,omitempty"  gorm:"column:id;primaryKey"        json:"ID"`
    DocumentID    string    `bson:"documentId"     gorm:"column:document_id"          json:"documentId"`
    URL           string    `bson:"url"            gorm:"column:url"                  json:"url"`
    ThumbnailURL  string    `bson:"thumbnailUrl"   gorm:"column:thumbnail_url"        json:"thumbnailUrl"`
    PublicID      string    `bson:"publicId"       gorm:"column:public_id"            json:"publicId"`
    FileName      string    `bson:"fileName"       gorm:"column:file_name"            json:"fileName"`
    FileExt       string    `bson:"fileExt"        gorm:"column:file_ext"             json:"fileExt"`
    Hash          string    `bson:"hash"           gorm:"column:hash"                 json:"hash"`
    Width         int       `bson:"width"          gorm:"column:width"                json:"width"`
    Height        int       `bson:"height"         gorm:"column:height"               json:"height"`
    CreatedAt     time.Time `bson:"createdAt"      gorm:"column:created_at"           json:"createdAt"`
}
```

**Removed fields:** `ContentTypeID` and `DocumentRef` (media assets are now referenced by their `document_id` from document media fields, not by explicit back-references).

---

## 4. Repository Interfaces

### MediaAssetRepository

```go
type MediaAssetRepository interface {
    Create(ctx context.Context, asset *entity.MediaAsset) error
    FindByID(ctx context.Context, id string) (*entity.MediaAsset, error)
    FindByDocumentID(ctx context.Context, documentID string) (*entity.MediaAsset, error)
    FindAll(ctx context.Context) ([]*entity.MediaAsset, error)
    Delete(ctx context.Context, id string) error
}
```

**Changes:** Removed `DeleteByDocumentRef`; added `FindByDocumentID` (used by GraphQL resolvers to resolve media fields into full `MediaAsset` objects).

### StorageAdapter

```go
type StorageAdapter interface {
    Upload(ctx context.Context, file multipart.File, header *multipart.FileHeader) (*UploadResult, error)
    Delete(ctx context.Context, publicID string) error
}

type UploadResult struct {
    URL          string
    ThumbnailURL string
    PublicID     string
}
```

---

## 5. Use Cases

### Media UseCase (`usecase/media/`)

| Method | Signature | Description |
|---|---|---|
| `Upload` | `(ctx, file, header) → (*MediaAsset, err)` | Upload to storage, create DB record |
| `List` | `(ctx) → ([]*MediaAsset, err)` | List all media assets |
| `Delete` | `(ctx, id) → err` | Delete from storage then DB |

### Delete Flow

```
Delete(ctx, id):
  1. asset ← assetRepo.FindByID(ctx, id)       // propagate not-found as-is
  2. storage.Delete(ctx, asset.PublicID)        // remove from storage provider
  3. assetRepo.Delete(ctx, id)                 // remove DB record
  return error
```

If `storage.Delete` fails, do **not** call `assetRepo.Delete` — storage is the source of truth. Orphaned DB records are harder to clean up than orphaned storage files.

---

## 6. API Contracts

### REST — Media Routes

| Method | Route | Permission | Response | Description |
|---|---|---|---|---|
| `GET` | `/api/media` | `media:read` | `MediaAsset[]` | List all assets |
| `POST` | `/api/media/upload` | `media:upload` | `MediaAsset` (201) | Upload file |
| `DELETE` | `/api/media/:id` | `media:delete` | `204` | Delete asset |

**Delete responses:**

| Status | Body | Condition |
|---|---|---|
| `204 No Content` | — | Deleted from storage and DB |
| `404 Not Found` | `{"error": "not found"}` | Asset ID doesn't exist |
| `500 Internal Server Error` | `{"error": "..."}` | Storage or DB failure |

### gRPC — MediaService

```protobuf
service MediaService {
    rpc ListMedia(ListMediaRequest) returns (ListMediaResponse);
    rpc DeleteMedia(DeleteMediaRequest) returns (DeleteMediaResponse);
}
```

Media upload is REST-only (multipart upload not supported via gRPC).

---

## 7. Infrastructure

### Storage Adapter Selection

Configured via `STORAGE_PROVIDER` env var:
- `s3` → S3 adapter (`internal/infrastructure/s3/`)
- `cloudinary` → Cloudinary adapter (`internal/infrastructure/cloudinary/`)

Both implement the `StorageAdapter` interface. Selection happens in `main.go`:

```go
func newStorageAdapter(ctx context.Context, cfg *config.Config) (repository.StorageAdapter, error) {
    switch cfg.Media.Driver {
    case "s3":
        return s3adapter.New(ctx, cfg.Media.S3.Bucket, cfg.Media.S3.Region)
    default:
        return cloudinaryadapter.NewCloudinaryAdapter(cloudName, apiKey, apiSecret)
    }
}
```

### Cloudinary Adapter

- Upload: sends file to Cloudinary API, returns URL + thumbnail URL + public ID
- Delete: removes file by public ID via Cloudinary Admin API

### S3 Adapter

- Upload: puts object to S3 bucket, returns URL + public ID
- Delete: removes object from S3 bucket by key

---

## 8. Testing

**UseCase (`media_usecase_test.go`):**
- `TestDelete_CallsStorageAndRepo` — FindByID → storage.Delete with correct PublicID → assetRepo.Delete
- `TestDelete_AssetNotFound_ReturnsError` — FindByID returns not-found; storage.Delete never called
- `TestDelete_StorageError_DoesNotDeleteFromRepo` — storage.Delete fails; assetRepo.Delete never called
- `TestDelete_RepoDeleteError_ReturnsError` — storage succeeds, assetRepo.Delete fails; error propagated

**Handler (`media_handler_test.go`):**
- `DELETE /api/media/:id` → 204 on success
- `DELETE /api/media/:id` → 404 for unknown ID
- `DELETE /api/media/:id` → 500 on usecase error

---

## 9. Boundaries

| Rule | Detail |
|---|---|
| **Always** | Call `storage.Delete` before `assetRepo.Delete` — storage is source of truth |
| **Always** | Return 404 (not 500) when asset ID is not found |
| **Always** | Invalidate `['media', 'list']` query on successful delete (FE) |
| **Never** | Bulk-delete — single asset at a time only |
| **Never** | Skip storage delete (no DB-only or soft-delete removal) |
| **Never** | Expose media upload via gRPC — multipart upload is REST-only |
| **Ask first** | Choosing which storage adapter is active per environment |
| **Ask first** | Cascade-deleting assets referenced by documents (touches `DeleteByDocumentRef` — cross-module) |

---

## 10. Resolved Decisions

1. **Media storage**: Support both AWS S3 and Cloudinary from day one, behind the `StorageAdapter` interface, selectable via `STORAGE_PROVIDER` env var.

---

## 11. Changelog

| Date | Change | Source |
|------|--------|--------|
| v1.0 | Media upload + list (Cloudinary) | §1 |
| v1.1 | S3 adapter added behind StorageAdapter interface | Resolved Decision #1 |
| v1.2 | Delete media asset with inline-confirm UX | §8 |
| v1.3 | gRPC MediaService (list + delete) | §11.7 |
| v1.4 | Removed `ContentTypeID` and `DocumentRef` fields from MediaAsset entity | sync-table-fields |
| v1.5 | Added `FindByDocumentID` method to MediaAssetRepository; added `Width`/`Height` fields | sync-table-fields, graphql-overhaul |
| v1.6 | Media fields in documents now store the media asset's `document_id` (UUID reference), not URL strings | sync-table-fields |
