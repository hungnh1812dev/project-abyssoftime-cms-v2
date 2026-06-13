package mock

import (
	"context"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
)

var _ repository.MediaAssetRepository = (*MediaAssetRepository)(nil)

// MediaAssetRepository is a test double for repository.MediaAssetRepository.
// Set each Fn field to a stub before calling the method under test.
type MediaAssetRepository struct {
	CreateFn               func(ctx context.Context, asset *entity.MediaAsset) error
	FindByIDFn             func(ctx context.Context, id string) (*entity.MediaAsset, error)
	FindByDocumentRefFn    func(ctx context.Context, documentRef string) ([]*entity.MediaAsset, error)
	DeleteByDocumentRefFn  func(ctx context.Context, documentRef string) error
	DeleteFn               func(ctx context.Context, id string) error
}

func (m *MediaAssetRepository) Create(ctx context.Context, asset *entity.MediaAsset) error {
	return m.CreateFn(ctx, asset)
}

func (m *MediaAssetRepository) FindByID(ctx context.Context, id string) (*entity.MediaAsset, error) {
	return m.FindByIDFn(ctx, id)
}

func (m *MediaAssetRepository) FindByDocumentRef(ctx context.Context, documentRef string) ([]*entity.MediaAsset, error) {
	return m.FindByDocumentRefFn(ctx, documentRef)
}

func (m *MediaAssetRepository) DeleteByDocumentRef(ctx context.Context, documentRef string) error {
	return m.DeleteByDocumentRefFn(ctx, documentRef)
}

func (m *MediaAssetRepository) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}
