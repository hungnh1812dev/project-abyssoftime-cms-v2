package invite

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

const inviteExpiry = 7 * 24 * time.Hour

type UseCase struct {
	inviteRepo repository.InviteRepository
	userRepo   repository.UserRepository
}

func New(inviteRepo repository.InviteRepository, userRepo repository.UserRepository) *UseCase {
	return &UseCase{inviteRepo: inviteRepo, userRepo: userRepo}
}

func (uc *UseCase) Create(ctx context.Context, actorID, email string, role entity.Role) (*entity.Invite, string, error) {
	actor, err := uc.userRepo.FindByID(ctx, actorID)
	if err != nil {
		return nil, "", err
	}
	if entity.RoleLevel(actor.Role) <= entity.RoleLevel(role) {
		return nil, "", pkgerrors.ErrForbidden
	}

	existing, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil && !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
		return nil, "", err
	}
	if existing != nil {
		return nil, "", pkgerrors.ErrConflict
	}

	if pending, err := uc.inviteRepo.FindByEmail(ctx, email); err == nil {
		_ = uc.inviteRepo.Delete(ctx, pending.ID)
	}

	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return nil, "", fmt.Errorf("generate token: %w", err)
	}
	plaintext := base64.URLEncoding.EncodeToString(raw)
	hash := sha256.Sum256([]byte(plaintext))
	tokenHash := fmt.Sprintf("%x", hash)

	inv := &entity.Invite{
		Email:     email,
		Role:      role,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().UTC().Add(inviteExpiry),
		CreatedBy: actorID,
	}
	if err := uc.inviteRepo.Create(ctx, inv); err != nil {
		return nil, "", err
	}
	return inv, plaintext, nil
}

func (uc *UseCase) List(ctx context.Context) ([]*entity.Invite, error) {
	return uc.inviteRepo.FindAll(ctx)
}

func (uc *UseCase) Revoke(ctx context.Context, id string) error {
	return uc.inviteRepo.Delete(ctx, id)
}

func (uc *UseCase) Accept(ctx context.Context, token, password string) (*entity.User, error) {
	hash := sha256.Sum256([]byte(token))
	tokenHash := fmt.Sprintf("%x", hash)

	inv, err := uc.inviteRepo.FindByHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}

	if time.Now().UTC().After(inv.ExpiresAt) {
		return nil, pkgerrors.ErrValidation
	}

	existing, err := uc.userRepo.FindByEmail(ctx, inv.Email)
	if err != nil && !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, pkgerrors.ErrConflict
	}

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &entity.User{
		Email:        inv.Email,
		PasswordHash: string(passHash),
		Role:         inv.Role,
	}
	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	_ = uc.inviteRepo.Delete(ctx, inv.ID)
	return user, nil
}
