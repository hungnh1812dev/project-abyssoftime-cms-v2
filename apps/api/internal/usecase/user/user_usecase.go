package user

import (
	"context"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

type UseCase struct {
	userRepo repository.UserRepository
}

func New(userRepo repository.UserRepository) *UseCase {
	return &UseCase{userRepo: userRepo}
}

func (uc *UseCase) List(ctx context.Context, page, limit int) ([]*entity.User, int64, error) {
	return uc.userRepo.FindAll(ctx, page, limit)
}

func (uc *UseCase) GetByID(ctx context.Context, id string) (*entity.User, error) {
	return uc.userRepo.FindByID(ctx, id)
}

func (uc *UseCase) UpdateRole(ctx context.Context, actorID, targetID string, newRole entity.Role) error {
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

	actorLevel := entity.RoleLevel(actor.Role)
	if actorLevel <= entity.RoleLevel(target.Role) {
		return pkgerrors.ErrForbidden
	}
	if actorLevel <= entity.RoleLevel(newRole) {
		return pkgerrors.ErrForbidden
	}

	target.Role = newRole
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

	if entity.RoleLevel(actor.Role) <= entity.RoleLevel(target.Role) {
		return pkgerrors.ErrForbidden
	}

	return uc.userRepo.Delete(ctx, targetID)
}
