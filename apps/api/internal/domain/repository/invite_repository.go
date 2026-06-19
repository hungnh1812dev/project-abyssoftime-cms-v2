package repository

import (
	"context"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

type InviteRepository interface {
	Create(ctx context.Context, invite *entity.Invite) error
	FindByHash(ctx context.Context, tokenHash string) (*entity.Invite, error)
	FindByEmail(ctx context.Context, email string) (*entity.Invite, error)
	Delete(ctx context.Context, id string) error
	DeleteExpired(ctx context.Context) error
	FindAll(ctx context.Context) ([]*entity.Invite, error)
}
