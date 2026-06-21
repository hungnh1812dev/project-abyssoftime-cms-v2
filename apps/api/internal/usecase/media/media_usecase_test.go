package media_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"testing"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
	repomock "project-abyssoftime-cms-v2/api/internal/domain/repository/mock"
	mediauc "project-abyssoftime-cms-v2/api/internal/usecase/media"
)

var ctx = context.Background()

// ---- Upload ----------------------------------------------------------------

func TestUpload_CreatesMediaAsset(t *testing.T) {
	storage := &repomock.StorageAdapter{}
	storage.UploadFn = func(_ context.Context, _ io.Reader, filename string, _ bool) (*repository.UploadResult, error) {
		return &repository.UploadResult{URL: "https://cdn.example.com/photo.jpg", ThumbnailURL: "https://cdn.example.com/photo.jpg", PublicID: "photo-id"}, nil
	}

	var capturedAsset *entity.MediaAsset
	assetRepo := &repomock.MediaAssetRepository{}
	assetRepo.CreateFn = func(_ context.Context, asset *entity.MediaAsset) error {
		capturedAsset = asset
		return nil
	}

	uc := mediauc.New(assetRepo, storage, false)
	got, err := uc.Upload(ctx, bytes.NewReader([]byte("img")), "photo.jpg")
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	if got.URL != "https://cdn.example.com/photo.jpg" {
		t.Errorf("Upload() URL = %v, want https://cdn.example.com/photo.jpg", got.URL)
	}
	if capturedAsset == nil {
		t.Fatal("Upload() did not call repo.Create")
	}
}

func TestUpload_PersistsThumbnailURL(t *testing.T) {
	storage := &repomock.StorageAdapter{}
	storage.UploadFn = func(_ context.Context, _ io.Reader, _ string, _ bool) (*repository.UploadResult, error) {
		return &repository.UploadResult{
			URL:          "https://cdn.example.com/photo.jpg",
			ThumbnailURL: "https://cdn.example.com/thumb_photo.jpg",
			PublicID:     "photo-id",
		}, nil
	}
	var capturedAsset *entity.MediaAsset
	assetRepo := &repomock.MediaAssetRepository{}
	assetRepo.CreateFn = func(_ context.Context, asset *entity.MediaAsset) error {
		capturedAsset = asset
		return nil
	}

	uc := mediauc.New(assetRepo, storage, true)
	_, err := uc.Upload(ctx, bytes.NewReader([]byte("img")), "photo.jpg")
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	if capturedAsset.ThumbnailURL != "https://cdn.example.com/thumb_photo.jpg" {
		t.Errorf("Upload() ThumbnailURL = %q, want %q", capturedAsset.ThumbnailURL, "https://cdn.example.com/thumb_photo.jpg")
	}
}

func TestUpload_AutoThumbnailEnabled_PassesGenerateThumbnailTrue(t *testing.T) {
	var gotGenerateThumbnail bool
	storage := &repomock.StorageAdapter{}
	storage.UploadFn = func(_ context.Context, _ io.Reader, _ string, generateThumbnail bool) (*repository.UploadResult, error) {
		gotGenerateThumbnail = generateThumbnail
		return &repository.UploadResult{URL: "u", ThumbnailURL: "t", PublicID: "p"}, nil
	}
	assetRepo := &repomock.MediaAssetRepository{}
	assetRepo.CreateFn = func(_ context.Context, _ *entity.MediaAsset) error { return nil }

	uc := mediauc.New(assetRepo, storage, true)
	_, _ = uc.Upload(ctx, bytes.NewReader([]byte("img")), "photo.jpg")
	if !gotGenerateThumbnail {
		t.Error("Upload() did not pass generateThumbnail=true to storage when MediaAutoThumbnail=true")
	}
}

func TestUpload_AutoThumbnailDisabled_PassesGenerateThumbnailFalse(t *testing.T) {
	var gotGenerateThumbnail bool
	storage := &repomock.StorageAdapter{}
	storage.UploadFn = func(_ context.Context, _ io.Reader, _ string, generateThumbnail bool) (*repository.UploadResult, error) {
		gotGenerateThumbnail = generateThumbnail
		return &repository.UploadResult{URL: "u", ThumbnailURL: "u", PublicID: "p"}, nil
	}
	assetRepo := &repomock.MediaAssetRepository{}
	assetRepo.CreateFn = func(_ context.Context, _ *entity.MediaAsset) error { return nil }

	uc := mediauc.New(assetRepo, storage, false)
	_, _ = uc.Upload(ctx, bytes.NewReader([]byte("img")), "photo.jpg")
	if gotGenerateThumbnail {
		t.Error("Upload() passed generateThumbnail=true to storage when MediaAutoThumbnail=false")
	}
}

func TestUpload_StorageError_ReturnsError(t *testing.T) {
	storage := &repomock.StorageAdapter{}
	storageErr := errors.New("cloudinary unavailable")
	storage.UploadFn = func(_ context.Context, _ io.Reader, _ string, _ bool) (*repository.UploadResult, error) {
		return nil, storageErr
	}
	assetRepo := &repomock.MediaAssetRepository{}

	uc := mediauc.New(assetRepo, storage, false)
	_, err := uc.Upload(ctx, bytes.NewReader([]byte("img")), "photo.jpg")
	if !errors.Is(err, storageErr) {
		t.Errorf("Upload() error = %v, want %v", err, storageErr)
	}
}

func TestList_ReturnsPaginatedAssets(t *testing.T) {
	assetRepo := &repomock.MediaAssetRepository{}
	assetRepo.FindAllFn = func(_ context.Context, page, limit int) ([]*entity.MediaAsset, int64, error) {
		return []*entity.MediaAsset{
			{DocumentID: "a1", URL: "https://cdn/a1.jpg"},
			{DocumentID: "a2", URL: "https://cdn/a2.jpg"},
		}, 10, nil
	}
	storage := &repomock.StorageAdapter{}
	uc := mediauc.New(assetRepo, storage, false)

	items, total, err := uc.List(ctx, 1, 2)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if total != 10 {
		t.Errorf("List() total = %d, want 10", total)
	}
	if len(items) != 2 {
		t.Errorf("List() items count = %d, want 2", len(items))
	}
}

func TestUpload_BuildsHashedFilename(t *testing.T) {
	content := []byte("deterministic content")
	sum := sha256.Sum256(content)
	expectedHash := fmt.Sprintf("%x", sum)[:12]

	var capturedFilename string
	storage := &repomock.StorageAdapter{}
	storage.UploadFn = func(_ context.Context, _ io.Reader, filename string, _ bool) (*repository.UploadResult, error) {
		capturedFilename = filename
		return &repository.UploadResult{URL: "u", ThumbnailURL: "t", PublicID: "p"}, nil
	}
	assetRepo := &repomock.MediaAssetRepository{}
	assetRepo.CreateFn = func(_ context.Context, _ *entity.MediaAsset) error { return nil }

	uc := mediauc.New(assetRepo, storage, false)
	if _, err := uc.Upload(ctx, bytes.NewReader(content), "photo.jpg"); err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	want := "photo_" + expectedHash + ".jpg"
	if capturedFilename != want {
		t.Errorf("Upload() filename passed to storage = %q, want %q", capturedFilename, want)
	}
}

func TestUpload_HashedFilenameIsDeterministic(t *testing.T) {
	content := []byte("same bytes")
	var names []string
	storage := &repomock.StorageAdapter{}
	storage.UploadFn = func(_ context.Context, _ io.Reader, filename string, _ bool) (*repository.UploadResult, error) {
		names = append(names, filename)
		return &repository.UploadResult{URL: "u", ThumbnailURL: "t", PublicID: "p"}, nil
	}
	assetRepo := &repomock.MediaAssetRepository{}
	assetRepo.CreateFn = func(_ context.Context, _ *entity.MediaAsset) error { return nil }

	uc := mediauc.New(assetRepo, storage, false)
	for i := 0; i < 3; i++ {
		if _, err := uc.Upload(ctx, bytes.NewReader(content), "photo.jpg"); err != nil {
			t.Fatalf("Upload() run %d error = %v", i, err)
		}
	}
	if names[0] != names[1] || names[1] != names[2] {
		t.Errorf("Upload() filenames not deterministic: %v", names)
	}
}

func TestUpload_PopulatesFileFields(t *testing.T) {
	content := []byte("file content")
	storage := &repomock.StorageAdapter{}
	storage.UploadFn = func(_ context.Context, _ io.Reader, _ string, _ bool) (*repository.UploadResult, error) {
		return &repository.UploadResult{URL: "u", ThumbnailURL: "t", PublicID: "p"}, nil
	}
	var capturedAsset *entity.MediaAsset
	assetRepo := &repomock.MediaAssetRepository{}
	assetRepo.CreateFn = func(_ context.Context, asset *entity.MediaAsset) error {
		capturedAsset = asset
		return nil
	}

	uc := mediauc.New(assetRepo, storage, false)
	if _, err := uc.Upload(ctx, bytes.NewReader(content), "photo.jpg"); err != nil {
		t.Fatalf("Upload() error = %v", err)
	}

	sum := sha256.Sum256(content)
	wantHash := fmt.Sprintf("%x", sum)[:12]

	if capturedAsset.FileName != "photo_"+wantHash+".jpg" {
		t.Errorf("Upload() asset.FileName = %q, want %q", capturedAsset.FileName, "photo_"+wantHash+".jpg")
	}
	if capturedAsset.FileExt != "jpg" {
		t.Errorf("Upload() asset.FileExt = %q, want %q", capturedAsset.FileExt, "jpg")
	}
	if capturedAsset.Hash != wantHash {
		t.Errorf("Upload() asset.Hash = %q, want %q", capturedAsset.Hash, wantHash)
	}
}

func TestUpload_RepoError_ReturnsError(t *testing.T) {
	storage := &repomock.StorageAdapter{}
	storage.UploadFn = func(_ context.Context, _ io.Reader, _ string, _ bool) (*repository.UploadResult, error) {
		return &repository.UploadResult{URL: "https://cdn.example.com/photo.jpg", ThumbnailURL: "https://cdn.example.com/photo.jpg", PublicID: "photo-id"}, nil
	}
	repoErr := errors.New("mongo write failed")
	assetRepo := &repomock.MediaAssetRepository{}
	assetRepo.CreateFn = func(_ context.Context, _ *entity.MediaAsset) error {
		return repoErr
	}

	uc := mediauc.New(assetRepo, storage, false)
	_, err := uc.Upload(ctx, bytes.NewReader([]byte("img")), "photo.jpg")
	if !errors.Is(err, repoErr) {
		t.Errorf("Upload() error = %v, want %v", err, repoErr)
	}
}

// ---- Delete ----------------------------------------------------------------

func TestDelete_CallsStorageAndRepo(t *testing.T) {
	var storageDeleteID string
	var repoDeleteID string

	assetRepo := &repomock.MediaAssetRepository{}
	assetRepo.FindByIDFn = func(_ context.Context, id string) (*entity.MediaAsset, error) {
		return &entity.MediaAsset{DocumentID: id, PublicID: "pub-123"}, nil
	}
	assetRepo.DeleteFn = func(_ context.Context, id string) error {
		repoDeleteID = id
		return nil
	}

	storage := &repomock.StorageAdapter{}
	storage.DeleteFn = func(_ context.Context, publicID string) error {
		storageDeleteID = publicID
		return nil
	}

	uc := mediauc.New(assetRepo, storage, false)
	if err := uc.Delete(ctx, "asset-1"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if storageDeleteID != "pub-123" {
		t.Errorf("Delete() storage.Delete publicID = %q, want %q", storageDeleteID, "pub-123")
	}
	if repoDeleteID != "asset-1" {
		t.Errorf("Delete() repo.Delete id = %q, want %q", repoDeleteID, "asset-1")
	}
}

func TestDelete_AssetNotFound_ReturnsError(t *testing.T) {
	notFound := errors.New("not found")
	storageDeleteCalled := false

	assetRepo := &repomock.MediaAssetRepository{}
	assetRepo.FindByIDFn = func(_ context.Context, _ string) (*entity.MediaAsset, error) {
		return nil, notFound
	}

	storage := &repomock.StorageAdapter{}
	storage.DeleteFn = func(_ context.Context, _ string) error {
		storageDeleteCalled = true
		return nil
	}

	uc := mediauc.New(assetRepo, storage, false)
	err := uc.Delete(ctx, "asset-1")
	if !errors.Is(err, notFound) {
		t.Errorf("Delete() error = %v, want %v", err, notFound)
	}
	if storageDeleteCalled {
		t.Error("Delete() called storage.Delete after FindByID returned an error")
	}
}

func TestDelete_StorageError_DoesNotDeleteFromRepo(t *testing.T) {
	storageErr := errors.New("cloudinary error")
	repoDeleteCalled := false

	assetRepo := &repomock.MediaAssetRepository{}
	assetRepo.FindByIDFn = func(_ context.Context, id string) (*entity.MediaAsset, error) {
		return &entity.MediaAsset{DocumentID: id, PublicID: "pub-123"}, nil
	}
	assetRepo.DeleteFn = func(_ context.Context, _ string) error {
		repoDeleteCalled = true
		return nil
	}

	storage := &repomock.StorageAdapter{}
	storage.DeleteFn = func(_ context.Context, _ string) error {
		return storageErr
	}

	uc := mediauc.New(assetRepo, storage, false)
	err := uc.Delete(ctx, "asset-1")
	if !errors.Is(err, storageErr) {
		t.Errorf("Delete() error = %v, want %v", err, storageErr)
	}
	if repoDeleteCalled {
		t.Error("Delete() called repo.Delete after storage.Delete returned an error")
	}
}

func TestDelete_RepoDeleteError_ReturnsError(t *testing.T) {
	repoErr := errors.New("mongo write failed")

	assetRepo := &repomock.MediaAssetRepository{}
	assetRepo.FindByIDFn = func(_ context.Context, id string) (*entity.MediaAsset, error) {
		return &entity.MediaAsset{DocumentID: id, PublicID: "pub-123"}, nil
	}
	assetRepo.DeleteFn = func(_ context.Context, _ string) error {
		return repoErr
	}

	storage := &repomock.StorageAdapter{}
	storage.DeleteFn = func(_ context.Context, _ string) error {
		return nil
	}

	uc := mediauc.New(assetRepo, storage, false)
	err := uc.Delete(ctx, "asset-1")
	if !errors.Is(err, repoErr) {
		t.Errorf("Delete() error = %v, want %v", err, repoErr)
	}
}
