package gormdb

import (
	"context"
	"testing"
	"time"

	"gorm.io/gorm"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

func setupRoleDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := NewClient("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if err := db.AutoMigrate(&entity.RoleEntity{}); err != nil {
		t.Fatalf("AutoMigrate: %v", err)
	}
	return db
}

func TestRoleRepository_Create_And_FindByID(t *testing.T) {
	db := setupRoleDB(t)
	repo := NewRoleRepository(db)
	ctx := context.Background()

	role := &entity.RoleEntity{
		ID:          "r1",
		DocumentID:  "rdoc1",
		Name:        "Editor",
		Slug:        "editor",
		Permissions: []string{"content:read", "content:create"},
		Level:       60,
		IsDefault:   true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := repo.Create(ctx, role); err != nil {
		t.Fatalf("Create: %v", err)
	}

	found, err := repo.FindByID(ctx, "rdoc1")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.Name != "Editor" {
		t.Errorf("Name = %q, want %q", found.Name, "Editor")
	}
	if found.Slug != "editor" {
		t.Errorf("Slug = %q, want %q", found.Slug, "editor")
	}
	if len(found.Permissions) != 2 {
		t.Errorf("len(Permissions) = %d, want 2", len(found.Permissions))
	}
}

func TestRoleRepository_FindByID_NotFound(t *testing.T) {
	db := setupRoleDB(t)
	repo := NewRoleRepository(db)

	_, err := repo.FindByID(context.Background(), "nonexistent")
	if !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestRoleRepository_FindBySlug(t *testing.T) {
	db := setupRoleDB(t)
	repo := NewRoleRepository(db)
	ctx := context.Background()

	_ = repo.Create(ctx, &entity.RoleEntity{
		ID: "r1", DocumentID: "rdoc1", Name: "Guest", Slug: "guest",
		Permissions: []string{"content:read"}, Level: 20,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	})

	found, err := repo.FindBySlug(ctx, "guest")
	if err != nil {
		t.Fatalf("FindBySlug: %v", err)
	}
	if found.DocumentID != "rdoc1" {
		t.Errorf("DocumentID = %q, want %q", found.DocumentID, "rdoc1")
	}
}

func TestRoleRepository_FindBySlug_NotFound(t *testing.T) {
	db := setupRoleDB(t)
	repo := NewRoleRepository(db)

	_, err := repo.FindBySlug(context.Background(), "nonexistent")
	if !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestRoleRepository_FindAll(t *testing.T) {
	db := setupRoleDB(t)
	repo := NewRoleRepository(db)
	ctx := context.Background()

	now := time.Now()
	_ = repo.Create(ctx, &entity.RoleEntity{ID: "r1", DocumentID: "d1", Name: "Admin", Slug: "admin", Permissions: []string{"content:read"}, Level: 80, CreatedAt: now, UpdatedAt: now})
	_ = repo.Create(ctx, &entity.RoleEntity{ID: "r2", DocumentID: "d2", Name: "Guest", Slug: "guest", Permissions: []string{"content:read"}, Level: 20, CreatedAt: now, UpdatedAt: now})

	roles, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll: %v", err)
	}
	if len(roles) != 2 {
		t.Errorf("len(roles) = %d, want 2", len(roles))
	}
	if roles[0].Level < roles[1].Level {
		t.Error("expected roles sorted by level descending")
	}
}

func TestRoleRepository_Update(t *testing.T) {
	db := setupRoleDB(t)
	repo := NewRoleRepository(db)
	ctx := context.Background()

	now := time.Now()
	_ = repo.Create(ctx, &entity.RoleEntity{ID: "r1", DocumentID: "d1", Name: "Editor", Slug: "editor", Permissions: []string{"content:read"}, Level: 60, CreatedAt: now, UpdatedAt: now})

	role, _ := repo.FindByID(ctx, "d1")
	role.Permissions = []string{"content:read", "content:create", "content:update"}
	role.UpdatedAt = time.Now()
	if err := repo.Update(ctx, role); err != nil {
		t.Fatalf("Update: %v", err)
	}

	updated, _ := repo.FindByID(ctx, "d1")
	if len(updated.Permissions) != 3 {
		t.Errorf("len(Permissions) = %d, want 3", len(updated.Permissions))
	}
}

func TestRoleRepository_Delete(t *testing.T) {
	db := setupRoleDB(t)
	repo := NewRoleRepository(db)
	ctx := context.Background()

	now := time.Now()
	_ = repo.Create(ctx, &entity.RoleEntity{ID: "r1", DocumentID: "d1", Name: "Custom", Slug: "custom", Permissions: []string{"content:read"}, Level: 50, CreatedAt: now, UpdatedAt: now})

	if err := repo.Delete(ctx, "d1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := repo.FindByID(ctx, "d1")
	if !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestRoleRepository_Delete_NotFound(t *testing.T) {
	db := setupRoleDB(t)
	repo := NewRoleRepository(db)

	err := repo.Delete(context.Background(), "nonexistent")
	if !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestRoleRepository_HasAny(t *testing.T) {
	db := setupRoleDB(t)
	repo := NewRoleRepository(db)
	ctx := context.Background()

	has, err := repo.HasAny(ctx)
	if err != nil {
		t.Fatalf("HasAny: %v", err)
	}
	if has {
		t.Error("expected false when no roles exist")
	}

	now := time.Now()
	_ = repo.Create(ctx, &entity.RoleEntity{ID: "r1", DocumentID: "d1", Name: "Admin", Slug: "admin", Permissions: []string{}, Level: 80, CreatedAt: now, UpdatedAt: now})

	has, err = repo.HasAny(ctx)
	if err != nil {
		t.Fatalf("HasAny: %v", err)
	}
	if !has {
		t.Error("expected true when roles exist")
	}
}
