package mock

import (
	"context"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
)

var _ repository.LocaleRepository = (*LocaleRepository)(nil)

type LocaleRepository struct {
	CreateFn       func(ctx context.Context, locale *entity.Locale) error
	FindByCodeFn   func(ctx context.Context, code string) (*entity.Locale, error)
	FindAllFn      func(ctx context.Context) ([]*entity.Locale, error)
	FindDefaultFn  func(ctx context.Context) (*entity.Locale, error)
	UpdateFn       func(ctx context.Context, locale *entity.Locale) error
	DeleteFn       func(ctx context.Context, code string) error
	ClearDefaultFn func(ctx context.Context) error
}

func (mock *LocaleRepository) Create(ctx context.Context, locale *entity.Locale) error {
	return mock.CreateFn(ctx, locale)
}

func (mock *LocaleRepository) FindByCode(ctx context.Context, code string) (*entity.Locale, error) {
	return mock.FindByCodeFn(ctx, code)
}

func (mock *LocaleRepository) FindAll(ctx context.Context) ([]*entity.Locale, error) {
	return mock.FindAllFn(ctx)
}

func (mock *LocaleRepository) FindDefault(ctx context.Context) (*entity.Locale, error) {
	return mock.FindDefaultFn(ctx)
}

func (mock *LocaleRepository) Update(ctx context.Context, locale *entity.Locale) error {
	return mock.UpdateFn(ctx, locale)
}

func (mock *LocaleRepository) Delete(ctx context.Context, code string) error {
	return mock.DeleteFn(ctx, code)
}

func (mock *LocaleRepository) ClearDefault(ctx context.Context) error {
	return mock.ClearDefaultFn(ctx)
}
