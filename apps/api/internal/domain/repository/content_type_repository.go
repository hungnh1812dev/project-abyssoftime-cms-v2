package repository

import (
	"context"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

type ContentTypeRepository interface {
	Create(ctx context.Context, ct *entity.ContentType) error
	FindByID(ctx context.Context, id string) (*entity.ContentType, error)
	FindBySlug(ctx context.Context, slug string) (*entity.ContentType, error)
	FindAll(ctx context.Context) ([]*entity.ContentType, error)
	Update(ctx context.Context, ct *entity.ContentType) error
	Delete(ctx context.Context, id string) error
}
