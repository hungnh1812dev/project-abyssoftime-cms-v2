package media

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"path/filepath"
	"strings"

	_ "golang.org/x/image/webp"

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

func (uc *UseCase) Delete(ctx context.Context, id string) error {
	asset, err := uc.assetRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if err := uc.storage.Delete(ctx, asset.PublicID); err != nil {
		return err
	}
	return uc.assetRepo.Delete(ctx, id)
}

func (uc *UseCase) Upload(ctx context.Context, file io.Reader, filename string) (*entity.MediaAsset, error) {
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
	var width, height int
	if cfg, _, err := image.DecodeConfig(bytes.NewReader(data)); err == nil {
		width = cfg.Width
		height = cfg.Height
	}

	asset := &entity.MediaAsset{
		URL:          result.URL,
		ThumbnailURL: result.ThumbnailURL,
		PublicID:     result.PublicID,
		FileName:     hashedFilename,
		FileExt:      ext,
		Hash:         hash12,
		Width:        width,
		Height:       height,
	}
	if err := uc.assetRepo.Create(ctx, asset); err != nil {
		return nil, err
	}
	return asset, nil
}
