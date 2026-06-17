package content_type_test

import (
	"context"
	"testing"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	repomock "project-abyssoftime-cms-v2/api/internal/domain/repository/mock"
	contenttype "project-abyssoftime-cms-v2/api/internal/usecase/content_type"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

// fakeEntryManager is a minimal test double for content_type.EntryManager.
type fakeEntryManager struct {
	saveFn   func(ctx context.Context, doc *entity.Document, userID string) (*entity.Document, error)
	getAllFn func(ctx context.Context, contentTypeID string) ([]*entity.Document, error)
	deleteFn func(ctx context.Context, id string) error
	saved    []*entity.Document
	deleted  []string
}

func (f *fakeEntryManager) Save(ctx context.Context, doc *entity.Document, userID string) (*entity.Document, error) {
	f.saved = append(f.saved, doc)
	if f.saveFn != nil {
		return f.saveFn(ctx, doc, userID)
	}
	return doc, nil
}

func (f *fakeEntryManager) GetAll(ctx context.Context, contentTypeID string) ([]*entity.Document, error) {
	return f.getAllFn(ctx, contentTypeID)
}

func (f *fakeEntryManager) Delete(ctx context.Context, id string) error {
	f.deleted = append(f.deleted, id)
	return f.deleteFn(ctx, id)
}

func TestSync_CreatesNewDefinitions(t *testing.T) {
	repo := &repomock.ContentTypeRepository{}
	repo.FindAllFn = func(_ context.Context) ([]*entity.ContentType, error) { return nil, nil }
	repo.FindBySlugFn = func(_ context.Context, _ string) (*entity.ContentType, error) {
		return nil, pkgerrors.ErrNotFound
	}
	var created []*entity.ContentType
	repo.CreateFn = func(_ context.Context, ct *entity.ContentType) error {
		ct.ID = "ct-" + ct.Slug
		created = append(created, ct)
		return nil
	}

	entries := &fakeEntryManager{}
	syncer := contenttype.NewSyncer(contenttype.New(repo), entries)

	defs := []contenttype.ContentTypeDefinition{
		{Slug: "homepage", Name: "Homepage", Kind: "single"},
		{Slug: "blog-post", Name: "Blog Post", Kind: "collection"},
	}
	if err := syncer.Sync(ctx, defs); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}
	if len(created) != 2 {
		t.Fatalf("Sync() created %d content types, want 2", len(created))
	}
}

func TestSync_CreatesSingletonEntryForNewSingleType(t *testing.T) {
	repo := &repomock.ContentTypeRepository{}
	repo.FindAllFn = func(_ context.Context) ([]*entity.ContentType, error) { return nil, nil }
	repo.FindBySlugFn = func(_ context.Context, _ string) (*entity.ContentType, error) {
		return nil, pkgerrors.ErrNotFound
	}
	repo.CreateFn = func(_ context.Context, ct *entity.ContentType) error {
		ct.ID = "ct-homepage"
		return nil
	}

	entries := &fakeEntryManager{}
	syncer := contenttype.NewSyncer(contenttype.New(repo), entries)

	defs := []contenttype.ContentTypeDefinition{
		{Slug: "homepage", Name: "Homepage", Kind: "single"},
	}
	if err := syncer.Sync(ctx, defs); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}
	if len(entries.saved) != 1 {
		t.Fatalf("Sync() saved %d singleton entries, want 1", len(entries.saved))
	}
	if entries.saved[0].EntryID != "ct-homepage" {
		t.Errorf("Sync() singleton EntryID = %q, want ct-homepage (= contentTypeID)", entries.saved[0].EntryID)
	}
}

func TestSync_DoesNotCreateSingletonForNewCollectionType(t *testing.T) {
	repo := &repomock.ContentTypeRepository{}
	repo.FindAllFn = func(_ context.Context) ([]*entity.ContentType, error) { return nil, nil }
	repo.FindBySlugFn = func(_ context.Context, _ string) (*entity.ContentType, error) {
		return nil, pkgerrors.ErrNotFound
	}
	repo.CreateFn = func(_ context.Context, ct *entity.ContentType) error {
		ct.ID = "ct-blog"
		return nil
	}

	entries := &fakeEntryManager{}
	syncer := contenttype.NewSyncer(contenttype.New(repo), entries)

	defs := []contenttype.ContentTypeDefinition{
		{Slug: "blog-post", Name: "Blog Post", Kind: "collection"},
	}
	if err := syncer.Sync(ctx, defs); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}
	if len(entries.saved) != 0 {
		t.Errorf("Sync() saved %d entries for a collection type, want 0", len(entries.saved))
	}
}

func TestSync_UpdatesChangedDefinitions(t *testing.T) {
	existing := &entity.ContentType{ID: "ct-1", Slug: "homepage", Name: "Old Name", Kind: entity.KindSingle}

	repo := &repomock.ContentTypeRepository{}
	repo.FindAllFn = func(_ context.Context) ([]*entity.ContentType, error) {
		return []*entity.ContentType{existing}, nil
	}
	repo.FindBySlugFn = func(_ context.Context, slug string) (*entity.ContentType, error) {
		if slug == existing.Slug {
			return existing, nil
		}
		return nil, pkgerrors.ErrNotFound
	}
	var updated *entity.ContentType
	repo.UpdateFn = func(_ context.Context, ct *entity.ContentType) error {
		updated = ct
		return nil
	}

	entries := &fakeEntryManager{}
	syncer := contenttype.NewSyncer(contenttype.New(repo), entries)

	defs := []contenttype.ContentTypeDefinition{
		{Slug: "homepage", Name: "New Name", Kind: "single"},
	}
	if err := syncer.Sync(ctx, defs); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}
	if updated == nil || updated.Name != "New Name" {
		t.Fatalf("Sync() did not update content type, got %+v", updated)
	}
	if len(entries.saved) != 0 {
		t.Errorf("Sync() should not create a singleton when updating an already-existing content type, saved %d", len(entries.saved))
	}
}

func TestSync_UnchangedDefinition_NoOp(t *testing.T) {
	existing := &entity.ContentType{ID: "ct-1", Slug: "homepage", Name: "Homepage", Kind: entity.KindSingle}

	repo := &repomock.ContentTypeRepository{}
	repo.FindAllFn = func(_ context.Context) ([]*entity.ContentType, error) {
		return []*entity.ContentType{existing}, nil
	}
	repo.FindBySlugFn = func(_ context.Context, slug string) (*entity.ContentType, error) {
		return existing, nil
	}
	repo.UpdateFn = func(_ context.Context, _ *entity.ContentType) error {
		t.Error("Update should not be called for an unchanged definition")
		return nil
	}
	repo.CreateFn = func(_ context.Context, _ *entity.ContentType) error {
		t.Error("Create should not be called for an existing definition")
		return nil
	}

	entries := &fakeEntryManager{}
	syncer := contenttype.NewSyncer(contenttype.New(repo), entries)

	defs := []contenttype.ContentTypeDefinition{
		{Slug: "homepage", Name: "Homepage", Kind: "single"},
	}
	if err := syncer.Sync(ctx, defs); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}
}

func TestSync_CreatesContentTypeWithFields(t *testing.T) {
	repo := &repomock.ContentTypeRepository{}
	repo.FindAllFn = func(_ context.Context) ([]*entity.ContentType, error) { return nil, nil }
	repo.FindBySlugFn = func(_ context.Context, _ string) (*entity.ContentType, error) {
		return nil, pkgerrors.ErrNotFound
	}
	var created *entity.ContentType
	repo.CreateFn = func(_ context.Context, ct *entity.ContentType) error {
		ct.ID = "ct-new"
		created = ct
		return nil
	}

	entries := &fakeEntryManager{}
	syncer := contenttype.NewSyncer(contenttype.New(repo), entries)

	fields := []entity.FieldDefinition{
		{Name: "title", Type: "text"},
		{Name: "body", Type: "richtext"},
	}
	defs := []contenttype.ContentTypeDefinition{
		{Slug: "article", Name: "Article", Kind: "collection", Fields: fields},
	}
	if err := syncer.Sync(ctx, defs); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}
	if created == nil {
		t.Fatal("Sync() did not create a content type")
	}
	if len(created.Fields) != 2 {
		t.Fatalf("Sync() created ContentType.Fields len = %d, want 2", len(created.Fields))
	}
	if created.Fields[0].Name != "title" || created.Fields[1].Name != "body" {
		t.Errorf("Sync() created ContentType.Fields = %+v, want [{title text} {body richtext}]", created.Fields)
	}
}

func TestSync_UpdatesFieldsWhenChanged(t *testing.T) {
	existing := &entity.ContentType{
		ID:   "ct-1",
		Slug: "article",
		Name: "Article",
		Kind: entity.KindCollection,
		Fields: []entity.FieldDefinition{
			{Name: "title", Type: "text"},
		},
	}

	repo := &repomock.ContentTypeRepository{}
	repo.FindAllFn = func(_ context.Context) ([]*entity.ContentType, error) {
		return []*entity.ContentType{existing}, nil
	}
	repo.FindBySlugFn = func(_ context.Context, _ string) (*entity.ContentType, error) {
		return existing, nil
	}
	var updated *entity.ContentType
	repo.UpdateFn = func(_ context.Context, ct *entity.ContentType) error {
		updated = ct
		return nil
	}

	entries := &fakeEntryManager{}
	syncer := contenttype.NewSyncer(contenttype.New(repo), entries)

	newFields := []entity.FieldDefinition{
		{Name: "title", Type: "text"},
		{Name: "body", Type: "richtext"},
	}
	defs := []contenttype.ContentTypeDefinition{
		{Slug: "article", Name: "Article", Kind: "collection", Fields: newFields},
	}
	if err := syncer.Sync(ctx, defs); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}
	if updated == nil {
		t.Fatal("Sync() did not update the content type when Fields changed")
	}
	if len(updated.Fields) != 2 {
		t.Fatalf("Sync() updated ContentType.Fields len = %d, want 2", len(updated.Fields))
	}
}

func TestSync_RemovesMissingDefinitions_CascadesEntries(t *testing.T) {
	stale := &entity.ContentType{ID: "ct-stale", Slug: "old-type", Name: "Old Type", Kind: entity.KindCollection}

	repo := &repomock.ContentTypeRepository{}
	repo.FindAllFn = func(_ context.Context) ([]*entity.ContentType, error) {
		return []*entity.ContentType{stale}, nil
	}
	var deletedContentTypeID string
	repo.DeleteFn = func(_ context.Context, id string) error {
		deletedContentTypeID = id
		return nil
	}

	entries := &fakeEntryManager{
		getAllFn: func(_ context.Context, contentTypeID string) ([]*entity.Document, error) {
			// EntryID, not ID (the record's own Mongo _id), is what Delete must receive.
			return []*entity.Document{
				{ID: "rec-1", EntryID: "entry-1", ContentTypeID: contentTypeID},
				{ID: "rec-2", EntryID: "entry-2", ContentTypeID: contentTypeID},
			}, nil
		},
		deleteFn: func(_ context.Context, _ string) error { return nil },
	}
	syncer := contenttype.NewSyncer(contenttype.New(repo), entries)

	// No definitions at all → "old-type" is no longer defined anywhere.
	if err := syncer.Sync(ctx, nil); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}
	if deletedContentTypeID != "ct-stale" {
		t.Errorf("Sync() deleted content type = %q, want ct-stale", deletedContentTypeID)
	}
	if len(entries.deleted) != 2 {
		t.Fatalf("Sync() cascaded to %d entries, want 2", len(entries.deleted))
	}
	if entries.deleted[0] != "entry-1" || entries.deleted[1] != "entry-2" {
		t.Errorf("Sync() deleted entry IDs = %v, want [entry-1 entry-2] (must use EntryID, not ID)", entries.deleted)
	}
}
