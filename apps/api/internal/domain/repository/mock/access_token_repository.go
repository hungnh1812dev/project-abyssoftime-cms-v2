package mock

import (
	"context"
	"time"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
)

var _ repository.AccessTokenRepository = (*AccessTokenRepository)(nil)

type AccessTokenRepository struct {
	CreateFn       func(ctx context.Context, token *entity.AccessToken) error
	FindByHashFn   func(ctx context.Context, tokenHash string) (*entity.AccessToken, error)
	FindAllFn      func(ctx context.Context, page, limit int) ([]*entity.AccessToken, int64, error)
	DeleteFn       func(ctx context.Context, id string) error
	UpdateLastUsedFn func(ctx context.Context, id string, at time.Time) error
}

func (m *AccessTokenRepository) Create(ctx context.Context, token *entity.AccessToken) error {
	return m.CreateFn(ctx, token)
}

func (m *AccessTokenRepository) FindByHash(ctx context.Context, tokenHash string) (*entity.AccessToken, error) {
	return m.FindByHashFn(ctx, tokenHash)
}

func (m *AccessTokenRepository) FindAll(ctx context.Context, page, limit int) ([]*entity.AccessToken, int64, error) {
	return m.FindAllFn(ctx, page, limit)
}

func (m *AccessTokenRepository) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}

func (m *AccessTokenRepository) UpdateLastUsed(ctx context.Context, id string, at time.Time) error {
	return m.UpdateLastUsedFn(ctx, id, at)
}
