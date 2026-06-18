package repository

import (
	"context"
	"io"
)

type UploadResult struct {
	URL          string
	ThumbnailURL string
	PublicID     string
}

type StorageAdapter interface {
	Upload(ctx context.Context, file io.Reader, filename string, generateThumbnail bool) (*UploadResult, error)
	Delete(ctx context.Context, publicID string) error
}
