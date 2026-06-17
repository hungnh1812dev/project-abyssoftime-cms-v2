package media

import (
	"context"
	"io"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
)

type UseCase struct {
	assetRepo          repository.MediaAssetRepository
	storage            repository.StorageAdapter
	mediaAutoThumbnail bool
}

func New(assetRepo repository.MediaAssetRepository, storage repository.StorageAdapter, mediaAutoThumbnail bool) *UseCase {
	return &UseCase{assetRepo: assetRepo, storage: storage, mediaAutoThumbnail: mediaAutoThumbnail}
}

func (uc *UseCase) Upload(ctx context.Context, file io.Reader, filename, documentRef, contentTypeID string) (*entity.MediaAsset, error) {
	result, err := uc.storage.Upload(ctx, file, filename, uc.mediaAutoThumbnail)
	if err != nil {
		return nil, err
	}
	asset := &entity.MediaAsset{
		URL:           result.URL,
		ThumbnailURL:  result.ThumbnailURL,
		PublicID:      result.PublicID,
		DocumentRef:   documentRef,
		ContentTypeID: contentTypeID,
	}
	if err := uc.assetRepo.Create(ctx, asset); err != nil {
		return nil, err
	}
	return asset, nil
}
