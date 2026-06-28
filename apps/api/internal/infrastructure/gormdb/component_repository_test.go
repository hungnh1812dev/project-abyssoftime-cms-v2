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
	db, err := NewClient("sqlite", ":memory:", false)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	repo := &componentRepository{database: db}
	if err := repo.EnsureCollection(context.Background(), slug, comp, compTestFields, false); err != nil {
		t.Fatalf("EnsureCollection: %v", err)
	}
	return repo
}

func TestComponentRepository_EnsureCollection_CreatesTable(t *testing.T) {
	db, err := NewClient("sqlite", ":memory:", false)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	repo := NewComponentRepository(db)
	ctx := context.Background()

	if err := repo.EnsureCollection(ctx, "blog-posts", "banner", compTestFields, false); err != nil {
		t.Fatalf("EnsureCollection: %v", err)
	}

	if !db.Migrator().HasTable(componentTableName("blog-posts", "banner")) {
		t.Error("expected table components_blog_posts_banner to exist")
	}

	// idempotent
	if err := repo.EnsureCollection(ctx, "blog-posts", "banner", compTestFields, false); err != nil {
		t.Fatalf("EnsureCollection (2nd): %v", err)
	}
}

func TestComponentRepository_DropCollection(t *testing.T) {
	db, err := NewClient("sqlite", ":memory:", false)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	repo := NewComponentRepository(db)
	ctx := context.Background()

	_ = repo.EnsureCollection(ctx, "blog-posts", "banner", compTestFields, false)
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

func TestComponentRepository_EnsureCollection_PreservesData(t *testing.T) {
	db, err := NewClient("sqlite", ":memory:", false)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	repo := &componentRepository{database: db}
	ctx := context.Background()
	now := time.Now().UTC()

	if err := repo.EnsureCollection(ctx, "blog", "seo", compTestFields, false); err != nil {
		t.Fatalf("EnsureCollection: %v", err)
	}

	components := []*entity.Component{
		{ComponentID: "c1", Fields: map[string]any{"title": "Keep me"}, CreatedAt: now, UpdatedAt: now},
	}
	if err := repo.UpsertAll(ctx, "blog", "seo", "d1", "en", entity.VersionDraft, components); err != nil {
		t.Fatalf("UpsertAll: %v", err)
	}

	// Re-run EnsureCollection — data must survive
	if err := repo.EnsureCollection(ctx, "blog", "seo", compTestFields, false); err != nil {
		t.Fatalf("EnsureCollection (2nd): %v", err)
	}

	found, err := repo.FindByDocumentID(ctx, "blog", "seo", "d1", "en", entity.VersionDraft)
	if err != nil {
		t.Fatalf("FindByDocumentID: %v", err)
	}
	if len(found) != 1 {
		t.Fatalf("count = %d, want 1", len(found))
	}
	title, _ := found[0].Fields["title"].(string)
	if title != "Keep me" {
		t.Errorf("title = %q, want %q", title, "Keep me")
	}
}

func TestComponentRepository_EnsureCollection_AddsNewColumn(t *testing.T) {
	db, err := NewClient("sqlite", ":memory:", false)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	repo := &componentRepository{database: db}
	ctx := context.Background()
	now := time.Now().UTC()

	if err := repo.EnsureCollection(ctx, "blog", "seo", compTestFields, false); err != nil {
		t.Fatalf("EnsureCollection: %v", err)
	}

	components := []*entity.Component{
		{ComponentID: "c1", Fields: map[string]any{"title": "Hello"}, CreatedAt: now, UpdatedAt: now},
	}
	if err := repo.UpsertAll(ctx, "blog", "seo", "d1", "en", entity.VersionDraft, components); err != nil {
		t.Fatalf("UpsertAll: %v", err)
	}

	extendedFields := append(compTestFields, entity.FieldDefinition{Name: "keywords", Type: "text"})
	if err := repo.EnsureCollection(ctx, "blog", "seo", extendedFields, false); err != nil {
		t.Fatalf("EnsureCollection (extended): %v", err)
	}

	// Old data still exists
	found, err := repo.FindByDocumentID(ctx, "blog", "seo", "d1", "en", entity.VersionDraft)
	if err != nil {
		t.Fatalf("FindByDocumentID: %v", err)
	}
	if len(found) != 1 {
		t.Fatalf("count = %d, want 1", len(found))
	}

	// New column exists
	cols, err := existingColumns(db, componentTableName("blog", "seo"))
	if err != nil {
		t.Fatalf("existingColumns: %v", err)
	}
	if !cols["keywords"] {
		t.Error("expected 'keywords' column to exist after extending fields")
	}
}

func TestComponentRepository_EnsureCollection_IgnoresRemovedField(t *testing.T) {
	db, err := NewClient("sqlite", ":memory:", false)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	repo := &componentRepository{database: db}
	ctx := context.Background()
	now := time.Now().UTC()

	extendedFields := append(compTestFields, entity.FieldDefinition{Name: "keywords", Type: "text"})
	if err := repo.EnsureCollection(ctx, "blog", "seo", extendedFields, false); err != nil {
		t.Fatalf("EnsureCollection: %v", err)
	}

	components := []*entity.Component{
		{ComponentID: "c1", Fields: map[string]any{"title": "Hello", "keywords": "test"}, CreatedAt: now, UpdatedAt: now},
	}
	if err := repo.UpsertAll(ctx, "blog", "seo", "d1", "en", entity.VersionDraft, components); err != nil {
		t.Fatalf("UpsertAll: %v", err)
	}

	// Re-ensure with fewer fields — keywords column must remain
	if err := repo.EnsureCollection(ctx, "blog", "seo", compTestFields, false); err != nil {
		t.Fatalf("EnsureCollection (reduced): %v", err)
	}

	found, err := repo.FindByDocumentID(ctx, "blog", "seo", "d1", "en", entity.VersionDraft)
	if err != nil {
		t.Fatalf("FindByDocumentID: %v", err)
	}
	if len(found) != 1 {
		t.Fatalf("count = %d, want 1", len(found))
	}

	cols, err := existingColumns(db, componentTableName("blog", "seo"))
	if err != nil {
		t.Fatalf("existingColumns: %v", err)
	}
	if !cols["keywords"] {
		t.Error("expected 'keywords' column to still exist after reducing fields")
	}
}

func TestComponentRepository_EnsureCollection_CreatesSortOrderColumn(t *testing.T) {
	db, err := NewClient("sqlite", ":memory:", false)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	repo := &componentRepository{database: db}
	ctx := context.Background()

	if err := repo.EnsureCollection(ctx, "blog", "seo", compTestFields, false); err != nil {
		t.Fatalf("EnsureCollection: %v", err)
	}

	cols, err := existingColumns(db, componentTableName("blog", "seo"))
	if err != nil {
		t.Fatalf("existingColumns: %v", err)
	}
	if !cols["sort_order"] {
		t.Error("expected 'sort_order' column to exist in new table")
	}
}

func TestComponentRepository_EnsureCollection_AddsSortOrderToExistingTable(t *testing.T) {
	db, err := NewClient("sqlite", ":memory:", false)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	repo := &componentRepository{database: db}
	ctx := context.Background()
	table := componentTableName("blog", "seo")

	// Create table WITHOUT sort_order (simulating legacy schema)
	sql := "CREATE TABLE " + table + " (gorm_id INTEGER PRIMARY KEY AUTOINCREMENT, component_id TEXT, document_id TEXT, version TEXT, locale TEXT, title TEXT, description TEXT, created_at TIMESTAMP, updated_at TIMESTAMP)"
	if err := db.Exec(sql).Error; err != nil {
		t.Fatalf("create legacy table: %v", err)
	}

	// EnsureCollection should add the missing sort_order column
	if err := repo.EnsureCollection(ctx, "blog", "seo", compTestFields, false); err != nil {
		t.Fatalf("EnsureCollection: %v", err)
	}

	cols, err := existingColumns(db, table)
	if err != nil {
		t.Fatalf("existingColumns: %v", err)
	}
	if !cols["sort_order"] {
		t.Error("expected 'sort_order' column to be added to existing table")
	}
}

func TestComponentRepository_SortOrder_PreservedThroughUpsert(t *testing.T) {
	repo := setupCompDB(t, "blog", "seo")
	ctx := context.Background()
	now := time.Now().UTC()

	components := []*entity.Component{
		{ComponentID: "c1", SortOrder: 1, Fields: map[string]any{"title": "Second"}, CreatedAt: now, UpdatedAt: now},
		{ComponentID: "c2", SortOrder: 0, Fields: map[string]any{"title": "First"}, CreatedAt: now, UpdatedAt: now},
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
	if found[0].SortOrder != 0 {
		t.Errorf("found[0].SortOrder = %d, want 0", found[0].SortOrder)
	}
	if found[1].SortOrder != 1 {
		t.Errorf("found[1].SortOrder = %d, want 1", found[1].SortOrder)
	}
	title0, _ := found[0].Fields["title"].(string)
	if title0 != "First" {
		t.Errorf("found[0].title = %q, want %q", title0, "First")
	}
	title1, _ := found[1].Fields["title"].(string)
	if title1 != "Second" {
		t.Errorf("found[1].title = %q, want %q", title1, "Second")
	}
}

// --- Nested component table tests ---

func setupNestedCompDB(t *testing.T, slug, comp string) *componentRepository {
	t.Helper()
	db, err := NewClient("sqlite", ":memory:", false)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	repo := &componentRepository{database: db}
	if err := repo.EnsureCollection(context.Background(), slug, comp, compTestFields, true); err != nil {
		t.Fatalf("EnsureCollection(isNested=true): %v", err)
	}
	return repo
}

func TestComponentRepository_EnsureCollection_TopLevel_HasDocumentID(t *testing.T) {
	db, err := NewClient("sqlite", ":memory:", false)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	repo := &componentRepository{database: db}
	ctx := context.Background()

	if err := repo.EnsureCollection(ctx, "blog", "banner", compTestFields, false); err != nil {
		t.Fatalf("EnsureCollection: %v", err)
	}

	cols, err := existingColumns(db, componentTableName("blog", "banner"))
	if err != nil {
		t.Fatalf("existingColumns: %v", err)
	}
	if !cols["document_id"] {
		t.Error("top-level table should have document_id column")
	}
	if cols["parent_component_id"] {
		t.Error("top-level table should NOT have parent_component_id column")
	}
}

func TestComponentRepository_EnsureCollection_Nested_HasParentComponentID(t *testing.T) {
	db, err := NewClient("sqlite", ":memory:", false)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	repo := &componentRepository{database: db}
	ctx := context.Background()

	if err := repo.EnsureCollection(ctx, "blog", "banner_child", compTestFields, true); err != nil {
		t.Fatalf("EnsureCollection: %v", err)
	}

	cols, err := existingColumns(db, componentTableName("blog", "banner_child"))
	if err != nil {
		t.Fatalf("existingColumns: %v", err)
	}
	if !cols["parent_component_id"] {
		t.Error("nested table should have parent_component_id column")
	}
	if cols["document_id"] {
		t.Error("nested table should NOT have document_id column")
	}
}

func TestComponentRepository_UpsertAllByParent_And_FindByParentComponentID(t *testing.T) {
	repo := setupNestedCompDB(t, "blog", "seo_og")
	ctx := context.Background()
	now := time.Now().UTC()

	components := []*entity.Component{
		{ComponentID: "child1", SortOrder: 0, Fields: map[string]any{"title": "OG Title"}, CreatedAt: now, UpdatedAt: now},
		{ComponentID: "child2", SortOrder: 1, Fields: map[string]any{"title": "OG Desc"}, CreatedAt: now, UpdatedAt: now},
	}

	if err := repo.UpsertAllByParent(ctx, "blog", "seo_og", "parent-comp-1", "en", entity.VersionDraft, components); err != nil {
		t.Fatalf("UpsertAllByParent: %v", err)
	}

	found, err := repo.FindByParentComponentID(ctx, "blog", "seo_og", "parent-comp-1", "en", entity.VersionDraft)
	if err != nil {
		t.Fatalf("FindByParentComponentID: %v", err)
	}
	if len(found) != 2 {
		t.Fatalf("count = %d, want 2", len(found))
	}
	if found[0].ParentComponentID != "parent-comp-1" {
		t.Errorf("ParentComponentID = %q, want %q", found[0].ParentComponentID, "parent-comp-1")
	}
	if found[0].DocumentID != "" {
		t.Errorf("DocumentID should be empty for nested component, got %q", found[0].DocumentID)
	}
	title, _ := found[0].Fields["title"].(string)
	if title != "OG Title" {
		t.Errorf("title = %q, want %q", title, "OG Title")
	}
	if found[0].SortOrder != 0 || found[1].SortOrder != 1 {
		t.Errorf("sort_order = [%d %d], want [0 1]", found[0].SortOrder, found[1].SortOrder)
	}
}

func TestComponentRepository_UpsertAllByParent_Replaces(t *testing.T) {
	repo := setupNestedCompDB(t, "blog", "seo_og")
	ctx := context.Background()
	now := time.Now().UTC()

	initial := []*entity.Component{
		{ComponentID: "c1", Fields: map[string]any{"title": "v1"}, CreatedAt: now, UpdatedAt: now},
		{ComponentID: "c2", Fields: map[string]any{"title": "v2"}, CreatedAt: now, UpdatedAt: now},
	}
	_ = repo.UpsertAllByParent(ctx, "blog", "seo_og", "parent-1", "en", entity.VersionDraft, initial)

	replacement := []*entity.Component{
		{ComponentID: "c3", Fields: map[string]any{"title": "v3"}, CreatedAt: now, UpdatedAt: now},
	}
	if err := repo.UpsertAllByParent(ctx, "blog", "seo_og", "parent-1", "en", entity.VersionDraft, replacement); err != nil {
		t.Fatalf("UpsertAllByParent replace: %v", err)
	}

	found, _ := repo.FindByParentComponentID(ctx, "blog", "seo_og", "parent-1", "en", entity.VersionDraft)
	if len(found) != 1 {
		t.Fatalf("count = %d, want 1", len(found))
	}
	if found[0].ComponentID != "c3" {
		t.Errorf("ComponentID = %q, want c3", found[0].ComponentID)
	}
}

func TestComponentRepository_DeleteByParentComponentID(t *testing.T) {
	repo := setupNestedCompDB(t, "blog", "seo_og")
	ctx := context.Background()
	now := time.Now().UTC()

	_ = repo.UpsertAllByParent(ctx, "blog", "seo_og", "parent-1", "en", entity.VersionDraft, []*entity.Component{
		{ComponentID: "c1", Fields: map[string]any{}, CreatedAt: now, UpdatedAt: now},
	})

	if err := repo.DeleteByParentComponentID(ctx, "blog", "seo_og", "parent-1", "en"); err != nil {
		t.Fatalf("DeleteByParentComponentID: %v", err)
	}

	found, _ := repo.FindByParentComponentID(ctx, "blog", "seo_og", "parent-1", "en", entity.VersionDraft)
	if len(found) != 0 {
		t.Errorf("count = %d, want 0", len(found))
	}
}

func TestComponentRepository_NestedLocaleIsolation(t *testing.T) {
	repo := setupNestedCompDB(t, "blog", "seo_og")
	ctx := context.Background()
	now := time.Now().UTC()

	_ = repo.UpsertAllByParent(ctx, "blog", "seo_og", "parent-1", "en", entity.VersionDraft, []*entity.Component{
		{ComponentID: "en1", Fields: map[string]any{"title": "English"}, CreatedAt: now, UpdatedAt: now},
	})
	_ = repo.UpsertAllByParent(ctx, "blog", "seo_og", "parent-1", "vi", entity.VersionDraft, []*entity.Component{
		{ComponentID: "vi1", Fields: map[string]any{"title": "Vietnamese"}, CreatedAt: now, UpdatedAt: now},
	})

	enComps, _ := repo.FindByParentComponentID(ctx, "blog", "seo_og", "parent-1", "en", entity.VersionDraft)
	viComps, _ := repo.FindByParentComponentID(ctx, "blog", "seo_og", "parent-1", "vi", entity.VersionDraft)

	if len(enComps) != 1 || len(viComps) != 1 {
		t.Fatalf("expected 1 en and 1 vi component, got %d and %d", len(enComps), len(viComps))
	}

	enTitle, _ := enComps[0].Fields["title"].(string)
	viTitle, _ := viComps[0].Fields["title"].(string)
	if enTitle != "English" {
		t.Errorf("en title = %q, want English", enTitle)
	}
	if viTitle != "Vietnamese" {
		t.Errorf("vi title = %q, want Vietnamese", viTitle)
	}
}
