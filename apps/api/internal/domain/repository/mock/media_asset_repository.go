package mock

import (
	"context"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
)

var _ repository.MediaAssetRepository = (*MediaAssetRepository)(nil)

type MediaAssetRepository struct {
	CreateFn         func(ctx context.Context, asset *entity.MediaAsset) error
	FindByIDFn       func(ctx context.Context, id string) (*entity.MediaAsset, error)
	FindByDocumentIDFn func(ctx context.Context, documentID string) (*entity.MediaAsset, error)
	FindAllFn        func(ctx context.Context, page, limit int) ([]*entity.MediaAsset, int64, error)
	DeleteFn         func(ctx context.Context, id string) error
}

func (m *MediaAssetRepository) Create(ctx context.Context, asset *entity.MediaAsset) error {
	return m.CreateFn(ctx, asset)
}

func (m *MediaAssetRepository) FindByID(ctx context.Context, id string) (*entity.MediaAsset, error) {
	return m.FindByIDFn(ctx, id)
}

func (m *MediaAssetRepository) FindByDocumentID(ctx context.Context, documentID string) (*entity.MediaAsset, error) {
	if m.FindByDocumentIDFn != nil {
		return m.FindByDocumentIDFn(ctx, documentID)
	}
	return nil, nil
}

func (m *MediaAssetRepository) FindAll(ctx context.Context, page, limit int) ([]*entity.MediaAsset, int64, error) {
	return m.FindAllFn(ctx, page, limit)
}

func (m *MediaAssetRepository) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}
