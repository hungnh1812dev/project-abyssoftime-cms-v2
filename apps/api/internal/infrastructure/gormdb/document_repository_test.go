package gormdb

import (
	"context"
	"testing"
	"time"

	"gorm.io/gorm"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

func ptrTime(t time.Time) *time.Time { return &t }

func setupDocDB(t *testing.T, slugs ...string) *gorm.DB {
	t.Helper()
	db, err := NewClient("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	repo := NewDocumentRepository(db)
	ctx := context.Background()
	for _, slug := range slugs {
		if err := repo.EnsureCollection(ctx, slug); err != nil {
			t.Fatalf("EnsureCollection(%s): %v", slug, err)
		}
	}
	return db
}

func TestDocumentRepository_UpsertDraft_And_FindDraft(t *testing.T) {
	db := setupDocDB(t, "blog")
	repo := NewDocumentRepository(db)
	ctx := context.Background()

	doc := &entity.Document{
		DocumentID: "d1",
		Version:    entity.VersionDraft,
		Fields:       map[string]any{"title": "Hello"},
		Locale:     "en",
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
	if err := repo.UpsertDraft(ctx, "blog", doc); err != nil {
		t.Fatalf("UpsertDraft: %v", err)
	}

	found, err := repo.FindDraftByDocumentID(ctx, "blog", "d1", "en")
	if err != nil {
		t.Fatalf("FindDraftByDocumentID: %v", err)
	}
	if found.DocumentID != "d1" {
		t.Errorf("DocumentID = %q, want %q", found.DocumentID, "d1")
	}
	title, _ := found.Fields["title"].(string)
	if title != "Hello" {
		t.Errorf("Data.title = %q, want %q", title, "Hello")
	}
}

func TestDocumentRepository_UpsertDraft_Updates(t *testing.T) {
	db := setupDocDB(t, "blog")
	repo := NewDocumentRepository(db)
	ctx := context.Background()

	doc := &entity.Document{
		DocumentID: "d1", Version: entity.VersionDraft,
		Fields: map[string]any{"title": "v1"}, Locale: "en",
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	_ = repo.UpsertDraft(ctx, "blog", doc)

	doc.Fields = map[string]any{"title": "v2"}
	doc.UpdatedAt = time.Now().UTC()
	if err := repo.UpsertDraft(ctx, "blog", doc); err != nil {
		t.Fatalf("UpsertDraft update: %v", err)
	}

	found, _ := repo.FindDraftByDocumentID(ctx, "blog", "d1", "en")
	title, _ := found.Fields["title"].(string)
	if title != "v2" {
		t.Errorf("Data.title = %q, want %q", title, "v2")
	}
}

func TestDocumentRepository_FindDraft_NotFound(t *testing.T) {
	db := setupDocDB(t, "blog")
	repo := NewDocumentRepository(db)

	_, err := repo.FindDraftByDocumentID(context.Background(), "blog", "nope", "en")
	if !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestDocumentRepository_UpsertPublished_And_FindPublished(t *testing.T) {
	db := setupDocDB(t, "blog")
	repo := NewDocumentRepository(db)
	ctx := context.Background()

	doc := &entity.Document{
		DocumentID: "d1", Version: entity.VersionPublished,
		Fields: map[string]any{"title": "Pub"}, Locale: "en",
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
		PublishedAt: ptrTime(time.Now().UTC()), PublishedBy: "admin",
	}
	if err := repo.UpsertPublished(ctx, "blog", doc); err != nil {
		t.Fatalf("UpsertPublished: %v", err)
	}

	found, err := repo.FindPublishedByDocumentID(ctx, "blog", "d1", "en")
	if err != nil {
		t.Fatalf("FindPublishedByDocumentID: %v", err)
	}
	if found.PublishedBy != "admin" {
		t.Errorf("PublishedBy = %q, want %q", found.PublishedBy, "admin")
	}
}

func TestDocumentRepository_FindDraftsByContentTypePaginated(t *testing.T) {
	db := setupDocDB(t, "blog")
	repo := NewDocumentRepository(db)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		doc := &entity.Document{
			DocumentID: "d" + string(rune('0'+i)),
			Version:    entity.VersionDraft,
			Fields:       map[string]any{"i": i},
			Locale:     "en",
			CreatedAt:  time.Now().UTC().Add(time.Duration(i) * time.Second),
			UpdatedAt:  time.Now().UTC(),
		}
		_ = repo.UpsertDraft(ctx, "blog", doc)
	}

	docs, total, err := repo.FindDraftsByContentTypePaginated(ctx, "blog", 0, 2, "en", "createdAt", -1)
	if err != nil {
		t.Fatalf("FindDraftsByContentTypePaginated: %v", err)
	}
	if total != 5 {
		t.Errorf("total = %d, want 5", total)
	}
	if len(docs) != 2 {
		t.Errorf("page size = %d, want 2", len(docs))
	}
}

func TestDocumentRepository_FindPublishedByDocumentIDs(t *testing.T) {
	db := setupDocDB(t, "blog")
	repo := NewDocumentRepository(db)
	ctx := context.Background()

	for _, id := range []string{"d1", "d2", "d3"} {
		doc := &entity.Document{
			DocumentID: id, Version: entity.VersionPublished,
			Fields: map[string]any{}, Locale: "en",
			CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
		}
		_ = repo.UpsertPublished(ctx, "blog", doc)
	}

	found, err := repo.FindPublishedByDocumentIDs(ctx, "blog", []string{"d1", "d3"}, "en")
	if err != nil {
		t.Fatalf("FindPublishedByDocumentIDs: %v", err)
	}
	if len(found) != 2 {
		t.Errorf("count = %d, want 2", len(found))
	}
}

func TestDocumentRepository_DeleteByDocumentID(t *testing.T) {
	db := setupDocDB(t, "blog")
	repo := NewDocumentRepository(db)
	ctx := context.Background()

	doc := &entity.Document{
		DocumentID: "d1", Version: entity.VersionDraft,
		Fields: map[string]any{}, Locale: "en",
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	_ = repo.UpsertDraft(ctx, "blog", doc)

	if err := repo.DeleteByDocumentID(ctx, "blog", "d1", "en"); err != nil {
		t.Fatalf("DeleteByDocumentID: %v", err)
	}

	_, err := repo.FindDraftByDocumentID(ctx, "blog", "d1", "en")
	if !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
		t.Errorf("after delete, err = %v, want ErrNotFound", err)
	}
}

func TestDocumentRepository_DeleteAllByContentType(t *testing.T) {
	db := setupDocDB(t, "blog", "other")
	repo := NewDocumentRepository(db)
	ctx := context.Background()

	for _, id := range []string{"d1", "d2"} {
		_ = repo.UpsertDraft(ctx, "blog", &entity.Document{
			DocumentID: id, Version: entity.VersionDraft,
			Fields: map[string]any{}, Locale: "en",
			CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
		})
	}
	_ = repo.UpsertDraft(ctx, "other", &entity.Document{
		DocumentID: "d3", Version: entity.VersionDraft,
		Fields: map[string]any{}, Locale: "en",
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	})

	if err := repo.DeleteAllByContentType(ctx, "blog"); err != nil {
		t.Fatalf("DeleteAllByContentType: %v", err)
	}

	docs, total, _ := repo.FindDraftsByContentTypePaginated(ctx, "blog", 0, 100, "en", "createdAt", -1)
	if total != 0 || len(docs) != 0 {
		t.Errorf("blog docs after delete: total=%d, len=%d", total, len(docs))
	}

	docs, total, _ = repo.FindDraftsByContentTypePaginated(ctx, "other", 0, 100, "en", "createdAt", -1)
	if total != 1 {
		t.Errorf("other docs after delete: total=%d, want 1", total)
	}
}

func TestDocumentRepository_EnsureCollection_CreatesTable(t *testing.T) {
	db, err := NewClient("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	repo := NewDocumentRepository(db)
	ctx := context.Background()

	if err := repo.EnsureCollection(ctx, "blog-posts"); err != nil {
		t.Fatalf("EnsureCollection: %v", err)
	}

	if !db.Migrator().HasTable(documentTableName("blog-posts")) {
		t.Error("expected table documents_blog_posts to exist")
	}

	// idempotent
	if err := repo.EnsureCollection(ctx, "blog-posts"); err != nil {
		t.Fatalf("EnsureCollection (2nd call): %v", err)
	}
}

func TestDocumentRepository_DropCollection_DropsTable(t *testing.T) {
	db, err := NewClient("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	repo := NewDocumentRepository(db)
	ctx := context.Background()

	_ = repo.EnsureCollection(ctx, "blog-posts")
	if err := repo.DropCollection(ctx, "blog-posts"); err != nil {
		t.Fatalf("DropCollection: %v", err)
	}

	if db.Migrator().HasTable(documentTableName("blog-posts")) {
		t.Error("expected table documents_blog_posts to not exist after drop")
	}
}

func TestDocumentRepository_FindDraftsByContentType(t *testing.T) {
	db := setupDocDB(t, "blog")
	repo := NewDocumentRepository(db)
	ctx := context.Background()

	_ = repo.UpsertDraft(ctx, "blog", &entity.Document{
		DocumentID: "d1", Version: entity.VersionDraft,
		Fields: map[string]any{}, Locale: "en",
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	})

	docs, err := repo.FindDraftsByContentType(ctx, "blog")
	if err != nil {
		t.Fatalf("FindDraftsByContentType: %v", err)
	}
	if len(docs) != 1 {
		t.Errorf("count = %d, want 1", len(docs))
	}
}

func TestDocumentRepository_DeletePublishedByDocumentID(t *testing.T) {
	db := setupDocDB(t, "blog")
	repo := NewDocumentRepository(db)
	ctx := context.Background()

	_ = repo.UpsertPublished(ctx, "blog", &entity.Document{
		DocumentID: "d1", Version: entity.VersionPublished,
		Fields: map[string]any{}, Locale: "en",
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	})

	if err := repo.DeletePublishedByDocumentID(ctx, "blog", "d1", "en"); err != nil {
		t.Fatalf("DeletePublishedByDocumentID: %v", err)
	}

	_, err := repo.FindPublishedByDocumentID(ctx, "blog", "d1", "en")
	if !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
		t.Errorf("after delete, err = %v, want ErrNotFound", err)
	}
}
