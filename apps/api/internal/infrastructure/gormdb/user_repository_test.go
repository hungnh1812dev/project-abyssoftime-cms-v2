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
		DocumentID: "doc1",
		Email:        "test@example.com",
		PasswordHash: "hash",
		Role:         entity.RoleAdmin,
	}
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Create: %v", err)
	}

	found, err := repo.FindByID(ctx, "doc1")
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

	user := &entity.User{DocumentID: "u1", Email: "a@b.com", Role: entity.RoleGuest}
	_ = repo.Create(ctx, user)

	found, err := repo.FindByEmail(ctx, "a@b.com")
	if err != nil {
		t.Fatalf("FindByEmail: %v", err)
	}
	if found.DocumentID != "u1" {
		t.Errorf("DocumentID = %q, want %q", found.DocumentID, "u1")
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

func TestUserRepository_HasSuperAdmin(t *testing.T) {
	db := setupUserDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	has, err := repo.HasSuperAdmin(ctx)
	if err != nil {
		t.Fatalf("HasSuperAdmin: %v", err)
	}
	if has {
		t.Error("expected false when no users exist")
	}

	_ = repo.Create(ctx, &entity.User{DocumentID: "d1", Email: "admin@x.com", Role: entity.RoleAdmin})

	has, err = repo.HasSuperAdmin(ctx)
	if err != nil {
		t.Fatalf("HasSuperAdmin: %v", err)
	}
	if has {
		t.Error("expected false when only admin exists (no super_admin)")
	}

	_ = repo.Create(ctx, &entity.User{DocumentID: "d2", Email: "sa@x.com", Role: entity.RoleSuperAdmin})

	has, err = repo.HasSuperAdmin(ctx)
	if err != nil {
		t.Fatalf("HasSuperAdmin: %v", err)
	}
	if !has {
		t.Error("expected true when super_admin exists")
	}
}

func TestUserRepository_FindAll(t *testing.T) {
	db := setupUserDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	_ = repo.Create(ctx, &entity.User{DocumentID: "d1", Email: "a@x.com", Role: entity.RoleGuest})
	_ = repo.Create(ctx, &entity.User{DocumentID: "d2", Email: "b@x.com", Role: entity.RoleAdmin})
	_ = repo.Create(ctx, &entity.User{DocumentID: "d3", Email: "c@x.com", Role: entity.RoleEditor})

	users, total, err := repo.FindAll(ctx, 1, 2)
	if err != nil {
		t.Fatalf("FindAll: %v", err)
	}
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
	if len(users) != 2 {
		t.Errorf("len(users) = %d, want 2", len(users))
	}
}

func TestUserRepository_Update(t *testing.T) {
	db := setupUserDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	_ = repo.Create(ctx, &entity.User{DocumentID: "d1", Email: "a@x.com", Role: entity.RoleGuest})

	user, _ := repo.FindByID(ctx, "d1")
	user.Role = entity.RoleEditor
	if err := repo.Update(ctx, user); err != nil {
		t.Fatalf("Update: %v", err)
	}

	updated, _ := repo.FindByID(ctx, "d1")
	if updated.Role != entity.RoleEditor {
		t.Errorf("role = %q, want %q", updated.Role, entity.RoleEditor)
	}
}

func TestUserRepository_Delete(t *testing.T) {
	db := setupUserDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	_ = repo.Create(ctx, &entity.User{DocumentID: "d1", Email: "a@x.com", Role: entity.RoleGuest})

	if err := repo.Delete(ctx, "d1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := repo.FindByID(ctx, "d1")
	if !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestUserRepository_Delete_NotFound(t *testing.T) {
	db := setupUserDB(t)
	repo := NewUserRepository(db)

	err := repo.Delete(context.Background(), "nonexistent")
	if !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}
