package document_test

import (
	"context"
	"testing"
	"time"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	repomock "project-abyssoftime-cms-v2/api/internal/domain/repository/mock"
	docuc "project-abyssoftime-cms-v2/api/internal/usecase/document"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

var ctx = context.Background()

// ---- Save --------------------------------------------------------------

func TestSave_NewEntry_GeneratesEntryIDAndSetsAudit(t *testing.T) {
	repo := &repomock.DocumentRepository{}
	repo.FindDraftByEntryIDFn = func(_ context.Context, _ string) (*entity.Document, error) {
		return nil, pkgerrors.ErrNotFound
	}
	var upserted *entity.Document
	repo.UpsertDraftFn = func(_ context.Context, doc *entity.Document) error {
		upserted = doc
		return nil
	}
	uc := docuc.New(repo, &repomock.MediaAssetRepository{})

	doc := &entity.Document{ContentTypeID: "ct-1", Data: map[string]any{"title": "Hello"}}
	saved, err := uc.Save(ctx, doc, "user-1")
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if saved.EntryID == "" {
		t.Error("Save() did not generate an EntryID")
	}
	if saved.CreatedBy != "user-1" || saved.UpdatedBy != "user-1" {
		t.Errorf("Save() CreatedBy/UpdatedBy = %q/%q, want user-1/user-1", saved.CreatedBy, saved.UpdatedBy)
	}
	if saved.Locale != "en" {
		t.Errorf("Save() Locale = %q, want en", saved.Locale)
	}
	if upserted != saved {
		t.Error("Save() did not call UpsertDraft with the same document it returned")
	}
}

func TestSave_ExistingEntry_PreservesCreatedAtAndCreatedBy(t *testing.T) {
	createdAt := time.Now().UTC().Add(-time.Hour)
	existing := &entity.Document{
		ID: "rec-1", EntryID: "entry-1", ContentTypeID: "ct-1",
		CreatedAt: createdAt, CreatedBy: "original-author",
	}

	repo := &repomock.DocumentRepository{}
	repo.FindDraftByEntryIDFn = func(_ context.Context, entryID string) (*entity.Document, error) {
		if entryID == "entry-1" {
			return existing, nil
		}
		return nil, pkgerrors.ErrNotFound
	}
	var upserted *entity.Document
	repo.UpsertDraftFn = func(_ context.Context, doc *entity.Document) error {
		upserted = doc
		return nil
	}
	uc := docuc.New(repo, &repomock.MediaAssetRepository{})

	doc := &entity.Document{EntryID: "entry-1", Data: map[string]any{"title": "Updated"}}
	saved, err := uc.Save(ctx, doc, "editor-2")
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if !saved.CreatedAt.Equal(createdAt) {
		t.Errorf("Save() CreatedAt = %v, want %v (preserved)", saved.CreatedAt, createdAt)
	}
	if saved.CreatedBy != "original-author" {
		t.Errorf("Save() CreatedBy = %q, want original-author (preserved)", saved.CreatedBy)
	}
	if saved.UpdatedBy != "editor-2" {
		t.Errorf("Save() UpdatedBy = %q, want editor-2", saved.UpdatedBy)
	}
	if saved.ContentTypeID != "ct-1" {
		t.Errorf("Save() ContentTypeID = %q, want ct-1 (preserved from existing)", saved.ContentTypeID)
	}
	if upserted.ID != "rec-1" {
		t.Errorf("Save() did not preserve the existing record's Mongo _id, got %q", upserted.ID)
	}
}

// ---- Status (pure computation) ------------------------------------------

func TestStatus(t *testing.T) {
	now := time.Now().UTC()
	tests := []struct {
		name      string
		draft     *entity.Document
		published *entity.Document
		want      string
	}{
		{
			name:  "no published record",
			draft: &entity.Document{UpdatedAt: now},
			want:  "draft",
		},
		{
			name:      "draft newer than published",
			draft:     &entity.Document{UpdatedAt: now},
			published: &entity.Document{UpdatedAt: now.Add(-time.Minute)},
			want:      "modified",
		},
		{
			name:      "timestamps match",
			draft:     &entity.Document{UpdatedAt: now},
			published: &entity.Document{UpdatedAt: now},
			want:      "published",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := docuc.Status(tt.draft, tt.published)
			if got != tt.want {
				t.Errorf("Status() = %q, want %q", got, tt.want)
			}
		})
	}
}

// ---- GetForEdit ----------------------------------------------------------

func TestGetForEdit(t *testing.T) {
	now := time.Now().UTC()

	t.Run("draft only", func(t *testing.T) {
		repo := &repomock.DocumentRepository{}
		repo.FindDraftByEntryIDFn = func(_ context.Context, _ string) (*entity.Document, error) {
			return &entity.Document{EntryID: "e1", UpdatedAt: now}, nil
		}
		repo.FindPublishedByEntryIDFn = func(_ context.Context, _ string) (*entity.Document, error) {
			return nil, pkgerrors.ErrNotFound
		}
		uc := docuc.New(repo, &repomock.MediaAssetRepository{})

		draft, status, err := uc.GetForEdit(ctx, "e1")
		if err != nil {
			t.Fatalf("GetForEdit() error = %v", err)
		}
		if draft.EntryID != "e1" {
			t.Errorf("GetForEdit() draft.EntryID = %q, want e1", draft.EntryID)
		}
		if status != "draft" {
			t.Errorf("GetForEdit() status = %q, want draft", status)
		}
	})

	t.Run("draft not found", func(t *testing.T) {
		repo := &repomock.DocumentRepository{}
		repo.FindDraftByEntryIDFn = func(_ context.Context, _ string) (*entity.Document, error) {
			return nil, pkgerrors.ErrNotFound
		}
		uc := docuc.New(repo, &repomock.MediaAssetRepository{})

		_, _, err := uc.GetForEdit(ctx, "missing")
		if !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
			t.Errorf("GetForEdit() error = %v, want ErrNotFound", err)
		}
	})
}

// ---- GetPublished ----------------------------------------------------------

func TestGetPublished(t *testing.T) {
	tests := []struct {
		name    string
		repoErr error
		wantErr error
	}{
		{"found", nil, nil},
		{"not found", pkgerrors.ErrNotFound, pkgerrors.ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repomock.DocumentRepository{}
			repo.FindPublishedByEntryIDFn = func(_ context.Context, entryID string) (*entity.Document, error) {
				if tt.repoErr != nil {
					return nil, tt.repoErr
				}
				return &entity.Document{EntryID: entryID}, nil
			}
			uc := docuc.New(repo, &repomock.MediaAssetRepository{})

			_, err := uc.GetPublished(ctx, "e1")
			if !pkgerrors.Is(err, tt.wantErr) {
				t.Errorf("GetPublished() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

// ---- GetAll ----------------------------------------------------------------

func TestGetAll_ReturnsEntryDrafts(t *testing.T) {
	repo := &repomock.DocumentRepository{}
	repo.FindEntryDraftsByContentTypeFn = func(_ context.Context, contentTypeID string) ([]*entity.Document, error) {
		return []*entity.Document{
			{EntryID: "1", ContentTypeID: contentTypeID},
			{EntryID: "2", ContentTypeID: contentTypeID},
		}, nil
	}
	uc := docuc.New(repo, &repomock.MediaAssetRepository{})

	docs, err := uc.GetAll(ctx, "ct-1")
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}
	if len(docs) != 2 {
		t.Errorf("GetAll() count = %d, want 2", len(docs))
	}
}

// ---- Publish ----------------------------------------------------------------

func TestPublish_CopiesDraftAndSyncsTimestamps(t *testing.T) {
	draftUpdatedAt := time.Now().UTC()
	draft := &entity.Document{
		EntryID: "e1", ContentTypeID: "ct-1", Data: map[string]any{"title": "v2"},
		Locale: "en", UpdatedAt: draftUpdatedAt, UpdatedBy: "editor-1",
	}

	repo := &repomock.DocumentRepository{}
	repo.FindDraftByEntryIDFn = func(_ context.Context, _ string) (*entity.Document, error) {
		return draft, nil
	}
	var published *entity.Document
	repo.UpsertPublishedFn = func(_ context.Context, doc *entity.Document) error {
		published = doc
		return nil
	}
	uc := docuc.New(repo, &repomock.MediaAssetRepository{})

	if err := uc.Publish(ctx, "e1", "publisher-1"); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	if published == nil {
		t.Fatal("Publish() did not call UpsertPublished")
	}
	if !published.UpdatedAt.Equal(draftUpdatedAt) {
		t.Errorf("Publish() published.UpdatedAt = %v, want %v (synced from draft)", published.UpdatedAt, draftUpdatedAt)
	}
	if published.PublishedBy != "publisher-1" {
		t.Errorf("Publish() published.PublishedBy = %q, want publisher-1", published.PublishedBy)
	}
	if published.PublishedAt.IsZero() {
		t.Error("Publish() did not set PublishedAt")
	}
	if published.Data["title"] != "v2" {
		t.Errorf("Publish() did not copy draft data, got %v", published.Data)
	}
}

func TestPublish_DraftNotFound(t *testing.T) {
	repo := &repomock.DocumentRepository{}
	repo.FindDraftByEntryIDFn = func(_ context.Context, _ string) (*entity.Document, error) {
		return nil, pkgerrors.ErrNotFound
	}
	uc := docuc.New(repo, &repomock.MediaAssetRepository{})

	if err := uc.Publish(ctx, "missing", "publisher-1"); !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
		t.Errorf("Publish() error = %v, want ErrNotFound", err)
	}
}

// ---- Unpublish ---------------------------------------------------------------

func TestUnpublish_DeletesPublishedRecord(t *testing.T) {
	var deletedEntryID string
	repo := &repomock.DocumentRepository{}
	repo.DeletePublishedByEntryIDFn = func(_ context.Context, entryID string) error {
		deletedEntryID = entryID
		return nil
	}
	uc := docuc.New(repo, &repomock.MediaAssetRepository{})

	if err := uc.Unpublish(ctx, "e1"); err != nil {
		t.Fatalf("Unpublish() error = %v", err)
	}
	if deletedEntryID != "e1" {
		t.Errorf("Unpublish() deleted entry = %q, want e1", deletedEntryID)
	}
}

// ---- Delete (cascade) ---------------------------------------------------------

func TestDelete_CascadeOrder(t *testing.T) {
	var callOrder []string

	mediaRepo := &repomock.MediaAssetRepository{}
	mediaRepo.DeleteByDocumentRefFn = func(_ context.Context, ref string) error {
		callOrder = append(callOrder, "media:"+ref)
		return nil
	}

	docRepo := &repomock.DocumentRepository{}
	docRepo.DeleteByEntryIDFn = func(_ context.Context, entryID string) error {
		callOrder = append(callOrder, "entry:"+entryID)
		return nil
	}

	uc := docuc.New(docRepo, mediaRepo)
	if err := uc.Delete(ctx, "e1"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if len(callOrder) != 2 {
		t.Fatalf("Delete() called %d operations, want 2", len(callOrder))
	}
	if callOrder[0] != "media:e1" {
		t.Errorf("Delete() first call = %q, want media:e1", callOrder[0])
	}
	if callOrder[1] != "entry:e1" {
		t.Errorf("Delete() second call = %q, want entry:e1", callOrder[1])
	}
}

func TestDelete_MediaError_Aborts(t *testing.T) {
	mediaRepo := &repomock.MediaAssetRepository{}
	mediaRepo.DeleteByDocumentRefFn = func(_ context.Context, _ string) error {
		return pkgerrors.ErrNotFound
	}
	entryDeleteCalled := false
	docRepo := &repomock.DocumentRepository{}
	docRepo.DeleteByEntryIDFn = func(_ context.Context, _ string) error {
		entryDeleteCalled = true
		return nil
	}

	uc := docuc.New(docRepo, mediaRepo)
	err := uc.Delete(ctx, "e1")
	if err == nil {
		t.Error("Delete() should have returned error when media delete fails")
	}
	if entryDeleteCalled {
		t.Error("Delete() must not delete the entry when media cascade fails")
	}
}
