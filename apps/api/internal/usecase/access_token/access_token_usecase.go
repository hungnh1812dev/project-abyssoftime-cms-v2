package access_token

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

type UseCase struct {
	repo repository.AccessTokenRepository
}

func New(repo repository.AccessTokenRepository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) Create(ctx context.Context, name string, scopes []string, expiresAt *time.Time, createdBy string) (*entity.AccessToken, string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return nil, "", fmt.Errorf("generate token: %w", err)
	}
	plaintext := hex.EncodeToString(raw)
	hash := sha256.Sum256([]byte(plaintext))
	tokenHash := hex.EncodeToString(hash[:])

	token := &entity.AccessToken{
		DocumentID: uuid.NewString(),
		Name:      name,
		TokenHash: tokenHash,
		Prefix:    plaintext[:6],
		Scopes:    scopes,
		ExpiresAt: expiresAt,
		CreatedBy: createdBy,
	}
	if err := uc.repo.Create(ctx, token); err != nil {
		return nil, "", err
	}
	return token, plaintext, nil
}

func (uc *UseCase) List(ctx context.Context, page, limit int) ([]*entity.AccessToken, int64, error) {
	return uc.repo.FindAll(ctx, page, limit)
}

func (uc *UseCase) Delete(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}

func (uc *UseCase) Validate(ctx context.Context, rawToken string) (*entity.AccessToken, error) {
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	token, err := uc.repo.FindByHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}

	if token.ExpiresAt != nil && time.Now().UTC().After(*token.ExpiresAt) {
		return nil, pkgerrors.ErrUnauthorized
	}

	err = uc.repo.UpdateLastUsed(ctx, token.DocumentID, time.Now().UTC())

	return token, nil
}
