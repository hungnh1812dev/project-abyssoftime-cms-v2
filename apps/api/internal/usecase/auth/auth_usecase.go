package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
	pkgjwt "project-abyssoftime-cms-v2/api/pkg/jwt"
)

type UseCase struct {
	repo     repository.UserRepository
	roleRepo repository.RoleRepository
}

func New(repo repository.UserRepository, roleRepo repository.RoleRepository) *UseCase {
	return &UseCase{repo: repo, roleRepo: roleRepo}
}

func (uc *UseCase) Register(ctx context.Context, email, password, displayName string) (*entity.User, error) {
	if displayName == "" || len(displayName) > 100 {
		return nil, pkgerrors.ErrValidation
	}

	hasSA, err := uc.repo.HasSuperAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if hasSA {
		return nil, pkgerrors.ErrForbidden
	}

	existing, err := uc.repo.FindByEmail(ctx, email)
	if err != nil && !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, pkgerrors.ErrConflict
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	saRole, err := uc.roleRepo.FindBySlug(ctx, "super_admin")
	if err != nil {
		return nil, fmt.Errorf("find super_admin role: %w", err)
	}

	user := &entity.User{
		ID:           uuid.New().String(),
		DocumentID:   uuid.New().String(),
		Email:        email,
		DisplayName:  displayName,
		PasswordHash: string(hash),
		Role:         entity.RoleSuperAdmin,
		RoleID:       saRole.DocumentID,
		CreatedAt:    time.Now().UTC(),
	}
	if err := uc.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (uc *UseCase) SetupStatus(ctx context.Context) (bool, error) {
	return uc.repo.HasSuperAdmin(ctx)
}

func (uc *UseCase) Login(ctx context.Context, email, password string) (accessToken, refreshToken string, err error) {
	user, err := uc.repo.FindByEmail(ctx, email)
	if pkgerrors.Is(err, pkgerrors.ErrNotFound) {
		return "", "", pkgerrors.ErrUnauthorized
	}
	if err != nil {
		return "", "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", "", pkgerrors.ErrUnauthorized
	}

	roleSlug, err := uc.resolveRoleSlug(ctx, user)
	if err != nil {
		return "", "", err
	}

	access, err := pkgjwt.GenerateAccessToken(user.ID, roleSlug)
	if err != nil {
		return "", "", err
	}
	refresh, err := pkgjwt.GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", err
	}
	return access, refresh, nil
}

func (uc *UseCase) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	claims, err := pkgjwt.ValidateToken(refreshToken)
	if err != nil {
		return "", "", pkgerrors.ErrUnauthorized
	}

	user, err := uc.repo.FindByID(ctx, claims.UserID)
	if err != nil {
		return "", "", pkgerrors.ErrUnauthorized
	}

	roleSlug, err := uc.resolveRoleSlug(ctx, user)
	if err != nil {
		return "", "", pkgerrors.ErrUnauthorized
	}

	access, err := pkgjwt.GenerateAccessToken(user.ID, roleSlug)
	if err != nil {
		return "", "", err
	}

	newRefresh, err := pkgjwt.GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", err
	}

	return access, newRefresh, nil
}

func (uc *UseCase) resolveRoleSlug(ctx context.Context, user *entity.User) (string, error) {
	if user.RoleID != "" {
		r, err := uc.roleRepo.FindByID(ctx, user.RoleID)
		if err != nil {
			return "", err
		}
		return r.Slug, nil
	}
	// Fallback for legacy users not yet migrated
	return string(user.Role), nil
}

func (uc *UseCase) Logout(_ context.Context, _ string) error {
	// Stateless JWT: session clearing is handled by the HTTP handler (cookie removal).
	return nil
}
