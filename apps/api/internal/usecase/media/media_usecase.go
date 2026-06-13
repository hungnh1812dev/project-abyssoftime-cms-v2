package media

import (
	"context"
	"io"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
)

type UseCase struct {
	assetRepo repository.MediaAssetRepository
	storage   repository.StorageAdapter
}

func New(assetRepo repository.MediaAssetRepository, storage repository.StorageAdapter) *UseCase {
	return &UseCase{assetRepo: assetRepo, storage: storage}
}

func (uc *UseCase) Upload(ctx context.Context, file io.Reader, filename, documentRef, contentTypeID string) (*entity.MediaAsset, error) {
	result, err := uc.storage.Upload(ctx, file, filename)
	if err != nil {
		return nil, err
	}
	asset := &entity.MediaAsset{
		URL:           result.URL,
		PublicID:      result.PublicID,
		DocumentRef:   documentRef,
		ContentTypeID: contentTypeID,
	}
	if err := uc.assetRepo.Create(ctx, asset); err != nil {
		return nil, err
	}
	return asset, nil
}
