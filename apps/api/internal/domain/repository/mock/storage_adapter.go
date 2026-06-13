package mock

import (
	"context"
	"io"

	"project-abyssoftime-cms-v2/api/internal/domain/repository"
)

var _ repository.StorageAdapter = (*StorageAdapter)(nil)

// StorageAdapter is a test double for repository.StorageAdapter.
// Set each Fn field to a stub before calling the method under test.
type StorageAdapter struct {
	UploadFn func(ctx context.Context, file io.Reader, filename string) (*repository.UploadResult, error)
	DeleteFn func(ctx context.Context, publicID string) error
}

func (m *StorageAdapter) Upload(ctx context.Context, file io.Reader, filename string) (*repository.UploadResult, error) {
	return m.UploadFn(ctx, file, filename)
}

func (m *StorageAdapter) Delete(ctx context.Context, publicID string) error {
	return m.DeleteFn(ctx, publicID)
}
