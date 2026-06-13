package auth_test

import (
	"context"
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	repomock "project-abyssoftime-cms-v2/api/internal/domain/repository/mock"
	"project-abyssoftime-cms-v2/api/internal/usecase/auth"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
	pkgjwt "project-abyssoftime-cms-v2/api/pkg/jwt"
)

// ---- helpers ---------------------------------------------------------------

// mustHash generates a bcrypt hash; panics on error (safe for test init).
func mustHash(password string) string {
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		panic(err)
	}
	return string(h)
}

func validRefreshToken(t *testing.T, userID string) string {
	t.Helper()
	tok, err := pkgjwt.GenerateRefreshToken(userID)
	if err != nil {
		t.Fatalf("GenerateRefreshToken: %v", err)
	}
	return tok
}

// ---- Register --------------------------------------------------------------

func TestRegister(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		email      string
		password   string
		setupRepo  func(*repomock.UserRepository)
		wantErr    error
		wantUserID bool
	}{
		{
			name:     "success — creates user and returns it",
			email:    "new@example.com",
			password: "secret123",
			setupRepo: func(r *repomock.UserRepository) {
				r.FindByEmailFn = func(_ context.Context, _ string) (*entity.User, error) {
					return nil, pkgerrors.ErrNotFound
				}
				r.CreateFn = func(_ context.Context, u *entity.User) error {
					u.ID = "gen-id-1"
					return nil
				}
			},
			wantUserID: true,
		},
		{
			name:     "conflict — email already registered",
			email:    "dup@example.com",
			password: "secret123",
			setupRepo: func(r *repomock.UserRepository) {
				r.FindByEmailFn = func(_ context.Context, _ string) (*entity.User, error) {
					return &entity.User{ID: "existing"}, nil
				}
			},
			wantErr: pkgerrors.ErrConflict,
		},
		{
			name:     "repo FindByEmail returns unexpected error",
			email:    "x@example.com",
			password: "secret123",
			setupRepo: func(r *repomock.UserRepository) {
				r.FindByEmailFn = func(_ context.Context, _ string) (*entity.User, error) {
					return nil, errors.New("db down")
				}
			},
			wantErr: errors.New("db down"),
		},
		{
			name:     "repo Create returns error",
			email:    "y@example.com",
			password: "secret123",
			setupRepo: func(r *repomock.UserRepository) {
				r.FindByEmailFn = func(_ context.Context, _ string) (*entity.User, error) {
					return nil, pkgerrors.ErrNotFound
				}
				r.CreateFn = func(_ context.Context, _ *entity.User) error {
					return errors.New("insert failed")
				}
			},
			wantErr: errors.New("insert failed"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &repomock.UserRepository{}
			tc.setupRepo(repo)

			uc := auth.New(repo)
			user, err := uc.Register(ctx, tc.email, tc.password)

			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantUserID && user == nil {
				t.Fatal("expected non-nil user")
			}
			// password must be hashed, not stored in plain text
			if user != nil && user.PasswordHash == tc.password {
				t.Error("PasswordHash must not equal plain-text password")
			}
		})
	}
}

// ---- Login -----------------------------------------------------------------

func TestLogin(t *testing.T) {
	ctx := context.Background()
	const plainPass = "correct-horse"

	tests := []struct {
		name      string
		password  string
		setupRepo func(*repomock.UserRepository)
		wantErr   error
	}{
		{
			name:     "success — returns access and refresh tokens",
			password: plainPass,
			setupRepo: func(r *repomock.UserRepository) {
				r.FindByEmailFn = func(_ context.Context, _ string) (*entity.User, error) {
					return &entity.User{
						ID:           "u1",
						Role:         entity.RoleAdmin,
						PasswordHash: mustHash(plainPass),
					}, nil
				}
			},
		},
		{
			name:     "user not found → ErrUnauthorized",
			password: plainPass,
			setupRepo: func(r *repomock.UserRepository) {
				r.FindByEmailFn = func(_ context.Context, _ string) (*entity.User, error) {
					return nil, pkgerrors.ErrNotFound
				}
			},
			wantErr: pkgerrors.ErrUnauthorized,
		},
		{
			name:     "repo returns unexpected error",
			password: plainPass,
			setupRepo: func(r *repomock.UserRepository) {
				r.FindByEmailFn = func(_ context.Context, _ string) (*entity.User, error) {
					return nil, errors.New("db error")
				}
			},
			wantErr: errors.New("db error"),
		},
		{
			name:     "wrong password → ErrUnauthorized",
			password: "wrong-password",
			setupRepo: func(r *repomock.UserRepository) {
				r.FindByEmailFn = func(_ context.Context, _ string) (*entity.User, error) {
					return &entity.User{
						ID:           "u1",
						Role:         entity.RoleAdmin,
						PasswordHash: mustHash(plainPass),
					}, nil
				}
			},
			wantErr: pkgerrors.ErrUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &repomock.UserRepository{}
			tc.setupRepo(repo)

			uc := auth.New(repo)
			access, refresh, err := uc.Login(ctx, "user@example.com", tc.password)

			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if access == "" {
				t.Error("access token must not be empty")
			}
			if refresh == "" {
				t.Error("refresh token must not be empty")
			}
		})
	}
}

// ---- RefreshToken ----------------------------------------------------------

func TestRefreshToken(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		buildToken   func() string
		setupRepo    func(*repomock.UserRepository)
		wantErr      error
		wantNonEmpty bool
	}{
		{
			name: "success — returns new access token",
			buildToken: func() string { return validRefreshToken(t, "u1") },
			setupRepo: func(r *repomock.UserRepository) {
				r.FindByIDFn = func(_ context.Context, id string) (*entity.User, error) {
					return &entity.User{ID: id, Role: entity.RoleAdmin}, nil
				}
			},
			wantNonEmpty: true,
		},
		{
			name:       "invalid token → ErrUnauthorized",
			buildToken: func() string { return "not.a.valid.token" },
			setupRepo:  func(r *repomock.UserRepository) {},
			wantErr:    pkgerrors.ErrUnauthorized,
		},
		{
			name: "user no longer exists → ErrUnauthorized",
			buildToken: func() string { return validRefreshToken(t, "deleted-user") },
			setupRepo: func(r *repomock.UserRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.User, error) {
					return nil, pkgerrors.ErrNotFound
				}
			},
			wantErr: pkgerrors.ErrUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &repomock.UserRepository{}
			tc.setupRepo(repo)

			uc := auth.New(repo)
			token, err := uc.RefreshToken(ctx, tc.buildToken())

			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantNonEmpty && token == "" {
				t.Error("expected non-empty access token")
			}
		})
	}
}

// ---- Logout ----------------------------------------------------------------

func TestLogout(t *testing.T) {
	ctx := context.Background()
	repo := &repomock.UserRepository{}
	uc := auth.New(repo)

	if err := uc.Logout(ctx, "any-user-id"); err != nil {
		t.Fatalf("Logout returned unexpected error: %v", err)
	}
}
