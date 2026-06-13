package media_test

import (
	"bytes"
	"context"
	"errors"
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
	storage.UploadFn = func(_ context.Context, _ io.Reader, filename string) (*repository.UploadResult, error) {
		return &repository.UploadResult{URL: "https://cdn.example.com/photo.jpg", PublicID: "photo-id"}, nil
	}

	var capturedAsset *entity.MediaAsset
	assetRepo := &repomock.MediaAssetRepository{}
	assetRepo.CreateFn = func(_ context.Context, asset *entity.MediaAsset) error {
		capturedAsset = asset
		return nil
	}

	uc := mediauc.New(assetRepo, storage)
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

func TestUpload_StorageError_ReturnsError(t *testing.T) {
	storage := &repomock.StorageAdapter{}
	storageErr := errors.New("cloudinary unavailable")
	storage.UploadFn = func(_ context.Context, _ io.Reader, _ string) (*repository.UploadResult, error) {
		return nil, storageErr
	}
	assetRepo := &repomock.MediaAssetRepository{}

	uc := mediauc.New(assetRepo, storage)
	_, err := uc.Upload(ctx, bytes.NewReader([]byte("img")), "photo.jpg", "doc-1", "ct-1")
	if !errors.Is(err, storageErr) {
		t.Errorf("Upload() error = %v, want %v", err, storageErr)
	}
}

func TestUpload_RepoError_ReturnsError(t *testing.T) {
	storage := &repomock.StorageAdapter{}
	storage.UploadFn = func(_ context.Context, _ io.Reader, _ string) (*repository.UploadResult, error) {
		return &repository.UploadResult{URL: "https://cdn.example.com/photo.jpg", PublicID: "photo-id"}, nil
	}
	repoErr := errors.New("mongo write failed")
	assetRepo := &repomock.MediaAssetRepository{}
	assetRepo.CreateFn = func(_ context.Context, _ *entity.MediaAsset) error {
		return repoErr
	}

	uc := mediauc.New(assetRepo, storage)
	_, err := uc.Upload(ctx, bytes.NewReader([]byte("img")), "photo.jpg", "doc-1", "ct-1")
	if !errors.Is(err, repoErr) {
		t.Errorf("Upload() error = %v, want %v", err, repoErr)
	}
}
