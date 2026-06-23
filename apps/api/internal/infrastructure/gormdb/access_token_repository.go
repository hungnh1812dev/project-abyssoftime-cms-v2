package gormdb

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

var _ repository.AccessTokenRepository = (*accessTokenRepository)(nil)

type accessTokenRepository struct {
	database *gorm.DB
}

func NewAccessTokenRepository(database *gorm.DB) repository.AccessTokenRepository {
	return &accessTokenRepository{database: database}
}

func (r *accessTokenRepository) Create(ctx context.Context, token *entity.AccessToken) error {
	return r.database.WithContext(ctx).Create(token).Error
}

func (r *accessTokenRepository) FindByHash(ctx context.Context, tokenHash string) (*entity.AccessToken, error) {
	var token entity.AccessToken
	if err := r.database.WithContext(ctx).Where("token_hash = ?", tokenHash).First(&token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, err
	}
	return &token, nil
}

func (r *accessTokenRepository) FindAll(ctx context.Context, page, limit int) ([]*entity.AccessToken, int64, error) {
	var total int64
	if err := r.database.WithContext(ctx).Model(&entity.AccessToken{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var tokens []*entity.AccessToken
	if err := r.database.WithContext(ctx).Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&tokens).Error; err != nil {
		return nil, 0, err
	}
	return tokens, total, nil
}

func (r *accessTokenRepository) Delete(ctx context.Context, id string) error {
	result := r.database.WithContext(ctx).Where("document_id = ?", id).Delete(&entity.AccessToken{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return pkgerrors.ErrNotFound
	}
	return nil
}

func (r *accessTokenRepository) UpdateLastUsed(ctx context.Context, id string, at time.Time) error {
	return r.database.WithContext(ctx).Model(&entity.AccessToken{}).Where("gorm_id = ?", id).Update("last_used_at", at).Error
}
