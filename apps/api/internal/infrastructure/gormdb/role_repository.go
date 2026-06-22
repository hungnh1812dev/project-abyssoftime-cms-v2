package gormdb

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

var _ repository.RoleRepository = (*roleRepository)(nil)

type roleRepository struct {
	database *gorm.DB
}

func NewRoleRepository(database *gorm.DB) repository.RoleRepository {
	return &roleRepository{database: database}
}

func (r *roleRepository) Create(ctx context.Context, role *entity.RoleEntity) error {
	return r.database.WithContext(ctx).Create(role).Error
}

func (r *roleRepository) FindByID(ctx context.Context, documentID string) (*entity.RoleEntity, error) {
	var role entity.RoleEntity
	if err := r.database.WithContext(ctx).Where("document_id = ?", documentID).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) FindBySlug(ctx context.Context, slug string) (*entity.RoleEntity, error) {
	var role entity.RoleEntity
	if err := r.database.WithContext(ctx).Where("slug = ?", slug).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) FindAll(ctx context.Context) ([]*entity.RoleEntity, error) {
	var roles []*entity.RoleEntity
	if err := r.database.WithContext(ctx).Order("level DESC").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *roleRepository) Update(ctx context.Context, role *entity.RoleEntity) error {
	result := r.database.WithContext(ctx).Save(role)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return pkgerrors.ErrNotFound
	}
	return nil
}

func (r *roleRepository) Delete(ctx context.Context, documentID string) error {
	result := r.database.WithContext(ctx).Where("document_id = ?", documentID).Delete(&entity.RoleEntity{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return pkgerrors.ErrNotFound
	}
	return nil
}

func (r *roleRepository) HasAny(ctx context.Context) (bool, error) {
	var count int64
	if err := r.database.WithContext(ctx).Model(&entity.RoleEntity{}).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
