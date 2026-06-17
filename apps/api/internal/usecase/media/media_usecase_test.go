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
	got, err := uc.Upload(ctx, bytes.NewReader([]byte("img")), "photo.jpg", "doc-1", "ct-1")
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	if got.URL != "https://cdn.example.com/photo.jpg" {
		t.Errorf("Upload() URL = %v, want https://cdn.example.com/photo.jpg", got.URL)
	}
	if capturedAsset == nil {
		t.Fatal("Upload() did not call repo.Create")
	}
	if capturedAsset.DocumentRef != "doc-1" {
		t.Errorf("Upload() DocumentRef = %v, want doc-1", capturedAsset.DocumentRef)
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
	_, err := uc.Upload(ctx, bytes.NewReader([]byte("img")), "photo.jpg", "doc-1", "ct-1")
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
	_, _ = uc.Upload(ctx, bytes.NewReader([]byte("img")), "photo.jpg", "doc-1", "ct-1")
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
	_, _ = uc.Upload(ctx, bytes.NewReader([]byte("img")), "photo.jpg", "doc-1", "ct-1")
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
	_, err := uc.Upload(ctx, bytes.NewReader([]byte("img")), "photo.jpg", "doc-1", "ct-1")
	if !errors.Is(err, storageErr) {
		t.Errorf("Upload() error = %v, want %v", err, storageErr)
	}
}

func TestList_ReturnsPaginatedAssets(t *testing.T) {
	assetRepo := &repomock.MediaAssetRepository{}
	assetRepo.FindAllFn = func(_ context.Context, page, limit int) ([]*entity.MediaAsset, int64, error) {
		return []*entity.MediaAsset{
			{ID: "a1", URL: "https://cdn/a1.jpg"},
			{ID: "a2", URL: "https://cdn/a2.jpg"},
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
	if _, err := uc.Upload(ctx, bytes.NewReader(content), "photo.jpg", "doc-1", "ct-1"); err != nil {
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
		if _, err := uc.Upload(ctx, bytes.NewReader(content), "photo.jpg", "doc", "ct"); err != nil {
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
	if _, err := uc.Upload(ctx, bytes.NewReader(content), "photo.jpg", "doc", "ct"); err != nil {
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
	_, err := uc.Upload(ctx, bytes.NewReader([]byte("img")), "photo.jpg", "doc-1", "ct-1")
	if !errors.Is(err, repoErr) {
		t.Errorf("Upload() error = %v, want %v", err, repoErr)
	}
}
