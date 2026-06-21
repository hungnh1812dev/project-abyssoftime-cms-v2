package gormdb

import (
	"context"
	"testing"
	"time"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

var compTestFields = []entity.FieldDefinition{
	{Name: "title", Type: "text"},
	{Name: "description", Type: "text"},
}

func setupCompDB(t *testing.T, slug, comp string) *componentRepository {
	t.Helper()
	db, err := NewClient("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	repo := &componentRepository{db: db}
	if err := repo.EnsureCollection(context.Background(), slug, comp, compTestFields); err != nil {
		t.Fatalf("EnsureCollection: %v", err)
	}
	return repo
}

func TestComponentRepository_EnsureCollection_CreatesTable(t *testing.T) {
	db, err := NewClient("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	repo := NewComponentRepository(db)
	ctx := context.Background()

	if err := repo.EnsureCollection(ctx, "blog-posts", "banner", compTestFields); err != nil {
		t.Fatalf("EnsureCollection: %v", err)
	}

	if !db.Migrator().HasTable(componentTableName("blog-posts", "banner")) {
		t.Error("expected table components_blog_posts_banner to exist")
	}

	// idempotent
	if err := repo.EnsureCollection(ctx, "blog-posts", "banner", compTestFields); err != nil {
		t.Fatalf("EnsureCollection (2nd): %v", err)
	}
}

func TestComponentRepository_DropCollection(t *testing.T) {
	db, err := NewClient("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	repo := NewComponentRepository(db)
	ctx := context.Background()

	_ = repo.EnsureCollection(ctx, "blog-posts", "banner", compTestFields)
	if err := repo.DropCollection(ctx, "blog-posts", "banner"); err != nil {
		t.Fatalf("DropCollection: %v", err)
	}

	if db.Migrator().HasTable(componentTableName("blog-posts", "banner")) {
		t.Error("expected table to not exist after drop")
	}
}

func TestComponentRepository_UpsertAll_And_FindByDocumentID(t *testing.T) {
	repo := setupCompDB(t, "blog", "seo")
	ctx := context.Background()
	now := time.Now().UTC()

	components := []*entity.Component{
		{ComponentID: "c1", Fields: map[string]any{"title": "SEO Title"}, CreatedAt: now, UpdatedAt: now},
		{ComponentID: "c2", Fields: map[string]any{"description": "SEO Desc"}, CreatedAt: now, UpdatedAt: now},
	}

	if err := repo.UpsertAll(ctx, "blog", "seo", "d1", "en", entity.VersionDraft, components); err != nil {
		t.Fatalf("UpsertAll: %v", err)
	}

	found, err := repo.FindByDocumentID(ctx, "blog", "seo", "d1", "en", entity.VersionDraft)
	if err != nil {
		t.Fatalf("FindByDocumentID: %v", err)
	}
	if len(found) != 2 {
		t.Fatalf("count = %d, want 2", len(found))
	}
	title, _ := found[0].Fields["title"].(string)
	if title != "SEO Title" {
		t.Errorf("Data.title = %q, want %q", title, "SEO Title")
	}
}

func TestComponentRepository_UpsertAll_Replaces(t *testing.T) {
	repo := setupCompDB(t, "blog", "seo")
	ctx := context.Background()
	now := time.Now().UTC()

	initial := []*entity.Component{
		{ComponentID: "c1", Fields: map[string]any{"title": "v1"}, CreatedAt: now, UpdatedAt: now},
		{ComponentID: "c2", Fields: map[string]any{"title": "v2"}, CreatedAt: now, UpdatedAt: now},
	}
	_ = repo.UpsertAll(ctx, "blog", "seo", "d1", "en", entity.VersionDraft, initial)

	replacement := []*entity.Component{
		{ComponentID: "c3", Fields: map[string]any{"title": "v3"}, CreatedAt: now, UpdatedAt: now},
	}
	if err := repo.UpsertAll(ctx, "blog", "seo", "d1", "en", entity.VersionDraft, replacement); err != nil {
		t.Fatalf("UpsertAll replace: %v", err)
	}

	found, _ := repo.FindByDocumentID(ctx, "blog", "seo", "d1", "en", entity.VersionDraft)
	if len(found) != 1 {
		t.Fatalf("count = %d, want 1", len(found))
	}
	if found[0].ComponentID != "c3" {
		t.Errorf("ComponentID = %q, want c3", found[0].ComponentID)
	}
}

func TestComponentRepository_DeleteByDocumentID(t *testing.T) {
	repo := setupCompDB(t, "blog", "seo")
	ctx := context.Background()
	now := time.Now().UTC()

	_ = repo.UpsertAll(ctx, "blog", "seo", "d1", "en", entity.VersionDraft, []*entity.Component{
		{ComponentID: "c1", Fields: map[string]any{}, CreatedAt: now, UpdatedAt: now},
	})

	if err := repo.DeleteByDocumentID(ctx, "blog", "seo", "d1", "en"); err != nil {
		t.Fatalf("DeleteByDocumentID: %v", err)
	}

	found, _ := repo.FindByDocumentID(ctx, "blog", "seo", "d1", "en", entity.VersionDraft)
	if len(found) != 0 {
		t.Errorf("count = %d, want 0", len(found))
	}
}

func TestComponentRepository_DeleteAllByContentType(t *testing.T) {
	repo := setupCompDB(t, "blog", "seo")
	ctx := context.Background()
	now := time.Now().UTC()

	_ = repo.UpsertAll(ctx, "blog", "seo", "d1", "en", entity.VersionDraft, []*entity.Component{
		{ComponentID: "c1", Fields: map[string]any{}, CreatedAt: now, UpdatedAt: now},
	})
	_ = repo.UpsertAll(ctx, "blog", "seo", "d2", "en", entity.VersionDraft, []*entity.Component{
		{ComponentID: "c2", Fields: map[string]any{}, CreatedAt: now, UpdatedAt: now},
	})

	if err := repo.DeleteAllByContentType(ctx, "blog", "seo"); err != nil {
		t.Fatalf("DeleteAllByContentType: %v", err)
	}

	found, _ := repo.FindByDocumentID(ctx, "blog", "seo", "d1", "en", entity.VersionDraft)
	if len(found) != 0 {
		t.Errorf("d1 count = %d, want 0", len(found))
	}
	found, _ = repo.FindByDocumentID(ctx, "blog", "seo", "d2", "en", entity.VersionDraft)
	if len(found) != 0 {
		t.Errorf("d2 count = %d, want 0", len(found))
	}
}
