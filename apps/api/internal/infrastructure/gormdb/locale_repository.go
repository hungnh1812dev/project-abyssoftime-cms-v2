package gormdb

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

var _ repository.LocaleRepository = (*localeRepository)(nil)

type localeRepository struct {
	database *gorm.DB
}

func NewLocaleRepository(database *gorm.DB) repository.LocaleRepository {
	return &localeRepository{database: database}
}

func (repo *localeRepository) Create(ctx context.Context, locale *entity.Locale) error {
	return repo.database.WithContext(ctx).Create(locale).Error
}

func (repo *localeRepository) FindByCode(ctx context.Context, code string) (*entity.Locale, error) {
	var locale entity.Locale
	if err := repo.database.WithContext(ctx).Where("code = ?", code).First(&locale).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, err
	}
	return &locale, nil
}

func (repo *localeRepository) FindAll(ctx context.Context) ([]*entity.Locale, error) {
	var locales []*entity.Locale
	if err := repo.database.WithContext(ctx).Order("code ASC").Find(&locales).Error; err != nil {
		return nil, err
	}
	return locales, nil
}

func (repo *localeRepository) FindDefault(ctx context.Context) (*entity.Locale, error) {
	var locale entity.Locale
	if err := repo.database.WithContext(ctx).Where("is_default = ?", true).First(&locale).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, err
	}
	return &locale, nil
}

func (repo *localeRepository) Update(ctx context.Context, locale *entity.Locale) error {
	result := repo.database.WithContext(ctx).Save(locale)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return pkgerrors.ErrNotFound
	}
	return nil
}

func (repo *localeRepository) Delete(ctx context.Context, code string) error {
	result := repo.database.WithContext(ctx).Where("code = ?", code).Delete(&entity.Locale{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return pkgerrors.ErrNotFound
	}
	return nil
}

func (repo *localeRepository) ClearDefault(ctx context.Context) error {
	return repo.database.WithContext(ctx).
		Model(&entity.Locale{}).
		Where("is_default = ?", true).
		Update("is_default", false).Error
}
