package gormdb

import (
	"context"
	"testing"
	"time"

	"gorm.io/gorm"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

func setupLocaleDB(t *testing.T) *gorm.DB {
	t.Helper()
	database, err := NewClient("sqlite", ":memory:", false)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if err := database.AutoMigrate(&entity.Locale{}); err != nil {
		t.Fatalf("AutoMigrate: %v", err)
	}
	return database
}

func TestLocaleRepository_Create_And_FindByCode(t *testing.T) {
	database := setupLocaleDB(t)
	repo := NewLocaleRepository(database)
	ctx := context.Background()

	locale := &entity.Locale{
		Code:      "en",
		Name:      "English",
		IsDefault: true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := repo.Create(ctx, locale); err != nil {
		t.Fatalf("Create: %v", err)
	}

	found, err := repo.FindByCode(ctx, "en")
	if err != nil {
		t.Fatalf("FindByCode: %v", err)
	}
	if found.Name != "English" {
		t.Errorf("Name = %q, want %q", found.Name, "English")
	}
	if !found.IsDefault {
		t.Error("expected IsDefault = true")
	}
}

func TestLocaleRepository_FindByCode_NotFound(t *testing.T) {
	database := setupLocaleDB(t)
	repo := NewLocaleRepository(database)

	_, err := repo.FindByCode(context.Background(), "nonexistent")
	if !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestLocaleRepository_FindAll_OrderedByCode(t *testing.T) {
	database := setupLocaleDB(t)
	repo := NewLocaleRepository(database)
	ctx := context.Background()

	now := time.Now()
	_ = repo.Create(ctx, &entity.Locale{Code: "vi", Name: "Tiếng Việt", CreatedAt: now, UpdatedAt: now})
	_ = repo.Create(ctx, &entity.Locale{Code: "en", Name: "English", CreatedAt: now, UpdatedAt: now})
	_ = repo.Create(ctx, &entity.Locale{Code: "ja", Name: "Japanese", CreatedAt: now, UpdatedAt: now})

	locales, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll: %v", err)
	}
	if len(locales) != 3 {
		t.Fatalf("len(locales) = %d, want 3", len(locales))
	}
	if locales[0].Code != "en" || locales[1].Code != "ja" || locales[2].Code != "vi" {
		t.Errorf("expected alphabetical order, got %s, %s, %s", locales[0].Code, locales[1].Code, locales[2].Code)
	}
}

func TestLocaleRepository_FindDefault(t *testing.T) {
	database := setupLocaleDB(t)
	repo := NewLocaleRepository(database)
	ctx := context.Background()

	now := time.Now()
	_ = repo.Create(ctx, &entity.Locale{Code: "en", Name: "English", IsDefault: true, CreatedAt: now, UpdatedAt: now})
	_ = repo.Create(ctx, &entity.Locale{Code: "vi", Name: "Tiếng Việt", IsDefault: false, CreatedAt: now, UpdatedAt: now})

	found, err := repo.FindDefault(ctx)
	if err != nil {
		t.Fatalf("FindDefault: %v", err)
	}
	if found.Code != "en" {
		t.Errorf("Code = %q, want %q", found.Code, "en")
	}
}

func TestLocaleRepository_FindDefault_NotFound(t *testing.T) {
	database := setupLocaleDB(t)
	repo := NewLocaleRepository(database)
	ctx := context.Background()

	now := time.Now()
	_ = repo.Create(ctx, &entity.Locale{Code: "en", Name: "English", IsDefault: false, CreatedAt: now, UpdatedAt: now})

	_, err := repo.FindDefault(ctx)
	if !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestLocaleRepository_ClearDefault(t *testing.T) {
	database := setupLocaleDB(t)
	repo := NewLocaleRepository(database)
	ctx := context.Background()

	now := time.Now()
	_ = repo.Create(ctx, &entity.Locale{Code: "en", Name: "English", IsDefault: true, CreatedAt: now, UpdatedAt: now})
	_ = repo.Create(ctx, &entity.Locale{Code: "vi", Name: "Tiếng Việt", IsDefault: true, CreatedAt: now, UpdatedAt: now})

	if err := repo.ClearDefault(ctx); err != nil {
		t.Fatalf("ClearDefault: %v", err)
	}

	_, err := repo.FindDefault(ctx)
	if !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
		t.Errorf("expected no default after ClearDefault, got err = %v", err)
	}
}

func TestLocaleRepository_Update(t *testing.T) {
	database := setupLocaleDB(t)
	repo := NewLocaleRepository(database)
	ctx := context.Background()

	now := time.Now()
	_ = repo.Create(ctx, &entity.Locale{Code: "en", Name: "English", IsDefault: false, CreatedAt: now, UpdatedAt: now})

	locale, _ := repo.FindByCode(ctx, "en")
	locale.Name = "English (US)"
	locale.IsDefault = true
	locale.UpdatedAt = time.Now()
	if err := repo.Update(ctx, locale); err != nil {
		t.Fatalf("Update: %v", err)
	}

	updated, _ := repo.FindByCode(ctx, "en")
	if updated.Name != "English (US)" {
		t.Errorf("Name = %q, want %q", updated.Name, "English (US)")
	}
	if !updated.IsDefault {
		t.Error("expected IsDefault = true after update")
	}
}

func TestLocaleRepository_Delete(t *testing.T) {
	database := setupLocaleDB(t)
	repo := NewLocaleRepository(database)
	ctx := context.Background()

	now := time.Now()
	_ = repo.Create(ctx, &entity.Locale{Code: "en", Name: "English", CreatedAt: now, UpdatedAt: now})

	if err := repo.Delete(ctx, "en"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := repo.FindByCode(ctx, "en")
	if !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestLocaleRepository_Delete_NotFound(t *testing.T) {
	database := setupLocaleDB(t)
	repo := NewLocaleRepository(database)

	err := repo.Delete(context.Background(), "nonexistent")
	if !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}
