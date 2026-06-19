package repository

import (
	"context"
	"time"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

type AccessTokenRepository interface {
	Create(ctx context.Context, token *entity.AccessToken) error
	FindByHash(ctx context.Context, tokenHash string) (*entity.AccessToken, error)
	FindAll(ctx context.Context, page, limit int) ([]*entity.AccessToken, int64, error)
	Delete(ctx context.Context, id string) error
	UpdateLastUsed(ctx context.Context, id string, at time.Time) error
}
