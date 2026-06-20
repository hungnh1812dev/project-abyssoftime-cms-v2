package repository

import (
	"context"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

type RoleRepository interface {
	Create(ctx context.Context, role *entity.RoleEntity) error
	FindByID(ctx context.Context, documentID string) (*entity.RoleEntity, error)
	FindBySlug(ctx context.Context, slug string) (*entity.RoleEntity, error)
	FindAll(ctx context.Context) ([]*entity.RoleEntity, error)
	Update(ctx context.Context, role *entity.RoleEntity) error
	Delete(ctx context.Context, documentID string) error
	HasAny(ctx context.Context) (bool, error)
}
