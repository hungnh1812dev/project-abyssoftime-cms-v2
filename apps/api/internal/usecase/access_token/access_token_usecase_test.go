package access_token_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	repomock "project-abyssoftime-cms-v2/api/internal/domain/repository/mock"
	accesstoken "project-abyssoftime-cms-v2/api/internal/usecase/access_token"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

func TestCreate(t *testing.T) {
	var stored *entity.AccessToken
	repo := &repomock.AccessTokenRepository{
		CreateFn: func(_ context.Context, tok *entity.AccessToken) error {
			tok.DocumentID = "tok-1"
			stored = tok
			return nil
		},
	}

	uc := accesstoken.New(repo)
	tok, plaintext, err := uc.Create(context.Background(), "My Token", []string{"documents:read"}, nil, "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plaintext == "" {
		t.Error("plaintext must not be empty")
	}
	if len(tok.Prefix) != 6 {
		t.Errorf("prefix length = %d, want 6", len(tok.Prefix))
	}
	if tok.TokenHash == plaintext {
		t.Error("stored hash must not equal plaintext")
	}

	hash := sha256.Sum256([]byte(plaintext))
	expectedHash := hex.EncodeToString(hash[:])
	if stored.TokenHash != expectedHash {
		t.Errorf("stored hash = %q, want SHA-256 of plaintext", stored.TokenHash)
	}
}

func TestValidate_Success(t *testing.T) {
	plaintext := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	hash := sha256.Sum256([]byte(plaintext))
	tokenHash := hex.EncodeToString(hash[:])

	var updatedID string
	repo := &repomock.AccessTokenRepository{
		FindByHashFn: func(_ context.Context, h string) (*entity.AccessToken, error) {
			if h != tokenHash {
				return nil, pkgerrors.ErrNotFound
			}
			return &entity.AccessToken{
				DocumentID: "tok-1",
				TokenHash: tokenHash,
				Scopes:    []string{"documents:read"},
			}, nil
		},
		UpdateLastUsedFn: func(_ context.Context, id string, _ time.Time) error {
			updatedID = id
			return nil
		},
	}

	uc := accesstoken.New(repo)
	tok, err := uc.Validate(context.Background(), plaintext)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok.DocumentID != "tok-1" {
		t.Errorf("ID = %q, want %q", tok.DocumentID, "tok-1")
	}
	if updatedID != "tok-1" {
		t.Errorf("lastUsedAt not updated, got ID %q", updatedID)
	}
}

func TestValidate_Expired(t *testing.T) {
	plaintext := "expired-token-value"
	hash := sha256.Sum256([]byte(plaintext))
	tokenHash := hex.EncodeToString(hash[:])
	expired := time.Now().UTC().Add(-time.Hour)

	repo := &repomock.AccessTokenRepository{
		FindByHashFn: func(_ context.Context, _ string) (*entity.AccessToken, error) {
			return &entity.AccessToken{
				DocumentID: "tok-1",
				TokenHash: tokenHash,
				ExpiresAt: &expired,
			}, nil
		},
	}

	uc := accesstoken.New(repo)
	_, err := uc.Validate(context.Background(), plaintext)
	if !pkgerrors.Is(err, pkgerrors.ErrUnauthorized) {
		t.Errorf("err = %v, want ErrUnauthorized", err)
	}
}

func TestValidate_NotFound(t *testing.T) {
	repo := &repomock.AccessTokenRepository{
		FindByHashFn: func(_ context.Context, _ string) (*entity.AccessToken, error) {
			return nil, pkgerrors.ErrNotFound
		},
	}

	uc := accesstoken.New(repo)
	_, err := uc.Validate(context.Background(), "unknown-token")
	if !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}
