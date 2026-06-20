package mock

import (
	"context"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
)

var _ repository.RoleRepository = (*RoleRepository)(nil)

type RoleRepository struct {
	CreateFn     func(ctx context.Context, role *entity.RoleEntity) error
	FindByIDFn   func(ctx context.Context, documentID string) (*entity.RoleEntity, error)
	FindBySlugFn func(ctx context.Context, slug string) (*entity.RoleEntity, error)
	FindAllFn    func(ctx context.Context) ([]*entity.RoleEntity, error)
	UpdateFn     func(ctx context.Context, role *entity.RoleEntity) error
	DeleteFn     func(ctx context.Context, documentID string) error
	HasAnyFn     func(ctx context.Context) (bool, error)
}

func (m *RoleRepository) Create(ctx context.Context, role *entity.RoleEntity) error {
	return m.CreateFn(ctx, role)
}

func (m *RoleRepository) FindByID(ctx context.Context, documentID string) (*entity.RoleEntity, error) {
	return m.FindByIDFn(ctx, documentID)
}

func (m *RoleRepository) FindBySlug(ctx context.Context, slug string) (*entity.RoleEntity, error) {
	return m.FindBySlugFn(ctx, slug)
}

func (m *RoleRepository) FindAll(ctx context.Context) ([]*entity.RoleEntity, error) {
	return m.FindAllFn(ctx)
}

func (m *RoleRepository) Update(ctx context.Context, role *entity.RoleEntity) error {
	return m.UpdateFn(ctx, role)
}

func (m *RoleRepository) Delete(ctx context.Context, documentID string) error {
	return m.DeleteFn(ctx, documentID)
}

func (m *RoleRepository) HasAny(ctx context.Context) (bool, error) {
	return m.HasAnyFn(ctx)
}
