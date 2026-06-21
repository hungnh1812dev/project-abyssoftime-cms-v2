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
		DocumentID:   "m1",
		URL:          "https://example.com/m1.jpg",
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
	if found.URL != "https://example.com/m1.jpg" {
		t.Errorf("URL = %q, want %q", found.URL, "https://example.com/m1.jpg")
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

func TestMediaAssetRepository_FindByDocumentID(t *testing.T) {
	db := setupMediaDB(t)
	repo := NewMediaAssetRepository(db)
	ctx := context.Background()

	_ = repo.Create(ctx, &entity.MediaAsset{DocumentID: "doc-uuid-1", URL: "https://cdn/a.jpg", CreatedAt: time.Now().UTC()})
	_ = repo.Create(ctx, &entity.MediaAsset{DocumentID: "doc-uuid-2", URL: "https://cdn/b.jpg", CreatedAt: time.Now().UTC()})

	found, err := repo.FindByDocumentID(ctx, "doc-uuid-1")
	if err != nil {
		t.Fatalf("FindByDocumentID: %v", err)
	}
	if found.URL != "https://cdn/a.jpg" {
		t.Errorf("URL = %q, want %q", found.URL, "https://cdn/a.jpg")
	}

	_, err = repo.FindByDocumentID(ctx, "nonexistent")
	if !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
		t.Errorf("FindByDocumentID(nonexistent) err = %v, want ErrNotFound", err)
	}
}

func TestMediaAssetRepository_FindAll_Paginated(t *testing.T) {
	db := setupMediaDB(t)
	repo := NewMediaAssetRepository(db)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		_ = repo.Create(ctx, &entity.MediaAsset{
			DocumentID: "doc-" + string(rune('0'+i)),
			CreatedAt:  time.Now().UTC().Add(time.Duration(i) * time.Second),
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

func TestMediaAssetRepository_Delete(t *testing.T) {
	db := setupMediaDB(t)
	repo := NewMediaAssetRepository(db)
	ctx := context.Background()
	m1 := &entity.MediaAsset{DocumentID: "m1", URL: "url1", CreatedAt: time.Now().UTC().Add(-time.Hour)}
	m2 := &entity.MediaAsset{DocumentID: "m2", URL: "url2", CreatedAt: time.Now().UTC()}

	_ = repo.Create(ctx, m1)
	_ = repo.Create(ctx, m2)

	if err := repo.Delete(ctx, "m1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := repo.FindByID(ctx, "m1")
	if !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
		t.Errorf("after delete, err = %v, want ErrNotFound", err)
	}
}
