package mock

import (
	"context"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
)

var _ repository.InviteRepository = (*InviteRepository)(nil)

type InviteRepository struct {
	CreateFn       func(ctx context.Context, invite *entity.Invite) error
	FindByHashFn   func(ctx context.Context, tokenHash string) (*entity.Invite, error)
	FindByEmailFn  func(ctx context.Context, email string) (*entity.Invite, error)
	DeleteFn       func(ctx context.Context, id string) error
	DeleteExpiredFn func(ctx context.Context) error
	FindAllFn      func(ctx context.Context) ([]*entity.Invite, error)
}

func (m *InviteRepository) Create(ctx context.Context, invite *entity.Invite) error {
	return m.CreateFn(ctx, invite)
}

func (m *InviteRepository) FindByHash(ctx context.Context, tokenHash string) (*entity.Invite, error) {
	return m.FindByHashFn(ctx, tokenHash)
}

func (m *InviteRepository) FindByEmail(ctx context.Context, email string) (*entity.Invite, error) {
	return m.FindByEmailFn(ctx, email)
}

func (m *InviteRepository) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}

func (m *InviteRepository) DeleteExpired(ctx context.Context) error {
	return m.DeleteExpiredFn(ctx)
}

func (m *InviteRepository) FindAll(ctx context.Context) ([]*entity.Invite, error) {
	return m.FindAllFn(ctx)
}
