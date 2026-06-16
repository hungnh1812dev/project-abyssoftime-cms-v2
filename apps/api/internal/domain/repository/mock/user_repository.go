package mock

import (
	"context"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
)

var _ repository.UserRepository = (*UserRepository)(nil)

// UserRepository is a test double for repository.UserRepository.
// Set each Fn field to a stub before calling the method under test.
type UserRepository struct {
	CreateFn       func(ctx context.Context, user *entity.User) error
	FindByEmailFn  func(ctx context.Context, email string) (*entity.User, error)
	FindByIDFn     func(ctx context.Context, id string) (*entity.User, error)
	CountAdminsFn  func(ctx context.Context) (int64, error)
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

func (m *UserRepository) CountAdmins(ctx context.Context) (int64, error) {
	return m.CountAdminsFn(ctx)
}
