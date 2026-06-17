package repository

import (
	"context"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

type MediaAssetRepository interface {
	Create(ctx context.Context, asset *entity.MediaAsset) error
	FindByID(ctx context.Context, id string) (*entity.MediaAsset, error)
	FindByDocumentRef(ctx context.Context, documentRef string) ([]*entity.MediaAsset, error)
	FindAll(ctx context.Context, page, limit int) ([]*entity.MediaAsset, int64, error)
	DeleteByDocumentRef(ctx context.Context, documentRef string) error
	Delete(ctx context.Context, id string) error
}
