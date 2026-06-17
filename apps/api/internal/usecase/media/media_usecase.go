package media

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"path/filepath"
	"strings"

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

func (uc *UseCase) List(ctx context.Context, page, limit int) ([]*entity.MediaAsset, int64, error) {
	return uc.assetRepo.FindAll(ctx, page, limit)
}

func (uc *UseCase) Upload(ctx context.Context, file io.Reader, filename, documentRef, contentTypeID string) (*entity.MediaAsset, error) {
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	sum := sha256.Sum256(data)
	hash12 := fmt.Sprintf("%x", sum)[:12]

	ext := strings.TrimPrefix(filepath.Ext(filename), ".")
	stem := strings.TrimSuffix(filename, filepath.Ext(filename))
	hashedFilename := stem + "_" + hash12 + "." + ext

	result, err := uc.storage.Upload(ctx, bytes.NewReader(data), hashedFilename, uc.mediaAutoThumbnail)
	if err != nil {
		return nil, err
	}
	asset := &entity.MediaAsset{
		URL:           result.URL,
		ThumbnailURL:  result.ThumbnailURL,
		PublicID:      result.PublicID,
		FileName:      hashedFilename,
		FileExt:       ext,
		Hash:          hash12,
		DocumentRef:   documentRef,
		ContentTypeID: contentTypeID,
	}
	if err := uc.assetRepo.Create(ctx, asset); err != nil {
		return nil, err
	}
	return asset, nil
}
