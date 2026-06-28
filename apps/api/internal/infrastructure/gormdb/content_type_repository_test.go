package gormdb

import (
	"context"
	"testing"
	"time"

	"gorm.io/gorm"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

func setupCTDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := NewClient("sqlite", ":memory:", false)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if err := db.AutoMigrate(&entity.ContentType{}); err != nil {
		t.Fatalf("AutoMigrate: %v", err)
	}
	return db
}

func TestContentTypeRepository_Create_And_FindByID(t *testing.T) {
	db := setupCTDB(t)
	repo := NewContentTypeRepository(db)
	ctx := context.Background()

	ct := &entity.ContentType{
		DocumentID: "ct1",
		Name:       "Test Content Type",
		Slug:       "blog",
		Kind:       entity.KindCollection,
		Fields: []entity.FieldDefinition{
			{Name: "title", Type: "text"},
			{Name: "body", Type: "richtext"},
		},
		ListFields: []string{"title"},
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
	if err := repo.Create(ctx, ct); err != nil {
		t.Fatalf("Create: %v", err)
	}

	found, err := repo.FindByID(ctx, "ct1")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.Name != "Test Content Type" {
		t.Errorf("Name = %q, want %q", found.Name, "Test Content Type")
	}
	if len(found.Fields) != 2 {
		t.Errorf("Fields count = %d, want 2", len(found.Fields))
	}
	if len(found.ListFields) != 1 || found.ListFields[0] != "title" {
		t.Errorf("ListFields = %v, want [title]", found.ListFields)
	}
}

func TestContentTypeRepository_FindBySlug(t *testing.T) {
	db := setupCTDB(t)
	repo := NewContentTypeRepository(db)
	ctx := context.Background()

	_ = repo.Create(ctx, &entity.ContentType{DocumentID: "ct1", Name: "Blog", Slug: "blog", Kind: entity.KindCollection})

	found, err := repo.FindBySlug(ctx, "blog")
	if err != nil {
		t.Fatalf("FindBySlug: %v", err)
	}
	if found.DocumentID != "ct1" {
		t.Errorf("DocumentID = %q, want %q", found.DocumentID, "ct1")
	}
}

func TestContentTypeRepository_FindByID_NotFound(t *testing.T) {
	db := setupCTDB(t)
	repo := NewContentTypeRepository(db)

	_, err := repo.FindByID(context.Background(), "nonexistent")
	if !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestContentTypeRepository_FindBySlug_NotFound(t *testing.T) {
	db := setupCTDB(t)
	repo := NewContentTypeRepository(db)

	_, err := repo.FindBySlug(context.Background(), "missing")
	if !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestContentTypeRepository_FindAll(t *testing.T) {
	db := setupCTDB(t)
	repo := NewContentTypeRepository(db)
	ctx := context.Background()

	_ = repo.Create(ctx, &entity.ContentType{DocumentID: "ct1", Name: "Blog", Slug: "blog", Kind: entity.KindCollection})
	_ = repo.Create(ctx, &entity.ContentType{DocumentID: "ct2", Name: "Homepage", Slug: "homepage", Kind: entity.KindSingle})

	all, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("count = %d, want 2", len(all))
	}
}

func TestContentTypeRepository_Update(t *testing.T) {
	db := setupCTDB(t)
	repo := NewContentTypeRepository(db)
	ctx := context.Background()

	ct := &entity.ContentType{DocumentID: "ct1", Name: "Blog", Slug: "blog", Kind: entity.KindCollection}
	_ = repo.Create(ctx, ct)

	ct.Name = "Articles"
	ct.UpdatedAt = time.Now().UTC()
	if err := repo.Update(ctx, ct); err != nil {
		t.Fatalf("Update: %v", err)
	}

	found, _ := repo.FindByID(ctx, "ct1")
	if found.Name != "Articles" {
		t.Errorf("Name = %q, want %q", found.Name, "Articles")
	}
}

func TestContentTypeRepository_Delete(t *testing.T) {
	db := setupCTDB(t)
	repo := NewContentTypeRepository(db)
	ctx := context.Background()

	_ = repo.Create(ctx, &entity.ContentType{DocumentID: "ct1", Name: "Blog", Slug: "blog", Kind: entity.KindCollection})

	if err := repo.Delete(ctx, "ct1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := repo.FindByID(ctx, "ct1")
	if !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
		t.Errorf("after Delete, FindByID err = %v, want ErrNotFound", err)
	}
}
