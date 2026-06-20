package role

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"time"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

var slugRe = regexp.MustCompile(`^[a-z0-9]+(?:[-_][a-z0-9]+)*$`)

type UseCase struct {
	roleRepo repository.RoleRepository
	userRepo repository.UserRepository
}

func New(roleRepo repository.RoleRepository, userRepo repository.UserRepository) *UseCase {
	return &UseCase{roleRepo: roleRepo, userRepo: userRepo}
}

type CreateRoleInput struct {
	Name        string
	Slug        string
	Permissions []string
	Level       int
}

type UpdateRoleInput struct {
	Name        *string
	Permissions *[]string
	Level       *int
}

func generateDocID() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (uc *UseCase) SeedDefaults(ctx context.Context) error {
	has, err := uc.roleRepo.HasAny(ctx)
	if err != nil {
		return err
	}
	if has {
		return nil
	}

	now := time.Now().UTC()
	for _, def := range entity.DefaultRoles {
		r := def
		r.ID = generateDocID()
		r.DocumentID = generateDocID()
		r.CreatedAt = now
		r.UpdatedAt = now
		if err := uc.roleRepo.Create(ctx, &r); err != nil {
			return fmt.Errorf("seed role %s: %w", r.Slug, err)
		}
	}
	return nil
}

func (uc *UseCase) Create(ctx context.Context, input CreateRoleInput, callerLevel int) (*entity.RoleEntity, error) {
	if err := validateSlug(input.Slug); err != nil {
		return nil, err
	}
	if err := validatePermissions(input.Permissions); err != nil {
		return nil, err
	}
	if input.Level < 1 || input.Level > 99 {
		return nil, fmt.Errorf("%w: level must be 1-99", pkgerrors.ErrValidation)
	}
	if input.Level >= callerLevel {
		return nil, fmt.Errorf("%w: cannot create role with level >= your own", pkgerrors.ErrForbidden)
	}
	if input.Name == "" || len(input.Name) > 100 {
		return nil, fmt.Errorf("%w: name must be 1-100 characters", pkgerrors.ErrValidation)
	}

	now := time.Now().UTC()
	role := &entity.RoleEntity{
		ID:          generateDocID(),
		DocumentID:  generateDocID(),
		Name:        input.Name,
		Slug:        input.Slug,
		Permissions: input.Permissions,
		Level:       input.Level,
		IsDefault:   false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := uc.roleRepo.Create(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

func (uc *UseCase) FindAll(ctx context.Context) ([]*entity.RoleEntity, error) {
	return uc.roleRepo.FindAll(ctx)
}

func (uc *UseCase) FindByID(ctx context.Context, documentID string) (*entity.RoleEntity, error) {
	return uc.roleRepo.FindByID(ctx, documentID)
}

func (uc *UseCase) Update(ctx context.Context, documentID string, input UpdateRoleInput, callerLevel int) (*entity.RoleEntity, error) {
	role, err := uc.roleRepo.FindByID(ctx, documentID)
	if err != nil {
		return nil, err
	}

	if role.IsDefault {
		if input.Name != nil || input.Level != nil {
			return nil, fmt.Errorf("%w: cannot change name or level of a default role", pkgerrors.ErrValidation)
		}
	}

	if input.Name != nil {
		if *input.Name == "" || len(*input.Name) > 100 {
			return nil, fmt.Errorf("%w: name must be 1-100 characters", pkgerrors.ErrValidation)
		}
		role.Name = *input.Name
	}
	if input.Permissions != nil {
		if err := validatePermissions(*input.Permissions); err != nil {
			return nil, err
		}
		role.Permissions = *input.Permissions
	}
	if input.Level != nil {
		if *input.Level < 1 || *input.Level > 99 {
			return nil, fmt.Errorf("%w: level must be 1-99", pkgerrors.ErrValidation)
		}
		if *input.Level >= callerLevel {
			return nil, fmt.Errorf("%w: cannot set level >= your own", pkgerrors.ErrForbidden)
		}
		role.Level = *input.Level
	}

	role.UpdatedAt = time.Now().UTC()
	if err := uc.roleRepo.Update(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

func (uc *UseCase) Delete(ctx context.Context, documentID string) error {
	role, err := uc.roleRepo.FindByID(ctx, documentID)
	if err != nil {
		return err
	}

	if role.IsDefault {
		return fmt.Errorf("%w: cannot delete a default role", pkgerrors.ErrValidation)
	}

	users, count, err := uc.userRepo.FindAll(ctx, 1, 1)
	if err != nil {
		return err
	}
	_ = users
	if count > 0 {
		for _, u := range users {
			if u.RoleID == documentID {
				return fmt.Errorf("%w: role is assigned to users", pkgerrors.ErrConflict)
			}
		}
		// Need to check all users, not just first page
		allUsers, _, err := uc.userRepo.FindAll(ctx, 1, 10000)
		if err != nil {
			return err
		}
		for _, u := range allUsers {
			if u.RoleID == documentID {
				return fmt.Errorf("%w: role is assigned to users", pkgerrors.ErrConflict)
			}
		}
	}

	return uc.roleRepo.Delete(ctx, documentID)
}

func validateSlug(slug string) error {
	if len(slug) == 0 || len(slug) > 63 || !slugRe.MatchString(slug) {
		return fmt.Errorf("%w: slug must be 1-63 lowercase alphanumeric chars, hyphens or underscores", pkgerrors.ErrValidation)
	}
	return nil
}

func validatePermissions(perms []string) error {
	for _, p := range perms {
		if !entity.IsValidPermission(p) {
			return fmt.Errorf("%w: unknown permission %q", pkgerrors.ErrValidation, p)
		}
	}
	return nil
}
