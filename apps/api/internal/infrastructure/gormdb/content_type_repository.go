package gormdb

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

var _ repository.ContentTypeRepository = (*contentTypeRepository)(nil)

type contentTypeRepository struct {
	db *gorm.DB
}

func NewContentTypeRepository(db *gorm.DB) repository.ContentTypeRepository {
	return &contentTypeRepository{db: db}
}

func (r *contentTypeRepository) Create(ctx context.Context, ct *entity.ContentType) error {
	return r.db.WithContext(ctx).Create(ct).Error
}

func (r *contentTypeRepository) FindByID(ctx context.Context, id string) (*entity.ContentType, error) {
	var ct entity.ContentType
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&ct).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, err
	}
	return &ct, nil
}

func (r *contentTypeRepository) FindBySlug(ctx context.Context, slug string) (*entity.ContentType, error) {
	var ct entity.ContentType
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&ct).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, err
	}
	return &ct, nil
}

func (r *contentTypeRepository) FindAll(ctx context.Context) ([]*entity.ContentType, error) {
	var cts []*entity.ContentType
	if err := r.db.WithContext(ctx).Find(&cts).Error; err != nil {
		return nil, err
	}
	return cts, nil
}

func (r *contentTypeRepository) Update(ctx context.Context, ct *entity.ContentType) error {
	return r.db.WithContext(ctx).Save(ct).Error
}

func (r *contentTypeRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.ContentType{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return pkgerrors.ErrNotFound
	}
	return nil
}
