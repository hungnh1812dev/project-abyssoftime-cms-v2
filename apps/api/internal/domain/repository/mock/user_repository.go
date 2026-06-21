package mock

import (
	"context"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
)

var _ repository.UserRepository = (*UserRepository)(nil)

type UserRepository struct {
	CreateFn        func(ctx context.Context, user *entity.User) error
	FindByEmailFn   func(ctx context.Context, email string) (*entity.User, error)
	FindByIDFn      func(ctx context.Context, id string) (*entity.User, error)
	FindByIDsFn     func(ctx context.Context, ids []string) ([]*entity.User, error)
	HasSuperAdminFn func(ctx context.Context) (bool, error)
	FindAllFn       func(ctx context.Context, page, limit int) ([]*entity.User, int64, error)
	UpdateFn        func(ctx context.Context, user *entity.User) error
	DeleteFn        func(ctx context.Context, id string) error
}

func (m *UserRepository) Create(ctx context.Context, user *entity.User) error {
	return m.CreateFn(ctx, user)
}

func (m *UserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	return m.FindByEmailFn(ctx, email)
}

func (m *UserRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	return m.FindByIDFn(ctx, id)
}

func (m *UserRepository) FindByIDs(ctx context.Context, ids []string) ([]*entity.User, error) {
	if m.FindByIDsFn != nil {
		return m.FindByIDsFn(ctx, ids)
	}
	return nil, nil
}

func (m *UserRepository) HasSuperAdmin(ctx context.Context) (bool, error) {
	return m.HasSuperAdminFn(ctx)
}

func (m *UserRepository) FindAll(ctx context.Context, page, limit int) ([]*entity.User, int64, error) {
	return m.FindAllFn(ctx, page, limit)
}

func (m *UserRepository) Update(ctx context.Context, user *entity.User) error {
	return m.UpdateFn(ctx, user)
}

func (m *UserRepository) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}
