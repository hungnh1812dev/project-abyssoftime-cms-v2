package repository

import (
	"context"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

type LocaleRepository interface {
	Create(ctx context.Context, locale *entity.Locale) error
	FindByCode(ctx context.Context, code string) (*entity.Locale, error)
	FindAll(ctx context.Context) ([]*entity.Locale, error)
	FindDefault(ctx context.Context) (*entity.Locale, error)
	Update(ctx context.Context, locale *entity.Locale) error
	Delete(ctx context.Context, code string) error
	ClearDefault(ctx context.Context) error
}
