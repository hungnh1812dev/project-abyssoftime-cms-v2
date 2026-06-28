package gormdb

import (
	"context"
	"fmt"
	"testing"
	"time"

	"gorm.io/gorm"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

func ptrTime(t time.Time) *time.Time { return &t }

var testFields = []entity.FieldDefinition{
	{Name: "title", Type: "text"},
	{Name: "body", Type: "richtext"},
}

func setupDocDB(t *testing.T, slugs ...string) *gorm.DB {
	t.Helper()
	db, err := NewClient("sqlite", ":memory:", false)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	repo := NewDocumentRepository(db)
	ctx := context.Background()
	for _, slug := range slugs {
		if err := repo.EnsureCollection(ctx, slug, testFields); err != nil {
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
			DocumentID: fmt.Sprintf("d%d", i),
			Version:    entity.VersionDraft,
			Fields:     map[string]any{"title": fmt.Sprintf("Post %d", i)},
			Locale:     "en",
			CreatedAt:  time.Now().UTC().Add(time.Duration(i) * time.Second),
			UpdatedAt:  time.Now().UTC(),
		}
		_ = repo.UpsertDraft(ctx, "blog", doc)
	}

	docs, total, err := repo.FindDraftsByContentTypePaginated(ctx, "blog", 0, 2, "en", "createdAt", -1, nil)
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

	docs, total, _ := repo.FindDraftsByContentTypePaginated(ctx, "blog", 0, 100, "en", "createdAt", -1, nil)
	if total != 0 || len(docs) != 0 {
		t.Errorf("blog docs after delete: total=%d, len=%d", total, len(docs))
	}

	docs, total, _ = repo.FindDraftsByContentTypePaginated(ctx, "other", 0, 100, "en", "createdAt", -1, nil)
	if total != 1 {
		t.Errorf("other docs after delete: total=%d, want 1", total)
	}
}

func TestDocumentRepository_EnsureCollection_CreatesTable(t *testing.T) {
	db, err := NewClient("sqlite", ":memory:", false)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	repo := NewDocumentRepository(db)
	ctx := context.Background()

	if err := repo.EnsureCollection(ctx, "blog-posts", testFields); err != nil {
		t.Fatalf("EnsureCollection: %v", err)
	}

	if !db.Migrator().HasTable(documentTableName("blog-posts")) {
		t.Error("expected table documents_blog_posts to exist")
	}

	// idempotent
	if err := repo.EnsureCollection(ctx, "blog-posts", testFields); err != nil {
		t.Fatalf("EnsureCollection (2nd call): %v", err)
	}
}

func TestDocumentRepository_DropCollection_DropsTable(t *testing.T) {
	db, err := NewClient("sqlite", ":memory:", false)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	repo := NewDocumentRepository(db)
	ctx := context.Background()

	_ = repo.EnsureCollection(ctx, "blog-posts", testFields)
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

func TestExistingColumns(t *testing.T) {
	db, err := NewClient("sqlite", ":memory:", false)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	repo := NewDocumentRepository(db)
	ctx := context.Background()

	if err := repo.EnsureCollection(ctx, "blog", testFields); err != nil {
		t.Fatalf("EnsureCollection: %v", err)
	}

	cols, err := existingColumns(db, documentTableName("blog"))
	if err != nil {
		t.Fatalf("existingColumns: %v", err)
	}

	for _, want := range []string{"gorm_id", "document_id", "version", "locale", "title", "body", "created_at", "updated_at"} {
		if !cols[want] {
			t.Errorf("missing column %q", want)
		}
	}

	if cols["nonexistent"] {
		t.Error("unexpected column 'nonexistent'")
	}
}

func TestDocumentRepository_EnsureCollection_PreservesData(t *testing.T) {
	db, err := NewClient("sqlite", ":memory:", false)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	repo := NewDocumentRepository(db)
	ctx := context.Background()

	if err := repo.EnsureCollection(ctx, "blog", testFields); err != nil {
		t.Fatalf("EnsureCollection: %v", err)
	}

	doc := &entity.Document{
		DocumentID: "d1", Version: entity.VersionDraft,
		Fields: map[string]any{"title": "Keep me"}, Locale: "en",
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	if err := repo.UpsertDraft(ctx, "blog", doc); err != nil {
		t.Fatalf("UpsertDraft: %v", err)
	}

	// Re-run EnsureCollection — data must survive
	if err := repo.EnsureCollection(ctx, "blog", testFields); err != nil {
		t.Fatalf("EnsureCollection (2nd): %v", err)
	}

	found, err := repo.FindDraftByDocumentID(ctx, "blog", "d1", "en")
	if err != nil {
		t.Fatalf("FindDraftByDocumentID after re-ensure: %v", err)
	}
	title, _ := found.Fields["title"].(string)
	if title != "Keep me" {
		t.Errorf("title = %q, want %q", title, "Keep me")
	}
}

func TestDocumentRepository_EnsureCollection_AddsNewColumn(t *testing.T) {
	db, err := NewClient("sqlite", ":memory:", false)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	repo := NewDocumentRepository(db)
	ctx := context.Background()

	if err := repo.EnsureCollection(ctx, "blog", testFields); err != nil {
		t.Fatalf("EnsureCollection: %v", err)
	}

	doc := &entity.Document{
		DocumentID: "d1", Version: entity.VersionDraft,
		Fields: map[string]any{"title": "Hello"}, Locale: "en",
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	if err := repo.UpsertDraft(ctx, "blog", doc); err != nil {
		t.Fatalf("UpsertDraft: %v", err)
	}

	extendedFields := append(testFields, entity.FieldDefinition{Name: "summary", Type: "text"})
	if err := repo.EnsureCollection(ctx, "blog", extendedFields); err != nil {
		t.Fatalf("EnsureCollection (extended): %v", err)
	}

	// Old data still exists
	found, err := repo.FindDraftByDocumentID(ctx, "blog", "d1", "en")
	if err != nil {
		t.Fatalf("FindDraftByDocumentID: %v", err)
	}
	title, _ := found.Fields["title"].(string)
	if title != "Hello" {
		t.Errorf("title = %q, want %q", title, "Hello")
	}

	// New column exists
	cols, err := existingColumns(db, documentTableName("blog"))
	if err != nil {
		t.Fatalf("existingColumns: %v", err)
	}
	if !cols["summary"] {
		t.Error("expected 'summary' column to exist after extending fields")
	}
}

func TestDocumentRepository_EnsureCollection_IgnoresRemovedField(t *testing.T) {
	db, err := NewClient("sqlite", ":memory:", false)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	repo := NewDocumentRepository(db)
	ctx := context.Background()

	extendedFields := append(testFields, entity.FieldDefinition{Name: "summary", Type: "text"})
	if err := repo.EnsureCollection(ctx, "blog", extendedFields); err != nil {
		t.Fatalf("EnsureCollection: %v", err)
	}

	doc := &entity.Document{
		DocumentID: "d1", Version: entity.VersionDraft,
		Fields: map[string]any{"title": "Hello", "summary": "World"}, Locale: "en",
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	if err := repo.UpsertDraft(ctx, "blog", doc); err != nil {
		t.Fatalf("UpsertDraft: %v", err)
	}

	// Re-ensure with fewer fields — summary column must remain
	if err := repo.EnsureCollection(ctx, "blog", testFields); err != nil {
		t.Fatalf("EnsureCollection (reduced): %v", err)
	}

	found, err := repo.FindDraftByDocumentID(ctx, "blog", "d1", "en")
	if err != nil {
		t.Fatalf("FindDraftByDocumentID: %v", err)
	}
	title, _ := found.Fields["title"].(string)
	if title != "Hello" {
		t.Errorf("title = %q, want %q", title, "Hello")
	}

	cols, err := existingColumns(db, documentTableName("blog"))
	if err != nil {
		t.Fatalf("existingColumns: %v", err)
	}
	if !cols["summary"] {
		t.Error("expected 'summary' column to still exist after reducing fields")
	}
}

func TestDocumentRepository_WidthFieldsRoundTrip(t *testing.T) {
	db, err := NewClient("sqlite", ":memory:", false)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	repo := NewDocumentRepository(db)
	ctx := context.Background()

	widthFields := []entity.FieldDefinition{
		{Name: "packName", Type: "text", Width: "50%"},
		{Name: "packTitle", Type: "text", Width: "50%"},
		{Name: "words", Type: "json"},
	}

	if err := repo.EnsureCollection(ctx, "en-vocab-pack", widthFields); err != nil {
		t.Fatalf("EnsureCollection: %v", err)
	}

	cols, err := existingColumns(db, documentTableName("en-vocab-pack"))
	if err != nil {
		t.Fatalf("existingColumns: %v", err)
	}
	if !cols["pack_name"] {
		t.Error("expected 'pack_name' column to exist")
	}
	if !cols["pack_title"] {
		t.Error("expected 'pack_title' column to exist")
	}

	doc := &entity.Document{
		DocumentID: "d1", Version: entity.VersionDraft,
		Fields:    map[string]any{"packName": "Pack 1", "packTitle": "Title 1", "words": `["hello"]`},
		Locale:    "en",
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	if err := repo.UpsertDraft(ctx, "en-vocab-pack", doc); err != nil {
		t.Fatalf("UpsertDraft: %v", err)
	}

	found, err := repo.FindDraftByDocumentID(ctx, "en-vocab-pack", "d1", "en")
	if err != nil {
		t.Fatalf("FindDraftByDocumentID: %v", err)
	}
	if got := found.Fields["packName"]; got != "Pack 1" {
		t.Errorf("packName = %v, want %q", got, "Pack 1")
	}
	if got := found.Fields["packTitle"]; got != "Title 1" {
		t.Errorf("packTitle = %v, want %q", got, "Title 1")
	}

	doc.Fields["packName"] = "Pack 2"
	doc.Fields["packTitle"] = "Title 2"
	if err := repo.UpsertDraft(ctx, "en-vocab-pack", doc); err != nil {
		t.Fatalf("UpsertDraft (update): %v", err)
	}

	found, err = repo.FindDraftByDocumentID(ctx, "en-vocab-pack", "d1", "en")
	if err != nil {
		t.Fatalf("FindDraftByDocumentID (after update): %v", err)
	}
	if got := found.Fields["packName"]; got != "Pack 2" {
		t.Errorf("packName after update = %v, want %q", got, "Pack 2")
	}
	if got := found.Fields["packTitle"]; got != "Title 2" {
		t.Errorf("packTitle after update = %v, want %q", got, "Title 2")
	}
}

func TestDocumentRepository_WidthFieldsMigrateExistingTable(t *testing.T) {
	db, err := NewClient("sqlite", ":memory:", false)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	repo := NewDocumentRepository(db)
	ctx := context.Background()

	oldSQL := fmt.Sprintf("CREATE TABLE %s (gorm_id INTEGER PRIMARY KEY AUTOINCREMENT, document_id TEXT, version TEXT, locale TEXT, words TEXT, created_at TIMESTAMP, updated_at TIMESTAMP, published_at TIMESTAMP, created_by TEXT, updated_by TEXT, published_by TEXT)",
		documentTableName("en-vocab-pack"))
	if err := db.Exec(oldSQL).Error; err != nil {
		t.Fatalf("create old table: %v", err)
	}

	insertSQL := fmt.Sprintf("INSERT INTO %s (document_id, version, locale, words, created_at, updated_at, created_by, updated_by) VALUES ('d1', 'draft', 'en', '[\"hello\"]', datetime('now'), datetime('now'), 'u1', 'u1')",
		documentTableName("en-vocab-pack"))
	if err := db.Exec(insertSQL).Error; err != nil {
		t.Fatalf("insert: %v", err)
	}

	widthFields := []entity.FieldDefinition{
		{Name: "packName", Type: "text", Width: "50%"},
		{Name: "packTitle", Type: "text", Width: "50%"},
		{Name: "words", Type: "json"},
	}
	if err := repo.EnsureCollection(ctx, "en-vocab-pack", widthFields); err != nil {
		t.Fatalf("EnsureCollection: %v", err)
	}

	cols, err := existingColumns(db, documentTableName("en-vocab-pack"))
	if err != nil {
		t.Fatalf("existingColumns: %v", err)
	}
	if !cols["pack_name"] {
		t.Error("expected 'pack_name' column after migration")
	}
	if !cols["pack_title"] {
		t.Error("expected 'pack_title' column after migration")
	}

	doc := &entity.Document{
		DocumentID: "d1", Version: entity.VersionDraft,
		Fields:    map[string]any{"packName": "New Pack", "packTitle": "New Title", "words": `["hello"]`},
		Locale:    "en",
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
		CreatedBy: "u1", UpdatedBy: "u1",
	}
	if err := repo.UpsertDraft(ctx, "en-vocab-pack", doc); err != nil {
		t.Fatalf("UpsertDraft: %v", err)
	}

	found, err := repo.FindDraftByDocumentID(ctx, "en-vocab-pack", "d1", "en")
	if err != nil {
		t.Fatalf("FindDraftByDocumentID: %v", err)
	}
	if got := found.Fields["packName"]; got != "New Pack" {
		t.Errorf("packName = %v, want %q", got, "New Pack")
	}
	if got := found.Fields["packTitle"]; got != "New Title" {
		t.Errorf("packTitle = %v, want %q", got, "New Title")
	}
}

func TestDocumentRepository_TableInfo(t *testing.T) {
	db, err := NewClient("sqlite", ":memory:", false)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	repo := NewDocumentRepository(db)
	ctx := context.Background()

	// Table doesn't exist yet
	exists, count, err := repo.TableInfo(ctx, "blog")
	if err != nil {
		t.Fatalf("TableInfo: %v", err)
	}
	if exists {
		t.Error("expected exists=false before EnsureCollection")
	}

	// Create table and insert data
	if err := repo.EnsureCollection(ctx, "blog", testFields); err != nil {
		t.Fatalf("EnsureCollection: %v", err)
	}
	for _, id := range []string{"d1", "d2", "d3"} {
		_ = repo.UpsertDraft(ctx, "blog", &entity.Document{
			DocumentID: id, Version: entity.VersionDraft,
			Fields: map[string]any{"title": id}, Locale: "en",
			CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
		})
	}

	exists, count, err = repo.TableInfo(ctx, "blog")
	if err != nil {
		t.Fatalf("TableInfo: %v", err)
	}
	if !exists {
		t.Error("expected exists=true after EnsureCollection")
	}
	if count != 3 {
		t.Errorf("count = %d, want 3", count)
	}
}
