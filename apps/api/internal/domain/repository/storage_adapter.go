package repository

import (
	"context"
	"io"
)

type UploadResult struct {
	URL      string
	PublicID string
}

type StorageAdapter interface {
	Upload(ctx context.Context, file io.Reader, filename string) (*UploadResult, error)
	Delete(ctx context.Context, publicID string) error
}
