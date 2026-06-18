package repository

import (
	"context"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	FindByID(ctx context.Context, id string) (*entity.User, error)
	CountAdmins(ctx context.Context) (int64, error)
}
