package content_type_test

import (
	"context"
	"testing"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	repomock "project-abyssoftime-cms-v2/api/internal/domain/repository/mock"
	contenttype "project-abyssoftime-cms-v2/api/internal/usecase/content_type"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

type fakeEntryManager struct {
	getAllFn func(ctx context.Context, contentTypeSlug string) ([]*entity.Document, error)
	deleteFn func(ctx context.Context, contentTypeSlug, documentID string) error
	deleted  []string
}

func (f *fakeEntryManager) GetAll(ctx context.Context, contentTypeSlug string) ([]*entity.Document, error) {
	return f.getAllFn(ctx, contentTypeSlug)
}

func (f *fakeEntryManager) Delete(ctx context.Context, contentTypeSlug, documentID string) error {
	f.deleted = append(f.deleted, documentID)
	return f.deleteFn(ctx, contentTypeSlug, documentID)
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
	docRepo := &repomock.DocumentRepository{}
	syncer := contenttype.NewSyncer(contenttype.New(repo), entries, docRepo)

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

func TestSync_DoesNotCreateSingletonForSingleType(t *testing.T) {
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
	docRepo := &repomock.DocumentRepository{}
	syncer := contenttype.NewSyncer(contenttype.New(repo), entries, docRepo)

	defs := []contenttype.ContentTypeDefinition{
		{Slug: "homepage", Name: "Homepage", Kind: "single"},
	}
	if err := syncer.Sync(ctx, defs); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}
	if len(entries.deleted) != 0 {
		t.Errorf("Sync() should not auto-create entries for single types, deleted %d", len(entries.deleted))
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
	docRepo := &repomock.DocumentRepository{}
	syncer := contenttype.NewSyncer(contenttype.New(repo), entries, docRepo)

	defs := []contenttype.ContentTypeDefinition{
		{Slug: "homepage", Name: "New Name", Kind: "single"},
	}
	if err := syncer.Sync(ctx, defs); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}
	if updated == nil || updated.Name != "New Name" {
		t.Fatalf("Sync() did not update content type, got %+v", updated)
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
	docRepo := &repomock.DocumentRepository{}
	syncer := contenttype.NewSyncer(contenttype.New(repo), entries, docRepo)

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
	docRepo := &repomock.DocumentRepository{}
	syncer := contenttype.NewSyncer(contenttype.New(repo), entries, docRepo)

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
	docRepo := &repomock.DocumentRepository{}
	syncer := contenttype.NewSyncer(contenttype.New(repo), entries, docRepo)

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

func TestSync_CreatesContentTypeWithListFields(t *testing.T) {
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
	docRepo := &repomock.DocumentRepository{}
	syncer := contenttype.NewSyncer(contenttype.New(repo), entries, docRepo)

	defs := []contenttype.ContentTypeDefinition{
		{
			Slug: "articles", Name: "Articles", Kind: "collection",
			ListFields: []string{"title", "slug"},
			Fields: []entity.FieldDefinition{
				{Name: "title", Type: "text"},
				{Name: "slug", Type: "text"},
				{Name: "body", Type: "richtext"},
			},
		},
	}
	if err := syncer.Sync(ctx, defs); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}
	if created == nil {
		t.Fatal("Sync() did not create a content type")
	}
	if len(created.ListFields) != 2 || created.ListFields[0] != "title" || created.ListFields[1] != "slug" {
		t.Errorf("Sync() created ContentType.ListFields = %v, want [title slug]", created.ListFields)
	}
}

func TestSync_UpdatesWhenListFieldsChanged(t *testing.T) {
	existing := &entity.ContentType{
		ID:         "ct-1",
		Slug:       "articles",
		Name:       "Articles",
		Kind:       entity.KindCollection,
		Fields:     []entity.FieldDefinition{{Name: "title", Type: "text"}, {Name: "body", Type: "richtext"}},
		ListFields: []string{"title"},
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
	docRepo := &repomock.DocumentRepository{}
	syncer := contenttype.NewSyncer(contenttype.New(repo), entries, docRepo)

	defs := []contenttype.ContentTypeDefinition{
		{
			Slug: "articles", Name: "Articles", Kind: "collection",
			ListFields: []string{"title", "body"},
			Fields:     []entity.FieldDefinition{{Name: "title", Type: "text"}, {Name: "body", Type: "richtext"}},
		},
	}
	if err := syncer.Sync(ctx, defs); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}
	if updated == nil {
		t.Fatal("Sync() did not update the content type when ListFields changed")
	}
	if len(updated.ListFields) != 2 || updated.ListFields[1] != "body" {
		t.Errorf("Sync() updated ContentType.ListFields = %v, want [title body]", updated.ListFields)
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
		getAllFn: func(_ context.Context, slug string) ([]*entity.Document, error) {
			return []*entity.Document{
				{DocumentID: "doc-1", ContentTypeID: "ct-stale"},
				{DocumentID: "doc-2", ContentTypeID: "ct-stale"},
			}, nil
		},
		deleteFn: func(_ context.Context, _, _ string) error { return nil },
	}
	docRepo := &repomock.DocumentRepository{}
	var droppedSlug string
	docRepo.DropCollectionFn = func(_ context.Context, slug string) error {
		droppedSlug = slug
		return nil
	}
	syncer := contenttype.NewSyncer(contenttype.New(repo), entries, docRepo)

	if err := syncer.Sync(ctx, nil); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}
	if deletedContentTypeID != "ct-stale" {
		t.Errorf("Sync() deleted content type = %q, want ct-stale", deletedContentTypeID)
	}
	if len(entries.deleted) != 2 {
		t.Fatalf("Sync() cascaded to %d entries, want 2", len(entries.deleted))
	}
	if entries.deleted[0] != "doc-1" || entries.deleted[1] != "doc-2" {
		t.Errorf("Sync() deleted document IDs = %v, want [doc-1 doc-2]", entries.deleted)
	}
	if droppedSlug != "old-type" {
		t.Errorf("Sync() dropped collection for slug = %q, want old-type", droppedSlug)
	}
}
