package invite_test

import (
	"context"
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	repomock "project-abyssoftime-cms-v2/api/internal/domain/repository/mock"
	"project-abyssoftime-cms-v2/api/internal/usecase/invite"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

func TestCreate_HierarchyEnforcement(t *testing.T) {
	tests := []struct {
		name      string
		actorRole entity.Role
		invRole   entity.Role
		wantErr   bool
	}{
		{"super_admin invites admin", entity.RoleSuperAdmin, entity.RoleAdmin, false},
		{"admin invites editor", entity.RoleAdmin, entity.RoleEditor, false},
		{"admin cannot invite admin", entity.RoleAdmin, entity.RoleAdmin, true},
		{"admin cannot invite super_admin", entity.RoleAdmin, entity.RoleSuperAdmin, true},
		{"editor cannot invite editor (equal to own role)", entity.RoleEditor, entity.RoleEditor, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &repomock.UserRepository{
				FindByIDFn: func(_ context.Context, _ string) (*entity.User, error) {
					return &entity.User{ID: "actor", Role: tt.actorRole}, nil
				},
				FindByEmailFn: func(_ context.Context, _ string) (*entity.User, error) {
					return nil, pkgerrors.ErrNotFound
				},
				CreateFn: func(_ context.Context, u *entity.User) error {
					u.ID = "new-user"
					return nil
				},
			}
			inviteRepo := &repomock.InviteRepository{
				FindByEmailFn: func(_ context.Context, _ string) (*entity.Invite, error) {
					return nil, pkgerrors.ErrNotFound
				},
				CreateFn: func(_ context.Context, inv *entity.Invite) error {
					inv.ID = "inv-1"
					return nil
				},
			}

			uc := invite.New(inviteRepo, userRepo)
			_, _, err := uc.Create(context.Background(), "actor", "new@test.com", tt.invRole)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestCreate_DuplicateEmailReplacesInvite(t *testing.T) {
	var deletedID string
	userRepo := &repomock.UserRepository{
		FindByIDFn: func(_ context.Context, _ string) (*entity.User, error) {
			return &entity.User{ID: "sa", Role: entity.RoleSuperAdmin}, nil
		},
		FindByEmailFn: func(_ context.Context, _ string) (*entity.User, error) {
			return nil, pkgerrors.ErrNotFound
		},
	}
	inviteRepo := &repomock.InviteRepository{
		FindByEmailFn: func(_ context.Context, _ string) (*entity.Invite, error) {
			return &entity.Invite{ID: "old-invite"}, nil
		},
		DeleteFn: func(_ context.Context, id string) error {
			deletedID = id
			return nil
		},
		CreateFn: func(_ context.Context, inv *entity.Invite) error {
			inv.ID = "new-invite"
			return nil
		},
	}

	uc := invite.New(inviteRepo, userRepo)
	inv, plaintext, err := uc.Create(context.Background(), "sa", "dup@test.com", entity.RoleEditor)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deletedID != "old-invite" {
		t.Errorf("old invite not deleted, got %q", deletedID)
	}
	if inv == nil || plaintext == "" {
		t.Error("expected non-nil invite and non-empty plaintext")
	}
}

func TestCreate_UserAlreadyRegistered(t *testing.T) {
	userRepo := &repomock.UserRepository{
		FindByIDFn: func(_ context.Context, _ string) (*entity.User, error) {
			return &entity.User{ID: "sa", Role: entity.RoleSuperAdmin}, nil
		},
		FindByEmailFn: func(_ context.Context, _ string) (*entity.User, error) {
			return &entity.User{ID: "existing"}, nil
		},
	}
	inviteRepo := &repomock.InviteRepository{}

	uc := invite.New(inviteRepo, userRepo)
	_, _, err := uc.Create(context.Background(), "sa", "exists@test.com", entity.RoleEditor)
	if !pkgerrors.Is(err, pkgerrors.ErrConflict) {
		t.Errorf("err = %v, want ErrConflict", err)
	}
}

func TestAccept_Success(t *testing.T) {
	plaintext := "test-token-plaintext"
	hash := sha256.Sum256([]byte(plaintext))
	tokenHash := fmt.Sprintf("%x", hash)

	var createdUser *entity.User
	var deletedInviteID string

	inviteRepo := &repomock.InviteRepository{
		FindByHashFn: func(_ context.Context, h string) (*entity.Invite, error) {
			if h != tokenHash {
				return nil, pkgerrors.ErrNotFound
			}
			return &entity.Invite{
				ID:        "inv-1",
				Email:     "new@test.com",
				Role:      entity.RoleEditor,
				TokenHash: tokenHash,
				ExpiresAt: time.Now().UTC().Add(time.Hour),
			}, nil
		},
		DeleteFn: func(_ context.Context, id string) error {
			deletedInviteID = id
			return nil
		},
	}
	userRepo := &repomock.UserRepository{
		FindByEmailFn: func(_ context.Context, _ string) (*entity.User, error) {
			return nil, pkgerrors.ErrNotFound
		},
		CreateFn: func(_ context.Context, u *entity.User) error {
			u.ID = "new-user-id"
			createdUser = u
			return nil
		},
	}

	uc := invite.New(inviteRepo, userRepo)
	user, err := uc.Accept(context.Background(), plaintext, "password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email != "new@test.com" {
		t.Errorf("email = %q, want %q", user.Email, "new@test.com")
	}
	if user.Role != entity.RoleEditor {
		t.Errorf("role = %q, want %q", user.Role, entity.RoleEditor)
	}
	if createdUser.PasswordHash == "password123" {
		t.Error("password must be hashed")
	}
	if deletedInviteID != "inv-1" {
		t.Errorf("invite not deleted, got %q", deletedInviteID)
	}
}

func TestAccept_ExpiredToken(t *testing.T) {
	plaintext := "expired-token"
	hash := sha256.Sum256([]byte(plaintext))
	tokenHash := fmt.Sprintf("%x", hash)

	inviteRepo := &repomock.InviteRepository{
		FindByHashFn: func(_ context.Context, _ string) (*entity.Invite, error) {
			return &entity.Invite{
				ID:        "inv-1",
				TokenHash: tokenHash,
				ExpiresAt: time.Now().UTC().Add(-time.Hour),
			}, nil
		},
	}
	userRepo := &repomock.UserRepository{}

	uc := invite.New(inviteRepo, userRepo)
	_, err := uc.Accept(context.Background(), plaintext, "password123")
	if !pkgerrors.Is(err, pkgerrors.ErrValidation) {
		t.Errorf("err = %v, want ErrValidation", err)
	}
}

func TestAccept_InvalidToken(t *testing.T) {
	inviteRepo := &repomock.InviteRepository{
		FindByHashFn: func(_ context.Context, _ string) (*entity.Invite, error) {
			return nil, pkgerrors.ErrNotFound
		},
	}
	userRepo := &repomock.UserRepository{}

	uc := invite.New(inviteRepo, userRepo)
	_, err := uc.Accept(context.Background(), "invalid-token", "password123")
	if !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}
