package gormdb

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

var _ repository.InviteRepository = (*inviteRepository)(nil)

type inviteRepository struct {
	database *gorm.DB
}

func NewInviteRepository(database *gorm.DB) repository.InviteRepository {
	return &inviteRepository{database: database}
}

func (r *inviteRepository) Create(ctx context.Context, invite *entity.Invite) error {
	return r.database.WithContext(ctx).Create(invite).Error
}

func (r *inviteRepository) FindByHash(ctx context.Context, tokenHash string) (*entity.Invite, error) {
	var inv entity.Invite
	if err := r.database.WithContext(ctx).Where("token_hash = ?", tokenHash).First(&inv).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, err
	}
	return &inv, nil
}

func (r *inviteRepository) FindByEmail(ctx context.Context, email string) (*entity.Invite, error) {
	var inv entity.Invite
	if err := r.database.WithContext(ctx).Where("email = ?", email).First(&inv).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, err
	}
	return &inv, nil
}

func (r *inviteRepository) Delete(ctx context.Context, id string) error {
	result := r.database.WithContext(ctx).Where("document_id = ?", id).Delete(&entity.Invite{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return pkgerrors.ErrNotFound
	}
	return nil
}

func (r *inviteRepository) DeleteExpired(ctx context.Context) error {
	return r.database.WithContext(ctx).Where("expires_at < ?", time.Now().UTC()).Delete(&entity.Invite{}).Error
}

func (r *inviteRepository) FindAll(ctx context.Context) ([]*entity.Invite, error) {
	var invites []*entity.Invite
	if err := r.database.WithContext(ctx).Order("created_at DESC").Find(&invites).Error; err != nil {
		return nil, err
	}
	return invites, nil
}
