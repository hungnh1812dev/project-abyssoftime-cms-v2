package auth

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
	pkgjwt "project-abyssoftime-cms-v2/api/pkg/jwt"
)

type UseCase struct {
	repo repository.UserRepository
}

func New(repo repository.UserRepository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) Register(ctx context.Context, email, password string) (*entity.User, error) {
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

	user := &entity.User{
		Email:        email,
		PasswordHash: string(hash),
		Role:         entity.RoleSuperAdmin,
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

	access, err := pkgjwt.GenerateAccessToken(user.ID, string(user.Role))
	if err != nil {
		return "", "", err
	}
	refresh, err := pkgjwt.GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", err
	}
	return access, refresh, nil
}

func (uc *UseCase) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	claims, err := pkgjwt.ValidateToken(refreshToken)
	if err != nil {
		return "", pkgerrors.ErrUnauthorized
	}

	user, err := uc.repo.FindByID(ctx, claims.UserID)
	if err != nil {
		return "", pkgerrors.ErrUnauthorized
	}

	return pkgjwt.GenerateAccessToken(user.ID, string(user.Role))
}

func (uc *UseCase) Logout(_ context.Context, _ string) error {
	// Stateless JWT: session clearing is handled by the HTTP handler (cookie removal).
	return nil
}
