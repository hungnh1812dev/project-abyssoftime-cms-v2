package user

import (
	"context"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

type UseCase struct {
	userRepo repository.UserRepository
	roleRepo repository.RoleRepository
}

func New(userRepo repository.UserRepository, roleRepo repository.RoleRepository) *UseCase {
	return &UseCase{userRepo: userRepo, roleRepo: roleRepo}
}

func (uc *UseCase) List(ctx context.Context, page, limit int) ([]*entity.User, int64, error) {
	return uc.userRepo.FindAll(ctx, page, limit)
}

func (uc *UseCase) GetByID(ctx context.Context, id string) (*entity.User, error) {
	return uc.userRepo.FindByID(ctx, id)
}

func (uc *UseCase) UpdateRole(ctx context.Context, actorID, targetID, newRoleID string) error {
	if actorID == targetID {
		return pkgerrors.ErrForbidden
	}

	actor, err := uc.userRepo.FindByID(ctx, actorID)
	if err != nil {
		return err
	}
	target, err := uc.userRepo.FindByID(ctx, targetID)
	if err != nil {
		return err
	}

	actorLevel, err := uc.resolveLevel(ctx, actor)
	if err != nil {
		return err
	}
	targetLevel, err := uc.resolveLevel(ctx, target)
	if err != nil {
		return err
	}
	if actorLevel <= targetLevel {
		return pkgerrors.ErrForbidden
	}

	newRole, err := uc.roleRepo.FindByID(ctx, newRoleID)
	if err != nil {
		return err
	}
	if actorLevel <= newRole.Level {
		return pkgerrors.ErrForbidden
	}

	target.RoleID = newRoleID
	target.Role = entity.Role(newRole.Slug)
	return uc.userRepo.Update(ctx, target)
}

func (uc *UseCase) Delete(ctx context.Context, actorID, targetID string) error {
	if actorID == targetID {
		return pkgerrors.ErrForbidden
	}

	actor, err := uc.userRepo.FindByID(ctx, actorID)
	if err != nil {
		return err
	}
	target, err := uc.userRepo.FindByID(ctx, targetID)
	if err != nil {
		return err
	}

	actorLevel, err := uc.resolveLevel(ctx, actor)
	if err != nil {
		return err
	}
	targetLevel, err := uc.resolveLevel(ctx, target)
	if err != nil {
		return err
	}
	if actorLevel <= targetLevel {
		return pkgerrors.ErrForbidden
	}

	return uc.userRepo.Delete(ctx, targetID)
}

func (uc *UseCase) resolveLevel(ctx context.Context, u *entity.User) (int, error) {
	if u.RoleID != "" {
		r, err := uc.roleRepo.FindByID(ctx, u.RoleID)
		if err != nil {
			return 0, err
		}
		return r.Level, nil
	}
	return entity.RoleLevel(u.Role), nil
}
