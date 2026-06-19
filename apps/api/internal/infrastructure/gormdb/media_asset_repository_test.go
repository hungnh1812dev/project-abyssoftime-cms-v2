package gormdb

import (
	"context"
	"testing"
	"time"

	"gorm.io/gorm"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

func setupMediaDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := NewClient("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if err := db.AutoMigrate(&entity.MediaAsset{}); err != nil {
		t.Fatalf("AutoMigrate: %v", err)
	}
	return db
}

func TestMediaAssetRepository_Create_And_FindByID(t *testing.T) {
	db := setupMediaDB(t)
	repo := NewMediaAssetRepository(db)
	ctx := context.Background()

	asset := &entity.MediaAsset{
		ID:        "m1",
		URL:       "https://cdn/photo.jpg",
		FileName:  "photo.jpg",
		CreatedAt: time.Now().UTC(),
	}
	if err := repo.Create(ctx, asset); err != nil {
		t.Fatalf("Create: %v", err)
	}

	found, err := repo.FindByID(ctx, "m1")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.URL != "https://cdn/photo.jpg" {
		t.Errorf("URL = %q, want %q", found.URL, "https://cdn/photo.jpg")
	}
}

func TestMediaAssetRepository_FindByID_NotFound(t *testing.T) {
	db := setupMediaDB(t)
	repo := NewMediaAssetRepository(db)

	_, err := repo.FindByID(context.Background(), "nope")
	if !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestMediaAssetRepository_FindByDocumentRef(t *testing.T) {
	db := setupMediaDB(t)
	repo := NewMediaAssetRepository(db)
	ctx := context.Background()

	_ = repo.Create(ctx, &entity.MediaAsset{ID: "m1", DocumentRef: "doc1", CreatedAt: time.Now().UTC()})
	_ = repo.Create(ctx, &entity.MediaAsset{ID: "m2", DocumentRef: "doc1", CreatedAt: time.Now().UTC()})
	_ = repo.Create(ctx, &entity.MediaAsset{ID: "m3", DocumentRef: "doc2", CreatedAt: time.Now().UTC()})

	found, err := repo.FindByDocumentRef(ctx, "doc1")
	if err != nil {
		t.Fatalf("FindByDocumentRef: %v", err)
	}
	if len(found) != 2 {
		t.Errorf("count = %d, want 2", len(found))
	}
}

func TestMediaAssetRepository_FindAll_Paginated(t *testing.T) {
	db := setupMediaDB(t)
	repo := NewMediaAssetRepository(db)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		_ = repo.Create(ctx, &entity.MediaAsset{
			ID:        "m" + string(rune('0'+i)),
			CreatedAt: time.Now().UTC().Add(time.Duration(i) * time.Second),
		})
	}

	items, total, err := repo.FindAll(ctx, 1, 2)
	if err != nil {
		t.Fatalf("FindAll: %v", err)
	}
	if total != 5 {
		t.Errorf("total = %d, want 5", total)
	}
	if len(items) != 2 {
		t.Errorf("page size = %d, want 2", len(items))
	}
}

func TestMediaAssetRepository_DeleteByDocumentRef(t *testing.T) {
	db := setupMediaDB(t)
	repo := NewMediaAssetRepository(db)
	ctx := context.Background()

	_ = repo.Create(ctx, &entity.MediaAsset{ID: "m1", DocumentRef: "doc1", CreatedAt: time.Now().UTC()})
	_ = repo.Create(ctx, &entity.MediaAsset{ID: "m2", DocumentRef: "doc1", CreatedAt: time.Now().UTC()})

	if err := repo.DeleteByDocumentRef(ctx, "doc1"); err != nil {
		t.Fatalf("DeleteByDocumentRef: %v", err)
	}

	found, _ := repo.FindByDocumentRef(ctx, "doc1")
	if len(found) != 0 {
		t.Errorf("count after delete = %d, want 0", len(found))
	}
}

func TestMediaAssetRepository_Delete(t *testing.T) {
	db := setupMediaDB(t)
	repo := NewMediaAssetRepository(db)
	ctx := context.Background()

	_ = repo.Create(ctx, &entity.MediaAsset{ID: "m1", CreatedAt: time.Now().UTC()})

	if err := repo.Delete(ctx, "m1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := repo.FindByID(ctx, "m1")
	if !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
		t.Errorf("after delete, err = %v, want ErrNotFound", err)
	}
}
