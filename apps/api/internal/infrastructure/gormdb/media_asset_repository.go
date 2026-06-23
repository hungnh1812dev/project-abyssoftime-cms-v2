package gormdb

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

var _ repository.MediaAssetRepository = (*mediaAssetRepository)(nil)

type mediaAssetRepository struct {
	database *gorm.DB
}

func NewMediaAssetRepository(database *gorm.DB) repository.MediaAssetRepository {
	return &mediaAssetRepository{database: database}
}

func (r *mediaAssetRepository) Create(ctx context.Context, asset *entity.MediaAsset) error {
	return r.database.WithContext(ctx).Create(asset).Error
}

func (r *mediaAssetRepository) FindByID(ctx context.Context, id string) (*entity.MediaAsset, error) {
	var asset entity.MediaAsset
	if err := r.database.WithContext(ctx).Where("document_id = ?", id).First(&asset).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, err
	}
	return &asset, nil
}

func (r *mediaAssetRepository) FindByDocumentID(ctx context.Context, documentID string) (*entity.MediaAsset, error) {
	var asset entity.MediaAsset
	if err := r.database.WithContext(ctx).Where("document_id = ?", documentID).First(&asset).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, err
	}
	return &asset, nil
}

func (r *mediaAssetRepository) FindAll(ctx context.Context, page, limit int) ([]*entity.MediaAsset, int64, error) {
	var total int64
	if err := r.database.WithContext(ctx).Model(&entity.MediaAsset{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	var assets []*entity.MediaAsset
	if err := r.database.WithContext(ctx).Order("created_at DESC").Offset(offset).Limit(limit).Find(&assets).Error; err != nil {
		return nil, 0, err
	}
	return assets, total, nil
}

func (r *mediaAssetRepository) Delete(ctx context.Context, id string) error {
	return r.database.WithContext(ctx).Where("document_id = ?", id).Delete(&entity.MediaAsset{}).Error
}
