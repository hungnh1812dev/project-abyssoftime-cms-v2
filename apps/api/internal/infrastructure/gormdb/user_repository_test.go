package gormdb

import (
	"context"
	"testing"

	"gorm.io/gorm"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

func setupUserDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := NewClient("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if err := db.AutoMigrate(&entity.User{}); err != nil {
		t.Fatalf("AutoMigrate: %v", err)
	}
	return db
}

func TestUserRepository_Create_And_FindByID(t *testing.T) {
	db := setupUserDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &entity.User{
		ID:           "u1",
		DocumentID:   "doc1",
		Email:        "test@example.com",
		PasswordHash: "hash",
		Role:         entity.RoleAdmin,
	}
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Create: %v", err)
	}

	found, err := repo.FindByID(ctx, "u1")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.Email != "test@example.com" {
		t.Errorf("Email = %q, want %q", found.Email, "test@example.com")
	}
}

func TestUserRepository_FindByEmail(t *testing.T) {
	db := setupUserDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &entity.User{ID: "u1", Email: "a@b.com", Role: entity.RoleGuest}
	_ = repo.Create(ctx, user)

	found, err := repo.FindByEmail(ctx, "a@b.com")
	if err != nil {
		t.Fatalf("FindByEmail: %v", err)
	}
	if found.ID != "u1" {
		t.Errorf("ID = %q, want %q", found.ID, "u1")
	}
}

func TestUserRepository_FindByID_NotFound(t *testing.T) {
	db := setupUserDB(t)
	repo := NewUserRepository(db)

	_, err := repo.FindByID(context.Background(), "nonexistent")
	if !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestUserRepository_FindByEmail_NotFound(t *testing.T) {
	db := setupUserDB(t)
	repo := NewUserRepository(db)

	_, err := repo.FindByEmail(context.Background(), "nope@x.com")
	if !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestUserRepository_CountAdmins(t *testing.T) {
	db := setupUserDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	count, err := repo.CountAdmins(ctx)
	if err != nil {
		t.Fatalf("CountAdmins: %v", err)
	}
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}

	_ = repo.Create(ctx, &entity.User{ID: "u1", Email: "admin@x.com", Role: entity.RoleAdmin})
	_ = repo.Create(ctx, &entity.User{ID: "u2", Email: "guest@x.com", Role: entity.RoleGuest})

	count, err = repo.CountAdmins(ctx)
	if err != nil {
		t.Fatalf("CountAdmins: %v", err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
}
